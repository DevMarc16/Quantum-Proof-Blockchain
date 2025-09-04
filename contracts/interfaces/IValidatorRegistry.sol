// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title IValidatorRegistry
 * @notice Interface for the Quantum Blockchain Validator Registry
 */
interface IValidatorRegistry {
    
    // Validator Management
    function registerValidator(
        bytes calldata quantumPublicKey,
        uint8 sigAlgorithm,
        uint256 initialStake,
        uint256 commissionRate,
        string calldata metadata
    ) external;
    
    function updateValidatorKeys(
        bytes calldata newPublicKey,
        uint8 sigAlgorithm,
        bytes calldata proof
    ) external;
    
    function addStake(uint256 amount) external;
    function withdrawStake(uint256 amount) external;
    function updateCommissionRate(uint256 newRate) external;
    
    // Delegation Management
    function delegate(address validator, uint256 amount) external;
    function undelegate(address validator, uint256 amount) external;
    function completeUnbonding(uint256 unbondingId) external;
    
    // Rewards
    function claimRewards() external;
    function distributeBlockReward(address validator, uint256 blockReward) external;
    
    // Slashing & Jail Management
    function slashValidator(address validator, uint8 reason) external;
    function unjailValidator() external;
    function updateValidatorPerformance(address validator, bool missed) external;
    
    // View Functions
    function getActiveValidators() external view returns (address[] memory);
    function getValidatorCount() external view returns (uint256 total, uint256 active);
    function getTotalStaked() external view returns (uint256);
    function getTotalDelegated() external view returns (uint256);
    function getValidatorPerformance(address validator) external view returns (uint256 missedBlocks, uint256 totalBlocks, uint256 uptime);
    function isValidatorActive(address validator) external view returns (bool);
    
    // Constants
    function MIN_STAKE() external view returns (uint256);
    function MAX_STAKE() external view returns (uint256);
    function MIN_DELEGATION() external view returns (uint256);
    function UNBONDING_PERIOD() external view returns (uint256);
    function MAX_VALIDATORS() external view returns (uint256);
}