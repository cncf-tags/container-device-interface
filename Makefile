# Copyright © The CDI Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO_CMD   := go
GO_BUILD := $(GO_CMD) build
GO_TEST  := $(GO_CMD) test -race -v -cover

GO_LINT  := golint -set_exit_status
GO_FMT   := gofmt
GO_VET   := $(GO_CMD) vet

CDI_PKG  := $(shell grep ^module go.mod | sed 's/^module *//g')

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
BINARIES := $(patsubst %,bin/%,$(CMDS))

ifneq ($(V),1)
  Q := @
endif


#
# top-level targets
#

all: build

build: $(BINARIES)

clean: clean-binaries clean-schema

test: test-gopkgs test-schema

#
# validation targets
#

pre-pr-checks pr-checks: test fmt lint vet

fmt format:
	$(Q)$(GO_FMT) -s -d -w -e .

lint:
	$(Q)$(GO_LINT) -set_exit_status ./...
vet:
	$(Q)$(GO_VET) ./...

#
# build targets
#

$(BINARIES): bin/%:
	$(Q)echo "Building $@..."
	$(Q)(cd cmd/$(*) && $(GO_BUILD) -o $(abspath $@) .)

#
# go module tidy and verify targets
#
.PHONY: mod-tidy
.PHONY: mod-verify

mod-tidy:
	$(Q)for mod in $$(find . -name go.mod); do \
	    echo "Tidying $$mod..."; ( \
	        cd $$(dirname $$mod) && go mod tidy \
            ) || exit 1; \
	done

mod-verify:
	$(Q)for mod in $$(find . -name go.mod); do \
	    echo "Verifying $$mod..."; ( \
	        cd $$(dirname $$mod) && go mod verify | sed 's/^/  /g' \
	    ) || exit 1; \
	done

#
# cleanup targets
#

# clean up binaries
clean-binaries:
	$(Q) rm -f $(BINARIES)

# clean up schema validator
clean-schema:
	$(Q)rm -f schema/validate

#
# test targets
#

# tests for go packages
test-gopkgs:
	$(Q)$(GO_TEST) ./...

# tests for CDI Spec JSON schema
test-schema: bin/validate
	$(Q)echo "Building in schema..."; \
	$(MAKE) -C schema test


#
# dependencies
#

bin/validate: $(wildcard schema/*.json) $(wildcard cmd/validate/*.go cmd/validate/cmd/*.go) $(shell \
            for dir in \
                $$(cd ./cmd/validate; $(GO_CMD) list -f '{{ join .Deps "\n"}}' ./... | \
                      grep $(CDI_PKG)/pkg/ | \
                      sed 's:$(CDI_PKG):.:g'); do \
                find $$dir -name \*.go; \
            done | sort | uniq)

# quasi-automatic dependency for bin/cdi
bin/cdi: $(wildcard cmd/cdi/*.go cmd/cdi/cmd/*.go) $(shell \
            for dir in \
                $$(cd ./cmd/cdi; $(GO_CMD) list -f '{{ join .Deps "\n"}}' ./... | \
                      grep $(CDI_PKG)/pkg/ | \
                      sed 's:$(CDI_PKG):.:g'); do \
                find $$dir -name \*.go; \
            done | sort | uniq)

