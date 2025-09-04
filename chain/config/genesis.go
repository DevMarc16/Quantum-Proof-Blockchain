package config

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"quantum-blockchain/chain/types"
)

// GenesisConfig represents the genesis block configuration
type GenesisConfig struct {
	Config     *ChainConfig               `json:"config"`
	Difficulty string                     `json:"difficulty"`
	GasLimit   string                     `json:"gasLimit"`
	Alloc      map[string]*GenesisAccount `json:"alloc"`
	Validators []GenesisValidator         `json:"validators,omitempty"`
}

// ChainConfig represents the chain configuration
type ChainConfig struct {
	ChainID             uint64 `json:"chainId"`
	HomesteadBlock      uint64 `json:"homesteadBlock"`
	EIP150Block         uint64 `json:"eip150Block"`
	EIP155Block         uint64 `json:"eip155Block"`
	EIP158Block         uint64 `json:"eip158Block"`
	ByzantiumBlock      uint64 `json:"byzantiumBlock"`
	ConstantinopleBlock uint64 `json:"constantinopleBlock"`
	PetersburgBlock     uint64 `json:"petersburgBlock"`
	IstanbulBlock       uint64 `json:"istanbulBlock"`
	BerlinBlock         uint64 `json:"berlinBlock"`
	LondonBlock         uint64 `json:"londonBlock"`
}

// GenesisAccount represents a genesis account allocation
type GenesisAccount struct {
	Balance string            `json:"balance"`
	Code    string            `json:"code,omitempty"`
	Storage map[string]string `json:"storage,omitempty"`
}

// GenesisValidator represents a genesis validator
type GenesisValidator struct {
	Address string `json:"address"`
	Stake   string `json:"stake"`
}

// LoadGenesisConfig loads the genesis configuration from a file
func LoadGenesisConfig(path string) (*GenesisConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("genesis config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis config: %w", err)
	}

	// Parse JSON
	var genesis GenesisConfig
	err = json.Unmarshal(data, &genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to parse genesis config: %w", err)
	}

	// Validate configuration
	err = genesis.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid genesis config: %w", err)
	}

	return &genesis, nil
}

// Validate validates the genesis configuration
func (g *GenesisConfig) Validate() error {
	if g.Config == nil {
		return fmt.Errorf("missing chain config")
	}

	if g.Config.ChainID == 0 {
		return fmt.Errorf("invalid chain ID: must be greater than 0")
	}

	if g.Difficulty == "" {
		return fmt.Errorf("missing difficulty")
	}

	if g.GasLimit == "" {
		return fmt.Errorf("missing gas limit")
	}

	// Validate difficulty
	_, success := new(big.Int).SetString(g.Difficulty, 0)
	if !success {
		return fmt.Errorf("invalid difficulty format: %s", g.Difficulty)
	}

	// Validate gas limit
	_, success = new(big.Int).SetString(g.GasLimit, 0)
	if !success {
		return fmt.Errorf("invalid gas limit format: %s", g.GasLimit)
	}

	// Validate allocations
	for addrStr, account := range g.Alloc {
		// Validate address format
		_, err := types.HexToAddress(addrStr)
		if err != nil {
			return fmt.Errorf("invalid address in alloc: %s", addrStr)
		}

		// Validate balance format
		if account.Balance == "" {
			return fmt.Errorf("missing balance for address %s", addrStr)
		}

		_, success := new(big.Int).SetString(account.Balance, 0)
		if !success {
			return fmt.Errorf("invalid balance format for address %s: %s", addrStr, account.Balance)
		}
	}

	// Validate validators
	for i, validator := range g.Validators {
		// Validate validator address
		_, err := types.HexToAddress(validator.Address)
		if err != nil {
			return fmt.Errorf("invalid validator address at index %d: %s", i, validator.Address)
		}

		// Validate stake amount
		if validator.Stake == "" {
			return fmt.Errorf("missing stake for validator at index %d", i)
		}

		_, success := new(big.Int).SetString(validator.Stake, 0)
		if !success {
			return fmt.Errorf("invalid stake format for validator at index %d: %s", i, validator.Stake)
		}
	}

	return nil
}

// GetDifficulty returns the genesis difficulty as big.Int
func (g *GenesisConfig) GetDifficulty() *big.Int {
	diff, _ := new(big.Int).SetString(g.Difficulty, 0)
	return diff
}

// GetGasLimit returns the genesis gas limit as big.Int
func (g *GenesisConfig) GetGasLimit() *big.Int {
	gasLimit, _ := new(big.Int).SetString(g.GasLimit, 0)
	return gasLimit
}

// GetAllocations returns the genesis allocations with proper type conversion
func (g *GenesisConfig) GetAllocations() (map[types.Address]*big.Int, error) {
	allocations := make(map[types.Address]*big.Int)

	for addrStr, account := range g.Alloc {
		addr, err := types.HexToAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid address: %s", addrStr)
		}

		balance, success := new(big.Int).SetString(account.Balance, 0)
		if !success {
			return nil, fmt.Errorf("invalid balance: %s", account.Balance)
		}

		allocations[addr] = balance
	}

	return allocations, nil
}

// GetValidators returns the genesis validators with proper type conversion
func (g *GenesisConfig) GetValidators() ([]ValidatorInfo, error) {
	validators := make([]ValidatorInfo, len(g.Validators))

	for i, validator := range g.Validators {
		addr, err := types.HexToAddress(validator.Address)
		if err != nil {
			return nil, fmt.Errorf("invalid validator address: %s", validator.Address)
		}

		stake, success := new(big.Int).SetString(validator.Stake, 0)
		if !success {
			return nil, fmt.Errorf("invalid validator stake: %s", validator.Stake)
		}

		validators[i] = ValidatorInfo{
			Address: addr,
			Stake:   stake,
		}
	}

	return validators, nil
}

// ValidatorInfo represents validator information with proper types
type ValidatorInfo struct {
	Address types.Address
	Stake   *big.Int
}

// DefaultGenesisConfig returns a default genesis configuration
func DefaultGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		Config: &ChainConfig{
			ChainID:             8888,
			HomesteadBlock:      0,
			EIP150Block:         0,
			EIP155Block:         0,
			EIP158Block:         0,
			ByzantiumBlock:      0,
			ConstantinopleBlock: 0,
			PetersburgBlock:     0,
			IstanbulBlock:       0,
			BerlinBlock:         0,
			LondonBlock:         0,
		},
		Difficulty: "0x1",
		GasLimit:   "0x47b760", // 4,700,000
		Alloc: map[string]*GenesisAccount{
			"0x0000000000000000000000000000000000000001": {
				Balance: "0xd3c21bcecceda1000000", // 1,000,000 tokens
			},
			// Test addresses with generous initial balances for development and testing
			"0x129b052af5f7858ab578c8c8f244eaac818fa504": {
				Balance: "0x56bc75e2d630eb20000000", // 100,000,000 QTM (test address from rpc_submit test)
			},
			"0x742d35Cc2cC0b34aC2F4a7770e6Bd4b7A00B7D8F": {
				Balance: "0x56bc75e2d630eb20000000", // 100,000,000 QTM (common test recipient)
			},
			"0x951a4aece2548a5a6ffd69bab3dee1d62a6c75c1": {
				Balance: "0x56bc75e2d630eb20000000", // 100,000,000 QTM (validator address)
			},
			// Additional test addresses for comprehensive testing
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266": {
				Balance: "0x56bc75e2d630eb20000000", // 100,000,000 QTM (common hardhat test address)
			},
			"0x70997970C51812dc3A010C7d01b50e0d17dc79C8": {
				Balance: "0x56bc75e2d630eb20000000", // 100,000,000 QTM (hardhat test address 2)
			},
			// Pre-funded quantum address for contract deployment
			"0x7889e2f42d63650635ad2987bd3582f7a183e6e9": {
				Balance: "5000000000000000000", // 5 ETH (quantum deployment address)
			},
		},
		Validators: []GenesisValidator{},
	}
}
