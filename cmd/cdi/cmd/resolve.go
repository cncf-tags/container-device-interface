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

	"github.com/spf13/cobra"
)

type resolveFlags struct {
	output string
}

// resolveCmd is our command for resolving CDI devices present in an OCI Spec.
var resolveCmd = &cobra.Command{
	Aliases: []string{"res"},
	Use:     "resolve",
	Short:   "Resolve CDI devices present in an OCI Spec",
	Long: `
The 'resolve' command takes an OCI Spec file (use "-" for stdin),
resolves any CDI Devices present in the Spec and dumps the result.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("OCI Spec argument(s) expected\n")
			os.Exit(1)
		}
		if err := cdiResolveDevices(args...); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

/*
func resolveDevices(ociSpecFiles ...string) error {
	for _, ociSpecFile := range ociSpecFiles {
		ociSpec, err := readOCISpec(ociSpecFile)
		if err != nil {
			return err
		}

		resolved, err := cdi.ResolveDevices(ociSpec)
		if err != nil {
			return errors.Wrapf(err, "CDI device resolution failed in %q",
				ociSpecFile)
		}

		output := injectCfg.output
		if output == "" {
			if filepath.Ext(ociSpecFile) == ".json" {
				output = "json"
			} else {
				output = "yaml"
			}
		}

		if resolved != nil {
			fmt.Printf("OCI Spec %q: resolved devices %q\n", ociSpecFile,
				strings.Join(resolved, ", "))
			fmt.Printf("%s", marshalObject(2, ociSpec, output))
		}
	}

	return nil
}
*/

var (
	resolveCfg resolveFlags
)

func init() {
	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().StringVarP(&injectCfg.output,
		"output", "o", "", "output format for OCI Spec (json|yaml)")
}
