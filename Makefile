.PHONY: build clean test run docker-build docker-run help lint fmt all

# Default Go binary name
BINARY_NAME=aggregator
# Docker image name
DOCKER_IMAGE=content-aggregator

# Build tags
BUILD_TAGS=

# Go build and test flags
GOFLAGS=-v
LDFLAGS=-ldflags "-w -s"
TESTFLAGS=-race -cover

# Output paths
BIN_DIR=./bin
OUTPUT_DIR=./output

# Source files
SRC_DIRS=./cmd ./internal ./pkg
GO_FILES=$(shell find $(SRC_DIRS) -name "*.go" -type f)

# Default target
all: clean build test

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build $(GOFLAGS) $(LDFLAGS) -tags "$(BUILD_TAGS)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/aggregator

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test $(TESTFLAGS) ./...

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)

# Run the application with specific configuration
run-custom: build
	@echo "Running $(BINARY_NAME) with custom config..."
	@$(BIN_DIR)/$(BINARY_NAME) --config=$(CONFIG) --sources=$(SOURCES)

# Run the web interface
run-web: build
	@echo "Running web interface..."
	@$(BIN_DIR)/$(BINARY_NAME) --web --port=8080

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE) .

# Run in Docker
docker-run: docker-build
	@echo "Running in Docker..."
	@docker run -p 8080:8080 -v $(shell pwd)/configs:/app/configs -v $(shell pwd)/output:/app/output $(DOCKER_IMAGE)

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w $(GO_FILES)

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create a new release
release:
	@echo "Creating release..."
	@goreleaser release --rm-dist

# Create initial project structure
init:
	@echo "Creating project structure..."
	@mkdir -p cmd/aggregator internal/fetcher internal/parser internal/model internal/coordinator internal/normalizer internal/aggregator pkg/config pkg/util web/templates web/static/{css,js} configs output
	@echo "package main\n\nfunc main() {\n\t// TODO: Implement\n}" > cmd/aggregator/main.go
	@echo "module github.com/CyberwizD/Concurrent-Web-Content-Aggregator\n\ngo 1.18" > go.mod
	@go mod tidy

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  run           - Build and run the application"
	@echo "  run-web       - Run the web interface"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linters"
	@echo "  deps          - Install dependencies"
	@echo "  init          - Create initial project structure"
	@echo "  release       - Create a new release"
	@echo "  all           - Clean, build, and test"
	@echo "  help          - Show this help"
