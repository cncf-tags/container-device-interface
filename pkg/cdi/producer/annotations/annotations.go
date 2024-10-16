/*
   Copyright © 2021-2022 The CDI Authors

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

package annotations

import (
	"errors"
	"fmt"
	"strings"

	"tags.cncf.io/container-device-interface/pkg/parser"
)

const (
	// Prefix is the prefix for CDI container annotation keys.
	Prefix = "cdi.k8s.io/"
)

// Update updates annotations with a plugin-specific CDI device
// injection request for the given devices. Upon any error a non-nil error
// is returned and annotations are left intact. By convention plugin should
// be in the format of "vendor.device-type".
func Update(annotations map[string]string, plugin string, deviceID string, devices []string) (map[string]string, error) {
	key, err := Key(plugin, deviceID)
	if err != nil {
		return annotations, fmt.Errorf("CDI annotation failed: %w", err)
	}
	if _, ok := annotations[key]; ok {
		return annotations, fmt.Errorf("CDI annotation failed, key %q used", key)
	}
	value, err := Value(devices)
	if err != nil {
		return annotations, fmt.Errorf("CDI annotation failed: %w", err)
	}

	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[key] = value

	return annotations, nil
}

// Parse parses annotations for CDI device injection requests.
// The keys and devices from all such requests are collected into slices
// which are returned as the result. All devices are expected to be fully
// qualified CDI device names. If any device fails this check empty slices
// are returned along with a non-nil error. The annotations are expected
// to be formatted by, or in a compatible fashion to UpdateAnnotations().
func Parse(annotations map[string]string) ([]string, []string, error) {
	var (
		keys    []string
		devices []string
	)

	for key, value := range annotations {
		if !strings.HasPrefix(key, Prefix) {
			continue
		}
		for _, d := range strings.Split(value, ",") {
			if !parser.IsQualifiedName(d) {
				return nil, nil, fmt.Errorf("invalid CDI device name %q", d)
			}
			devices = append(devices, d)
		}
		keys = append(keys, key)
	}

	return keys, devices, nil
}

// Key returns a unique annotation key for an device allocation
// by a K8s device plugin. pluginName should be in the format of
// "vendor.device-type". deviceID is the ID of the device the plugin is
// allocating. It is used to make sure that the generated key is unique
// even if multiple allocations by a single plugin needs to be annotated.
func Key(pluginName, deviceID string) (string, error) {
	const maxNameLen = 63

	if pluginName == "" {
		return "", errors.New("invalid plugin name, empty")
	}
	if deviceID == "" {
		return "", errors.New("invalid deviceID, empty")
	}

	name := pluginName + "_" + strings.ReplaceAll(deviceID, "/", "_")

	if len(name) > maxNameLen {
		return "", fmt.Errorf("invalid plugin+deviceID %q, too long", name)
	}

	if c := rune(name[0]); !parser.IsAlphaNumeric(c) {
		return "", fmt.Errorf("invalid name %q, first '%c' should be alphanumeric",
			name, c)
	}
	if len(name) > 2 {
		for _, c := range name[1 : len(name)-1] {
			switch {
			case parser.IsAlphaNumeric(c):
			case c == '_' || c == '-' || c == '.':
			default:
				return "", fmt.Errorf("invalid name %q, invalid character '%c'",
					name, c)
			}
		}
	}
	if c := rune(name[len(name)-1]); !parser.IsAlphaNumeric(c) {
		return "", fmt.Errorf("invalid name %q, last '%c' should be alphanumeric",
			name, c)
	}

	return Prefix + name, nil
}

// Value returns an annotation value for the given devices.
func Value(devices []string) (string, error) {
	value, sep := "", ""
	for _, d := range devices {
		if _, _, _, err := parser.ParseQualifiedName(d); err != nil {
			return "", err
		}
		value += sep + d
		sep = ","
	}

	return value, nil
}
