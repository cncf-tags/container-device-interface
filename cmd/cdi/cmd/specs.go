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
	"strings"

	"github.com/spf13/cobra"
	"tags.cncf.io/container-device-interface/pkg/cdi"
)

type specFlags struct {
	verbose bool
	output  string
}

// specsCmd is our command for listing Spec files.
var specsCmd = &cobra.Command{
	Use:   "specs [vendor-list]",
	Short: "List available CDI Specs",
	Long: fmt.Sprintf(`
The 'specs' command lists all CDI Specs present in the registry.
If a vendor list is given, only CDI Specs by the given vendors are
listed. The CDI Specs are discovered and loaded to the registry
from CDI Spec directories. The default CDI Spec directories are:
    %s.`, strings.Join(cdi.DefaultSpecDirs, ", ")),
	Run: func(cmd *cobra.Command, vendors []string) {
		cdiListSpecs(specCfg.verbose, specCfg.output, vendors...)
	},
}

var (
	specCfg specFlags
)

func init() {
	rootCmd.AddCommand(specsCmd)
	specsCmd.Flags().BoolVarP(&specCfg.verbose,
		"verbose", "v", false, "list CDI Spec details")
	specsCmd.Flags().StringVarP(&specCfg.output,
		"output", "o", "", "output format for details (json|yaml)")
}
