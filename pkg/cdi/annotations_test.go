/*
   Copyright © 2022 The CDI Authors

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
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnnotationKey(t *testing.T) {
	type testCase = struct {
		name    string
		plugin  string
		devID   string
		key     string
		invalid bool
	}

	for _, tc := range []*testCase{
		{
			name:    "invalid, empty plugin",
			plugin:  "",
			invalid: true,
		},
		{
			name:    "invalid, empty device ID",
			plugin:  "plugin",
			devID:   "",
			invalid: true,
		},
		{
			name:    "invalid, non-alphanumeric first character",
			plugin:  "_vendor.class",
			devID:   "device",
			invalid: true,
		},
		{
			name:    "invalid, non-alphanumeric last character",
			plugin:  "vendor.class",
			devID:   "device_",
			invalid: true,
		},
		{
			name:    "invalid, plugin contains invalid characters",
			plugin:  "ven.dor-cl+ass",
			devID:   "device",
			invalid: true,
		},
		{
			name:    "invalid, devID contains invalid characters",
			plugin:  "vendor.class",
			devID:   "dev+ice",
			invalid: true,
		},
		{
			name:    "invalid, too plugin long",
			plugin:  "123456789012345678901234567890123456789012345678901234567",
			devID:   "device",
			invalid: true,
		},
		{
			name:   "valid, simple",
			plugin: "vendor.class",
			devID:  "device",
			key:    AnnotationPrefix + "vendor.class" + "_" + "device",
		},
		{
			name:   "valid, with special characters",
			plugin: "v-e.n_d.or.cl-as_s",
			devID:  "d_e-v-i-c_e",
			key:    AnnotationPrefix + "v-e.n_d.or.cl-as_s" + "_" + "d_e-v-i-c_e",
		},
		{
			name:   "valid, with /'s replaced in devID",
			plugin: "v-e.n_d.or.cl-as_s",
			devID:  "d-e/v/i/c-e",
			key:    AnnotationPrefix + "v-e.n_d.or.cl-as_s" + "_" + "d-e_v_i_c-e",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, err := AnnotationKey(tc.plugin, tc.devID)
			if !tc.invalid {
				require.NoError(t, err, "annotation key")
				require.Equal(t, tc.key, key, "annotation key")
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateAnnotations(t *testing.T) {
	type inject = struct {
		plugin  string
		devID   string
		devices []string
	}
	type testCase = struct {
		name        string
		existing    map[string]string
		injections  []*inject
		annotations map[string]string
		parsed      []string
		invalid     bool
	}

	for _, tc := range []*testCase{
		{
			name: "one plugin, one device",
			injections: []*inject{
				{
					plugin: "vendor.class",
					devID:  "device",
					devices: []string{
						"vendor.com/class=device",
					},
				},
			},
			annotations: map[string]string{
				AnnotationPrefix + "vendor.class_device": "vendor.com/class=device",
			},
			parsed: []string{
				"vendor.com/class=device",
			},
		},
		{
			name: "one plugin, multiple devices",
			injections: []*inject{
				{
					plugin: "vendor.class",
					devID:  "device",
					devices: []string{
						"vendor.com/class=device1",
						"vendor.com/class=device2",
						"vendor.com/class=device3",
					},
				},
			},
			annotations: map[string]string{
				AnnotationPrefix + "vendor.class_device": "vendor.com/class=device1,vendor.com/class=device2,vendor.com/class=device3",
			},
			parsed: []string{
				"vendor.com/class=device1",
				"vendor.com/class=device2",
				"vendor.com/class=device3",
			},
		},
		{
			name: "multiple plugins, multiple devices",
			injections: []*inject{
				{
					plugin: "vendor1.class",
					devID:  "device1",
					devices: []string{
						"vendor1.com/class=device1",
					},
				},
				{
					plugin: "vendor1.class",
					devID:  "device2",
					devices: []string{
						"vendor2.com/class=device1",
						"vendor2.com/class=device2",
					},
				},
				{
					plugin: "vendor3.class2",
					devID:  "device",
					devices: []string{
						"vendor3.com/class2=device1",
						"vendor3.com/class2=device2",
						"vendor3.com/class2=device3",
					},
				},
			},
			annotations: map[string]string{
				AnnotationPrefix + "vendor1.class_device1": "vendor1.com/class=device1",
				AnnotationPrefix + "vendor1.class_device2": "vendor2.com/class=device1,vendor2.com/class=device2",
				AnnotationPrefix + "vendor3.class2_device": "vendor3.com/class2=device1,vendor3.com/class2=device2,vendor3.com/class2=device3",
			},
			parsed: []string{
				"vendor1.com/class=device1",
				"vendor2.com/class=device1",
				"vendor2.com/class=device2",
				"vendor3.com/class2=device1",
				"vendor3.com/class2=device2",
				"vendor3.com/class2=device3",
			},
		},
		{
			name: "invalid, empty plugin",
			injections: []*inject{
				{
					plugin: "vendor1.class",
					devID:  "device",
					devices: []string{
						"vendor1.com/class=device1",
					},
				},
				{
					plugin: "vendor2.class",
					devID:  "device",
					devices: []string{
						"vendor2.com/class=device1",
						"vendor2.com/class=device2",
					},
				},
				{
					plugin: "",
					devID:  "device",
					devices: []string{
						"vendor3.com/class2=device1",
						"vendor3.com/class2=device2",
						"vendor3.com/class2=device3",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid, malformed device reference",
			injections: []*inject{
				{
					plugin: "vendor1.class",
					devID:  "device",
					devices: []string{
						"vendor1.com/class=device1",
					},
				},
				{
					plugin: "vendor2.class",
					devID:  "device",
					devices: []string{
						"vendor2.com/class=device1",
						"vendor2.com/device2",
					},
				},
				{
					plugin: "vendor3.class2",
					devID:  "device",
					devices: []string{
						"vendor3.com/class2=device1",
						"vendor3.com/class2=device2",
						"vendor3.com/class2=device3",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid, pre-resolved device",
			injections: []*inject{
				{
					plugin: "vendor1.class",
					devID:  "device",
					devices: []string{
						"vendor1.com/class=device1",
					},
				},
				{
					plugin: "vendor2.class",
					devID:  "device",
					devices: []string{
						"vendor2.com/class=device1",
						"vendor2.com/class=device2",
					},
				},
				{
					plugin: "vendor3.class2",
					devID:  "device",
					devices: []string{
						"vendor3.com/class2=device1",
						"vendor3.com/class2=device2",
						"/dev/null",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid, conflicting keys",
			existing: map[string]string{
				AnnotationPrefix + "vendor3.class2_device": "vendor3.com/class2=device0",
			},
			injections: []*inject{
				{
					plugin: "vendor1.class",
					devID:  "device",
					devices: []string{
						"vendor1.com/class=device1",
					},
				},
				{
					plugin: "vendor2.class",
					devID:  "device",
					devices: []string{
						"vendor2.com/class=device1",
						"vendor2.com/class=device2",
					},
				},
				{
					plugin: "vendor3.class2",
					devID:  "device",
					devices: []string{
						"vendor3.com/class2=device1",
						"vendor3.com/class2=device2",
					},
				},
			},
			invalid: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				annotations map[string]string
				parsed      []string
				err         error
			)
			for _, i := range tc.injections {
				if tc.existing != nil {
					annotations = tc.existing
				}
				annotations, err = UpdateAnnotations(annotations, i.plugin, i.devID, i.devices)
				if !tc.invalid {
					require.NoError(t, err, "CDI device injection annotation")
				} else {
					if err != nil {
						break
					}
				}
			}
			if tc.invalid {
				require.Error(t, err, "invalid injection")
			} else {
				require.Equal(t, tc.annotations, annotations)
				_, parsed, err = ParseAnnotations(annotations)
				require.NoError(t, err, "annotation parsing")
				sort.Strings(tc.parsed)
				sort.Strings(parsed)
				require.Equal(t, tc.parsed, parsed, "parsed annotations")
			}
		})
	}
}

func TestParseAnnotation(t *testing.T) {
	type testCase = struct {
		name        string
		annotations map[string]string
		devices     []string
		invalid     bool
	}

	for _, tc := range []*testCase{
		{
			name: "one vendor, one device",
			annotations: map[string]string{
				AnnotationPrefix + "vendor.class_device": "vendor.com/class=device1",
			},
			devices: []string{
				"vendor.com/class=device1",
			},
		},
		{
			name: "one vendor, multiple devices",
			annotations: map[string]string{
				AnnotationPrefix + "vendor.class_device": "vendor.com/class=device1,vendor.com/class=device2,vendor.com/class=device3",
			},
			devices: []string{
				"vendor.com/class=device1",
				"vendor.com/class=device2",
				"vendor.com/class=device3",
			},
		},

		{
			name: "one plugin, one device",
			annotations: map[string]string{
				AnnotationPrefix + "vendor.class_device": "vendor.com/class=device",
			},
			devices: []string{
				"vendor.com/class=device",
			},
		},
		{
			name: "one plugin, multiple devices",
			annotations: map[string]string{
				AnnotationPrefix + "vendor.class_device": "vendor.com/class=device1,vendor.com/class=device2,vendor.com/class=device3",
			},
			devices: []string{
				"vendor.com/class=device1",
				"vendor.com/class=device2",
				"vendor.com/class=device3",
			},
		},
		{
			name: "multiple plugins, multiple devices",
			annotations: map[string]string{
				AnnotationPrefix + "vendor1.class_device":  "vendor1.com/class=device1",
				AnnotationPrefix + "vendor2.class_device":  "vendor2.com/class=device1,vendor2.com/class=device2",
				AnnotationPrefix + "vendor3.class2_device": "vendor3.com/class2=device1,vendor3.com/class2=device2,vendor3.com/class2=device3",
			},
			devices: []string{
				"vendor1.com/class=device1",
				"vendor2.com/class=device1",
				"vendor2.com/class=device2",
				"vendor3.com/class2=device1",
				"vendor3.com/class2=device2",
				"vendor3.com/class2=device3",
			},
		},
		{
			name: "invalid, malformed device reference",
			annotations: map[string]string{
				AnnotationPrefix + "vendor1.class1_device": "vendor1.com/class1=device1",
				AnnotationPrefix + "vendor2.class2_device": "vendor2.com/class2=device2",
				AnnotationPrefix + "vendor3.class_device":  "vendor3.com=device3",
			},
			invalid: true,
		},
		{
			name: "invalid, pre-resolved device",
			annotations: map[string]string{
				AnnotationPrefix + "vendor1.class2_device": "vendor1.com/class2=device1,vendor1.com/class2=device2",
				AnnotationPrefix + "vendor2.class_device":  "/dev/null",
			},
			invalid: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, devices, err := ParseAnnotations(tc.annotations)
			if !tc.invalid {
				require.NoError(t, err, "parsing annotations")
				sort.Strings(tc.devices)
				sort.Strings(devices)
				require.Equal(t, tc.devices, devices, "parsing annotations")
			} else {
				require.Error(t, err)
			}
		})
	}
}
