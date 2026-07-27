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
	CurrentVersion = "1.1.0"

	// vCurrent is the current version as a semver-comparable type
	vCurrent version = "v" + CurrentVersion

	// These represent the released versions of the CDI specification
	v010 version = "v0.1.0"
	v020 version = "v0.2.0"
	v030 version = "v0.3.0"
	v040 version = "v0.4.0"
	v050 version = "v0.5.0"
	v060 version = "v0.6.0"
	v070 version = "v0.7.0"
	v080 version = "v0.8.0"
	v100 version = "v1.0.0"
	v110 version = "v1.1.0"

	// vEarliest is the earliest supported version of the CDI specification
	vEarliest version = v030
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
	v100: requiresV100,
	v110: requiresV110,
}

// ValidateVersion checks whether the specified spec version is valid.
// In addition to checking whether the spec version is in the set of known versions,
// the spec is inspected to determine whether the features used are available in specified
// version.
func ValidateVersion(spec *Spec) error {
	if !validSpecVersions.isValidVersion(spec.Version) {
		return fmt.Errorf("invalid version %q", spec.Version)
	}
	if err := validateVersionedSpec(spec); err != nil {
		return err
	}
	minVersion, err := MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("could not determine minimum required version: %w", err)
	}
	if newVersion(minVersion).isGreaterThan(newVersion(spec.Version)) {
		return fmt.Errorf("the spec version must be at least v%v", minVersion)
	}
	return nil
}

// ValidateVersionContents checks versioned fields in raw CDI Spec contents.
// This complements ValidateVersion for fields whose mere presence matters but
// whose zero value is indistinguishable from omission after unmarshalling.
// Invalid or missing cdiVersion values are left for ValidateVersion to report.
func ValidateVersionContents(contents map[string]interface{}) error {
	specVersion, ok := getString(contents, "cdiVersion")
	if !ok {
		return nil
	}
	if !validSpecVersions.isValidVersion(specVersion) {
		return nil
	}

	version := newVersion(specVersion)
	if err := validateVersionedContainerEditsContents(version, "containerEdits", getMap(contents, "containerEdits")); err != nil {
		return err
	}

	devices, _ := contents["devices"].([]interface{})
	for idx, device := range devices {
		deviceContents, _ := device.(map[string]interface{})
		path := fmt.Sprintf("devices[%d].containerEdits", idx)
		if err := validateVersionedContainerEditsContents(version, path, getMap(deviceContents, "containerEdits")); err != nil {
			return err
		}
	}

	return nil
}

// MinimumRequiredVersion determines the minimum spec version for the input spec.
func MinimumRequiredVersion(spec *Spec) (string, error) {
	minVersion := validSpecVersions.requiredVersion(spec)
	return minVersion.String(), nil
}

// version represents a semantic version string
type version string

// newVersion creates a version that can be used for semantic version comparisons.
func newVersion(v string) version {
	return version("v" + strings.TrimPrefix(v, "v"))
}

// String returns the string representation of the version.
// This trims a leading v if present.
func (v version) String() string {
	return strings.TrimPrefix(string(v), "v")
}

// isGreaterThan checks with a version is greater than the specified version.
func (v version) isGreaterThan(o version) bool {
	return semver.Compare(string(v), string(o)) > 0
}

// isLatest checks whether the version is the latest supported version
func (v version) isLatest() bool {
	return v == vCurrent
}

func getString(contents map[string]interface{}, field string) (string, bool) {
	if contents == nil {
		return "", false
	}
	value, ok := contents[field]
	if !ok {
		return "", false
	}
	valueString, ok := value.(string)
	return valueString, ok
}

func getMap(contents map[string]interface{}, field string) map[string]interface{} {
	if contents == nil {
		return nil
	}
	value, ok := contents[field]
	if !ok {
		return nil
	}
	valueMap, _ := value.(map[string]interface{})
	return valueMap
}

func validateVersionedContainerEditsContents(specVersion version, path string, edits map[string]interface{}) error {
	if edits != nil {
		if _, ok := edits["netDevices"]; ok && v110.isGreaterThan(specVersion) {
			return fmt.Errorf("%s.netDevices requires CDI spec version at least %s", path, v110)
		}
		if err := validateVersionedIntelRdtContents(specVersion, path+".intelRdt", getMap(edits, "intelRdt")); err != nil {
			return err
		}
	}

	return nil
}

func validateVersionedIntelRdtContents(specVersion version, path string, intelRdt map[string]interface{}) error {
	if intelRdt == nil {
		return nil
	}

	for _, field := range []string{"enableCMT", "enableMBM"} {
		if _, ok := intelRdt[field]; ok && !v110.isGreaterThan(specVersion) {
			return fmt.Errorf("%s.%s is not valid for CDI spec version %s", path, field, specVersion)
		}
	}
	for _, field := range []string{"schemata", "enableMonitoring"} {
		if _, ok := intelRdt[field]; ok && v110.isGreaterThan(specVersion) {
			return fmt.Errorf("%s.%s requires CDI spec version at least %s", path, field, v110)
		}
	}

	return nil
}

func validateVersionedSpec(spec *Spec) error {
	specVersion := newVersion(spec.Version)
	if err := validateVersionedIntelRdt(specVersion, "containerEdits.intelRdt", spec.ContainerEdits.IntelRdt); err != nil {
		return err
	}

	for idx, dev := range spec.Devices {
		path := fmt.Sprintf("devices[%d].containerEdits.intelRdt", idx)
		if err := validateVersionedIntelRdt(specVersion, path, dev.ContainerEdits.IntelRdt); err != nil {
			return err
		}
	}

	return nil
}

func validateVersionedIntelRdt(specVersion version, path string, intelRdt *IntelRdt) error {
	if intelRdt == nil || v110.isGreaterThan(specVersion) {
		return nil
	}

	switch {
	case intelRdt.EnableCMT:
		return fmt.Errorf("%s.enableCMT is not valid for CDI spec version %s", path, specVersion)
	case intelRdt.EnableMBM:
		return fmt.Errorf("%s.enableMBM is not valid for CDI spec version %s", path, specVersion)
	}

	return nil
}

type requiredFunc func(*Spec) bool

type requiredVersionMap map[version]requiredFunc

// isValidVersion checks whether the specified version is valid.
// A version is valid if it is contained in the required version map.
func (r requiredVersionMap) isValidVersion(specVersion string) bool {
	_, ok := validSpecVersions[newVersion(specVersion)]

	return ok
}

// requiredVersion returns the minimum version required for the given spec
func (r requiredVersionMap) requiredVersion(spec *Spec) version {
	minVersion := vEarliest

	for v, isRequired := range validSpecVersions {
		if isRequired == nil {
			continue
		}
		if isRequired(spec) && v.isGreaterThan(minVersion) {
			minVersion = v
		}
		// If we have already detected the latest version then no later version could be detected
		if minVersion.isLatest() {
			break
		}
	}

	return minVersion
}

// requiresV110 returns true if the spec uses v1.1.0 features.
func requiresV110(spec *Spec) bool {
	if i := spec.ContainerEdits.IntelRdt; i != nil {
		if i.Schemata != nil || i.EnableMonitoring {
			return true
		}
	}

	if len(spec.ContainerEdits.NetDevices) > 0 {
		return true
	}

	for _, dev := range spec.Devices {
		if i := dev.ContainerEdits.IntelRdt; i != nil {
			if i.Schemata != nil || i.EnableMonitoring {
				return true
			}
		}

		if len(dev.ContainerEdits.NetDevices) > 0 {
			return true
		}
	}

	return false
}

// requiresV100 returns true if the spec uses v1.0.0 features.
// Since the v1.0.0 spec bump was due to moving the minimum version checks to
// the spec package, there are no explicit spec changes.
func requiresV100(_ *Spec) bool {
	return false
}

// requiresV080 returns true if the spec uses v0.8.0 features.
// Since the v0.8.0 spec bump was due to the removed .ToOCI functions on the
// spec types, there are no explicit spec changes.
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
	if !strings.Contains(spec.Kind, "/") {
		return false
	}
	class := strings.SplitN(spec.Kind, "/", 2)[1]
	return strings.Contains(class, ".")
}

// requiresV050 returns true if the spec uses v0.5.0 features
func requiresV050(spec *Spec) bool {
	var edits []*ContainerEdits

	for _, d := range spec.Devices {
		// The v0.5.0 spec allowed device name to start with a digit
		if len(d.Name) > 0 && '0' <= d.Name[0] && d.Name[0] <= '9' {
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
