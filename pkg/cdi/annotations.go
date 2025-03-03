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

package cdi

import (
	"fmt"
	"strings"

	"tags.cncf.io/container-device-interface/api/producer"
	"tags.cncf.io/container-device-interface/pkg/parser"
)

const (
	// AnnotationPrefix is the prefix for CDI container annotation keys.
	AnnotationPrefix = producer.AnnotationPrefix
)

// UpdateAnnotations updates annotations with a plugin-specific CDI device
// injection request for the given devices. Upon any error a non-nil error
// is returned and annotations are left intact. By convention plugin should
// be in the format of "vendor.device-type".
//
// Deprecated: use producer.UpdateAnnotations directly.
func UpdateAnnotations(annotations map[string]string, plugin string, deviceID string, devices []string) (map[string]string, error) {
	return producer.UpdateAnnotations(annotations, plugin, deviceID, devices)
}

// ParseAnnotations parses annotations for CDI device injection requests.
// The keys and devices from all such requests are collected into slices
// which are returned as the result. All devices are expected to be fully
// qualified CDI device names. If any device fails this check empty slices
// are returned along with a non-nil error. The annotations are expected
// to be formatted by, or in a compatible fashion to UpdateAnnotations().
func ParseAnnotations(annotations map[string]string) ([]string, []string, error) {
	var (
		keys    []string
		devices []string
	)

	for key, value := range annotations {
		if !strings.HasPrefix(key, AnnotationPrefix) {
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

// AnnotationKey returns a unique annotation key for an device allocation
// by a K8s device plugin. pluginName should be in the format of
// "vendor.device-type". deviceID is the ID of the device the plugin is
// allocating. It is used to make sure that the generated key is unique
// even if multiple allocations by a single plugin needs to be annotated.
//
// Deprecated: use producer.AnnotationKey directly.
func AnnotationKey(pluginName, deviceID string) (string, error) {
	return producer.AnnotationKey(pluginName, deviceID)
}

// AnnotationValue returns an annotation value for the given devices.
//
// Deprecated: use producer.AnnotationValue directly.
func AnnotationValue(devices []string) (string, error) {
	return producer.AnnotationValue(devices)
}
