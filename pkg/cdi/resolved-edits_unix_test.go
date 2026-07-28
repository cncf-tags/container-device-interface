//go:build !windows

/*
   Copyright © 2026 The CDI Authors

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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveEditsResolvesDeviceNodesWithoutMutatingCache(t *testing.T) {
	dir, err := createSpecDirs(t, map[string]string{
		"vendor.yaml": `
cdiVersion: "0.5.0"
kind: vendor.example/gpu
devices:
- name: gpu0
  containerEdits:
    deviceNodes:
    - path: /container/null
      hostPath: /dev/null
`,
	}, nil)
	require.NoError(t, err)

	cache := newCache(
		WithAutoRefresh(false),
		WithSpecDirs(filepath.Join(dir, "etc")),
	)
	resolved, err := cache.ResolveEdits("vendor.example/gpu=gpu0")
	require.NoError(t, err)
	require.Len(t, resolved.DeviceNodes, 1)

	info, err := deviceInfoFromPath("/dev/null")
	require.NoError(t, err)
	node := resolved.DeviceNodes[0]
	require.Equal(t, "/dev/null", node.HostPath)
	require.Equal(t, charDevice, node.Type)
	require.Equal(t, info.major, node.Major)
	require.Equal(t, info.minor, node.Minor)
	require.NotNil(t, node.FileMode)

	source := cache.devices["vendor.example/gpu=gpu0"].ContainerEdits.DeviceNodes[0]
	require.Empty(t, source.Type)
	require.Zero(t, source.Major)
	require.Zero(t, source.Minor)
	require.Nil(t, source.FileMode)

	node.Permissions = "r"
	require.Empty(t, source.Permissions)

	*node.FileMode = 0777
	require.Nil(t, source.FileMode)
}

func TestResolveEditsKeepsFullySpecifiedDeviceWithoutHostNode(t *testing.T) {
	dir, err := createSpecDirs(t, map[string]string{
		"vendor.yaml": `
cdiVersion: "0.5.0"
kind: vendor.example/gpu
devices:
- name: gpu0
  containerEdits:
    deviceNodes:
    - path: /container/synthetic
      hostPath: /does/not/exist
      type: c
      major: 123
      minor: 4
`,
	}, nil)
	require.NoError(t, err)

	cache := newCache(
		WithAutoRefresh(false),
		WithSpecDirs(filepath.Join(dir, "etc")),
	)
	resolved, err := cache.ResolveEdits("vendor.example/gpu=gpu0")
	require.NoError(t, err)
	require.Len(t, resolved.DeviceNodes, 1)
	require.Equal(t, "/does/not/exist", resolved.DeviceNodes[0].HostPath)
	require.Equal(t, charDevice, resolved.DeviceNodes[0].Type)
	require.Equal(t, int64(123), resolved.DeviceNodes[0].Major)
	require.Equal(t, int64(4), resolved.DeviceNodes[0].Minor)
}
