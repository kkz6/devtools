.PHONY: build install clean test release help

# Variables
BINARY_NAME=devtools
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
INSTALL_PATH=/usr/local/bin

# Default target
help:
	@echo "DevTools Makefile"
	@echo "================="
	@echo ""
	@echo "Available targets:"
	@echo "  make build      - Build the devtools binary"
	@echo "  make install    - Build and install to $(INSTALL_PATH)"
	@echo "  make clean      - Remove built binaries"
	@echo "  make test       - Run tests"
	@echo "  make release    - Build release binaries for all platforms"
	@echo "  make run        - Build and run devtools"
	@echo ""

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo "Run 'devtools' to start using it."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f devtools-*
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build release binaries
release:
	@echo "Building release binaries..."
	@mkdir -p dist
	
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/devtools-darwin-amd64 .
	
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/devtools-darwin-arm64 .
	
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/devtools-linux-amd64 .
	
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/devtools-windows-amd64.exe .
	
	@echo "Release builds complete! Check the dist/ directory."

# Build and run
run: build
	@./$(BINARY_NAME) 