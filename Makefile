.DEFAULT_GOAL := help

SHELL := bash
PATH := $(CURDIR)/.dev/go-tools/bin:$(PATH)
COMMIT_HASH := $(shell git rev-parse --short HEAD)

BUILD_LDFLAGS = "-s -w -X github.com/kohkimakimoto/enclave/v3/internal/version.CommitHash=$(COMMIT_HASH)"

# Load .env file if it exists.
ifneq (,$(wildcard ./.env))
  include .env
  export
endif

.PHONY: help
help: ## Show help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[/0-9a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'


# --------------------------------------------------------------------------------------
# Development environment
# --------------------------------------------------------------------------------------
.PHONY: setup
setup: ## Setup development environment
	@echo "==> Setting up development environment..."
	@mkdir -p $(CURDIR)/.dev/go-tools
	@export GOPATH=$(CURDIR)/.dev/go-tools && \
		go install github.com/chmouel/go-better-html-coverage@latest
	@export GOPATH=$(CURDIR)/.dev/go-tools && go clean -modcache && rm -rf $(CURDIR)/.dev/go-tools/pkg

.PHONY: clean
clean: ## Clean up development environment
	@rm -rf .dev


# --------------------------------------------------------------------------------------
# Build
# --------------------------------------------------------------------------------------
.PHONY: build
build: ## Build dev binary
	@mkdir -p .dev/build/dev
	@CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o .dev/build/dev/enclave ./cmd/enclave

.PHONY: build-release
build-release: ## Build release binary
	@mkdir -p .dev/build/release
	@CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -trimpath -o .dev/build/release/enclave ./cmd/enclave

.PHONY: build-clean
build-clean: ## Clean up build artifacts
	@rm -rf .dev/build


# --------------------------------------------------------------------------------------
# Testing, Formatting and etc.
# --------------------------------------------------------------------------------------
.PHONY: format
format: ## Format source code
	@go fmt ./...

.PHONY: test
test: ## Run tests
	@go test -race -timeout 30m ./...

.PHONY: test-short
test-short: ## Run short tests
	@go test -short -race -timeout 30m ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose outputting
	@go test -race -timeout 30m -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@mkdir -p $(CURDIR)/.dev/test
	@go test -race -coverpkg=./... -coverprofile=$(CURDIR)/.dev/test/coverage.out ./...
	@go-better-html-coverage -profile $(CURDIR)/.dev/test/coverage.out -o $(CURDIR)/.dev/test/coverage.html -n

.PHONY: test-cover-open
test-cover-open: ## Open coverage report in browser
	@open $(CURDIR)/.dev/test/coverage.html
