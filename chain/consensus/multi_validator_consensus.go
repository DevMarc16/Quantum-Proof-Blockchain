package consensus

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// MultiValidatorConsensus implements production-ready consensus with 3-21 validators
type MultiValidatorConsensus struct {
	chainID             *big.Int
	validators          map[types.Address]*ValidatorState
	validatorList       []*ValidatorState
	delegations         map[types.Address]map[types.Address]*big.Int // delegator -> validator -> amount
	currentEpoch        uint64
	epochBlocks         uint64
	blockTime           time.Duration
	minValidators       int
	maxValidators       int
	minStake            *big.Int
	slashingPercentage  float64
	
	// Consensus state
	currentProposer     types.Address
	consensusMessages   map[uint64]map[types.Address]*ConsensusVote
	finalizationQuorum  float64 // 2/3+ required
	
	// Security and governance
	proposalTimeout     time.Duration
	maxMissedBlocks     uint64
	jailDuration        time.Duration
	unbondingPeriod     time.Duration
	
	// Performance tracking
	networkPerformance  *NetworkPerformance
	
	// Thread safety
	mu                  sync.RWMutex
	
	// Event handlers
	onSlash            func(types.Address, string, *big.Int)
	onJail             func(types.Address, time.Duration)
	onUnbond           func(types.Address, types.Address, *big.Int)
}

// ValidatorState represents enhanced validator state for production
type ValidatorState struct {
	Address            types.Address                 `json:"address"`
	PublicKey          []byte                       `json:"publicKey"`
	PrivateKey         []byte                       `json:"privateKey,omitempty"`
	SigAlgorithm       crypto.SignatureAlgorithm    `json:"sigAlgorithm"`
	
	// Staking
	SelfStake          *big.Int                     `json:"selfStake"`
	DelegatedStake     *big.Int                     `json:"delegatedStake"`
	TotalStake         *big.Int                     `json:"totalStake"`
	
	// Performance metrics
	Performance        *ValidatorPerformance        `json:"performance"`
	
	// Status
	Status             ValidatorStatus              `json:"status"`
	JailedUntil        time.Time                   `json:"jailedUntil"`
	LastActive         time.Time                   `json:"lastActive"`
	
	// Governance
	VotingPower        *big.Int                     `json:"votingPower"`
	Commission         float64                      `json:"commission"` // 0.0 to 1.0
}

// ValidatorPerformance tracks validator performance metrics
type ValidatorPerformance struct {
	BlocksProposed     uint64    `json:"blocksProposed"`
	BlocksProposedOK   uint64    `json:"blocksProposedOK"`
	BlocksMissed       uint64    `json:"blocksMissed"`
	AttestationsMissed uint64    `json:"attestationsMissed"`
	SlashCount         uint64    `json:"slashCount"`
	UptimeScore        float64   `json:"uptimeScore"`        // 0.0 to 1.0
	LatencyScore       float64   `json:"latencyScore"`       // 0.0 to 1.0
	ReliabilityScore   float64   `json:"reliabilityScore"`   // 0.0 to 1.0
	LastSlash          time.Time `json:"lastSlash"`
}

// ValidatorStatus represents validator status
type ValidatorStatus uint8

const (
	StatusActive ValidatorStatus = iota
	StatusJailed
	StatusUnbonding
	StatusUnbonded
	StatusSlashed
)

// ConsensusVote represents a consensus vote with quantum signature
type ConsensusVote struct {
	Validator     types.Address `json:"validator"`
	BlockHash     types.Hash    `json:"blockHash"`
	BlockHeight   uint64        `json:"blockHeight"`
	VoteType      VoteType      `json:"voteType"`
	Timestamp     time.Time     `json:"timestamp"`
	Signature     []byte        `json:"signature"`
	PublicKey     []byte        `json:"publicKey"`
	SigAlgorithm  crypto.SignatureAlgorithm `json:"sigAlgorithm"`
}

// VoteType represents different vote types
type VoteType uint8

const (
	VoteProposal VoteType = iota
	VotePreCommit
	VoteCommit
	VoteFinalize
)

// NetworkPerformance tracks overall network performance
type NetworkPerformance struct {
	BlockTime          time.Duration `json:"blockTime"`
	TransactionTPS     float64       `json:"transactionTPS"`
	NetworkLatency     time.Duration `json:"networkLatency"`
	ConsensusLatency   time.Duration `json:"consensusLatency"`
	FinalizationTime   time.Duration `json:"finalizationTime"`
	ValidatorOnline    int           `json:"validatorOnline"`
	NetworkLoad        float64       `json:"networkLoad"`
	LastUpdate         time.Time     `json:"lastUpdate"`
}

// NewMultiValidatorConsensus creates a new multi-validator consensus instance
func NewMultiValidatorConsensus(chainID *big.Int) *MultiValidatorConsensus {
	minStake := new(big.Int)
	minStake.SetString("100000000000000000000000", 10) // 100,000 QTM minimum

	return &MultiValidatorConsensus{
		chainID:            chainID,
		validators:         make(map[types.Address]*ValidatorState),
		validatorList:      make([]*ValidatorState, 0),
		delegations:        make(map[types.Address]map[types.Address]*big.Int),
		currentEpoch:       0,
		epochBlocks:        7200, // ~4 hours at 2-second blocks
		blockTime:          2 * time.Second,
		minValidators:      3,
		maxValidators:      21,
		minStake:           minStake,
		slashingPercentage: 0.05, // 5% slashing
		consensusMessages:  make(map[uint64]map[types.Address]*ConsensusVote),
		finalizationQuorum: 0.67, // 2/3+ required
		proposalTimeout:    8 * time.Second, // 4x block time
		maxMissedBlocks:    50,   // Jail after 50 missed blocks
		jailDuration:       24 * time.Hour,
		unbondingPeriod:    21 * 24 * time.Hour, // 21 days
		networkPerformance: &NetworkPerformance{
			BlockTime:        2 * time.Second,
			LastUpdate:       time.Now(),
		},
	}
}

// RegisterValidator registers a new validator with enhanced validation
func (mvc *MultiValidatorConsensus) RegisterValidator(
	address types.Address,
	publicKey []byte,
	selfStake *big.Int,
	sigAlg crypto.SignatureAlgorithm,
	commission float64,
) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()

	// Validation checks
	if _, exists := mvc.validators[address]; exists {
		return errors.New("validator already exists")
	}
	
	if selfStake.Cmp(mvc.minStake) < 0 {
		return fmt.Errorf("insufficient self-stake: minimum %s QTM required", mvc.minStake.String())
	}
	
	if commission < 0 || commission > 1.0 {
		return errors.New("commission must be between 0% and 100%")
	}
	
	if len(mvc.validatorList) >= mvc.maxValidators {
		return fmt.Errorf("maximum validators (%d) reached", mvc.maxValidators)
	}

	// Create validator state
	validator := &ValidatorState{
		Address:        address,
		PublicKey:      publicKey,
		SigAlgorithm:   sigAlg,
		SelfStake:      new(big.Int).Set(selfStake),
		DelegatedStake: big.NewInt(0),
		TotalStake:     new(big.Int).Set(selfStake),
		Performance: &ValidatorPerformance{
			UptimeScore:      1.0,
			LatencyScore:     1.0,
			ReliabilityScore: 1.0,
		},
		Status:      StatusActive,
		LastActive:  time.Now(),
		VotingPower: new(big.Int).Set(selfStake),
		Commission:  commission,
	}

	mvc.validators[address] = validator
	mvc.updateValidatorSet()
	
	return nil
}

// Delegate allows token holders to delegate stake to validators
func (mvc *MultiValidatorConsensus) Delegate(
	delegator types.Address,
	validator types.Address,
	amount *big.Int,
) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()
	
	validatorState, exists := mvc.validators[validator]
	if !exists {
		return errors.New("validator not found")
	}
	
	if validatorState.Status != StatusActive {
		return errors.New("cannot delegate to inactive validator")
	}
	
	if amount.Sign() <= 0 {
		return errors.New("delegation amount must be positive")
	}
	
	// Initialize delegator map if needed
	if mvc.delegations[delegator] == nil {
		mvc.delegations[delegator] = make(map[types.Address]*big.Int)
	}
	
	// Add delegation
	if existing := mvc.delegations[delegator][validator]; existing != nil {
		existing.Add(existing, amount)
	} else {
		mvc.delegations[delegator][validator] = new(big.Int).Set(amount)
	}
	
	// Update validator stake
	validatorState.DelegatedStake.Add(validatorState.DelegatedStake, amount)
	validatorState.TotalStake.Add(validatorState.TotalStake, amount)
	validatorState.VotingPower.Add(validatorState.VotingPower, amount)
	
	mvc.updateValidatorSet()
	return nil
}

// Undelegate initiates undelegation process
func (mvc *MultiValidatorConsensus) Undelegate(
	delegator types.Address,
	validator types.Address,
	amount *big.Int,
) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()
	
	if mvc.delegations[delegator] == nil {
		return errors.New("no delegations found")
	}
	
	currentDelegation := mvc.delegations[delegator][validator]
	if currentDelegation == nil || currentDelegation.Cmp(amount) < 0 {
		return errors.New("insufficient delegation to undelegate")
	}
	
	validatorState := mvc.validators[validator]
	if validatorState == nil {
		return errors.New("validator not found")
	}
	
	// Update delegation
	currentDelegation.Sub(currentDelegation, amount)
	if currentDelegation.Sign() == 0 {
		delete(mvc.delegations[delegator], validator)
	}
	
	// Update validator stake
	validatorState.DelegatedStake.Sub(validatorState.DelegatedStake, amount)
	validatorState.TotalStake.Sub(validatorState.TotalStake, amount)
	validatorState.VotingPower.Sub(validatorState.VotingPower, amount)
	
	// Trigger unbonding callback
	if mvc.onUnbond != nil {
		mvc.onUnbond(delegator, validator, amount)
	}
	
	mvc.updateValidatorSet()
	return nil
}

// SlashValidator slashes a validator for misbehavior
func (mvc *MultiValidatorConsensus) SlashValidator(
	validator types.Address,
	reason string,
	evidence []byte,
) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()
	
	validatorState, exists := mvc.validators[validator]
	if !exists {
		return errors.New("validator not found")
	}
	
	// Calculate slash amount
	slashAmount := new(big.Int).Set(validatorState.TotalStake)
	slashAmount.Mul(slashAmount, big.NewInt(int64(mvc.slashingPercentage*1000)))
	slashAmount.Div(slashAmount, big.NewInt(1000))
	
	// Apply slashing
	validatorState.TotalStake.Sub(validatorState.TotalStake, slashAmount)
	validatorState.VotingPower.Sub(validatorState.VotingPower, slashAmount)
	validatorState.Performance.SlashCount++
	validatorState.Performance.LastSlash = time.Now()
	validatorState.Status = StatusSlashed
	
	// Jail validator
	validatorState.JailedUntil = time.Now().Add(mvc.jailDuration)
	
	// Trigger slash callback
	if mvc.onSlash != nil {
		mvc.onSlash(validator, reason, slashAmount)
	}
	
	mvc.updateValidatorSet()
	return nil
}

// GetNextProposer determines next block proposer with enhanced selection
func (mvc *MultiValidatorConsensus) GetNextProposer(blockHeight uint64) (types.Address, error) {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()
	
	activeValidators := mvc.getActiveValidators()
	if len(activeValidators) == 0 {
		return types.Address{}, errors.New("no active validators")
	}
	
	// Use verifiable random function (simplified)
	seed := mvc.generateSeed(blockHeight)
	
	// Calculate weighted selection
	totalWeight := big.NewInt(0)
	for _, validator := range activeValidators {
		// Apply performance weighting
		weight := new(big.Int).Set(validator.VotingPower)
		performanceMultiplier := int64(validator.Performance.ReliabilityScore * 1000)
		weight.Mul(weight, big.NewInt(performanceMultiplier))
		weight.Div(weight, big.NewInt(1000))
		totalWeight.Add(totalWeight, weight)
	}
	
	if totalWeight.Sign() == 0 {
		return activeValidators[0].Address, nil
	}
	
	// Select proposer
	randomValue := new(big.Int).Mod(seed, totalWeight)
	currentWeight := big.NewInt(0)
	
	for _, validator := range activeValidators {
		weight := new(big.Int).Set(validator.VotingPower)
		performanceMultiplier := int64(validator.Performance.ReliabilityScore * 1000)
		weight.Mul(weight, big.NewInt(performanceMultiplier))
		weight.Div(weight, big.NewInt(1000))
		
		currentWeight.Add(currentWeight, weight)
		if currentWeight.Cmp(randomValue) > 0 {
			return validator.Address, nil
		}
	}
	
	return activeValidators[0].Address, nil
}

// SubmitConsensusVote submits a consensus vote
func (mvc *MultiValidatorConsensus) SubmitConsensusVote(
	validator types.Address,
	blockHash types.Hash,
	blockHeight uint64,
	voteType VoteType,
	privateKey []byte,
) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()
	
	validatorState, exists := mvc.validators[validator]
	if !exists || validatorState.Status != StatusActive {
		return errors.New("validator not active")
	}
	
	// Create vote message
	voteData := fmt.Sprintf("%s:%d:%d:%d", 
		blockHash.Hex(), blockHeight, voteType, time.Now().Unix())
	
	// Sign vote
	qrSig, err := crypto.SignMessage([]byte(voteData), validatorState.SigAlgorithm, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign vote: %w", err)
	}
	
	vote := &ConsensusVote{
		Validator:    validator,
		BlockHash:    blockHash,
		BlockHeight:  blockHeight,
		VoteType:     voteType,
		Timestamp:    time.Now(),
		Signature:    qrSig.Signature,
		PublicKey:    validatorState.PublicKey,
		SigAlgorithm: validatorState.SigAlgorithm,
	}
	
	// Store vote
	if mvc.consensusMessages[blockHeight] == nil {
		mvc.consensusMessages[blockHeight] = make(map[types.Address]*ConsensusVote)
	}
	mvc.consensusMessages[blockHeight][validator] = vote
	
	return nil
}

// CheckConsensus checks if consensus is reached for a block
func (mvc *MultiValidatorConsensus) CheckConsensus(blockHeight uint64) (bool, error) {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()
	
	votes, exists := mvc.consensusMessages[blockHeight]
	if !exists {
		return false, nil
	}
	
	activeValidators := mvc.getActiveValidators()
	totalVotingPower := big.NewInt(0)
	for _, validator := range activeValidators {
		totalVotingPower.Add(totalVotingPower, validator.VotingPower)
	}
	
	// Count voting power for this block
	votingPower := big.NewInt(0)
	for validatorAddr, vote := range votes {
		if validator := mvc.validators[validatorAddr]; validator != nil && validator.Status == StatusActive {
			// Verify vote signature
			voteData := fmt.Sprintf("%s:%d:%d:%d", 
				vote.BlockHash.Hex(), vote.BlockHeight, vote.VoteType, vote.Timestamp.Unix())
			
			qrSig := &crypto.QRSignature{
				Algorithm: vote.SigAlgorithm,
				Signature: vote.Signature,
				PublicKey: vote.PublicKey,
			}
			
			if valid, _ := crypto.VerifySignature([]byte(voteData), qrSig); valid {
				votingPower.Add(votingPower, validator.VotingPower)
			}
		}
	}
	
	// Check if we have 2/3+ voting power
	requiredPower := new(big.Int).Set(totalVotingPower)
	requiredPower.Mul(requiredPower, big.NewInt(67)) // 67% for safety
	requiredPower.Div(requiredPower, big.NewInt(100))
	
	return votingPower.Cmp(requiredPower) >= 0, nil
}

// updateValidatorSet updates the active validator set
func (mvc *MultiValidatorConsensus) updateValidatorSet() {
	mvc.validatorList = make([]*ValidatorState, 0)
	
	for _, validator := range mvc.validators {
		if validator.Status == StatusActive && 
		   validator.TotalStake.Cmp(mvc.minStake) >= 0 &&
		   time.Now().After(validator.JailedUntil) {
			mvc.validatorList = append(mvc.validatorList, validator)
		}
	}
	
	// Sort by total stake (descending)
	sort.Slice(mvc.validatorList, func(i, j int) bool {
		return mvc.validatorList[i].TotalStake.Cmp(mvc.validatorList[j].TotalStake) > 0
	})
	
	// Limit to max validators
	if len(mvc.validatorList) > mvc.maxValidators {
		mvc.validatorList = mvc.validatorList[:mvc.maxValidators]
	}
}

// getActiveValidators returns active validators (thread-safe helper)
func (mvc *MultiValidatorConsensus) getActiveValidators() []*ValidatorState {
	activeValidators := make([]*ValidatorState, 0)
	for _, validator := range mvc.validatorList {
		if validator.Status == StatusActive && time.Now().After(validator.JailedUntil) {
			activeValidators = append(activeValidators, validator)
		}
	}
	return activeValidators
}

// generateSeed generates a deterministic seed for proposer selection
func (mvc *MultiValidatorConsensus) generateSeed(blockHeight uint64) *big.Int {
	data := fmt.Sprintf("%d:%d:%d", mvc.chainID.Uint64(), blockHeight, mvc.currentEpoch)
	hash := sha256.Sum256([]byte(data))
	return new(big.Int).SetBytes(hash[:])
}

// GetValidatorSet returns the current active validator set
func (mvc *MultiValidatorConsensus) GetValidatorSet() []*ValidatorState {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()
	
	result := make([]*ValidatorState, len(mvc.validatorList))
	copy(result, mvc.validatorList)
	return result
}

// GetNetworkPerformance returns network performance metrics
func (mvc *MultiValidatorConsensus) GetNetworkPerformance() *NetworkPerformance {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()
	
	// Update current metrics
	mvc.networkPerformance.ValidatorOnline = len(mvc.getActiveValidators())
	mvc.networkPerformance.LastUpdate = time.Now()
	
	return mvc.networkPerformance
}

// GetConsensusInfo returns comprehensive consensus information
func (mvc *MultiValidatorConsensus) GetConsensusInfo() map[string]interface{} {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()
	
	return map[string]interface{}{
		"chainID":             mvc.chainID.String(),
		"currentEpoch":        mvc.currentEpoch,
		"epochBlocks":         mvc.epochBlocks,
		"blockTime":           mvc.blockTime.Seconds(),
		"activeValidators":    len(mvc.validatorList),
		"totalValidators":     len(mvc.validators),
		"minValidators":       mvc.minValidators,
		"maxValidators":       mvc.maxValidators,
		"minStake":           mvc.minStake.String(),
		"slashingPercentage":  mvc.slashingPercentage,
		"finalizationQuorum":  mvc.finalizationQuorum,
		"unbondingPeriod":     mvc.unbondingPeriod.Hours(),
		"networkPerformance":  mvc.networkPerformance,
	}
}

// SetEventHandlers sets event handlers for slashing, jailing, etc.
func (mvc *MultiValidatorConsensus) SetEventHandlers(
	onSlash func(types.Address, string, *big.Int),
	onJail func(types.Address, time.Duration),
	onUnbond func(types.Address, types.Address, *big.Int),
) {
	mvc.onSlash = onSlash
	mvc.onJail = onJail
	mvc.onUnbond = onUnbond
}