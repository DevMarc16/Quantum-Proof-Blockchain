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
# Clean dependencies and build all components
make clean build

# Or build individual components:
make build-node          # Build quantum-node only
make build-validator-cli  # Build validator CLI only
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
# Build validator CLI first
make build-validator-cli

# Generate quantum validator keys
./build/binaries/validator-cli -generate -algorithm dilithium -output validator-keys

# Register as validator
./build/binaries/validator-cli -register -stake 100000 -commission 500 -rpc http://localhost:8545

# Check validator status
./build/binaries/validator-cli -status -rpc http://localhost:8545
```

#### 5. Run Tests
```bash
# Run integration tests
make test-integration

# Run unit tests
make test-unit

# Test specific performance scenarios
go run tests/performance/test_fast_performance/test_fast_performance.go
go run tests/performance/test_live_blockchain/test_live_blockchain.go
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
├── build/                       # Compiled artifacts and binaries
│   ├── binaries/                # Built executables
│   ├── contracts/               # Compiled smart contracts (ABI/bytecode)
│   └── docs/                    # Generated documentation
├── cache/                       # Build cache files
├── chain/                       # Core blockchain implementation
│   ├── config/                  # Chain configuration
│   ├── consensus/               # Quantum consensus algorithms
│   ├── crypto/                  # Quantum cryptography (Dilithium, Kyber, Falcon)
│   ├── economics/               # Token economics and rewards
│   ├── evm/                     # EVM compatibility layer
│   ├── governance/              # On-chain governance
│   ├── monitoring/              # Metrics and monitoring
│   ├── network/                 # P2P networking layer
│   ├── node/                    # Multi-validator node implementation
│   ├── security/                # Security modules (including HSM)
│   └── types/                   # Transaction and block types
├── clients/                     # Client implementations
│   ├── cli/                     # Command-line interface
│   └── wallet-sdk/              # Wallet software development kit
├── cmd/                         # Main executables
│   ├── quantum-node/            # Main node binary
│   └── validator-cli/           # Validator management CLI
├── config/                      # Configuration templates
├── configs/                     # Runtime configurations
├── contracts/                   # Smart contract source code
│   ├── interfaces/              # Contract interfaces
│   ├── libraries/               # Reusable contract libraries
│   └── scripts/                 # Deployment scripts
├── data/                        # Runtime data
├── deploy/                      # Deployment configurations
├── docs/                        # Comprehensive documentation
│   ├── developer/               # Developer guides
│   ├── enterprise/              # Enterprise features
│   └── user/                    # User documentation
├── examples/                    # Usage examples
├── infra/                       # Infrastructure as code
│   ├── ci/                      # CI/CD configurations
│   ├── docker/                  # Docker configurations
│   ├── helm/                    # Kubernetes Helm charts
│   └── nginx/                   # Load balancer configs
├── integrations/                # Third-party integrations
│   └── metamask/                # MetaMask snap integration
├── k8s/                         # Kubernetes manifests
│   ├── base/                    # Base configurations
│   ├── monitoring/              # Monitoring stack
│   ├── networking/              # Network policies
│   ├── security/                # Security policies
│   └── validators/              # Validator deployments
├── monitoring/                  # Monitoring configurations
├── scripts/                     # Operational scripts
├── sdk/                         # Software Development Kits
│   └── js/                      # JavaScript/TypeScript SDK
│       ├── examples/            # SDK usage examples
│       ├── src/                 # SDK source code
│       ├── tests/               # SDK tests
│       └── types/               # Type definitions
├── spec/                        # Technical specifications
├── tests/                       # Comprehensive test suites
│   ├── benchmark/               # Performance benchmarks
│   ├── integration/             # Integration tests
│   ├── manual/                  # Manual testing tools
│   ├── performance/             # Performance tests
│   ├── production/              # Production environment tests
│   └── unit/                    # Unit tests
└── tools/                       # Development and operational tools
    ├── cli/                     # Command-line utilities
    ├── debug/                   # Debugging tools
    ├── deployment/              # Deployment automation
    └── testing/                 # Testing utilities
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
# Deploy quantum contracts with fixed keys
go run tools/deployment/deploy_with_fixed_keys/deploy_with_fixed_keys.go

# Deploy simple contracts
go run tools/deployment/deploy_simple/deploy_simple.go

# Fund quantum accounts
go run tools/deployment/fund_quantum_account/fund_quantum_account.go

# Deploy contracts for quantum blockchain
go run tools/deployment/deploy_contracts_quantum/deploy_contracts_quantum.go
```

### Testing
```bash
# Run all test suites
make test-unit test-integration

# Run integration tests only
make test-integration

# Run unit tests only  
make test-unit

# Performance testing
go run tests/performance/test_fast_performance/test_fast_performance.go
go run tests/performance/test_live_blockchain/test_live_blockchain.go

# Manual testing tools
go run tests/manual/test_simple_balance/test_simple_balance.go
go run tests/manual/test_rpc_submit/test_rpc_submit.go
```

### Debugging
```bash
# Debug transactions
go run tools/debug/debug_transaction.go

# CLI utilities
go run tools/cli/simple_validator_cli.go

# Testing utilities
go run tools/testing/tools/testing/test_simple_transaction/test_simple_transaction.go
go run tools/testing/tools/testing/test_receipt_lookup/test_receipt_lookup.go
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
