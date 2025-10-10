# Git TUI Makefile

.PHONY: build clean run test install uninstall help

# Binary name
BINARY_NAME=git-tui

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

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

## run: Build and run the application
run: build
	./$(BINARY_NAME)

## run-custom: Build and run with custom config
run-custom: build
	./$(BINARY_NAME) -c test_configs/custom.toml

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