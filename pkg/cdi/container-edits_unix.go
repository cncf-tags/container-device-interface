//go:build !windows

/*
   Copyright © 2021 The CDI Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cdi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

const (
	blockDevice = "b"
	charDevice  = "c" // or "u"
	fifoDevice  = "p"
)

type deviceInfo struct {
	// cgroup properties
	deviceType string
	major      int64
	minor      int64

	// device node properties
	fileMode os.FileMode
}

// deviceInfoFromPath takes the path to a device and returns its type,
// major and minor device numbers.
//
// It was adapted from https://github.com/opencontainers/runc/blob/v1.1.9/libcontainer/devices/device_unix.go#L30-L69
func deviceInfoFromPath(path string) (*deviceInfo, error) {
	var stat unix.Stat_t
	err := unix.Lstat(path, &stat)
	if err != nil {
		return nil, err
	}

	var devType string
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFBLK:
		devType = blockDevice
	case unix.S_IFCHR:
		devType = charDevice
	case unix.S_IFIFO:
		devType = fifoDevice
	default:
		return nil, errors.New("not a device node")
	}
	devNumber := uint64(stat.Rdev) //nolint:unconvert // Rdev is uint32 on e.g. MIPS.

	di := deviceInfo{
		deviceType: devType,
		major:      int64(unix.Major(devNumber)),
		minor:      int64(unix.Minor(devNumber)),
		fileMode:   os.FileMode(stat.Mode &^ unix.S_IFMT),
	}

	return &di, nil
}

// validateWildcards validates the wildcard patterns (if any) of a device node.
func (d *DeviceNode) validateWildcards() error {
	pathIsPattern, hostPathIsPattern := cdi.HasWildcards(d.Path), cdi.HasWildcards(d.HostPath)
	if !pathIsPattern && !hostPathIsPattern {
		return nil
	}

	for _, p := range []string{d.Path, d.HostPath} {
		if !cdi.HasWildcards(p) {
			continue
		}
		// Require absolute path, and a "clean" path (rule out paths like "/dev/../etc/*")
		if !filepath.IsAbs(p) || filepath.Clean(p) != p {
			return fmt.Errorf("device %q: wildcard pattern %q is not an absolute, cleaned path", d.Path, p)
		}
		if _, err := filepath.Match(p, ""); err != nil {
			return fmt.Errorf("device %q: invalid wildcard pattern %q: %w", d.Path, p, err)
		}
		if cdi.HasWildcards(filepath.Dir(p)) {
			return fmt.Errorf("device %q: wildcards are only allowed in the last element of the path %q",
				d.Path, p)
		}
	}

	// The host device node determines the type and the device numbers, thus these must not be set
	switch {
	case d.Type != "":
		return fmt.Errorf("device %q: type must not be set for a wildcard pattern", d.Path)
	case d.Major != 0 || d.Minor != 0:
		return fmt.Errorf("device %q: major/minor must not be set for a wildcard pattern", d.Path)
	}

	if d.HostPath == "" {
		return nil
	}

	// The patterns of path and host path must be identical so that the container
	// path of every match is unambiguous.
	if filepath.Base(d.Path) != filepath.Base(d.HostPath) {
		return fmt.Errorf("device %q: last element of path and hostPath %q must be an identical pattern",
			d.Path, d.HostPath)
	}

	return nil
}

// expandWildcards expands device nodes that have paths with wildcard patterns into
// actual device nodes to be injected into the container. Matches which are not
// device nodes, e.g. regular files, directories and symlinks, are ignored. A
// pattern matching no host device nodes expands to an empty list.
func expandWildcards(nodes []*cdi.DeviceNode) ([]*cdi.DeviceNode, error) {
	expanded := make([]*cdi.DeviceNode, 0, len(nodes))
	seen := make(map[string]bool, len(nodes))

	for _, d := range nodes {
		if !cdi.HasWildcards(d.Path) && !cdi.HasWildcards(d.HostPath) {
			// NOTE: we don't dedup explicit (non-wildcard) device nodes (to not change the existing behavior).
			// Should we drop that for wildcard expansion, too(?)
			seen[d.Path] = true
			expanded = append(expanded, d)
			continue
		}

		matches, err := (&DeviceNode{d}).expand()
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			// Don't inject a device node twice
			if !seen[m.Path] {
				seen[m.Path] = true
				expanded = append(expanded, m)
			}
		}
	}

	return expanded, nil
}

// expand returns one new device node per host device node matching the path pattern of the device node.
func (d *DeviceNode) expand() ([]*cdi.DeviceNode, error) {
	pattern := d.HostPath
	if pattern == "" {
		pattern = d.Path
	}

	// Glob returns matches in sorted order, making this deterministic
	matches, err := filepath.Glob(pattern)
	if err != nil {
		// NOTE: the pattern is checked by Validate() so we should never end up here
		return nil, fmt.Errorf("invalid device node pattern %q: %w", pattern, err)
	}

	expanded := make([]*cdi.DeviceNode, 0, len(matches))
	for _, hostPath := range matches {
		// Ignores matches which are not device nodes, also filtering out symlinks (like /dev/dri/by-path/*)
		if _, err := deviceInfoFromPath(hostPath); err != nil {
			continue
		}

		node := *d.DeviceNode
		if node.HostPath == "" {
			node.Path = hostPath
		} else {
			// NOTE: Validate() ensures that the patterns of path and hostPath are identical so we can safely do this
			node.Path = filepath.Join(filepath.Dir(d.Path), filepath.Base(hostPath))
			node.HostPath = hostPath
		}
		expanded = append(expanded, &node)
	}

	return expanded, nil
}

// fillMissingInfo fills in missing mandatory attributes from the host device.
func (d *DeviceNode) fillMissingInfo() error {
	hasMinimalSpecification := d.Type != "" && (d.Major != 0 || d.Type == fifoDevice)

	// Ensure that the host path and the container path match.
	if d.HostPath == "" {
		d.HostPath = d.Path
	}

	// Try to extract the device info from the host path.
	di, err := deviceInfoFromPath(d.HostPath)
	if err != nil {
		// The error is only considered fatal if the device is not already
		// minimally specified since it is allowed for a device vendor to fully
		// specify a device node specification.
		if !hasMinimalSpecification {
			return fmt.Errorf("failed to stat CDI host device %q: %w", d.HostPath, err)
		}
		return nil
	}

	// Even for minimally-specified device nodes, we update the file mode if
	// required. This is useful for rootless containers where device node
	// requests may be treated as bind mounts.
	if d.FileMode == nil {
		d.FileMode = &di.fileMode
	}

	// If the device is minimally specified, we make no further updates and
	// don't perform additional checks.
	if hasMinimalSpecification {
		return nil
	}

	if d.Type == "" {
		d.Type = di.deviceType
	}
	if d.Type != di.deviceType {
		return fmt.Errorf("CDI device (%q, %q), host type mismatch (%s, %s)",
			d.Path, d.HostPath, d.Type, di.deviceType)
	}

	// For a fifoDevice, we do not update the major and minor number.
	if d.Type == fifoDevice {
		return nil
	}

	// Update the major and minor number for the device node if required.
	if d.Major == 0 {
		d.Major = di.major
		d.Minor = di.minor
	}

	return nil
}
