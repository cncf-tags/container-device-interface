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

// dirsCmd is our command for listing CDI Spec directories in use.
var dirsCmd = &cobra.Command{
	Use:   "dirs",
	Short: "Show CDI Spec directories in use",
	Long: `
Show which directories are used by the registry to discover and
load CDI Specs. The later an entry is in the list the higher its
priority. This priority is inherited by Spec files loaded from
the directory and is used to resolve device conflicts. If there
are multiple definitions for a CDI device, the Spec file with
the highest priority takes precedence over the others.`,
	Run: func(cmd *cobra.Command, args []string) {
		cdiShowSpecDirs()
	},
}

func init() {
	rootCmd.AddCommand(dirsCmd)
}
