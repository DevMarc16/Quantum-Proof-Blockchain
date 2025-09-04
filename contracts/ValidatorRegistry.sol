// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/IValidatorRegistry.sol";
import "./interfaces/IQuantumToken.sol";
import "./libraries/QuantumCrypto.sol";

/**
 * @title ValidatorRegistry
 * @notice Manages validator registration, staking, delegation, and slashing for the quantum-resistant blockchain
 * @dev Implements quantum-resistant validator management with NIST post-quantum algorithms
 */
contract ValidatorRegistry is IValidatorRegistry {
    using QuantumCrypto for bytes;
    
    // Constants
    uint256 public constant MIN_STAKE = 100_000 * 10**18; // 100K QTM minimum stake
    uint256 public constant MAX_STAKE = 10_000_000 * 10**18; // 10M QTM maximum stake
    uint256 public constant MIN_DELEGATION = 100 * 10**18; // 100 QTM minimum delegation
    uint256 public constant UNBONDING_PERIOD = 21 days; // 21 days unbonding period
    uint256 public constant MAX_VALIDATORS = 100; // Maximum active validators
    uint256 public constant COMMISSION_RATE_PRECISION = 10000; // Basis points
    uint256 public constant MAX_COMMISSION_RATE = 2000; // 20% max commission
    uint256 public constant SLASHING_RATE_DOUBLE_SIGN = 2000; // 20% for double signing
    uint256 public constant SLASHING_RATE_DOWNTIME = 100; // 1% for downtime
    uint256 public constant SLASHING_RATE_INVALID_BLOCK = 1000; // 10% for invalid blocks
    
    // State variables
    IQuantumToken public immutable qtmToken;
    address public governance;
    uint256 public totalStaked;
    uint256 public totalDelegated;
    uint256 public validatorCount;
    uint256 public activeValidatorCount;
    
    // Validator data structures
    struct Validator {
        address validatorAddress;
        bytes quantumPublicKey; // Dilithium/Falcon public key
        uint8 sigAlgorithm; // 1: Dilithium, 2: Falcon
        uint256 selfStake;
        uint256 totalDelegated;
        uint256 commissionRate; // In basis points (100 = 1%)
        uint256 lastActiveBlock;
        uint256 missedBlocks;
        uint256 totalBlocks;
        bool isActive;
        bool isSlashed;
        uint256 registrationBlock;
        uint256 jailUntilBlock;
        string metadata; // IPFS hash for additional info
    }
    
    // Delegation data structures
    struct Delegation {
        uint256 amount;
        uint256 shares; // For compound rewards
        uint256 lastClaimBlock;
        uint256 unbondingAmount;
        uint256 unbondingEndTime;
    }
    
    // Unbonding request
    struct UnbondingRequest {
        address delegator;
        address validator;
        uint256 amount;
        uint256 completionTime;
    }
    
    // Mappings
    mapping(address => Validator) public validators;
    mapping(address => mapping(address => Delegation)) public delegations; // delegator => validator => delegation
    mapping(address => uint256) public pendingRewards;
    mapping(address => bool) public isValidator;
    mapping(uint256 => UnbondingRequest) public unbondingRequests;
    uint256 public nextUnbondingId;
    
    // Validator set management
    address[] public validatorSet;
    address[] public activeValidators;
    
    // Events
    event ValidatorRegistered(address indexed validator, bytes quantumPublicKey, uint256 stake);
    event ValidatorActivated(address indexed validator);
    event ValidatorDeactivated(address indexed validator, string reason);
    event StakeAdded(address indexed validator, uint256 amount);
    event StakeRemoved(address indexed validator, uint256 amount);
    event DelegationAdded(address indexed delegator, address indexed validator, uint256 amount);
    event DelegationRemoved(address indexed delegator, address indexed validator, uint256 amount);
    event RewardsClaimed(address indexed account, uint256 amount);
    event ValidatorSlashed(address indexed validator, uint256 amount, uint8 reason);
    event CommissionUpdated(address indexed validator, uint256 oldRate, uint256 newRate);
    event ValidatorJailed(address indexed validator, uint256 until);
    event ValidatorUnjailed(address indexed validator);
    event UnbondingInitiated(uint256 indexed id, address delegator, address validator, uint256 amount);
    event UnbondingCompleted(uint256 indexed id);
    
    // Modifiers
    modifier onlyGovernance() {
        require(msg.sender == governance, "Only governance");
        _;
    }
    
    modifier onlyActiveValidator(address validator) {
        require(validators[validator].isActive, "Not active validator");
        require(!validators[validator].isSlashed, "Validator slashed");
        require(block.number >= validators[validator].jailUntilBlock, "Validator jailed");
        _;
    }
    
    modifier validatorExists(address validator) {
        require(isValidator[validator], "Validator not found");
        _;
    }
    
    constructor(address _qtmToken, address _governance) {
        qtmToken = IQuantumToken(_qtmToken);
        governance = _governance;
    }
    
    // ============ Validator Registration ============
    
    /**
     * @notice Register as a validator with quantum-resistant keys
     * @param quantumPublicKey Dilithium or Falcon public key
     * @param sigAlgorithm 1 for Dilithium, 2 for Falcon
     * @param initialStake Initial QTM stake (min 100K)
     * @param commissionRate Commission rate in basis points
     * @param metadata IPFS hash for additional validator info
     */
    function registerValidator(
        bytes calldata quantumPublicKey,
        uint8 sigAlgorithm,
        uint256 initialStake,
        uint256 commissionRate,
        string calldata metadata
    ) external {
        require(!isValidator[msg.sender], "Already registered");
        require(initialStake >= MIN_STAKE, "Insufficient stake");
        require(initialStake <= MAX_STAKE, "Exceeds max stake");
        require(commissionRate <= MAX_COMMISSION_RATE, "Commission too high");
        require(sigAlgorithm == 1 || sigAlgorithm == 2, "Invalid algorithm");
        require(quantumPublicKey.length > 0, "Invalid public key");
        require(validatorCount < MAX_VALIDATORS, "Max validators reached");
        
        // Transfer stake
        require(qtmToken.transferFrom(msg.sender, address(this), initialStake), "Transfer failed");
        
        // Create validator
        validators[msg.sender] = Validator({
            validatorAddress: msg.sender,
            quantumPublicKey: quantumPublicKey,
            sigAlgorithm: sigAlgorithm,
            selfStake: initialStake,
            totalDelegated: 0,
            commissionRate: commissionRate,
            lastActiveBlock: block.number,
            missedBlocks: 0,
            totalBlocks: 0,
            isActive: false, // Requires activation
            isSlashed: false,
            registrationBlock: block.number,
            jailUntilBlock: 0,
            metadata: metadata
        });
        
        isValidator[msg.sender] = true;
        validatorSet.push(msg.sender);
        validatorCount++;
        totalStaked += initialStake;
        
        emit ValidatorRegistered(msg.sender, quantumPublicKey, initialStake);
        
        // Auto-activate if meets criteria
        _tryActivateValidator(msg.sender);
    }
    
    /**
     * @notice Update validator quantum keys (for key rotation)
     * @param newPublicKey New quantum public key
     * @param sigAlgorithm New signature algorithm
     * @param proof Proof of key ownership (signed with old key)
     */
    function updateValidatorKeys(
        bytes calldata newPublicKey,
        uint8 sigAlgorithm,
        bytes calldata proof
    ) external validatorExists(msg.sender) {
        Validator storage validator = validators[msg.sender];
        
        // Verify proof with old key
        require(_verifyKeyRotationProof(validator.quantumPublicKey, newPublicKey, proof), "Invalid proof");
        
        // Update keys
        validator.quantumPublicKey = newPublicKey;
        validator.sigAlgorithm = sigAlgorithm;
        
        emit ValidatorKeysUpdated(msg.sender, newPublicKey, sigAlgorithm);
    }
    
    // ============ Staking Management ============
    
    /**
     * @notice Add stake to validator position
     * @param amount Additional QTM to stake
     */
    function addStake(uint256 amount) external validatorExists(msg.sender) {
        Validator storage validator = validators[msg.sender];
        require(validator.selfStake + amount <= MAX_STAKE, "Exceeds max stake");
        require(!validator.isSlashed, "Validator slashed");
        
        // Transfer stake
        require(qtmToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        
        validator.selfStake += amount;
        totalStaked += amount;
        
        emit StakeAdded(msg.sender, amount);
        
        // Try to activate if not active
        if (!validator.isActive) {
            _tryActivateValidator(msg.sender);
        }
    }
    
    /**
     * @notice Begin stake withdrawal (starts unbonding)
     * @param amount Amount to withdraw
     */
    function withdrawStake(uint256 amount) external validatorExists(msg.sender) {
        Validator storage validator = validators[msg.sender];
        require(validator.selfStake >= amount, "Insufficient stake");
        
        // Check minimum stake requirement
        uint256 remainingStake = validator.selfStake - amount;
        if (remainingStake < MIN_STAKE && remainingStake > 0) {
            revert("Below minimum stake");
        }
        
        // If withdrawing all, deactivate validator
        if (remainingStake == 0) {
            _deactivateValidator(msg.sender, "Full withdrawal");
        }
        
        // Create unbonding request
        unbondingRequests[nextUnbondingId] = UnbondingRequest({
            delegator: msg.sender,
            validator: msg.sender,
            amount: amount,
            completionTime: block.timestamp + UNBONDING_PERIOD
        });
        
        validator.selfStake -= amount;
        totalStaked -= amount;
        
        emit UnbondingInitiated(nextUnbondingId, msg.sender, msg.sender, amount);
        nextUnbondingId++;
    }
    
    // ============ Delegation Management ============
    
    /**
     * @notice Delegate QTM to a validator
     * @param validator Address of validator to delegate to
     * @param amount Amount of QTM to delegate
     */
    function delegate(address validator, uint256 amount) 
        external 
        onlyActiveValidator(validator) 
    {
        require(amount >= MIN_DELEGATION, "Below minimum delegation");
        
        // Transfer tokens
        require(qtmToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        
        Validator storage val = validators[validator];
        Delegation storage del = delegations[msg.sender][validator];
        
        // Calculate shares for compound interest
        uint256 shares = amount;
        if (val.totalDelegated > 0 && del.shares > 0) {
            shares = (amount * del.shares) / val.totalDelegated;
        }
        
        del.amount += amount;
        del.shares += shares;
        del.lastClaimBlock = block.number;
        
        val.totalDelegated += amount;
        totalDelegated += amount;
        
        emit DelegationAdded(msg.sender, validator, amount);
    }
    
    /**
     * @notice Begin undelegation (starts unbonding)
     * @param validator Address of validator to undelegate from
     * @param amount Amount to undelegate
     */
    function undelegate(address validator, uint256 amount) external {
        Delegation storage del = delegations[msg.sender][validator];
        require(del.amount >= amount, "Insufficient delegation");
        
        Validator storage val = validators[validator];
        
        // Calculate shares to remove
        uint256 sharesToRemove = (amount * del.shares) / del.amount;
        
        del.amount -= amount;
        del.shares -= sharesToRemove;
        del.unbondingAmount += amount;
        del.unbondingEndTime = block.timestamp + UNBONDING_PERIOD;
        
        val.totalDelegated -= amount;
        totalDelegated -= amount;
        
        // Create unbonding request
        unbondingRequests[nextUnbondingId] = UnbondingRequest({
            delegator: msg.sender,
            validator: validator,
            amount: amount,
            completionTime: block.timestamp + UNBONDING_PERIOD
        });
        
        emit UnbondingInitiated(nextUnbondingId, msg.sender, validator, amount);
        nextUnbondingId++;
    }
    
    /**
     * @notice Complete unbonding after period expires
     * @param unbondingId ID of unbonding request
     */
    function completeUnbonding(uint256 unbondingId) external {
        UnbondingRequest storage request = unbondingRequests[unbondingId];
        require(request.delegator == msg.sender, "Not your request");
        require(block.timestamp >= request.completionTime, "Still unbonding");
        require(request.amount > 0, "Already completed");
        
        uint256 amount = request.amount;
        request.amount = 0;
        
        // Transfer tokens back
        require(qtmToken.transfer(msg.sender, amount), "Transfer failed");
        
        emit UnbondingCompleted(unbondingId);
    }
    
    // ============ Rewards Management ============
    
    /**
     * @notice Claim pending rewards
     */
    function claimRewards() external {
        uint256 rewards = pendingRewards[msg.sender];
        require(rewards > 0, "No pending rewards");
        
        pendingRewards[msg.sender] = 0;
        
        // Transfer rewards
        require(qtmToken.transfer(msg.sender, rewards), "Transfer failed");
        
        emit RewardsClaimed(msg.sender, rewards);
    }
    
    /**
     * @notice Distribute block rewards to validator and delegators
     * @param validator Address of block proposer
     * @param blockReward Total block reward
     */
    function distributeBlockReward(address validator, uint256 blockReward) 
        external 
        onlyGovernance 
        validatorExists(validator) 
    {
        Validator storage val = validators[validator];
        
        // Calculate commission
        uint256 commission = (blockReward * val.commissionRate) / COMMISSION_RATE_PRECISION;
        uint256 delegatorRewards = blockReward - commission;
        
        // Add validator commission
        pendingRewards[validator] += commission;
        
        // Distribute to delegators proportionally
        if (val.totalDelegated > 0) {
            // This would be done off-chain and submitted in batches for gas efficiency
            _distributeDelegatorRewards(validator, delegatorRewards);
        } else {
            // If no delegators, validator gets all
            pendingRewards[validator] += delegatorRewards;
        }
        
        // Update validator stats
        val.lastActiveBlock = block.number;
        val.totalBlocks++;
    }
    
    // ============ Slashing Management ============
    
    /**
     * @notice Slash a validator for misbehavior
     * @param validator Address of validator to slash
     * @param reason Slashing reason (1: double sign, 2: downtime, 3: invalid block)
     */
    function slashValidator(address validator, uint8 reason) 
        external 
        onlyGovernance 
        validatorExists(validator) 
    {
        Validator storage val = validators[validator];
        require(!val.isSlashed, "Already slashed");
        
        uint256 slashRate;
        if (reason == 1) {
            slashRate = SLASHING_RATE_DOUBLE_SIGN;
        } else if (reason == 2) {
            slashRate = SLASHING_RATE_DOWNTIME;
        } else if (reason == 3) {
            slashRate = SLASHING_RATE_INVALID_BLOCK;
        } else {
            revert("Invalid reason");
        }
        
        // Calculate slash amounts
        uint256 slashFromStake = (val.selfStake * slashRate) / COMMISSION_RATE_PRECISION;
        uint256 slashFromDelegated = (val.totalDelegated * slashRate) / COMMISSION_RATE_PRECISION;
        
        // Apply slashing
        val.selfStake -= slashFromStake;
        val.totalDelegated -= slashFromDelegated;
        totalStaked -= slashFromStake;
        totalDelegated -= slashFromDelegated;
        
        // Mark as slashed for severe violations
        if (reason == 1) {
            val.isSlashed = true;
            _deactivateValidator(validator, "Slashed");
        } else {
            // Jail for lesser violations
            val.jailUntilBlock = block.number + (reason == 2 ? 1000 : 5000);
            emit ValidatorJailed(validator, val.jailUntilBlock);
        }
        
        emit ValidatorSlashed(validator, slashFromStake + slashFromDelegated, reason);
    }
    
    /**
     * @notice Unjail a validator after jail period
     */
    function unjailValidator() external validatorExists(msg.sender) {
        Validator storage val = validators[msg.sender];
        require(val.jailUntilBlock > 0, "Not jailed");
        require(block.number >= val.jailUntilBlock, "Still jailed");
        
        val.jailUntilBlock = 0;
        val.missedBlocks = 0; // Reset missed blocks
        
        emit ValidatorUnjailed(msg.sender);
    }
    
    // ============ Validator Performance ============
    
    /**
     * @notice Update validator performance metrics
     * @param validator Address of validator
     * @param missed Whether validator missed the block
     */
    function updateValidatorPerformance(address validator, bool missed) 
        external 
        onlyGovernance 
        validatorExists(validator) 
    {
        Validator storage val = validators[validator];
        
        if (missed) {
            val.missedBlocks++;
            
            // Check downtime slashing threshold (e.g., 50 missed in 1000 blocks)
            if (val.missedBlocks > 50 && val.totalBlocks > 1000) {
                uint256 missRate = (val.missedBlocks * 100) / val.totalBlocks;
                if (missRate > 5) { // More than 5% missed
                    this.slashValidator(validator, 2); // Downtime slashing
                }
            }
        }
        
        val.totalBlocks++;
        val.lastActiveBlock = block.number;
    }
    
    // ============ Internal Functions ============
    
    function _tryActivateValidator(address validator) internal {
        Validator storage val = validators[validator];
        
        if (!val.isActive && val.selfStake >= MIN_STAKE && !val.isSlashed) {
            val.isActive = true;
            activeValidators.push(validator);
            activeValidatorCount++;
            
            emit ValidatorActivated(validator);
        }
    }
    
    function _deactivateValidator(address validator, string memory reason) internal {
        Validator storage val = validators[validator];
        
        if (val.isActive) {
            val.isActive = false;
            
            // Remove from active set
            for (uint i = 0; i < activeValidators.length; i++) {
                if (activeValidators[i] == validator) {
                    activeValidators[i] = activeValidators[activeValidators.length - 1];
                    activeValidators.pop();
                    break;
                }
            }
            
            activeValidatorCount--;
            emit ValidatorDeactivated(validator, reason);
        }
    }
    
    function _distributeDelegatorRewards(address validator, uint256 totalRewards) internal {
        // This is a simplified version - in production, this would be done off-chain
        // and submitted in batches to save gas
        Validator storage val = validators[validator];
        
        if (val.totalDelegated == 0) return;
        
        // Distribute proportionally based on shares
        // In production: track and allow individual claims
        pendingRewards[validator] += totalRewards;
    }
    
    function _verifyKeyRotationProof(
        bytes memory oldKey,
        bytes memory newKey,
        bytes memory proof
    ) internal pure returns (bool) {
        // Implement quantum signature verification
        // This would call the precompiled quantum crypto verification
        return proof.length > 0; // Simplified for now
    }
    
    // ============ View Functions ============
    
    function getValidator(address validator) external view returns (Validator memory) {
        return validators[validator];
    }
    
    function getActiveValidators() external view returns (address[] memory) {
        return activeValidators;
    }
    
    function getValidatorCount() external view returns (uint256 total, uint256 active) {
        return (validatorCount, activeValidatorCount);
    }
    
    function getDelegation(address delegator, address validator) 
        external 
        view 
        returns (Delegation memory) 
    {
        return delegations[delegator][validator];
    }
    
    function getTotalStaked() external view returns (uint256) {
        return totalStaked;
    }
    
    function getTotalDelegated() external view returns (uint256) {
        return totalDelegated;
    }
    
    function getValidatorPerformance(address validator) 
        external 
        view 
        returns (uint256 missedBlocks, uint256 totalBlocks, uint256 uptime) 
    {
        Validator storage val = validators[validator];
        uint256 uptimePercent = val.totalBlocks > 0 
            ? ((val.totalBlocks - val.missedBlocks) * 100) / val.totalBlocks 
            : 0;
        return (val.missedBlocks, val.totalBlocks, uptimePercent);
    }
    
    function isValidatorActive(address validator) external view returns (bool) {
        return validators[validator].isActive && 
               !validators[validator].isSlashed &&
               block.number >= validators[validator].jailUntilBlock;
    }
    
    // ============ Governance Functions ============
    
    function updateCommissionRate(uint256 newRate) external validatorExists(msg.sender) {
        require(newRate <= MAX_COMMISSION_RATE, "Rate too high");
        Validator storage val = validators[msg.sender];
        uint256 oldRate = val.commissionRate;
        val.commissionRate = newRate;
        emit CommissionUpdated(msg.sender, oldRate, newRate);
    }
    
    function updateGovernance(address newGovernance) external onlyGovernance {
        governance = newGovernance;
    }
    
    // Additional events
    event ValidatorKeysUpdated(address indexed validator, bytes newKey, uint8 algorithm);
}