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

package schema

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"sigs.k8s.io/yaml"

	"tags.cncf.io/container-device-interface/internal/validation"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// currently loaded schema, builtin by default
var current atomic.Pointer[Schema]

func init() {
	current.Store(BuiltinSchema())
}

const (
	// BuiltinSchemaName names the builtin schema for Load()/Set().
	BuiltinSchemaName = "builtin"
	// NoneSchemaName names the NOP-schema for Load()/Set().
	NoneSchemaName = "none"
	// builtinSchemaFile is the builtin schema URI in our embedded FS.
	builtinSchemaFile = "file:///schema.json"
)

// Schema is a JSON validation schema.
type Schema struct {
	schema *jsonschema.Schema
}

// Validate applies a schema validation on the supplied CDI specification.
// If the Schema is nil, no validation is performed.
func (s *Schema) Validate(spec *cdi.Spec) error {
	if s == nil {
		return nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to load JSON data for validation: %w", err)
	}
	return s.validate(bytes.NewReader(data))
}

// Set sets the default validating JSON schema.
func Set(s *Schema) {
	if s == nil {
		s = NopSchema()
	}
	current.Store(s)
}

// Get returns the active validating JSON schema.
func Get() *Schema {
	return current.Load()
}

// BuiltinSchema returns the builtin schema if we have a valid one. Otherwise
// it falls back to NopSchema().
func BuiltinSchema() *Schema {
	return builtinSchema()
}

// builtinURLLoader loads schemas from the embedded builtin filesystem.
type builtinURLLoader struct{}

func (builtinURLLoader) Load(url string) (any, error) {
	f, err := builtinFS.Open(strings.TrimPrefix(url, "file:///"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return jsonschema.UnmarshalJSON(f)
}

var builtinSchema = sync.OnceValue(func() *Schema {
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{
		"file": builtinURLLoader{},
	})

	s, err := compiler.Compile(builtinSchemaFile)
	if err != nil {
		return NopSchema()
	}
	return &Schema{schema: s}
})

// NopSchema returns an validating JSON Schema that does no real validation.
func NopSchema() *Schema {
	return &Schema{}
}

// ReadAndValidate all data from the given reader, using the active schema for validation.
func ReadAndValidate(r io.Reader) ([]byte, error) {
	return Get().ReadAndValidate(r)
}

// ValidateReader validates the data read from an io.Reader against the active schema.
func ValidateReader(r io.Reader) error {
	return Get().ValidateReader(r)
}

// ValidateData validates the given JSON document against the active schema.
func ValidateData(data []byte) error {
	return Get().ValidateData(data)
}

// ValidateFile validates the given JSON file against the active schema.
func ValidateFile(path string) error {
	return Get().ValidateFile(path)
}

// ValidateType validates a go object against the schema.
func ValidateType(obj any) error {
	return Get().ValidateType(obj)
}

// Load the given JSON Schema.
func Load(source string) (*Schema, error) {
	source = strings.TrimSpace(source)

	switch {
	case source == BuiltinSchemaName:
		return BuiltinSchema(), nil
	case source == NoneSchemaName, source == "":
		return NopSchema(), nil
	case strings.HasPrefix(source, "http://"):
	case strings.HasPrefix(source, "https://"):
	default:
		if !strings.Contains(source, "://") || strings.HasPrefix(source, "file://") {
			var err error
			source, err = filepath.Abs(strings.TrimPrefix(source, "file://"))
			if err != nil {
				return nil, fmt.Errorf("failed to get JSON schema absolute path for %s: %w",
					source, err)
			}
			source = "file://" + source
		}
	}

	compiler := jsonschema.NewCompiler()
	httpLoader := httpURLLoader(http.Client{
		Timeout: 15 * time.Second,
	})
	compiler.UseLoader(jsonschema.SchemeURLLoader{
		"file":  jsonschema.FileLoader{},
		"http":  &httpLoader,
		"https": &httpLoader,
	})

	s, err := compiler.Compile(source)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON schema: %w", err)
	}

	return &Schema{schema: s}, nil
}

// ReadAndValidate all data from the given reader, using the schema for validation.
func (s *Schema) ReadAndValidate(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data for validation: %w", err)
	}
	return data, s.validate(bytes.NewReader(data))
}

// ValidateReader validates the data read from an io.Reader against the schema.
func (s *Schema) ValidateReader(r io.Reader) error {
	return s.validate(r)
}

// ValidateData validates the given JSON data against the schema.
func (s *Schema) ValidateData(data []byte) error {
	var schemaData map[string]any
	if !bytes.HasPrefix(bytes.TrimSpace(data), []byte{'{'}) {
		var err error
		err = yaml.Unmarshal(data, &schemaData)
		if err != nil {
			return fmt.Errorf("failed to YAML unmarshal data for validation: %w", err)
		}
		data, err = json.Marshal(schemaData)
		if err != nil {
			return fmt.Errorf("failed to JSON remarshal data for validation: %w", err)
		}
	}

	if err := s.validate(bytes.NewReader(data)); err != nil {
		return err
	}

	return s.validateContents(schemaData)
}

// ValidateFile validates the given JSON file against the schema.
func (s *Schema) ValidateFile(path string) error {
	if filepath.Ext(path) == ".json" {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return s.validate(f)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return s.ValidateData(data)
}

// ValidateType validates a go object against the schema.
func (s *Schema) ValidateType(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to load JSON data for validation: %w", err)
	}

	return s.validate(bytes.NewReader(data))
}

// Validate the given doc against the schema.
func (s *Schema) validate(r io.Reader) error {
	if s == nil || s.schema == nil {
		return nil
	}

	doc, err := jsonschema.UnmarshalJSON(r)
	if err != nil {
		return fmt.Errorf("failed to load JSON data for validation: %w", err)
	}
	return s.schema.Validate(doc)
}

type schemaContents map[string]any

func asSchemaContents(i any) (schemaContents, error) {
	if i == nil {
		return nil, nil
	}

	if c, ok := i.(map[string]any); ok {
		return schemaContents(c), nil
	}

	return nil, fmt.Errorf("expected map[string]any but got %T", i)
}

func (c schemaContents) getFieldAsString(key string) (string, bool) {
	if c == nil {
		return "", false
	}
	if v, ok := c[key]; ok {
		if value, ok := v.(string); ok {
			return value, true
		}
	}
	return "", false
}

func (c schemaContents) getAnnotations() (map[string]any, bool) {
	if c == nil {
		return nil, false
	}
	if v, ok := c["annotations"]; ok {
		if annotations, ok := v.(map[string]any); ok {
			return annotations, true
		}
	}
	return nil, false
}

func (c schemaContents) getDevices() ([]schemaContents, error) {
	if c == nil {
		return nil, nil
	}
	devicesIfc, ok := c["devices"]
	if !ok {
		return nil, nil
	}

	devices, ok := devicesIfc.([]any)
	if !ok {
		return nil, nil
	}

	var deviceContents []schemaContents
	for _, device := range devices {
		sc, err := asSchemaContents(device)
		if err != nil {
			return nil, fmt.Errorf("failed to parse device: %w", err)
		}
		deviceContents = append(deviceContents, sc)
	}

	return deviceContents, nil
}

// validateContents performs additional validation against the schema contents.
func (s *Schema) validateContents(data map[string]any) error {
	if data == nil || s == nil {
		return nil
	}

	contents := schemaContents(data)

	if specAnnotations, ok := contents.getAnnotations(); ok {
		if err := validation.ValidateSpecAnnotations("", specAnnotations); err != nil {
			return err
		}
	}

	devices, err := contents.getDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		name, _ := device.getFieldAsString("name")
		if annotations, ok := device.getAnnotations(); ok {
			if err := validation.ValidateSpecAnnotations(name, annotations); err != nil {
				return err
			}
		}
	}

	return nil
}

type httpURLLoader http.Client

func (l *httpURLLoader) Load(url string) (any, error) {
	resp, err := (*http.Client)(l).Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status code %d", url, resp.StatusCode)
	}

	return jsonschema.UnmarshalJSON(resp.Body)
}

//go:embed *.json
var builtinFS embed.FS
