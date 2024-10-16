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

	"tags.cncf.io/container-device-interface/pkg/cdi/producer/validator"
)

// An Option defines a functional option for constructing a producer.
type Option func(*Producer) error

// WithSpecFormat sets the output format of a CDI specification.
func WithSpecFormat(format specFormat) Option {
	return func(p *Producer) error {
		switch format {
		case SpecFormatJSON, SpecFormatYAML:
			p.format = format
		default:
			return fmt.Errorf("invalid CDI spec format %v", format)
		}
		return nil
	}
}

// WithSpecValidator sets a validator to be used when writing an output spec.
func WithSpecValidator(v specValidator) Option {
	return func(p *Producer) error {
		if v == nil {
			v = validator.Disabled
		}
		p.validator = v
		return nil
	}
}

// WithOverwrite specifies whether a producer should overwrite a CDI spec when
// saving to file.
func WithOverwrite(overwrite bool) Option {
	return func(p *Producer) error {
		p.failIfExists = !overwrite
		return nil
	}
}
