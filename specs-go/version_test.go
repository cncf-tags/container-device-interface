/*
   Copyright © The CDI Authors

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

package specs_test

import (
	"strings"
	"testing"

	"tags.cncf.io/container-device-interface/specs-go"
)

func TestMinimumRequiredVersion(t *testing.T) {
	tests := []struct {
		doc  string
		spec *specs.Spec
		want string
	}{
		{
			doc:  "no version-specific features",
			spec: &specs.Spec{},
			want: "0.3.0",
		},
		{
			doc: "v0.4.0 mount type",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					Mounts: []*specs.Mount{{Type: "bind"}},
				},
			},
			want: "0.4.0",
		},
		{
			doc: "v0.5.0 device node host path",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					DeviceNodes: []*specs.DeviceNode{{HostPath: "/dev/example"}},
				},
			},
			want: "0.5.0",
		},
		{
			doc: "v0.5.0 numeric device name",
			spec: &specs.Spec{
				Devices: []specs.Device{{Name: "0example"}},
			},
			want: "0.5.0",
		},
		{
			doc: "v0.6.0 annotation",
			spec: &specs.Spec{
				Annotations: map[string]string{"example.com/key": "value"},
			},
			want: "0.6.0",
		},
		{
			doc: "v0.6.0 dot in class",
			spec: &specs.Spec{
				Kind: "example.com/example.class",
			},
			want: "0.6.0",
		},
		{
			doc: "v0.7.0 additional GIDs",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					AdditionalGIDs: []uint32{1},
				},
			},
			want: "0.7.0",
		},
		{
			doc: "v0.7.0 Intel RDT",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					IntelRdt: &specs.IntelRdt{},
				},
			},
			want: "0.7.0",
		},
		{
			doc: "v1.1.0 Intel RDT schemata",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					IntelRdt: &specs.IntelRdt{Schemata: []string{"L3:0=ff"}},
				},
			},
			want: "1.1.0",
		},
		{
			doc: "v1.1.0 Intel RDT monitoring",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					IntelRdt: &specs.IntelRdt{EnableMonitoring: true},
				},
			},
			want: "1.1.0",
		},
		{
			doc: "v1.1.0 network device",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					NetDevices: []*specs.LinuxNetDevice{{}},
				},
			},
			want: "1.1.0",
		},
		{
			doc: "device-scoped feature",
			spec: &specs.Spec{
				Devices: []specs.Device{{
					ContainerEdits: specs.ContainerEdits{
						AdditionalGIDs: []uint32{1},
					},
				}},
			},
			want: "0.7.0",
		},
		{
			doc: "newest feature wins",
			spec: &specs.Spec{
				Annotations: map[string]string{"example.com/key": "value"},
				ContainerEdits: specs.ContainerEdits{
					AdditionalGIDs: []uint32{1},
					NetDevices:     []*specs.LinuxNetDevice{{}},
				},
			},
			want: "1.1.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.doc, func(t *testing.T) {
			got, err := specs.MinimumRequiredVersion(tc.spec)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("MinimumRequiredVersion() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		doc     string
		version string
		spec    *specs.Spec
		wantErr string
	}{
		{
			doc:     "current version",
			version: "1.1.0",
			spec:    &specs.Spec{},
		},
		{
			doc:     "optional v prefix",
			version: "v1.1.0",
			spec:    &specs.Spec{},
		},
		{
			doc:     "earliest supported version",
			version: "0.3.0",
			spec:    &specs.Spec{},
		},
		{
			doc:     "known but unsupported v0.1.0",
			version: "0.1.0",
			spec:    &specs.Spec{},
			wantErr: "the spec version must be at least v0.3.0",
		},
		{
			doc:     "known but unsupported v0.2.0",
			version: "v0.2.0",
			spec:    &specs.Spec{},
			wantErr: "the spec version must be at least v0.3.0",
		},
		{
			doc:     "unknown version",
			version: "1.2.0",
			spec:    &specs.Spec{},
			wantErr: `invalid version "1.2.0"`,
		},
		{
			doc:     "malformed version",
			version: "not-a-version",
			spec:    &specs.Spec{},
			wantErr: `invalid version "not-a-version"`,
		},
		{
			doc:     "version too old for feature",
			version: "0.6.0",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					AdditionalGIDs: []uint32{1},
				},
			},
			wantErr: "the spec version must be at least v0.7.0",
		},
		{
			doc:     "minimum version for feature",
			version: "0.7.0",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					AdditionalGIDs: []uint32{1},
				},
			},
		},
		{
			doc:     "newer version for feature",
			version: "1.1.0",
			spec: &specs.Spec{
				ContainerEdits: specs.ContainerEdits{
					AdditionalGIDs: []uint32{1},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.doc, func(t *testing.T) {
			tc.spec.Version = tc.version
			err := specs.ValidateVersion(tc.spec)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateVersion() error = nil, want %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateVersion() error = %q, want error containing %q", err, tc.wantErr)
			}
		})
	}
}
