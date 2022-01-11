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
	"github.com/spf13/cobra"
)

type devicesFlags struct {
	verbose bool
	output  string
}

// devicesCmd is our command for listing devices found in the CDI registry.
var devicesCmd = &cobra.Command{
	Aliases: []string{"devs", "dev"},
	Use:     "devices",
	Short:   "List devices in the CDI registry",
	Long: `
The 'devices' command lists devices found in the CDI registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		cdiListDevices(devicesCfg.verbose, devicesCfg.output)
	},
}

var (
	devicesCfg devicesFlags
)

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.Flags().BoolVarP(&devicesCfg.verbose,
		"verbose", "v", false, "list CDI Spec details")
	devicesCmd.Flags().StringVarP(&devicesCfg.output,
		"output", "o", "", "output format for details (json|yaml)")
}
