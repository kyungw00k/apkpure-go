.PHONY: help build test lint clean install deps

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the project"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linter"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install binary to ~/.local/bin (use PREFIX=/path to override)"
	@echo "  deps      - Download dependencies"
	@echo "  check     - Run all checks (lint + test + build)"
	@echo "  ci        - Run CI checks (same as GitHub Actions)"

# Build the project
build:
	@echo "🔨 Building project..."
	go build -o bin/apkpure ./cmd/apkpure

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run linter
lint:
	@echo "🔍 Running linter..."
	golangci-lint run --timeout=5m

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f bin/apkpure
	go clean

# Install binary to local bin directory
# Usage: make install
#        PREFIX=/usr/local make install (to install to /usr/local/bin)
install: build
	@echo "📦 Installing apkpure-go..."
	@INSTALL_DIR=$${PREFIX:-$$HOME/.local/bin}; \
	mkdir -p $$INSTALL_DIR; \
	cp bin/apkpure $$INSTALL_DIR/apkpure; \
	chmod +x $$INSTALL_DIR/apkpure; \
	echo "✅ Installed to $$INSTALL_DIR/apkpure"; \
	echo "💡 Make sure $$INSTALL_DIR is in your PATH"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	go mod download
	go mod verify

# Run all checks
check: lint test build
	@echo "✅ All checks passed!"

# Run CI checks (same as GitHub Actions)
ci: deps test lint build
	@echo "✅ CI checks passed!"

# Install golangci-lint if not present
install-lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📦 Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.64.8; \
	else \
		echo "✅ golangci-lint already installed"; \
	fi