GO_CMD   := go
GO_BUILD := $(GO_CMD) build
GO_TEST  := $(GO_CMD) test -v -cover

GO_LINT  := golint -set_exit_status
GO_FMT   := gofmt
GO_VET   := $(GO_CMD) vet

CDI_PKG  := $(shell grep ^module go.mod | sed 's/^module *//g')
GO_MODS  := $(shell $(GO_CMD) list ./...)
GO_PKGS  := $(shell $(GO_CMD) list ./... | grep -v cmd/ | sed 's:$(CDI_PKG):.:g')

BINARIES := bin/cdi

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
# targets for running test prior to filing a PR
#

pre-pr-checks pr-checks: test fmt lint vet


fmt format:
	$(Q)report=$$($(GO_FMT) -s -d -w $$(find . -name *.go)); \
	    if [ -n "$$report" ]; then \
	        echo "$$report"; \
	        exit 1; \
	    fi

lint:
	$(Q)status=0; for f in $$(find . -name \*.go); do \
	    $(GO_LINT) $$f || status=1; \
	done; \
	exit $$status

vet:
	$(Q)$(GO_VET) $(GO_MODS)

#
# build targets
#

bin/%:
	$(Q)echo "Building $@..."; \
	$(GO_BUILD) -o $@ ./$(subst bin/,cmd/,$@)

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
	$(Q)status=0; for pkg in $(GO_PKGS); do \
	    $(GO_TEST) $$pkg; \
	    if [ $$? != 0 ]; then \
	        echo "*** Test FAILED for package $$pkg."; \
	        status=1; \
	    fi; \
	done; \
	exit $$status

# tests for CDI Spec JSON schema
test-schema:
	$(Q)echo "Building in schema..."; \
	$(MAKE) -C schema test


#
# dependencies
#

# quasi-automatic dependency for bin/cdi
bin/cdi: $(wildcard cmd/cdi/*.go cmd/cdi/cmd/*.go) $(shell \
            for dir in \
                $$($(GO_CMD) list -f '{{ join .Deps "\n"}}' ./cmd/cdi/... | \
                      grep $(CDI_PKG)/pkg/ | \
                      sed 's:$(CDI_PKG):.:g'); do \
                find $$dir -name \*.go; \
            done | sort | uniq)
