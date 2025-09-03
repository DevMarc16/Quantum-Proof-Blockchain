#!/bin/bash

# Quantum Blockchain Multi-Validator Network Deployment Script
# This script deploys a production-ready quantum blockchain network

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$DEPLOY_DIR/.." && pwd)"

NETWORK_NAME="${NETWORK_NAME:-quantum-mainnet}"
ENVIRONMENT="${ENVIRONMENT:-production}"
NUM_VALIDATORS="${NUM_VALIDATORS:-3}"
ENABLE_MONITORING="${ENABLE_MONITORING:-true}"
ENABLE_BACKUP="${ENABLE_BACKUP:-true}"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if running as root or with sudo
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root for security reasons"
        exit 1
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if user is in docker group
    if ! groups $USER | grep -q '\bdocker\b'; then
        print_error "User $USER is not in the docker group. Please add yourself to the docker group."
        exit 1
    fi
    
    # Check if Go is installed (for building)
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.19 or later."
        exit 1
    fi
    
    # Check available disk space (at least 1TB recommended)
    available_space=$(df / | awk 'NR==2{printf "%.0f", $4/1024/1024}')
    if [[ $available_space -lt 500 ]]; then
        print_warning "Available disk space is ${available_space}GB. At least 500GB recommended for production."
    fi
    
    # Check available memory (at least 16GB recommended)
    available_memory=$(free -g | awk 'NR==2{printf "%.0f", $2}')
    if [[ $available_memory -lt 16 ]]; then
        print_warning "Available memory is ${available_memory}GB. At least 16GB recommended for production."
    fi
    
    print_success "Prerequisites check completed"
}

# Function to create directory structure
setup_directories() {
    print_status "Setting up directory structure..."
    
    # Create data directories
    sudo mkdir -p /data/quantum/{validator-{1..3},seed-{1..2}}/{data,logs,keys}
    sudo mkdir -p /var/log/quantum
    
    # Create configuration directories
    mkdir -p "$DEPLOY_DIR"/config/{validator-{1..3},seed-{1..2}}
    mkdir -p "$DEPLOY_DIR"/keys/{validator-{1..3},seed-{1..2}}
    mkdir -p "$DEPLOY_DIR"/monitoring/{prometheus,grafana/{provisioning,dashboards}}
    mkdir -p "$DEPLOY_DIR"/nginx/{ssl,conf.d}
    mkdir -p "$DEPLOY_DIR"/backup/scripts
    
    # Set permissions
    sudo chown -R $USER:$USER /data/quantum
    sudo chown -R $USER:$USER /var/log/quantum
    
    print_success "Directory structure created"
}

# Function to build the quantum blockchain binary
build_quantum_node() {
    print_status "Building quantum blockchain node..."
    
    cd "$ROOT_DIR"
    
    # Clean previous builds
    rm -f build/quantum-node
    
    # Build the node
    go build -ldflags="-X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o build/quantum-node ./cmd/quantum-node
    
    if [[ ! -f "build/quantum-node" ]]; then
        print_error "Failed to build quantum-node binary"
        exit 1
    fi
    
    print_success "Quantum node binary built successfully"
}

# Function to build Docker image
build_docker_image() {
    print_status "Building Docker image..."
    
    cd "$ROOT_DIR"
    
    # Create Dockerfile if it doesn't exist
    if [[ ! -f "Dockerfile" ]]; then
        create_dockerfile
    fi
    
    # Build the image
    docker build -t quantum-blockchain:latest .
    
    print_success "Docker image built successfully"
}

# Function to create Dockerfile
create_dockerfile() {
    print_status "Creating Dockerfile..."
    
    cat > "$ROOT_DIR/Dockerfile" << 'EOF'
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev linux-headers

# Set working directory
WORKDIR /src

# Copy source code
COPY . .

# Build the application
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o quantum-node ./cmd/quantum-node

FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl jq
RUN adduser -D -s /bin/sh quantum

# Create directories
RUN mkdir -p /var/lib/quantum /var/log/quantum /etc/quantum
RUN chown -R quantum:quantum /var/lib/quantum /var/log/quantum /etc/quantum

# Copy binary
COPY --from=builder /src/quantum-node /usr/local/bin/quantum-node
RUN chmod +x /usr/local/bin/quantum-node

# Copy configuration templates
COPY deploy/config/ /etc/quantum/

# Switch to quantum user
USER quantum

# Expose ports
EXPOSE 8545 8546 30303 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:9090/health || exit 1

# Default command
CMD ["quantum-node", "--config", "/etc/quantum/config.yaml"]
EOF
    
    print_success "Dockerfile created"
}

# Function to generate validator keys
generate_validator_keys() {
    print_status "Generating validator keys..."
    
    cd "$ROOT_DIR"
    
    for i in $(seq 1 $NUM_VALIDATORS); do
        print_status "Generating keys for validator-$i..."
        
        # Generate keys using the quantum node
        ./build/quantum-node generate-key --output "$DEPLOY_DIR/keys/validator-$i/validator.key" --algorithm dilithium
        
        # Generate address from key
        address=$(./build/quantum-node address-from-key --key "$DEPLOY_DIR/keys/validator-$i/validator.key")
        echo "$address" > "$DEPLOY_DIR/keys/validator-$i/address.txt"
        
        print_success "Keys generated for validator-$i (Address: $address)"
    done
}

# Function to create genesis configuration
create_genesis_config() {
    print_status "Creating genesis configuration..."
    
    # Generate genesis configuration with validator addresses
    validator_addresses=()
    for i in $(seq 1 $NUM_VALIDATORS); do
        if [[ -f "$DEPLOY_DIR/keys/validator-$i/address.txt" ]]; then
            address=$(cat "$DEPLOY_DIR/keys/validator-$i/address.txt")
            validator_addresses+=("$address")
        fi
    done
    
    # Create genesis.json
    cat > "$DEPLOY_DIR/genesis.json" << EOF
{
  "config": {
    "chainId": 8888,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0
  },
  "difficulty": "0x1",
  "gasLimit": "0x47b760",
  "alloc": {
EOF
    
    # Add initial allocations
    for i in "${!validator_addresses[@]}"; do
        if [[ $i -gt 0 ]]; then
            echo "," >> "$DEPLOY_DIR/genesis.json"
        fi
        echo "    \"${validator_addresses[$i]}\": {" >> "$DEPLOY_DIR/genesis.json"
        echo "      \"balance\": \"0xd3c21bcecceda1000000\"" >> "$DEPLOY_DIR/genesis.json" # 1M QTM
        echo -n "    }" >> "$DEPLOY_DIR/genesis.json"
    done
    
    cat >> "$DEPLOY_DIR/genesis.json" << EOF

  },
  "validators": [
EOF
    
    # Add validator definitions
    for i in "${!validator_addresses[@]}"; do
        if [[ $i -gt 0 ]]; then
            echo "," >> "$DEPLOY_DIR/genesis.json"
        fi
        echo "    {" >> "$DEPLOY_DIR/genesis.json"
        echo "      \"address\": \"${validator_addresses[$i]}\"," >> "$DEPLOY_DIR/genesis.json"
        echo "      \"stake\": \"0xd3c21bcecceda1000000\"" >> "$DEPLOY_DIR/genesis.json" # 1M QTM
        echo -n "    }" >> "$DEPLOY_DIR/genesis.json"
    done
    
    cat >> "$DEPLOY_DIR/genesis.json" << EOF

  ]
}
EOF
    
    print_success "Genesis configuration created"
}

# Function to create validator configurations
create_validator_configs() {
    print_status "Creating validator configurations..."
    
    for i in $(seq 1 $NUM_VALIDATORS); do
        validator_name="validator-$i"
        
        # Get validator address
        address=$(cat "$DEPLOY_DIR/keys/$validator_name/address.txt" 2>/dev/null || echo "")
        
        # Create validator-specific configuration
        cat > "$DEPLOY_DIR/config/$validator_name/config.yaml" << EOF
# Quantum Blockchain Validator Configuration

# Network Configuration
network:
  chain_id: 8888
  network_id: 8888
  listen_addr: "0.0.0.0:30303"
  bootstrap_peers:
$(for j in $(seq 1 $NUM_VALIDATORS); do
    if [[ $j -ne $i ]]; then
        echo "    - \"validator-$j.quantum.network:30303\""
    fi
done)
    - "seed-1.quantum.network:30303"
    - "seed-2.quantum.network:30303"

# Validator Configuration
validator:
  enabled: true
  address: "$address"
  key_path: "/var/lib/quantum/keys/validator.key"
  mining: true

# RPC Configuration
rpc:
  http_addr: "0.0.0.0:8545"
  ws_addr: "0.0.0.0:8546"
  cors_origins: ["*"]
  
# Storage Configuration
storage:
  data_dir: "/var/lib/quantum/data"
  
# Logging Configuration
logging:
  level: "info"
  file: "/var/log/quantum/node.log"
  
# Monitoring Configuration
monitoring:
  metrics_addr: "0.0.0.0:9090"
  health_addr: "0.0.0.0:9090"
EOF
        
        print_success "Configuration created for $validator_name"
    done
}

# Function to create monitoring configurations
create_monitoring_configs() {
    if [[ "$ENABLE_MONITORING" != "true" ]]; then
        return
    fi
    
    print_status "Creating monitoring configurations..."
    
    # Create Prometheus configuration
    cat > "$DEPLOY_DIR/monitoring/prometheus.yml" << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "quantum_alerts.yml"

scrape_configs:
  - job_name: 'quantum-validators'
    static_configs:
$(for i in $(seq 1 $NUM_VALIDATORS); do
    echo "      - targets: ['validator-$i.quantum.network:9090']"
    echo "        labels:"
    echo "          validator: 'validator-$i'"
done)

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
EOF
    
    # Create Grafana provisioning
    mkdir -p "$DEPLOY_DIR/monitoring/grafana/provisioning/datasources"
    mkdir -p "$DEPLOY_DIR/monitoring/grafana/provisioning/dashboards"
    
    cat > "$DEPLOY_DIR/monitoring/grafana/provisioning/datasources/datasources.yml" << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
EOF
    
    print_success "Monitoring configurations created"
}

# Function to create SSL certificates
create_ssl_certificates() {
    print_status "Creating SSL certificates..."
    
    # Create self-signed certificate for development
    if [[ ! -f "$DEPLOY_DIR/nginx/ssl/quantum.crt" ]]; then
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "$DEPLOY_DIR/nginx/ssl/quantum.key" \
            -out "$DEPLOY_DIR/nginx/ssl/quantum.crt" \
            -subj "/C=US/ST=CA/L=SF/O=Quantum/OU=Blockchain/CN=quantum.network"
        
        print_success "SSL certificates created"
    else
        print_status "SSL certificates already exist"
    fi
}

# Function to create Nginx configuration
create_nginx_config() {
    print_status "Creating Nginx configuration..."
    
    cat > "$DEPLOY_DIR/nginx/nginx.conf" << EOF
events {
    worker_connections 1024;
}

http {
    upstream quantum_rpc {
$(for i in $(seq 1 $NUM_VALIDATORS); do
    echo "        server validator-$i.quantum.network:8545;"
done)
    }
    
    upstream quantum_ws {
$(for i in $(seq 1 $NUM_VALIDATORS); do
    echo "        server validator-$i.quantum.network:8546;"
done)
    }
    
    server {
        listen 80;
        server_name quantum.network;
        return 301 https://\$server_name\$request_uri;
    }
    
    server {
        listen 443 ssl http2;
        server_name quantum.network;
        
        ssl_certificate /etc/nginx/ssl/quantum.crt;
        ssl_certificate_key /etc/nginx/ssl/quantum.key;
        
        # RPC endpoint
        location /rpc {
            proxy_pass http://quantum_rpc;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
        }
        
        # WebSocket endpoint
        location /ws {
            proxy_pass http://quantum_ws;
            proxy_http_version 1.1;
            proxy_set_header Upgrade \$http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host \$host;
        }
        
        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
EOF
    
    print_success "Nginx configuration created"
}

# Function to deploy the network
deploy_network() {
    print_status "Deploying quantum blockchain network..."
    
    cd "$DEPLOY_DIR"
    
    # Start the network
    docker-compose up -d
    
    # Wait for services to be healthy
    print_status "Waiting for services to become healthy..."
    sleep 30
    
    # Check service status
    for i in $(seq 1 $NUM_VALIDATORS); do
        service_name="validator-$i"
        if docker-compose ps | grep -q "$service_name.*healthy"; then
            print_success "$service_name is healthy"
        else
            print_warning "$service_name may not be healthy yet"
        fi
    done
    
    print_success "Network deployment completed"
}

# Function to run post-deployment tests
run_post_deployment_tests() {
    print_status "Running post-deployment tests..."
    
    # Test RPC connectivity
    for i in $(seq 1 $NUM_VALIDATORS); do
        port=$((8544 + i))
        if curl -s -f "http://localhost:$port/health" > /dev/null; then
            print_success "Validator-$i RPC is responding"
        else
            print_error "Validator-$i RPC is not responding"
        fi
    done
    
    # Test consensus
    print_status "Checking consensus..."
    sleep 10
    
    # Get block heights from all validators
    block_heights=()
    for i in $(seq 1 $NUM_VALIDATORS); do
        port=$((8544 + i))
        height=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "http://localhost:$port" | jq -r '.result' | xargs printf "%d\n" 2>/dev/null || echo "0")
        block_heights+=($height)
        print_status "Validator-$i block height: $height"
    done
    
    # Check if all validators are in sync (within 2 blocks)
    max_height=$(printf '%s\n' "${block_heights[@]}" | sort -nr | head -1)
    min_height=$(printf '%s\n' "${block_heights[@]}" | sort -n | head -1)
    height_diff=$((max_height - min_height))
    
    if [[ $height_diff -le 2 ]]; then
        print_success "All validators are in sync (height difference: $height_diff)"
    else
        print_warning "Validators may be out of sync (height difference: $height_diff)"
    fi
    
    print_success "Post-deployment tests completed"
}

# Function to display network information
display_network_info() {
    print_status "Network deployment summary:"
    echo ""
    echo "Network Name: $NETWORK_NAME"
    echo "Environment: $ENVIRONMENT"
    echo "Number of Validators: $NUM_VALIDATORS"
    echo "Monitoring Enabled: $ENABLE_MONITORING"
    echo ""
    echo "Validator Endpoints:"
    for i in $(seq 1 $NUM_VALIDATORS); do
        port=$((8544 + i))
        echo "  Validator-$i: http://localhost:$port"
    done
    echo ""
    if [[ "$ENABLE_MONITORING" == "true" ]]; then
        echo "Monitoring Endpoints:"
        echo "  Grafana: http://localhost:3000 (admin/quantum_admin_2024)"
        echo "  Prometheus: http://localhost:9093"
        echo ""
    fi
    echo "Load Balancer: https://localhost (requires DNS setup)"
    echo ""
    print_success "Quantum Blockchain Network is running!"
}

# Main deployment function
main() {
    print_status "Starting Quantum Blockchain Network Deployment"
    print_status "============================================="
    
    check_prerequisites
    setup_directories
    build_quantum_node
    build_docker_image
    generate_validator_keys
    create_genesis_config
    create_validator_configs
    
    if [[ "$ENABLE_MONITORING" == "true" ]]; then
        create_monitoring_configs
    fi
    
    create_ssl_certificates
    create_nginx_config
    deploy_network
    run_post_deployment_tests
    display_network_info
    
    print_success "Deployment completed successfully!"
}

# Handle command line arguments
case "${1:-}" in
    "")
        main
        ;;
    "clean")
        print_status "Cleaning up deployment..."
        cd "$DEPLOY_DIR"
        docker-compose down -v
        docker rmi quantum-blockchain:latest 2>/dev/null || true
        sudo rm -rf /data/quantum
        rm -rf "$DEPLOY_DIR"/{config,keys,monitoring,nginx}
        print_success "Cleanup completed"
        ;;
    "status")
        cd "$DEPLOY_DIR"
        docker-compose ps
        ;;
    "logs")
        cd "$DEPLOY_DIR"
        docker-compose logs -f "${2:-}"
        ;;
    "restart")
        cd "$DEPLOY_DIR"
        docker-compose restart "${2:-}"
        ;;
    *)
        echo "Usage: $0 [clean|status|logs|restart]"
        echo ""
        echo "Options:"
        echo "  (none)   - Deploy the network"
        echo "  clean    - Clean up the deployment"
        echo "  status   - Show service status"
        echo "  logs     - Show logs (optionally for specific service)"
        echo "  restart  - Restart services (optionally specific service)"
        exit 1
        ;;
esac