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
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

func TestDeviceValidate(t *testing.T) {
	type testCase struct {
		name    string
		device  *Device
		invalid bool
	}
	for _, tc := range []*testCase{
		{
			name: "valid name, valid edits",
			device: &Device{
				Device: &cdi.Device{
					Name: "dev",
					ContainerEdits: cdi.ContainerEdits{
						Env: []string{"FOO=BAR"},
					},
				},
			},
		},
		{
			name: "valid name, invalid edits",
			device: &Device{
				Device: &cdi.Device{
					Name: "dev",
					ContainerEdits: cdi.ContainerEdits{
						Env: []string{"=BAR"},
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid name, valid edits",
			device: &Device{
				Device: &cdi.Device{
					Name: "a dev ice",
					ContainerEdits: cdi.ContainerEdits{
						Env: []string{"FOO=BAR"},
					},
				},
			},
			invalid: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.device.validate()
			if tc.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
