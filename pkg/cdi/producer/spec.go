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
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"

	cdi "tags.cncf.io/container-device-interface/specs-go"
)

type spec struct {
	*cdi.Spec
	format specFormat
}

// save saves a CDI spec to the specified filename.
func (s *spec) save(filename string, overwrite bool) error {
	data, err := s.contents()
	if err != nil {
		return fmt.Errorf("failed to marshal Spec file: %w", err)
	}

	dir := filepath.Dir(filename)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create Spec dir: %w", err)
		}
	}

	tmp, err := os.CreateTemp(dir, "spec.*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create Spec file: %w", err)
	}
	_, err = tmp.Write(data)
	tmp.Close()
	if err != nil {
		return fmt.Errorf("failed to write Spec file: %w", err)
	}

	err = renameIn(dir, filepath.Base(tmp.Name()), filepath.Base(filename), overwrite)
	if err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("failed to write Spec file: %w", err)
	}
	return nil
}

// contents returns the raw contents of a CDI specification.
func (s *spec) contents() ([]byte, error) {
	switch s.format {
	case SpecFormatYAML:
		data, err := yaml.Marshal(s.Spec)
		if err != nil {
			return nil, err
		}
		data = append([]byte("---\n"), data...)
		return data, nil
	case SpecFormatJSON:
		return json.Marshal(s.Spec)
	default:
		return nil, fmt.Errorf("undefined CDI spec format %v", s.format)
	}
}
