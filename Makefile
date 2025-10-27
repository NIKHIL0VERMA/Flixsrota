BINARY_NAME=flixsrota
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Go build flags
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED=0

# Build directory
BUILD_DIR=build
DIST_DIR=dist

# Default target
.PHONY: all
all: clean build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@go clean -cache

# Build for current platform
.PHONY: build
build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/flixsrota

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux x86_64
	@echo "Building for linux/amd64..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/flixsrota
	
	# Linux ARM64
	@echo "Building for linux/arm64..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/flixsrota
	
	# macOS x86_64
	@echo "Building for darwin/amd64..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/flixsrota
	
	# macOS ARM64 (Apple Silicon)
	@echo "Building for darwin/arm64..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/flixsrota
	
	# Windows x86_64
	@echo "Building for windows/amd64..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/flixsrota
	
	# Windows ARM64
	@echo "Building for windows/arm64..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/flixsrota
	
	@echo "Build complete! Binaries available in $(DIST_DIR)/"

# Build for specific platform
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/flixsrota
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/flixsrota

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/flixsrota
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/flixsrota

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/flixsrota
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/flixsrota

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate protobuf code
.PHONY: proto
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/flixsrota.proto

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run the application
.PHONY: run
run: build
	@echo "Running Flixsrota..."
	@./$(BUILD_DIR)/$(BINARY_NAME) serve

# Run with configuration wizard
.PHONY: run-config
run-config: build
	@echo "Running configuration wizard..."
	@./$(BUILD_DIR)/$(BINARY_NAME) config init

# Validate configuration
.PHONY: validate-config
validate-config: build
	@echo "Validating configuration..."
	@./$(BUILD_DIR)/$(BINARY_NAME) config validate

# Create release tarballs
.PHONY: release
release: build-all
	@echo "Creating release tarballs..."
	@cd $(DIST_DIR) && \
	tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && \
	zip $(BINARY_NAME)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe
	@echo "Release tarballs created in $(DIST_DIR)/"

# Show help
.PHONY: help
help:
	@echo "Flixsrota Makefile"
	@echo "=================="
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-all      - Build for all platforms (Linux, macOS, Windows, x86_64, ARM64)"
	@echo "  build-linux    - Build for Linux (x86_64, ARM64)"
	@echo "  build-darwin   - Build for macOS (x86_64, ARM64)"
	@echo "  build-windows  - Build for Windows (x86_64, ARM64)"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  proto          - Generate protobuf code"
	@echo "  lint           - Lint code"
	@echo "  fmt            - Format code"
	@echo "  run            - Build and run the application"
	@echo "  run-config     - Run configuration wizard"
	@echo "  validate-config - Validate configuration"
	@echo "  release        - Create release tarballs"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION        - Version string (default: git describe)"
	@echo "  GOOS           - Target OS (linux, darwin, windows)"
	@echo "  GOARCH         - Target architecture (amd64, arm64)"
	@echo "  CGO_ENABLED    - Enable CGO (default: 0)" 