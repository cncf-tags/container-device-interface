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
	"tags.cncf.io/container-device-interface/pkg/cdi/producer/annotations"
)

const (
	// AnnotationPrefix is the prefix for CDI container annotation keys.
	AnnotationPrefix = annotations.Prefix
)

// UpdateAnnotations updates annotations with a plugin-specific CDI device
// injection request for the given devices. Upon any error a non-nil error
// is returned and annotations are left intact. By convention plugin should
// be in the format of "vendor.device-type".
func UpdateAnnotations(a map[string]string, plugin string, deviceID string, devices []string) (map[string]string, error) {
	return annotations.Update(a, plugin, deviceID, devices)
}

// ParseAnnotations parses annotations for CDI device injection requests.
// The keys and devices from all such requests are collected into slices
// which are returned as the result. All devices are expected to be fully
// qualified CDI device names. If any device fails this check empty slices
// are returned along with a non-nil error. The annotations are expected
// to be formatted by, or in a compatible fashion to UpdateAnnotations().
func ParseAnnotations(a map[string]string) ([]string, []string, error) {
	return annotations.Parse(a)
}

// AnnotationKey returns a unique annotation key for an device allocation
// by a K8s device plugin. pluginName should be in the format of
// "vendor.device-type". deviceID is the ID of the device the plugin is
// allocating. It is used to make sure that the generated key is unique
// even if multiple allocations by a single plugin needs to be annotated.
func AnnotationKey(pluginName, deviceID string) (string, error) {
	return annotations.Key(pluginName, deviceID)
}

// AnnotationValue returns an annotation value for the given devices.
func AnnotationValue(devices []string) (string, error) {
	return annotations.Value(devices)
}
