/*
   Copyright © The CDI Authors

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

package specs_test

import (
	"testing"

	"tags.cncf.io/container-device-interface/specs-go"
)

func TestHasWildcards(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "", want: false},
		{path: "/dev/null", want: false},
		{path: "/dev/dri/card0", want: false},
		{path: "/dev/dri/card*", want: true},
		{path: "/dev/dri/renderD*", want: true},
		{path: "/dev/mei*", want: true},
		{path: "/dev/card?", want: true},
		{path: "/dev/card[0-9]", want: true},
		{path: "/dev/*/card0", want: true},
		// An escaped asterisk still makes the path a pattern.
		{path: `/dev/literal\*`, want: true},
		// A closing bracket alone is not a meta character.
		{path: "/dev/card]", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			if got := specs.HasWildcards(tc.path); got != tc.want {
				t.Errorf("HasWildcards(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
