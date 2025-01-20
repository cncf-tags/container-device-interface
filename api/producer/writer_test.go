/*
   Copyright © 2024 The CDI Authors

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

package producer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

func TestSave(t *testing.T) {
	testCases := []struct {
		description         string
		spec                cdi.Spec
		options             []Option
		filename            string
		expectedError       error
		expectedFilename    string
		expectedPermissions os.FileMode
		expectedOutput      string
	}{
		{
			description: "output as json",
			spec: cdi.Spec{
				Version: "0.3.0",
				Kind:    "example.com/class",
				Devices: []cdi.Device{
					{
						Name: "dev1",
						ContainerEdits: cdi.ContainerEdits{
							DeviceNodes: []*cdi.DeviceNode{
								{
									Path: "/dev/foo",
								},
							},
						},
					},
				},
			},
			options:             []Option{},
			filename:            "foo.json",
			expectedFilename:    "foo.json",
			expectedPermissions: 0600,
			expectedOutput:      `{"cdiVersion":"0.3.0","kind":"example.com/class","devices":[{"name":"dev1","containerEdits":{"deviceNodes":[{"path":"/dev/foo"}]}}],"containerEdits":{}}`,
		},
		{
			description: "output with permissions",
			spec: cdi.Spec{
				Version: "0.3.0",
				Kind:    "example.com/class",
				Devices: []cdi.Device{
					{
						Name: "dev1",
						ContainerEdits: cdi.ContainerEdits{
							DeviceNodes: []*cdi.DeviceNode{
								{
									Path: "/dev/foo",
								},
							},
						},
					},
				},
			},
			options:             []Option{WithPermissions(0644)},
			filename:            "foo.json",
			expectedFilename:    "foo.json",
			expectedPermissions: 0644,
			expectedOutput:      `{"cdiVersion":"0.3.0","kind":"example.com/class","devices":[{"name":"dev1","containerEdits":{"deviceNodes":[{"path":"/dev/foo"}]}}],"containerEdits":{}}`,
		},
		{
			description: "filename overwrites format",
			spec: cdi.Spec{
				Version: "0.3.0",
				Kind:    "example.com/class",
				Devices: []cdi.Device{
					{
						Name: "dev1",
						ContainerEdits: cdi.ContainerEdits{
							DeviceNodes: []*cdi.DeviceNode{
								{
									Path: "/dev/foo",
								},
							},
						},
					},
				},
			},
			options:             []Option{WithSpecFormat(SpecFormatJSON)},
			filename:            "foo.yaml",
			expectedFilename:    "foo.yaml",
			expectedPermissions: 0600,
			expectedOutput: `---
cdiVersion: 0.3.0
containerEdits: {}
devices:
- containerEdits:
    deviceNodes:
    - path: /dev/foo
  name: dev1
kind: example.com/class
`,
		},
		{
			description: "filename is inferred from format",
			spec: cdi.Spec{
				Version: "0.3.0",
				Kind:    "example.com/class",
				Devices: []cdi.Device{
					{
						Name: "dev1",
						ContainerEdits: cdi.ContainerEdits{
							DeviceNodes: []*cdi.DeviceNode{
								{
									Path: "/dev/foo",
								},
							},
						},
					},
				},
			},
			options:             []Option{WithSpecFormat(SpecFormatYAML)},
			filename:            "foo",
			expectedFilename:    "foo.yaml",
			expectedPermissions: 0600,
			expectedOutput: `---
cdiVersion: 0.3.0
containerEdits: {}
devices:
- containerEdits:
    deviceNodes:
    - path: /dev/foo
  name: dev1
kind: example.com/class
`,
		},
		{
			description: "minimum version is detected",
			spec: cdi.Spec{
				Version: cdi.CurrentVersion,
				Kind:    "example.com/class",
				Devices: []cdi.Device{
					{
						Name: "dev1",
						ContainerEdits: cdi.ContainerEdits{
							DeviceNodes: []*cdi.DeviceNode{
								{
									Path: "/dev/foo",
								},
							},
						},
					},
				},
			},
			options:             []Option{WithDetectMinimumVersion(true)},
			filename:            "foo",
			expectedFilename:    "foo.yaml",
			expectedPermissions: 0600,
			expectedOutput: `---
cdiVersion: 0.3.0
containerEdits: {}
devices:
- containerEdits:
    deviceNodes:
    - path: /dev/foo
  name: dev1
kind: example.com/class
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			outputDir := t.TempDir()

			p, err := NewSpecWriter(tc.options...)
			require.NoError(t, err)

			f, err := p.Save(&tc.spec, filepath.Join(outputDir, tc.filename))
			require.ErrorIs(t, err, tc.expectedError)
			if tc.expectedError != nil {
				return
			}

			require.Equal(t, filepath.Join(outputDir, tc.expectedFilename), f)
			info, err := os.Stat(f)
			require.NoError(t, err)

			require.Equal(t, tc.expectedPermissions, info.Mode())

			contents, _ := os.ReadFile(f)
			require.Equal(t, tc.expectedOutput, string(contents))
		})
	}
}

type validatorWithError struct {
	err error
}

func (v *validatorWithError) Validate(*cdi.Spec) error {
	return error(v.err)
}
