/*
   Copyright © 2021 The CDI Authors

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

package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

func chooseFormat(format string, path string) string {
	if format == "" {
		if ext := filepath.Ext(path); ext == ".json" || ext == ".yaml" {
			format = ext[1:]
		} else {
			format = "yaml"
		}
	}
	return format
}

func marshalObject(level int, obj interface{}, format string) string {
	var (
		raw []byte
		err error
		out string
	)

	if format == "json" {
		raw, err = json.MarshalIndent(obj, "", "  ")
	} else {
		raw, err = yaml.Marshal(obj)
	}

	if err != nil {
		return fmt.Sprintf("%s<failed to dump object: %v\n", indent(level), err)
	}

	for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		out += indent(level) + line + "\n"
	}

	return out
}

func indent(level int) string {
	format := fmt.Sprintf("%%%d.%ds", level, level)
	return fmt.Sprintf(format, "")
}
