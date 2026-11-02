GO      ?= go
BINARY  ?= taskgo
BIN_DIR ?= bin
UI      ?= ./scripts/ui.sh

VERSION ?= 0.1.0-dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -X github.com/0xengr/taskgo/internal/cli.version=$(VERSION) -X github.com/0xengr/taskgo/internal/cli.commit=$(COMMIT)

FMT  = $(GO) fmt ./...
VET  = $(GO) vet ./...
TEST = $(GO) test ./...

.PHONY: all build build-all run test fmt vet check tidy clean bench help

all: check build

help: ## Show this help
	@$(UI) banner "help"
	@printf '  Available targets\n\n'
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

build: ## Build the binary into bin/
	@$(UI) banner "build"
	@$(UI) spin "compiling taskgo" -- $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/taskgo
	@$(UI) ok "bin/$(BINARY) is ready"

build-all: ## Cross-compile for Linux, macOS, and Windows
	@$(UI) banner "cross-compile"
	@$(UI) spin "linux/amd64" -- env GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taskgo-linux-amd64 ./cmd/taskgo
	@$(UI) spin "darwin/amd64" -- env GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taskgo-darwin-amd64 ./cmd/taskgo
	@$(UI) spin "windows/amd64" -- env GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taskgo-windows-amd64.exe ./cmd/taskgo
	@$(UI) ok "three binaries in $(BIN_DIR)/"

run: ## Run the program
	@$(UI) banner "run"
	@$(UI) run "taskgo" -- $(GO) run -ldflags "$(LDFLAGS)" ./cmd/taskgo

test: ## Run all tests
	@$(UI) banner "test"
	@$(UI) run "testing all packages" -- $(TEST)

fmt: ## Format source files
	@$(UI) spin "formatting" -- $(FMT)

vet: ## Run go vet
	@$(UI) spin "vetting" -- $(VET)

check: ## Format, vet, and test
	@$(UI) banner "check"
	@$(UI) spin "formatting" -- $(FMT)
	@$(UI) spin "vetting" -- $(VET)
	@$(UI) run "testing" -- $(TEST)
	@$(UI) ok "all checks passed"

bench: build ## Run process-level benchmarks on 10,000 tasks
	@$(UI) banner "benchmark"
	@$(UI) run "10,000 tasks" -- $(GO) run ./cmd/taskgo-bench -binary $(BIN_DIR)/$(BINARY)

tidy: ## Sync go.mod and go.sum
	@$(UI) spin "tidying modules" -- $(GO) mod tidy
	@$(UI) ok "go.mod and go.sum are tidy"

clean: ## Remove build artifacts
	@$(UI) banner "clean"
	@rm -rf $(BIN_DIR)
	@$(UI) ok "removed $(BIN_DIR)/"
