GOTOOLS := \
	github.com/alecthomas/gometalinter \
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \

DIRS     ?= $(shell find . -name '*.go' | grep --invert-match 'vendor' | xargs -n 1 dirname | sort --unique)
PKG_NAME ?= sentry-echo

BFLAGS ?=
LFLAGS ?=
TFLAGS ?=

COVERAGE_PROFILE ?= coverage.out

## default command
.DEFAULT_GOAL := help

.PHONY: clean
clean: ## Removes Go temporary build files build directory
	@echo "---> Cleaning"
	rm -rf ./vendor

.PHONY: enforce
enforce: ## Enforces code coverage
	@echo "---> Enforcing coverage"
	./scripts/coverage.sh $(COVERAGE_PROFILE)

.PHONY: html
html: ## Generates an HTML coverage report
	@echo "---> Generating HTML coverage report"
	go tool cover -html $(COVERAGE_PROFILE)

.PHONY: lint
lint: ## Runs all linters
	@echo "---> Linting"
	gometalinter --vendor --tests --deadline=2m $(LFLAGS) $(DIRS)

.PHONY: setup
setup: ## Installs all development dependencies
	@echo "--> Installing linter"
	go get -u -v $(GOTOOLS)
	gometalinter --install

.PHONY: install
install: ## Installs dependencies
	@echo "---> Installing dependencies"
	dep ensure

.PHONY: test
test: ## Runs all the tests and outputs the coverage report
	@echo "---> Testing"
	ENVIRONMENT=test go test ./... -race -coverprofile $(COVERAGE_PROFILE) $(TFLAGS)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
