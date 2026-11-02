GO      ?= go
BINARY  ?= taskgo
BIN_DIR ?= bin

.PHONY: all build run test fmt vet check tidy clean help

all: check build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-8s %s\n", $$1, $$2}'

build: ## Build the binary into bin/
	$(GO) build -o $(BIN_DIR)/$(BINARY) ./cmd/taskgo

run: ## Run the program
	$(GO) run ./cmd/taskgo

test: ## Run all tests
	$(GO) test ./...

fmt: ## Format source files
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

check: fmt vet test ## Format, vet, and test

tidy: ## Sync go.mod and go.sum
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)
