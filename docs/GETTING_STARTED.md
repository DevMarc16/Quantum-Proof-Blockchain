# 🚀 Getting Started with Quantum Blockchain

## What You Need (FREE - No Payment Required!)

### Prerequisites
- **Go 1.23+** (free from https://golang.org/dl/)
- **Git** (free)
- **Your computer** (Windows/Mac/Linux)

**💡 That's it! No cloud services, no payments, no subscriptions needed to run locally.**

## Quick Start (5 Minutes)

### Step 1: Build the Quantum Node
```bash
# Clone if you haven't already
cd /path/to/quantum

# Build the blockchain node (takes ~30 seconds)
go build -o build/quantum-node ./cmd/quantum-node
```

### Step 2: Start Single Validator (Simplest)
```bash
# Start one validator node
./build/quantum-node --data-dir ./data --rpc-port 8545

# In another terminal, test it works:
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545
```

### Step 3: Start Multi-Validator Network (Recommended)
```bash
# Kill single validator first
pkill -f quantum-node

# Start 3-validator network (our current setup)
./deploy_multi_validators.sh

# Check all validators are working:
curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
curl -s http://localhost:8547 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'  
curl -s http://localhost:8549 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
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