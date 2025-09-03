package consensus

import (
	"errors"
	"math/big"
	"sort"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// ValidatorInfo represents a validator in the system
type ValidatorInfo struct {
	Address    types.Address           `json:"address"`
	PublicKey  []byte                  `json:"publicKey"`
	SigAlg     crypto.SignatureAlgorithm `json:"sigAlg"`
	Stake      *big.Int                `json:"stake"`
	LastActive uint64                  `json:"lastActive"`
	Slashed    bool                    `json:"slashed"`
}

// ValidatorSet represents a set of validators
type ValidatorSet struct {
	Validators []*ValidatorInfo `json:"validators"`
	TotalStake *big.Int         `json:"totalStake"`
	Proposer   *ValidatorInfo   `json:"proposer"`
}

// NewValidatorSet creates a new validator set
func NewValidatorSet(validators []*ValidatorInfo) *ValidatorSet {
	vs := &ValidatorSet{
		Validators: make([]*ValidatorInfo, len(validators)),
		TotalStake: big.NewInt(0),
	}
	
	copy(vs.Validators, validators)
	
	// Calculate total stake
	for _, val := range vs.Validators {
		if !val.Slashed {
			vs.TotalStake.Add(vs.TotalStake, val.Stake)
		}
	}
	
	// Sort validators by stake (descending) for deterministic ordering
	sort.Slice(vs.Validators, func(i, j int) bool {
		return vs.Validators[i].Stake.Cmp(vs.Validators[j].Stake) > 0
	})
	
	return vs
}

// GetProposer returns the block proposer for a given height using VRF-like selection
func (vs *ValidatorSet) GetProposer(height uint64, seed []byte) *ValidatorInfo {
	if len(vs.Validators) == 0 {
		return nil
	}
	
	// Use deterministic pseudo-random selection based on height and seed
	// This simulates VRF behavior for proposer selection
	combinedSeed := append(types.SHA256(seed), types.SHA256(types.Uint64ToBytes(height))...)
	hash := types.SHA256(combinedSeed)
	
	// Convert hash to big int
	hashBig := new(big.Int).SetBytes(hash)
	
	// Select proposer weighted by stake
	target := new(big.Int).Mod(hashBig, vs.TotalStake)
	
	cumulative := big.NewInt(0)
	for _, val := range vs.Validators {
		if val.Slashed {
			continue
		}
		
		cumulative.Add(cumulative, val.Stake)
		if target.Cmp(cumulative) < 0 {
			return val
		}
	}
	
	// Fallback to first validator
	for _, val := range vs.Validators {
		if !val.Slashed {
			return val
		}
	}
	
	return nil
}

// Size returns the number of active validators
func (vs *ValidatorSet) Size() int {
	count := 0
	for _, val := range vs.Validators {
		if !val.Slashed {
			count++
		}
	}
	return count
}

// GetByAddress returns a validator by address
func (vs *ValidatorSet) GetByAddress(addr types.Address) *ValidatorInfo {
	for _, val := range vs.Validators {
		if val.Address.Equal(addr) {
			return val
		}
	}
	return nil
}

// IsValidator checks if an address is a validator
func (vs *ValidatorSet) IsValidator(addr types.Address) bool {
	val := vs.GetByAddress(addr)
	return val != nil && !val.Slashed
}

// AddValidator adds a new validator to the set
func (vs *ValidatorSet) AddValidator(val *ValidatorInfo) error {
	if vs.GetByAddress(val.Address) != nil {
		return errors.New("validator already exists")
	}
	
	vs.Validators = append(vs.Validators, val)
	if !val.Slashed {
		vs.TotalStake.Add(vs.TotalStake, val.Stake)
	}
	
	// Re-sort validators
	sort.Slice(vs.Validators, func(i, j int) bool {
		return vs.Validators[i].Stake.Cmp(vs.Validators[j].Stake) > 0
	})
	
	return nil
}

// RemoveValidator removes a validator from the set
func (vs *ValidatorSet) RemoveValidator(addr types.Address) error {
	for i, val := range vs.Validators {
		if val.Address.Equal(addr) {
			// Remove from slice
			vs.Validators = append(vs.Validators[:i], vs.Validators[i+1:]...)
			
			// Subtract stake if not slashed
			if !val.Slashed {
				vs.TotalStake.Sub(vs.TotalStake, val.Stake)
			}
			
			return nil
		}
	}
	
	return errors.New("validator not found")
}

// SlashValidator slashes a validator for misbehavior
func (vs *ValidatorSet) SlashValidator(addr types.Address) error {
	val := vs.GetByAddress(addr)
	if val == nil {
		return errors.New("validator not found")
	}
	
	if !val.Slashed {
		val.Slashed = true
		vs.TotalStake.Sub(vs.TotalStake, val.Stake)
	}
	
	return nil
}

// UpdateStake updates a validator's stake
func (vs *ValidatorSet) UpdateStake(addr types.Address, newStake *big.Int) error {
	val := vs.GetByAddress(addr)
	if val == nil {
		return errors.New("validator not found")
	}
	
	if !val.Slashed {
		vs.TotalStake.Sub(vs.TotalStake, val.Stake)
		vs.TotalStake.Add(vs.TotalStake, newStake)
	}
	
	val.Stake = new(big.Int).Set(newStake)
	
	// Re-sort validators
	sort.Slice(vs.Validators, func(i, j int) bool {
		return vs.Validators[i].Stake.Cmp(vs.Validators[j].Stake) > 0
	})
	
	return nil
}

// QuantumPoSConsensus implements Proof of Stake with quantum-resistant signatures
type QuantumPoSConsensus struct {
	validatorSet *ValidatorSet
	privateKey   []byte
	algorithm    crypto.SignatureAlgorithm
	address      types.Address
	
	// Configuration
	blockTime      time.Duration
	minValidators  int
	slashingWindow uint64
}

// NewQuantumPoSConsensus creates a new quantum PoS consensus engine
func NewQuantumPoSConsensus(privateKey []byte, algorithm crypto.SignatureAlgorithm, address types.Address) *QuantumPoSConsensus {
	return &QuantumPoSConsensus{
		privateKey:     privateKey,
		algorithm:      algorithm,
		address:        address,
		blockTime:      time.Second * 12, // 12 second block time
		minValidators:  1,                // Minimum for single-node testing
		slashingWindow: 100,              // Slash within 100 blocks
	}
}

// SetValidatorSet sets the active validator set
func (c *QuantumPoSConsensus) SetValidatorSet(vs *ValidatorSet) {
	c.validatorSet = vs
}

// ValidateBlock validates a block according to quantum PoS rules
func (c *QuantumPoSConsensus) ValidateBlock(block *types.Block, parentBlock *types.Block) error {
	header := block.Header
	
	// Basic validation
	if header.Number.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("invalid block number")
	}
	
	if parentBlock != nil {
		// Check parent hash
		if !header.ParentHash.Equal(parentBlock.Hash()) {
			return errors.New("invalid parent hash")
		}
		
		// Check block number progression
		expectedNumber := new(big.Int).Add(parentBlock.Number(), big.NewInt(1))
		if header.Number.Cmp(expectedNumber) != 0 {
			return errors.New("invalid block number progression")
		}
		
		// Check timestamp
		if header.Time <= parentBlock.Time() {
			return errors.New("block timestamp must be greater than parent")
		}
	}
	
	// Validate validator signature
	if header.ValidatorSig == nil {
		return errors.New("missing validator signature")
	}
	
	// Check if signer is a valid validator
	if !c.validatorSet.IsValidator(header.ValidatorAddr) {
		return errors.New("invalid validator")
	}
	
	// Verify quantum-resistant signature
	valid, err := header.VerifyValidatorSignature()
	if err != nil {
		return err
	}
	
	if !valid {
		return errors.New("invalid validator signature")
	}
	
	// Check if this validator should be the proposer
	// (simplified - in practice, you'd check against VRF proof)
	seed := parentBlock.Hash().Bytes()
	if parentBlock == nil {
		seed = types.ZeroHash.Bytes()
	}
	
	proposer := c.validatorSet.GetProposer(header.Number.Uint64(), seed)
	if proposer == nil || !proposer.Address.Equal(header.ValidatorAddr) {
		return errors.New("invalid proposer for this height")
	}
	
	return nil
}

// PrepareBlock prepares a block for proposal
func (c *QuantumPoSConsensus) PrepareBlock(parentBlock *types.Block, transactions []*types.QuantumTransaction, coinbase types.Address, gasLimit uint64) (*types.Block, error) {
	if c.validatorSet == nil {
		return nil, errors.New("validator set not initialized")
	}
	
	parentHash := types.ZeroHash
	number := big.NewInt(0)
	
	if parentBlock != nil {
		parentHash = parentBlock.Hash()
		number = new(big.Int).Add(parentBlock.Number(), big.NewInt(1))
	}
	
	// Check if we're the proposer for this height
	seed := parentHash.Bytes()
	proposer := c.validatorSet.GetProposer(number.Uint64(), seed)
	if proposer == nil || !proposer.Address.Equal(c.address) {
		return nil, errors.New("not the proposer for this height")
	}
	
	// Create block header
	header := types.NewBlockHeader(
		parentHash,
		coinbase,
		types.ZeroHash, // State root will be set by state processor
		number,
		gasLimit,
		uint64(time.Now().Unix()),
	)
	
	// Create block
	block := types.NewBlock(header, transactions, []*types.BlockHeader{})
	
	// Sign the block
	err := header.SignBlock(c.privateKey, c.algorithm, c.address)
	if err != nil {
		return nil, err
	}
	
	return block, nil
}

// GetValidatorSet returns the current validator set
func (c *QuantumPoSConsensus) GetValidatorSet() *ValidatorSet {
	return c.validatorSet
}

// IsValidator checks if the given address is a validator
func (c *QuantumPoSConsensus) IsValidator(addr types.Address) bool {
	if c.validatorSet == nil {
		return false
	}
	return c.validatorSet.IsValidator(addr)
}

// GetBlockTime returns the target block time
func (c *QuantumPoSConsensus) GetBlockTime() time.Duration {
	return c.blockTime
}

// Utility function for uint64 to bytes conversion
func uint64ToBytes(n uint64) []byte {
	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = byte(n & 0xff)
		n >>= 8
	}
	return result
}

// Add to types package helper
func init() {
	// Ensure the utility function is available in types package
	types.Uint64ToBytes = uint64ToBytes
}