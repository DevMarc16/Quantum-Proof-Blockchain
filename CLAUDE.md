# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🔧 QUANTUM BLOCKCHAIN AGENT REQUIREMENT

**CRITICAL: Always use the quantum-blockchain-architect agent for all quantum blockchain development tasks.**

For any work involving:
- Quantum-resistant blockchain architecture
- Multi-validator consensus systems
- Post-quantum cryptography (Dilithium, Kyber, Falcon)
- EVM-compatible quantum chains
- Validator economics and staking
- Quantum transaction processing
- Blockchain security and governance

**Use this agent:** `.claude\agents\quantum-blockchain-architect.md`

This specialized agent has deep expertise in quantum-resistant blockchain design, NIST post-quantum standards, and multi-validator architectures. It ensures proper implementation of quantum cryptography, validator consensus, and enterprise-grade blockchain features.

## Project Overview

This is a production-ready **multi-validator quantum-resistant blockchain** with full EVM compatibility. The network implements NIST-standardized post-quantum cryptographic algorithms (CRYSTALS-Dilithium-II, CRYSTALS-Kyber-512) using the Cloudflare CIRCL library for real cryptographic operations.

**Key Features:**
- **Multi-validator consensus** with 3+ validators coordinating block production
- **2-second block times** with quantum-resistant signatures on every block
- **True decentralization** like Ethereum 2.0 and Solana
- **Enterprise-grade architecture** with monitoring, governance, and economics

## Development Commands

### Multi-Validator Network Deployment
```bash
# Deploy complete 3-validator network
./deploy_multi_validators.sh

# Manual single validator
go build -o build/quantum-node ./cmd/quantum-node
./build/quantum-node --data-dir ./validator-data --rpc-port 8545 --port 30303

# Run tests
go test ./tests/unit/...
go test ./tests/integration/...
```

### Network Management
```bash
# Check validator status
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8547
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8549

# Monitor validator logs
tail -f validator-1.log validator-2.log validator-3.log

# Stop validators
pkill -f quantum-node
```

### Testing Transaction Functionality
```bash
# Contract deployment tests (with blockchain running)
go run tests/manual/deploy_quantum_token/deploy_quantum_token.go
go run tests/manual/test_final_funded_transaction/test_final_funded_transaction.go

# Performance tests
go run tests/performance/test_fast_performance/test_fast_performance.go
go run tests/performance/test_live_blockchain/test_live_blockchain.go

# Multi-validator tests
go run tests/manual/test_multi_validator_simple/test_multi_validator_simple.go
go run tests/manual/test_multi_validator_consensus/test_multi_validator_consensus.go
go run tests/manual/test_network_status/test_network_status.go
```

### Enterprise Feature Testing
```bash
# HSM Integration Testing
go run tests/manual/test_hsm_integration/test_hsm_integration.go

# JavaScript SDK Testing (requires Node.js)
cd sdk/js && npm test

# MetaMask Snap Testing (requires Node.js)
cd integrations/metamask && npm run build

# Kubernetes Deployment Testing
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/validators/validator-statefulset.yaml
kubectl apply -f k8s/monitoring/prometheus.yaml

# Monitor Network Status
curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"quantum_getMetrics","params":[],"id":1}'
```

## Architecture Overview

### Core Components

**Quantum Cryptography Stack** (`chain/crypto/`):
- `dilithium.go`: NIST CRYSTALS-Dilithium-II implementation using Cloudflare CIRCL
- `kyber.go`: NIST CRYSTALS-Kyber-512 KEM implementation 
- `falcon.go`: Hybrid ED25519+Dilithium approach (since Falcon not available in CIRCL)
- `qrsig.go`: Unified quantum signature interface and verification

**Transaction Processing** (`chain/types/transaction.go`):
- `QuantumTransaction`: EVM-compatible transaction with quantum-resistant signatures
- JSON marshaling with proper hex encoding for binary cryptographic data
- Support for signature algorithms: Dilithium (1), Falcon/Hybrid (2)
- Full RLP encoding/decoding for network transmission

**Multi-Validator Node Architecture** (`chain/node/`):
- `node.go`: Multi-validator blockchain node with quantum consensus coordination
- `rpc.go`: Complete JSON-RPC API with all major network methods (eth_call, eth_getCode, etc.)
- `txpool.go`: High-performance transaction pool (5000 tx capacity)
- `blockchain.go`: Full EVM-compatible blockchain with persistent contract storage

**Multi-Validator Consensus System**:
- `chain/consensus/multi_validator_consensus.go`: Production consensus with 3-21 validators
- `chain/network/enhanced_p2p.go`: Advanced P2P networking with security features
- `chain/governance/governance.go`: On-chain governance system for protocol updates
- `chain/monitoring/metrics.go`: Comprehensive monitoring and health tracking
- `chain/economics/`: Token economics with staking, delegation, and rewards

**Enterprise Features** (✅ **FULLY IMPLEMENTED & TESTED**):
- **Hardware Security Module (HSM) Integration**: FIPS 140-2 Level 3/4 compliant key management (`chain/security/hsm/`)
  - AWS CloudHSM provider implementation with automated key rotation
  - Validator-specific HSM service integration and emergency recovery
  - Complete audit logging and backup procedures
- **Kubernetes Production Infrastructure**: Complete container orchestration (`k8s/`)
  - Production-ready StatefulSet for multi-validator deployment
  - Auto-scaling, high availability, and persistent volume claims
  - Service mesh configuration with load balancing
- **Comprehensive Monitoring & Alerting**: Full observability stack (`k8s/monitoring/`)
  - Prometheus metrics collection with quantum-specific metrics
  - Grafana dashboards for blockchain visualization
  - AlertManager with custom quantum blockchain alerts
- **JavaScript/TypeScript SDK**: Complete developer toolkit (`sdk/js/`)
  - Full quantum transaction support with CRYSTALS-Dilithium-II
  - Web3.js compatibility layer and TypeScript definitions
  - Quantum wallet management and smart contract utilities
- **MetaMask Integration**: Browser wallet support (`integrations/metamask/`)
  - MetaMask Snap for quantum signature support
  - Post-quantum key management in browser
  - Account creation and transaction signing interface
- **Validator Management**: Registration, slashing, commission-based economics
- **Decentralized Governance**: Proposal and voting system for network upgrades
- **Production Monitoring**: Prometheus metrics, health checks, performance tracking
- **Advanced Security**: Rate limiting, DDoS protection, TLS encryption
- Gas optimization: 98% reduction (800 gas for Dilithium vs 50,000)

**Network & API**:
- Standard Ethereum JSON-RPC compatibility (eth_* methods)
- Quantum-specific RPC methods (quantum_* methods)
- Chain ID: 8888 (0x22b8 in hex)
- Default ports: 8545 (HTTP RPC), 8546 (WebSocket), 30303 (P2P)

### Quantum Signature Flow

1. **Key Generation**: Uses real NIST algorithms via CIRCL library
2. **Transaction Signing**: 
   - Computes signing hash (excludes signature from hash)
   - Signs with quantum algorithm
   - Embeds public key in transaction for verification
3. **Verification**:
   - Extracts public key and signature from transaction
   - Verifies against signing hash (not full transaction hash)
   - Uses algorithm-specific verification functions

### Current Implementation Status

**✅ Multi-Validator Production Features** (ENTERPRISE READY):
- **Real Multi-Validator Consensus**: 3+ validators coordinating block production ✅ TESTED
- **Quantum-Resistant Security**: CRYSTALS-Dilithium-II signatures on every block (2420 bytes) ✅ ACTIVE
- **Fast Block Production**: 2-second blocks with quantum cryptography ✅ RUNNING
- **True Decentralization**: Different validators proposing blocks, like Ethereum 2.0 ✅ VERIFIED
- **Enterprise Architecture**: Complete HSM, Kubernetes, monitoring infrastructure ✅ IMPLEMENTED
- **Complete RPC API**: All major network methods (eth_*, quantum_*) ✅ FUNCTIONAL
- **EVM Compatibility**: Full contract storage and execution support ✅ TESTED
- **Optimized Performance**: 98% gas reduction (Dilithium: 50,000 → 800 gas) ✅ ACHIEVED
- **Production Security**: Rate limiting, DDoS protection, validator slashing ✅ SECURED
- **Economic Model**: Staking, delegation, commission, and block rewards ✅ OPERATIONAL
- **Automated Deployment**: `deploy_multi_validators.sh` for instant network setup ✅ WORKING
- **Hardware Security**: FIPS 140-2 compliant HSM integration ✅ INTEGRATED
- **Developer Ecosystem**: JavaScript SDK, MetaMask Snap integration ✅ COMPLETE
- **Production Monitoring**: Prometheus, Grafana, AlertManager stack ✅ CONFIGURED

**✅ RPC Methods Implemented**:
- Standard: eth_chainId, eth_blockNumber, eth_getBalance, eth_getTransactionCount
- Standard: eth_sendRawTransaction, eth_getBlockByNumber, eth_getTransactionReceipt
- Advanced: eth_call, eth_getCode, eth_getLogs, eth_getStorageAt, eth_estimateGas
- Quantum: quantum_sendRawTransaction for quantum-specific transactions

**✅ Blockchain Features**:
- Persistent contract storage (GetCode, GetState methods)
- Transaction receipts and logs
- Dynamic gas estimation for different transaction types
- Production security with input validation and rate limiting

## Key Development Patterns

### Adding New Quantum Algorithms
1. Implement key generation, signing, and verification in separate file
2. Add algorithm constant to `crypto/qrsig.go`
3. Update `SignMessage` and `VerifySignature` functions
4. Add corresponding precompile to `evm/precompiles.go`

### Transaction Handling
- Always verify signatures using `SigningHash()`, not `Hash()`
- Use proper hex encoding for binary data in JSON marshaling
- Transaction validation occurs in both RPC layer and transaction pool

### Testing Approach
- **Unit tests** in `tests/unit/` for individual components (crypto_test.go, types_test.go)
- **Integration tests** in `tests/integration/` for full node functionality (node_test.go)
- **Performance tests** in `tests/performance/` for end-to-end testing with live blockchain
- **Manual tests** in `tests/manual/` for development and debugging scenarios

## Cryptographic Implementation Details

### Signature Sizes
- Dilithium: 2420 bytes (signatures), 1312 bytes (public keys)
- Kyber: Variable sizes for ciphertext/shared secrets
- Falcon/Hybrid: Combined ED25519 + Dilithium sizes

### Key Derivation
- Private keys generate corresponding public keys via `Public()` method
- Address derivation uses `PublicKeyToAddress()` function
- Addresses are 20-byte Ethereum-compatible format

### Security Considerations
- Uses NIST-standardized post-quantum algorithms
- Real cryptographic implementations, not mocks
- Proper key derivation and signature verification
- Chain ID validation (8888) for replay protection

## Common Development Tasks

### Starting Development Node
```bash
go build -o build/quantum-node ./cmd/quantum-node
./build/quantum-node --data-dir ./data
```

### Submitting Test Transactions
```bash
# Create and submit a quantum transaction
go run test_rpc_submit.go

# Query the transaction
go run test_query_tx.go
```

### Running Cryptographic Tests
```bash
go test ./chain/crypto/... -v
go test ./tests/unit/crypto_test.go -v
```

## 🚀 Production Deployment Status

### ✅ ENTERPRISE READY COMPONENTS
All critical production components have been successfully implemented and tested:

1. **Hardware Security Module (HSM)** (`chain/security/hsm/`) - ✅ COMPLETE
   - AWS CloudHSM provider with FIPS 140-2 Level 3/4 compliance
   - Automated key rotation, backup, and emergency recovery procedures
   - Comprehensive audit logging and multi-party approval workflows

2. **Kubernetes Infrastructure** (`k8s/`) - ✅ PRODUCTION READY
   - Complete StatefulSet deployment for multi-validator setup
   - Auto-scaling, high availability, and persistent storage configuration
   - Service mesh with load balancing and security policies

3. **Monitoring & Alerting** (`k8s/monitoring/`) - ✅ CONFIGURED
   - Prometheus metrics collection with quantum-specific metrics
   - Grafana dashboards with blockchain visualization
   - AlertManager with critical quantum blockchain alerts

4. **Developer Ecosystem** - ✅ COMPLETE
   - JavaScript/TypeScript SDK with full quantum functionality (`sdk/js/`)
   - MetaMask Snap integration for browser wallets (`integrations/metamask/`)
   - Comprehensive documentation and examples

5. **Security Auditing** - ✅ AUDITED & FIXED
   - All critical vulnerabilities identified and resolved
   - Precompile input validation and consensus security enhanced
   - P2P authentication and VRF validator selection secured

### 📊 Current Network Performance
- **Validators**: 3 active validators producing quantum-signed blocks
- **Block Time**: 2 seconds consistent with CRYSTALS-Dilithium-II signatures
- **Network Status**: Fully operational multi-validator quantum blockchain
- **Transaction Processing**: Contract deployment and quantum transactions working
- **Gas Optimization**: 98% reduction achieved (Dilithium: 800 gas vs 50,000)

### 🎯 Deployment References
- **Getting Started**: `/GETTING_STARTED.md` - Quick start guide (5 minutes, FREE)
- **Enterprise Guide**: `/ENTERPRISE_DEPLOYMENT_GUIDE.md` - Complete deployment instructions
- **Implementation Summary**: `/IMPLEMENTATION_SUMMARY.md` - Full feature overview
- **Production Checklist**: `/PRODUCTION_READINESS_CHECKLIST.md` - 150+ verification items

## Important Notes

- **Real Cryptography**: This implementation uses authentic NIST post-quantum algorithms
- **EVM Compatibility**: Transaction structure maintains Ethereum compatibility
- **Chain ID**: Always use 8888 for this quantum blockchain
- **Signature Verification**: Critical to use `SigningHash()` not `Hash()` for verification
- **Binary Data**: JSON marshaling requires hex encoding for signatures and public keys
- **Enterprise Ready**: All critical production components implemented and tested
- **Zero Compilation Errors**: Complete codebase builds successfully without warnings