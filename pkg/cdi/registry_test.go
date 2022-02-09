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
	"path/filepath"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestRegistryResolver(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name       string
		cdiSpecs   specDirs
		ociSpec    *oci.Spec
		devices    []string
		result     *oci.Spec
		unresolved []string
	}
	for _, tc := range []*testCase{
		{
			name: "empty OCI Spec, inject one device",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
  env:
  - VENDOR1_SPEC_VAR1=VAL1
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "VENDOR1_VAR1=VAL1"
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
				},
			},
			ociSpec: &oci.Spec{},
			devices: []string{
				"vendor1.com/device=dev1",
			},
			result: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"VENDOR1_SPEC_VAR1=VAL1",
						"VENDOR1_VAR1=VAL1",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/vendor1-dev1",
							Type:  "b",
							Major: 10,
							Minor: 1,
						},
					},
					Resources: &oci.LinuxResources{
						Devices: []oci.LinuxDeviceCgroup{
							{
								Allow:  true,
								Type:   "b",
								Major:  int64ptr(10),
								Minor:  int64ptr(1),
								Access: "rwm",
							},
						},
					},
				},
			},
		},
		{
			name: "non-empty OCI Spec, inject one device",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
  env:
  - VENDOR1_SPEC_VAR1=VAL1
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "VENDOR1_VAR1=VAL1"
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
				},
			},
			ociSpec: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"ORIG_VAR1=VAL1",
						"ORIG_VAR2=VAL2",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path: "/dev/null",
						},
						{
							Path: "/dev/zero",
						},
					},
				},
			},
			devices: []string{
				"vendor1.com/device=dev1",
			},
			result: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"ORIG_VAR1=VAL1",
						"ORIG_VAR2=VAL2",
						"VENDOR1_SPEC_VAR1=VAL1",
						"VENDOR1_VAR1=VAL1",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path: "/dev/null",
						},
						{
							Path: "/dev/zero",
						},
						{
							Path:  "/dev/vendor1-dev1",
							Type:  "b",
							Major: 10,
							Minor: 1,
						},
					},
					Resources: &oci.LinuxResources{
						Devices: []oci.LinuxDeviceCgroup{
							{
								Allow:  true,
								Type:   "b",
								Major:  int64ptr(10),
								Minor:  int64ptr(1),
								Access: "rwm",
							},
						},
					},
				},
			},
		},
		{
			name: "non-empty OCI Spec, inject several devices, hooks",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
  env:
  - VENDOR1_SPEC_VAR1=VAL1
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "VENDOR1_DEV1=VAL1"
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
      hooks:
      - hookName: prestart
        path: "/usr/local/bin/prestart-vendor-hook"
        args:
        - "--verbose"
        env:
        - "HOOK_ENV1=PRESTART_VAL1"
      - hookName: createRuntime
        path: "/usr/local/bin/cr-vendor-hook"
        args:
        - "--debug"
        env:
        - "HOOK_ENV1=CREATE_RUNTIME_VAL1"
  - name: "dev3"
    containerEdits:
      env:
      - "VENDOR1_DEV3=VAL3"
      deviceNodes:
      - path: "/dev/vendor1-dev3"
        type: b
        major: 10
        minor: 3
`,
				},
			},
			ociSpec: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"ORIG_VAR1=VAL1",
						"ORIG_VAR2=VAL2",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path: "/dev/null",
						},
						{
							Path: "/dev/zero",
						},
					},
				},
			},
			devices: []string{
				"vendor1.com/device=dev1",
				"vendor1.com/device=dev2",
				"vendor1.com/device=dev3",
			},
			result: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"ORIG_VAR1=VAL1",
						"ORIG_VAR2=VAL2",
						"VENDOR1_SPEC_VAR1=VAL1",
						"VENDOR1_DEV1=VAL1",
						"VENDOR1_DEV2=VAL2",
						"VENDOR1_DEV3=VAL3",
					},
				},
				Hooks: &oci.Hooks{
					Prestart: []oci.Hook{
						{
							Path: "/usr/local/bin/prestart-vendor-hook",
							Args: []string{"--verbose"},
							Env:  []string{"HOOK_ENV1=PRESTART_VAL1"},
						},
					},
					CreateRuntime: []oci.Hook{
						{
							Path: "/usr/local/bin/cr-vendor-hook",
							Args: []string{"--debug"},
							Env:  []string{"HOOK_ENV1=CREATE_RUNTIME_VAL1"},
						},
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path: "/dev/null",
						},
						{
							Path: "/dev/zero",
						},
						{
							Path:  "/dev/vendor1-dev1",
							Type:  "b",
							Major: 10,
							Minor: 1,
						},
						{
							Path:  "/dev/vendor1-dev2",
							Type:  "b",
							Major: 10,
							Minor: 2,
						},
						{
							Path:  "/dev/vendor1-dev3",
							Type:  "b",
							Major: 10,
							Minor: 3,
						},
					},
					Resources: &oci.LinuxResources{
						Devices: []oci.LinuxDeviceCgroup{
							{
								Allow:  true,
								Type:   "b",
								Major:  int64ptr(10),
								Minor:  int64ptr(1),
								Access: "rwm",
							},
							{
								Allow:  true,
								Type:   "b",
								Major:  int64ptr(10),
								Minor:  int64ptr(2),
								Access: "rwm",
							},
							{
								Allow:  true,
								Type:   "b",
								Major:  int64ptr(10),
								Minor:  int64ptr(3),
								Access: "rwm",
							},
						},
					},
				},
			},
		},
		{
			name: "empty OCI Spec, non-existent device",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
  env:
  - VENDOR1_SPEC_VAR1=VAL1
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "VENDOR1_VAR1=VAL1"
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
				},
			},
			ociSpec: &oci.Spec{},
			devices: []string{
				"vendor1.com/device=dev2",
			},
			result: &oci.Spec{},
			unresolved: []string{
				"vendor1.com/device=dev2",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir string
				err error
				reg Registry
			)
			dir, err = createSpecDirs(t, tc.cdiSpecs.etc, tc.cdiSpecs.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}
			reg = GetRegistry(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)
			require.Nil(t, err)
			require.NotNil(t, reg)

			unresolved, err := reg.InjectDevices(tc.ociSpec, tc.devices...)
			if len(tc.unresolved) != 0 {
				require.NotNil(t, err)
				require.Equal(t, tc.unresolved, unresolved)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.result, tc.ociSpec)
		})
	}
}

func TestRegistrySpecDB(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name     string
		cdiSpecs specDirs
		vendors  []string
		classes  []string
	}
	for _, tc := range []*testCase{
		{
			name: "no vendors, no classes",
		},
		{
			name: "one vendor, one class",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
  env:
  - VENDOR1_SPEC_VAR1=VAL1
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "VENDOR1_VAR1=VAL1"
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
				},
			},
			vendors: []string{
				"vendor1.com",
			},
			classes: []string{
				"device",
			},
		},
		{
			name: "one vendor, multiple classes",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
					"vendor1-other.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/other-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-other-dev1"
        type: b
        major: 12
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-other-dev2"
        type: b
        major: 11
        minor: 2
`,
				},
			},
			vendors: []string{
				"vendor1.com",
			},
			classes: []string{
				"device",
				"other-device",
			},
		},
		{
			name: "multiple vendor, multiple classes",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
					"vendor2.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor2.com/other-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev2"
`,
					"vendor2-other.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor2.com/another-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-another-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-another-dev2"
`,
					"vendor3.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor3.com/yet-another-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor3-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor3-dev2"
`,
				},
			},
			vendors: []string{
				"vendor1.com",
				"vendor2.com",
				"vendor3.com",
			},
			classes: []string{
				"another-device",
				"device",
				"other-device",
				"yet-another-device",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir string
				err error
				reg Registry
			)
			dir, err = createSpecDirs(t, tc.cdiSpecs.etc, tc.cdiSpecs.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}
			reg = GetRegistry(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)
			require.Nil(t, err)
			require.NotNil(t, reg)

			vendors := reg.SpecDB().ListVendors()
			require.Equal(t, tc.vendors, vendors)
			classes := reg.SpecDB().ListClasses()
			require.Equal(t, tc.classes, classes)
		})
	}
}

func TestRegistryDeviceDB(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name     string
		cdiSpecs specDirs
		devices  []string
	}
	for _, tc := range []*testCase{
		{
			name: "no vendors, no classes",
		},
		{
			name: "one vendor, one class",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
  env:
  - VENDOR1_SPEC_VAR1=VAL1
devices:
  - name: "dev1"
    containerEdits:
      env:
      - "VENDOR1_VAR1=VAL1"
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
				},
			},
			devices: []string{
				"vendor1.com/device=dev1",
			},
		},
		{
			name: "one vendor, multiple classes",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
					"vendor1-other.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/other-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-other-dev1"
        type: b
        major: 11
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-other-dev2"
        type: b
        major: 11
        minor: 2
`,
				},
			},
			devices: []string{
				"vendor1.com/device=dev1",
				"vendor1.com/device=dev2",
				"vendor1.com/other-device=dev1",
				"vendor1.com/other-device=dev2",
			},
		},
		{
			name: "multiple vendor, multiple classes",
			cdiSpecs: specDirs{
				etc: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
  - name: "dev2"
    containerEdits:
      env:
      - "VENDOR1_DEV2=VAL2"
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
					"vendor2.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor2.com/other-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev2"
`,
					"vendor2-other.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor2.com/another-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-another-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-another-dev2"
`,
					"vendor3.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor3.com/yet-another-device"
containerEdits:
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor3-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor3-dev2"
`,
				},
			},
			devices: []string{
				"vendor1.com/device=dev1",
				"vendor1.com/device=dev2",
				"vendor2.com/another-device=dev1",
				"vendor2.com/another-device=dev2",
				"vendor2.com/other-device=dev1",
				"vendor2.com/other-device=dev2",
				"vendor3.com/yet-another-device=dev1",
				"vendor3.com/yet-another-device=dev2",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir string
				err error
				reg Registry
			)
			dir, err = createSpecDirs(t, tc.cdiSpecs.etc, tc.cdiSpecs.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}
			reg = GetRegistry(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)
			require.Nil(t, err)
			require.NotNil(t, reg)

			devices := reg.DeviceDB().ListDevices()
			require.Equal(t, tc.devices, devices)
		})
	}
}
