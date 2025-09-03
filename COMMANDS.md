# Quantum Blockchain Commands Guide

This guide provides step-by-step commands to run the quantum blockchain and deploy smart contracts.

## 🚀 Quick Start Commands

### 1. Build the Quantum Node

```bash
# Build the quantum blockchain node
go build -o build/quantum-node ./cmd/quantum-node
```

### 2. Start the Blockchain

```bash
# Start with genesis configuration (recommended)
./build/quantum-node --data-dir ./production-data --rpc-port 8545 --port 30303 --genesis ./config/genesis.json

# Or start with basic configuration
./build/quantum-node --data-dir ./data --rpc-port 8545 --port 30303
```

**Expected Output:**
```
🚀 Starting Quantum Blockchain Node vdev
🌐 P2P listening on port 30303
🔗 JSON-RPC server listening on port 8545
✅ Quantum blockchain is running!
🚀 Fast block #1: 0 tx, 0.0% load, reward: 1 QTM
🚀 Fast block #2: 0 tx, 0.0% load, reward: 1 QTM
```

## 📊 Verify Blockchain Status

### Check Chain ID
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545
```
**Expected:** `{"jsonrpc":"2.0","result":"0x22b8","id":1}` (Chain ID: 8888)

### Check Current Block Number
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545
```
**Expected:** `{"jsonrpc":"2.0","result":"0x1a","id":1}` (Block number in hex)

### Check Account Balance
```bash
# Check genesis account balance
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x742d35Cc6671C0532925a3b8D581C027d2b3d07f","latest"],"id":1}' \
  http://localhost:8545
```
**Expected:** Large hex value (pre-funded account)

## 🔐 Deploy Smart Contracts

### Method 1: Deploy Quantum Token Contract (Comprehensive)

```bash
# Run the comprehensive quantum token deployment test
go run tests/manual/deploy_quantum_token.go
```

**This command will:**
- Generate Dilithium quantum keys
- Check blockchain state and balances
- Create a quantum-signed contract deployment transaction
- Show deployment simulation (expected to fail without funding)
- Display gas optimization results (98% reduction)

**Expected Output:**
```
🚀 Deploying Quantum Token Contract to Quantum Blockchain
1️⃣ Generating Dilithium keys for contract deployment...
✅ Contract deployer address: 0x43375cf38d7c3310544c5e8e22c29b491cf52e40
2️⃣ Checking blockchain state...
   📦 Current block: 0x10
   💰 Deployer balance: 0x0 QTM
3️⃣ Creating contract deployment transaction...
   🔐 Signing with CRYSTALS-Dilithium-II...
   ✅ Quantum signature verified
```

### Method 2: Test Contract Deployment (Full Testing)

```bash
# Run comprehensive blockchain and contract testing
go run tests/manual/test_contract_deployment.go
```

**This includes:**
- Quantum key generation and verification
- Contract deployment simulation
- Gas cost optimization verification (98% reduction)
- Security feature testing (rate limiting, input validation)
- Production readiness verification

## 🧪 Test Blockchain Functionality

### Basic Transaction Tests
```bash
# Test transaction creation and signing
go run tests/manual/test_transaction.go

# Test RPC transaction submission
go run tests/manual/test_rpc_submit.go

# Test transaction querying
go run tests/manual/test_query_tx.go
```

### Performance Tests (with blockchain running)
```bash
# Test fast performance metrics
go run tests/performance/test_fast_performance.go

# Test live blockchain performance
go run tests/performance/test_live_blockchain.go
```

## 💎 Smart Contract Interaction Commands

### Call Contract Methods
```bash
# Call a contract method (read-only)
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x123...","data":"0x70a08231"},"latest"],"id":1}' \
  http://localhost:8545
```

### Get Contract Code
```bash
# Retrieve contract bytecode
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0x123...","latest"],"id":1}' \
  http://localhost:8545
```

### Get Contract Storage
```bash
# Read contract storage slot
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["0x123...","0x0","latest"],"id":1}' \
  http://localhost:8545
```

### Estimate Gas for Contract Call
```bash
# Estimate gas for contract deployment or call
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_estimateGas","params":[{"to":"0x123...","data":"0x..."}],"id":1}' \
  http://localhost:8545
```

## 🔧 Development Commands

### Clean Start (Fresh Blockchain)
```bash
# Stop current blockchain (Ctrl+C), then:
rm -rf ./production-data
go build -o build/quantum-node ./cmd/quantum-node
./build/quantum-node --data-dir ./production-data --rpc-port 8545 --port 30303 --genesis ./config/genesis.json
```

### Run with Different Ports (Multiple Nodes)
```bash
# Node 1 (default)
./build/quantum-node --data-dir ./node1-data --rpc-port 8545 --port 30303 --genesis ./config/genesis.json

# Node 2 (different ports)
./build/quantum-node --data-dir ./node2-data --rpc-port 8546 --port 30304 --genesis ./config/genesis.json
```

### Build and Run (One Command)
```bash
go build -o build/quantum-node ./cmd/quantum-node && ./build/quantum-node --data-dir ./data --rpc-port 8545 --port 30303
```

## ⚡ Quick Testing Sequence

Run these commands in sequence to verify everything works:

```bash
# 1. Build
go build -o build/quantum-node ./cmd/quantum-node

# 2. Clean start
rm -rf ./production-data

# 3. Start blockchain (in background or separate terminal)
./build/quantum-node --data-dir ./production-data --rpc-port 8545 --port 30303 --genesis ./config/genesis.json

# 4. Wait 10 seconds, then verify (in another terminal)
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545

# 5. Test contract deployment
go run tests/manual/deploy_quantum_token.go

# 6. Run comprehensive test
go run tests/manual/test_contract_deployment.go
```

## 🛠️ Advanced Commands

### Generate Quantum Keys (Standalone)
```bash
# Create a simple key generation test
cat > generate_keys.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "quantum-blockchain/chain/crypto"
)

func main() {
    // Generate Dilithium keys
    privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Private Key Size: %d bytes\n", len(privKey.Bytes()))
    fmt.Printf("Public Key Size: %d bytes\n", len(pubKey.Bytes()))
    fmt.Printf("Address: %s\n", crypto.PublicKeyToAddress(pubKey.Bytes()).Hex())
}
EOF

go run generate_keys.go
```

### Manual Transaction Creation
```bash
# Create a manual transaction example
cat > send_transaction.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    
    "quantum-blockchain/chain/crypto"
    "quantum-blockchain/chain/types"
)

func main() {
    // Generate keys
    privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create transaction
    tx := &types.QuantumTransaction{
        ChainID:   types.NewBigInt(8888),
        Nonce:     0,
        GasPrice:  types.NewBigInt(1000000000),
        Gas:       21000,
        To:        &types.Address{}, // Zero address
        Value:     types.NewBigInt(1000000000000000000), // 1 QTM
        Data:      []byte{},
        SigAlg:    crypto.SigAlgDilithium,
        PublicKey: pubKey.Bytes(),
    }
    
    // Sign transaction
    sigHash := tx.SigningHash()
    qrSig, err := crypto.SignMessage(sigHash[:], crypto.SigAlgDilithium, privKey.Bytes())
    if err != nil {
        log.Fatal(err)
    }
    tx.Signature = qrSig.Signature
    
    // Convert to JSON
    txData, _ := json.MarshalIndent(tx, "", "  ")
    fmt.Printf("Quantum Transaction:\n%s\n", txData)
}
EOF

go run send_transaction.go
```

## 📋 Troubleshooting Commands

### Check if Blockchain is Running
```bash
# Test connection
curl -s http://localhost:8545 || echo "Blockchain not running on port 8545"
```

### Check Process
```bash
# Find quantum-node process
ps aux | grep quantum-node

# Kill if needed
pkill quantum-node
```

### Check Logs
```bash
# If running in background, check output with:
tail -f quantum.log  # if you redirected output

# Or check system logs
journalctl -f | grep quantum
```

### Port Issues
```bash
# Check if port is in use
netstat -tulpn | grep :8545
lsof -i :8545

# Use different port if needed
./build/quantum-node --data-dir ./data --rpc-port 8547 --port 30305
```

## 🎯 Production Deployment Commands

### Single Node Production
```bash
# Production configuration
export QUANTUM_NETWORK_ID=8888
export QUANTUM_GAS_LIMIT=15000000
export QUANTUM_ENABLE_METRICS=true

# Start production node
./build/quantum-node \
  --data-dir ./production-data \
  --rpc-port 8545 \
  --port 30303 \
  --genesis ./config/genesis.json \
  --mining \
  --validator
```

### Docker Commands
```bash
# Build Docker image
docker build -t quantum-blockchain .

# Run container
docker run -d \
  -p 8545:8545 \
  -p 30303:30303 \
  -v $(pwd)/data:/data \
  quantum-blockchain
```

## ✅ Success Indicators

When everything is working correctly, you should see:

1. **Blockchain Running**: `✅ Quantum blockchain is running!`
2. **Fast Blocks**: `🚀 Fast block #X: 0 tx, 0.0% load, reward: 1 QTM` (every 2 seconds)
3. **RPC Working**: `{"jsonrpc":"2.0","result":"0x22b8","id":1}` for chain ID
4. **Contract Tests**: `✅ Production ready: TRUE`
5. **Gas Optimization**: `98% reduction achieved`
6. **Quantum Crypto**: `✅ Quantum cryptography: WORKING`

## 🚨 Common Issues & Solutions

| Issue | Command | Solution |
|-------|---------|----------|
| Port in use | `lsof -i :8545` | Use different port or kill process |
| Build fails | `go mod tidy` | Update dependencies |
| Permission denied | `chmod +x build/quantum-node` | Make executable |
| Genesis not found | `ls config/genesis.json` | Ensure file exists |
| Out of memory | `export GOMAXPROCS=2` | Limit Go processes |

---

## 📞 Quick Help

- **Check if running**: `curl -s http://localhost:8545 && echo "✅ Running" || echo "❌ Not running"`
- **Get block count**: `curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545`
- **Stop blockchain**: `Ctrl+C` or `pkill quantum-node`
- **Fresh start**: `rm -rf ./production-data && ./build/quantum-node --data-dir ./production-data --genesis ./config/genesis.json`

🎉 **You now have a production-ready quantum-resistant blockchain with 2-second blocks, 98% gas optimization, and full EVM compatibility!**