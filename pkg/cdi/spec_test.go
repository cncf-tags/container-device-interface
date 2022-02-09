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
	"testing"

	"github.com/pkg/errors"

	cdi "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"github.com/stretchr/testify/require"
)

func TestReadSpec(t *testing.T) {
	type testCase struct {
		name       string
		data       string
		unparsable bool
		invalid    bool
	}
	for _, tc := range []*testCase{
		{
			name: "unparsable",
			data: `
xyzzy: garbled
`,
			unparsable: true,
		},
		{
			name: "invalid, missing CDI version",
			data: `
kind:    "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name: "valid",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := mkTestFile(t, []byte(tc.data))
			if err != nil {
				t.Errorf("failed to create CDI Spec test file: %v", err)
				return
			}

			spec, err := ReadSpec(file, 0)
			if tc.unparsable {
				require.Error(t, err)
				require.Nil(t, spec)
				return
			}
			if tc.invalid {
				require.Error(t, err)
				require.Nil(t, spec)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, spec)
		})
	}
}

func TestNewSpec(t *testing.T) {
	type testCase struct {
		name       string
		data       string
		unparsable bool
		invalid    bool
	}
	for _, tc := range []*testCase{
		{
			name: "unparsable",
			data: `
xyzzy: garbled
`,
			unparsable: true,
		},
		{
			name: "invalid, missing CDI version",
			data: `
kind:    "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name: "invalid, invalid CDI version",
			data: `
cdiVersion: "0.0.0"
kind:       "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name: "invalid, missing vendor/class",
			data: `
cdiVersion: "0.3.0"
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name: "invalid, invalid vendor",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com-/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name: "invalid, invalid class",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device=
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name: "invalid, invalid device",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
`,
			invalid: true,
		},
		{
			name: "invalid, conflicting devices",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
  - name: "dev2"
    containerEdits:
      env:
        - "BAR=FOO"
  - name: "dev1"
    containerEdits:
      env:
        - "SPACE=BAR"
`,
			invalid: true,
		},
		{
			name: "valid",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
  - name: "dev2"
    containerEdits:
      env:
        - "BAR=FOO"
  - name: "dev3"
    containerEdits:
      env:
        - "SPACE=BAR"
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				raw  *cdi.Spec
				spec *Spec
				err  error
			)

			raw, err = parseSpec([]byte(tc.data))
			if tc.unparsable {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			spec, err = NewSpec(raw, tc.name, 0)
			if tc.invalid {
				require.Error(t, err)
				require.Nil(t, spec)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, spec)
		})
	}
}

func TestGetters(t *testing.T) {
	type testCase struct {
		name     string
		priority int
		data     string
		invalid  bool
		vendor   string
		class    string
	}
	for _, tc := range []*testCase{
		{
			name: "unparsable",
			data: `
xyzzy: garbled
`,
			invalid: true,
		},
		{
			name: "invalid, missing CDI version",
			data: `
kind:    "vendor.com/device"
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			invalid: true,
		},
		{
			name:     "valid",
			priority: 1,
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
			vendor: "vendor.com",
			class:  "device",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := mkTestFile(t, []byte(tc.data))
			if err != nil {
				t.Errorf("failed to create CDI Spec test file: %v", err)
				return
			}

			spec, err := ReadSpec(file, tc.priority)
			if tc.invalid {
				require.Error(t, err)
				require.Nil(t, spec)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, spec)
			require.Equal(t, file, spec.GetPath())
			require.Equal(t, tc.priority, spec.GetPriority())

			vendor, class := spec.GetVendor(), spec.GetClass()
			require.Equal(t, tc.vendor, vendor)
			require.Equal(t, tc.class, class)

			for name, d := range spec.devices {
				require.Equal(t, spec, d.GetSpec())
				require.Equal(t, d, spec.GetDevice(name))
				require.Equal(t, QualifiedName(vendor, class, name), d.GetQualifiedName())
			}
		})
	}
}

// Create an automatically cleaned up temporary file for a test.
func mkTestFile(t *testing.T, data []byte) (string, error) {
	tmp, err := ioutil.TempFile("", ".cdi-test.*")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create test file")
	}

	if data != nil {
		_, err := tmp.Write(data)
		if err != nil {
			return "", errors.Wrap(err, "failed to write test file content")
		}
	}

	file := tmp.Name()
	t.Cleanup(func() {
		os.Remove(file)
	})

	tmp.Close()
	return file, nil
}
