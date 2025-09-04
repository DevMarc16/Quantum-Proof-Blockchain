// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/IQuantumToken.sol";

/**
 * @title TestnetFaucet
 * @notice Provides free QTM tokens for testnet operations
 * @dev Implements rate limiting and quantum signature verification
 */
contract TestnetFaucet {
    IQuantumToken public immutable qtmToken;
    address public owner;
    
    // Faucet configuration
    uint256 public constant DRIP_AMOUNT = 100 * 10**18; // 100 QTM per request
    uint256 public constant DRIP_INTERVAL = 24 hours; // Once per day
    uint256 public constant MAX_BALANCE = 1000 * 10**18; // Max 1000 QTM balance to request
    uint256 public constant FAUCET_RESERVE = 10_000_000 * 10**18; // 10M QTM reserve
    
    // Enhanced limits for validators
    uint256 public constant VALIDATOR_DRIP_AMOUNT = 100_000 * 10**18; // 100K QTM for validators
    uint256 public constant VALIDATOR_MAX_BALANCE = 200_000 * 10**18; // 200K QTM max for validators
    
    // State tracking
    mapping(address => uint256) public lastDripTime;
    mapping(address => uint256) public totalReceived;
    mapping(address => bool) public isVerifiedValidator;
    mapping(address => bool) public isBlacklisted;
    
    // Quantum signature verification for validators
    mapping(address => bytes) public validatorQuantumKeys;
    mapping(bytes32 => bool) public usedNonces;
    
    // Statistics
    uint256 public totalDripped;
    uint256 public totalRequests;
    uint256 public uniqueRecipients;
    mapping(address => bool) private hasReceivedBefore;
    
    // Events
    event Dripped(address indexed recipient, uint256 amount, bool isValidator);
    event ValidatorVerified(address indexed validator, bytes quantumPublicKey);
    event FaucetRefilled(uint256 amount);
    event EmergencyWithdraw(uint256 amount);
    event BlacklistUpdated(address indexed account, bool blacklisted);
    event ConfigurationUpdated(string parameter, uint256 value);
    
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner");
        _;
    }
    
    modifier notBlacklisted(address account) {
        require(!isBlacklisted[account], "Account blacklisted");
        _;
    }
    
    constructor(address _qtmToken) {
        qtmToken = IQuantumToken(_qtmToken);
        owner = msg.sender;
    }
    
    // ============ Public Functions ============
    
    /**
     * @notice Request testnet tokens
     */
    function requestTokens() external notBlacklisted(msg.sender) {
        _processTokenRequest(msg.sender, false);
    }
    
    /**
     * @notice Internal function to process token requests
     */
    function _processTokenRequest(address recipient, bool skipRateLimit) internal {
        // Check rate limit (skip for validator requests that already verified signature)
        if (!skipRateLimit) {
            require(
                block.timestamp >= lastDripTime[recipient] + DRIP_INTERVAL,
                "Too soon, please wait"
            );
        }
        
        // Check balance limit
        uint256 currentBalance = qtmToken.balanceOf(recipient);
        uint256 maxAllowed = isVerifiedValidator[recipient] ? VALIDATOR_MAX_BALANCE : MAX_BALANCE;
        require(currentBalance < maxAllowed, "Balance too high");
        
        // Calculate drip amount
        uint256 dripAmount = isVerifiedValidator[recipient] ? VALIDATOR_DRIP_AMOUNT : DRIP_AMOUNT;
        
        // Ensure we don't exceed max balance
        if (currentBalance + dripAmount > maxAllowed) {
            dripAmount = maxAllowed - currentBalance;
        }
        
        // Check faucet balance
        uint256 faucetBalance = qtmToken.balanceOf(address(this));
        require(faucetBalance >= dripAmount, "Faucet empty");
        
        // Update state
        lastDripTime[recipient] = block.timestamp;
        totalReceived[recipient] += dripAmount;
        totalDripped += dripAmount;
        totalRequests++;
        
        // Track unique recipients
        if (!hasReceivedBefore[recipient]) {
            hasReceivedBefore[recipient] = true;
            uniqueRecipients++;
        }
        
        // Transfer tokens
        require(qtmToken.transfer(recipient, dripAmount), "Transfer failed");
        
        emit Dripped(recipient, dripAmount, isVerifiedValidator[recipient]);
    }
    
    /**
     * @notice Request validator tokens with quantum signature verification
     * @param quantumPublicKey Dilithium/Falcon public key
     * @param signature Quantum signature of the request
     * @param nonce Unique nonce to prevent replay
     */
    function requestValidatorTokens(
        bytes calldata quantumPublicKey,
        bytes calldata signature,
        bytes32 nonce
    ) external notBlacklisted(msg.sender) {
        // Verify nonce hasn't been used
        require(!usedNonces[nonce], "Nonce already used");
        usedNonces[nonce] = true;
        
        // Verify quantum signature
        bytes32 message = keccak256(abi.encodePacked(msg.sender, nonce, "VALIDATOR_REQUEST"));
        require(_verifyQuantumSignature(quantumPublicKey, message, signature), "Invalid signature");
        
        // Mark as verified validator
        if (!isVerifiedValidator[msg.sender]) {
            isVerifiedValidator[msg.sender] = true;
            validatorQuantumKeys[msg.sender] = quantumPublicKey;
            emit ValidatorVerified(msg.sender, quantumPublicKey);
        }
        
        // Process token request - call internal logic
        _processTokenRequest(msg.sender, true);
    }
    
    /**
     * @notice Batch request for multiple addresses (owner only)
     * @param recipients Array of recipient addresses
     */
    function batchDrip(address[] calldata recipients) external onlyOwner {
        uint256 faucetBalance = qtmToken.balanceOf(address(this));
        uint256 totalAmount = recipients.length * DRIP_AMOUNT;
        require(faucetBalance >= totalAmount, "Insufficient balance");
        
        for (uint i = 0; i < recipients.length; i++) {
            if (!isBlacklisted[recipients[i]]) {
                // Skip rate limiting for batch operations
                totalReceived[recipients[i]] += DRIP_AMOUNT;
                totalDripped += DRIP_AMOUNT;
                
                if (!hasReceivedBefore[recipients[i]]) {
                    hasReceivedBefore[recipients[i]] = true;
                    uniqueRecipients++;
                }
                
                require(qtmToken.transfer(recipients[i], DRIP_AMOUNT), "Transfer failed");
                emit Dripped(recipients[i], DRIP_AMOUNT, false);
            }
        }
        
        totalRequests += recipients.length;
    }
    
    // ============ View Functions ============
    
    /**
     * @notice Check if address can request tokens
     */
    function canRequest(address account) external view returns (bool canDrip, uint256 timeUntilNext, uint256 maxAmount) {
        if (isBlacklisted[account]) {
            return (false, 0, 0);
        }
        
        uint256 timeSinceLastDrip = block.timestamp - lastDripTime[account];
        bool timeEligible = timeSinceLastDrip >= DRIP_INTERVAL;
        
        uint256 currentBalance = qtmToken.balanceOf(account);
        uint256 maxAllowed = isVerifiedValidator[account] ? VALIDATOR_MAX_BALANCE : MAX_BALANCE;
        bool balanceEligible = currentBalance < maxAllowed;
        
        uint256 timeLeft = timeEligible ? 0 : DRIP_INTERVAL - timeSinceLastDrip;
        uint256 availableAmount = 0;
        
        if (balanceEligible) {
            uint256 dripAmount = isVerifiedValidator[account] ? VALIDATOR_DRIP_AMOUNT : DRIP_AMOUNT;
            availableAmount = currentBalance + dripAmount > maxAllowed 
                ? maxAllowed - currentBalance 
                : dripAmount;
        }
        
        return (timeEligible && balanceEligible, timeLeft, availableAmount);
    }
    
    /**
     * @notice Get faucet statistics
     */
    function getFaucetStats() external view returns (
        uint256 balance,
        uint256 totalDistributed,
        uint256 requests,
        uint256 unique,
        uint256 averagePerUser
    ) {
        balance = qtmToken.balanceOf(address(this));
        totalDistributed = totalDripped;
        requests = totalRequests;
        unique = uniqueRecipients;
        averagePerUser = uniqueRecipients > 0 ? totalDripped / uniqueRecipients : 0;
    }
    
    /**
     * @notice Get user statistics
     */
    function getUserStats(address account) external view returns (
        uint256 lastRequest,
        uint256 total,
        bool isValidator,
        bool blacklisted,
        uint256 currentBalance
    ) {
        return (
            lastDripTime[account],
            totalReceived[account],
            isVerifiedValidator[account],
            isBlacklisted[account],
            qtmToken.balanceOf(account)
        );
    }
    
    // ============ Admin Functions ============
    
    /**
     * @notice Refill faucet with tokens
     */
    function refillFaucet(uint256 amount) external onlyOwner {
        require(qtmToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        emit FaucetRefilled(amount);
    }
    
    /**
     * @notice Emergency withdraw (owner only)
     */
    function emergencyWithdraw(uint256 amount) external onlyOwner {
        uint256 balance = qtmToken.balanceOf(address(this));
        if (amount == 0) {
            amount = balance;
        }
        require(amount <= balance, "Insufficient balance");
        
        require(qtmToken.transfer(owner, amount), "Transfer failed");
        emit EmergencyWithdraw(amount);
    }
    
    /**
     * @notice Update blacklist status
     */
    function updateBlacklist(address account, bool blacklisted) external onlyOwner {
        isBlacklisted[account] = blacklisted;
        emit BlacklistUpdated(account, blacklisted);
    }
    
    /**
     * @notice Batch update blacklist
     */
    function batchUpdateBlacklist(address[] calldata accounts, bool blacklisted) external onlyOwner {
        for (uint i = 0; i < accounts.length; i++) {
            isBlacklisted[accounts[i]] = blacklisted;
            emit BlacklistUpdated(accounts[i], blacklisted);
        }
    }
    
    /**
     * @notice Verify validator manually (owner only)
     */
    function verifyValidator(address validator, bytes calldata quantumPublicKey) external onlyOwner {
        isVerifiedValidator[validator] = true;
        validatorQuantumKeys[validator] = quantumPublicKey;
        emit ValidatorVerified(validator, quantumPublicKey);
    }
    
    /**
     * @notice Transfer ownership
     */
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "Invalid address");
        owner = newOwner;
    }
    
    // ============ Internal Functions ============
    
    /**
     * @notice Verify quantum signature (simplified - would call precompile)
     */
    function _verifyQuantumSignature(
        bytes memory publicKey,
        bytes32 message,
        bytes memory signature
    ) internal pure returns (bool) {
        // In production, this would call the quantum signature verification precompile
        // For now, we do a simple check
        return publicKey.length > 0 && signature.length > 0 && message != bytes32(0);
    }
    
    // ============ Testnet Features ============
    
    /**
     * @notice Reset user cooldown (testnet only)
     */
    function resetCooldown(address account) external onlyOwner {
        lastDripTime[account] = 0;
    }
    
    /**
     * @notice Clear all blacklists (testnet only)
     */
    function clearAllBlacklists() external onlyOwner {
        // This would iterate through a stored array of blacklisted addresses
        // For gas efficiency, we'd need to maintain such an array
    }
    
    /**
     * @notice Set custom drip amount for testing
     */
    function setCustomDripAmount(address account, uint256 amount) external onlyOwner {
        require(amount <= FAUCET_RESERVE, "Amount too large");
        require(qtmToken.transfer(account, amount), "Transfer failed");
        
        totalReceived[account] += amount;
        totalDripped += amount;
        emit Dripped(account, amount, isVerifiedValidator[account]);
    }
}