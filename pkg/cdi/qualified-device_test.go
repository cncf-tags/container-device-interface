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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQualifiedName(t *testing.T) {
	type testCase = struct {
		device      string
		vendor      string
		class       string
		name        string
		isQualified bool
		isParsable  bool
	}

	for _, tc := range []*testCase{
		{
			device:      "vendor.com/class=dev",
			vendor:      "vendor.com",
			class:       "class",
			name:        "dev",
			isQualified: true,
		},
		{
			device:      "vendor.com/class=0",
			vendor:      "vendor.com",
			class:       "class",
			name:        "0",
			isQualified: true,
		},
		{
			device:      "vendor1.com/class1=dev1",
			vendor:      "vendor1.com",
			class:       "class1",
			name:        "dev1",
			isQualified: true,
		},
		{
			device:      "other-vendor1.com/class_1=dev_1",
			vendor:      "other-vendor1.com",
			class:       "class_1",
			name:        "dev_1",
			isQualified: true,
		},
		{
			device:      "yet_another-vendor2.com/c-lass_2=dev_1:2.3",
			vendor:      "yet_another-vendor2.com",
			class:       "c-lass_2",
			name:        "dev_1:2.3",
			isQualified: true,
		},
		{
			device:     "_invalid.com/class=dev",
			vendor:     "_invalid.com",
			class:      "class",
			name:       "dev",
			isParsable: true,
		},
		{
			device:     "invalid2.com-/class=dev",
			vendor:     "invalid2.com-",
			class:      "class",
			name:       "dev",
			isParsable: true,
		},
		{
			device:     "invalid3.com/_class=dev",
			vendor:     "invalid3.com",
			class:      "_class",
			name:       "dev",
			isParsable: true,
		},
		{
			device:     "invalid4.com/class_=dev",
			vendor:     "invalid4.com",
			class:      "class_",
			name:       "dev",
			isParsable: true,
		},
		{
			device:     "invalid5.com/class=-dev",
			vendor:     "invalid5.com",
			class:      "class",
			name:       "-dev",
			isParsable: true,
		},
		{
			device:     "invalid6.com/class=dev:",
			vendor:     "invalid6.com",
			class:      "class",
			name:       "dev:",
			isParsable: true,
		},
		{
			device:     "*.com/*dev=*gpu*",
			vendor:     "*.com",
			class:      "*dev",
			name:       "*gpu*",
			isParsable: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vendor, class, name, err := ParseQualifiedName(tc.device)
			if tc.isQualified {
				require.True(t, IsQualifiedName(tc.device), "qualified name %q", tc.device)
				require.NoError(t, err)
				require.Equal(t, tc.vendor, vendor, "qualified name %q", tc.device)
				require.Equal(t, tc.class, class, "qualified name %q", tc.device)
				require.Equal(t, tc.name, name, "qualified name %q", tc.device)

				vendor, class, name = ParseDevice(tc.device)
				require.Equal(t, tc.vendor, vendor, "parsed name %q", tc.device)
				require.Equal(t, tc.class, class, "parse name %q", tc.device)
				require.Equal(t, tc.name, name, "parsed name %q", tc.device)

				device := QualifiedName(vendor, class, name)
				require.Equal(t, tc.device, device, "constructed device %q", tc.device)
			} else if tc.isParsable {
				require.False(t, IsQualifiedName(tc.device), "parsed name %q", tc.device)
				vendor, class, name = ParseDevice(tc.device)
				require.Equal(t, tc.vendor, vendor, "parsed name %q", tc.device)
				require.Equal(t, tc.class, class, "parse name %q", tc.device)
				require.Equal(t, tc.name, name, "parsed name %q", tc.device)
			}
		})
	}
}
