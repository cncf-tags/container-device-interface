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

package cdi

import (
	"strings"

	"golang.org/x/mod/semver"

	cdi "github.com/container-orchestrated-devices/container-device-interface/specs-go"
)

const (
	// CurrentVersion is the current vesion of the CDI Spec.
	CurrentVersion = cdi.CurrentVersion

	vCurrent version = "v" + CurrentVersion

	vEarliest version = v020

	v020 version = "v0.2.0"
	v030 version = "v0.3.0"
	v040 version = "v0.4.0"
	v050 version = "v0.5.0"
)

// version represents a semantic version string
type version string

// String returns the string representation of the version.
// This trims a leading v if present.
func (v version) String() string {
	return strings.TrimPrefix(string(v), "v")
}

// LT checks whether a version is less than the specified version.
// Semantic versioning is used to perform the comparison.
func (v version) LT(o version) bool {
	return semver.Compare(string(v), string(o)) < 0
}

// IsLatest checks whether the version is the latest supported version
func (v version) IsLatest() bool {
	return v == vCurrent
}

type requiredFunc func(*cdi.Spec) bool

type requiredVersionMap map[version]requiredFunc

// required stores a map of spec versions to functions to check the required versions.
// Adding new fields / spec versions requires that a `requiredFunc` be implemented and
// this map be updated.
var required = requiredVersionMap{
	v050: requiresV050,
	v040: requiresV040,
}

// minVersion returns the minimum version required for the given spec
func (r requiredVersionMap) minVersion(spec *cdi.Spec) version {
	minVersion := vEarliest

	for specVersion := range validSpecVersions {
		v := version("v" + strings.TrimPrefix(specVersion, "v"))
		if f, ok := r[v]; ok {
			if f(spec) && minVersion.LT(v) {
				minVersion = v
			}
		}
		// If we have already detected the latest version then no later version could be detected
		if minVersion.IsLatest() {
			break
		}
	}

	return minVersion
}

// requiresV050 returns true if the spec uses v0.5.0 features
func requiresV050(spec *cdi.Spec) bool {
	var edits []*cdi.ContainerEdits

	for _, d := range spec.Devices {
		// The v0.5.0 spec allowed device names to start with a digit instead of requiring a letter
		if len(d.Name) > 0 && !isLetter(rune(d.Name[0])) {
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
func requiresV040(spec *cdi.Spec) bool {
	var edits []*cdi.ContainerEdits

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
