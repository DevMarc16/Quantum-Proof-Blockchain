# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a production-ready quantum-resistant blockchain implementation with full EVM compatibility. The project implements NIST-standardized post-quantum cryptographic algorithms (Dilithium, Kyber) using the Cloudflare CIRCL library for real cryptographic operations.

## Development Commands

### Building and Running
```bash
# Build the quantum node
go build -o build/quantum-node ./cmd/quantum-node

# Run the node (basic)
./build/quantum-node --data-dir ./data

# Run with mining enabled
./build/quantum-node --data-dir ./data --mining --validator

# Run tests
go test ./tests/unit/...
go test ./tests/integration/...
```

### Testing Transaction Functionality
```bash
# Contract deployment tests (with blockchain running)
go run tests/manual/deploy_quantum_token.go
go run tests/manual/test_contract_deployment.go

# Performance tests
go run tests/performance/test_fast_performance.go
go run tests/performance/test_live_blockchain.go

# Basic transaction tests
go run tests/manual/test_transaction.go
go run tests/manual/test_rpc_submit.go
go run tests/manual/test_query_tx.go
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

**Node Architecture** (`chain/node/`):
- `node.go`: Core blockchain node with fast consensus, mining, and QTM token management
- `rpc.go`: Complete JSON-RPC API with all major network methods (eth_call, eth_getCode, etc.)
- `txpool.go`: High-performance transaction pool (5000 tx capacity)
- `blockchain.go`: Full EVM-compatible blockchain with persistent contract storage

**Fast Consensus & Token System**:
- `chain/consensus/fast_consensus.go`: Flare-like consensus with 2-second blocks
- `chain/types/token.go`: Native QTM token with 1B supply and rewards
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

**✅ Production Ready Features**:
- Real NIST quantum cryptography (CRYSTALS-Dilithium-II, CRYSTALS-Kyber-512) 
- Complete RPC API suite with all major network methods
- Full contract storage and EVM execution support
- 98% gas cost optimization (Dilithium: 50,000 → 800 gas)
- 2-second fast block production with Flare-like consensus
- Persistent LevelDB storage for state, blocks, and contracts
- Rate limiting and production security features
- Native QTM token with 1 billion supply and block rewards
- Genesis configuration with pre-funded accounts
- Complete transaction pool with signature validation

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

## Important Notes

- **Real Cryptography**: This implementation uses authentic NIST post-quantum algorithms
- **EVM Compatibility**: Transaction structure maintains Ethereum compatibility
- **Chain ID**: Always use 8888 for this quantum blockchain
- **Signature Verification**: Critical to use `SigningHash()` not `Hash()` for verification
- **Binary Data**: JSON marshaling requires hex encoding for signatures and public keys