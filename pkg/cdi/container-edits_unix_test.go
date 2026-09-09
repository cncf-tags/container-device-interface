//go:build !windows

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

package cdi

import (
	"os"
	"path/filepath"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// mockDevDir creates a directory with FIFOs acting as device nodes.
func mockDevDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for _, name := range []string{"dev0", "dev1", "dev2"} {
		require.NoError(t, unix.Mkfifo(filepath.Join(dir, name), 0o600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dev-regular"), nil, 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "dev-subdir"), 0o700))
	require.NoError(t, os.Symlink(filepath.Join(dir, "dev0"), filepath.Join(dir, "dev-link")))

	return dir
}

func TestValidateWildcardContainerEdits(t *testing.T) {
	for _, tc := range []struct {
		name    string
		edits   *cdi.ContainerEdits
		invalid bool
	}{
		{
			name: "valid wildcard device",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/dri/card*",
					},
					{
						Path:        "/dev/mei[0-9]",
						Permissions: "rw",
					},
				},
			},
		},
		{
			name: "invalid wildcard device, in a non-final path element",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/bus/usb/*/*",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, in a non-final path element only",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/dri/*/card0",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid wildcard device, with host path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:     "/dev/card*",
						HostPath: "/vendorroot/dev/card*",
					},
				},
			},
		},
		{
			name: "invalid wildcard device, malformed pattern",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/card[0-9",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, relative path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "dev/card*",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, uncleaned path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/../etc/host*",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, uncleaned host path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:     "/dev/host*",
						HostPath: "/vendorroot/../etc/host*",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, type set",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/card*",
						Type: "c",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, device numbers set",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:  "/dev/card*",
						Major: 226,
						Minor: 0,
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, wildcards only in host path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:     "/dev/card0",
						HostPath: "/vendorroot/dev/card*",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, mismatching patterns",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:     "/dev/gpu*",
						HostPath: "/vendorroot/dev/card*",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, wildcard directory only in host path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:     "/dev/card0",
						HostPath: "/vendorroot/*/card0",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid wildcard device, wildcard directory only in path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:     "/dev/*/card0",
						HostPath: "/vendorroot/dev/card0",
					},
				},
			},
			invalid: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edits := ContainerEdits{tc.edits}
			err := edits.Validate()
			if tc.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExpandWildcards(t *testing.T) {
	dir := mockDevDir(t)
	mode := os.FileMode(0o600)

	type result struct {
		path     string
		hostPath string
	}
	type testCase struct {
		name  string
		nodes []*cdi.DeviceNode
		want  []result
	}
	for _, tc := range []*testCase{
		{
			name: "no wildcards, returned as is",
			nodes: []*cdi.DeviceNode{
				{Path: "/dev/null"},
				{Path: "/dev/card0", HostPath: "/vendorroot/dev/card0"},
			},
			want: []result{
				{"/dev/null", ""},
				{"/dev/card0", "/vendorroot/dev/card0"},
			},
		},
		{
			name: "wildcard expanded to matching device nodes",
			nodes: []*cdi.DeviceNode{
				{Path: dir + "/dev*"},
			},
			want: []result{
				{dir + "/dev0", ""},
				{dir + "/dev1", ""},
				{dir + "/dev2", ""},
			},
		},
		{
			name: "wildcard with a separate host path",
			nodes: []*cdi.DeviceNode{
				{Path: "/dev/vendor/dev*", HostPath: dir + "/dev*"},
			},
			want: []result{
				{"/dev/vendor/dev0", dir + "/dev0"},
				{"/dev/vendor/dev1", dir + "/dev1"},
				{"/dev/vendor/dev2", dir + "/dev2"},
			},
		},
		{
			name: "wildcards matching nothing are dropped",
			nodes: []*cdi.DeviceNode{
				{Path: dir + "/nosuchdev*"},
				{Path: dir + "/nosuchdir/dev*"},
			},
			want: nil,
		},
		{
			name: "overlapping wildcards are deduplicated",
			nodes: []*cdi.DeviceNode{
				{Path: dir + "/dev[01]"},
				{Path: dir + "/dev*"},
			},
			want: []result{
				{dir + "/dev0", ""},
				{dir + "/dev1", ""},
				{dir + "/dev2", ""},
			},
		},
		{
			name: "wildcard does not duplicate an explicit device node",
			nodes: []*cdi.DeviceNode{
				{Path: dir + "/dev0"},
				{Path: dir + "/dev*"},
			},
			want: []result{
				{dir + "/dev0", ""},
				{dir + "/dev1", ""},
				{dir + "/dev2", ""},
			},
		},
		{
			name: "explicit device nodes are not deduplicated",
			nodes: []*cdi.DeviceNode{
				{Path: dir + "/dev*"},
				{Path: dir + "/dev0"},
			},
			want: []result{
				{dir + "/dev0", ""},
				{dir + "/dev1", ""},
				{dir + "/dev2", ""},
				{dir + "/dev0", ""},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expanded, err := expandWildcards(tc.nodes)
			require.NoError(t, err)

			got := []result(nil)
			for _, d := range expanded {
				got = append(got, result{d.Path, d.HostPath})
			}
			require.Equal(t, tc.want, got)
		})
	}

	t.Run("malformed pattern is an error", func(t *testing.T) {
		// Validate() normally catches this before we ever get here
		_, err := expandWildcards([]*cdi.DeviceNode{{Path: "/dev/card[0-9"}})
		require.Error(t, err)
	})

	t.Run("properties are inherited from host", func(t *testing.T) {
		uid, gid := uint32(1000), uint32(2000)
		pattern := &cdi.DeviceNode{
			Path:        dir + "/dev*",
			Permissions: "rw",
			FileMode:    &mode,
			UID:         &uid,
			GID:         &gid,
		}

		expanded, err := expandWildcards([]*cdi.DeviceNode{pattern})
		require.NoError(t, err)
		require.Len(t, expanded, 3)
		for _, d := range expanded {
			require.Equal(t, "rw", d.Permissions)
			require.Equal(t, &mode, d.FileMode)
			require.Equal(t, &uid, d.UID)
			require.Equal(t, &gid, d.GID)
		}
	})

	t.Run("fill missing info does not modify the pattern", func(t *testing.T) {
		pattern := &cdi.DeviceNode{Path: dir + "/dev*"}
		edits := &ContainerEdits{&cdi.ContainerEdits{DeviceNodes: []*cdi.DeviceNode{pattern}}}

		for range 2 {
			spec := &oci.Spec{}
			require.NoError(t, edits.Apply(spec))
			require.Len(t, spec.Linux.Devices, 3)

			require.Equal(t, dir+"/dev*", pattern.Path)
			require.Equal(t, "", pattern.HostPath)
			require.Equal(t, "", pattern.Type)
			require.Nil(t, pattern.FileMode)
		}
	})
}
