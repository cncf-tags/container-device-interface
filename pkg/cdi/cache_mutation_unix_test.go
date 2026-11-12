//go:build unix

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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// TestInjectDevicesDoesNotMutateCache verifies that applying container
// edits fills in missing device node information in the generated OCI Spec
// only, and never writes host-derived values back into the cached CDI Spec.
func TestInjectDevicesDoesNotMutateCache(t *testing.T) {
	hostDev := filepath.Join(t.TempDir(), "dev0")
	require.NoError(t, unix.Mkfifo(hostDev, 0o600))

	dir, err := createSpecDirs(t, map[string]string{
		"vendor1.yaml": fmt.Sprintf(`
cdiVersion: "0.3.0"
kind: "vendor1.com/device"
devices:
  - name: "dev0"
    containerEdits:
      deviceNodes:
      - path: %q
`, hostDev),
	}, nil)
	require.NoError(t, err)

	cache := newCache(WithSpecDirs(
		filepath.Join(dir, "etc"),
		filepath.Join(dir, "run"),
	))
	require.NotNil(t, cache)

	ociSpec := &oci.Spec{}
	unresolved, err := cache.InjectDevices(ociSpec, "vendor1.com/device=dev0")
	require.NoError(t, err)
	require.Nil(t, unresolved)

	// The OCI Spec got the information from the host device...
	require.Len(t, ociSpec.Linux.Devices, 1)
	require.Equal(t, hostDev, ociSpec.Linux.Devices[0].Path)
	require.Equal(t, fifoDevice, ociSpec.Linux.Devices[0].Type)
	require.NotNil(t, ociSpec.Linux.Devices[0].FileMode)

	// ...but the cached device node is left exactly as the Spec file had it.
	dev := cache.GetDevice("vendor1.com/device=dev0")
	require.NotNil(t, dev)
	require.Len(t, dev.ContainerEdits.DeviceNodes, 1)
	cached := dev.ContainerEdits.DeviceNodes[0]
	require.Equal(t, hostDev, cached.Path)
	require.Empty(t, cached.HostPath)
	require.Empty(t, cached.Type)
	require.Zero(t, cached.Major)
	require.Zero(t, cached.Minor)
	require.Nil(t, cached.FileMode)

	// A change on the host is noticed by the next injection, because the
	// device node is re-checked rather than served from mutated cache state.
	require.NoError(t, os.Remove(hostDev))
	require.NoError(t, os.WriteFile(hostDev, nil, 0o600))
	_, err = cache.InjectDevices(&oci.Spec{}, "vendor1.com/device=dev0")
	require.Error(t, err)
}
