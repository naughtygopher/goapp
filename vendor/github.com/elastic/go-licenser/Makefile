export VERSION := v0.4.0
export GO111MODULE ?= on
export GOBIN = $(shell pwd)/bin
OWNER ?= elastic
REPO ?= go-licenser
TEST_UNIT_FLAGS ?= -timeout 10s -p 4 -race -cover
TEST_UNIT_PACKAGE ?= ./...
GOLINT_PRESENT := $(shell command -v golint 2> /dev/null)
GOIMPORTS_PRESENT := $(shell command -v goimports 2> /dev/null)
RELEASED = $(shell git tag -l $(VERSION))
DEFAULT_LDFLAGS ?= -X main.version=$(VERSION)-dev -X main.commit=$(shell git rev-parse HEAD)
include build/Makefile.deps

define HELP
/////////////////////////////////////////
/\t$(REPO) Makefile \t\t/
/////////////////////////////////////////

## Build target

- build:                  It will build $(REPO) for the current architecture in bin/$(REPO).
- install:                It will install $(REPO) in the current system (by default in $(GOPATH)/bin/$(REPO)).

## Development targets

- deps:                   It will install the dependencies required to run developemtn targets.
- unit:                   Runs the unit tests.
- lint:                   Runs the linters.
- format:                 Formats the source files according to gofmt, goimports and go-licenser.
- update-golden-files:    Updates the test golden files.

## Release targets

- release:                Creates and publishes a new release matching the VERSION variable.
- snapshot:               Creates a snapshot locally in the dist/ folder.

endef
export HELP

.DEFAULT: help
.PHONY: help
help:
	@ echo "$$HELP"

.PHONY: deps
deps:
ifndef GOLINT_PRESENT
	@ go get -u golang.org/x/lint/golint
endif
ifndef GOIMPORTS_PRESENT
	@ go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: release_deps
release_deps: $(GOBIN)/goreleaser

.PHONY: update-golden-files
update-golden-files:
	$(eval GOLDEN_FILE_PACKAGES := "github.com/$(OWNER)/$(REPO)")
	@ go test $(GOLDEN_FILE_PACKAGES) -update

.PHONY: unit
unit:
	@ go test $(TEST_UNIT_FLAGS) $(TEST_UNIT_PACKAGE)

.PHONY: build
build: deps
	@ go build -o bin/$(REPO) -ldflags="$(DEFAULT_LDFLAGS)"

.PHONY: install
install: deps
	@ go install

.PHONY: lint
lint: build
	@ $(GOBIN)/golint -set_exit_status $(shell go list ./...)
	@ gofmt -d -e -s .
	@ $(GOBIN)/go-licenser -d -exclude golden

.PHONY: format
format: deps build
	@ gofmt -e -w -s .
	@ $(GOBIN)/goimports -w .
	@ $(GOBIN)/go-licenser -exclude golden

.PHONY: release
release: deps release_deps
	@ echo "-> Releasing $(REPO) $(VERSION)..."
	@ git fetch upstream
ifeq ($(strip $(RELEASED)),)
	@ echo "-> Creating and pushing a new tag $(VERSION)..."
	@ git tag $(VERSION)
	@ git push upstream $(VERSION)
	@ $(GOBIN)/goreleaser release --skip-validate --rm-dist
else
	@ echo "-> git tag $(VERSION) already present, skipping release..."
endif

.PHONY: snapshot
snapshot: deps release_deps
	@ echo "-> Snapshotting $(REPO) $(VERSION)..."
	@ $(GOBIN)/goreleaser release --snapshot --rm-dist
