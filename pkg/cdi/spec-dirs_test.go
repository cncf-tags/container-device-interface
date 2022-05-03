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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestScanSpecDirs(t *testing.T) {
	type testCase struct {
		name    string
		files   map[string]string
		success map[string]struct{}
		failure map[string]struct{}
		vendors map[string]string
		classes map[string]string
		abort   bool
	}
	for _, tc := range []*testCase{
		{
			name:  "no directory",
			files: nil,
		},
		{
			name:    "no files",
			files:   map[string]string{},
			success: map[string]struct{}{},
			failure: map[string]struct{}{},
			vendors: map[string]string{},
			classes: map[string]string{},
		},
		{
			name: "one valid file",
			files: map[string]string{
				"valid.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			},
			success: map[string]struct{}{
				"valid.yaml": {},
			},
			failure: map[string]struct{}{},
			vendors: map[string]string{
				"valid.yaml": "vendor.com",
			},
			classes: map[string]string{
				"valid.yaml": "device",
			},
		},
		{
			name: "one invalid file",
			files: map[string]string{
				"invalid.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
`,
			},
			success: map[string]struct{}{},
			failure: map[string]struct{}{
				"invalid.yaml": {},
			},
			vendors: map[string]string{},
			classes: map[string]string{},
		},
		{
			name: "two valid files, one invalid file",
			files: map[string]string{
				"valid1.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device1
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
				"valid2.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device2
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
				"invalid.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
`,
			},
			success: map[string]struct{}{
				"valid1.yaml": {},
				"valid2.yaml": {},
			},
			failure: map[string]struct{}{
				"invalid.yaml": {},
			},
			vendors: map[string]string{
				"valid1.yaml": "vendor.com",
				"valid2.yaml": "vendor.com",
			},
			classes: map[string]string{
				"valid1.yaml": "device1",
				"valid2.yaml": "device2",
			},
		},
		{
			// we assume running on an fs with sorted readdir()...
			name: "one valid file, one invalid file, abort on first error",
			files: map[string]string{
				"valid.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
				"invalid.yaml": `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
`,
			},
			success: map[string]struct{}{},
			failure: map[string]struct{}{
				"invalid.yaml": {},
			},
			vendors: map[string]string{},
			classes: map[string]string{},
			abort:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				dir     string
				err     error
				success map[string]struct{}
				failure map[string]struct{}
				vendors map[string]string
				classes map[string]string
			)
			if tc.files != nil {
				dir, err = mkTestDir(t, map[string]map[string]string{
					"etc": tc.files,
				})
				if err != nil {
					t.Errorf("failed to populate test directory: %v", err)
				}
				dir = filepath.Join(dir, "etc")
				success = map[string]struct{}{}
				failure = map[string]struct{}{}
				vendors = map[string]string{}
				classes = map[string]string{}
			}

			dirs := []string{dir}
			err = scanSpecDirs(dirs, func(path string, prio int, spec *Spec, err error) error {
				name := filepath.Base(path)
				if err != nil {
					failure[name] = struct{}{}
					if tc.abort {
						return err
					}
				} else {
					success[name] = struct{}{}
					vendors[name] = spec.GetVendor()
					classes[name] = spec.GetClass()
				}
				return nil
			})

			if tc.files == nil {
				require.Error(t, err)
				require.True(t, os.IsNotExist(err))
				return
			}

			require.Equal(t, tc.success, success)
			require.Equal(t, tc.failure, failure)
			require.Equal(t, tc.vendors, vendors)
			require.Equal(t, tc.classes, classes)
		})
	}
}

// Create an automatically cleaned up temporary directory, with optional content.
func mkTestDir(t *testing.T, dirs map[string]map[string]string) (string, error) {
	tmp, err := ioutil.TempDir("", ".cache-test*")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create test directory")
	}

	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	if err = updateTestDir(t, tmp, dirs); err != nil {
		return "", err
	}

	return tmp, nil
}

func updateTestDir(t *testing.T, tmp string, dirs map[string]map[string]string) error {
	for sub, content := range dirs {
		dir := filepath.Join(tmp, sub)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create directory %q", dir)
		}
		for file, data := range content {
			file := filepath.Join(dir, file)
			tmp := file + ".tmp"
			if err := ioutil.WriteFile(tmp, []byte(data), 0644); err != nil {
				return errors.Wrapf(err, "failed to write file %q", tmp)
			}
			if err := os.Rename(tmp, file); err != nil {
				return errors.Wrapf(err, "failed to rename %q to %q", tmp, file)
			}
		}
	}
	return nil
}
