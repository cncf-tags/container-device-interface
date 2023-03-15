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
	"fmt"

	"github.com/container-orchestrated-devices/container-device-interface/internal/validation"
	cdi "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	oci "github.com/opencontainers/runtime-spec/specs-go"
)

// Device represents a CDI device of a Spec.
type Device struct {
	*cdi.Device
	spec *Spec
}

// Create a new Device, associate it with the given Spec.
func newDevice(spec *Spec, d cdi.Device) (*Device, error) {
	dev := &Device{
		Device: &d,
		spec:   spec,
	}

	if err := dev.validate(); err != nil {
		return nil, err
	}

	return dev, nil
}

// GetSpec returns the Spec this device is defined in.
func (d *Device) GetSpec() *Spec {
	return d.spec
}

// GetNames returns the possible names for the device including aliases
func (d *Device) GetNames() []string {
	names := []string{d.Name}
	return append(names, d.Aliases...)
}

// GetQualifiedName returns the qualified name for this device.
func (d *Device) GetQualifiedName() string {
	return QualifiedName(d.spec.GetVendor(), d.spec.GetClass(), d.Name)
}

// GetQualifiedNames returns the qualified names for this device and its aliases.
func (d *Device) GetQualifiedNames() []string {
	var qualifiedNames []string
	for _, name := range d.GetNames() {
		qualified := QualifiedName(d.spec.GetVendor(), d.spec.GetClass(), name)
		qualifiedNames = append(qualifiedNames, qualified)
	}
	return qualifiedNames
}

// ApplyEdits applies the device-speific container edits to an OCI Spec.
func (d *Device) ApplyEdits(ociSpec *oci.Spec) error {
	return d.edits().Apply(ociSpec)
}

// edits returns the applicable container edits for this spec.
func (d *Device) edits() *ContainerEdits {
	return &ContainerEdits{&d.ContainerEdits}
}

// Validate the device.
func (d *Device) validate() error {
	for _, name := range d.GetNames() {
		if err := ValidateDeviceName(name); err != nil {
			return err
		}
	}
	name := d.Name
	if d.spec != nil {
		name = d.GetQualifiedName()
	}
	if err := validation.ValidateSpecAnnotations(name, d.Annotations); err != nil {
		return err
	}
	edits := d.edits()
	if edits.isEmpty() {
		return fmt.Errorf("invalid device, empty device edits")
	}
	if err := edits.Validate(); err != nil {
		return fmt.Errorf("invalid device %q: %w", d.Name, err)
	}
	return nil
}
