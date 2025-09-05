# Makefile for TMForum Proxy Package

.PHONY: help build test clean run-proxy build-proxy

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the reporting package"
	@echo "  test         - Run tests for the reporting package"
	@echo "  build-proxy  - Build the command-line reporting tool"
	@echo "  run-proxy    - Run the reporting tool with example configuration"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install the reporting tool"

# Build the reporting package
build:
	go build ./reporting/...

# Run tests for the reporting package
test:
	go test -v ./reporting/...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./reporting/...
	go tool cover -html=coverage.out

# Build the command-line reporting tool
build-proxy:
	@echo "Building tmfproxy binary..."
	go build -o bin/tmfproxy cmd/reporting/main.go
	@echo "Binary created at bin/tmfproxy"

# Run the reporting tool with example configuration
run-proxy: build-proxy
	@echo "Running reporting tool..."
	./bin/tmfproxy -help

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out
	go clean

# Install the reporting tool
install:
	go install ./cmd/reporting

# Create output directory
setup:
	mkdir -p bin
	mkdir -p reports

# Run with example configuration
example: setup build-proxy
	@echo "Running reporting tool with example configuration..."
	./bin/tmfproxy -base-url "https://tmf.example.com" -progress

# Format code
fmt:
	go fmt ./reporting/...
	go fmt ./cmd/...

# Lint code
lint:
	golangci-lint run ./reporting/...
	golangci-lint run ./cmd/...

# Check for security vulnerabilities
security:
	gosec ./reporting/...
	gosec ./cmd/...

# Generate documentation
docs:
	@echo "Documentation is available in reporting/README.md"
	@echo "Example configurations are in reporting/config.example.*"

# Show package information
info:
	@echo "TMForum Reporting Package"
	@echo "========================"
	@echo "Package: github.com/hesusruiz/isbetmf/reporting"
	@echo "Command: cmd/reporting"
	@echo "Configuration: reporting/config.example.*"
	@echo "Documentation: reporting/README.md"
