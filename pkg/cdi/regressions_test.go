package cdi

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCDIInjectionRace(t *testing.T) {
	// This is a gutted version of a containerd test case which triggered
	// read/write data race in the Cache.

	for _, test := range []struct {
		description   string
		cdiSpecFiles  []string
		annotations   map[string]string
		expectError   bool
		expectDevices []oci.LinuxDevice
		expectEnv     []string
	}{
		{description: "expect no CDI error for nil annotations"},
		{description: "expect no CDI error for empty annotations",
			annotations: map[string]string{},
		},
		{description: "expect CDI error for invalid CDI device reference in annotations",
			annotations: map[string]string{
				AnnotationPrefix + "devices": "foobar",
			},
			expectError: true,
		},
		{description: "expect CDI error for unresolvable devices",
			annotations: map[string]string{
				AnnotationPrefix + "vendor1_devices": "vendor1.com/device=no-such-dev",
			},
			expectError: true,
		},
		{description: "expect properly injected resolvable CDI devices",
			cdiSpecFiles: []string{
				`
cdiVersion: "0.3.0"
kind: "vendor1.com/device"
devices:
  - name: foo
    containerEdits:
      deviceNodes:
        - path: /dev/loop8
          type: b
          major: 7
          minor: 8
      env:
        - FOO=injected
containerEdits:
  env:
    - "VENDOR1=present"
`,
				`
cdiVersion: "0.3.0"
kind: "vendor2.com/device"
devices:
  - name: bar
    containerEdits:
      deviceNodes:
        - path: /dev/loop9
          type: b
          major: 7
          minor: 9
      env:
        - BAR=injected
containerEdits:
  env:
    - "VENDOR2=present"
`,
			},
			annotations: map[string]string{
				AnnotationPrefix + "vendor1_devices": "vendor1.com/device=foo",
				AnnotationPrefix + "vendor2_devices": "vendor2.com/device=bar",
			},
			expectDevices: []oci.LinuxDevice{
				{
					Path:  "/dev/loop8",
					Type:  "b",
					Major: 7,
					Minor: 8,
				},
				{
					Path:  "/dev/loop9",
					Type:  "b",
					Major: 7,
					Minor: 9,
				},
			},
			expectEnv: []string{
				"FOO=injected",
				"VENDOR1=present",
				"BAR=injected",
				"VENDOR2=present",
			},
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			var (
				err  error
				spec = &oci.Spec{}
			)

			cdiDir, err := writeFilesToTempDir("containerd-test-CDI-injections-", test.cdiSpecFiles)
			if cdiDir != "" {
				defer os.RemoveAll(cdiDir)
			}
			require.NoError(t, err)

			injectFun := withCDI(t, test.annotations, []string{cdiDir})
			err = injectFun(spec)
			assert.Equal(t, test.expectError, err != nil)

			if err != nil {
				if test.expectEnv != nil {
					for _, expectedEnv := range test.expectEnv {
						assert.Contains(t, spec.Process.Env, expectedEnv)
					}
				}
				if test.expectDevices != nil {
					for _, expectedDev := range test.expectDevices {
						assert.Contains(t, spec.Linux.Devices, expectedDev)
					}
				}
			}
		})
	}
}

type specOpts func(*oci.Spec) error

// withCDI (WithCDI) SpecOpt adopted from containerd.
func withCDI(t *testing.T, annotations map[string]string, cdiSpecDirs []string) specOpts {
	return func(s *oci.Spec) error {
		_, cdiDevices, err := ParseAnnotations(annotations)
		if err != nil {
			return fmt.Errorf("failed to parse CDI device annotations: %w", err)
		}
		if cdiDevices == nil {
			return nil
		}

		registry := GetRegistry(WithSpecDirs(cdiSpecDirs...))
		if err = registry.Refresh(); err != nil {
			t.Logf("CDI registry refresh failed: %v", err)
		}

		if _, err := registry.InjectDevices(s, cdiDevices...); err != nil {
			return fmt.Errorf("CDI device injection failed: %w", err)
		}

		return nil
	}
}

func writeFilesToTempDir(tmpDirPattern string, content []string) (string, error) {
	if len(content) == 0 {
		return "", nil
	}

	dir, err := os.MkdirTemp("", tmpDirPattern)
	if err != nil {
		return "", err
	}

	for idx, data := range content {
		file := filepath.Join(dir, fmt.Sprintf("spec-%d.yaml", idx))
		err := os.WriteFile(file, []byte(data), 0644)
		if err != nil {
			return "", err
		}
	}

	return dir, nil
}
