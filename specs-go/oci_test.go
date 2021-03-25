package specs

import (
	"testing"

	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestApplyEditsToOCISpec(t *testing.T) {
	testCases := []struct {
		name           string
		config         *spec.Spec
		edits          *ContainerEdits
		expectedResult spec.Spec
		expectedError  bool
	}{
		{
			name:          "nil spec",
			expectedError: true,
		},
		{
			name:   "nil edits",
			config: &spec.Spec{},
			edits:  nil,
		},
		{
			name:   "add env to the empty spec",
			config: &spec.Spec{},
			edits: &ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
			},
			expectedResult: spec.Spec{
				Process: &spec.Process{
					Env: []string{"BAR=BARVALUE1"},
				},
			},
		},
		{
			name:   "add device nodes to the empty spec",
			config: &spec.Spec{},
			edits: &ContainerEdits{
				DeviceNodes: []*DeviceNode{
					{
						Path: "/dev/vendorctl",
					},
				},
			},
			expectedResult: spec.Spec{
				Linux: &spec.Linux{
					Devices: []spec.LinuxDevice{
						{Path: "/dev/vendorctl"},
					},
				},
			},
		},
		{
			name:   "add mounts to the empty spec",
			config: &spec.Spec{},
			edits: &ContainerEdits{
				Mounts: []*Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
			expectedResult: spec.Spec{
				Mounts: []spec.Mount{
					{
						Source:      "/dev/vendorctl",
						Destination: "/dev/vendorctl",
					},
				},
			},
		},
		{
			name:   "add hooks to the empty spec",
			config: &spec.Spec{},
			edits: &ContainerEdits{
				Hooks: []*Hook{
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
			expectedResult: spec.Spec{
				Hooks: &spec.Hooks{
					Prestart: []spec.Hook{
						{
							Path: "/usr/local/bin/prestart-vendor-hook",
							Args: []string{"--verbose"},
							Env:  []string{"VENDOR_ENV1=value1"},
						},
					},
					CreateRuntime: []spec.Hook{
						{
							Path: "/usr/local/bin/cr-vendor-hook",
							Args: []string{"--debug"},
							Env:  []string{"VENDOR_ENV2=value2"},
						},
					},
					CreateContainer: []spec.Hook{
						{
							Path: "/usr/local/bin/cc-vendor-hook",
							Args: []string{"--create"},
							Env:  []string{"VENDOR_ENV3=value3"},
						},
					},
					StartContainer: []spec.Hook{
						{
							Path: "/usr/local/bin/sc-vendor-hook",
							Args: []string{"--start"},
							Env:  []string{"VENDOR_ENV4=value4"},
						},
					},
					Poststart: []spec.Hook{
						{
							Path: "/usr/local/bin/poststart-vendor-hook",
							Env:  []string{"VENDOR_ENV5=value5"},
						},
					},
					Poststop: []spec.Hook{
						{
							Path: "/usr/local/bin/poststop-vendor-hook",
						},
					},
				},
			},
		},
		{
			name:   "unknown hook",
			config: &spec.Spec{},
			edits: &ContainerEdits{
				Hooks: []*Hook{
					{
						HookName: "unknown",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
				},
			},
			expectedResult: spec.Spec{
				Hooks: &spec.Hooks{},
			},
		},
		{
			name: "multiple edits",
			config: &spec.Spec{
				Version: "1.0.2",
				Process: &spec.Process{
					Env: []string{"ENV=value"},
				},
				Root: &spec.Root{
					Path:     "/chroot/root1",
					Readonly: true,
				},
				Hostname: "some.host.com",
				Mounts: []spec.Mount{
					{
						Source:      "/source",
						Destination: "/destination",
						Type:        "tmpfs",
						Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
					},
				},
				Hooks: &spec.Hooks{
					Prestart: []spec.Hook{
						{
							Path: "/bin/hook",
							Args: []string{"--prestart"},
							Env:  []string{"HOOKENV=hookval"},
						},
					},
				},
			},
			edits: &ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
				DeviceNodes: []*DeviceNode{
					{
						Path: "/dev/device1",
					},
				},
				Hooks: []*Hook{
					{
						HookName: "prestart",
						Path:     "/bin/vendor-hook",
					},
					{
						HookName: "poststart",
						Path:     "/bin/poststart",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
				},
				Mounts: []*Mount{
					{
						HostPath:      "/mnt/mount1",
						ContainerPath: "/mnt/mount1",
						Options:       []string{"noexec", "noatime"},
					},
				},
			},
			expectedResult: spec.Spec{
				Version: "1.0.2",
				Process: &spec.Process{
					Env: []string{"ENV=value", "BAR=BARVALUE1"},
				},
				Root: &spec.Root{
					Path:     "/chroot/root1",
					Readonly: true,
				},
				Hostname: "some.host.com",
				Linux: &spec.Linux{
					Devices: []spec.LinuxDevice{
						{Path: "/dev/device1"},
					},
				},
				Mounts: []spec.Mount{
					{
						Source:      "/source",
						Destination: "/destination",
						Type:        "tmpfs",
						Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
					},
					{
						Source:      "/mnt/mount1",
						Destination: "/mnt/mount1",
						Options:     []string{"noexec", "noatime"},
					},
				},
				Hooks: &spec.Hooks{
					Prestart: []spec.Hook{
						{
							Path: "/bin/hook",
							Args: []string{"--prestart"},
							Env:  []string{"HOOKENV=hookval"},
						},
						{
							Path: "/bin/vendor-hook",
						},
					},
					Poststart: []spec.Hook{
						{
							Path: "/bin/poststart",
							Args: []string{"--verbose"},
							Env:  []string{"VENDOR_ENV1=value1"},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ApplyEditsToOCISpec(tc.config, tc.edits)
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.edits != nil {
				require.Equal(t, tc.expectedResult, *tc.config)
			}
		})
	}
}
func TestApplyOCIEdits(t *testing.T) {
	testCases := []struct {
		name           string
		config         *spec.Spec
		cdiSpec        *Spec
		expectedResult spec.Spec
	}{
		{
			name:   "edit empty spec",
			config: &spec.Spec{},
			cdiSpec: &Spec{
				Version: "0.2.0",
				Kind:    "vendor.com/device",
				Devices: []Devices{},
				ContainerEdits: ContainerEdits{
					Env: []string{"FOO=VALID_SPEC", "BAR=BARVALUE1"},
					DeviceNodes: []*DeviceNode{
						{
							Path: "/dev/device1",
						},
					},
				},
			},
			expectedResult: spec.Spec{
				Process: &spec.Process{
					Env: []string{"FOO=VALID_SPEC", "BAR=BARVALUE1"},
				},
				Linux: &spec.Linux{
					Devices: []spec.LinuxDevice{
						{Path: "/dev/device1"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ApplyOCIEdits(tc.config, tc.cdiSpec)
			require.NoError(t, err)
			require.Equal(t, tc.expectedResult, *tc.config)
		})
	}
}

func TestApplyOCIEditsForDevice(t *testing.T) {
	testCases := []struct {
		name           string
		config         *spec.Spec
		cdiSpec        *Spec
		dev            string
		expectedResult spec.Spec
		expectedError  bool
	}{
		{
			name:   "add device to the empty spec",
			config: &spec.Spec{},
			cdiSpec: &Spec{
				Version: "0.2.0",
				Kind:    "vendor.com/device",
				Devices: []Devices{
					{
						Name: "Vendor device XYZ",
						ContainerEdits: ContainerEdits{
							Env: []string{"FOO=VALID_SPEC", "BAR=BARVALUE1"},
							DeviceNodes: []*DeviceNode{
								{
									Path: "/dev/device1",
								},
							},
						},
					},
					{
						Name: "Vendor device ABC",
						ContainerEdits: ContainerEdits{
							DeviceNodes: []*DeviceNode{
								{
									Path: "/dev/devABC",
								},
							},
						},
					},
				},
			},
			expectedResult: spec.Spec{
				Process: &spec.Process{
					Env: []string{"FOO=VALID_SPEC", "BAR=BARVALUE1"},
				},
				Linux: &spec.Linux{
					Devices: []spec.LinuxDevice{
						{Path: "/dev/device1"},
					},
				},
			},
		},
		{
			name:   "device not found",
			config: &spec.Spec{},
			cdiSpec: &Spec{
				Version: "0.2.0",
				Kind:    "vendor.com/device",
				Devices: []Devices{
					{
						Name: "Vendor device ABC",
						ContainerEdits: ContainerEdits{
							DeviceNodes: []*DeviceNode{
								{
									Path: "/dev/devABC",
								},
							},
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ApplyOCIEditsForDevice(tc.config, tc.cdiSpec, "Vendor device XYZ")
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedResult, *tc.config)
		})
	}
}
