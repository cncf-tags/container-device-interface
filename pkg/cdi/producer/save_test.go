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

package producer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

func TestSave(t *testing.T) {
	testCases := []struct {
		description   string
		spec          *cdi.Spec
		filename      string
		options       []Option
		expectedError string
		assert        func(*testing.T, string)
	}{
		{
			description:   "empty filename returns directory error",
			filename:      "",
			expectedError: "specified path is a directory",
		},
		{
			description: "spec is written with default permissions",
			spec: &cdi.Spec{
				Version: "v1.1.0",
			},
			filename: "test.yaml",
			assert: func(t *testing.T, fullpath string) {
				require.FileExists(t, fullpath)
				info, err := os.Stat(fullpath)
				require.NoError(t, err)
				require.EqualValues(t, os.FileMode(0644), info.Mode().Perm())
				contents, err := os.ReadFile(fullpath)
				require.NoError(t, err)
				expectedContents := `---
cdiVersion: v1.1.0
kind: ""
devices: []
`
				require.EqualValues(t, expectedContents, string(contents))
			},
		},
		{
			description: "spec is written as json with default permissions",
			spec: &cdi.Spec{
				Version: "v1.1.0",
			},
			filename: "test.json",
			assert: func(t *testing.T, fullpath string) {
				require.FileExists(t, fullpath)
				info, err := os.Stat(fullpath)
				require.NoError(t, err)
				require.EqualValues(t, os.FileMode(0644), info.Mode().Perm())
				contents, err := os.ReadFile(fullpath)
				require.NoError(t, err)
				expectedContents := `{"cdiVersion":"v1.1.0","kind":"","devices":null,"containerEdits":{}}`
				require.EqualValues(t, expectedContents, string(contents))
			},
		},
		{
			description: "spec is written with format specified as json",
			spec: &cdi.Spec{
				Version: "v1.1.0",
			},
			filename: "test",
			options:  []Option{WithOutputFormat("json")},
			assert: func(t *testing.T, fullpath string) {
				require.FileExists(t, fullpath)
				info, err := os.Stat(fullpath)
				require.NoError(t, err)
				require.EqualValues(t, os.FileMode(0644), info.Mode().Perm())
				contents, err := os.ReadFile(fullpath)
				require.NoError(t, err)
				expectedContents := `{"cdiVersion":"v1.1.0","kind":"","devices":null,"containerEdits":{}}`
				require.EqualValues(t, expectedContents, string(contents))
			},
		},
		{
			description: "spec is written with format specified as yaml",
			spec: &cdi.Spec{
				Version: "v1.1.0",
			},
			filename: "test",
			options:  []Option{WithOutputFormat("yaml")},
			assert: func(t *testing.T, fullpath string) {
				require.FileExists(t, fullpath)
				info, err := os.Stat(fullpath)
				require.NoError(t, err)
				require.EqualValues(t, os.FileMode(0644), info.Mode().Perm())
				contents, err := os.ReadFile(fullpath)
				require.NoError(t, err)
				expectedContents := `---
cdiVersion: v1.1.0
kind: ""
devices: []
`
				require.EqualValues(t, expectedContents, string(contents))
			},
		},
		{
			description: "spec is written with specified permissions",
			spec: &cdi.Spec{
				Version: "v1.1.0",
			},
			filename: "test.yaml",
			options: []Option{
				WithPermissions(os.FileMode(0666)),
			},
			assert: func(t *testing.T, fullpath string) {
				require.FileExists(t, fullpath)
				info, err := os.Stat(fullpath)
				require.NoError(t, err)
				require.EqualValues(t, os.FileMode(0666), info.Mode().Perm())
				contents, err := os.ReadFile(fullpath)
				require.NoError(t, err)
				expectedContents := `---
cdiVersion: v1.1.0
kind: ""
devices: []
`
				require.EqualValues(t, expectedContents, string(contents))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dir := t.TempDir()

			fullpath := filepath.Join(dir, tc.filename)
			err := Save(tc.spec, fullpath, tc.options...)

			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedError)
			}
			if tc.assert != nil {
				tc.assert(t, fullpath)
			}
		})
	}

}
