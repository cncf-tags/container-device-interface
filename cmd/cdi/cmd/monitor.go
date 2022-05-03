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
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/fsnotify.v1"

	cdi "github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
)

type monitorFlags struct {
	verbose bool
	output  string
}

// monitorCmd is our command for monitoring CDI Spec refreshes.
var monitorCmd = &cobra.Command{
	Use:   "monitor [specs] [vendors] [classes] [devices] [all]",
	Short: "Monitor CDI Spec directories and refresh on changes",
	Long: `
The 'monitor' command monitors the CDI Spec directories and refreshes
the cache upon changes. The arguments passed to monitor control what
information to show upon each refresh.`,
	Run: func(cmd *cobra.Command, args []string) {
		monitorSpecDirs(args...)
	},
}

func monitorSpecDirs(args ...string) {
	var (
		registry = cdi.GetRegistry()
		specDirs = registry.GetSpecDirectories()
		dirWatch *fsnotify.Watcher
		err      error
		done     chan error
	)

	dirWatch, err = monitorDirectories(specDirs...)
	if err != nil {
		fmt.Printf("failed to set up CDI Spec dir monitoring: %v\n", err)
		os.Exit(1)
	}

	for _, dir := range specDirs {
		if _, err = os.Stat(dir); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("failed to stat CDI Spec directory %s: %v\n", dir, err)
				os.Exit(1)
			}
			fmt.Printf("WARNING: CDI Spec directory %s does not exist...\n", dir)
			continue
		}

		if err = dirWatch.Add(dir); err != nil {
			fmt.Printf("failed to watch CDI directory %q: %v\n", dir, err)
			os.Exit(1)
		}
	}

	done = make(chan error, 1)

	go func() {
		var (
			// don't print registry content more often than this
			oneSecond = 1 * time.Second
			timer     *time.Timer
		)

		if len(args) == 0 {
			args = []string{"all"}
		}

		timer = time.NewTimer(oneSecond)
		refresh := timer.C

		defer func() {
			timer.Stop()
		}()

		for {
			select {
			case _, ok := <-dirWatch.Events:
				if !ok {
					close(done)
					return
				}

				if refresh != nil {
					if !timer.Stop() {
						<-timer.C
					}
				}
				timer.Reset(oneSecond)
				refresh = timer.C

			case err, ok := <-dirWatch.Errors:
				if ok {
					done <- err
				}
				return

			case _ = <-refresh:
				refresh = nil
				cdiPrintRegistry(args...)
			}
		}
	}()

	err = <-done
	if err != nil {
		fmt.Printf("CDI Spec watch failed: %v\n", err)
		os.Exit(1)
	}
}

func monitorDirectories(dirs ...string) (*fsnotify.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create directory watch for %s",
			strings.Join(dirs, ","))
	}

	for _, dir := range dirs {
		if _, err = os.Stat(dir); err != nil {
			fmt.Printf("WARNING: failed to stat dir %q, NOT watching it...", dir)
			continue
		}

		if err = w.Add(dir); err != nil {
			return nil, errors.Wrapf(err, "failed to add %q to fsnotify watch", dir)
		}
	}

	return w, nil
}

var (
	monitorCfg monitorFlags
)

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().BoolVarP(&monitorCfg.verbose,
		"verbose", "v", false, "print details")
	monitorCmd.Flags().StringVarP(&monitorCfg.output,
		"output", "o", "", "output format for details (json|yaml)")
}
