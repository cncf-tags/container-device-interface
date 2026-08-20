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
	"os"
	"path/filepath"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestResolveEditsReturnsCompleteDetachedEdits(t *testing.T) {
	dir, err := createSpecDirs(t, map[string]string{
		"vendor.yaml": `
cdiVersion: "1.1.0"
kind: vendor.example/gpu
containerEdits:
  env: [SPEC=1]
  mounts:
  - hostPath: /spec-host
    containerPath: /spec-container
    type: bind
    options: [rbind, ro]
  hooks:
  - hookName: prestart
    path: /bin/true
    args: [spec]
    timeout: 1
  netDevices:
  - hostInterfaceName: eth0
    name: eth0
  intelRdt:
    closID: spec
    l3CacheSchema: L3:0=ffff
    memBwSchema: MB:0=100
    schemata: [spec]
    enableMonitoring: true
  additionalGids: [0, 1]
devices:
- name: gpu0
  containerEdits:
    env: [GPU0=1]
    mounts:
    - hostPath: /gpu0-host
      containerPath: /gpu0-container
      type: tmpfs
      options: [rw, nosuid]
    hooks:
    - hookName: prestart
      path: /bin/true
      args: [gpu0]
      env: [HOOK=gpu0]
      timeout: 2
    netDevices:
    - hostInterfaceName: eth1
      name: eth1
    intelRdt:
      closID: gpu0
      l3CacheSchema: L3:0=ffff
      memBwSchema: MB:0=100
      schemata: [gpu0]
      enableMonitoring: true
    additionalGids: [2]
- name: gpu1
  containerEdits:
    env: [GPU1=1]
    mounts:
    - hostPath: /gpu1-host
      containerPath: /gpu1-container
      options: [ro]
    hooks:
    - hookName: prestart
      path: /bin/true
      args: [gpu1]
      timeout: 3
    netDevices:
    - hostInterfaceName: eth2
      name: eth2
    intelRdt:
      closID: gpu1
      schemata: [gpu1]
    additionalGids: [3]
`,
	}, nil)
	require.NoError(t, err)

	cache := newCache(
		WithAutoRefresh(false),
		WithSpecDirs(filepath.Join(dir, "etc")),
	)

	gpu0 := "vendor.example/gpu=gpu0"
	gpu1 := "vendor.example/gpu=gpu1"
	resolved, err := cache.ResolveEdits(gpu0, gpu1, gpu0)
	require.NoError(t, err)
	require.Equal(t, []string{"SPEC=1", "GPU0=1", "GPU1=1", "GPU0=1"}, resolved.Env)
	require.Len(t, resolved.DeviceNodes, 0)
	require.Len(t, resolved.Mounts, 4)
	require.Len(t, resolved.Hooks, 4)
	require.Len(t, resolved.NetDevices, 4)
	require.Equal(t, []uint32{0, 1, 2, 3, 2}, resolved.AdditionalGIDs)
	require.Equal(t, "gpu0", resolved.IntelRdt.ClosID)
	require.Equal(t, "L3:0=ffff", resolved.IntelRdt.L3CacheSchema)
	require.Equal(t, "MB:0=100", resolved.IntelRdt.MemBwSchema)
	require.True(t, resolved.IntelRdt.EnableMonitoring)
	require.Equal(t, []string{"gpu0"}, resolved.IntelRdt.Schemata)
	require.Equal(t, []string{"rbind", "ro"}, resolved.Mounts[0].Options)
	require.Equal(t, []string{"rw", "nosuid"}, resolved.Mounts[1].Options)
	require.Equal(t, []string{"spec", "gpu0", "gpu1", "gpu0"}, []string{
		resolved.Hooks[0].Args[0],
		resolved.Hooks[1].Args[0],
		resolved.Hooks[2].Args[0],
		resolved.Hooks[3].Args[0],
	})
	require.Equal(t, []string{"eth0", "eth1", "eth2", "eth1"}, []string{
		resolved.NetDevices[0].HostInterfaceName,
		resolved.NetDevices[1].HostInterfaceName,
		resolved.NetDevices[2].HostInterfaceName,
		resolved.NetDevices[3].HostInterfaceName,
	})

	resolved.Env[1] = "changed"
	resolved.Mounts[1].Options[0] = "changed"
	resolved.Hooks[1].Args[0] = "changed"
	resolved.Hooks[1].Env[0] = "changed"
	*resolved.Hooks[1].Timeout = 99
	resolved.NetDevices[1].Name = "changed"
	resolved.IntelRdt.Schemata[0] = "changed"
	resolved.AdditionalGIDs[2] = 99

	source := cache.devices[gpu0].ContainerEdits
	require.Equal(t, "GPU0=1", source.Env[0])
	require.Equal(t, "rw", source.Mounts[0].Options[0])
	require.Equal(t, "gpu0", source.Hooks[0].Args[0])
	require.Equal(t, "HOOK=gpu0", source.Hooks[0].Env[0])
	require.Equal(t, 2, *source.Hooks[0].Timeout)
	require.Equal(t, "eth1", source.NetDevices[0].Name)
	require.Equal(t, "gpu0", source.IntelRdt.Schemata[0])
	require.Equal(t, uint32(2), source.AdditionalGIDs[0])
}

func TestResolveEditsEmptyAndUnresolvable(t *testing.T) {
	cache := newCache(WithAutoRefresh(false))

	empty, err := cache.ResolveEdits()
	require.NoError(t, err)
	require.NotNil(t, empty.ContainerEdits)

	resolved, err := cache.ResolveEdits("vendor.example/gpu=missing0", "vendor.example/gpu=missing1")
	require.Nil(t, resolved)
	require.EqualError(t, err, "unresolvable CDI devices vendor.example/gpu=missing0, vendor.example/gpu=missing1")
}

func TestCollectContainerEditsReturnsAggregateInSelectionOrder(t *testing.T) {
	dir, err := createSpecDirs(t, map[string]string{
		"vendor.yaml": `
cdiVersion: "1.1.0"
kind: vendor.example/gpu
containerEdits:
  env: [SPEC=1]
devices:
- name: gpu0
  containerEdits:
    env: [GPU0=1]
- name: gpu1
  containerEdits:
    env: [GPU1=1]
`,
	}, nil)
	require.NoError(t, err)

	cache := newCache(
		WithAutoRefresh(false),
		WithSpecDirs(filepath.Join(dir, "etc")),
	)

	cache.Lock()
	defer cache.Unlock()
	selected, unresolved := cache.collectContainerEdits([]string{
		"vendor.example/gpu=gpu0",
		"vendor.example/gpu=gpu1",
		"vendor.example/gpu=gpu0",
	})
	require.Empty(t, unresolved)
	require.Equal(t, []string{"SPEC=1", "GPU0=1", "GPU1=1", "GPU0=1"}, selected.Env)
}

func TestResolveEditsPropagatesAutomaticRefreshError(t *testing.T) {
	dir, err := createSpecDirs(t, map[string]string{
		"vendor.yaml": `
cdiVersion: "0.3.0"
kind: vendor.example/gpu
devices:
- name: gpu0
`,
	}, nil)
	require.NoError(t, err)

	cache := newCache(WithSpecDirs(filepath.Join(dir, "etc")))
	cache.watch.stop()
	cache.watch = &watch{}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "etc", "vendor.yaml"), []byte("invalid"), 0600))

	resolved, err := cache.ResolveEdits()
	require.Nil(t, resolved)
	require.Error(t, err)

	_, err = cache.InjectDevices(&oci.Spec{}, "vendor.example/gpu=gpu0")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "failed to parse")
}

func TestInjectDevicesPreservesSpecIntelRdtAfterSelectionRefactor(t *testing.T) {
	dir, err := createSpecDirs(t, map[string]string{
		"vendor.yaml": `
cdiVersion: "0.7.0"
kind: vendor.example/gpu
containerEdits:
  intelRdt:
    closID: spec
devices:
- name: gpu0
  containerEdits:
    env: [GPU0=1]
`,
	}, nil)
	require.NoError(t, err)

	cache := newCache(
		WithAutoRefresh(false),
		WithSpecDirs(filepath.Join(dir, "etc")),
	)
	ociSpec := &oci.Spec{}
	unresolved, err := cache.InjectDevices(ociSpec, "vendor.example/gpu=gpu0")
	require.NoError(t, err)
	require.Empty(t, unresolved)
	require.Equal(t, "spec", ociSpec.Linux.IntelRdt.ClosID)
}
