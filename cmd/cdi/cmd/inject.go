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
	"io"
	"os"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
)

type injectFlags struct {
	output string
}

// injectCmd is our command for injecting CDI devices into an OCI Spec.
var injectCmd = &cobra.Command{
	Aliases: []string{"inj", "in", "oci"},
	Use:     "inject <OCI Spec File> <CDI-device-list>",
	Short:   "Inject CDI devices into an OCI Spec",
	Long: `
The 'inject' command reads an OCI Spec from a file (use "-" for stdin),
injects a requested set of CDI devices into it and dumps the resulting
updated OCI Spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Printf("OCI Spec argument and devices expected\n")
			os.Exit(1)
		}

		ociSpec, err := readOCISpec(args[0])
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		if err := cdiInjectDevices(injectCfg.output, ociSpec, args[1:]); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func readOCISpec(path string) (*oci.Spec, error) {
	var (
		spec *oci.Spec
		data []byte
		err  error
	)

	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read OCI Spec (%q): %w", path, err)
	}

	spec = &oci.Spec{}
	if err = yaml.Unmarshal(data, spec); err != nil {
		return nil, fmt.Errorf("failed to parse OCI Spec (%q): %w", path, err)
	}

	return spec, nil
}

var (
	injectCfg injectFlags
)

func init() {
	rootCmd.AddCommand(injectCmd)
	injectCmd.Flags().StringVarP(&injectCfg.output,
		"output", "o", "", "output format for OCI Spec (json|yaml)")
}
