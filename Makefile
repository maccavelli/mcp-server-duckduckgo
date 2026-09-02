MOD_VERSION := 1.26.6
BINARY_NAME=mcp-server-duckduckgo
DIST_DIR=dist
GIT_VERSION=$(shell git describe --tags --always --dirty 2>/dev/null)
VERSION?=$(GIT_VERSION)
# BUILD_KIND is "release" only for a tag build; the release workflow passes it
# explicitly. A local build stays "local" even when `git describe` resembles a
# tag, so self-update refuses to replace it without --force (MADR 0005).
BUILD_KIND?=local
# Prefer the user's Go toolchain install (go install ...), then PATH.
GOPATH_BIN     := $(shell go env GOPATH)/bin
GOLANGCI_LINT  ?= $(GOPATH_BIN)/golangci-lint
FLEET_LINT_CFG := .golangci.yml

.PHONY: all build clean test run install version build-all linux darwin-arm64 windows-amd64 help fmt vet lint vendor

all: help build-all

build: ## Compiles the Go application for the local OS/Arch
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=0 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)" -o $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ./cmd/$(BINARY_NAME)

build-all: linux darwin-arm64 windows-amd64 ## Compiles for multiple platforms

linux: ## Compiles for Linux AMD64
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)

darwin-arm64: ## Compiles for macOS ARM64
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)

windows-amd64: ## Compiles for Windows AMD64
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)

clean: ## Removes all build artifacts
	rm -rf $(DIST_DIR)

test: ## Runs all tests with verbose output
	go test -v ./...

fmt: ## Formats all Go source files
	go fmt ./...

vet: ## Runs go vet on the project
	go vet ./...

lint: ## Runs golangci-lint from $(go env GOPATH)/bin (override with GOLANGCI_LINT=)
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found at $(GOLANGCI_LINT)"; \
		echo "Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run -c $(FLEET_LINT_CFG) ./...

vendor: ## Re-vendor dependencies
	go mod vendor
	go mod tidy

run: build ## Builds and executes the local binary
	@BIN_NAME=$(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ; \
	$$BIN_NAME

install: build ## Copies the local binary to ~/.local/bin/
	@BIN_NAME=$(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ; \
	INSTALL_PATH=$(HOME)/.local/bin/$(BINARY_NAME) ; \
	if [ -f "$$INSTALL_PATH" ]; then mv "$$INSTALL_PATH" "$$INSTALL_PATH.old"; fi ; \
	cp $$BIN_NAME $$INSTALL_PATH ; \
	rm -f "$$INSTALL_PATH.old" ; \
	echo "Installed $(BINARY_NAME) to ~/.local/bin/"

version: build ## Displays the version of the local binary
	@BIN_NAME=$(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ; \
	$$BIN_NAME --version

help: ## Displays this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
