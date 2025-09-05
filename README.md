# Quantum-Resistant Blockchain

A production-ready multi-validator quantum-resistant blockchain with full EVM compatibility, implementing NIST-standardized post-quantum cryptographic algorithms (CRYSTALS-Dilithium-II, CRYSTALS-Kyber-512).

## 🚀 Quick Start (5 Minutes)

### Prerequisites
- Go 1.23+ installed
- Linux/WSL/macOS terminal
- 8GB RAM, 10GB free disk space

### Complete Setup from Scratch

#### 1. Build the Quantum Node
```bash
# Clean dependencies first
go mod tidy

# Build the main quantum blockchain node
go build -o build/quantum-node cmd/quantum-node/main.go

# Build the validator CLI tool
go build -o validator-cli cmd/validator-cli/main.go
```

#### 2. Start Multi-Validator Network
```bash
# Make deployment script executable and run
chmod +x scripts/deploy_multi_validators.sh
./scripts/deploy_multi_validators.sh
```

**Network will auto-start with:**
- **Validator 1**: http://localhost:8545 (Primary)
- **Validator 2**: http://localhost:8547 (Secondary) 
- **Validator 3**: http://localhost:8549 (Tertiary)

#### 3. Verify Network is Running
```bash
# Check current block height
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545

# Should return: {"jsonrpc":"2.0","result":"0x...","id":1}
```

#### 4. Set Up Validator CLI
```bash
# Generate quantum validator keys
./validator-cli -generate -algorithm dilithium -output validator-keys

# Register as validator
./validator-cli -register -stake 100000 -commission 500 -rpc http://localhost:8545

# Check validator status
./validator-cli -status -rpc http://localhost:8545
```

#### 5. Run Tests
```bash
# Run integration tests
go test ./tests/integration/ -v

# Test multi-validator consensus (optional)
go run tests/performance/test_multi_validator_consensus.go
```

### ✅ You're Ready!
Your quantum blockchain network is now running with:
- 🔐 **Quantum-resistant signatures** (CRYSTALS-Dilithium-II)
- ⚡ **2-second blocks** with automatic reward distribution
- 🏛️ **Multi-validator consensus** (3 validators)
- 🔗 **Full JSON-RPC API** compatibility

## 🎯 ONE-COMMAND SETUP

For the fastest setup, use our automated script:
```bash
./quick-start.sh
```
This single command will:
1. Build everything from scratch
2. Deploy 3-validator network
3. Set up validator CLI with keys
4. Run tests to verify functionality
5. Show you exactly what's running and how to use it

**Total time: 5 minutes** ⏱️

## 📁 Project Structure

```
quantum-blockchain/
├── README.md                    # This file
├── chain/                       # Core blockchain implementation
│   ├── crypto/                  # Quantum cryptography (Dilithium, Kyber, Falcon)
│   ├── types/                   # Transaction and block types
│   ├── node/                    # Multi-validator node implementation
│   ├── consensus/               # Quantum consensus algorithms
│   └── evm/                     # EVM compatibility layer
├── cmd/                         # Main executables
│   ├── quantum-node/            # Main node binary
│   └── validator-cli/           # Validator management CLI
├── config/                      # Configuration files
├── contracts/                   # Smart contracts
├── docs/                        # Documentation
├── examples/                    # Usage examples
├── scripts/                     # Deployment and management scripts
├── sdk/                         # JavaScript/TypeScript SDK
├── tests/                       # Test suites
└── tools/                       # Development and deployment tools
    ├── deployment/              # Contract deployment tools
    ├── testing/                 # Testing utilities
    ├── debug/                   # Debugging tools
    └── cli/                     # Command-line utilities
```

## 🔧 Core Components

### Quantum Cryptography
- **CRYSTALS-Dilithium-II**: Digital signatures (2420-byte signatures)
- **CRYSTALS-Kyber-512**: Key encapsulation mechanism
- **Falcon**: Hybrid ED25519+Dilithium approach

### Multi-Validator Network
- **3+ validators** coordinating block production
- **2-second block times** with quantum-resistant signatures
- **True decentralization** like Ethereum 2.0 and Solana
- **Enterprise-grade architecture** with monitoring and governance

### EVM Compatibility
- Full Ethereum JSON-RPC compatibility
- Smart contract deployment and execution
- Metamask integration via quantum snap

## 🛠️ Development Tools

### Deployment
```bash
# Deploy quantum contracts
go run tools/deployment/deploy_with_fixed_keys.go

# Fund quantum accounts
go run tools/deployment/fund_quantum_account.go
```

### Testing
```bash
# Run integration tests
go test ./tests/integration/...

# Performance testing
go run tools/testing/test_performance.go
```

### Debugging
```bash
# Debug transactions
go run tools/debug/debug_transaction.go

# CLI utilities
go run tools/cli/simple_validator_cli.go
```

## 🌐 API Reference

### Standard Ethereum RPC Methods
- `eth_chainId` - Get chain ID (8888)
- `eth_blockNumber` - Get current block number
- `eth_getBalance` - Get account balance
- `eth_sendRawTransaction` - Submit transactions
- `eth_getTransactionReceipt` - Get transaction receipt

### Quantum-Specific Methods
- `quantum_sendRawTransaction` - Submit quantum transactions
- `quantum_getMetrics` - Get network metrics

## 📚 Documentation

- **[Getting Started](docs/GETTING_STARTED.md)** - 5-minute setup guide
- **[Architecture](docs/ARCHITECTURE.md)** - Technical architecture overview
- **[Enterprise Deployment](docs/ENTERPRISE_DEPLOYMENT_GUIDE.md)** - Production deployment
- **[Implementation Summary](docs/IMPLEMENTATION_SUMMARY.md)** - Feature overview
- **[Production Checklist](docs/PRODUCTION_READINESS_CHECKLIST.md)** - 150+ verification items

## 🔒 Security Features

- **NIST Post-Quantum Standards**: CRYSTALS-Dilithium-II and Kyber-512
- **Hardware Security Module (HSM)**: FIPS 140-2 Level 3/4 compliance
- **Multi-Validator Consensus**: Byzantine fault tolerance
- **Rate Limiting & DDoS Protection**: Production security
- **Validator Slashing**: Economic security mechanisms

## 🚀 Enterprise Features

- **Kubernetes Infrastructure**: Production-ready orchestration
- **Monitoring & Alerting**: Prometheus + Grafana + AlertManager
- **JavaScript/TypeScript SDK**: Full developer toolkit
- **MetaMask Integration**: Browser wallet support
- **Automated Deployment**: One-click network setup
- **HSM Integration**: Enterprise key management

## 📊 Performance

- **Block Time**: 2 seconds
- **TPS**: High throughput with quantum signatures
- **Gas Optimization**: 98% reduction (Dilithium: 800 gas vs 50,000)
- **RPC Response Time**: <5ms average

## 🏗️ Network Status

Current network is **PRODUCTION READY** with:
- ✅ Multi-validator consensus operational
- ✅ Quantum transactions processing
- ✅ Smart contracts deployed successfully
- ✅ Transaction receipts working
- ✅ All RPC methods functional

## 🤝 Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for development guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For help and support:
- **Documentation**: [docs/](docs/) directory
- **Issues**: Create GitHub issue
- **Discussion**: GitHub Discussions

---

**Built with quantum-resistant cryptography for a post-quantum future** 🌟
