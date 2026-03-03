.PHONY: all build build-prod run cursor test clean docker-build docker-run dev help version

# Variables
BINARY_NAME=gomcp
BINARY_PATH=bin/$(BINARY_NAME)
DOCKER_IMAGE=gomcp-server
DOCKER_TAG=latest

# Version information
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "\
	-X github.com/NP-compete/gomcp/internal/version.Version=$(VERSION) \
	-X github.com/NP-compete/gomcp/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/NP-compete/gomcp/internal/version.BuildTime=$(BUILD_TIME)"

# Build the application (development)
build:
	@echo "Building $(BINARY_NAME) (development)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) ./cmd/server
	@echo "Build complete: $(BINARY_PATH)"

# Build the application (production - optimized)
build-prod:
	@echo "Building $(BINARY_NAME) (production)..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@mkdir -p bin
	@CGO_ENABLED=0 go build \
		$(LDFLAGS) \
		-trimpath \
		-o $(BINARY_PATH) \
		./cmd/server
	@echo "Production build complete: $(BINARY_PATH)"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_PATH)

# Run server for Cursor IDE (HTTP/SSE with compatibility mode)
cursor: build
	@echo "Starting server for Cursor IDE..."
	@echo "Config: HTTP/SSE transport on port 8081 with Cursor compatibility"
	@MCP_TRANSPORT_PROTOCOL=http \
	 MCP_PORT=8081 \
	 CURSOR_COMPATIBLE_SSE=true \
	 ENABLE_AUTH=false \
	 LOG_LEVEL=DEBUG \
	 ./$(BINARY_PATH)

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with short flag
test-short:
	@echo "Running short tests..."
	@go test -v -short ./...

# Run linter
lint:
	@echo "Running linter..."
	@go vet ./...
	@go fmt ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Build with Podman
podman-build:
	@echo "Building Podman image..."
	@podman build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run with Podman
podman-run: podman-build
	@echo "Running Podman container..."
	@podman run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development mode..."
	@chmod +x scripts/run_dev.sh
	@./scripts/run_dev.sh

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ tmp/
	@rm -f coverage.out coverage.html
	@rm -f build-errors.log
	@go clean

# Clean all artifacts including dependencies
clean-all: clean
	@echo "Deep cleaning..."
	@rm -rf vendor/
	@go clean -modcache

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Generate code (if needed)
generate:
	@echo "Generating code..."
	@go generate ./...

# Install development tools
tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@if [ -f "$(BINARY_PATH)" ]; then \
		echo ""; \
		echo "Binary version:"; \
		./$(BINARY_PATH) --version 2>/dev/null || echo "Binary does not support --version flag"; \
	fi

# Docker compose commands
compose-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

compose-down:
	@echo "Stopping services..."
	@docker-compose down

compose-logs:
	@docker-compose logs -f gomcp

compose-ps:
	@docker-compose ps

# Security scan
security-scan:
	@echo "Running security scan..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Benchmark
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application binary (development)"
	@echo "  build-prod     - Build optimized production binary with version info"
	@echo "  run            - Build and run the application"
	@echo "  test           - Run tests with coverage"
	@echo "  test-short     - Run short tests"
	@echo "  lint           - Run linter and formatter"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Build and run Docker container"
	@echo "  podman-build   - Build Podman image"
	@echo "  podman-run     - Build and run Podman container"
	@echo "  compose-up     - Start services with docker-compose"
	@echo "  compose-down   - Stop docker-compose services"
	@echo "  compose-logs   - View docker-compose logs"
	@echo "  compose-ps     - Show docker-compose processes"
	@echo "  dev            - Run in development mode with hot reload"
	@echo "  clean          - Clean build artifacts"
	@echo "  clean-all      - Deep clean including dependencies"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  generate       - Generate code"
	@echo "  tools          - Install development tools"
	@echo "  version        - Show version information"
	@echo "  security-scan  - Run security vulnerability scan"
	@echo "  benchmark      - Run benchmarks"
	@echo "  help           - Show this help message"

# Default target
all: build

