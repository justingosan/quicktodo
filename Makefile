# QuickTodo CLI Makefile

# Variables
BINARY_NAME=quicktodo
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | cut -d' ' -f3)
LDFLAGS=-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GoVersion=$(GO_VERSION)'

# Build directory
BUILD_DIR=build

# Installation paths
PREFIX=/usr/local
BINDIR=$(PREFIX)/bin

# Platform-specific settings
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    INSTALL_CMD=install
else
    INSTALL_CMD=install
endif

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Install the binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR)
	$(INSTALL_CMD) $(BUILD_DIR)/$(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(BINDIR)/$(BINARY_NAME)"
	@echo ""
	@echo "You can now run: $(BINARY_NAME) --help"

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(BINDIR)..."
	rm -f $(BINDIR)/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME)"

# Development build (same as build but in current directory)
.PHONY: dev
dev:
	@echo "Building $(BINARY_NAME) for development..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "Built $(BINARY_NAME)"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2" && exit 1)
	golangci-lint run

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	@echo "Building for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	@echo "Building for Windows AMD64..."
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Cross-platform builds complete in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Create release archives
.PHONY: release
release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/releases
	
	# Linux AMD64
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64 -C ../. README.md QUICKTODO.md
	
	# Linux ARM64
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-arm64 -C ../. README.md QUICKTODO.md
	
	# macOS AMD64
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64 -C ../. README.md QUICKTODO.md
	
	# macOS ARM64
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64 -C ../. README.md QUICKTODO.md
	
	# Windows AMD64
	cd $(BUILD_DIR) && zip releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe ../README.md ../QUICKTODO.md
	
	@echo "Release archives created in $(BUILD_DIR)/releases/"
	@ls -la $(BUILD_DIR)/releases/

# Check if binary works after build
.PHONY: check
check: build
	@echo "Testing built binary..."
	@$(BUILD_DIR)/$(BINARY_NAME) version
	@echo "Binary check passed"

# Full development workflow
.PHONY: ci
ci: fmt vet test build check
	@echo "CI pipeline completed successfully"

# Show help
.PHONY: help
help:
	@echo "QuickTodo CLI Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the binary for current platform"
	@echo "  dev         - Build binary in current directory (for development)"
	@echo "  install     - Build and install to $(BINDIR)"
	@echo "  uninstall   - Remove installed binary"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code (requires golangci-lint)"
	@echo "  vet         - Vet code"
	@echo "  build-all   - Cross-compile for all platforms"
	@echo "  release     - Create release archives for all platforms"
	@echo "  check       - Test that built binary works"
	@echo "  ci          - Run full CI pipeline (fmt, vet, test, build, check)"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     = $(VERSION)"
	@echo "  BUILD_TIME  = $(BUILD_TIME)"
	@echo "  GO_VERSION  = $(GO_VERSION)"
	@echo "  PREFIX      = $(PREFIX)"
	@echo "  BINDIR      = $(BINDIR)"

# Default help if no target specified
.DEFAULT_GOAL := help