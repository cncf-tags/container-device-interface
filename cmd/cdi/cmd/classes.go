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

// classesCmd is our command for listing device classes in the registry.
var classesCmd = &cobra.Command{
	Use:   "classes",
	Short: "List CDI device classes",
	Long:  `List CDI device classes found in the registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		cdiListClasses()
	},
}

func init() {
	rootCmd.AddCommand(classesCmd)
}
