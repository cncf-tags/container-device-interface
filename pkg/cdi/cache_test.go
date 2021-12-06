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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	type testCase struct {
		name    string
		etc     map[string]string
		run     map[string]string
		sources map[string]string
		errors  map[string]struct{}
	}
	for _, tc := range []*testCase{
		{
			name: "no spec dirs",
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
cdiVersion: "0.2.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
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
cdiVersion: "0.2.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
`,
			},
			run: map[string]string{
				"vendor1.yaml": `
cdiVersion: "0.2.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
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
cdiVersion: "0.2.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
  - name: "dev2"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev2"
`,
				"vendor1-other.yaml": `
cdiVersion: "0.2.0"
kind:       "vendor1.com/device"
devices:
  - name: "dev1"
    containerEdits:
      deviceNodes:
      - path: "/dev/vendor1-dev1"
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
				dir   string
				err   error
				cache *Cache
			)
			if tc.etc != nil || tc.run != nil {
				dir, err = createSpecDirs(t, tc.etc, tc.run)
				if err != nil {
					t.Errorf("failed to create test directory: %v", err)
					return
				}
			}
			cache, err = NewCache(WithSpecDirs(
				filepath.Join(dir, "etc"),
				filepath.Join(dir, "run")),
			)

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
