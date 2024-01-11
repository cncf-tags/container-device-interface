/*
   Copyright © 2022 The CDI Authors

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

package schema_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/schema"
)

var (
	unloadable = map[string]bool{
		"empty.json": true,
	}

	none = schema.NopSchema()
)

func TestLoad(t *testing.T) {
	type testCase struct {
		testName   string
		schemaName string
	}
	for _, tc := range []*testCase{
		{
			testName:   "builtin schema",
			schemaName: "builtin",
		},
		{
			testName:   "externally loaded schema.json",
			schemaName: "file://./schema.json",
		},
		{
			testName:   "disabled/none schema",
			schemaName: "none",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			scm, err := schema.Load(tc.schemaName)
			require.NoError(t, err)
			require.NotNil(t, scm)
		})
	}
}

func TestValidateFile(t *testing.T) {
	type testCase struct {
		testName   string
		schemaName string
	}
	for _, tc := range []*testCase{
		{
			testName:   "builtin schema",
			schemaName: "builtin",
		},
		{
			testName:   "externally loaded schema.json",
			schemaName: "file://./schema.json",
		},
		{
			testName:   "disabled/none schema",
			schemaName: "none",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			scm := loadSchema(t, tc.schemaName)

			scanAndValidate(t, scm, "./testdata/good", true, validateFile)
			scanAndValidate(t, scm, "./testdata/bad", false, validateFile)
			old := schema.Get()
			schema.Set(scm)
			scanAndValidate(t, nil, "./testdata/good", true, validateFile)
			scanAndValidate(t, nil, "./testdata/bad", false, validateFile)
			schema.Set(old)
		})
	}
}

func TestValidateData(t *testing.T) {
	type testCase struct {
		testName   string
		schemaName string
	}
	for _, tc := range []*testCase{
		{
			testName:   "builtin schema",
			schemaName: "builtin",
		},
		{
			testName:   "externally loaded schema.json",
			schemaName: "file://./schema.json",
		},
		{
			testName:   "disabled/none schema",
			schemaName: "none",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			scm := loadSchema(t, tc.schemaName)

			scanAndValidate(t, scm, "./testdata/good", true, validateData)
			scanAndValidate(t, scm, "./testdata/bad", false, validateData)
			old := schema.Get()
			schema.Set(scm)
			scanAndValidate(t, nil, "./testdata/good", true, validateData)
			scanAndValidate(t, nil, "./testdata/bad", false, validateData)
			schema.Set(old)
		})
	}
}

func TestValidateReader(t *testing.T) {
	type testCase struct {
		testName   string
		schemaName string
	}
	for _, tc := range []*testCase{
		{
			testName:   "builtin schema",
			schemaName: "builtin",
		},
		{
			testName:   "externally loaded schema.json",
			schemaName: "file://./schema.json",
		},
		{
			testName:   "disabled/none schema",
			schemaName: "none",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			scm := loadSchema(t, tc.schemaName)

			scanAndValidate(t, scm, "./testdata/good", true, validateRead)
			scanAndValidate(t, scm, "./testdata/bad", false, validateRead)
			old := schema.Get()
			schema.Set(scm)
			scanAndValidate(t, nil, "./testdata/good", true, validateRead)
			scanAndValidate(t, nil, "./testdata/bad", false, validateRead)
			schema.Set(old)
		})
	}
}

func TestValidateReadAndValidate(t *testing.T) {
	type testCase struct {
		testName   string
		schemaName string
	}
	for _, tc := range []*testCase{
		{
			testName:   "builtin schema",
			schemaName: "builtin",
		},
		{
			testName:   "externally loaded schema.json",
			schemaName: "file://./schema.json",
		},
		{
			testName:   "disabled/none schema",
			schemaName: "none",
		},
	} {

		t.Run(tc.testName, func(t *testing.T) {
			scm := loadSchema(t, tc.schemaName)

			scanAndValidate(t, scm, "./testdata/good", true, readAndValidate)
			scanAndValidate(t, scm, "./testdata/bad", false, readAndValidate)
			old := schema.Get()
			schema.Set(scm)
			scanAndValidate(t, nil, "./testdata/good", true, readAndValidate)
			scanAndValidate(t, nil, "./testdata/bad", false, readAndValidate)
			schema.Set(old)
		})
	}
}

func TestValidateSpec(t *testing.T) {
	type testCase struct {
		testName   string
		schemaName string
	}
	for _, tc := range []*testCase{
		{
			testName:   "builtin schema",
			schemaName: "builtin",
		},
		{
			testName:   "externally loaded schema.json",
			schemaName: "file://./schema.json",
		},
		{
			testName:   "disabled/none schema",
			schemaName: "none",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			scm := loadSchema(t, tc.schemaName)

			scanAndValidate(t, scm, "./testdata/good", true, validateSpec)
			scanAndValidate(t, scm, "./testdata/bad", false, validateSpec)
			old := schema.Get()
			schema.Set(scm)
			scanAndValidate(t, nil, "./testdata/good", true, validateSpec)
			scanAndValidate(t, nil, "./testdata/bad", false, validateSpec)
			schema.Set(old)
		})
	}
}

func scanAndValidate(t *testing.T, scm *schema.Schema, dir string, isValid bool,
	validateFn func(t *testing.T, scm *schema.Schema, path string, shouldLoad, isValid bool)) {
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path == dir {
				return nil
			}
			scanAndValidate(t, scm, path, isValid, validateFn)
		} else {
			if name := info.Name(); filepath.Ext(name) != ".json" || name == "empty.json" {
				return nil
			}
			//fmt.Printf("*** processing %s...\n", path)
			validateFn(t, scm, path, !unloadable[filepath.Base(path)], isValid)
		}

		return nil
	})
	require.NoError(t, err)
}

func validateFile(t *testing.T, scm *schema.Schema, path string, shouldLoad, isValid bool) {
	var err error

	if scm != nil {
		err = scm.ValidateFile(path)
	} else {
		err = schema.ValidateFile(path)
	}

	verifyResult(t, scm, err, shouldLoad, isValid)
}

func validateData(t *testing.T, scm *schema.Schema, path string, shouldLoad, isValid bool) {
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	if scm != nil {
		err = scm.ValidateData(data)
	} else {
		err = schema.ValidateData(data)
	}

	verifyResult(t, scm, err, shouldLoad, isValid)
}

func readAndValidate(t *testing.T, scm *schema.Schema, path string, shouldLoad, isValid bool) {
	var (
		data []byte
	)

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	if scm != nil {
		data, err = scm.ReadAndValidate(f)
	} else {
		data, err = schema.ReadAndValidate(f)
	}

	verifyResult(t, scm, err, shouldLoad, isValid)

	if scm != nil {
		err = scm.Validate(bytes.NewReader(data))
	} else {
		err = schema.Validate(bytes.NewReader(data))
	}

	verifyResult(t, scm, err, shouldLoad, isValid)
}

func validateRead(t *testing.T, scm *schema.Schema, path string, shouldLoad, isValid bool) {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	buf := &bytes.Buffer{}
	r := io.TeeReader(f, buf)

	if scm != nil {
		err = scm.Validate(r)
	} else {
		err = schema.Validate(r)
	}

	verifyResult(t, scm, err, shouldLoad, isValid)

	if scm != nil {
		err = scm.Validate(bytes.NewReader(buf.Bytes()))
	} else {
		err = schema.Validate(bytes.NewReader(buf.Bytes()))
	}

	verifyResult(t, scm, err, shouldLoad, isValid)
}

func validateSpec(t *testing.T, scm *schema.Schema, path string, shouldLoad, isValid bool) {
	var old *schema.Schema

	if scm != nil {
		old = schema.Get()
		schema.Set(scm)
	}
	spec, err := cdi.ReadSpec(path, 0)
	if scm != nil {
		schema.Set(old)
	}

	if !shouldLoad || !isValid {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
		require.NotNil(t, spec)
	}
}

func loadSchema(t *testing.T, name string) *schema.Schema {
	if name == schema.NoneSchemaName {
		return none
	}

	scm, err := schema.Load(name)
	require.NoError(t, err)
	require.NotNil(t, scm)
	return scm
}

func verifyResult(t *testing.T, s *schema.Schema, err error, shouldLoad, isValid bool) {
	if s == none || (s == nil && schema.Get() == none) {
		require.NoError(t, err)
		return
	}

	if !isValid || !shouldLoad {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}
