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

package parser

import (
	"fmt"
	"strings"

	"tags.cncf.io/container-device-interface/api/validator"
)

// QualifiedName returns the qualified name for a device.
// The syntax for a qualified device names is
//
//	"<vendor>/<class>=<name>".
//
// A valid vendor and class name may contain the following runes:
//
//	'A'-'Z', 'a'-'z', '0'-'9', '.', '-', '_'.
//
// A valid device name may contain the following runes:
//
//	'A'-'Z', 'a'-'z', '0'-'9', '-', '_', '.', ':'
func QualifiedName(vendor, class, name string) string {
	return vendor + "/" + class + "=" + name
}

// IsQualifiedName tests if a device name is qualified.
func IsQualifiedName(device string) bool {
	_, _, _, err := ParseQualifiedName(device)
	return err == nil
}

// ParseQualifiedName splits a qualified name into device vendor, class,
// and name. If the device fails to parse as a qualified name, or if any
// of the split components fail to pass syntax validation, vendor and
// class are returned as empty, together with the verbatim input as the
// name and an error describing the reason for failure.
func ParseQualifiedName(device string) (string, string, string, error) {
	vendor, class, name := ParseDevice(device)

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

// ParseDevice tries to split a device name into vendor, class, and name.
// If this fails, for instance in the case of unqualified device names,
// ParseDevice returns an empty vendor and class together with name set
// to the verbatim input.
func ParseDevice(device string) (string, string, string) {
	if device == "" || device[0] == '/' {
		return "", "", device
	}

	parts := strings.SplitN(device, "=", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", device
	}

	name := parts[1]
	vendor, class := ParseQualifier(parts[0])
	if vendor == "" {
		return "", "", device
	}

	return vendor, class, name
}

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

// ValidateVendorName checks the validity of a vendor name.
// A vendor name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, and dot ('_', '-', and '.')
//
// Deprecated: use validator.ValidateVendorName instead
func ValidateVendorName(vendor string) error {
	return validator.ValidateVendorName(vendor)
}

// ValidateClassName checks the validity of class name.
// A class name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, and dot ('_', '-', and '.')
//
// Deprecated: use validator.ValidateClassName instead
func ValidateClassName(class string) error {
	return validator.ValidateClassName(class)
}

// ValidateDeviceName checks the validity of a device name.
// A device name may contain the following ASCII characters:
//   - upper- and lowercase letters ('A'-'Z', 'a'-'z')
//   - digits ('0'-'9')
//   - underscore, dash, dot, colon ('_', '-', '.', ':')
//
// Deprecated: use validator.ValidateDeviceName instead
func ValidateDeviceName(name string) error {
	return validator.ValidateDeviceName(name)
}

// IsLetter reports whether the rune is a letter.
//
// Deprecated: use validator.IsLetter instead
func IsLetter(c rune) bool {
	return validator.IsLetter(c)
}

// IsDigit reports whether the rune is a digit.
//
// Deprecated: use validator.IsDigit instead
func IsDigit(c rune) bool {
	return validator.IsDigit(c)
}

// IsAlphaNumeric reports whether the rune is a letter or digit.
//
// Deprecated: use validator.IsAlphaNumeric instead
func IsAlphaNumeric(c rune) bool {
	return validator.IsAlphaNumeric(c)
}
