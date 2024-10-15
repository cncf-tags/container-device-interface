/*
   Copyright © 2024 The CDI Authors

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
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"sigs.k8s.io/yaml"

	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// A SpecFormat defines the encoding to use when reading or writing a CDI specification.
type SpecFormat string

// A specFormatter wraps a raw CDI specification and allows it to be formatted
// using the specified options.
type specFormatter struct {
	*cdi.Spec
	options
}

// WriteTo writes the spec to the specified writer.
func (p *specFormatter) WriteTo(w io.Writer) (int64, error) {
	data, err := p.contents()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal Spec file: %w", err)
	}

	n, err := w.Write(data)
	return int64(n), err
}

// marshal returns the raw contents of a CDI specification.
// No validation is performed.
func (p SpecFormat) marshal(spec *cdi.Spec) ([]byte, error) {
	switch p {
	case SpecFormatYAML:
		data, err := yaml.Marshal(spec)
		if err != nil {
			return nil, err
		}
		data = append([]byte("---\n"), data...)
		return data, nil
	case SpecFormatJSON:
		return json.Marshal(spec)
	default:
		return nil, fmt.Errorf("undefined CDI spec format %v", p)
	}
}

// normalizeFilename ensures that the specified filename ends in a supported extension.
func (p SpecFormat) normalizeFilename(filename string) (string, SpecFormat) {
	switch filepath.Ext(filename) {
	case ".json":
		return filename, SpecFormatJSON
	case ".yaml":
		return filename, SpecFormatYAML
	default:
		return filename + string(p), p
	}
}

// validate performs an explicit validation of the spec.
// This is currently a placeholder for validation that should be performed when
// saving a spec.
func (p *specFormatter) validate() error {
	return nil
}

// contents returns the raw contents of a CDI specification.
// Validation is performed before marshalling the contentent based on the spec format.
func (p *specFormatter) contents() ([]byte, error) {
	if err := p.validate(); err != nil {
		return nil, fmt.Errorf("spec validation failed: %w", err)
	}
	return p.specFormat.marshal(p.Spec)
}
