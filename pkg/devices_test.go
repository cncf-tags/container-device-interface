package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"

	cdispec "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

func TestExtractVendor(t *testing.T) {
	testcases := []struct {
		dev            string
		expectedVendor string
		expectedDevice string
	}{
		{
			dev:            "vendor.com/device=myDevice",
			expectedVendor: "vendor.com/device",
			expectedDevice: "myDevice",
		},
		{
			dev:            "vendor.com/device=all", // make sure this isn't a special case
			expectedVendor: "vendor.com/device",
			expectedDevice: "all",
		},
		{
			dev:            "myDevice",
			expectedVendor: "",
			expectedDevice: "myDevice",
		},
	}

	for _, test := range testcases {
		v, d := extractVendor(test.dev)
		require.Equal(t, v, test.expectedVendor)
		require.Equal(t, d, test.expectedDevice)
	}
}

func TestGetCDIForDevice(t *testing.T) {
	testcases := []struct {
		testname string
		getdev   string
		specs    map[string]*cdispec.Spec

		expectedKind  string
		expectedError bool
	}{
		{
			testname: "simple - Get Existing device",
			getdev:   "vendor.com/device=myDevice",

			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
			},
			expectedKind:  "vendor.com/device",
			expectedError: false,
		},
		{
			testname: "simple - Get non existing vendor",
			getdev:   "foovendor.com/device=myDevice3",

			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
			},
			expectedKind:  "",
			expectedError: true,
		},
		{
			testname: "simple - Get non existing device",
			getdev:   "vendor.com/device=myDevice3",

			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
			},
			expectedKind:  "",
			expectedError: true,
		},
		{
			testname: "simple - Get CDI with only device name",
			getdev:   "myDevice",

			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
			},
			expectedKind:  "vendor.com/device",
			expectedError: false,
		},
		{
			testname: "simple - Get non existing device and no vendor name",
			getdev:   "myDevice3",

			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
			},
			expectedKind:  "",
			expectedError: true,
		},
		{
			testname: "medium - device name is ambiguous",
			getdev:   "myDevice",

			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
				"bar-vendor.com/device": {
					Kind: "vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice"},
						{Name: "myDevice-2"},
					},
				},
			},
			expectedKind:  "",
			expectedError: true,
		},
		{
			testname: "medium - get device multiple vendors",
			getdev:   "myDevice-bar-2",

			specs: map[string]*cdispec.Spec{
				"foo-vendor.com/device": {
					Kind: "foo-vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "myDevice-foo-1"},
						{Name: "myDevice-foo-2"},
					},
				},
				"bar-vendor.com/device": {
					Kind: "bar-vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "mydevice-bar-1"},
						{Name: "myDevice-bar-2"},
					},
				},
				"baz-vendor.com/device": {
					Kind: "baz-vendor.com/device",
					Devices: []cdispec.Device{
						{Name: "mydevice-baz-1"},
						{Name: "myDevice-baz-2"},
					},
				},
			},
			expectedKind:  "bar-vendor.com/device",
			expectedError: false,
		},
	}

	for _, test := range testcases {
		s, err := GetCDIForDevice(test.getdev, test.specs)
		if test.expectedError == true {
			require.Error(t, err)
			continue
		}

		require.NotNil(t, s)
		require.Equal(t, test.expectedKind, s.Kind)
	}
}

func TestUpdateOCISpecForDevicesWithSpecs(t *testing.T) {
	testCases := []struct {
		name           string
		config         *spec.Spec
		devices        []string
		specs          map[string]*cdispec.Spec
		expectedResult spec.Spec
		expectedError  bool
	}{
		{
			name:    "add fully qualified device",
			config:  &spec.Spec{},
			devices: []string{"vendor.com/device=ABC", "vstartand.org/prodsnow=XYZ"},
			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Version: "0.3.0",
					Kind:    "vendor.com/device",
					Devices: []cdispec.Device{
						{
							Name: "XYZ",
							ContainerEdits: cdispec.ContainerEdits{
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/iceXYZ",
									},
								},
							},
						},
						{
							Name: "ABC",
							ContainerEdits: cdispec.ContainerEdits{
								Env: []string{"OR=ICE"},
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/iceABC",
									},
								},
							},
						},
					},
				},
				"vstartand.org/prodsnow": {
					Version: "0.3.0",
					Kind:    "vstartand.org/prodsnow",
					Devices: []cdispec.Device{
						{
							Name: "XYZ",
							ContainerEdits: cdispec.ContainerEdits{
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/snowXYZ",
									},
								},
							},
						},
						{
							Name: "ABC",
							ContainerEdits: cdispec.ContainerEdits{
								Env: []string{"AND=SNOW"},
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/snowABC",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: spec.Spec{
				Process: &spec.Process{
					Env: []string{"OR=ICE"},
				},
				Linux: &spec.Linux{
					Devices: []spec.LinuxDevice{
						{Path: "/dev/iceABC"},
						{Path: "/dev/snowXYZ"},
					},
				},
			},
		},
		{
			name:    "add unqualified device",
			config:  &spec.Spec{},
			devices: []string{"BAR", "ABC"},
			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Version: "0.3.0",
					Kind:    "vendor.com/device",
					Devices: []cdispec.Device{
						{
							Name: "XYZ",
							ContainerEdits: cdispec.ContainerEdits{
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/iceXYZ",
									},
								},
							},
						},
						{
							Name: "ABC",
							ContainerEdits: cdispec.ContainerEdits{
								Env: []string{"OR=ICE"},
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/iceABC",
									},
								},
							},
						},
					},
				},
				"vstartand.org/prodsnow": {
					Version: "0.3.0",
					Kind:    "vstartand.org/prodsnow",
					Devices: []cdispec.Device{
						{
							Name: "FOO",
							ContainerEdits: cdispec.ContainerEdits{
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/snowFOO",
									},
								},
							},
						},
						{
							Name: "BAR",
							ContainerEdits: cdispec.ContainerEdits{
								Env: []string{"AND=SNOW"},
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/snowBAR",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: spec.Spec{
				Process: &spec.Process{
					Env: []string{"AND=SNOW", "OR=ICE"},
				},
				Linux: &spec.Linux{
					Devices: []spec.LinuxDevice{
						{Path: "/dev/snowBAR"},
						{Path: "/dev/iceABC"},
					},
				},
			},
		},
		{
			name:    "nonexistent device",
			config:  &spec.Spec{},
			devices: []string{"BAR"},
			specs: map[string]*cdispec.Spec{
				"vendor.com/device": {
					Version: "0.3.0",
					Kind:    "vendor.com/device",
					Devices: []cdispec.Device{
						{
							Name: "XYZ",
							ContainerEdits: cdispec.ContainerEdits{
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/iceXYZ",
									},
								},
							},
						},
						{
							Name: "ABC",
							ContainerEdits: cdispec.ContainerEdits{
								Env: []string{"END=ICE"},
								DeviceNodes: []*cdispec.DeviceNode{
									{
										Path: "/dev/iceABC",
									},
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
			err := UpdateOCISpecForDevicesWithSpec(tc.config, tc.devices, tc.specs)
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedResult, *tc.config)
		})
	}
}
