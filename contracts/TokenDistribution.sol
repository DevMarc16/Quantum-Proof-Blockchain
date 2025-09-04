// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/IQuantumToken.sol";

/**
 * @title TokenDistribution
 * @notice Manages QTM token distribution including genesis allocation, vesting, and public sale
 * @dev Implements a comprehensive token distribution strategy for the quantum blockchain
 */
contract TokenDistribution {
    // Token allocation percentages (out of 1000 for precision)
    uint256 public constant GENESIS_VALIDATORS = 150; // 15% - Initial validators
    uint256 public constant PUBLIC_SALE = 250; // 25% - Public sale/IDO
    uint256 public constant ECOSYSTEM_FUND = 200; // 20% - Ecosystem development
    uint256 public constant TEAM_ALLOCATION = 150; // 15% - Team (vested)
    uint256 public constant ADVISORS = 50; // 5% - Advisors (vested)
    uint256 public constant LIQUIDITY_PROVISION = 100; // 10% - DEX liquidity
    uint256 public constant STAKING_REWARDS = 50; // 5% - Initial staking rewards
    uint256 public constant COMMUNITY_AIRDROPS = 50; // 5% - Community airdrops
    
    uint256 public constant TOTAL_SUPPLY = 1_000_000_000 * 10**18; // 1B QTM
    uint256 public constant VESTING_DURATION = 365 days * 2; // 2 years vesting
    uint256 public constant CLIFF_DURATION = 180 days; // 6 months cliff
    uint256 public constant TGE_UNLOCK = 100; // 10% unlock at TGE for vested tokens
    
    IQuantumToken public immutable qtmToken;
    address public governance;
    uint256 public tgeTimestamp; // Token Generation Event timestamp
    bool public initialized;
    
    // Distribution tracking
    mapping(DistributionCategory => uint256) public categoryAllocations;
    mapping(DistributionCategory => uint256) public categoryDistributed;
    
    // Vesting schedules
    struct VestingSchedule {
        uint256 totalAmount;
        uint256 startTime;
        uint256 cliffTime;
        uint256 duration;
        uint256 released;
        bool revocable;
        bool revoked;
    }
    
    mapping(address => VestingSchedule) public vestingSchedules;
    mapping(address => bool) public hasVesting;
    
    // Public sale parameters
    struct PublicSaleConfig {
        uint256 startTime;
        uint256 endTime;
        uint256 minContribution;
        uint256 maxContribution;
        uint256 tokenPrice; // In wei per token
        uint256 hardCap;
        uint256 totalRaised;
        bool finalized;
    }
    
    PublicSaleConfig public publicSale;
    mapping(address => uint256) public publicSaleContributions;
    mapping(address => bool) public publicSaleWhitelist;
    bool public whitelistRequired = true;
    
    // Airdrop management
    struct AirdropCampaign {
        string name;
        uint256 totalAmount;
        uint256 distributed;
        uint256 claimDeadline;
        bytes32 merkleRoot;
        mapping(address => bool) claimed;
    }
    
    mapping(uint256 => AirdropCampaign) public airdropCampaigns;
    uint256 public nextAirdropId;
    
    // Genesis validator allocation
    mapping(address => uint256) public genesisValidatorAllocations;
    address[] public genesisValidators;
    
    // Liquidity provision
    struct LiquidityProvision {
        address dexAddress;
        uint256 tokenAmount;
        uint256 ethAmount;
        uint256 unlockTime;
        bool provided;
    }
    
    LiquidityProvision[] public liquidityProvisions;
    
    enum DistributionCategory {
        GenesisValidators,
        PublicSale,
        Ecosystem,
        Team,
        Advisors,
        Liquidity,
        StakingRewards,
        Airdrops
    }
    
    // Events
    event TokensAllocated(DistributionCategory category, uint256 amount);
    event VestingScheduleCreated(address beneficiary, uint256 amount);
    event TokensReleased(address beneficiary, uint256 amount);
    event PublicSaleConfigured(uint256 startTime, uint256 endTime, uint256 tokenPrice);
    event PublicSaleContribution(address contributor, uint256 ethAmount, uint256 tokenAmount);
    event PublicSaleFinalized(uint256 totalRaised, uint256 tokensSold);
    event AirdropCreated(uint256 campaignId, string name, uint256 amount);
    event AirdropClaimed(uint256 campaignId, address recipient, uint256 amount);
    event GenesisValidatorRegistered(address validator, uint256 allocation);
    event LiquidityProvided(address dex, uint256 tokenAmount, uint256 ethAmount);
    
    modifier onlyGovernance() {
        require(msg.sender == governance, "Only governance");
        _;
    }
    
    modifier afterTGE() {
        require(tgeTimestamp > 0 && block.timestamp >= tgeTimestamp, "TGE not started");
        _;
    }
    
    constructor(address _qtmToken, address _governance) {
        qtmToken = IQuantumToken(_qtmToken);
        governance = _governance;
        
        // Initialize category allocations
        categoryAllocations[DistributionCategory.GenesisValidators] = (TOTAL_SUPPLY * GENESIS_VALIDATORS) / 1000;
        categoryAllocations[DistributionCategory.PublicSale] = (TOTAL_SUPPLY * PUBLIC_SALE) / 1000;
        categoryAllocations[DistributionCategory.Ecosystem] = (TOTAL_SUPPLY * ECOSYSTEM_FUND) / 1000;
        categoryAllocations[DistributionCategory.Team] = (TOTAL_SUPPLY * TEAM_ALLOCATION) / 1000;
        categoryAllocations[DistributionCategory.Advisors] = (TOTAL_SUPPLY * ADVISORS) / 1000;
        categoryAllocations[DistributionCategory.Liquidity] = (TOTAL_SUPPLY * LIQUIDITY_PROVISION) / 1000;
        categoryAllocations[DistributionCategory.StakingRewards] = (TOTAL_SUPPLY * STAKING_REWARDS) / 1000;
        categoryAllocations[DistributionCategory.Airdrops] = (TOTAL_SUPPLY * COMMUNITY_AIRDROPS) / 1000;
    }
    
    // ============ Initialization ============
    
    /**
     * @notice Initialize token distribution (one-time setup)
     * @param _tgeTimestamp Token Generation Event timestamp
     */
    function initialize(uint256 _tgeTimestamp) external onlyGovernance {
        require(!initialized, "Already initialized");
        require(_tgeTimestamp >= block.timestamp, "Invalid TGE timestamp");
        
        tgeTimestamp = _tgeTimestamp;
        initialized = true;
    }
    
    // ============ Genesis Validator Allocation ============
    
    /**
     * @notice Register genesis validators and their allocations
     * @param validators Array of validator addresses
     * @param allocations Array of token allocations
     */
    function registerGenesisValidators(
        address[] calldata validators,
        uint256[] calldata allocations
    ) external onlyGovernance {
        require(validators.length == allocations.length, "Length mismatch");
        
        uint256 totalAllocation;
        for (uint i = 0; i < validators.length; i++) {
            require(genesisValidatorAllocations[validators[i]] == 0, "Already registered");
            
            genesisValidatorAllocations[validators[i]] = allocations[i];
            genesisValidators.push(validators[i]);
            totalAllocation += allocations[i];
            
            emit GenesisValidatorRegistered(validators[i], allocations[i]);
        }
        
        require(
            categoryDistributed[DistributionCategory.GenesisValidators] + totalAllocation 
            <= categoryAllocations[DistributionCategory.GenesisValidators],
            "Exceeds allocation"
        );
        
        categoryDistributed[DistributionCategory.GenesisValidators] += totalAllocation;
    }
    
    /**
     * @notice Claim genesis validator allocation
     */
    function claimGenesisAllocation() external afterTGE {
        uint256 allocation = genesisValidatorAllocations[msg.sender];
        require(allocation > 0, "No allocation");
        
        genesisValidatorAllocations[msg.sender] = 0;
        
        require(qtmToken.transfer(msg.sender, allocation), "Transfer failed");
        emit TokensReleased(msg.sender, allocation);
    }
    
    // ============ Vesting Management ============
    
    /**
     * @notice Create vesting schedule for team/advisors
     * @param beneficiary Address to receive vested tokens
     * @param amount Total amount to vest
     * @param category Distribution category (Team or Advisors)
     * @param revocable Whether vesting can be revoked
     */
    function createVestingSchedule(
        address beneficiary,
        uint256 amount,
        DistributionCategory category,
        bool revocable
    ) external onlyGovernance {
        require(!hasVesting[beneficiary], "Already has vesting");
        require(category == DistributionCategory.Team || category == DistributionCategory.Advisors, "Invalid category");
        
        require(
            categoryDistributed[category] + amount <= categoryAllocations[category],
            "Exceeds allocation"
        );
        
        vestingSchedules[beneficiary] = VestingSchedule({
            totalAmount: amount,
            startTime: tgeTimestamp,
            cliffTime: tgeTimestamp + CLIFF_DURATION,
            duration: VESTING_DURATION,
            released: 0,
            revocable: revocable,
            revoked: false
        });
        
        hasVesting[beneficiary] = true;
        categoryDistributed[category] += amount;
        
        emit VestingScheduleCreated(beneficiary, amount);
    }
    
    /**
     * @notice Release vested tokens
     */
    function releaseVestedTokens() external {
        require(hasVesting[msg.sender], "No vesting schedule");
        VestingSchedule storage schedule = vestingSchedules[msg.sender];
        require(!schedule.revoked, "Vesting revoked");
        
        uint256 releasable = _computeReleasableAmount(schedule);
        require(releasable > 0, "No tokens to release");
        
        schedule.released += releasable;
        require(qtmToken.transfer(msg.sender, releasable), "Transfer failed");
        
        emit TokensReleased(msg.sender, releasable);
    }
    
    function _computeReleasableAmount(VestingSchedule memory schedule) internal view returns (uint256) {
        if (block.timestamp < schedule.cliffTime) {
            return 0;
        }
        
        uint256 tgeRelease = (schedule.totalAmount * TGE_UNLOCK) / 1000;
        uint256 vestedAmount;
        
        if (block.timestamp >= schedule.startTime + schedule.duration) {
            vestedAmount = schedule.totalAmount;
        } else {
            uint256 timeFromStart = block.timestamp - schedule.startTime;
            uint256 vestingAmount = schedule.totalAmount - tgeRelease;
            vestedAmount = tgeRelease + (vestingAmount * timeFromStart) / schedule.duration;
        }
        
        return vestedAmount - schedule.released;
    }
    
    // ============ Public Sale Management ============
    
    /**
     * @notice Configure public sale parameters
     */
    function configurePublicSale(
        uint256 startTime,
        uint256 endTime,
        uint256 minContribution,
        uint256 maxContribution,
        uint256 tokenPrice,
        uint256 hardCap
    ) external onlyGovernance {
        require(startTime > block.timestamp, "Invalid start time");
        require(endTime > startTime, "Invalid end time");
        require(tokenPrice > 0, "Invalid price");
        
        publicSale = PublicSaleConfig({
            startTime: startTime,
            endTime: endTime,
            minContribution: minContribution,
            maxContribution: maxContribution,
            tokenPrice: tokenPrice,
            hardCap: hardCap,
            totalRaised: 0,
            finalized: false
        });
        
        emit PublicSaleConfigured(startTime, endTime, tokenPrice);
    }
    
    /**
     * @notice Contribute to public sale
     */
    function contributeToPublicSale() external payable {
        PublicSaleConfig memory sale = publicSale;
        require(block.timestamp >= sale.startTime, "Sale not started");
        require(block.timestamp <= sale.endTime, "Sale ended");
        require(msg.value >= sale.minContribution, "Below minimum");
        require(publicSaleContributions[msg.sender] + msg.value <= sale.maxContribution, "Exceeds maximum");
        require(sale.totalRaised + msg.value <= sale.hardCap, "Exceeds hard cap");
        
        if (whitelistRequired) {
            require(publicSaleWhitelist[msg.sender], "Not whitelisted");
        }
        
        uint256 tokenAmount = (msg.value * 10**18) / sale.tokenPrice;
        
        publicSaleContributions[msg.sender] += msg.value;
        publicSale.totalRaised += msg.value;
        
        require(qtmToken.transfer(msg.sender, tokenAmount), "Transfer failed");
        
        emit PublicSaleContribution(msg.sender, msg.value, tokenAmount);
    }
    
    /**
     * @notice Finalize public sale and withdraw funds
     */
    function finalizePublicSale() external onlyGovernance {
        require(block.timestamp > publicSale.endTime, "Sale not ended");
        require(!publicSale.finalized, "Already finalized");
        
        publicSale.finalized = true;
        
        uint256 tokensSold = (publicSale.totalRaised * 10**18) / publicSale.tokenPrice;
        categoryDistributed[DistributionCategory.PublicSale] += tokensSold;
        
        // Transfer raised ETH to governance
        payable(governance).transfer(publicSale.totalRaised);
        
        emit PublicSaleFinalized(publicSale.totalRaised, tokensSold);
    }
    
    // ============ Airdrop Management ============
    
    /**
     * @notice Create new airdrop campaign
     */
    function createAirdropCampaign(
        string calldata name,
        uint256 totalAmount,
        uint256 claimDuration,
        bytes32 merkleRoot
    ) external onlyGovernance {
        require(
            categoryDistributed[DistributionCategory.Airdrops] + totalAmount 
            <= categoryAllocations[DistributionCategory.Airdrops],
            "Exceeds allocation"
        );
        
        uint256 campaignId = nextAirdropId++;
        AirdropCampaign storage campaign = airdropCampaigns[campaignId];
        
        campaign.name = name;
        campaign.totalAmount = totalAmount;
        campaign.claimDeadline = block.timestamp + claimDuration;
        campaign.merkleRoot = merkleRoot;
        
        categoryDistributed[DistributionCategory.Airdrops] += totalAmount;
        
        emit AirdropCreated(campaignId, name, totalAmount);
    }
    
    /**
     * @notice Claim airdrop tokens with Merkle proof
     */
    function claimAirdrop(
        uint256 campaignId,
        uint256 amount,
        bytes32[] calldata merkleProof
    ) external {
        AirdropCampaign storage campaign = airdropCampaigns[campaignId];
        require(block.timestamp <= campaign.claimDeadline, "Claim period ended");
        require(!campaign.claimed[msg.sender], "Already claimed");
        
        // Verify Merkle proof
        bytes32 leaf = keccak256(abi.encodePacked(msg.sender, amount));
        require(_verifyMerkleProof(merkleProof, campaign.merkleRoot, leaf), "Invalid proof");
        
        campaign.claimed[msg.sender] = true;
        campaign.distributed += amount;
        require(campaign.distributed <= campaign.totalAmount, "Exceeds campaign allocation");
        
        require(qtmToken.transfer(msg.sender, amount), "Transfer failed");
        
        emit AirdropClaimed(campaignId, msg.sender, amount);
    }
    
    function _verifyMerkleProof(
        bytes32[] memory proof,
        bytes32 root,
        bytes32 leaf
    ) internal pure returns (bool) {
        bytes32 computedHash = leaf;
        
        for (uint256 i = 0; i < proof.length; i++) {
            bytes32 proofElement = proof[i];
            
            if (computedHash <= proofElement) {
                computedHash = keccak256(abi.encodePacked(computedHash, proofElement));
            } else {
                computedHash = keccak256(abi.encodePacked(proofElement, computedHash));
            }
        }
        
        return computedHash == root;
    }
    
    // ============ Liquidity Provision ============
    
    /**
     * @notice Schedule liquidity provision
     */
    function scheduleLiquidityProvision(
        address dexAddress,
        uint256 tokenAmount,
        uint256 ethAmount,
        uint256 unlockTime
    ) external onlyGovernance {
        require(
            categoryDistributed[DistributionCategory.Liquidity] + tokenAmount 
            <= categoryAllocations[DistributionCategory.Liquidity],
            "Exceeds allocation"
        );
        
        liquidityProvisions.push(LiquidityProvision({
            dexAddress: dexAddress,
            tokenAmount: tokenAmount,
            ethAmount: ethAmount,
            unlockTime: unlockTime,
            provided: false
        }));
        
        categoryDistributed[DistributionCategory.Liquidity] += tokenAmount;
    }
    
    /**
     * @notice Provide liquidity to DEX
     */
    function provideLiquidity(uint256 provisionIndex) external onlyGovernance {
        LiquidityProvision storage provision = liquidityProvisions[provisionIndex];
        require(!provision.provided, "Already provided");
        require(block.timestamp >= provision.unlockTime, "Not unlocked");
        
        provision.provided = true;
        
        // Transfer tokens to DEX
        require(qtmToken.transfer(provision.dexAddress, provision.tokenAmount), "Token transfer failed");
        
        // Send ETH to DEX
        if (provision.ethAmount > 0) {
            payable(provision.dexAddress).transfer(provision.ethAmount);
        }
        
        emit LiquidityProvided(provision.dexAddress, provision.tokenAmount, provision.ethAmount);
    }
    
    // ============ Ecosystem & Staking Rewards ============
    
    /**
     * @notice Allocate tokens to ecosystem fund
     */
    function allocateEcosystemFunds(address recipient, uint256 amount) external onlyGovernance {
        require(
            categoryDistributed[DistributionCategory.Ecosystem] + amount 
            <= categoryAllocations[DistributionCategory.Ecosystem],
            "Exceeds allocation"
        );
        
        categoryDistributed[DistributionCategory.Ecosystem] += amount;
        require(qtmToken.transfer(recipient, amount), "Transfer failed");
        
        emit TokensAllocated(DistributionCategory.Ecosystem, amount);
    }
    
    /**
     * @notice Allocate tokens for staking rewards
     */
    function allocateStakingRewards(address stakingContract, uint256 amount) external onlyGovernance {
        require(
            categoryDistributed[DistributionCategory.StakingRewards] + amount 
            <= categoryAllocations[DistributionCategory.StakingRewards],
            "Exceeds allocation"
        );
        
        categoryDistributed[DistributionCategory.StakingRewards] += amount;
        require(qtmToken.transfer(stakingContract, amount), "Transfer failed");
        
        emit TokensAllocated(DistributionCategory.StakingRewards, amount);
    }
    
    // ============ View Functions ============
    
    function getCategoryAllocation(DistributionCategory category) external view returns (uint256 allocated, uint256 distributed) {
        return (categoryAllocations[category], categoryDistributed[category]);
    }
    
    function getVestingSchedule(address beneficiary) external view returns (VestingSchedule memory) {
        return vestingSchedules[beneficiary];
    }
    
    function getReleasableAmount(address beneficiary) external view returns (uint256) {
        if (!hasVesting[beneficiary]) return 0;
        return _computeReleasableAmount(vestingSchedules[beneficiary]);
    }
    
    function hasClaimedAirdrop(uint256 campaignId, address recipient) external view returns (bool) {
        return airdropCampaigns[campaignId].claimed[recipient];
    }
    
    // ============ Admin Functions ============
    
    function updateWhitelistRequirement(bool required) external onlyGovernance {
        whitelistRequired = required;
    }
    
    function addToWhitelist(address[] calldata addresses) external onlyGovernance {
        for (uint i = 0; i < addresses.length; i++) {
            publicSaleWhitelist[addresses[i]] = true;
        }
    }
    
    function removeFromWhitelist(address[] calldata addresses) external onlyGovernance {
        for (uint i = 0; i < addresses.length; i++) {
            publicSaleWhitelist[addresses[i]] = false;
        }
    }
    
    function revokeVesting(address beneficiary) external onlyGovernance {
        VestingSchedule storage schedule = vestingSchedules[beneficiary];
        require(schedule.revocable, "Not revocable");
        require(!schedule.revoked, "Already revoked");
        
        uint256 releasable = _computeReleasableAmount(schedule);
        if (releasable > 0) {
            schedule.released += releasable;
            require(qtmToken.transfer(beneficiary, releasable), "Transfer failed");
        }
        
        schedule.revoked = true;
        
        // Return unvested tokens to treasury
        uint256 unvested = schedule.totalAmount - schedule.released;
        if (unvested > 0) {
            require(qtmToken.transfer(governance, unvested), "Transfer failed");
        }
    }
    
    function updateGovernance(address newGovernance) external onlyGovernance {
        governance = newGovernance;
    }
    
    // Receive ETH for liquidity provision
    receive() external payable {}
}