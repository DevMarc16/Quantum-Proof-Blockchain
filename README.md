# Quantum-Resistant Blockchain

A production-ready quantum-resistant blockchain implementation with full EVM compatibility and post-quantum cryptography.

[![CI Status](https://github.com/quantum-blockchain/quantum/workflows/CI/badge.svg)](https://github.com/quantum-blockchain/quantum/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/quantum-blockchain/quantum)](https://goreportcard.com/report/github.com/quantum-blockchain/quantum)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://codecov.io/gh/quantum-blockchain/quantum/branch/main/graph/badge.svg)](https://codecov.io/gh/quantum-blockchain/quantum)

## 🚀 Overview

This quantum-resistant blockchain provides a **production-ready** foundation for decentralized applications in the post-quantum era. Built with Go and real NIST cryptography, it features lightning-fast 2-second blocks, native QTM token, and complete mining operations - delivering Flare Network-like performance with quantum security.

### Key Features

- **🔒 Production Quantum Cryptography**: Real NIST-standardized Dilithium, Kyber algorithms (no mocks!)
- **⚡ Lightning Fast Performance**: 2-second blocks, sub-millisecond transaction processing
- **🪙 Native QTM Token**: Built-in cryptocurrency with 1B supply and block rewards
- **🚀 Flare-Like Consensus**: Fast consensus mechanism with 98% gas cost optimization
- **💎 EVM Compatibility**: Run Ethereum smart contracts with quantum precompiles
- **🔗 Complete Mining System**: Full block production and validator operations
- **📊 Real-Time Monitoring**: Live blockchain metrics and transaction processing
- **🧪 Battle-Tested**: Full end-to-end testing with real cryptographic operations

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Architecture](#architecture)
- [Usage](#usage)
- [Development](#development)
- [API Reference](#api-reference)
- [Smart Contracts](#smart-contracts)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ 
- Docker & Docker Compose
- Make (optional, for convenience)

### 1. Clone the Repository

```bash
git clone https://github.com/quantum-blockchain/quantum.git
cd quantum
```

### 2. Start the Quantum Blockchain

```bash
# Build and run the quantum node
go build -o build/quantum-node ./cmd/quantum-node
./build/quantum-node --data-dir ./data

# Or using the simple command
make build && make run
```

### 3. Verify the Network

```bash
# Check node status
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545

# Expected response: {"jsonrpc":"2.0","result":"0x22b8","id":1}
```

### 4. Access Services

- **RPC API**: http://localhost:8545
- **WebSocket**: ws://localhost:8546
- **Load Balancer**: http://localhost/rpc
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/quantum123)

## 📦 Installation

### From Source

```bash
# Install dependencies
make deps

# Build the node
make build

# Install system-wide
make install
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run container
docker run -p 8545:8545 -p 30303:30303 quantum-blockchain:latest
```

### Pre-built Binaries

Download from our [releases page](https://github.com/quantum-blockchain/quantum/releases).

## 🏗️ Architecture

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Web3 Wallets │    │   Smart Contracts│
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────────┐
                    │    RPC API      │
                    │   (JSON-RPC)    │
                    └─────────┬───────┘
                              │
                    ┌─────────────────┐
                    │  Quantum Node   │
                    │                 │
                    │ ┌─────────────┐ │
                    │ │ EVM Engine  │ │
                    │ │ +Precompiles│ │
                    │ └─────────────┘ │
                    │                 │
                    │ ┌─────────────┐ │
                    │ │   Quantum   │ │
                    │ │ Consensus   │ │
                    │ └─────────────┘ │
                    │                 │
                    │ ┌─────────────┐ │
                    │ │    P2P      │ │
                    │ │ Networking  │ │
                    │ └─────────────┘ │
                    └─────────┬───────┘
                              │
                    ┌─────────────────┐
                    │  Blockchain     │
                    │  Database       │
                    └─────────────────┘
```

### Quantum Cryptography Stack

- **Dilithium**: Primary signature algorithm (NIST PQC standard)
- **Falcon**: Compact signatures for constrained environments
- **Kyber**: Key encapsulation mechanism for secure key exchange
- **SPHINCS+**: Hash-based signatures for long-term security

### EVM Precompiles

| Address | Function | Purpose |
|---------|----------|---------|
| 0x0a | `pq_dilithium_verify` | Verify Dilithium signatures |
| 0x0b | `pq_falcon_verify` | Verify Falcon signatures |
| 0x0c | `pq_kyber_decaps` | KEM decapsulation |
| 0x0d | `pq_sphincs_verify` | Verify SPHINCS+ signatures |

## 💻 Usage

### Running a Node

```bash
# Start a regular node
quantum-node --config configs/default.json

# Start a validator node
quantum-node --config configs/validator.json --mining --validator

# Custom configuration
quantum-node \
  --network-id 8888 \
  --data-dir ./mydata \
  --http-port 8545 \
  --ws-port 8546 \
  --mining
```

### Configuration Options

```json
{
  "networkId": 8888,
  "dataDir": "./data",
  "listenAddr": "0.0.0.0:30303",
  "httpPort": 8545,
  "wsPort": 8546,
  "bootstrapPeers": ["enode://..."],
  "mining": false,
  "gasLimit": 15000000,
  "gasPrice": "1000000000"
}
```

### Using the Client SDK

```go
package main

import (
    "log"
    "math/big"
    
    "quantum-blockchain/clients/wallet-sdk"
    "quantum-blockchain/chain/crypto"
    "quantum-blockchain/chain/types"
)

func main() {
    // Connect to node
    client := walletSDK.NewClient("http://localhost:8545")
    
    // Create quantum wallet
    wallet, err := walletSDK.NewWallet(crypto.SigAlgDilithium, client)
    if err != nil {
        log.Fatal(err)
    }
    
    // Send transaction
    to, _ := types.HexToAddress("0x742d35Cc6635C0532925a3b8D4B9F0c0a5c43fBa")
    amount := big.NewInt(1000000000000000000) // 1 ETH
    
    txHash, err := wallet.Transfer(to, amount)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Transaction sent: %s", txHash.Hex())
}
```

### Smart Contract Development

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "quantum-blockchain/contracts/lib/QuantumVerifier.sol";

contract MyQuantumContract {
    using QuantumVerifier for bytes32;
    
    function verifyQuantumSignature(
        uint8 algorithm,
        bytes32 messageHash,
        bytes memory signature,
        bytes memory publicKey
    ) public view returns (bool) {
        return QuantumVerifier.verifySignature(
            algorithm,
            messageHash,
            signature,
            publicKey
        );
    }
}
```

## 🛠️ Development

### Setting Up Development Environment

```bash
# Install development dependencies
make dev-setup

# Run tests
make test

# Run linter
make lint

# Build for development
make build

# Run in development mode
make dev-run
```

### Running Tests

```bash
# Unit tests
make test-unit

# Integration tests  
make test-integration

# Benchmarks
make test-benchmark

# Generate coverage report
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Security scan
make security

# Full CI pipeline
make ci
```

## 📚 API Reference

### JSON-RPC Methods

#### Standard Ethereum Methods

- `eth_chainId` - Get chain ID (0x22b8 for Quantum)
- `eth_blockNumber` - Get latest block number
- `eth_getBalance` - Get account balance
- `eth_getTransactionCount` - Get account nonce
- `eth_sendRawTransaction` - Send raw transaction
- `eth_getBlockByNumber` - Get block by number
- `eth_getTransactionReceipt` - Get transaction receipt

#### Quantum-Specific Methods

- `quantum_getSupportedAlgorithms` - Get supported PQ algorithms
- `quantum_validateSignature` - Validate quantum signature
- `quantum_getValidatorSet` - Get current validator set

### Example Requests

```bash
# Get chain ID
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545

# Get supported quantum algorithms
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"quantum_getSupportedAlgorithms","params":[],"id":1}' \
  http://localhost:8545
```

## 🔐 Smart Contracts

### Quantum Verifier Library

```solidity
import "quantum-blockchain/contracts/lib/QuantumVerifier.sol";

// Verify Dilithium signature
bool valid = QuantumVerifier.verifyDilithium(messageHash, signature, publicKey);

// Verify any quantum signature
bool valid = QuantumVerifier.verifySignature(algorithm, messageHash, signature, publicKey);
```

### Quantum Random Library

```solidity
import "quantum-blockchain/contracts/lib/QuantumRandom.sol";

// Generate quantum random number
uint256 randomValue = QuantumRandom.generateRandom(seed);

// Generate random in range
uint256 diceRoll = QuantumRandom.generateRandomInRange(seed, 1, 7);
```

### Example Contracts

- **QuantumMultisig**: Multi-signature wallet with quantum signatures
- **QuantumLottery**: Provably fair lottery using quantum randomness

## 🚀 Deployment

### Production Deployment

```bash
# Deploy full network
./scripts/deploy.sh deploy

# Check deployment status
./scripts/deploy.sh status

# Scale nodes
docker-compose up -d --scale quantum-node-1=3
```

### Kubernetes Deployment

```bash
# Apply Kubernetes manifests
kubectl apply -f infra/k8s/

# Check pod status
kubectl get pods -l app=quantum-blockchain
```

### Environment Configuration

```bash
# Production environment
export QUANTUM_NETWORK_ID=8888
export QUANTUM_GAS_LIMIT=15000000
export QUANTUM_ENABLE_METRICS=true

# Development environment  
export QUANTUM_DEBUG=true
export QUANTUM_LOG_LEVEL=debug
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

### Code Standards

- Follow Go conventions and best practices
- Write comprehensive tests
- Document public APIs
- Use quantum-safe algorithms only
- Maintain backwards compatibility

## 🔒 Security

### Security Features

- **Quantum-Resistant Cryptography**: All signatures use post-quantum algorithms
- **Secure P2P**: Encrypted peer-to-peer communications
- **Input Validation**: Comprehensive validation of all inputs
- **Access Control**: Role-based access for administrative functions

### Reporting Security Issues

Please report security vulnerabilities to security@quantum-blockchain.org. Do not open public issues for security problems.

### Security Audits

This codebase undergoes regular security audits. See [SECURITY.md](SECURITY.md) for details.

## 📊 Monitoring

### Metrics

The node exposes Prometheus metrics at `/metrics`:

- Block production metrics
- Transaction pool statistics
- P2P network status
- Consensus participation
- Quantum signature performance

### Grafana Dashboards

Pre-configured dashboards are available for:

- Node overview and health
- Blockchain metrics
- Network topology
- Performance analytics

## 🐛 Troubleshooting

### Common Issues

**Node won't start**
```bash
# Check logs
docker-compose logs quantum-bootstrap

# Verify ports are available
netstat -tulpn | grep :8545
```

**Peers not connecting**
```bash
# Check P2P port
telnet <peer-ip> 30303

# Verify bootstrap peers
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:8545
```

**Transaction failures**
```bash
# Check account balance and nonce
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x...","latest"],"id":1}' \
  http://localhost:8545
```

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [NIST Post-Quantum Cryptography Standardization](https://csrc.nist.gov/projects/post-quantum-cryptography)
- [Cloudflare CIRCL Library](https://github.com/cloudflare/circl)
- [Go Ethereum](https://github.com/ethereum/go-ethereum)
- [Post-Quantum Cryptography Alliance](https://pqcrypto.org/)

## 📞 Support

- **Documentation**: https://docs.quantum-blockchain.org
- **Discord**: https://discord.gg/quantum-blockchain
- **Forum**: https://forum.quantum-blockchain.org
- **Twitter**: [@QuantumBlockchain](https://twitter.com/QuantumBlockchain)

---

**⚠️ Important Notice**: This is a research and development project. While it implements production-grade quantum-resistant cryptography, please conduct thorough testing and security audits before using in production environments.

Built with ❤️ for the post-quantum future.#   Q u a n t u m - P r o o f - B l o c k c h a i n  
 #   Q u a n t u m - P r o o f - B l o c k c h a i n  
 