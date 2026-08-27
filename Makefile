BINARY := repokit
BUILD_DIR := dist
LOCAL_BIN := $(CURDIR)/bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILD_DATE)

# ==============================================================================
# Tools
# ==============================================================================

GOLANGCI := $(LOCAL_BIN)/golangci-lint
GORELEASER := $(LOCAL_BIN)/goreleaser
COMMITKIT := $(LOCAL_BIN)/commitkit

GOLANGCI_VERSION := v2.12.2
GORELEASER_VERSION := v2.18.0
COMMITKIT_VERSION := v0.1.0

# ==============================================================================
# Commands
# ==============================================================================

.PHONY: setup devtools hooks build install test lint clean release

setup:
	go mod download
	$(MAKE) devtools
	$(MAKE) hooks

devtools:
	@mkdir -p $(LOCAL_BIN)

	@if [ ! -x "$(GOLANGCI)" ]; then \
		echo "Installing golangci-lint $(GOLANGCI_VERSION)..."; \
		GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION); \
	fi

	@if [ ! -x "$(COMMITKIT)" ]; then \
		echo "Installing commitkit $(COMMITKIT_VERSION)..."; \
		GOBIN=$(LOCAL_BIN) go install github.com/destyk/commitkit@$(COMMITKIT_VERSION); \
	fi

	@if [ ! -x "$(GORELEASER)" ]; then \
		echo "Installing goreleaser $(GORELEASER_VERSION)..."; \
		GOBIN=$(LOCAL_BIN) go install github.com/goreleaser/goreleaser/v2@$(GOLANGCI_VERSION); \
	fi

hooks:
	@echo "Installing hooks..."
	@./bin/commitkit install-hook
	@echo "Hooks installed successfully"

build:
	go build \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BINARY) \
		./

install:
	go install \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		./

test:
	go test ./...

lint:
	go vet ./...

release:
	@mkdir -p $(BUILD_DIR)

	GOOS=darwin GOARCH=arm64 go build \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY)-darwin-arm64 \
		./

	GOOS=darwin GOARCH=amd64 go build \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY)-darwin-amd64 \
		./

	GOOS=linux GOARCH=amd64 go build \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY)-linux-amd64 \
		./

	GOOS=linux GOARCH=arm64 go build \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY)-linux-arm64 \
		./

	GOOS=windows GOARCH=amd64 go build \
		-trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe \
		./

release-snapshot:
	$(GORELEASER) release --snapshot --clean

clean:
	rm -rf $(BINARY) $(BUILD_DIR)