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
	"sigs.k8s.io/yaml"

	"github.com/container-orchestrated-devices/container-device-interface/pkg/cdi/validate"
	"github.com/container-orchestrated-devices/container-device-interface/schema"
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
			name:       "empty",
			data:       "",
			unparsable: true,
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
		schemaFail bool
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
			schemaFail: true,
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
			schemaFail: true,
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
			name: "invalid, missing required edits",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev1"
`,
			schemaFail: true,
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

			raw, err = ParseSpec([]byte(tc.data))
			if tc.unparsable {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			spec, err = NewSpec(raw, tc.name, 0)
			if tc.invalid || tc.schemaFail {
				require.Error(t, err)
				require.Nil(t, spec)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, spec)
		})
	}
}

func TestWriteSpec(t *testing.T) {
	type testCase struct {
		name    string
		data    string
		invalid bool
	}
	for _, tc := range []*testCase{
		{
			name:    "invalid-spec1.yaml",
			invalid: true,
			data: `
cdiVersion: ""
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      env:
        - "FOO=BAR"
`,
		},
		{
			name:    "invalid-spec2.yaml",
			invalid: true,
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
containerEdits:
  env:
    - "FOO=BAR"
`,
		},
		{
			name: "spec1.yaml",
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
		{
			name: "spec2.yaml",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev4"
    containerEdits:
      env:
        - "BAR=FOO"
  - name: "dev5"
    containerEdits:
      env:
        - "XYZ=ZY"
  - name: "dev6"
    containerEdits:
      env:
        - "BAR=SPACE"
`,
		},
		{
			name: "spec3.yaml",
			data: `
cdiVersion: "0.3.0"
kind: vendor.com/device
devices:
  - name: "dev7"
    containerEdits:
      env:
        - "FOO=BAR"
  - name: "dev8"
    containerEdits:
      env:
        - "FOOBAR=BARFOO"
  - name: "dev9"
    containerEdits:
      env:
        - "SPACE=BAR"
`,
		},
	} {
		dir, err := mkTestDir(t, nil)
		require.NoError(t, err)

		SetSpecValidator(validate.WithDefaultSchema())
		defer SetSpecValidator(validate.WithSchema(schema.NopSchema()))

		t.Run(tc.name, func(t *testing.T) {
			var (
				raw  = &cdi.Spec{}
				spec *Spec
				chk  *Spec
				err  error
			)

			err = yaml.Unmarshal([]byte(tc.data), raw)
			require.NoError(t, err)

			spec, err = NewSpec(raw, filepath.Join(dir, tc.name), 0)
			if tc.invalid {
				require.Error(t, err, "NewSpec with invalid data")
				require.Nil(t, spec, "NewSpec with invalid data")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, spec)

			err = spec.Write(true)
			require.NoError(t, err)
			_, err = os.Stat(spec.GetPath())
			require.NoError(t, err, "spec.Write destination file")

			err = spec.Write(false)
			require.Error(t, err)

			chk, err = ReadSpec(spec.GetPath(), spec.GetPriority())
			require.NoError(t, err)
			require.NotNil(t, chk)
			require.Equal(t, spec.Spec, chk.Spec)
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
		{
			name:     "valid",
			priority: 1,
			data: `
cdiVersion: "0.4.0"
kind: vendor.com/device
devices:
  - name: "dev1"
    containerEdits:
      mounts:
        - hostPath: "tmpfs"
          containerPath: "/usr/local/container"
          type: "tmpfs"
          options:
            - "ro"
            - "mode=755"
            - "size=65536k"
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

func TestCurrentVersionIsValid(t *testing.T) {
	require.NoError(t, validateVersion(cdi.CurrentVersion))
}
