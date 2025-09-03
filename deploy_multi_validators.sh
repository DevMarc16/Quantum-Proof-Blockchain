#!/bin/bash

# Multi-Validator Quantum Blockchain Network Deployment
# This script deploys 3 validators running simultaneously

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting Multi-Validator Quantum Network Deployment${NC}"
echo "=================================================="

# Clean up any existing processes
echo -e "${YELLOW}Cleaning up existing processes...${NC}"
pkill -f "quantum-node" || true
sleep 2

# Build the latest quantum-node binary
echo -e "${BLUE}Building quantum-node binary...${NC}"
go build -o build/quantum-node ./cmd/quantum-node
if [ ! -f "build/quantum-node" ]; then
    echo -e "${RED}Failed to build quantum-node binary${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Binary built successfully${NC}"

# Create data directories for 3 validators
echo -e "${BLUE}Setting up validator data directories...${NC}"
for i in 1 2 3; do
    rm -rf ./validator-$i-data
    mkdir -p ./validator-$i-data
    echo -e "${GREEN}✅ Created data directory for Validator $i${NC}"
done

# Deploy Validator 1 (Primary)
echo -e "${BLUE}🔗 Starting Validator 1 (Primary) on ports RPC:8545, P2P:30303${NC}"
./build/quantum-node --data-dir ./validator-1-data --rpc-port 8545 --port 30303 > validator-1.log 2>&1 &
VALIDATOR1_PID=$!
echo "Validator 1 PID: $VALIDATOR1_PID"

# Wait for Validator 1 to initialize
sleep 5

# Deploy Validator 2 (Secondary)  
echo -e "${BLUE}🔗 Starting Validator 2 (Secondary) on ports RPC:8547, P2P:30304${NC}"
./build/quantum-node --data-dir ./validator-2-data --rpc-port 8547 --port 30304 > validator-2.log 2>&1 &
VALIDATOR2_PID=$!
echo "Validator 2 PID: $VALIDATOR2_PID"

# Deploy Validator 3 (Tertiary)
echo -e "${BLUE}🔗 Starting Validator 3 (Tertiary) on ports RPC:8549, P2P:30305${NC}"
./build/quantum-node --data-dir ./validator-3-data --rpc-port 8549 --port 30305 > validator-3.log 2>&1 &
VALIDATOR3_PID=$!
echo "Validator 3 PID: $VALIDATOR3_PID"

echo -e "${GREEN}🎉 All validators started! PIDs: $VALIDATOR1_PID, $VALIDATOR2_PID, $VALIDATOR3_PID${NC}"

# Wait for network to initialize
echo -e "${BLUE}Waiting for network initialization...${NC}"
sleep 10

# Test network connectivity
echo -e "${BLUE}Testing network connectivity...${NC}"

echo "Testing Validator 1 (Port 8545):"
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545 | jq .

echo -e "\nTesting Validator 2 (Port 8547):"
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8547 | jq .

echo -e "\nTesting Validator 3 (Port 8549):"
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8549 | jq .

# Monitor block heights
echo -e "\n${BLUE}Monitoring block heights across validators...${NC}"
for round in 1 2 3 4 5; do
    echo -e "\n--- Round $round ---"
    
    # Get block height from each validator
    HEIGHT1=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
      http://localhost:8545 | jq -r '.result' | xargs printf "%d\n" 2>/dev/null || echo "0")
    
    HEIGHT2=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
      http://localhost:8547 | jq -r '.result' | xargs printf "%d\n" 2>/dev/null || echo "0")
    
    HEIGHT3=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
      http://localhost:8549 | jq -r '.result' | xargs printf "%d\n" 2>/dev/null || echo "0")
    
    echo "Validator 1 Height: $HEIGHT1"
    echo "Validator 2 Height: $HEIGHT2" 
    echo "Validator 3 Height: $HEIGHT3"
    
    # Check if validators are in sync
    if [ "$HEIGHT1" -eq "$HEIGHT2" ] && [ "$HEIGHT2" -eq "$HEIGHT3" ] && [ "$HEIGHT1" -gt 0 ]; then
        echo -e "${GREEN}✅ All validators are in sync at height $HEIGHT1${NC}"
    else
        echo -e "${YELLOW}⚠️  Validators syncing... Heights: $HEIGHT1, $HEIGHT2, $HEIGHT3${NC}"
    fi
    
    sleep 5
done

echo -e "\n${GREEN}🎯 Multi-Validator Network Deployment Complete!${NC}"
echo "=============================================="
echo -e "${BLUE}Network Status:${NC}"
echo "• Validator 1: http://localhost:8545 (P2P: 30303) - PID: $VALIDATOR1_PID"
echo "• Validator 2: http://localhost:8547 (P2P: 30304) - PID: $VALIDATOR2_PID" 
echo "• Validator 3: http://localhost:8549 (P2P: 30305) - PID: $VALIDATOR3_PID"
echo ""
echo -e "${BLUE}Monitoring:${NC}"
echo "• tail -f validator-1.log (Validator 1 logs)"
echo "• tail -f validator-2.log (Validator 2 logs)"
echo "• tail -f validator-3.log (Validator 3 logs)"
echo ""
echo -e "${BLUE}Network Testing:${NC}"
echo "• go run tests/performance/test_multi_validator_consensus.go"
echo "• go run tests/manual/test_multi_validator_transactions.go"
echo ""
echo -e "${GREEN}Multi-validator quantum blockchain network is now running!${NC}"
echo -e "${YELLOW}Press Ctrl+C to monitor logs or run 'pkill -f quantum-node' to stop${NC}"

# Keep script running to show logs
trap 'echo -e "\n${RED}Shutting down validators...${NC}"; kill $VALIDATOR1_PID $VALIDATOR2_PID $VALIDATOR3_PID 2>/dev/null; exit 0' INT

# Show live logs from all validators
echo -e "\n${BLUE}Live logs from all validators (Ctrl+C to exit):${NC}"
tail -f validator-1.log validator-2.log validator-3.log