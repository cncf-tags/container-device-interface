GO_CMD   := go
GO_BUILD := $(GO_CMD) build
GO_TEST  := $(GO_CMD) test -v -cover

GO_LINT  := golint -set_exit_status
GO_FMT   := gofmt
GO_VET   := $(GO_CMD) vet

CDI_PKG  := $(shell grep ^module go.mod | sed 's/^module *//g')

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
	$(Q)$(GO_TEST) ./...

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
