GO      ?= go
BINARY  ?= taskgo
BIN_DIR ?= bin

VERSION ?= 0.1.0-dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -X github.com/0xengr/taskgo/internal/cli.version=$(VERSION) -X github.com/0xengr/taskgo/internal/cli.commit=$(COMMIT)

.PHONY: all build build-all run test fmt vet check tidy clean bench help

all: check build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-8s %s\n", $$1, $$2}'

build: ## Build the binary into bin/
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/taskgo

build-all: ## Cross-compile for Linux, macOS, and Windows
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taskgo-linux-amd64 ./cmd/taskgo
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taskgo-darwin-amd64 ./cmd/taskgo
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taskgo-windows-amd64.exe ./cmd/taskgo

run: ## Run the program
	$(GO) run -ldflags "$(LDFLAGS)" ./cmd/taskgo

test: ## Run all tests
	$(GO) test ./...

fmt: ## Format source files
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

check: fmt vet test ## Format, vet, and test

bench: build ## Run process-level benchmarks on 10,000 tasks
	$(GO) run ./cmd/taskgo-bench -binary $(BIN_DIR)/$(BINARY)

tidy: ## Sync go.mod and go.sum
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)
