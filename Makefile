# Quantum Blockchain Makefile

.PHONY: all build clean test test-unit test-integration test-benchmark lint fmt vet deps docker docker-build docker-up docker-down deploy help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build parameters
BINARY_NAME=quantum-node
BINARY_PATH=./cmd/quantum-node
BUILD_DIR=./build
DOCKER_IMAGE=quantum-blockchain
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
COMMIT=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Colors
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[0;37m
NC=\033[0m # No Color

# Default target
all: deps lint test build

# Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(CYAN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Dependencies
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify

# Build targets
build: deps ## Build the quantum node binary
	@echo "$(BLUE)Building quantum node...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-validator-cli: deps ## Build validator CLI
	@echo "$(BLUE)Building validator CLI...$(NC)"
	@mkdir -p $(BUILD_DIR)/binaries
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/binaries/validator-cli ./cmd/validator-cli
	@echo "$(GREEN)Validator CLI build completed: $(BUILD_DIR)/binaries/validator-cli$(NC)"

build-cross: deps ## Build for multiple platforms
	@echo "$(BLUE)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_PATH)
	
	# Linux ARM64
	@echo "Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(BINARY_PATH)
	
	# macOS AMD64
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_PATH)
	
	# macOS ARM64
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_PATH)
	
	# Windows AMD64
	@echo "Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_PATH)
	
	@echo "$(GREEN)Cross-platform build completed$(NC)"

# Clean targets
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean completed$(NC)"

# Test targets
test: test-unit test-integration ## Run all tests

test-unit: deps ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./tests/unit/...
	@echo "$(GREEN)Unit tests completed$(NC)"

test-integration: deps ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GOTEST) -v -timeout=10m ./tests/integration/...
	@echo "$(GREEN)Integration tests completed$(NC)"

test-benchmark: deps ## Run benchmark tests
	@echo "$(BLUE)Running benchmark tests...$(NC)"
	$(GOTEST) -bench=. -benchmem -run=^# ./tests/unit/...
	@echo "$(GREEN)Benchmark tests completed$(NC)"

coverage: test-unit ## Generate and display coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Code quality targets
lint: deps ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	$(GOLINT) run --timeout=5m
	@echo "$(GREEN)Linting completed$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)Vet completed$(NC)"

# Security targets
security: ## Run security checks
	@echo "$(BLUE)Running security checks...$(NC)"
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...
	@echo "$(GREEN)Security checks completed$(NC)"

# Docker targets
docker: docker-build ## Build Docker image

docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(VERSION)$(NC)"

docker-up: ## Start Docker Compose services
	@echo "$(BLUE)Starting Docker Compose services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Docker Compose services started$(NC)"

docker-down: ## Stop Docker Compose services
	@echo "$(YELLOW)Stopping Docker Compose services...$(NC)"
	docker-compose down
	@echo "$(GREEN)Docker Compose services stopped$(NC)"

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f