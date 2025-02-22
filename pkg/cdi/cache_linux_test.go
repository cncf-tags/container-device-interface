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
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestTooManyOpenFiles(t *testing.T) {
	em, err := triggerEmfile()
	require.NoError(t, err)
	require.NotNil(t, em)
	defer func() {
		require.NoError(t, em.undo())
	}()

	_, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	require.Equal(t, syscall.EMFILE, err)

	cache := newCache(
		WithAutoRefresh(true),
	)
	require.NotNil(t, cache)

	// try to trigger original crash with a nil fsnotify.Watcher
	_, _ = cache.InjectDevices(&oci.Spec{}, "vendor1.com/device=dev1")
}

type emfile struct {
	limit  syscall.Rlimit
	fds    []int
	undone bool
}

// getFdTableSize reads the process' FD table size.
func getFdTableSize() (uint64, error) {
	status, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, err
	}

	const fdSizeTag = "FDSize:"

	for _, line := range strings.Split(string(status), "\n") {
		if strings.HasPrefix(line, fdSizeTag) {
			value := strings.TrimSpace(strings.TrimPrefix(line, fdSizeTag))
			size, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return 0, err
			}
			return size, nil
		}
	}

	return 0, fmt.Errorf("tag %s not found in /proc/self/status", fdSizeTag)
}

// triggerEmfile exhausts the file descriptors of the process and triggers
// a failure with an EMFILE for any syscall that needs to create a new fd.
// On success, the returned emfile's undo() method can be used to undo the
// exhausted table and restore everything to a working state.
func triggerEmfile() (*emfile, error) {
	// We exhaust our file descriptors by
	//   - checking the size of our current fd table
	//   - setting our soft RLIMIT_NOFILE limit to the table size
	//   - ensuring the fd table is full by creating new fd's
	//
	// We also save our original RLIMIT_NOFILE limit and any fd's we
	// might need to create, so we can eventually restore everything
	// to its original state.

	fdsize, err := getFdTableSize()
	if err != nil {
		return nil, err
	}

	em := &emfile{}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &em.limit); err != nil {
		return nil, err
	}

	limit := em.limit
	limit.Cur = fdsize

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return nil, err
	}

	for i := uint64(0); i < fdsize; i++ {
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
		if err != nil {
			return em, nil
		}
		em.fds = append(em.fds, fd)
	}

	return nil, fmt.Errorf("failed to trigger EMFILE")
}

// undo restores the process' state to its pre-EMFILE condition.
func (em *emfile) undo() error {
	if em == nil || em.undone {
		return nil
	}

	// we restore the process' state to pre-EMFILE condition by
	//   - restoring our saved RLIMIT_NOFILE
	//   - closing any extra file descriptors we might have created

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &em.limit); err != nil {
		return err
	}
	for _, fd := range em.fds {
		syscall.Close(fd)
	}
	em.undone = true

	return nil
}
