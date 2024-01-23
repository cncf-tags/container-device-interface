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
	"path/filepath"
	"testing"
	"time"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestDefaultCacheConfigure(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name    string
		entries specDirs
		devices []string
	}
	for _, tc := range []*testCase{
		{
			name: "one CDI Spec",
			entries: specDirs{
				run: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
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
			name: "two CDI Specs",
			entries: specDirs{
				run: map[string]string{
					"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
					"vendor1-other.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
				},
			},
			devices: []string{
				"vendor1.com/device=dev1",
				"vendor1.com/device=dev2",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir   string
				err   error
				opts  []Option
				cache = GetDefaultCache()
			)

			dir, err = createSpecDirs(t, tc.entries.etc, tc.entries.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}
			opts = []Option{
				WithAutoRefresh(false),
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			}
			Configure(opts...)

			devices := cache.ListDevices()
			if len(tc.devices) == 0 {
				require.True(t, len(devices) == 0)
			} else {
				require.Equal(t, tc.devices, devices)
			}
		})
	}
}

func TestDefaultCacheRefresh(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name    string
		updates []specDirs
		errors  []map[string]struct{}
		devices [][]string
		devprio []map[string]int
	}
	for _, tc := range []*testCase{
		{
			name: "empty cache, add one Spec",
			updates: []specDirs{
				{},
				{
					run: map[string]string{
						"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
					},
				},
			},
			devices: [][]string{
				nil,
				{
					"vendor1.com/device=dev1",
				},
			},
			devprio: []map[string]int{
				{},
				{
					"vendor1.com/device=dev1": 1,
				},
			},
			errors: []map[string]struct{}{
				{},
				{},
			},
		},
		{
			name: "two Specs, remove one",
			updates: []specDirs{
				{
					run: map[string]string{
						"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
						"vendor1-other.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
					},
				},
				{
					run: map[string]string{
						"vendor1.yaml": "remove",
					},
				},
			},
			devices: [][]string{
				{
					"vendor1.com/device=dev1",
					"vendor1.com/device=dev2",
				},
				{
					"vendor1.com/device=dev2",
				},
			},
			devprio: []map[string]int{
				{
					"vendor1.com/device=dev1": 1,
					"vendor1.com/device=dev2": 1,
				},
				{
					"vendor1.com/device=dev2": 1,
				},
			},
			errors: []map[string]struct{}{
				{},
				{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir   string
				err   error
				opts  []Option
				cache = GetDefaultCache()
			)
			for _, selfRefresh := range []bool{false, true} {
				for idx, update := range tc.updates {
					if idx == 0 {
						dir, err = createSpecDirs(t, update.etc, update.run)
						if err != nil {
							t.Errorf("failed to create test directory: %v", err)
							return
						}
						opts = []Option{
							WithAutoRefresh(selfRefresh),
							WithSpecDirs(
								filepath.Join(dir, "etc"),
								filepath.Join(dir, "run"),
							),
						}
						Configure(opts...)
					} else {
						err = updateSpecDirs(t, dir, update.etc, update.run)
						if err != nil {
							t.Errorf("failed to update test directory: %v", err)
							return
						}
					}

					if !selfRefresh {
						if idx > 0 { // Configure implies Refresh(), so omit it here
							err = Refresh()
							if len(tc.errors[idx]) == 0 {
								require.Nil(t, err, fmt.Sprintf("unexpected errors: %v", err))
							} else {
								require.NotNil(t, err)
							}
						}
					} else {
						time.Sleep(10 * time.Millisecond)
					}

					devices := cache.ListDevices()
					if len(tc.devices[idx]) == 0 {
						require.True(t, len(devices) == 0)
					} else {
						require.Equal(t, tc.devices[idx], devices)
					}

					for name, prio := range tc.devprio[idx] {
						dev := cache.GetDevice(name)
						require.NotNil(t, dev)
						require.Equal(t, dev.GetSpec().GetPriority(), prio)
					}

					for _, v := range cache.ListVendors() {
						for _, spec := range cache.GetVendorSpecs(v) {
							err := cache.GetSpecErrors(spec)
							relSpecPath, _ := filepath.Rel(dir, spec.GetPath())
							_, ok := tc.errors[idx][relSpecPath]
							require.True(t, (err == nil && !ok) || (err != nil && ok))
						}
					}
				}
			}
		})
	}
}

func TestDefaultCacheInjectDevice(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name        string
		cdiSpecs    specDirs
		ociSpec     *oci.Spec
		devices     []string
		result      *oci.Spec
		unresolved  []string
		expectedErr error
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir string
				err error
			)
			dir, err = createSpecDirs(t, tc.cdiSpecs.etc, tc.cdiSpecs.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}

			Configure(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)

			unresolved, err := InjectDevices(tc.ociSpec, tc.devices...)
			if len(tc.unresolved) != 0 {
				require.NotNil(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Equal(t, tc.unresolved, unresolved)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.result, tc.ociSpec)
		})
	}
}
