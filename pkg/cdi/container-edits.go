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
	"sort"
	"strings"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	ocigen "github.com/opencontainers/runtime-tools/generate"

	"tags.cncf.io/container-device-interface/pkg/cdi/producer/validator"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

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

var (
	// Names of recognized hooks.
	validHookNames = map[string]struct{}{
		PrestartHook:        {},
		CreateRuntimeHook:   {},
		CreateContainerHook: {},
		StartContainerHook:  {},
		PoststartHook:       {},
		PoststopHook:        {},
	}
)

// ContainerEdits represent updates to be applied to an OCI Spec.
// These updates can be specific to a CDI device, or they can be
// specific to a CDI Spec. In the former case these edits should
// be applied to all OCI Specs where the corresponding CDI device
// is injected. In the latter case, these edits should be applied
// to all OCI Specs where at least one devices from the CDI Spec
// is injected.
type ContainerEdits struct {
	*cdi.ContainerEdits
}

// Apply edits to the given OCI Spec. Updates the OCI Spec in place.
// Returns an error if the update fails.
func (e *ContainerEdits) Apply(spec *oci.Spec) error {
	if spec == nil {
		return errors.New("can't edit nil OCI Spec")
	}
	if e == nil || e.ContainerEdits == nil {
		return nil
	}

	specgen := ocigen.NewFromSpec(spec)
	if len(e.Env) > 0 {
		specgen.AddMultipleProcessEnv(e.Env)
	}

	for _, d := range e.DeviceNodes {
		dn := DeviceNode{d}

		err := dn.fillMissingInfo()
		if err != nil {
			return err
		}
		dev := dn.toOCI()
		if dev.UID == nil && spec.Process != nil {
			if uid := spec.Process.User.UID; uid > 0 {
				dev.UID = &uid
			}
		}
		if dev.GID == nil && spec.Process != nil {
			if gid := spec.Process.User.GID; gid > 0 {
				dev.GID = &gid
			}
		}

		specgen.RemoveDevice(dev.Path)
		specgen.AddDevice(dev)

		if dev.Type == "b" || dev.Type == "c" {
			access := d.Permissions
			if access == "" {
				access = "rwm"
			}
			specgen.AddLinuxResourcesDevice(true, dev.Type, &dev.Major, &dev.Minor, access)
		}
	}

	if len(e.Mounts) > 0 {
		for _, m := range e.Mounts {
			specgen.RemoveMount(m.ContainerPath)
			specgen.AddMount((&Mount{m}).toOCI())
		}
		sortMounts(&specgen)
	}

	for _, h := range e.Hooks {
		ociHook := (&Hook{h}).toOCI()
		switch h.HookName {
		case PrestartHook:
			specgen.AddPreStartHook(ociHook)
		case PoststartHook:
			specgen.AddPostStartHook(ociHook)
		case PoststopHook:
			specgen.AddPostStopHook(ociHook)
			// TODO: Maybe runtime-tools/generate should be updated with these...
		case CreateRuntimeHook:
			ensureOCIHooks(spec)
			spec.Hooks.CreateRuntime = append(spec.Hooks.CreateRuntime, ociHook)
		case CreateContainerHook:
			ensureOCIHooks(spec)
			spec.Hooks.CreateContainer = append(spec.Hooks.CreateContainer, ociHook)
		case StartContainerHook:
			ensureOCIHooks(spec)
			spec.Hooks.StartContainer = append(spec.Hooks.StartContainer, ociHook)
		default:
			return fmt.Errorf("unknown hook name %q", h.HookName)
		}
	}

	if e.IntelRdt != nil {
		// The specgen is missing functionality to set all parameters so we
		// just piggy-back on it to initialize all structs and the copy over.
		specgen.SetLinuxIntelRdtClosID(e.IntelRdt.ClosID)
		spec.Linux.IntelRdt = (&IntelRdt{e.IntelRdt}).toOCI()
	}

	for _, additionalGID := range e.AdditionalGIDs {
		if additionalGID == 0 {
			continue
		}
		specgen.AddProcessAdditionalGid(additionalGID)
	}

	return nil
}

// Validate container edits.
func (e *ContainerEdits) Validate() error {
	if e == nil || e.ContainerEdits == nil {
		return nil
	}
	return validator.Default.ValidateAny(e.ContainerEdits)
}

// Append other edits into this one. If called with a nil receiver,
// allocates and returns newly allocated edits.
func (e *ContainerEdits) Append(o *ContainerEdits) *ContainerEdits {
	if o == nil || o.ContainerEdits == nil {
		return e
	}
	if e == nil {
		e = &ContainerEdits{}
	}
	if e.ContainerEdits == nil {
		e.ContainerEdits = &cdi.ContainerEdits{}
	}

	e.Env = append(e.Env, o.Env...)
	e.DeviceNodes = append(e.DeviceNodes, o.DeviceNodes...)
	e.Hooks = append(e.Hooks, o.Hooks...)
	e.Mounts = append(e.Mounts, o.Mounts...)
	if o.IntelRdt != nil {
		e.IntelRdt = o.IntelRdt
	}
	e.AdditionalGIDs = append(e.AdditionalGIDs, o.AdditionalGIDs...)

	return e
}

// DeviceNode is a CDI Spec DeviceNode wrapper, used for validating DeviceNodes.
type DeviceNode struct {
	*cdi.DeviceNode
}

// Validate a CDI Spec DeviceNode.
func (d *DeviceNode) Validate() error {
	return validator.Default.ValidateAny(d.DeviceNode)
}

// Hook is a CDI Spec Hook wrapper, used for validating hooks.
type Hook struct {
	*cdi.Hook
}

// Validate a hook.
func (h *Hook) Validate() error {
	return validator.Default.ValidateAny(h.Hook)
}

// Mount is a CDI Mount wrapper, used for validating mounts.
type Mount struct {
	*cdi.Mount
}

// Validate a mount.
func (m *Mount) Validate() error {
	return validator.Default.ValidateAny(m.Mount)
}

// IntelRdt is a CDI IntelRdt wrapper.
// This is used for validation and conversion to OCI specifications.
type IntelRdt struct {
	*cdi.IntelRdt
}

// ValidateIntelRdt validates the IntelRdt configuration.
//
// Deprecated: ValidateIntelRdt is deprecated use IntelRdt.Validate() instead.
func ValidateIntelRdt(i *cdi.IntelRdt) error {
	return (&IntelRdt{i}).Validate()
}

// Validate validates the IntelRdt configuration.
func (i *IntelRdt) Validate() error {
	return validator.Default.ValidateAny(i.IntelRdt)
}

// Ensure OCI Spec hooks are not nil so we can add hooks.
func ensureOCIHooks(spec *oci.Spec) {
	if spec.Hooks == nil {
		spec.Hooks = &oci.Hooks{}
	}
}

// sortMounts sorts the mounts in the given OCI Spec.
func sortMounts(specgen *ocigen.Generator) {
	mounts := specgen.Mounts()
	specgen.ClearMounts()
	sort.Sort(orderedMounts(mounts))
	specgen.Config.Mounts = mounts
}

// orderedMounts defines how to sort an OCI Spec Mount slice.
// This is the almost the same implementation sa used by CRI-O and Docker,
// with a minor tweak for stable sorting order (easier to test):
//
//	https://github.com/moby/moby/blob/17.05.x/daemon/volumes.go#L26
type orderedMounts []oci.Mount

// Len returns the number of mounts. Used in sorting.
func (m orderedMounts) Len() int {
	return len(m)
}

// Less returns true if the number of parts (a/b/c would be 3 parts) in the
// mount indexed by parameter 1 is less than that of the mount indexed by
// parameter 2. Used in sorting.
func (m orderedMounts) Less(i, j int) bool {
	ip, jp := m.parts(i), m.parts(j)
	if ip < jp {
		return true
	}
	if jp < ip {
		return false
	}
	return m[i].Destination < m[j].Destination
}

// Swap swaps two items in an array of mounts. Used in sorting
func (m orderedMounts) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// parts returns the number of parts in the destination of a mount. Used in sorting.
func (m orderedMounts) parts(i int) int {
	return strings.Count(filepath.Clean(m[i].Destination), string(os.PathSeparator))
}
