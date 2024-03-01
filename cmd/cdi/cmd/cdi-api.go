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

package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	gen "github.com/opencontainers/runtime-tools/generate"
	"tags.cncf.io/container-device-interface/pkg/cdi"
)

func cdiListVendors() {
	var (
		cache   = cdi.GetDefaultCache()
		vendors = cache.ListVendors()
	)

	if len(vendors) == 0 {
		fmt.Printf("No CDI vendors found.\n")
		return
	}

	fmt.Printf("CDI vendors found:\n")
	for idx, vendor := range vendors {
		fmt.Printf("  %d. %q (%d CDI Spec Files)\n", idx, vendor,
			len(cache.GetVendorSpecs(vendor)))
	}
}

func cdiListClasses() {
	var (
		cache   = cdi.GetDefaultCache()
		vendors = map[string][]string{}
	)

	for _, class := range cache.ListClasses() {
		vendors[class] = []string{}
		for _, vendor := range cache.ListVendors() {
			for _, spec := range cache.GetVendorSpecs(vendor) {
				if spec.GetClass() == class {
					vendors[class] = append(vendors[class], vendor)
				}
			}
		}
	}

	if len(vendors) == 0 {
		fmt.Printf("No CDI device classes found.\n")
		return
	}

	fmt.Printf("CDI device classes found:\n")
	for idx, class := range cache.ListClasses() {
		sort.Strings(vendors[class])
		fmt.Printf("  %d. %s (%d vendors: %s)\n", idx, class,
			len(vendors[class]), strings.Join(vendors[class], ", "))
	}
}

func cdiListDevices(verbose bool, format string) {
	var (
		cache   = cdi.GetDefaultCache()
		devices = cache.ListDevices()
	)

	if len(devices) == 0 {
		fmt.Printf("No CDI devices found.\n")
		return
	}

	fmt.Printf("CDI devices found:\n")
	for idx, device := range devices {
		cdiPrintDevice(idx, cache.GetDevice(device), verbose, format, 2)
	}
}

func cdiPrintDevice(idx int, dev *cdi.Device, verbose bool, format string, level int) {
	if !verbose {
		if idx >= 0 {
			fmt.Printf("%s%d. %s\n", indent(level), idx, dev.GetQualifiedName())
			return
		}
		fmt.Printf("%s%s\n", indent(level), dev.GetQualifiedName())
		return
	}

	var (
		spec = dev.GetSpec()
	)

	format = chooseFormat(format, spec.GetPath())

	fmt.Printf("  %s (%s)\n", dev.GetQualifiedName(), spec.GetPath())
	fmt.Printf("%s", marshalObject(level+2, dev.Device, format))
	edits := spec.ContainerEdits
	if len(edits.Env)+len(edits.DeviceNodes)+len(edits.Hooks)+len(edits.Mounts) > 0 {
		fmt.Printf("%s global Spec containerEdits:\n", indent(level+2))
		fmt.Printf("%s", marshalObject(level+4, spec.ContainerEdits, format))
	}
}

func cdiShowSpecDirs() {
	var (
		cache     = cdi.GetDefaultCache()
		specDirs  = cache.GetSpecDirectories()
		cdiErrors = cache.GetErrors()
	)
	fmt.Printf("CDI Spec directories in use:\n")
	for prio, dir := range specDirs {
		fmt.Printf("  %s (priority %d)\n", dir, prio)
		for path, specErrors := range cdiErrors {
			if filepath.Dir(path) != dir {
				continue
			}
			for _, err := range specErrors {
				fmt.Printf("    - has error %v\n", err)
			}
		}
	}
}

func cdiInjectDevices(format string, ociSpec *oci.Spec, patterns []string) error {
	var (
		cache   = cdi.GetDefaultCache()
		matches = map[string]struct{}{}
		devices = []string{}
	)

	for _, device := range cache.ListDevices() {
		for _, glob := range patterns {
			match, err := filepath.Match(glob, device)
			if err != nil {
				return fmt.Errorf("failed to match pattern %q against %q: %w",
					glob, device, err)
			}
			if match {
				matches[device] = struct{}{}
			}
		}
	}
	for device := range matches {
		devices = append(devices, device)
	}
	sort.Strings(devices)

	unresolved, err := cache.InjectDevices(ociSpec, devices...)

	if len(unresolved) > 0 {
		fmt.Printf("Unresolved CDI devices:\n")
		for idx, device := range unresolved {
			fmt.Printf("  %d. %s\n", idx, device)
		}
	}
	if err != nil {
		return fmt.Errorf("OCI device injection failed: %w", err)
	}

	fmt.Printf("Updated OCI Spec:\n")
	fmt.Printf("%s", marshalObject(2, ociSpec, format))

	return nil
}

func cdiResolveDevices(ociSpecFiles ...string) error {
	var (
		cache      *cdi.Cache
		ociSpec    *oci.Spec
		devices    []string
		unresolved []string
		err        error
	)

	cache, _ = cdi.NewCache()

	for _, ociSpecFile := range ociSpecFiles {
		ociSpec, err = readOCISpec(ociSpecFile)
		if err != nil {
			return err
		}

		devices = collectCDIDevicesFromOCISpec(ociSpec)

		unresolved, err = cache.InjectDevices(ociSpec, devices...)
		if len(unresolved) > 0 {
			fmt.Printf("Unresolved CDI devices:\n")
			for idx, device := range unresolved {
				fmt.Printf("  %d. %s\n", idx, device)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to resolve devices for OCI Spec %q: %w", ociSpecFile, err)
		}

		format := chooseFormat(injectCfg.output, ociSpecFile)
		fmt.Printf("%s", marshalObject(2, ociSpec, format))
	}

	return nil
}

func collectCDIDevicesFromOCISpec(spec *oci.Spec) []string {
	var (
		cdiDevs []string
	)

	if spec.Linux == nil || len(spec.Linux.Devices) == 0 {
		return nil
	}

	devices := spec.Linux.Devices
	g := gen.NewFromSpec(spec)
	g.ClearLinuxDevices()

	for _, d := range devices {
		if !cdi.IsQualifiedName(d.Path) {
			g.AddDevice(d)
			continue
		}
		cdiDevs = append(cdiDevs, d.Path)
	}

	return cdiDevs
}

func cdiListSpecs(verbose bool, format string, vendors ...string) {
	var (
		cache = cdi.GetDefaultCache()
	)

	format = chooseFormat(format, "format-as.yaml")

	if len(vendors) == 0 {
		vendors = cache.ListVendors()
	}

	if len(vendors) == 0 {
		fmt.Printf("No CDI Specs found.\n")
		cdiErrors := cache.GetErrors()
		if len(cdiErrors) > 0 {
			for path, specErrors := range cdiErrors {
				fmt.Printf("%s has errors:\n", path)
				for idx, err := range specErrors {
					fmt.Printf("  %d. %v\n", idx, err)
				}
			}
		}
		return
	}

	fmt.Printf("CDI Specs found:\n")
	for _, vendor := range cache.ListVendors() {
		fmt.Printf("Vendor %s:\n", vendor)
		for _, spec := range cache.GetVendorSpecs(vendor) {
			cdiPrintSpec(spec, verbose, format, 2)
			cdiPrintSpecErrors(spec, verbose, 2)
		}
	}
}

func cdiPrintSpec(spec *cdi.Spec, verbose bool, format string, level int) {
	fmt.Printf("%sSpec File %s\n", indent(level), spec.GetPath())

	if verbose {
		fmt.Printf("%s", marshalObject(level+2, spec.Spec, format))
	}
}

func cdiPrintSpecErrors(spec *cdi.Spec, verbose bool, level int) {
	var (
		cache     = cdi.GetDefaultCache()
		cdiErrors = cache.GetErrors()
	)

	if len(cdiErrors) > 0 {
		for path, specErrors := range cdiErrors {
			if len(specErrors) == 0 {
				continue
			}
			fmt.Printf("%s%s has %d errors:\n", indent(level), path, len(specErrors))
			for idx, err := range specErrors {
				fmt.Printf("%s%d. %v\n", indent(level+2), idx, err)
			}
		}
	}
}

func cdiPrintCache(args ...string) {
	if len(args) == 0 {
		args = []string{"all"}
	}

	for _, what := range args {
		switch what {
		case "vendors", "vendor":
			cdiListVendors()
		case "classes", "class":
			cdiListClasses()
		case "specs", "spec":
			cdiListSpecs(monitorCfg.verbose, monitorCfg.output)
		case "devices", "device":
			cdiListDevices(monitorCfg.verbose, monitorCfg.output)
		case "all":
			cdiListVendors()
			cdiListClasses()
			cdiListSpecs(monitorCfg.verbose, monitorCfg.output)
			cdiListDevices(monitorCfg.verbose, monitorCfg.output)
		default:
			fmt.Printf("Unrecognized CDI aspect/object %q... ignoring it\n", what)
		}
	}
}

func cdiPrintCacheErrors() {
	var (
		cache     = cdi.GetDefaultCache()
		cdiErrors = cache.GetErrors()
	)

	if len(cdiErrors) == 0 {
		return
	}

	fmt.Printf("CDI Cache has errors:\n")
	for path, specErrors := range cdiErrors {
		fmt.Printf("Spec file %s:\n", path)
		for idx, err := range specErrors {
			fmt.Printf("  %d: %v\n", idx, strings.TrimRight(err.Error(), "\n"))
		}
	}
}
