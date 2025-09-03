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
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
COMMIT=$(shell git rev-parse --short HEAD)

# Linker flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Colors for terminal output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

all: deps lint test build

# Help target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Dependencies
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify

# Build targets
build: deps ## Build the quantum node binary
	@echo "$(BLUE)Building quantum node...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if [ ! -f "$(BINARY_PATH)/main.go" ]; then \
		mkdir -p $(BINARY_PATH); \
		cat > $(BINARY_PATH)/main.go << 'EOF'; \
package main\
\
import (\
	"context"\
	"fmt"\
	"log"\
	"os"\
	"os/signal"\
	"syscall"\
\
	"quantum-blockchain/chain/node"\
\
	"github.com/spf13/cobra"\
	"github.com/spf13/viper"\
)\
\
var (\
	Version   = "dev"\
	BuildTime = "unknown"\
	Commit    = "unknown"\
)\
\
var rootCmd = &cobra.Command{\
	Use:   "quantum-node",\
	Short: "Quantum-resistant blockchain node",\
	Long:  "A quantum-resistant blockchain node with EVM compatibility",\
	Run:   runNode,\
}\
\
func init() {\
	cobra.OnInitialize(initConfig)\
	\
	rootCmd.PersistentFlags().String("config", "", "config file")\
	rootCmd.PersistentFlags().String("data-dir", "./data", "data directory")\
	rootCmd.PersistentFlags().Uint64("network-id", 8888, "network identifier")\
	rootCmd.PersistentFlags().String("listen-addr", "0.0.0.0:30303", "listen address")\
	rootCmd.PersistentFlags().Int("http-port", 8545, "HTTP-RPC server listening port")\
	rootCmd.PersistentFlags().Int("ws-port", 8546, "WS-RPC server listening port")\
	rootCmd.PersistentFlags().StringSlice("bootstrap-peers", []string{}, "bootstrap peers")\
	rootCmd.PersistentFlags().Bool("mining", false, "enable mining")\
	rootCmd.PersistentFlags().Bool("validator", false, "enable validator mode")\
	\
	viper.BindPFlags(rootCmd.PersistentFlags())\
}\
\
func initConfig() {\
	if cfgFile := viper.GetString("config"); cfgFile != "" {\
		viper.SetConfigFile(cfgFile)\
	} else {\
		viper.AddConfigPath(".")\
		viper.AddConfigPath("./configs")\
		viper.SetConfigType("json")\
		viper.SetConfigName("default")\
	}\
	\
	viper.AutomaticEnv()\
	viper.ReadInConfig()\
}\
\
func runNode(cmd *cobra.Command, args []string) {\
	fmt.Printf("Quantum Node %s (built %s, commit %s)\n", Version, BuildTime, Commit)\
	\
	config := &node.Config{\
		DataDir:         viper.GetString("data-dir"),\
		NetworkID:       viper.GetUint64("network-id"),\
		ListenAddr:      viper.GetString("listen-addr"),\
		HTTPPort:        viper.GetInt("http-port"),\
		WSPort:          viper.GetInt("ws-port"),\
		BootstrapPeers:  viper.GetStringSlice("bootstrap-peers"),\
		Mining:          viper.GetBool("mining"),\
	}\
	\
	node, err := node.NewNode(config)\
	if err != nil {\
		log.Fatalf("Failed to create node: %v", err)\
	}\
	\
	err = node.Start()\
	if err != nil {\
		log.Fatalf("Failed to start node: %v", err)\
	}\
	\
	sigCh := make(chan os.Signal, 1)\
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)\
	\
	log.Println("Node started. Press Ctrl+C to stop...")\
	<-sigCh\
	\
	log.Println("Shutting down node...")\
	node.Stop()\
}\
\
func main() {\
	if err := rootCmd.Execute(); err != nil {\
		fmt.Println(err)\
		os.Exit(1)\
	}\
}\
EOF; \
	fi
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

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

# Deployment targets
deploy: ## Deploy using deployment script
	@echo "$(BLUE)Running deployment...$(NC)"
	./scripts/deploy.sh deploy
	@echo "$(GREEN)Deployment completed$(NC)"

deploy-build: ## Build components for deployment
	@echo "$(BLUE)Building components for deployment...$(NC)"
	./scripts/deploy.sh build
	@echo "$(GREEN)Build completed$(NC)"

deploy-status: ## Show deployment status
	./scripts/deploy.sh status

deploy-clean: ## Clean up deployment
	@echo "$(YELLOW)Cleaning up deployment...$(NC)"
	./scripts/deploy.sh cleanup
	@echo "$(GREEN)Cleanup completed$(NC)"

# Development targets
dev-setup: deps ## Set up development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	
	# Install development tools
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	
	# Create directories
	@mkdir -p data logs configs build
	
	# Create default config if it doesn't exist
	@if [ ! -f "configs/default.json" ]; then \
		echo "Creating default config..."; \
		mkdir -p configs; \
		cat > configs/default.json << 'EOF'; \
{\
  "networkId": 8888,\
  "dataDir": "./data",\
  "listenAddr": "0.0.0.0:30303",\
  "httpPort": 8545,\
  "wsPort": 8546,\
  "bootstrapPeers": [],\
  "mining": false,\
  "gasLimit": 15000000,\
  "gasPrice": "1000000000"\
}\
EOF; \
	fi
	
	@echo "$(GREEN)Development environment setup completed$(NC)"

dev-run: build ## Run node in development mode
	@echo "$(BLUE)Starting quantum node in development mode...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) --config configs/default.json --mining

dev-run-validator: build ## Run node as validator
	@echo "$(BLUE)Starting quantum node as validator...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) --config configs/default.json --mining --validator

# Maintenance targets
update-deps: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

check-deps: ## Check for dependency vulnerabilities
	@echo "$(BLUE)Checking dependencies for vulnerabilities...$(NC)"
	@if ! command -v nancy >/dev/null 2>&1; then \
		echo "Installing nancy..."; \
		$(GOCMD) install github.com/sonatypecommunity/nancy@latest; \
	fi
	$(GOCMD) list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)Dependency check completed$(NC)"

mod-verify: ## Verify module dependencies
	@echo "$(BLUE)Verifying module dependencies...$(NC)"
	$(GOMOD) verify
	@echo "$(GREEN)Module verification completed$(NC)"

# Release targets
release-prepare: ## Prepare for release
	@echo "$(BLUE)Preparing for release...$(NC)"
	@$(MAKE) clean
	@$(MAKE) deps
	@$(MAKE) lint
	@$(MAKE) test
	@$(MAKE) security
	@$(MAKE) build-cross
	@echo "$(GREEN)Release preparation completed$(NC)"

# Documentation targets
docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	@if ! command -v godoc >/dev/null 2>&1; then \
		echo "Installing godoc..."; \
		$(GOCMD) install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Visit http://localhost:6060/pkg/quantum-blockchain/ to view docs"
	godoc -http=:6060

# Utility targets
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT)"

env: ## Show environment information
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Go path: $(shell $(GOCMD) env GOPATH)"
	@echo "Go root: $(shell $(GOCMD) env GOROOT)"
	@echo "OS/Arch: $(shell $(GOCMD) env GOOS)/$(shell $(GOCMD) env GOARCH)"

# Quick development workflow
quick: fmt vet test-unit build ## Quick development check (fmt, vet, test, build)

# Full CI workflow
ci: clean deps lint security test build ## Full CI workflow

# Installation target
install: build ## Install binary to GOPATH/bin
	@echo "$(BLUE)Installing quantum-node...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(shell $(GOCMD) env GOPATH)/bin/
	@echo "$(GREEN)quantum-node installed to $(shell $(GOCMD) env GOPATH)/bin/$(NC)"

# Uninstall target
uninstall: ## Remove installed binary
	@echo "$(YELLOW)Removing quantum-node...$(NC)"
	@rm -f $(shell $(GOCMD) env GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)quantum-node removed$(NC)"