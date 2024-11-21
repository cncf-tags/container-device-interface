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

package validator

import (
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// A disabledValidator performs no validation.
type disabledValidator string

// Validate always passes for a disabledValidator.
func (v disabledValidator) Validate(*cdi.Spec) error {
	return nil
}
