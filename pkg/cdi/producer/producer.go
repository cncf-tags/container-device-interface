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
	"fmt"
	"path/filepath"

	"tags.cncf.io/container-device-interface/pkg/cdi/producer/validator"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// A Producer defines a structure for outputting CDI specifications.
type Producer struct {
	format       specFormat
	failIfExists bool
	validator    specValidator
}

// New creates a new producer with the supplied options.
func New(opts ...Option) (*Producer, error) {
	p := &Producer{
		format:    DefaultSpecFormat,
		validator: validator.Default,
	}
	for _, opt := range opts {
		err := opt(p)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

// SaveSpec writes the specified CDI spec to a file with the specified name.
// If the filename ends in a supported extension, the format implied by the
// extension takes precedence over the format with which the Producer was
// configured.
func (p *Producer) SaveSpec(s *cdi.Spec, filename string) (string, error) {
	if err := p.Validate(s); err != nil {
		return "", fmt.Errorf("spec validation failed: %w", err)
	}

	filename = p.normalizeFilename(filename)
	sp := spec{
		Spec:   s,
		format: p.specFormatFromFilename(filename),
	}

	if err := sp.save(filename, !p.failIfExists); err != nil {
		return "", err
	}

	return filename, nil
}

// Validate performs a validation on a CDI spec using the configured validator.
// If no validator is configured, the spec is considered unconditionally valid.
func (p *Producer) Validate(s *cdi.Spec) error {
	if p == nil || p.validator == nil {
		return nil
	}
	return p.validator.Validate(s)
}

// specFormatFromFilename determines the CDI spec format for the given filename.
func (p *Producer) specFormatFromFilename(filename string) specFormat {
	switch filepath.Ext(filename) {
	case ".json":
		return SpecFormatJSON
	case ".yaml", ".yml":
		return SpecFormatYAML
	default:
		return p.format
	}
}

// normalizeFilename ensures that the specified filename ends in a supported extension.
func (p *Producer) normalizeFilename(filename string) string {
	switch filepath.Ext(filename) {
	case ".json":
		fallthrough
	case ".yaml", ".yml":
		return filename
	default:
		return filename + string(p.format)
	}
}
