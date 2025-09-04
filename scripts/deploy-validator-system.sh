#!/bin/bash

# Deploy Validator Onboarding and Token Distribution System
# For Quantum-Resistant Blockchain

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NETWORK=${NETWORK:-"localhost"}
RPC_URL=${RPC_URL:-"http://localhost:8545"}
PRIVATE_KEY=${PRIVATE_KEY:-""}
GAS_LIMIT=${GAS_LIMIT:-"8000000"}
CONTRACTS_DIR="./contracts"
BUILD_DIR="./build"
DOCS_DIR="./docs"

print_header() {
    echo -e "${BLUE}"
    echo "=================================================="
    echo "   Quantum Blockchain Validator System Deployment"
    echo "=================================================="
    echo -e "${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.19 or later."
        exit 1
    fi
    
    # Check if Foundry is installed
    if ! command -v forge &> /dev/null; then
        print_error "Foundry is not installed. Please install Foundry for smart contract deployment."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed. Please install jq for JSON processing."
        exit 1
    fi
    
    print_success "All dependencies found"
}

create_directories() {
    print_info "Creating necessary directories..."
    
    mkdir -p $BUILD_DIR/contracts
    mkdir -p $BUILD_DIR/binaries
    mkdir -p $BUILD_DIR/docs
    mkdir -p ./validator-keys
    mkdir -p ./deployment-logs
    
    print_success "Directories created"
}

compile_contracts() {
    print_info "Compiling smart contracts..."
    
    # Initialize Foundry project if foundry.toml doesn't exist
    if [ ! -f foundry.toml ]; then
        cat > foundry.toml << EOF
[profile.default]
src = "contracts"
out = "build/contracts"
libs = ["lib"]
remappings = []

[rpc_endpoints]
localhost = "http://localhost:8545"
testnet = "https://testnet-rpc.quantum-blockchain.io"
mainnet = "https://rpc.quantum-blockchain.io"
EOF
    fi
    
    # Compile contracts
    forge build --out $BUILD_DIR/contracts
    
    if [ $? -eq 0 ]; then
        print_success "Smart contracts compiled successfully"
    else
        print_error "Smart contract compilation failed"
        exit 1
    fi
}

build_cli_tools() {
    print_info "Building CLI tools..."
    
    # Build validator CLI
    echo "Building validator-cli..."
    go build -o $BUILD_DIR/binaries/validator-cli ./cmd/validator-cli/
    
    if [ $? -eq 0 ]; then
        print_success "validator-cli built successfully"
    else
        print_error "validator-cli build failed"
        exit 1
    fi
    
    # Build quantum node
    echo "Building quantum-node..."
    go build -o $BUILD_DIR/binaries/quantum-node ./cmd/quantum-node/
    
    if [ $? -eq 0 ]; then
        print_success "quantum-node built successfully"
    else
        print_error "quantum-node build failed"
        exit 1
    fi
}

deploy_contracts() {
    print_info "Deploying smart contracts to $NETWORK..."
    
    # Check if blockchain is running
    if ! curl -s $RPC_URL > /dev/null; then
        print_error "Blockchain not accessible at $RPC_URL"
        print_info "Please start the blockchain first: ./deploy_multi_validators.sh"
        exit 1
    fi
    
    # Deploy contracts using Foundry
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    DEPLOYMENT_LOG="./deployment-logs/deployment_$TIMESTAMP.json"
    
    echo "{" > $DEPLOYMENT_LOG
    echo '  "network": "'$NETWORK'",' >> $DEPLOYMENT_LOG
    echo '  "timestamp": "'$TIMESTAMP'",' >> $DEPLOYMENT_LOG
    echo '  "contracts": {' >> $DEPLOYMENT_LOG
    
    # Deploy QTM Token (mock for testing)
    print_info "Deploying QTM Token..."
    QTM_ADDRESS=$(forge create $CONTRACTS_DIR/QTMToken.sol:QTMToken \
        --rpc-url $RPC_URL \
        --constructor-args "Quantum Token" "QTM" 1000000000000000000000000000 \
        --json | jq -r '.deployedTo' 2>/dev/null || echo "")
    
    if [ -n "$QTM_ADDRESS" ]; then
        print_success "QTM Token deployed to: $QTM_ADDRESS"
        echo '    "QTMToken": "'$QTM_ADDRESS'",' >> $DEPLOYMENT_LOG
    else
        print_error "Failed to deploy QTM Token"
        exit 1
    fi
    
    # Deploy ValidatorRegistry
    print_info "Deploying ValidatorRegistry..."
    VALIDATOR_REGISTRY_ADDRESS=$(forge create $CONTRACTS_DIR/ValidatorRegistry.sol:ValidatorRegistry \
        --rpc-url $RPC_URL \
        --constructor-args $QTM_ADDRESS "0x$(printf '%040x' 1)" \
        --json | jq -r '.deployedTo' 2>/dev/null || echo "")
    
    if [ -n "$VALIDATOR_REGISTRY_ADDRESS" ]; then
        print_success "ValidatorRegistry deployed to: $VALIDATOR_REGISTRY_ADDRESS"
        echo '    "ValidatorRegistry": "'$VALIDATOR_REGISTRY_ADDRESS'",' >> $DEPLOYMENT_LOG
    else
        print_error "Failed to deploy ValidatorRegistry"
        exit 1
    fi
    
    # Deploy TokenDistribution
    print_info "Deploying TokenDistribution..."
    TOKEN_DISTRIBUTION_ADDRESS=$(forge create $CONTRACTS_DIR/TokenDistribution.sol:TokenDistribution \
        --rpc-url $RPC_URL \
        --constructor-args $QTM_ADDRESS "0x$(printf '%040x' 1)" \
        --json | jq -r '.deployedTo' 2>/dev/null || echo "")
    
    if [ -n "$TOKEN_DISTRIBUTION_ADDRESS" ]; then
        print_success "TokenDistribution deployed to: $TOKEN_DISTRIBUTION_ADDRESS"
        echo '    "TokenDistribution": "'$TOKEN_DISTRIBUTION_ADDRESS'",' >> $DEPLOYMENT_LOG
    else
        print_error "Failed to deploy TokenDistribution"
        exit 1
    fi
    
    # Deploy TestnetFaucet
    print_info "Deploying TestnetFaucet..."
    TESTNET_FAUCET_ADDRESS=$(forge create $CONTRACTS_DIR/TestnetFaucet.sol:TestnetFaucet \
        --rpc-url $RPC_URL \
        --constructor-args $QTM_ADDRESS \
        --json | jq -r '.deployedTo' 2>/dev/null || echo "")
    
    if [ -n "$TESTNET_FAUCET_ADDRESS" ]; then
        print_success "TestnetFaucet deployed to: $TESTNET_FAUCET_ADDRESS"
        echo '    "TestnetFaucet": "'$TESTNET_FAUCET_ADDRESS'"' >> $DEPLOYMENT_LOG
    else
        print_error "Failed to deploy TestnetFaucet"
        exit 1
    fi
    
    echo '  }' >> $DEPLOYMENT_LOG
    echo '}' >> $DEPLOYMENT_LOG
    
    print_success "All contracts deployed successfully!"
    print_info "Deployment details saved to: $DEPLOYMENT_LOG"
}

setup_initial_configuration() {
    print_info "Setting up initial configuration..."
    
    # Load deployment addresses
    DEPLOYMENT_LOG=$(ls -t ./deployment-logs/deployment_*.json | head -1)
    QTM_ADDRESS=$(cat $DEPLOYMENT_LOG | jq -r '.contracts.QTMToken')
    VALIDATOR_REGISTRY_ADDRESS=$(cat $DEPLOYMENT_LOG | jq -r '.contracts.ValidatorRegistry')
    TOKEN_DISTRIBUTION_ADDRESS=$(cat $DEPLOYMENT_LOG | jq -r '.contracts.TokenDistribution')
    TESTNET_FAUCET_ADDRESS=$(cat $DEPLOYMENT_LOG | jq -r '.contracts.TestnetFaucet')
    
    # Grant minter role to contracts
    print_info "Configuring token permissions..."
    
    # Create configuration file
    cat > $BUILD_DIR/config.json << EOF
{
  "network": "$NETWORK",
  "rpcUrl": "$RPC_URL",
  "contracts": {
    "QTMToken": "$QTM_ADDRESS",
    "ValidatorRegistry": "$VALIDATOR_REGISTRY_ADDRESS",
    "TokenDistribution": "$TOKEN_DISTRIBUTION_ADDRESS",
    "TestnetFaucet": "$TESTNET_FAUCET_ADDRESS"
  },
  "parameters": {
    "minStake": "100000000000000000000000",
    "maxStake": "10000000000000000000000000",
    "minDelegation": "100000000000000000000",
    "unbondingPeriod": 1814400,
    "blockTime": 2,
    "maxValidators": 100
  }
}
EOF
    
    print_success "Configuration file created: $BUILD_DIR/config.json"
}

create_genesis_validators() {
    print_info "Creating genesis validators..."
    
    # Create 3 genesis validators for initial testing
    for i in {1..3}; do
        VALIDATOR_DIR="./validator-keys/genesis-validator-$i"
        mkdir -p $VALIDATOR_DIR
        
        print_info "Generating keys for genesis validator $i..."
        $BUILD_DIR/binaries/validator-cli \
            -generate \
            -algorithm dilithium \
            -output $VALIDATOR_DIR \
            -mnemonic
        
        if [ $? -eq 0 ]; then
            print_success "Genesis validator $i keys generated"
        else
            print_error "Failed to generate keys for genesis validator $i"
        fi
    done
}

generate_documentation() {
    print_info "Generating deployment documentation..."
    
    # Copy documentation
    cp -r $DOCS_DIR/* $BUILD_DIR/docs/ 2>/dev/null || true
    
    # Generate API documentation
    cat > $BUILD_DIR/docs/API_REFERENCE.md << EOF
# Quantum Blockchain API Reference

## Contract Addresses

$(cat $BUILD_DIR/config.json | jq -r '.contracts | to_entries[] | "- **\(.key)**: \(.value)"')

## RPC Endpoints

- **JSON-RPC**: $RPC_URL
- **WebSocket**: ${RPC_URL/http/ws}

## CLI Commands

### Validator CLI

\`\`\`bash
# Generate validator keys
./validator-cli -generate -algorithm dilithium

# Register as validator
./validator-cli -register -stake 100000 -commission 500

# Check status
./validator-cli -status

# Delegate to validator
./validator-cli -delegate -validator 0x... -amount 1000
\`\`\`

### Node Commands

\`\`\`bash
# Start validator node
./quantum-node --validator --key-file ./keys/dilithium.key

# Start full node
./quantum-node --data-dir ./data --rpc-port 8545
\`\`\`

## Contract Interactions

### ValidatorRegistry

\`\`\`bash
# Register validator
cast send $VALIDATOR_REGISTRY_ADDRESS \\
  "registerValidator(bytes,uint8,uint256,uint256,string)" \\
  \$PUBLIC_KEY 1 \$STAKE \$COMMISSION \$METADATA

# Delegate tokens
cast send $VALIDATOR_REGISTRY_ADDRESS \\
  "delegate(address,uint256)" \\
  \$VALIDATOR_ADDRESS \$AMOUNT
\`\`\`

### TestnetFaucet

\`\`\`bash
# Request testnet tokens
cast send $TESTNET_FAUCET_ADDRESS "requestTokens()"

# Request validator tokens
cast send $TESTNET_FAUCET_ADDRESS \\
  "requestValidatorTokens(bytes,bytes,bytes32)" \\
  \$PUBLIC_KEY \$SIGNATURE \$NONCE
\`\`\`

Generated on: $(date)
EOF
    
    print_success "Documentation generated in $BUILD_DIR/docs/"
}

run_tests() {
    print_info "Running system tests..."
    
    # Run smart contract tests
    if [ -d "./test" ]; then
        print_info "Running smart contract tests..."
        forge test --match-path "./test/*" -vv
    fi
    
    # Run Go tests
    if [ -d "./tests" ]; then
        print_info "Running Go tests..."
        go test ./tests/... -v
    fi
    
    print_success "Tests completed"
}

print_deployment_summary() {
    print_success "🎉 Quantum Blockchain Validator System Deployed Successfully!"
    
    echo -e "${BLUE}"
    echo "=================================================="
    echo "           DEPLOYMENT SUMMARY"
    echo "=================================================="
    echo -e "${NC}"
    
    echo "📋 Contract Addresses:"
    cat $BUILD_DIR/config.json | jq -r '.contracts | to_entries[] | "   \(.key): \(.value)"'
    
    echo ""
    echo "🔧 Built Binaries:"
    echo "   validator-cli: $BUILD_DIR/binaries/validator-cli"
    echo "   quantum-node:  $BUILD_DIR/binaries/quantum-node"
    
    echo ""
    echo "📚 Documentation:"
    echo "   Validator Guide: $BUILD_DIR/docs/VALIDATOR_ONBOARDING.md"
    echo "   API Reference:   $BUILD_DIR/docs/API_REFERENCE.md"
    echo "   Configuration:   $BUILD_DIR/config.json"
    
    echo ""
    echo "🔑 Genesis Validators:"
    for i in {1..3}; do
        if [ -f "./validator-keys/genesis-validator-$i/validator-profile.json" ]; then
            ADDRESS=$(cat "./validator-keys/genesis-validator-$i/validator-profile.json" | jq -r '.config.address')
            echo "   Validator $i: $ADDRESS"
        fi
    done
    
    echo ""
    echo "🚀 Next Steps:"
    echo "   1. Fund genesis validators with QTM tokens"
    echo "   2. Register validators using the CLI"
    echo "   3. Start validator nodes to begin consensus"
    echo "   4. Open faucet for community testing"
    
    echo ""
    echo "📖 Quick Start Commands:"
    echo "   # Check system status"
    echo "   curl $RPC_URL -X POST -H 'Content-Type: application/json' -d '{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"id\":1}'"
    echo ""
    echo "   # Generate validator keys"
    echo "   $BUILD_DIR/binaries/validator-cli -generate -algorithm dilithium"
    echo ""
    echo "   # Request testnet tokens"
    echo "   cast send $(cat $BUILD_DIR/config.json | jq -r '.contracts.TestnetFaucet') 'requestTokens()'"
    
    echo -e "${BLUE}"
    echo "=================================================="
    echo -e "${NC}"
}

main() {
    print_header
    
    # Parse command line arguments
    RUN_TESTS=false
    SKIP_CONTRACTS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --test)
                RUN_TESTS=true
                shift
                ;;
            --skip-contracts)
                SKIP_CONTRACTS=true
                shift
                ;;
            --network)
                NETWORK="$2"
                shift 2
                ;;
            --rpc-url)
                RPC_URL="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --test            Run tests after deployment"
                echo "  --skip-contracts  Skip contract deployment"
                echo "  --network NAME    Target network (default: localhost)"
                echo "  --rpc-url URL     RPC endpoint (default: http://localhost:8545)"
                echo "  --help            Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment steps
    check_dependencies
    create_directories
    compile_contracts
    build_cli_tools
    
    if [ "$SKIP_CONTRACTS" = false ]; then
        deploy_contracts
        setup_initial_configuration
        create_genesis_validators
    fi
    
    generate_documentation
    
    if [ "$RUN_TESTS" = true ]; then
        run_tests
    fi
    
    print_deployment_summary
}

# Handle script termination
cleanup() {
    print_warning "Deployment interrupted"
    exit 1
}

trap cleanup SIGINT SIGTERM

# Run main function
main "$@"