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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

// validateCmd is our CDI command for validating CDI Spec files in the cache.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "List CDI cache errors",
	Long: `
The 'validate' command lists errors encountered during the population
of the CDI cache. It exits with an exit status of 1 if any errors
were reported by the cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		cache := cdi.GetDefaultCache()
		cdiErrors := cache.GetErrors()
		if len(cdiErrors) == 0 {
			fmt.Printf("No CDI cache errors.\n")
			return
		}

		fmt.Printf("CDI cache has errors:\n")
		for path, specErrors := range cdiErrors {
			fmt.Printf("Spec file %s:\n", path)
			for idx, err := range specErrors {
				fmt.Printf("  %2d: %v\n", idx, strings.TrimSpace(err.Error()))
			}
		}

		for _, v := range cache.ListVendors() {
			for _, s := range cache.GetVendorSpecs(v) {
				if err := specs.ValidateVersion(s.Spec); err != nil {
					fmt.Printf("Spec file %s failed version validation: %v\n", s.GetPath(), err)
				}
			}
		}

		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
