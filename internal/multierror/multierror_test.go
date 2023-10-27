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

package multierror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, nil, New())
	assert.Equal(t, nil, New(nil))
	assert.Equal(t, nil, New(nil, nil))
	assert.Equal(t, "hello\nworld", New(errors.New("hello"), errors.New("world")).Error())
}

func TestAppend(t *testing.T) {
	assert.Equal(t, nil, Append(nil))
	assert.Equal(t, nil, Append(nil, nil))
	assert.Equal(t, multiError{errors.New("hello"), errors.New("world"), errors.New("x"), errors.New("y")},
		Append(New(errors.New("hello"), errors.New("world")), New(errors.New("x"), nil, errors.New("y"))), nil)
}
