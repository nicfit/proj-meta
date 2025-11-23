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

all: clean build test

.PHONY: build
build:
	go build ./...
	go build -o bin/proj-meta ./cmd.go

.PHONY: clean
clean:
	rm -rf bin
	rm -f RELEASE_NOTES.md.* $(TEST_JUNIT) $(TEST_COVERAGE)

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

#github-release: extract-release-notes
#	@if [ -n "`./bin/loadctl version --prerelease`" ]; then \
#       prerelease="--prerelease"; \
#    else \
#       latest="--latest"; \
#    fi ;\
#	gh release create $(VERSION) --title "Loadctl Daemon $(VERSION)" \
#         $${prerelease} $${latest} \
#         --notes-file RELEASE_NOTES.md.$(VERSION) \
#         dist/*.zip dist/*.tgz

#extract-release-notes:
#	@sed -n '/^## $(VERSION)/,/^## /p' RELEASE_NOTES.md | grep -ve '^## ' >| RELEASE_NOTES.md.$(VERSION)

GOTESTSUM_VERSION ?= latest
GOTESTSUM = $(LOCALBIN)/gotestsum-$(GOTESTSUM_VERSION)

.PHONY: gotestsum
gotestsum: $(GOTESTSUM)
$(GOTESTSUM): $(LOCALBIN)
	$(call go-install-tool,$(GOTESTSUM),gotest.tools/gotestsum,$(GOTESTSUM_VERSION))

GOLANGCI_LINT_VERSION ?= latest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

lint: golangci-lint  ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) version
	$(GOLANGCI_LINT) run
lint-fix: golangci-lint  ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin/tools
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef