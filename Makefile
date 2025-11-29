SHELL = /usr/bin/env bash
.DEFAULT_GOAL := build

VERSION := $(shell cat version.txt | xargs)

TESTS ?= ./...
TEST_COVERAGE := coverage.out
TEST_JUNIT ?= .test-results.xml
TEST_OPTS ?= -coverprofile=$(TEST_COVERAGE)

CGO_ENABLED ?=
export CGO_ENABLED
GOOS ?=
export GOOS
GOARCH ?=
export GOARCH

EXAMPLES := bin/cli bin/cobra

all: clean build test

.PHONY: build
build:
	go build ./...
	for ex in $(EXAMPLES); do \
  	  base=$$(basename $$ex); \
  	  go build -o $$ex examples/$$base/$$base.go; \
	done

.PHONY: clean
clean:
	rm -f $(EXAMPLES)
	rm -f RELEASE_NOTES.md.* $(TEST_JUNIT) $(TEST_COVERAGE)

clean-all: clean
	rm -rf ./bin

export GOTESTSUM_FORMAT ?= dots-v2
test: lint test-unit
test-unit: gotestsum
	go vet ./...
	$(GOTESTSUM) --format-hide-empty-pkg  --junitfile $(TEST_JUNIT) -- $(TEST_OPTS) $(TESTS)
test-coverage: test
	go tool cover -html=$(TEST_COVERAGE)

release-tag: build
	@if [ `git branch --show-current` != "main" ]; then \
       echo "!!! Not on main branch. !!!"; \
       exit 1; \
    fi

	@(git diff --quiet && git diff --quiet --staged) || \
     (printf "\n!!! Working repo has uncommitted/un-staged changes. !!!\n" && \
      printf "\nCommit and try again.\n" && false)

	@if ! git tag --annotate $(VERSION) 2> /dev/null; then \
       echo "!!! $(VERSION) already exists; update server/version.txt !!!"; \
       exit 1; \
    fi
	git push origin $(VERSION)

update-deps:
	go get -u ./...
	go mod tidy
	git status --porcelain | grep -q 'go\.mod\|go\.sum'

github-release:
	@stage="--latest"; \
	if [ -n "`./bin/cobra version --prerelease`" ]; then \
       stage="--prerelease"; \
    else \
       stage="--latest"; \
    fi ;\
	gh release create $(VERSION) --title "proj-meta $(VERSION)" $${stage}

GOTESTSUM_VERSION ?= latest
GOTESTSUM = $(TOOLSBIN)/gotestsum-$(GOTESTSUM_VERSION)

.PHONY: gotestsum
gotestsum: $(GOTESTSUM)
$(GOTESTSUM): $(TOOLSBIN)
	$(call go-install-tool,$(GOTESTSUM),gotest.tools/gotestsum,$(GOTESTSUM_VERSION))

GOLANGCI_LINT_VERSION ?= latest
GOLANGCI_LINT = $(TOOLSBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(TOOLSBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

lint: golangci-lint  ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) version
	$(GOLANGCI_LINT) run
lint-fix: golangci-lint  ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

## Location to install dependencies to
TOOLSBIN ?= $(shell pwd)/bin/tools
$(TOOLSBIN):
	mkdir -p $(TOOLSBIN)

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(TOOLSBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef