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

package specs

import "strings"

// ParseQualifier splits a device qualifier into vendor and class.
// The syntax for a device qualifier is
//
//	"<vendor>/<class>"
//
// If parsing fails, an empty vendor and the class set to the
// verbatim input is returned.
func ParseQualifier(kind string) (string, string) {
	parts := strings.SplitN(kind, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", kind
	}
	return parts[0], parts[1]
}

// IsLetter reports whether the rune is a letter.
func IsLetter(c rune) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}
