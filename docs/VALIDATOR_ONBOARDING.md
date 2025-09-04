# Quantum Blockchain Validator Onboarding & Token Distribution

## Overview

The Quantum Blockchain implements a comprehensive validator onboarding and token distribution system designed for enterprise-grade security, decentralization, and quantum resistance. This document provides complete guidance for validators, delegators, and token holders.

## Table of Contents

1. [Validator Onboarding](#validator-onboarding)
2. [Token Distribution](#token-distribution)
3. [Smart Contracts](#smart-contracts)
4. [CLI Tools](#cli-tools)
5. [Security Best Practices](#security-best-practices)

---

## Validator Onboarding

### Requirements

- **Minimum Stake**: 100,000 QTM
- **Maximum Stake**: 10,000,000 QTM
- **Quantum Keys**: CRYSTALS-Dilithium-II or Falcon-512
- **Hardware**: 16GB RAM, 4 CPU cores, 500GB SSD
- **Network**: 100 Mbps dedicated bandwidth
- **Uptime**: 95% minimum (penalties for downtime)

### Step-by-Step Onboarding Process

#### 1. Generate Quantum-Resistant Keys

```bash
# Generate Dilithium keys (recommended)
./validator-cli -generate -algorithm dilithium -output ./my-validator-keys

# Generate Falcon/Hybrid keys (alternative)
./validator-cli -generate -algorithm falcon -output ./my-validator-keys

# With mnemonic for recovery
./validator-cli -generate -algorithm dilithium -mnemonic -password "secure-password"
```

Output:
- `dilithium.key` / `falcon.key` - Private key (encrypted)
- `validator-profile.json` - Validator configuration
- `mnemonic.txt` - Recovery phrase (if requested)

#### 2. Fund Your Validator Address

After key generation, fund your validator address with at least 100,000 QTM:

```bash
# Check your validator address
cat ./my-validator-keys/validator-profile.json | grep address

# For testnet, use the faucet
curl -X POST https://faucet.quantum-blockchain.io/request \
  -H "Content-Type: application/json" \
  -d '{"address": "YOUR_VALIDATOR_ADDRESS", "type": "validator"}'
```

#### 3. Register as Validator

```bash
# Register with 100K QTM stake and 5% commission
./validator-cli -register \
  -output ./my-validator-keys \
  -stake 100000 \
  -commission 500 \
  -metadata "ipfs://QmValidator..." \
  -rpc http://localhost:8545
```

Registration parameters:
- **stake**: Initial stake amount (min 100K QTM)
- **commission**: Commission rate in basis points (500 = 5%)
- **metadata**: IPFS hash or URL with validator details

#### 4. Start Your Validator Node

```bash
# Build the node
go build -o quantum-node ./cmd/quantum-node

# Start with validator keys
./quantum-node \
  --validator \
  --key-file ./my-validator-keys/dilithium.key \
  --data-dir ./validator-data \
  --rpc-port 8545 \
  --p2p-port 30303 \
  --metrics
```

#### 5. Monitor Performance

```bash
# Check validator status
./validator-cli -status -output ./my-validator-keys

# View on-chain metrics
curl http://localhost:8545/metrics | grep validator

# Check block production
tail -f validator.log | grep "block_produced"
```

### Delegation System

#### For Delegators

Minimum delegation: 100 QTM

```bash
# Delegate to a validator
./validator-cli -delegate \
  -validator 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1 \
  -amount 1000 \
  -rpc http://localhost:8545

# Check delegation status
cast call 0xValidatorRegistry \
  "getDelegation(address,address)" \
  $YOUR_ADDRESS $VALIDATOR_ADDRESS
```

#### For Validators

Validators earn commission on delegated stakes:
- Default commission: 5%
- Maximum commission: 20%
- Can be updated once per epoch (7 days)

### Rewards & Economics

#### Block Rewards
- **Base Reward**: 1 QTM per block (2-second blocks)
- **Annual Decay**: 5% reduction per year
- **Early Adopter Bonus**: 20% extra for first year

#### Staking Rewards
- **Base APY**: 10% for validators
- **Performance Multiplier**: Up to 1.5x for 100% uptime
- **Delegation APY**: 9.5% (after 5% commission)

#### Reward Distribution

```solidity
Total Block Reward = Base Reward + Transaction Fees
├── Validator Commission (5%): To validator
├── Validator Stake Share: Proportional to stake
└── Delegator Rewards (95%): Distributed to delegators
```

### Slashing Conditions

Validators can be slashed for misbehavior:

| Violation | Penalty | Jail Time | Description |
|-----------|---------|-----------|-------------|
| Double Signing | 20% | Permanent | Signing conflicting blocks |
| Downtime | 1% | 1000 blocks | Missing >5% of blocks |
| Invalid Block | 10% | 5000 blocks | Proposing invalid blocks |
| General Misbehavior | 5% | 2500 blocks | Other protocol violations |

#### Unjailing Process

After jail period expires:
```bash
# Unjail validator
cast send 0xValidatorRegistry \
  "unjailValidator()" \
  --from $VALIDATOR_ADDRESS
```

---

## Token Distribution

### Total Supply: 1 Billion QTM

| Category | Allocation | Vesting | Purpose |
|----------|------------|---------|---------|
| Genesis Validators | 15% | None | Initial validator incentives |
| Public Sale | 25% | None | Community distribution |
| Ecosystem Fund | 20% | None | Grants, partnerships, development |
| Team | 15% | 2 years | Core team allocation |
| Advisors | 5% | 2 years | Advisory board |
| Liquidity | 10% | Scheduled | DEX liquidity provision |
| Staking Rewards | 5% | None | Initial staking incentives |
| Community Airdrops | 5% | None | Community building |

### Vesting Schedule

Team and Advisor tokens:
- **TGE Unlock**: 10%
- **Cliff Period**: 6 months
- **Linear Vesting**: 24 months total

```bash
# Check vesting status
cast call 0xTokenDistribution \
  "getVestingSchedule(address)" \
  $BENEFICIARY_ADDRESS

# Claim vested tokens
cast send 0xTokenDistribution \
  "releaseVestedTokens()" \
  --from $BENEFICIARY_ADDRESS
```

### Public Sale Participation

#### Whitelist Registration
```javascript
// Register for whitelist
await tokenDistribution.addToWhitelist([userAddress]);
```

#### Contribution
```bash
# Contribute ETH (min 0.1, max 10 ETH)
cast send 0xTokenDistribution \
  "contributeToPublicSale()" \
  --value 1ether \
  --from $YOUR_ADDRESS
```

### Airdrop Campaigns

#### Claiming Airdrops
```javascript
// Generate Merkle proof off-chain
const proof = getMerkleProof(userAddress, amount, merkleTree);

// Claim airdrop
await tokenDistribution.claimAirdrop(
  campaignId,
  amount,
  proof
);
```

### Testnet Faucet

#### Regular Users
- **Amount**: 100 QTM per day
- **Max Balance**: 1,000 QTM
- **Rate Limit**: 24 hours

```bash
# Request testnet tokens
curl -X POST http://faucet.quantum-blockchain.io/request \
  -H "Content-Type: application/json" \
  -d '{"address": "0x..."}'
```

#### Verified Validators
- **Amount**: 100,000 QTM per day
- **Max Balance**: 200,000 QTM
- **Verification**: Quantum signature required

```javascript
// Request validator tokens
const signature = await signWithQuantumKey(message, privateKey);
await faucet.requestValidatorTokens(
  quantumPublicKey,
  signature,
  nonce
);
```

---

## Smart Contracts

### ValidatorRegistry.sol

Main contract for validator management:

```solidity
interface IValidatorRegistry {
    // Register as validator
    function registerValidator(
        bytes calldata quantumPublicKey,
        uint8 sigAlgorithm,
        uint256 initialStake,
        uint256 commissionRate,
        string calldata metadata
    ) external;
    
    // Delegation functions
    function delegate(address validator, uint256 amount) external;
    function undelegate(address validator, uint256 amount) external;
    function claimRewards() external;
    
    // View functions
    function getValidator(address) external view returns (Validator memory);
    function isValidatorActive(address) external view returns (bool);
}
```

### TokenDistribution.sol

Manages token allocation and distribution:

```solidity
interface ITokenDistribution {
    // Vesting management
    function createVestingSchedule(
        address beneficiary,
        uint256 amount,
        DistributionCategory category,
        bool revocable
    ) external;
    
    // Public sale
    function contributeToPublicSale() external payable;
    
    // Airdrops
    function claimAirdrop(
        uint256 campaignId,
        uint256 amount,
        bytes32[] calldata merkleProof
    ) external;
}
```

### TestnetFaucet.sol

Provides testnet tokens:

```solidity
interface ITestnetFaucet {
    function requestTokens() external;
    function requestValidatorTokens(
        bytes calldata quantumPublicKey,
        bytes calldata signature,
        bytes32 nonce
    ) external;
}
```

### Deployment Addresses (Testnet)

```javascript
const contracts = {
  ValidatorRegistry: "0x1234567890123456789012345678901234567890",
  TokenDistribution: "0x2345678901234567890123456789012345678901",
  TestnetFaucet: "0x3456789012345678901234567890123456789012",
  QTMToken: "0x4567890123456789012345678901234567890123"
};
```

---

## CLI Tools

### Validator CLI

Complete validator management tool:

```bash
# Installation
go install ./cmd/validator-cli

# Key generation
validator-cli -generate -algorithm dilithium

# Registration
validator-cli -register -stake 100000 -commission 500

# Status check
validator-cli -status

# Delegation
validator-cli -delegate -validator 0x... -amount 1000

# Backup/Restore
validator-cli -backup
validator-cli -restore backup-20240101.tar
```

### Contract Interaction

Using Foundry Cast:

```bash
# Check validator info
cast call $VALIDATOR_REGISTRY \
  "getValidator(address)(tuple)" \
  $VALIDATOR_ADDRESS

# Delegate tokens
cast send $VALIDATOR_REGISTRY \
  "delegate(address,uint256)" \
  $VALIDATOR_ADDRESS \
  100000000000000000000 \
  --from $YOUR_ADDRESS

# Claim rewards
cast send $VALIDATOR_REGISTRY \
  "claimRewards()" \
  --from $YOUR_ADDRESS
```

---

## Security Best Practices

### Key Management

1. **Hardware Security Module (HSM)**
   - Store validator keys in HSM for production
   - Use encrypted backups with strong passwords
   - Implement key rotation every 6 months

2. **Multi-Signature Setup**
   ```solidity
   // Use multisig for large stakes
   MultiSigWallet validatorWallet = new MultiSigWallet(
       owners,
       requiredSignatures
   );
   ```

3. **Key Backup Strategy**
   - Keep 3 encrypted backups in separate locations
   - Store mnemonic phrase in bank safety deposit box
   - Test recovery process quarterly

### Operational Security

1. **Node Security**
   - Run validator behind firewall
   - Use VPN for remote access
   - Enable DDoS protection
   - Regular security updates

2. **Monitoring & Alerts**
   ```yaml
   # Prometheus alerts
   - alert: ValidatorOffline
     expr: validator_block_production == 0
     for: 5m
     
   - alert: HighMissedBlocks
     expr: validator_missed_blocks > 50
     for: 10m
   ```

3. **Slashing Prevention**
   - Never run same validator keys on multiple nodes
   - Implement double-sign protection in client
   - Use remote signer with slashing protection

### Smart Contract Security

1. **Audit Recommendations**
   - All contracts audited by 2+ firms
   - Formal verification of critical functions
   - Bug bounty program with Immunefi

2. **Upgrade Patterns**
   ```solidity
   // Use proxy pattern for upgrades
   contract ValidatorRegistryV2 is ValidatorRegistryV1 {
       // New functionality
   }
   ```

3. **Emergency Procedures**
   - Pause mechanism for critical issues
   - Governance timelock for changes
   - Emergency withdrawal for users

### Quantum Security

1. **Algorithm Selection**
   - Primary: CRYSTALS-Dilithium-II (NIST Level 2)
   - Backup: Falcon-512 (smaller signatures)
   - Future: SPHINCS+ for long-term security

2. **Key Sizes & Performance**
   | Algorithm | Public Key | Signature | Verification |
   |-----------|------------|-----------|--------------|
   | Dilithium-II | 1.3 KB | 2.4 KB | 2.5 ms |
   | Falcon-512 | 897 B | 690 B | 0.7 ms |
   | SPHINCS+ | 32 B | 17 KB | 7.5 ms |

3. **Migration Plan**
   - Hybrid signatures during transition
   - Automatic algorithm upgrade mechanism
   - Backwards compatibility for 2 years

---

## Support & Resources

### Documentation
- Technical Specs: [docs.quantum-blockchain.io](https://docs.quantum-blockchain.io)
- API Reference: [api.quantum-blockchain.io](https://api.quantum-blockchain.io)
- Video Tutorials: [youtube.com/quantum-blockchain](https://youtube.com/quantum-blockchain)

### Community
- Discord: [discord.gg/quantum-blockchain](https://discord.gg/quantum-blockchain)
- Telegram: [t.me/quantum_blockchain](https://t.me/quantum_blockchain)
- Forum: [forum.quantum-blockchain.io](https://forum.quantum-blockchain.io)

### Technical Support
- Email: validators@quantum-blockchain.io
- Emergency Hotline: +1-800-QUANTUM
- Office Hours: Mon-Fri 9am-5pm EST

### Bug Bounty
- Critical: Up to $100,000
- High: Up to $25,000
- Medium: Up to $5,000
- Low: Up to $1,000

Report at: [security@quantum-blockchain.io](mailto:security@quantum-blockchain.io)

---

## Conclusion

The Quantum Blockchain validator onboarding and token distribution system provides a secure, decentralized, and quantum-resistant foundation for the next generation of blockchain technology. With comprehensive tooling, clear economics, and strong security practices, validators can confidently participate in securing the network while earning rewards for their contributions.

For the latest updates and announcements, follow [@QuantumBlockchain](https://twitter.com/QuantumBlockchain) on Twitter.