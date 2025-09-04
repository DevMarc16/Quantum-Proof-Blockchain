# Getting Started Guide

## 🎯 Complete Setup from Scratch (5 Minutes)

This guide walks you through setting up and running the quantum-resistant blockchain from a fresh start.

### Prerequisites Check
```bash
# Verify Go installation (required: 1.21+)
go version

# Check available memory (recommended: 8GB+)
free -h

# Check disk space (required: 10GB+)
df -h .
```

## Step 1: Build Everything

### Build the Core Components
```bash
# Navigate to project directory
cd quantum-blockchain

# Build the main quantum node
go build -o build/quantum-node cmd/quantum-node/main.go

# Build the validator CLI
go build -o validator-cli cmd/validator-cli/main.go

# Verify builds
ls -la build/quantum-node validator-cli
```

**Expected Output:**
```
-rwxr-xr-x quantum-node
-rwxr-xr-x validator-cli
```

## Step 2: Start the Network

### Deploy Multi-Validator Network
```bash
# Make script executable
chmod +x scripts/deploy_multi_validators.sh

# Start 3-validator network
./scripts/deploy_multi_validators.sh
```

**What Happens:**
- ✅ Builds quantum-node binary
- ✅ Creates validator data directories
- ✅ Starts 3 validators on different ports
- ✅ Begins block production (2-second intervals)
- ✅ Shows live network status

**Expected Output:**
```
🚀 Starting Multi-Validator Quantum Network Deployment
==================================================
✅ Binary built successfully
✅ Created data directory for Validator 1
✅ Created data directory for Validator 2  
✅ Created data directory for Validator 3
🔗 Starting Validator 1 (Primary) on ports RPC:8545, P2P:30303
🔗 Starting Validator 2 (Secondary) on ports RPC:8547, P2P:30304
🔗 Starting Validator 3 (Tertiary) on ports RPC:8549, P2P:30305
🎉 All validators started!
```

## What You Get (Running Locally)

✅ **Full quantum-resistant blockchain**
- CRYSTALS-Dilithium-II signatures (2420 bytes each)
- 2-second block times
- Multi-validator consensus (3 validators)
- EVM-compatible smart contracts

✅ **Complete JSON-RPC API**
- All standard Ethereum methods (eth_*)
- Quantum-specific methods (quantum_*)
- Web3.js compatible

✅ **Testing & Development**
- Deploy smart contracts
- Send quantum transactions
- Full development environment

## Test the Network

### Deploy a Smart Contract
```bash
# With network running, deploy test contract
go run tests/manual/deploy_quantum_token/deploy_quantum_token.go
```

### Send Quantum Transactions
```bash
# Test funded transaction
go run tests/manual/test_final_funded_transaction/test_final_funded_transaction.go
```

### Check Network Status
```bash
# See detailed network status
go run tests/manual/test_network_status/test_network_status.go
```

## What's Optional (Enterprise Features)

The following are **enterprise-ready but NOT required** to run the blockchain:

### 🏢 Enterprise Features (Optional - Costs Money)
- **AWS CloudHSM**: Hardware security for production ($1000s/month)
- **Kubernetes Cluster**: For cloud deployment (varies by provider)
- **Monitoring Services**: Prometheus/Grafana hosting (can run free locally)

### 🆓 Free Alternatives
- **Local Development**: Everything runs on your computer (FREE)
- **Testing Environment**: Full multi-validator network locally (FREE)
- **Smart Contracts**: Deploy and test locally (FREE)

## Common Issues & Solutions

### Build Errors
```bash
# If you get module errors:
go mod tidy
go build -o build/quantum-node ./cmd/quantum-node
```

### Port Conflicts
```bash
# If ports are in use, kill existing processes:
pkill -f quantum-node

# Or use different ports:
./build/quantum-node --rpc-port 8546 --port 30304
```

### Clean Restart
```bash
# Clean everything and restart:
pkill -f quantum-node
rm -rf validator-*-data/ validator-*.log
./deploy_multi_validators.sh
```

## Next Steps

### 1. Explore the Network
- **View Logs**: `tail -f validator-1.log validator-2.log validator-3.log`
- **Check Blocks**: Watch blocks being produced every 2 seconds
- **Test APIs**: Try different JSON-RPC calls

### 2. Smart Contract Development
- Deploy contracts using standard tools
- Test quantum-resistant transactions
- Explore EVM compatibility

### 3. SDK Integration (Optional)
```bash
# Test JavaScript SDK (requires Node.js)
cd sdk/js
npm install  # if packages missing
npm test
```

## Summary: What Costs Money?

### ✅ FREE (Everything You Need)
- Running quantum blockchain locally
- Multi-validator network
- Smart contract development
- All testing and development
- Complete blockchain functionality

### 💰 PAID (Enterprise Production Only)
- AWS CloudHSM for production security
- Cloud hosting (AWS, Google Cloud, etc.)
- Professional monitoring services
- Production-grade infrastructure

**💡 Bottom Line: You can run everything locally for FREE! The paid services are only for large-scale production deployment.**

## Quick Status Check

```bash
# Check if your network is running:
curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' | grep -o '"result":"0x22b8"' && echo "✅ Quantum Blockchain Running!" || echo "❌ Network not running"
```

If you see `✅ Quantum Blockchain Running!`, you're ready to go! 🎉