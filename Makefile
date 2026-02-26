# Makefile for HTTP Client Tool
# AI Agent Note: Simple build automation

.PHONY: build test clean install examples help format lint check

# Default target
all: build

# Format code
format:
	@echo "Formatting code..."
	gofmt -w .
	@echo "Code formatted."

# Lint code
lint:
	@echo "Linting code..."
	go vet ./...
	@echo "Linting complete."

# Check formatting, linting, and tests
check: format lint
	@echo "Running test suite..."
	./test.sh

# Build the binary
build:
	@echo "Building http client..."
	go build -o http

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o http-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o http-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o http-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o http-windows-amd64.exe

# Test with examples
test: build
	@echo "Testing simple GET request..."
	@cat examples/simple-get.http | ./http -dry-run
	@echo ""
	@echo "Testing POST with JSON..."
	@cat examples/post-json.http | ./http -dry-run
	@echo ""
	@echo "Testing environment variables..."
	@export API_TOKEN="test-token" API_KEY="test-key" && cat examples/with-env-vars.http | ./http -dry-run
	@echo ""
	@echo "Testing shell commands..."
	@cat examples/with-shell-commands.http | ./http -dry-run

# Run all examples
examples: build
	@echo "Running all examples..."
	@for file in examples/*.http; do \
		echo "=== $$file ===" ; \
		cat $$file | ./http -dry-run ; \
		echo "" ; \
	done

# Install to PATH
install: build
	@echo "Installing to /usr/local/bin/..."
	@sudo mv http /usr/local/bin/
	@echo "Installed! You can now use 'http' from anywhere."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f http http-*
	@echo "Clean complete."

# Show help
help:
	@echo "HTTP Client Tool - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build      - Build the http client binary"
	@echo "  make build-all  - Build for multiple platforms"
	@echo "  make format     - Format all Go files with gofmt"
	@echo "  make lint       - Lint code with go vet"
	@echo "  make check      - Format, lint, and run all tests"
	@echo "  make test       - Run tests with example files"
	@echo "  make examples   - Run all example requests"
	@echo "  make install    - Install to /usr/local/bin"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Before submitting a PR, always run: make check"
	@echo ""
	@echo "Usage examples:"
	@echo "  cat request.http | ./http"
	@echo "  cat request.http | ./http -dry-run"
	@echo "  cat request.http | ./http -no-color"
