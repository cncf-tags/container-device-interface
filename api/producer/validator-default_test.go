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
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	cdi "tags.cncf.io/container-device-interface/specs-go"
)

func TestValidateSpec(t *testing.T) {
	testCases := []struct {
		name    string
		spec    string
		invalid bool
	}{
		{
			name: "invalid kind",
			spec: `
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
			invalid: true,
		},
		{
			name: "no device name",
			spec: `
cdiVersion: "0.3.0"
kind:       "vendor3.com/device"
containerEdits:
  deviceNodes:
  - path: "/dev/vendor3-dev1"
    type: b
    major: 10
    minor: 1
`,
			invalid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := cdi.Spec{}
			err := yaml.Unmarshal([]byte(tc.spec), &spec)
			require.NoError(t, err)
			require.NotNil(t, spec)

			err = DefaultValidator.Validate(&spec)
			if tc.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateContainerEdits(t *testing.T) {
	type testCase struct {
		name    string
		edits   *cdi.ContainerEdits
		invalid bool
	}
	for _, tc := range []*testCase{
		{
			name:  "valid, empty edits",
			edits: &cdi.ContainerEdits{},
		},
		{
			name: "valid, env var",
			edits: &cdi.ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
			},
		},
		{
			name: "invalid env, empty var",
			edits: &cdi.ContainerEdits{
				Env: []string{""},
			},
			invalid: true,
		},
		{
			name: "invalid env, no var name",
			edits: &cdi.ContainerEdits{
				Env: []string{"=foo"},
			},
			invalid: true,
		},
		{
			name: "invalid env, no assignment",
			edits: &cdi.ContainerEdits{
				Env: []string{"FOOBAR"},
			},
			invalid: true,
		},
		{
			name: "valid device, path only",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
					},
				},
			},
		},
		{
			name: "valid device, path+type",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
						Type: "c",
					},
				},
			},
		},
		{
			name: "valid device, path+type+permissions",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:        "/dev/null",
						Type:        "b",
						Permissions: "rwm",
					},
				},
			},
		},
		{
			name: "invalid device, empty path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid device, wrong type",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/vendorctl",
						Type: "f",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid device, wrong permissions",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:        "/dev/vendorctl",
						Type:        "b",
						Permissions: "to land",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid mount",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
		},
		{
			name: "invalid mount, empty host path",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid mount, empty container path",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid hooks",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
					{
						HookName: "createRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
					{
						HookName: "createContainer",
						Path:     "/usr/local/bin/cc-vendor-hook",
						Args:     []string{"--create"},
						Env:      []string{"VENDOR_ENV3=value3"},
					},
					{
						HookName: "startContainer",
						Path:     "/usr/local/bin/sc-vendor-hook",
						Args:     []string{"--start"},
						Env:      []string{"VENDOR_ENV4=value4"},
					},
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/poststart-vendor-hook",
						Env:      []string{"VENDOR_ENV5=value5"},
					},
					{
						HookName: "poststop",
						Path:     "/usr/local/bin/poststop-vendor-hook",
					},
				},
			},
		},
		{
			name: "invalid hook, empty path",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid hook, wrong hook name",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "misCreateRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid hook, wrong env",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"=value2"},
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid rdt config",
			edits: &cdi.ContainerEdits{
				IntelRdt: &cdi.IntelRdt{
					ClosID: "foo.bar",
				},
			},
		},
		{
			name: "invalid rdt config, invalid closID (slash)",
			edits: &cdi.ContainerEdits{
				IntelRdt: &cdi.IntelRdt{
					ClosID: "foo/bar",
				},
			},
			invalid: true,
		},
		{
			name: "invalid rdt config, invalid closID (dot)",
			edits: &cdi.ContainerEdits{
				IntelRdt: &cdi.IntelRdt{
					ClosID: ".",
				},
			},
			invalid: true,
		},
		{
			name: "invalid rdt config, invalid closID (double dot)",
			edits: &cdi.ContainerEdits{
				IntelRdt: &cdi.IntelRdt{
					ClosID: "..",
				},
			},
			invalid: true,
		},
		{
			name: "invalid rdt config, invalid closID (newline)",
			edits: &cdi.ContainerEdits{
				IntelRdt: &cdi.IntelRdt{
					ClosID: "foo\nbar",
				},
			},
			invalid: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := defaultValidator("default").validateEdits(tc.edits)
			if tc.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
