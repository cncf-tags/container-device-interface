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
	"strings"
)

// ValidateQualifiedName checks the validity of a fully-qualified device name.
func ValidateQualifiedName(name string) error {
	_, _, _, err := ParseFullyQualifiedName(name)
	return err
}

// ValidateKind checks the validity of a CDI kind.
// The syntax for a device kind“ is
//
//	"<vendor>/<class>"
func ValidateKind(kind string) error {
	parts := strings.SplitN(kind, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("kind %s does not contain a / %w", kind, errInvalid)
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
	if !isLetter(rune(name[0])) {
		return fmt.Errorf("%q, should start with letter", name)
	}
	for _, c := range string(name[1 : len(name)-1]) {
		switch {
		case isAlphaNumeric(c):
		case c == '_' || c == '-' || c == '.':
		default:
			return fmt.Errorf("invalid character '%c' in name %q",
				c, name)
		}
	}
	if !isAlphaNumeric(rune(name[len(name)-1])) {
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
	if !isAlphaNumeric(rune(name[0])) {
		return fmt.Errorf("invalid class %q, should start with a letter or digit", name)
	}
	if len(name) == 1 {
		return nil
	}
	for _, c := range string(name[1 : len(name)-1]) {
		switch {
		case isAlphaNumeric(c):
		case c == '_' || c == '-' || c == '.' || c == ':':
		default:
			return fmt.Errorf("invalid character '%c' in device name %q",
				c, name)
		}
	}
	if !isAlphaNumeric(rune(name[len(name)-1])) {
		return fmt.Errorf("invalid name %q, should end with a letter or digit", name)
	}
	return nil
}

// ParseFullyQualifiedName splits a fully-qualified name into device vendor, class,
// and name. If the device fails to parse as a qualified name, or if any
// of the split components fail to pass syntax validation, vendor and
// class are returned as empty, together with the verbatim input as the
// name and an error describing the reason for failure.
func ParseFullyQualifiedName(device string) (string, string, string, error) {
	vendor, class, name := parseDevice(device)

	if vendor == "" {
		return "", "", device, fmt.Errorf("unqualified device %q, missing vendor", device)
	}
	if class == "" {
		return "", "", device, fmt.Errorf("unqualified device %q, missing class", device)
	}
	if name == "" {
		return "", "", device, fmt.Errorf("unqualified device %q, missing device name", device)
	}

	if err := ValidateVendorName(vendor); err != nil {
		return "", "", device, fmt.Errorf("invalid device %q: %w", device, err)
	}
	if err := ValidateClassName(class); err != nil {
		return "", "", device, fmt.Errorf("invalid device %q: %w", device, err)
	}
	if err := ValidateDeviceName(name); err != nil {
		return "", "", device, fmt.Errorf("invalid device %q: %w", device, err)
	}

	return vendor, class, name, nil
}

// parseDevice tries to split a device name into vendor, class, and name.
// If this fails, for instance in the case of unqualified device names,
// parseDevice returns an empty vendor and class together with name set
// to the verbatim input.
func parseDevice(device string) (string, string, string) {
	if device == "" || device[0] == '/' {
		return "", "", device
	}

	parts := strings.SplitN(device, "=", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", device
	}

	name := parts[1]
	vendor, class := ParseKind(parts[0])
	if vendor == "" {
		return "", "", device
	}

	return vendor, class, name
}

// ParseKind splits a device qualifier into vendor and class.
// The syntax for a device qualifier is
//
//	"<vendor>/<class>"
//
// If parsing fails, an empty vendor and the class set to the
// verbatim input is returned.
func ParseKind(kind string) (string, string) {
	parts := strings.SplitN(kind, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", kind
	}
	return parts[0], parts[1]
}

// isLetter reports whether the rune is a letter.
func isLetter(c rune) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}

// IsDigit reports whether the rune is a digit.
func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

// isAlphaNumeric reports whether the rune is a letter or digit.
func isAlphaNumeric(c rune) bool {
	return isLetter(c) || isDigit(c)
}
