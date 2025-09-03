#!/bin/bash

# Quantum Blockchain Deployment Script
# This script deploys a quantum-resistant blockchain network

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NETWORK_NAME="quantum-network"
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_warning "Go is not installed. Building from source will not work."
    fi
    
    log_success "Prerequisites check completed"
}

# Create necessary directories
create_directories() {
    log_info "Creating necessary directories..."
    
    mkdir -p data/{bootstrap,node1,node2}
    mkdir -p logs/{bootstrap,node1,node2}
    mkdir -p configs
    mkdir -p infra/nginx/ssl
    mkdir -p monitoring/grafana/{dashboards,provisioning}
    
    log_success "Directories created"
}

# Generate configuration files
generate_configs() {
    log_info "Generating configuration files..."
    
    # Generate default config
    cat > configs/default.json << EOF
{
  "networkId": 8888,
  "dataDir": "/root/data",
  "listenAddr": "0.0.0.0:30303",
  "httpPort": 8545,
  "wsPort": 8546,
  "bootstrapPeers": [],
  "mining": false,
  "gasLimit": 15000000,
  "gasPrice": "1000000000"
}
EOF

    # Generate validator config
    cat > configs/validator.json << EOF
{
  "networkId": 8888,
  "dataDir": "/root/data",
  "listenAddr": "0.0.0.0:30303",
  "httpPort": 8545,
  "wsPort": 8546,
  "bootstrapPeers": [],
  "mining": true,
  "validatorAlg": "Dilithium",
  "gasLimit": 15000000,
  "gasPrice": "1000000000"
}
EOF

    # Generate environment file
    cat > $ENV_FILE << EOF
# Quantum Blockchain Environment Configuration
COMPOSE_PROJECT_NAME=quantum-blockchain
QUANTUM_NETWORK_ID=8888
QUANTUM_CHAIN_ID=8888

# Node configuration
QUANTUM_GAS_LIMIT=15000000
QUANTUM_GAS_PRICE=1000000000
QUANTUM_BLOCK_TIME=12

# Security
QUANTUM_ENABLE_CORS=true
QUANTUM_CORS_ORIGINS=*

# Monitoring
GRAFANA_ADMIN_PASSWORD=quantum123
PROMETHEUS_RETENTION=15d

# Development settings (disable in production)
QUANTUM_DEBUG=false
QUANTUM_METRICS_ENABLED=true
EOF

    log_success "Configuration files generated"
}

# Build the quantum node
build_node() {
    log_info "Building quantum blockchain node..."
    
    if [ ! -f "cmd/quantum-node/main.go" ]; then
        log_info "Creating main.go file..."
        mkdir -p cmd/quantum-node
        
        cat > cmd/quantum-node/main.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"quantum-blockchain/chain/node"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "quantum-node",
	Short: "Quantum-resistant blockchain node",
	Long:  "A quantum-resistant blockchain node with EVM compatibility",
	Run:   runNode,
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.quantum.yaml)")
	rootCmd.PersistentFlags().String("data-dir", "./data", "data directory")
	rootCmd.PersistentFlags().Uint64("network-id", 8888, "network identifier")
	rootCmd.PersistentFlags().String("listen-addr", "0.0.0.0:30303", "listen address")
	rootCmd.PersistentFlags().Int("http-port", 8545, "HTTP-RPC server listening port")
	rootCmd.PersistentFlags().Int("ws-port", 8546, "WS-RPC server listening port")
	rootCmd.PersistentFlags().StringSlice("bootstrap-peers", []string{}, "bootstrap peers")
	rootCmd.PersistentFlags().Bool("mining", false, "enable mining")
	rootCmd.PersistentFlags().Bool("validator", false, "enable validator mode")
	
	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initConfig() {
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("json")
		viper.SetConfigName("default")
	}
	
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func runNode(cmd *cobra.Command, args []string) {
	config := &node.Config{
		DataDir:         viper.GetString("data-dir"),
		NetworkID:       viper.GetUint64("network-id"),
		ListenAddr:      viper.GetString("listen-addr"),
		HTTPPort:        viper.GetInt("http-port"),
		WSPort:          viper.GetInt("ws-port"),
		BootstrapPeers:  viper.GetStringSlice("bootstrap-peers"),
		Mining:          viper.GetBool("mining"),
	}
	
	log.Printf("Starting Quantum Node with config: %+v", config)
	
	node, err := node.NewNode(config)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	
	err = node.Start()
	if err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}
	
	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Node started. Press Ctrl+C to stop...")
	<-sigCh
	
	log.Println("Shutting down node...")
	node.Stop()
	log.Println("Node stopped")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
EOF
        
        log_success "main.go created"
    fi
    
    # Build using Go if available, otherwise rely on Docker build
    if command -v go &> /dev/null; then
        log_info "Building with Go..."
        CGO_ENABLED=1 go build -o quantum-node ./cmd/quantum-node
        log_success "Binary built successfully"
    else
        log_info "Will build inside Docker container"
    fi
}

# Deploy the network
deploy_network() {
    log_info "Deploying quantum blockchain network..."
    
    # Pull/build images
    docker-compose build
    
    # Start the network
    docker-compose up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 30
    
    # Check service status
    if docker-compose ps | grep -q "Up (healthy)"; then
        log_success "Network deployed successfully"
    else
        log_error "Some services are not healthy"
        docker-compose ps
        return 1
    fi
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check if bootstrap node is responding
    if curl -s -X POST -H "Content-Type: application/json" \
       --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
       http://localhost:8545 | grep -q "22b8"; then
        log_success "Bootstrap node is responding"
    else
        log_error "Bootstrap node is not responding"
        return 1
    fi
    
    # Check peer count
    peer_count=$(curl -s -X POST -H "Content-Type: application/json" \
                 --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
                 http://localhost:8545 | grep -o '"result":"0x[0-9]*"' | grep -o '0x[0-9]*' | tail -1)
    
    if [ "$peer_count" != "0x0" ]; then
        log_success "Nodes are connected (peer count: $peer_count)"
    else
        log_warning "No peers connected yet"
    fi
    
    # Check if mining is working
    block_number=$(curl -s -X POST -H "Content-Type: application/json" \
                   --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
                   http://localhost:8545 | grep -o '"result":"0x[0-9a-f]*"' | grep -o '0x[0-9a-f]*')
    
    if [ "$block_number" != "0x0" ]; then
        log_success "Mining is working (current block: $block_number)"
    else
        log_warning "No blocks mined yet"
    fi
}

# Show status
show_status() {
    log_info "Network Status:"
    echo "=================="
    docker-compose ps
    echo ""
    
    log_info "Access Points:"
    echo "HTTP RPC (Bootstrap): http://localhost:8545"
    echo "WebSocket (Bootstrap): ws://localhost:8546"
    echo "HTTP RPC (Node 1): http://localhost:8547"
    echo "HTTP RPC (Node 2): http://localhost:8549"
    echo "Load Balancer: http://localhost/rpc"
    echo "Prometheus: http://localhost:9090"
    echo "Grafana: http://localhost:3000 (admin/quantum123)"
    echo ""
    
    log_info "Sample RPC Call:"
    echo "curl -X POST -H \"Content-Type: application/json\" \\"
    echo "  --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_chainId\",\"params\":[],\"id\":1}' \\"
    echo "  http://localhost:8545"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up deployment..."
    docker-compose down -v
    docker system prune -f
    log_success "Cleanup completed"
}

# Main deployment logic
main() {
    case "${1:-deploy}" in
        "prereq"|"prerequisites")
            check_prerequisites
            ;;
        "build")
            check_prerequisites
            create_directories
            generate_configs
            build_node
            ;;
        "deploy")
            check_prerequisites
            create_directories
            generate_configs
            build_node
            deploy_network
            verify_deployment
            show_status
            ;;
        "status")
            show_status
            ;;
        "verify")
            verify_deployment
            ;;
        "cleanup"|"clean")
            cleanup
            ;;
        "restart")
            docker-compose restart
            verify_deployment
            show_status
            ;;
        *)
            echo "Usage: $0 {deploy|build|status|verify|cleanup|restart}"
            echo ""
            echo "Commands:"
            echo "  deploy    - Full deployment (default)"
            echo "  build     - Build components only"
            echo "  status    - Show network status"
            echo "  verify    - Verify deployment"
            echo "  cleanup   - Clean up deployment"
            echo "  restart   - Restart services"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"