/*
   Copyright © 2024 The CDI Authors

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

package producer

import (
	"errors"
	"fmt"
	"strings"

	cdi "tags.cncf.io/container-device-interface/specs-go"
)

const (
	// DefaultValidator implements the basic checks for a valid CDI spec in code.
	DefaultValidator = defaultValidator("default")
)

type defaultValidator string

var _ SpecValidator = (*defaultValidator)(nil)

// errInvalid can be returned if CDI spec validation fails.
var errInvalid = errors.New("invalid")

// Validate performs a default validation on a CDI spec.
func (v defaultValidator) Validate(s *cdi.Spec) (rerr error) {
	defer func() {
		if rerr != nil {
			rerr = errors.Join(rerr, errInvalid)
		}
	}()

	if s == nil {
		return fmt.Errorf("spec is nil")
	}
	if err := cdi.ValidateVersion(s); err != nil {
		return err
	}
	if err := ValidateKind(s.Kind); err != nil {
		return err
	}
	if err := ValidateSpecAnnotations(s.Kind, s.Annotations); err != nil {
		return err
	}
	if err := v.validateEdits(&s.ContainerEdits); err != nil {
		return err
	}

	seen := make(map[string]bool)
	for _, d := range s.Devices {
		if seen[d.Name] {
			return fmt.Errorf("invalid spec, multiple device %q", d.Name)
		}
		seen[d.Name] = true
		if err := v.ValidateDevice(&d, s.Kind); err != nil {
			return fmt.Errorf("invalid device %q: %w", d.Name, err)
		}
	}
	if len(seen) == 0 {
		return fmt.Errorf("invalid spec, no devices")
	}

	return nil
}

// ValidatedDevice validates a CDI device.
// The kind is optional and is used to provide additional context when validating annotations.
func (v defaultValidator) ValidateDevice(d *cdi.Device, kind string) error {
	if err := ValidateDeviceName(d.Name); err != nil {
		return err
	}

	name := d.Name
	if kind != "" {
		name = kind + "=" + name
	}
	if err := ValidateSpecAnnotations(name, d.Annotations); err != nil {
		return err
	}

	if err := v.assertNonEmptyEdits(&d.ContainerEdits); err != nil {
		return err
	}
	if err := v.validateEdits(&d.ContainerEdits); err != nil {
		return err
	}
	return nil
}

func (v defaultValidator) assertNonEmptyEdits(e *cdi.ContainerEdits) error {
	if e == nil {
		return nil
	}
	if len(e.Env) > 0 {
		return nil
	}
	if len(e.DeviceNodes) > 0 {
		return nil
	}
	if len(e.Hooks) > 0 {
		return nil
	}
	if len(e.Mounts) > 0 {
		return nil
	}
	if len(e.AdditionalGIDs) > 0 {
		return nil
	}
	if e.IntelRdt != nil {
		return nil
	}
	return errors.New("empty container edits")
}

func (v defaultValidator) validateEdits(e *cdi.ContainerEdits) error {
	if e == nil {
		return nil
	}
	if err := v.validateEnv(e.Env); err != nil {
		return fmt.Errorf("invalid container edits: %w", err)
	}
	for _, d := range e.DeviceNodes {
		if err := v.validateDeviceNode(d); err != nil {
			return err
		}
	}
	for _, h := range e.Hooks {
		if err := v.validateHook(h); err != nil {
			return err
		}
	}
	for _, m := range e.Mounts {
		if err := v.validateMount(m); err != nil {
			return err
		}
	}
	if err := v.validateIntelRdt(e.IntelRdt); err != nil {
		return err
	}
	return nil
}

func (v defaultValidator) validateEnv(env []string) error {
	for _, v := range env {
		if strings.IndexByte(v, byte('=')) <= 0 {
			return fmt.Errorf("invalid environment variable %q", v)
		}
	}
	return nil
}

func (v defaultValidator) validateDeviceNode(d *cdi.DeviceNode) error {
	validTypes := map[string]struct{}{
		"":  {},
		"b": {},
		"c": {},
		"u": {},
		"p": {},
	}

	if d.Path == "" {
		return errors.New("invalid (empty) device path")
	}
	if _, ok := validTypes[d.Type]; !ok {
		return fmt.Errorf("device %q: invalid type %q", d.Path, d.Type)
	}
	for _, bit := range d.Permissions {
		if bit != 'r' && bit != 'w' && bit != 'm' {
			return fmt.Errorf("device %q: invalid permissions %q",
				d.Path, d.Permissions)
		}
	}
	return nil
}

func (v defaultValidator) validateHook(h *cdi.Hook) error {
	const (
		// PrestartHook is the name of the OCI "prestart" hook.
		PrestartHook = "prestart"
		// CreateRuntimeHook is the name of the OCI "createRuntime" hook.
		CreateRuntimeHook = "createRuntime"
		// CreateContainerHook is the name of the OCI "createContainer" hook.
		CreateContainerHook = "createContainer"
		// StartContainerHook is the name of the OCI "startContainer" hook.
		StartContainerHook = "startContainer"
		// PoststartHook is the name of the OCI "poststart" hook.
		PoststartHook = "poststart"
		// PoststopHook is the name of the OCI "poststop" hook.
		PoststopHook = "poststop"
	)
	validHookNames := map[string]struct{}{
		PrestartHook:        {},
		CreateRuntimeHook:   {},
		CreateContainerHook: {},
		StartContainerHook:  {},
		PoststartHook:       {},
		PoststopHook:        {},
	}

	if _, ok := validHookNames[h.HookName]; !ok {
		return fmt.Errorf("invalid hook name %q", h.HookName)
	}
	if h.Path == "" {
		return fmt.Errorf("invalid hook %q with empty path", h.HookName)
	}
	if err := v.validateEnv(h.Env); err != nil {
		return fmt.Errorf("invalid hook %q: %w", h.HookName, err)
	}
	return nil
}

func (v defaultValidator) validateMount(m *cdi.Mount) error {
	if m.HostPath == "" {
		return errors.New("invalid mount, empty host path")
	}
	if m.ContainerPath == "" {
		return errors.New("invalid mount, empty container path")
	}
	return nil
}

func (v defaultValidator) validateIntelRdt(i *cdi.IntelRdt) error {
	if i == nil {
		return nil
	}
	// ClosID must be a valid Linux filename
	if len(i.ClosID) >= 4096 || i.ClosID == "." || i.ClosID == ".." || strings.ContainsAny(i.ClosID, "/\n") {
		return errors.New("invalid ClosID")
	}
	return nil
}
