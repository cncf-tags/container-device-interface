//go:build windows

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

package cdi

import (
	"fmt"

	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// validateWildcards is a no-op on Windows as device nodes are unsupported.
func (d *DeviceNode) validateWildcards() error {
	return nil
}

// expandWildcards is a no-op on Windows, as device nodes are unsupported.
func expandWildcards(nodes []*cdi.DeviceNode) ([]*cdi.DeviceNode, error) {
	return nodes, nil
}

// fillMissingInfo fills in missing mandatory attributes from the host device.
func (d *DeviceNode) fillMissingInfo() error {
	return fmt.Errorf("unimplemented")
}
