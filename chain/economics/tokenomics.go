package economics

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"quantum-blockchain/chain/types"
)

// TokenomicsEngine manages the economic model of the quantum blockchain
type TokenomicsEngine struct {
	// Core tokenomics parameters
	totalSupply          *big.Int // 1B QTM total supply
	circulatingSupply    *big.Int // Current circulating supply
	maxInflationRate     float64  // Maximum 5% annual inflation
	currentInflationRate float64  // Current inflation rate

	// Staking economics
	minStake          *big.Int      // Minimum stake for validation (100K QTM)
	maxStake          *big.Int      // Maximum stake per validator (10M QTM)
	stakingRewardRate float64       // Annual staking reward rate (8-12%)
	slashingRate      float64       // Slashing penalty rate (5%)
	unbondingPeriod   time.Duration // 21 days unbonding period

	// Block rewards and fees
	baseBlockReward     *big.Int // 1 QTM per block initially
	rewardDecayRate     float64  // 5% reduction per year
	burnRate            float64  // Fee burn rate (30% of tx fees)
	validatorCommission float64  // Default validator commission (5%)

	// Treasury and governance
	treasuryAllocation   float64  // 10% of block rewards to treasury
	governanceThreshold  *big.Int // Minimum QTM to create proposal (10K QTM)
	votingPowerThreshold *big.Int // Minimum QTM for voting (100 QTM)

	// Economic incentives
	delegatorRewardShare float64   // Delegators share of rewards (95%)
	validatorRewardShare float64   // Validators share of rewards (5%)
	earlyAdopterBonus    float64   // Early adopter staking bonus (20%)
	networkBootstrapEnd  time.Time // End of bootstrap period

	// Dynamic parameters
	networkUtilization float64       // Current network utilization (0-1)
	avgBlockTime       time.Duration // Average block time
	transactionVolume  uint64        // Recent transaction volume

	// Fee structure
	baseFee           *big.Int // Base transaction fee (10 Gwei)
	priorityFee       *big.Int // Priority fee for fast confirmation
	quantumVerifyFee  *big.Int // Additional fee for quantum signature verification
	contractDeployFee *big.Int // Contract deployment fee

	// State tracking
	totalStaked     *big.Int // Total QTM currently staked
	totalDelegated  *big.Int // Total QTM delegated
	totalBurned     *big.Int // Total QTM burned
	treasuryBalance *big.Int // Treasury balance

	// Reward distribution
	pendingRewards          map[types.Address]*big.Int                   // Pending validator rewards
	pendingDelegatorRewards map[types.Address]map[types.Address]*big.Int // delegator -> validator -> rewards

	// Economic metrics
	metrics *EconomicMetrics

	// Thread safety
	mu sync.RWMutex

	// Event handlers
	onRewardDistribution func(types.Address, *big.Int, RewardType)
	onSlashing           func(types.Address, *big.Int, SlashingReason)
	onTreasuryUpdate     func(*big.Int, TreasuryOperation)
}

// EconomicMetrics tracks key economic indicators
type EconomicMetrics struct {
	StakingRatio  float64   `json:"stakingRatio"`  // % of supply staked
	AverageStake  *big.Int  `json:"averageStake"`  // Average stake per validator
	RewardYield   float64   `json:"rewardYield"`   // Annualized reward yield
	NetworkValue  *big.Int  `json:"networkValue"`  // Total network value (TVL)
	InflationRate float64   `json:"inflationRate"` // Current inflation rate
	BurnRate      float64   `json:"burnRate"`      // Current burn rate
	TreasuryRatio float64   `json:"treasuryRatio"` // Treasury as % of supply
	NetworkHealth float64   `json:"networkHealth"` // Overall network health score
	LastUpdate    time.Time `json:"lastUpdate"`    // Last metrics update
}

// RewardType represents different types of rewards
type RewardType uint8

const (
	RewardTypeBlock RewardType = iota
	RewardTypeStaking
	RewardTypeDelegation
	RewardTypeGovernance
	RewardTypeEarlyAdopter
)

// SlashingReason represents reasons for slashing
type SlashingReason uint8

const (
	SlashingDoubleSign SlashingReason = iota
	SlashingDowntime
	SlashingInvalidBlock
	SlashingMisbehavior
)

// TreasuryOperation represents treasury operations
type TreasuryOperation uint8

const (
	TreasuryDeposit TreasuryOperation = iota
	TreasuryWithdraw
	TreasuryBurn
)

// NewTokenomicsEngine creates a new tokenomics engine
func NewTokenomicsEngine() *TokenomicsEngine {
	totalSupply := new(big.Int)
	totalSupply.SetString("1000000000000000000000000000", 10) // 1B QTM with 18 decimals

	minStake := new(big.Int)
	minStake.SetString("100000000000000000000000", 10) // 100K QTM

	maxStake := new(big.Int)
	maxStake.SetString("10000000000000000000000000", 10) // 10M QTM

	baseBlockReward := new(big.Int)
	baseBlockReward.SetString("1000000000000000000", 10) // 1 QTM

	governanceThreshold := new(big.Int)
	governanceThreshold.SetString("10000000000000000000000", 10) // 10K QTM

	votingPowerThreshold := new(big.Int)
	votingPowerThreshold.SetString("100000000000000000000", 10) // 100 QTM

	baseFee := new(big.Int)
	baseFee.SetString("10000000000", 10) // 10 Gwei

	priorityFee := new(big.Int)
	priorityFee.SetString("2000000000", 10) // 2 Gwei

	quantumVerifyFee := new(big.Int)
	quantumVerifyFee.SetString("5000000000", 10) // 5 Gwei

	contractDeployFee := new(big.Int)
	contractDeployFee.SetString("100000000000000000", 10) // 0.1 QTM

	return &TokenomicsEngine{
		totalSupply:          totalSupply,
		circulatingSupply:    new(big.Int).Set(totalSupply),
		maxInflationRate:     0.05, // 5%
		currentInflationRate: 0.03, // 3% initially

		minStake:          minStake,
		maxStake:          maxStake,
		stakingRewardRate: 0.10,                // 10% APY
		slashingRate:      0.05,                // 5% slashing
		unbondingPeriod:   21 * 24 * time.Hour, // 21 days

		baseBlockReward:     baseBlockReward,
		rewardDecayRate:     0.05, // 5% annual reduction
		burnRate:            0.30, // 30% of fees burned
		validatorCommission: 0.05, // 5% commission

		treasuryAllocation:   0.10, // 10% to treasury
		governanceThreshold:  governanceThreshold,
		votingPowerThreshold: votingPowerThreshold,

		delegatorRewardShare: 0.95,                                 // 95% to delegators
		validatorRewardShare: 0.05,                                 // 5% to validators
		earlyAdopterBonus:    0.20,                                 // 20% bonus
		networkBootstrapEnd:  time.Now().Add(365 * 24 * time.Hour), // 1 year bootstrap

		avgBlockTime: 2 * time.Second,

		baseFee:           baseFee,
		priorityFee:       priorityFee,
		quantumVerifyFee:  quantumVerifyFee,
		contractDeployFee: contractDeployFee,

		totalStaked:     big.NewInt(0),
		totalDelegated:  big.NewInt(0),
		totalBurned:     big.NewInt(0),
		treasuryBalance: big.NewInt(0),

		pendingRewards:          make(map[types.Address]*big.Int),
		pendingDelegatorRewards: make(map[types.Address]map[types.Address]*big.Int),

		metrics: &EconomicMetrics{
			LastUpdate: time.Now(),
		},
	}
}

// CalculateBlockReward calculates the block reward for current epoch
func (te *TokenomicsEngine) CalculateBlockReward(blockNumber uint64, validator types.Address) *big.Int {
	te.mu.RLock()
	defer te.mu.RUnlock()

	// Calculate years since genesis (assuming 2-second blocks)
	blocksPerYear := uint64((365 * 24 * 3600) / 2) // ~15.768M blocks per year
	years := float64(blockNumber) / float64(blocksPerYear)

	// Apply decay rate
	decayMultiplier := 1.0
	for i := 0; i < int(years); i++ {
		decayMultiplier *= (1.0 - te.rewardDecayRate)
	}

	baseReward := new(big.Int).Set(te.baseBlockReward)

	// Apply decay
	decayFloat := big.NewFloat(decayMultiplier)
	baseRewardFloat := new(big.Float).SetInt(baseReward)
	finalRewardFloat := new(big.Float).Mul(baseRewardFloat, decayFloat)

	finalReward, _ := finalRewardFloat.Int(nil)

	// Apply early adopter bonus if still in bootstrap period
	if time.Now().Before(te.networkBootstrapEnd) {
		bonusFloat := big.NewFloat(1.0 + te.earlyAdopterBonus)
		rewardFloat := new(big.Float).SetInt(finalReward)
		finalRewardFloat := new(big.Float).Mul(rewardFloat, bonusFloat)
		finalReward, _ = finalRewardFloat.Int(nil)
	}

	return finalReward
}

// CalculateStakingReward calculates staking rewards for a validator
func (te *TokenomicsEngine) CalculateStakingReward(
	validatorAddr types.Address,
	stakedAmount *big.Int,
	delegatedAmount *big.Int,
	commission float64,
	uptime float64,
	performance float64,
) (*big.Int, *big.Int, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()

	if stakedAmount.Sign() <= 0 {
		return big.NewInt(0), big.NewInt(0), errors.New("staked amount must be positive")
	}

	// Calculate total stake
	totalStake := new(big.Int).Add(stakedAmount, delegatedAmount)

	// Calculate annual reward
	rewardRate := te.stakingRewardRate

	// Apply performance multiplier
	rewardRate *= performance

	// Apply uptime multiplier
	rewardRate *= uptime

	// Calculate daily reward (assuming 365 days per year)
	dailyRewardRate := rewardRate / 365.0

	// Calculate total daily reward
	totalRewardFloat := new(big.Float).SetInt(totalStake)
	dailyRateFloat := big.NewFloat(dailyRewardRate)
	totalDailyRewardFloat := new(big.Float).Mul(totalRewardFloat, dailyRateFloat)

	totalDailyReward, _ := totalDailyRewardFloat.Int(nil)

	// Calculate validator commission
	commissionFloat := big.NewFloat(commission)
	totalRewardForCommissionFloat := new(big.Float).SetInt(totalDailyReward)
	validatorCommissionFloat := new(big.Float).Mul(totalRewardForCommissionFloat, commissionFloat)
	validatorCommission, _ := validatorCommissionFloat.Int(nil)

	// Remaining reward goes to delegators
	delegatorReward := new(big.Int).Sub(totalDailyReward, validatorCommission)

	// Validator also gets reward for their own stake
	if stakedAmount.Sign() > 0 {
		validatorStakeRewardFloat := new(big.Float).SetInt(delegatorReward)
		stakeRatioFloat := new(big.Float).Quo(new(big.Float).SetInt(stakedAmount), new(big.Float).SetInt(totalStake))
		validatorStakeRewardFloat.Mul(validatorStakeRewardFloat, stakeRatioFloat)
		validatorStakeReward, _ := validatorStakeRewardFloat.Int(nil)

		validatorCommission.Add(validatorCommission, validatorStakeReward)
		delegatorReward.Sub(delegatorReward, validatorStakeReward)
	}

	return validatorCommission, delegatorReward, nil
}

// CalculateTransactionFee calculates the transaction fee based on current network conditions
func (te *TokenomicsEngine) CalculateTransactionFee(
	txType TransactionType,
	gasUsed uint64,
	priority PriorityLevel,
) *big.Int {
	te.mu.RLock()
	defer te.mu.RUnlock()

	baseFee := new(big.Int).Set(te.baseFee)

	// Apply network utilization multiplier
	utilizationMultiplier := 1.0 + (te.networkUtilization * 2.0) // Up to 3x during high usage
	utilizationBig := big.NewFloat(utilizationMultiplier)
	baseFeeFloat := new(big.Float).SetInt(baseFee)
	adjustedBaseFeeFloat := new(big.Float).Mul(baseFeeFloat, utilizationBig)
	adjustedBaseFee, _ := adjustedBaseFeeFloat.Int(nil)

	// Calculate gas fee
	gasFee := new(big.Int).Mul(adjustedBaseFee, big.NewInt(int64(gasUsed)))

	// Add priority fee if requested
	if priority == PriorityHigh {
		priorityTotal := new(big.Int).Mul(te.priorityFee, big.NewInt(int64(gasUsed)))
		gasFee.Add(gasFee, priorityTotal)
	}

	// Add transaction type specific fees
	switch txType {
	case TxTypeQuantumSignature:
		quantumFee := new(big.Int).Mul(te.quantumVerifyFee, big.NewInt(int64(gasUsed)))
		gasFee.Add(gasFee, quantumFee)
	case TxTypeContractDeploy:
		gasFee.Add(gasFee, te.contractDeployFee)
	}

	return gasFee
}

// ProcessFeeBurn burns a portion of transaction fees
func (te *TokenomicsEngine) ProcessFeeBurn(totalFees *big.Int) (*big.Int, *big.Int) {
	te.mu.Lock()
	defer te.mu.Unlock()

	// Calculate burn amount
	burnRateFloat := big.NewFloat(te.burnRate)
	totalFeesFloat := new(big.Float).SetInt(totalFees)
	burnAmountFloat := new(big.Float).Mul(totalFeesFloat, burnRateFloat)
	burnAmount, _ := burnAmountFloat.Int(nil)

	// Remaining fees go to treasury
	treasuryAmount := new(big.Int).Sub(totalFees, burnAmount)

	// Update state
	te.totalBurned.Add(te.totalBurned, burnAmount)
	te.treasuryBalance.Add(te.treasuryBalance, treasuryAmount)
	te.circulatingSupply.Sub(te.circulatingSupply, burnAmount)

	return burnAmount, treasuryAmount
}

// CalculateSlashingPenalty calculates slashing penalty amount
func (te *TokenomicsEngine) CalculateSlashingPenalty(
	stakedAmount *big.Int,
	reason SlashingReason,
	severity SlashingSeverity,
) *big.Int {
	te.mu.RLock()
	defer te.mu.RUnlock()

	baseRate := te.slashingRate

	// Adjust rate based on reason
	switch reason {
	case SlashingDoubleSign:
		baseRate = 0.20 // 20% for double signing (severe)
	case SlashingDowntime:
		baseRate = 0.01 // 1% for downtime
	case SlashingInvalidBlock:
		baseRate = 0.10 // 10% for invalid blocks
	case SlashingMisbehavior:
		baseRate = 0.05 // 5% for general misbehavior
	}

	// Adjust for severity
	switch severity {
	case SeverityLow:
		baseRate *= 0.5
	case SeverityHigh:
		baseRate *= 1.5
	case SeverityCritical:
		baseRate *= 2.0
	}

	// Calculate penalty
	penaltyRateFloat := big.NewFloat(baseRate)
	stakedFloat := new(big.Float).SetInt(stakedAmount)
	penaltyFloat := new(big.Float).Mul(stakedFloat, penaltyRateFloat)
	penalty, _ := penaltyFloat.Int(nil)

	return penalty
}

// UpdateNetworkMetrics updates economic metrics
func (te *TokenomicsEngine) UpdateNetworkMetrics(
	networkUtilization float64,
	transactionVolume uint64,
	avgBlockTime time.Duration,
) {
	te.mu.Lock()
	defer te.mu.Unlock()

	te.networkUtilization = networkUtilization
	te.transactionVolume = transactionVolume
	te.avgBlockTime = avgBlockTime

	// Update metrics
	if te.totalSupply.Sign() > 0 {
		stakingRatioFloat := new(big.Float).Quo(
			new(big.Float).SetInt(te.totalStaked),
			new(big.Float).SetInt(te.totalSupply),
		)
		te.metrics.StakingRatio, _ = stakingRatioFloat.Float64()

		treasuryRatioFloat := new(big.Float).Quo(
			new(big.Float).SetInt(te.treasuryBalance),
			new(big.Float).SetInt(te.totalSupply),
		)
		te.metrics.TreasuryRatio, _ = treasuryRatioFloat.Float64()
	}

	te.metrics.InflationRate = te.currentInflationRate
	te.metrics.BurnRate = te.burnRate
	te.metrics.RewardYield = te.stakingRewardRate
	te.metrics.LastUpdate = time.Now()

	// Calculate network health score (0-1)
	healthScore := 1.0
	if networkUtilization > 0.8 {
		healthScore *= 0.8 // Reduce for high utilization
	}
	if te.metrics.StakingRatio < 0.3 {
		healthScore *= 0.9 // Reduce for low staking ratio
	}
	te.metrics.NetworkHealth = healthScore
}

// GetEconomicMetrics returns current economic metrics
func (te *TokenomicsEngine) GetEconomicMetrics() *EconomicMetrics {
	te.mu.RLock()
	defer te.mu.RUnlock()

	return te.metrics
}

// GetTokenomicsInfo returns comprehensive tokenomics information
func (te *TokenomicsEngine) GetTokenomicsInfo() map[string]interface{} {
	te.mu.RLock()
	defer te.mu.RUnlock()

	return map[string]interface{}{
		"totalSupply":          te.totalSupply.String(),
		"circulatingSupply":    te.circulatingSupply.String(),
		"totalStaked":          te.totalStaked.String(),
		"totalBurned":          te.totalBurned.String(),
		"treasuryBalance":      te.treasuryBalance.String(),
		"currentInflationRate": te.currentInflationRate,
		"stakingRewardRate":    te.stakingRewardRate,
		"minStake":             te.minStake.String(),
		"maxStake":             te.maxStake.String(),
		"baseBlockReward":      te.baseBlockReward.String(),
		"baseFee":              te.baseFee.String(),
		"burnRate":             te.burnRate,
		"metrics":              te.metrics,
	}
}

// Supporting types and enums
type TransactionType uint8

const (
	TxTypeStandard TransactionType = iota
	TxTypeQuantumSignature
	TxTypeContractDeploy
	TxTypeGovernance
	TxTypeStaking
)

type PriorityLevel uint8

const (
	PriorityLow PriorityLevel = iota
	PriorityNormal
	PriorityHigh
)

type SlashingSeverity uint8

const (
	SeverityLow SlashingSeverity = iota
	SeverityNormal
	SeverityHigh
	SeverityCritical
)

// SetEventHandlers sets economic event handlers
func (te *TokenomicsEngine) SetEventHandlers(
	onRewardDistribution func(types.Address, *big.Int, RewardType),
	onSlashing func(types.Address, *big.Int, SlashingReason),
	onTreasuryUpdate func(*big.Int, TreasuryOperation),
) {
	te.onRewardDistribution = onRewardDistribution
	te.onSlashing = onSlashing
	te.onTreasuryUpdate = onTreasuryUpdate
}
