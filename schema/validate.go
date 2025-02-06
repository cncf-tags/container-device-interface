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
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

const (
	// DefaultExternalSchema is the default JSON schema to load for validation.
	DefaultExternalSchema = "/etc/cdi/schema/schema.json"
)

// WithSchema returns a CDI Spec validator that uses the given Schema.
func WithSchema(s *Schema) func(*cdi.Spec) error {
	if s == nil {
		return func(*cdi.Spec) error {
			return nil
		}
	}
	return func(spec *cdi.Spec) error {
		return s.ValidateType(spec)
	}
}

// WithNamedSchema loads the named JSON schema and returns a CDI Spec
// validator for it. If loading the schema fails a dummy validator is
// returned.
func WithNamedSchema(name string) func(*cdi.Spec) error {
	s, _ := Load(name)
	return WithSchema(s)
}

// WithDefaultSchema returns a CDI Spec validator that uses the default
// external JSON schema, or the default builtin one if the external one
// fails to load.
func WithDefaultSchema() func(*cdi.Spec) error {
	s, err := Load(DefaultExternalSchema)
	if err == nil {
		return WithSchema(s)
	}
	return WithSchema(BuiltinSchema())
}
