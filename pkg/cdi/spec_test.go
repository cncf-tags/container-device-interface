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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/stretchr/testify/require"
	"tags.cncf.io/container-device-interface/pkg/cdi/validate"
	"tags.cncf.io/container-device-interface/pkg/parser"
	"tags.cncf.io/container-device-interface/schema"
	cdi "tags.cncf.io/container-device-interface/specs-go"
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
			file, err := mkTestSpec(t, []byte(tc.data))
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

			spec, err = newSpec(raw, tc.name, 0)
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

			spec, err = newSpec(raw, filepath.Join(dir, tc.name), 0)
			if tc.invalid {
				require.Error(t, err, "newSpec with invalid data")
				require.Nil(t, spec, "newSpec with invalid data")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, spec)

			err = spec.write(true)
			require.NoError(t, err)
			_, err = os.Stat(spec.GetPath())
			require.NoError(t, err, "spec.Write destination file")

			err = spec.write(false)
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
			file, err := mkTestSpec(t, []byte(tc.data))
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
				require.Equal(t, parser.QualifiedName(vendor, class, name), d.GetQualifiedName())
			}
		})
	}
}

// Create an automatically cleaned up temporary file for a test.
func mkTestSpec(t *testing.T, data []byte) (string, error) {
	tmp, err := os.CreateTemp("", ".cdi-test.*."+specType(data))
	if err != nil {
		return "", fmt.Errorf("failed to create test file: %w", err)
	}

	if data != nil {
		_, err := tmp.Write(data)
		if err != nil {
			return "", fmt.Errorf("failed to write test file content: %w", err)
		}
	}

	file := tmp.Name()
	t.Cleanup(func() {
		os.Remove(file)
	})

	tmp.Close()
	return file, nil
}

func specType(content []byte) string {
	spec := strings.TrimSpace(string(content))
	if spec != "" && spec[0] == '{' {
		return "json"
	}
	return "yaml"
}

func TestCurrentVersionIsValid(t *testing.T) {
	require.NoError(t, validateVersion(cdi.CurrentVersion))
}

func TestRequiredVersion(t *testing.T) {

	testCases := []struct {
		description     string
		spec            *cdi.Spec
		expectedVersion string
	}{
		{
			description:     "empty spec returns lowest version",
			spec:            &cdi.Spec{},
			expectedVersion: "0.3.0",
		},
		{
			description: "hostPath set returns version 0.5.0",
			spec: &cdi.Spec{
				ContainerEdits: cdi.ContainerEdits{
					DeviceNodes: []*cdi.DeviceNode{
						{
							HostPath: "/host/path/set",
						},
					},
				},
			},
			expectedVersion: "0.5.0",
		},
		{
			description: "hostPath equal to Path required v0.5.0",
			spec: &cdi.Spec{
				ContainerEdits: cdi.ContainerEdits{
					DeviceNodes: []*cdi.DeviceNode{
						{
							HostPath: "/some/path",
							Path:     "/some/path",
						},
					},
				},
			},
			expectedVersion: "0.5.0",
		},
		{
			description: "mount type set returns version 0.4.0",
			spec: &cdi.Spec{
				ContainerEdits: cdi.ContainerEdits{
					Mounts: []*cdi.Mount{
						{
							Type: "bind",
						},
					},
				},
			},
			expectedVersion: "0.4.0",
		},
		{
			description: "newest required version is selected",
			spec: &cdi.Spec{
				Annotations: map[string]string{
					"key": "value",
				},
				ContainerEdits: cdi.ContainerEdits{
					DeviceNodes: []*cdi.DeviceNode{
						{
							HostPath: "/host/path/set",
						},
					},
					Mounts: []*cdi.Mount{
						{
							Type: "bind",
						},
					},
				},
			},
			expectedVersion: "0.6.0",
		},
		{
			description: "device with name starting with digit requires v0.5.0",
			spec: &cdi.Spec{
				Devices: []cdi.Device{
					{
						Name: "0",
						ContainerEdits: cdi.ContainerEdits{
							Env: []string{
								"FOO=bar",
							},
						},
					},
				},
			},
			expectedVersion: "0.5.0",
		},
		{
			description: "device with name starting with letter requires minimum version",
			spec: &cdi.Spec{
				Devices: []cdi.Device{
					{
						Name: "device0",
						ContainerEdits: cdi.ContainerEdits{
							Env: []string{
								"FOO=bar",
							},
						},
					},
				},
			},
			expectedVersion: "0.3.0",
		},
		{
			description: "spec annotations require v0.6.0",
			spec: &cdi.Spec{
				Annotations: map[string]string{
					"key": "value",
				},
			},
			expectedVersion: "0.6.0",
		},
		{
			description: "device annotations require v0.6.0",
			spec: &cdi.Spec{
				Devices: []cdi.Device{
					{
						Name: "device0",
						Annotations: map[string]string{
							"key": "value",
						},
						ContainerEdits: cdi.ContainerEdits{
							Env: []string{
								"FOO=bar",
							},
						},
					},
				},
			},
			expectedVersion: "0.6.0",
		},
		{
			description: "dotted name (class) label require v0.6.0",
			spec: &cdi.Spec{
				Kind: "vendor.com/class.sub",
			},
			expectedVersion: "0.6.0",
		},
		{
			description: "IntelRdt requires v0.7.0",
			spec: &cdi.Spec{
				ContainerEdits: cdi.ContainerEdits{
					IntelRdt: &cdi.IntelRdt{
						ClosID: "foo",
					},
				},
			},
			expectedVersion: "0.7.0",
		},
		{
			description: "IntelRdt (on devices) requires v0.7.0",
			spec: &cdi.Spec{
				Devices: []cdi.Device{
					{
						Name: "device0",
						ContainerEdits: cdi.ContainerEdits{
							IntelRdt: &cdi.IntelRdt{
								ClosID: "foo",
							},
						},
					},
				},
			},
			expectedVersion: "0.7.0",
		},
		{
			description: "additionalGIDs in spec require v0.7.0",
			spec: &cdi.Spec{
				ContainerEdits: cdi.ContainerEdits{
					AdditionalGIDs: []uint32{5},
				},
			},
			expectedVersion: "0.7.0",
		},
		{

			description: "additionalGIDs in device require v0.7.0",
			spec: &cdi.Spec{
				Devices: []cdi.Device{
					{
						Name: "device0",
						ContainerEdits: cdi.ContainerEdits{
							AdditionalGIDs: []uint32{5},
						},
					},
				},
			},
			expectedVersion: "0.7.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			v, err := MinimumRequiredVersion(tc.spec)
			require.NoError(t, err)

			require.Equal(t, tc.expectedVersion, v)
		})
	}
}
