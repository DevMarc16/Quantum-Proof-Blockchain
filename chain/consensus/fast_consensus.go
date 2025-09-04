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

// FastConsensus implements a high-throughput consensus mechanism similar to Flare
// Features:
// - 2-second block times
// - Quantum-resistant validator signatures
// - Fast finality through aggregate signatures
// - Dynamic validator set based on stake
type FastConsensus struct {
	chainID       *big.Int
	validators    map[types.Address]*Validator
	validatorList []*Validator
	currentEpoch  uint64
	blockTime     time.Duration
	mu            sync.RWMutex
	tokenSupply   *types.TokenSupply
	minStake      *big.Int
	maxValidators int
}

// Validator represents a consensus validator
type Validator struct {
	Address        types.Address             `json:"address"`
	PublicKey      []byte                    `json:"publicKey"`
	PrivateKey     []byte                    `json:"privateKey,omitempty"` // Only for local validator
	Stake          *big.Int                  `json:"stake"`
	Performance    float64                   `json:"performance"` // Performance score (0.0 to 1.0)
	LastActive     time.Time                 `json:"lastActive"`
	SigAlgorithm   crypto.SignatureAlgorithm `json:"sigAlgorithm"`
	IsActive       bool                      `json:"isActive"`
	BlocksProduced uint64                    `json:"blocksProduced"`
	BlocksMissed   uint64                    `json:"blocksMissed"`
}

// ConsensusMessage represents a consensus message between validators
type ConsensusMessage struct {
	Type        MessageType   `json:"type"`
	Epoch       uint64        `json:"epoch"`
	BlockHeight uint64        `json:"blockHeight"`
	BlockHash   types.Hash    `json:"blockHash"`
	Validator   types.Address `json:"validator"`
	Signature   []byte        `json:"signature"`
	Timestamp   time.Time     `json:"timestamp"`
	Data        []byte        `json:"data,omitempty"`
}

type MessageType uint8

const (
	MessageTypePropose MessageType = iota
	MessageTypeVote
	MessageTypeCommit
	MessageTypeFinalize
)

// NewFastConsensus creates a new fast consensus instance
func NewFastConsensus(chainID *big.Int, tokenSupply *types.TokenSupply) *FastConsensus {
	minStake := new(big.Int)
	minStake.SetString("100000000000000000000000", 10) // 100,000 QTM minimum stake

	return &FastConsensus{
		chainID:       chainID,
		validators:    make(map[types.Address]*Validator),
		validatorList: make([]*Validator, 0),
		currentEpoch:  0,
		blockTime:     2 * time.Second, // Fast 2-second blocks like Flare
		tokenSupply:   tokenSupply,
		minStake:      minStake,
		maxValidators: 21, // Optimal number for fast consensus
	}
}

// RegisterValidator registers a new validator
func (fc *FastConsensus) RegisterValidator(address types.Address, publicKey []byte, stake *big.Int, sigAlg crypto.SignatureAlgorithm) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if stake.Cmp(fc.minStake) < 0 {
		return fmt.Errorf("insufficient stake: minimum %s QTM required", fc.minStake.String())
	}

	validator := &Validator{
		Address:        address,
		PublicKey:      publicKey,
		Stake:          new(big.Int).Set(stake),
		Performance:    1.0, // Start with perfect performance
		LastActive:     time.Now(),
		SigAlgorithm:   sigAlg,
		IsActive:       true,
		BlocksProduced: 0,
		BlocksMissed:   0,
	}

	fc.validators[address] = validator
	fc.updateValidatorList()

	return nil
}

// UpdateValidatorStake updates a validator's stake
func (fc *FastConsensus) UpdateValidatorStake(address types.Address, newStake *big.Int) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	validator, exists := fc.validators[address]
	if !exists {
		return errors.New("validator not found")
	}

	if newStake.Cmp(fc.minStake) < 0 {
		// Remove validator if stake is too low
		delete(fc.validators, address)
		fc.updateValidatorList()
		return nil
	}

	validator.Stake.Set(newStake)
	fc.updateValidatorList()

	return nil
}

// updateValidatorList updates the sorted validator list based on stake
func (fc *FastConsensus) updateValidatorList() {
	fc.validatorList = make([]*Validator, 0, len(fc.validators))

	for _, validator := range fc.validators {
		if validator.IsActive && validator.Stake.Cmp(fc.minStake) >= 0 {
			fc.validatorList = append(fc.validatorList, validator)
		}
	}

	// Sort by stake (highest first)
	sort.Slice(fc.validatorList, func(i, j int) bool {
		return fc.validatorList[i].Stake.Cmp(fc.validatorList[j].Stake) > 0
	})

	// Limit to max validators for optimal performance
	if len(fc.validatorList) > fc.maxValidators {
		fc.validatorList = fc.validatorList[:fc.maxValidators]
	}
}

// GetActiveValidators returns the current active validator set
func (fc *FastConsensus) GetActiveValidators() []*Validator {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	result := make([]*Validator, len(fc.validatorList))
	copy(result, fc.validatorList)
	return result
}

// ProposeBlock proposes a new block for consensus
func (fc *FastConsensus) ProposeBlock(block *types.Block, proposer types.Address) (*ConsensusMessage, error) {
	fc.mu.RLock()
	validator, exists := fc.validators[proposer]
	fc.mu.RUnlock()

	if !exists || !validator.IsActive {
		return nil, errors.New("invalid proposer")
	}

	// Create proposal message
	message := &ConsensusMessage{
		Type:        MessageTypePropose,
		Epoch:       fc.currentEpoch,
		BlockHeight: block.Number().Uint64(),
		BlockHash:   block.Hash(),
		Validator:   proposer,
		Timestamp:   time.Now(),
		Data:        []byte{}, // Simplified for now
	}

	// Sign the message with quantum-resistant signature
	messageHash := fc.hashMessage(message)
	qrSig, err := crypto.SignMessage(messageHash, validator.SigAlgorithm, validator.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign proposal: %w", err)
	}

	message.Signature = qrSig.Signature

	return message, nil
}

// VoteOnBlock votes on a proposed block
func (fc *FastConsensus) VoteOnBlock(blockHash types.Hash, voter types.Address, approve bool) (*ConsensusMessage, error) {
	fc.mu.RLock()
	validator, exists := fc.validators[voter]
	fc.mu.RUnlock()

	if !exists || !validator.IsActive {
		return nil, errors.New("invalid voter")
	}

	voteData := []byte{0}
	if approve {
		voteData = []byte{1}
	}

	message := &ConsensusMessage{
		Type:      MessageTypeVote,
		Epoch:     fc.currentEpoch,
		BlockHash: blockHash,
		Validator: voter,
		Timestamp: time.Now(),
		Data:      voteData,
	}

	// Sign the vote
	messageHash := fc.hashMessage(message)
	qrSig, err := crypto.SignMessage(messageHash, validator.SigAlgorithm, validator.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign vote: %w", err)
	}

	message.Signature = qrSig.Signature

	return message, nil
}

// AggregateVotes aggregates multiple votes into a single signature for efficiency
func (fc *FastConsensus) AggregateVotes(votes []*ConsensusMessage) (*crypto.AggregatedSignature, error) {
	if len(votes) == 0 {
		return nil, errors.New("no votes to aggregate")
	}

	signatures := make([]*crypto.QRSignature, len(votes))
	messageHashes := make([][]byte, len(votes))

	for i, vote := range votes {
		fc.mu.RLock()
		validator := fc.validators[vote.Validator]
		fc.mu.RUnlock()

		signatures[i] = &crypto.QRSignature{
			Algorithm: validator.SigAlgorithm,
			Signature: vote.Signature,
			PublicKey: validator.PublicKey,
		}
		messageHashes[i] = fc.hashMessage(vote)
	}

	return crypto.AggregateSignatures(signatures, messageHashes)
}

// ValidateBlock validates a block according to consensus rules
func (fc *FastConsensus) ValidateBlock(block *types.Block, proposer types.Address) error {
	fc.mu.RLock()
	validator, exists := fc.validators[proposer]
	fc.mu.RUnlock()

	if !exists || !validator.IsActive {
		return errors.New("invalid block proposer")
	}

	// Validate block timing - simplified for now
	// In production, would check actual block timestamp
	_ = fc.blockTime // Use the field to avoid unused variable warning

	// Validate transactions in block
	for _, tx := range block.Transactions {
		if err := fc.validateTransaction(tx); err != nil {
			return fmt.Errorf("invalid transaction in block: %w", err)
		}
	}

	return nil
}

// validateTransaction validates a quantum transaction
func (fc *FastConsensus) validateTransaction(tx *types.QuantumTransaction) error {
	// Verify quantum signature
	sigHash := tx.SigningHash()
	qrSig := &crypto.QRSignature{
		Algorithm: tx.SigAlg,
		Signature: tx.Signature,
		PublicKey: tx.PublicKey,
	}

	valid, err := crypto.VerifySignature(sigHash[:], qrSig)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	if !valid {
		return errors.New("invalid quantum signature")
	}

	// Validate chain ID
	if tx.ChainID.Cmp(fc.chainID) != 0 {
		return errors.New("invalid chain ID")
	}

	return nil
}

// FinalizeBlock finalizes a block after achieving consensus
func (fc *FastConsensus) FinalizeBlock(block *types.Block, votes *crypto.AggregatedSignature) error {
	// Verify aggregated votes
	valid, err := crypto.VerifyAggregatedSignature(votes)
	if err != nil {
		return fmt.Errorf("failed to verify aggregated votes: %w", err)
	}
	if !valid {
		return errors.New("invalid aggregated votes")
	}

	// Check if we have enough votes (2/3+ of active validators)
	requiredVotes := (len(fc.validatorList) * 2 / 3) + 1
	if len(votes.Signatures) < requiredVotes {
		return fmt.Errorf("insufficient votes: got %d, need %d", len(votes.Signatures), requiredVotes)
	}

	// Update validator performance metrics
	fc.updateValidatorPerformance(block, votes)

	return nil
}

// updateValidatorPerformance updates validator performance based on participation
func (fc *FastConsensus) updateValidatorPerformance(block *types.Block, votes *crypto.AggregatedSignature) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Track which validators participated
	participatedValidators := make(map[types.Address]bool)

	for i := 0; i < len(votes.PublicKeys); i++ {
		if (votes.Bitmap & (1 << i)) != 0 {
			// Find validator by public key
			for addr, validator := range fc.validators {
				if string(validator.PublicKey) == string(votes.PublicKeys[i]) {
					participatedValidators[addr] = true
					validator.BlocksProduced++
					validator.LastActive = time.Now()
					// Improve performance score slightly
					validator.Performance = min(1.0, validator.Performance+0.001)
					break
				}
			}
		}
	}

	// Penalize validators who didn't participate
	for addr, validator := range fc.validators {
		if !participatedValidators[addr] && validator.IsActive {
			validator.BlocksMissed++
			// Reduce performance score
			validator.Performance = max(0.0, validator.Performance-0.01)

			// Deactivate if performance drops too low
			if validator.Performance < 0.5 {
				validator.IsActive = false
			}
		}
	}
}

// GetNextProposer determines the next block proposer using weighted random selection
func (fc *FastConsensus) GetNextProposer(blockHeight uint64) (types.Address, error) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.validatorList) == 0 {
		return types.Address{}, errors.New("no active validators")
	}

	// Use deterministic randomness based on block height
	seed := big.NewInt(int64(blockHeight))

	// Calculate total weighted stake
	totalWeight := big.NewInt(0)
	for _, validator := range fc.validatorList {
		weight := new(big.Int).Set(validator.Stake)
		// Apply performance multiplier
		performanceMultiplier := int64(validator.Performance * 1000)
		weight.Mul(weight, big.NewInt(performanceMultiplier))
		weight.Div(weight, big.NewInt(1000))
		totalWeight.Add(totalWeight, weight)
	}

	// Generate random number
	randomValue := new(big.Int).Mod(seed, totalWeight)

	// Select validator based on weighted stake
	currentWeight := big.NewInt(0)
	for _, validator := range fc.validatorList {
		weight := new(big.Int).Set(validator.Stake)
		performanceMultiplier := int64(validator.Performance * 1000)
		weight.Mul(weight, big.NewInt(performanceMultiplier))
		weight.Div(weight, big.NewInt(1000))

		currentWeight.Add(currentWeight, weight)
		if currentWeight.Cmp(randomValue) > 0 {
			return validator.Address, nil
		}
	}

	// Fallback to first validator
	return fc.validatorList[0].Address, nil
}

// hashMessage creates a hash of a consensus message for signing
func (fc *FastConsensus) hashMessage(msg *ConsensusMessage) []byte {
	data := fmt.Sprintf("%d:%d:%d:%s:%s:%d",
		msg.Type, msg.Epoch, msg.BlockHeight,
		msg.BlockHash.Hex(), msg.Validator.Hex(), msg.Timestamp.Unix())

	if len(msg.Data) > 0 {
		data += ":" + string(msg.Data)
	}

	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// GetConsensusInfo returns information about the current consensus state
func (fc *FastConsensus) GetConsensusInfo() map[string]interface{} {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return map[string]interface{}{
		"currentEpoch":     fc.currentEpoch,
		"blockTime":        fc.blockTime.Seconds(),
		"activeValidators": len(fc.validatorList),
		"totalValidators":  len(fc.validators),
		"minStake":         fc.minStake.String(),
		"maxValidators":    fc.maxValidators,
	}
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
