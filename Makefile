# Makefile for TMForum Proxy Package

.PHONY: help build test clean run-proxy build-proxy

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the proxy package"
	@echo "  test         - Run tests for the proxy package"
	@echo "  build-proxy  - Build the command-line proxy tool"
	@echo "  run-proxy    - Run the proxy tool with example configuration"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install the proxy tool"

# Build the proxy package
build:
	go build ./proxy/...

# Run tests for the proxy package
test:
	go test -v ./proxy/...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./proxy/...
	go tool cover -html=coverage.out

# Build the command-line proxy tool
build-proxy:
	@echo "Building tmfproxy binary..."
	go build -o bin/tmfproxy cmd/tmfproxy/main.go
	@echo "Binary created at bin/tmfproxy"

# Run the proxy tool with example configuration
run-proxy: build-proxy
	@echo "Running proxy tool..."
	./bin/tmfproxy -help

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out
	go clean

# Install the proxy tool
install:
	go install ./cmd/tmfproxy

# Create output directory
setup:
	mkdir -p bin
	mkdir -p reports

# Run with example configuration
example: setup build-proxy
	@echo "Running proxy with example configuration..."
	./bin/tmfproxy -base-url "https://tmf.example.com" -progress

# Format code
fmt:
	go fmt ./proxy/...
	go fmt ./cmd/...

# Lint code
lint:
	golangci-lint run ./proxy/...
	golangci-lint run ./cmd/...

# Check for security vulnerabilities
security:
	gosec ./proxy/...
	gosec ./cmd/...

# Generate documentation
docs:
	@echo "Documentation is available in proxy/README.md"
	@echo "Example configurations are in proxy/config.example.*"

# Show package information
info:
	@echo "TMForum Proxy Package"
	@echo "===================="
	@echo "Package: github.com/hesusruiz/isbetmf/proxy"
	@echo "Command: cmd/tmfproxy"
	@echo "Configuration: proxy/config.example.*"
	@echo "Documentation: proxy/README.md"
