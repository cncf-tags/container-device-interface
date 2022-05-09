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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi/validate"
	cdi "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	type testCase struct {
		name      string
		etc       map[string]string
		run       map[string]string
		sources   map[string]string
		errors    map[string]struct{}
		dirErrors map[string]struct{}
	}
	for _, tc := range []*testCase{
		{
			name: "no spec dirs",
			dirErrors: map[string]struct{}{
				"etc": {},
				"run": {},
			},
		},
		{
			name: "no spec files",
			etc:  map[string]string{},
			run:  map[string]string{},
		},
		{
			name: "one spec file",
			etc: map[string]string{
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
			sources: map[string]string{
				"vendor1.com/device=dev1": "etc/vendor1.yaml",
			},
		},
		{
			name: "multiple spec files with override",
			etc: map[string]string{
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
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
			},
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
			sources: map[string]string{
				"vendor1.com/device=dev1": "run/vendor1.yaml",
				"vendor1.com/device=dev2": "etc/vendor1.yaml",
			},
		},
		{
			name: "multiple spec files, with conflicts",
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
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
				"vendor1-other.yaml": `
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
			sources: map[string]string{
				"vendor1.com/device=dev2": "run/vendor1.yaml",
			},
			errors: map[string]struct{}{
				"run/vendor1.yaml":       {},
				"run/vendor1-other.yaml": {},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir     string
				specDir string
				err     error
				cache   *Cache
			)
			if tc.etc != nil || tc.run != nil {
				dir, err = createSpecDirs(t, tc.etc, tc.run)
				if err != nil {
					t.Errorf("failed to create test directory: %v", err)
					return
				}
			} else {
				dir, err = mkTestDir(t, nil)
			}

			cache, err = NewCache(WithSpecDirs(
				filepath.Join(dir, "etc"),
				filepath.Join(dir, "run")),
			)

			if len(tc.dirErrors) != 0 {
				for specDir = range tc.dirErrors {
					specDir = filepath.Join(dir, specDir)
					require.NotNil(t, cache.GetSpecDirErrors()[specDir])
					return
				}
			}

			if len(tc.errors) == 0 {
				require.Nil(t, err)
			}
			require.NotNil(t, cache)

			for name, dev := range cache.devices {
				require.Equal(t, filepath.Join(dir, tc.sources[name]),
					dev.GetSpec().GetPath())
			}
			for name, path := range tc.sources {
				dev := cache.devices[name]
				require.NotNil(t, dev)
				require.Equal(t, filepath.Join(dir, path),
					dev.GetSpec().GetPath())
			}

			for path := range tc.errors {
				fullPath := filepath.Join(dir, path)
				_, ok := cache.errors[fullPath]
				require.True(t, ok)
			}
			for fullPath := range cache.errors {
				path, err := filepath.Rel(dir, fullPath)
				require.Nil(t, err)
				_, ok := tc.errors[path]
				require.True(t, ok)
			}
		})
	}
}

func TestRefreshCache(t *testing.T) {
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
			name: "one Spec, add another, no shadowing, no conflicts",
			updates: []specDirs{
				{
					etc: map[string]string{
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
				{
					run: map[string]string{
						"vendor1.yaml": `
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
			},
			devices: [][]string{
				{
					"vendor1.com/device=dev1",
				},
				{
					"vendor1.com/device=dev1",
					"vendor1.com/device=dev2",
				},
			},
			devprio: []map[string]int{
				{
					"vendor1.com/device=dev1": 0,
				},
				{
					"vendor1.com/device=dev1": 0,
					"vendor1.com/device=dev2": 1,
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
		{
			name: "one Spec, add another, shadowing",
			updates: []specDirs{
				{
					etc: map[string]string{
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
				{
					"vendor1.com/device=dev1",
				},
				{
					"vendor1.com/device=dev1",
				},
			},
			devprio: []map[string]int{
				{
					"vendor1.com/device=dev1": 0,
				},
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
			name: "one Spec, add another, conflicts",
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
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 1
`,
					},
				},
				{
					run: map[string]string{
						"vendor1-other.yaml": `
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
  - name: "dev3"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev3"
        type: b
        major: 10
        minor: 3
`,
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
					"vendor1.com/device=dev3",
				},
			},
			devprio: []map[string]int{
				{
					"vendor1.com/device=dev1": 1,
					"vendor1.com/device=dev2": 1,
				},
				{
					"vendor1.com/device=dev2": 1,
					"vendor1.com/device=dev3": 1,
				},
			},
			errors: []map[string]struct{}{
				{},
				{
					"run/vendor1.yaml":       {},
					"run/vendor1-other.yaml": {},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir   string
				err   error
				opts  []Option
				cache *Cache
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
							WithSpecDirs(
								filepath.Join(dir, "etc"),
								filepath.Join(dir, "run"),
							),
						}
						if !selfRefresh {
							opts = append(opts, WithAutoRefresh(false))
						}
						cache, err = NewCache(opts...)
						require.NotNil(t, cache)
					} else {
						err = updateSpecDirs(t, dir, update.etc, update.run)
						if err != nil {
							t.Errorf("failed to update test directory: %v", err)
							return
						}
						if selfRefresh {
							time.Sleep(100 * time.Millisecond)
						} else {
							err = cache.Refresh()

							if len(tc.errors[idx]) == 0 {
								require.Nil(t, err)
							} else {
								require.NotNil(t, err)
							}
						}
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

func TestFuzzSelfRefreshCache(t *testing.T) {
	type specDirs struct {
		etc map[string]string
		run map[string]string
	}
	type testCase struct {
		name    string
		updates []specDirs
	}

	for _, tc := range []*testCase{
		{
			name: "one device in /etc, update it, then shadow it in /run",
			updates: []specDirs{
				{
					etc: map[string]string{
						"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/original-vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
					},
				},
				{
					etc: map[string]string{
						"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/updated-vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
					},
				},
				{
					run: map[string]string{
						"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/shadowed-vendor1-dev1"
        type: b
        major: 10
        minor: 1
`,
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir    string
				cache  *Cache
				err    error
				errCh  chan error
				stopCh chan struct{}

				duration = 10 * time.Second
				wg       = &sync.WaitGroup{}
			)

			stopCh = make(chan struct{})
			errCh = make(chan error, 2)

			// injector: run injection loop until an error or request to stop
			injector := func() {
				var (
					inject = "vendor1.com/device=dev1"
					expect = map[string]struct{}{
						"/dev/original-vendor1-dev1": {},
						"/dev/updated-vendor1-dev1":  {},
						"/dev/shadowed-vendor1-dev1": {},
					}
					unresolved []string
					err        error
				)

				defer func() {
					errCh <- err
					wg.Done()
				}()

				for {
					ociSpec := &oci.Spec{}
					unresolved, err = cache.InjectDevices(ociSpec, inject)
					if err != nil {
						err = errors.Wrap(err, "device injection failed")
						return
					}
					if unresolved != nil {
						err = errors.Errorf("unresolved devices %s", strings.Join(unresolved, ","))
						return
					}

					result := ociSpec.Linux.Devices[0].Path
					if _, ok := expect[result]; !ok {
						err = errors.Errorf("unexpected device path %s", result)
						return
					}

					select {
					case _ = <-stopCh:
						return
					default:
					}
				}
			}

			// Run Spec update loop until an error or request to stop.
			updater := func() {
				var (
					idx = 1
					err error
				)

				defer func() {
					errCh <- err
					wg.Done()
				}()

				for {
					if idx >= len(tc.updates) {
						idx = 0
					}
					update := tc.updates[idx]
					err = updateSpecDirs(t, dir, update.etc, update.run)
					if err != nil {
						return
					}

					select {
					case _ = <-stopCh:
						return
					default:
					}

					idx++
				}
			}

			fssyncer := func() {
				var sync = time.NewTimer(2 * time.Second)

				defer func() {
					if !sync.Stop() {
						<-sync.C
					}
					wg.Done()
				}()

				// run sync loop until request to stop (trying to amortize the fs-hit
				// from updater()'s create+write+rename loop)
				for {
					select {
					case _ = <-stopCh:
						go syscall.Sync()
						return
					case _ = <-sync.C:
						go syscall.Sync()
						sync.Reset(2 * time.Second)
					}
				}
			}

			dir, err = createSpecDirs(t, tc.updates[0].etc, tc.updates[0].run)
			require.NoError(t, err)

			cache, err = NewCache(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)
			require.NoError(t, err)
			require.NotNil(t, cache)

			go injector()
			go updater()
			go fssyncer()
			wg.Add(3)

			done := time.After(duration)
			for {
				select {
				case err = <-errCh:
					require.NotNil(t, err)
				case _ = <-done:
					close(stopCh)
					wg.Wait()
					return
				}
			}
		})
	}
}

func TestInjectDevice(t *testing.T) {
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
				dir   string
				err   error
				cache *Cache
			)
			dir, err = createSpecDirs(t, tc.cdiSpecs.etc, tc.cdiSpecs.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}
			cache, err = NewCache(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)
			require.Nil(t, err)
			require.NotNil(t, cache)

			unresolved, err := cache.InjectDevices(tc.ociSpec, tc.devices...)
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

func TestListVendorsAndClasses(t *testing.T) {
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
        type: b
        major: 12
        minor: 1
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev2"
        type: b
        major: 12
        minor: 2
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
        type: b
        major: 13
        minor: 1
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-another-dev2"
        type: b
        major: 13
        minor: 2
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
        type: b
        major: 11
        minor: 1

  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor3-dev2"
        type: b
        major: 14
        minor: 2
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
				dir   string
				err   error
				cache *Cache
			)
			dir, err = createSpecDirs(t, tc.cdiSpecs.etc, tc.cdiSpecs.run)
			if err != nil {
				t.Errorf("failed to create test directory: %v", err)
				return
			}
			cache, err = NewCache(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
					filepath.Join(dir, "run"),
				),
			)
			require.Nil(t, err)
			require.NotNil(t, cache)

			vendors := cache.ListVendors()
			require.Equal(t, tc.vendors, vendors)
			classes := cache.ListClasses()
			require.Equal(t, tc.classes, classes)
		})
	}
}

func TestCacheWriteSpec(t *testing.T) {
	type testCase struct {
		name    string
		etc     map[string]string
		invalid map[string]bool
	}
	for _, tc := range []*testCase{
		{
			name: "one spec file",
			etc: map[string]string{
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
		{
			name: "multiple spec files",
			etc: map[string]string{
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
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
				"vendor2.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor2.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev1"
        type: b
        major: 10
        minor: 1
`,
				"vendor3.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor3.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor3-dev1"
        type: b
        major: 10
        minor: 1
`,
			},
		},

		{
			name: "multiple spec files/data, some invalid",
			etc: map[string]string{
				"vendor1.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor1.comdevice"
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
      deviceNodes:
      - path: "/dev/vendor1-dev2"
        type: b
        major: 10
        minor: 2
`,
				"vendor2.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor2.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor2-dev1"
        type: b
        major: 10
        minor: 1
`,
				"vendor3.yaml": `
cdiVersion: "0.3.0"
kind:       "vendor3.com/device"
containerEdits:
  deviceNodes:
  - path: "/dev/vendor3-dev1"
    type: b
    major: 10
    minor: 1
`,
			},
			invalid: map[string]bool{
				"vendor1.yaml": true,
				"vendor3.yaml": true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir   string
				etc   map[string]string
				raw   *cdi.Spec
				err   error
				cache *Cache
				other *Cache
			)

			SetSpecValidator(validate.WithNamedSchema("builtin"))

			if len(tc.invalid) != 0 {
				dir, err = createSpecDirs(t, nil, nil)
				require.NoError(t, err)
				cache, err = NewCache(
					WithSpecDirs(
						filepath.Join(dir, "etc"),
						filepath.Join(dir, "run"),
					),
					WithAutoRefresh(false),
				)

				require.NoError(t, err)
				require.NotNil(t, cache)

				etc = map[string]string{}
				for name, data := range tc.etc {
					raw, err = ParseSpec([]byte(data))
					require.NoError(t, err)
					require.NotNil(t, raw)

					err = cache.WriteSpec(raw, name)

					if tc.invalid[name] {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
						etc[name] = data
					}
				}
			} else {
				etc = tc.etc
			}

			dir, err = createSpecDirs(t, etc, nil)
			require.NoError(t, err)

			cache, err = NewCache(
				WithSpecDirs(
					filepath.Join(dir, "etc"),
				),
			)
			require.NoError(t, err)
			require.NotNil(t, cache)

			other, err = NewCache(
				WithSpecDirs(
					filepath.Join(dir, "run"),
				),
				WithAutoRefresh(false),
			)
			require.NoError(t, err)
			require.NotNil(t, other)

			cSpecs := map[string]*cdi.Spec{}
			for _, vendor := range cache.ListVendors() {
				for _, spec := range cache.GetVendorSpecs(vendor) {
					name := filepath.Base(spec.GetPath())
					cSpecs[name] = spec.Spec
					err = other.WriteSpec(spec.Spec, name)
					require.NoError(t, err)
				}
			}

			err = other.Refresh()
			require.NoError(t, err)

			for _, vendor := range other.ListVendors() {
				for _, spec := range other.GetVendorSpecs(vendor) {
					name := filepath.Base(spec.GetPath())
					require.Equal(t, spec.Spec, cSpecs[name])
				}
			}
		})
	}
}

// Create and populate automatically cleaned up spec directories.
func createSpecDirs(t *testing.T, etc, run map[string]string) (string, error) {
	return mkTestDir(t, map[string]map[string]string{
		"etc": etc,
		"run": run,
	})
}

// Update spec directories with new data.
func updateSpecDirs(t *testing.T, dir string, etc, run map[string]string) error {
	updates := map[string]map[string]string{
		"etc": {},
		"run": {},
	}
	for sub, entries := range map[string]map[string]string{
		"etc": etc,
		"run": run,
	} {
		path := filepath.Join(dir, sub)
		for name, data := range entries {
			if data == "remove" {
				os.Remove(filepath.Join(path, name))
			} else {
				updates[sub][name] = data
			}
		}
	}
	return updateTestDir(t, dir, updates)
}

func int64ptr(v int64) *int64 {
	return &v
}
