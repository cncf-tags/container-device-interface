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
	"fmt"
	"strings"
)

// ValidateKind checks the validity of a CDI kind.
// The syntax for a device kind“ is
//
//	"<vendor>/<class>"
func ValidateKind(kind string) error {
	parts := strings.SplitN(kind, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("kind %s does not contain a / %w", kind, ErrInvalid)
	}
	if err := ValidateVendorName(parts[0]); err != nil {
		return err
	}
	if err := ValidateClassName(parts[1]); err != nil {
		return err
	}
	return nil
}

// ValidateVendorName checks the validity of a vendor name.
// A vendor name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, and dot ('_', '-', and '.')
func ValidateVendorName(vendor string) error {
	err := validateVendorOrClassName(vendor)
	if err != nil {
		err = fmt.Errorf("invalid vendor. %w", err)
	}
	return err
}

// ValidateClassName checks the validity of class name.
// A class name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, and dot ('_', '-', and '.')
func ValidateClassName(class string) error {
	err := validateVendorOrClassName(class)
	if err != nil {
		err = fmt.Errorf("invalid class. %w", err)
	}
	return err
}

// validateVendorOrClassName checks the validity of vendor or class name.
// A name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, and dot ('_', '-', and '.')
func validateVendorOrClassName(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	if !IsLetter(rune(name[0])) {
		return fmt.Errorf("%q, should start with letter", name)
	}
	for _, c := range string(name[1 : len(name)-1]) {
		switch {
		case IsAlphaNumeric(c):
		case c == '_' || c == '-' || c == '.':
		default:
			return fmt.Errorf("invalid character '%c' in name %q",
				c, name)
		}
	}
	if !IsAlphaNumeric(rune(name[len(name)-1])) {
		return fmt.Errorf("%q, should end with a letter or digit", name)
	}

	return nil
}

// ValidateDeviceName checks the validity of a device name.
// A device name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, dot, colon ('_', '-', '.', ':')
func ValidateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid (empty) device name")
	}
	if !IsAlphaNumeric(rune(name[0])) {
		return fmt.Errorf("invalid class %q, should start with a letter or digit", name)
	}
	if len(name) == 1 {
		return nil
	}
	for _, c := range string(name[1 : len(name)-1]) {
		switch {
		case IsAlphaNumeric(c):
		case c == '_' || c == '-' || c == '.' || c == ':':
		default:
			return fmt.Errorf("invalid character '%c' in device name %q",
				c, name)
		}
	}
	if !IsAlphaNumeric(rune(name[len(name)-1])) {
		return fmt.Errorf("invalid name %q, should end with a letter or digit", name)
	}
	return nil
}

// IsLetter reports whether the rune is a letter.
func IsLetter(c rune) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}

// IsDigit reports whether the rune is a digit.
func IsDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

// IsAlphaNumeric reports whether the rune is a letter or digit.
func IsAlphaNumeric(c rune) bool {
	return IsLetter(c) || IsDigit(c)
}
