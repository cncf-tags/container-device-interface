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
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Option is an option to change some aspect of default CDI behavior.
type Option func(*Cache) error

// Cache stores CDI Specs loaded from Spec directories.
type Cache struct {
	sync.Mutex
	specDirs []string
	specs    map[string][]*Spec
	devices  map[string]*Device
	errors   map[string][]error
}

// NewCache creates a new CDI Cache. The cache is populated from a set
// of CDI Spec directories. These can be specified using a WithSpecDirs
// option. The default set of directories is exposed in DefaultSpecDirs.
func NewCache(options ...Option) (*Cache, error) {
	c := &Cache{}

	if err := c.Configure(options...); err != nil {
		return nil, err
	}
	if len(c.specDirs) == 0 {
		c.Configure(WithSpecDirs(DefaultSpecDirs...))
	}

	return c, c.Refresh()
}

// Configure applies options to the cache. Updates the cache if options have
// changed.
func (c *Cache) Configure(options ...Option) error {
	if len(options) == 0 {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	for _, o := range options {
		if err := o(c); err != nil {
			return errors.Wrapf(err, "failed to apply cache options")
		}
	}

	return nil
}

// Refresh rescans the CDI Spec directories and refreshes the Cache.
func (c *Cache) Refresh() error {
	var (
		specs      = map[string][]*Spec{}
		devices    = map[string]*Device{}
		conflicts  = map[string]struct{}{}
		specErrors = map[string][]error{}
		result     []error
	)

	// collect errors per spec file path and once globally
	collectError := func(err error, paths ...string) {
		result = append(result, err)
		for _, path := range paths {
			specErrors[path] = append(specErrors[path], err)
		}
	}
	// resolve conflicts based on device Spec priority (order of precedence)
	resolveConflict := func(name string, dev *Device, old *Device) bool {
		devSpec, oldSpec := dev.GetSpec(), old.GetSpec()
		devPrio, oldPrio := devSpec.GetPriority(), oldSpec.GetPriority()
		switch {
		case devPrio > oldPrio:
			return false
		case devPrio == oldPrio:
			devPath, oldPath := devSpec.GetPath(), oldSpec.GetPath()
			collectError(errors.Errorf("conflicting device %q (specs %q, %q)",
				name, devPath, oldPath), devPath, oldPath)
			conflicts[name] = struct{}{}
		}
		return true
	}

	_ = scanSpecDirs(c.specDirs, func(path string, priority int, spec *Spec, err error) error {
		path = filepath.Clean(path)
		if err != nil {
			collectError(errors.Wrapf(err, "failed to load CDI Spec"), path)
			return nil
		}

		vendor := spec.GetVendor()
		specs[vendor] = append(specs[vendor], spec)

		for _, dev := range spec.devices {
			qualified := dev.GetQualifiedName()
			other, ok := devices[qualified]
			if ok {
				if resolveConflict(qualified, dev, other) {
					continue
				}
			}
			devices[qualified] = dev
		}

		return nil
	})

	for conflict := range conflicts {
		delete(devices, conflict)
	}

	c.Lock()
	defer c.Unlock()

	c.specs = specs
	c.devices = devices
	c.errors = specErrors

	if len(result) > 0 {
		return multierror.Append(nil, result...)
	}

	return nil
}
