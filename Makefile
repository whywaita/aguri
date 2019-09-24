NAME := aguri
VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.revision=$(REVISION)'
GOIMPORTS ?= goimports
GOCILINT ?= golangci-lint
GO ?= GO111MODULE=on go
.DEFAULT_GOAL := help

.PHONY: test
test:  ## Run the tests.
	@$(GO) test ./...

.PHONY: build
build: main.go  ## Build a binary.
	$(GO) build -ldflags "$(LDFLAGS)"

.PHONY: linux
linux: main.go  ## Build a binary for linux.
	GOOS=linux GOARCH=amd64 make build

.PHONY: cross
cross: main.go  ## Build binaries for cross platform.
	mkdir -p pkg
	@# darwin
	@for arch in "amd64" "386"; do \
		GOOS=darwin GOARCH=$${arch} make build; \
		chmod +x aguri; \
		zip pkg/aguri_$(VERSION)_darwin_$${arch}.zip aguri; \
	done;
	@# linux
	@for arch in "amd64" "386" "arm64" "arm"; do \
		GOOS=linux GOARCH=$${arch} make build; \
		chmod +x aguri; \
		zip pkg/aguri_$(VERSION)_linux_$${arch}.zip aguri; \
	done;

.PHONY: help
help: ## Show help text
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[0m %s\n", $$1, $$2}'
