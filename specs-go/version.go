/*
   Copyright © The CDI Authors

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

package specs

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	// CurrentVersion is the current version of the Spec.
	CurrentVersion = "0.8.0"

	// vCurrent is the current version as a semver-comparable type
	vCurrent Version = "v" + CurrentVersion

	// These represent the released versions of the CDI specification
	v010 Version = "v0.1.0"
	v020 Version = "v0.2.0"
	v030 Version = "v0.3.0"
	v040 Version = "v0.4.0"
	v050 Version = "v0.5.0"
	v060 Version = "v0.6.0"
	v070 Version = "v0.7.0"
	v080 Version = "v0.8.0"

	// vEarliest is the earliest supported version of the CDI specification
	vEarliest Version = v030
)

// validSpecVersions stores a map of spec versions to functions to check the required versions.
// Adding new fields / spec versions requires that a `requiredFunc` be implemented and
// this map be updated.
var validSpecVersions = requiredVersionMap{
	v010: nil,
	v020: nil,
	v030: nil,
	v040: requiresV040,
	v050: requiresV050,
	v060: requiresV060,
	v070: requiresV070,
	v080: requiresV080,
}

// ValidateVersion checks whether the specified spec version is supported.
func ValidateVersion(version string) error {
	if !validSpecVersions.isValidVersion(version) {
		return fmt.Errorf("invalid version %q", version)
	}
	return nil
}

// MinimumRequiredVersion determines the minimum spec version for the input spec.
func MinimumRequiredVersion(spec *Spec) (string, error) {
	minVersion := validSpecVersions.requiredVersion(spec)
	return minVersion.String(), nil
}

// Version represents a semantic version string
type Version string

// NewVersion creates a version that can be used for semantic version comparisons.
func NewVersion(v string) Version {
	return Version("v" + strings.TrimPrefix(v, "v"))
}

// String returns the string representation of the version.
// This trims a leading v if present.
func (v Version) String() string {
	return strings.TrimPrefix(string(v), "v")
}

// IsGreaterThan checks with a version is greater than the specified version.
func (v Version) IsGreaterThan(o Version) bool {
	return semver.Compare(string(v), string(o)) > 0
}

// IsLatest checks whether the version is the latest supported version
func (v Version) IsLatest() bool {
	return v == vCurrent
}

type requiredFunc func(*Spec) bool

type requiredVersionMap map[Version]requiredFunc

// isValidVersion checks whether the specified version is valid.
// A version is valid if it is contained in the required version map.
func (r requiredVersionMap) isValidVersion(specVersion string) bool {
	_, ok := validSpecVersions[NewVersion(specVersion)]

	return ok
}

// requiredVersion returns the minimum version required for the given spec
func (r requiredVersionMap) requiredVersion(spec *Spec) Version {
	minVersion := vEarliest

	for v, isRequired := range validSpecVersions {
		if isRequired == nil {
			continue
		}
		if isRequired(spec) && v.IsGreaterThan(minVersion) {
			minVersion = v
		}
		// If we have already detected the latest version then no later version could be detected
		if minVersion.IsLatest() {
			break
		}
	}

	return minVersion
}

// requiresV080 returns true if the spec uses v0.8.0 features.
// Since the v0.8.0 spec bump was due to the removed .ToOCI functions on the
// spec types, there are explicit spec changes.
func requiresV080(_ *Spec) bool {
	return false
}

// requiresV070 returns true if the spec uses v0.7.0 features
func requiresV070(spec *Spec) bool {
	if spec.ContainerEdits.IntelRdt != nil {
		return true
	}
	// The v0.7.0 spec allows additional GIDs to be specified at a spec level.
	if len(spec.ContainerEdits.AdditionalGIDs) > 0 {
		return true
	}

	for _, d := range spec.Devices {
		if d.ContainerEdits.IntelRdt != nil {
			return true
		}
		// The v0.7.0 spec allows additional GIDs to be specified at a device level.
		if len(d.ContainerEdits.AdditionalGIDs) > 0 {
			return true
		}
	}

	return false
}

// requiresV060 returns true if the spec uses v0.6.0 features
func requiresV060(spec *Spec) bool {
	// The v0.6.0 spec allows annotations to be specified at a spec level
	for range spec.Annotations {
		return true
	}

	// The v0.6.0 spec allows annotations to be specified at a device level
	for _, d := range spec.Devices {
		for range d.Annotations {
			return true
		}
	}

	// The v0.6.0 spec allows dots "." in Kind name label (class)
	vendor, class := ParseQualifier(spec.Kind)
	if vendor != "" {
		if strings.ContainsRune(class, '.') {
			return true
		}
	}

	return false
}

// requiresV050 returns true if the spec uses v0.5.0 features
func requiresV050(spec *Spec) bool {
	var edits []*ContainerEdits

	for _, d := range spec.Devices {
		// The v0.5.0 spec allowed device names to start with a digit instead of requiring a letter
		if len(d.Name) > 0 && !IsLetter(rune(d.Name[0])) {
			return true
		}
		edits = append(edits, &d.ContainerEdits)
	}

	edits = append(edits, &spec.ContainerEdits)
	for _, e := range edits {
		for _, dn := range e.DeviceNodes {
			// The HostPath field was added in v0.5.0
			if dn.HostPath != "" {
				return true
			}
		}
	}
	return false
}

// requiresV040 returns true if the spec uses v0.4.0 features
func requiresV040(spec *Spec) bool {
	var edits []*ContainerEdits

	for _, d := range spec.Devices {
		edits = append(edits, &d.ContainerEdits)
	}

	edits = append(edits, &spec.ContainerEdits)
	for _, e := range edits {
		for _, m := range e.Mounts {
			// The Type field was added in v0.4.0
			if m.Type != "" {
				return true
			}
		}
	}
	return false
}
