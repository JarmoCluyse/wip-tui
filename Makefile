# Git TUI Makefile

.PHONY: build clean run test install uninstall help

# Binary name
BINARY_NAME=git-tui
CACHED_BINARY_NAME=git-tui-cached
OPTIMIZED_BINARY_NAME=git-tui-optimized

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Install path
INSTALL_PATH=/usr/local/bin

# Default target
all: build

## build: Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd

## build-cached: Build the cached binary with performance optimizations
build-cached:
	$(GOBUILD) -o $(CACHED_BINARY_NAME) ./cmd

## build-optimized: Build the optimized binary with worktree caching
build-optimized:
	$(GOBUILD) -o $(OPTIMIZED_BINARY_NAME) ./cmd

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(CACHED_BINARY_NAME) $(OPTIMIZED_BINARY_NAME)

## run: Build and run the application
run: build
	./$(BINARY_NAME)

## run-custom: Build and run with custom config
run-custom: build
	./$(BINARY_NAME) -c test_configs/custom.toml

## run-lot: Build and run with lot config (moderate repos)
run-lot: build
	./$(BINARY_NAME) -c test_configs/lot.toml

## run-alot: Build and run with alot config (many repos)  
run-alot: build
	./$(BINARY_NAME) -c test_configs/alot.toml

## run-cached: Build and run cached version
run-cached: build-cached
	./$(CACHED_BINARY_NAME)

## run-cached-custom: Build and run cached version with custom config
run-cached-custom: build-cached
	./$(CACHED_BINARY_NAME) -c test_configs/custom.toml

## run-cached-lot: Build and run cached version with lot config
run-cached-lot: build-cached
	./$(CACHED_BINARY_NAME) -c test_configs/lot.toml

## run-cached-alot: Build and run cached version with alot config
run-cached-alot: build-cached
	./$(CACHED_BINARY_NAME) -c test_configs/alot.toml

## run-optimized: Build and run optimized version
run-optimized: build-optimized
	./$(OPTIMIZED_BINARY_NAME)

## run-optimized-custom: Build and run optimized version with custom config
run-optimized-custom: build-optimized
	./$(OPTIMIZED_BINARY_NAME) -c test_configs/custom.toml

## run-optimized-lot: Build and run optimized version with lot config
run-optimized-lot: build-optimized
	./$(OPTIMIZED_BINARY_NAME) -c test_configs/lot.toml

## run-optimized-alot: Build and run optimized version with alot config
run-optimized-alot: build-optimized
	./$(OPTIMIZED_BINARY_NAME) -c test_configs/alot.toml

## perf-test: Performance comparison between regular and cached versions
perf-test: build build-cached
	@echo "Testing regular version with alot config..."
	@time timeout 3s ./$(BINARY_NAME) -c test_configs/alot.toml || true
	@echo "\nTesting cached version with alot config..."
	@time timeout 3s ./$(CACHED_BINARY_NAME) -c test_configs/alot.toml || true

## perf-test-optimized: Performance comparison between cached and optimized versions
perf-test-optimized: build-cached build-optimized
	@echo "Testing cached version with alot config..."
	@time timeout 3s ./$(CACHED_BINARY_NAME) -c test_configs/alot.toml || true
	@echo "\nTesting optimized version with alot config..."
	@time timeout 3s ./$(OPTIMIZED_BINARY_NAME) -c test_configs/alot.toml || true

## test: Run tests
test:
	$(GOTEST) -v ./...

## deps: Download dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

## install: Install binary to system path
install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed to $(INSTALL_PATH)"

## uninstall: Remove binary from system path
uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) removed from $(INSTALL_PATH)"

## dev: Run in development mode (with file watching if available)
dev: build
	@if command -v entr >/dev/null 2>&1; then \
		echo "Watching for changes... Press Ctrl+C to stop"; \
		find . -name "*.go" | entr -r make run; \
	else \
		echo "entr not found. Run 'make run' manually after changes."; \
		make run; \
	fi

## fmt: Format Go code
fmt:
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## check: Run all checks (fmt, vet, test)
check: fmt vet test

## release: Build optimized binary for release
release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_NAME) ./cmd

## help: Show this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'