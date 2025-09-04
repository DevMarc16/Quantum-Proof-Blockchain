package types

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

// Token-related errors
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// QTM is the native token of the quantum blockchain
const (
	QTMTokenName     = "Quantum Token"
	QTMTokenSymbol   = "QTM"
	QTMTokenDecimals = 18
	QTMTotalSupply   = "1000000000" // 1 billion QTM
)

// TokenSupply represents the native token supply management
type TokenSupply struct {
	TotalSupply   *big.Int          `json:"totalSupply"`
	Circulating   *big.Int          `json:"circulating"`
	Staked        *big.Int          `json:"staked"`
	Burned        *big.Int          `json:"burned"`
	Balances      map[Address]*big.Int `json:"balances"`
	LastUpdate    time.Time         `json:"lastUpdate"`
	
	// StateDB bridge for persistent state synchronization
	stateDB       StateDBInterface  `json:"-"`
}

// NewTokenSupply creates a new token supply instance
func NewTokenSupply() *TokenSupply {
	totalSupply := new(big.Int)
	totalSupply.SetString(QTMTotalSupply+"000000000000000000", 10) // Add 18 decimals

	return &TokenSupply{
		TotalSupply: totalSupply,
		Circulating: new(big.Int).Set(totalSupply),
		Staked:      big.NewInt(0),
		Burned:      big.NewInt(0),
		Balances:    make(map[Address]*big.Int),
		LastUpdate:  time.Now(),
		stateDB:     nil, // Will be set later
	}
}

// SetStateDB sets the persistent state database for the token supply
func (ts *TokenSupply) SetStateDB(stateDB StateDBInterface) {
	ts.stateDB = stateDB
}

// GetBalance returns the QTM balance for an address
func (ts *TokenSupply) GetBalance(addr Address) *big.Int {
	if balance, exists := ts.Balances[addr]; exists {
		return new(big.Int).Set(balance)
	}
	return big.NewInt(0)
}

// SetBalance sets the QTM balance for an address
func (ts *TokenSupply) SetBalance(addr Address, amount *big.Int) {
	if amount.Sign() <= 0 {
		delete(ts.Balances, addr)
	} else {
		ts.Balances[addr] = new(big.Int).Set(amount)
	}
	ts.LastUpdate = time.Now()
}

// Transfer transfers QTM between addresses
func (ts *TokenSupply) Transfer(from, to Address, amount *big.Int) error {
	fromBalance := ts.GetBalance(from)
	if fromBalance.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}

	// Deduct from sender
	newFromBalance := new(big.Int).Sub(fromBalance, amount)
	ts.SetBalance(from, newFromBalance)

	// Add to recipient
	toBalance := ts.GetBalance(to)
	newToBalance := new(big.Int).Add(toBalance, amount)
	ts.SetBalance(to, newToBalance)

	ts.LastUpdate = time.Now()
	return nil
}

// Stake stakes QTM tokens for consensus participation
func (ts *TokenSupply) Stake(addr Address, amount *big.Int) error {
	balance := ts.GetBalance(addr)
	if balance.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}

	// Move from circulating to staked
	newBalance := new(big.Int).Sub(balance, amount)
	ts.SetBalance(addr, newBalance)
	ts.Staked.Add(ts.Staked, amount)
	ts.Circulating.Sub(ts.Circulating, amount)

	ts.LastUpdate = time.Now()
	return nil
}

// Unstake unstakes QTM tokens
func (ts *TokenSupply) Unstake(addr Address, amount *big.Int) error {
	// In production, this would check staking records
	// For now, just move from staked back to circulating
	if ts.Staked.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}

	balance := ts.GetBalance(addr)
	newBalance := new(big.Int).Add(balance, amount)
	ts.SetBalance(addr, newBalance)
	ts.Staked.Sub(ts.Staked, amount)
	ts.Circulating.Add(ts.Circulating, amount)

	ts.LastUpdate = time.Now()
	return nil
}

// Burn burns QTM tokens (deflationary mechanism)
func (ts *TokenSupply) Burn(addr Address, amount *big.Int) error {
	balance := ts.GetBalance(addr)
	if balance.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}

	// Remove from circulation permanently
	newBalance := new(big.Int).Sub(balance, amount)
	ts.SetBalance(addr, newBalance)
	ts.Burned.Add(ts.Burned, amount)
	ts.Circulating.Sub(ts.Circulating, amount)

	ts.LastUpdate = time.Now()
	return nil
}

// StateDBInterface allows TokenSupply to update the persistent state
type StateDBInterface interface {
	GetBalance(addr Address) *big.Int
	SetBalance(addr Address, balance *big.Int)
}

// Mint mints new QTM tokens (for rewards, only by validators)
func (ts *TokenSupply) Mint(addr Address, amount *big.Int) error {
	// Update TokenSupply state
	balance := ts.GetBalance(addr)
	newBalance := new(big.Int).Add(balance, amount)
	ts.SetBalance(addr, newBalance)
	
	// Increase total supply
	ts.TotalSupply.Add(ts.TotalSupply, amount)
	ts.Circulating.Add(ts.Circulating, amount)

	// Also update persistent StateDB if available
	if ts.stateDB != nil {
		currentBalance := ts.stateDB.GetBalance(addr)
		newPersistentBalance := new(big.Int).Add(currentBalance, amount)
		ts.stateDB.SetBalance(addr, newPersistentBalance)
		
		fmt.Printf("💾 StateDB balance update: %s -> %s QTM (was %s QTM)\n", 
			addr.Hex()[:10]+"...", 
			new(big.Int).Div(newPersistentBalance, big.NewInt(1e18)).String(),
			new(big.Int).Div(currentBalance, big.NewInt(1e18)).String())
	} else {
		fmt.Printf("⚠️ StateDB not available for balance update\n")
	}

	ts.LastUpdate = time.Now()
	return nil
}

// MintToStateDB mints new QTM tokens and updates persistent state
func (ts *TokenSupply) MintToStateDB(addr Address, amount *big.Int, stateDB StateDBInterface) error {
	// Update TokenSupply state
	err := ts.Mint(addr, amount)
	if err != nil {
		return err
	}
	
	// Update persistent StateDB
	currentBalance := stateDB.GetBalance(addr)
	newBalance := new(big.Int).Add(currentBalance, amount)
	stateDB.SetBalance(addr, newBalance)
	
	return nil
}

// TokenInfo provides information about the native token
type TokenInfo struct {
	Name         string    `json:"name"`
	Symbol       string    `json:"symbol"`
	Decimals     uint8     `json:"decimals"`
	TotalSupply  *big.Int  `json:"totalSupply"`
	Circulating  *big.Int  `json:"circulating"`
	Staked       *big.Int  `json:"staked"`
	Burned       *big.Int  `json:"burned"`
	LastUpdate   time.Time `json:"lastUpdate"`
}

// GetTokenInfo returns information about QTM
func (ts *TokenSupply) GetTokenInfo() *TokenInfo {
	return &TokenInfo{
		Name:        QTMTokenName,
		Symbol:      QTMTokenSymbol,
		Decimals:    QTMTokenDecimals,
		TotalSupply: new(big.Int).Set(ts.TotalSupply),
		Circulating: new(big.Int).Set(ts.Circulating),
		Staked:      new(big.Int).Set(ts.Staked),
		Burned:      new(big.Int).Set(ts.Burned),
		LastUpdate:  ts.LastUpdate,
	}
}

// Economic parameters for fast transactions
const (
	// Base gas price in QTM (much lower than ETH)
	BaseGasPrice = 1000000 // 0.000001 QTM (1 micro-QTM)
	
	// Minimum gas price during congestion
	MinGasPrice = 100000   // 0.0000001 QTM (0.1 micro-QTM)
	
	// Maximum gas price cap
	MaxGasPrice = 10000000000 // 0.01 QTM (10 milli-QTM)
	
	// Gas limit for quantum signature verification
	QuantumSigGas = 5000     // Much lower than standard 21000
	
	// Gas limit for aggregated signatures
	AggregatedSigGas = 1000  // Even cheaper for aggregated sigs
	
	// Block gas limit (higher for throughput)
	DefaultBlockGasLimit = 50000000 // 50M gas per block
	
	// Block time target (fast like Flare)
	TargetBlockTime = 2 * time.Second // 2-second blocks
	
	// Validator rewards per block
	BlockReward = "1000000000000000000" // 1 QTM per block
)

// GasPriceCalculator calculates optimal gas prices for quantum transactions
type GasPriceCalculator struct {
	BasePrice    *big.Int
	MinPrice     *big.Int
	MaxPrice     *big.Int
	CurrentLoad  float64  // Network congestion (0.0 to 1.0)
}

// NewGasPriceCalculator creates a new gas price calculator
func NewGasPriceCalculator() *GasPriceCalculator {
	return &GasPriceCalculator{
		BasePrice:   big.NewInt(BaseGasPrice),
		MinPrice:    big.NewInt(MinGasPrice),
		MaxPrice:    big.NewInt(MaxGasPrice),
		CurrentLoad: 0.0,
	}
}

// CalculateGasPrice calculates gas price based on network load
func (gpc *GasPriceCalculator) CalculateGasPrice() *big.Int {
	// Dynamic pricing based on network congestion
	multiplier := 1.0 + (gpc.CurrentLoad * 9.0) // 1x to 10x multiplier
	
	price := new(big.Int).Set(gpc.BasePrice)
	multiplierBig := big.NewInt(int64(multiplier * 1000))
	price.Mul(price, multiplierBig)
	price.Div(price, big.NewInt(1000))
	
	// Apply bounds
	if price.Cmp(gpc.MinPrice) < 0 {
		price.Set(gpc.MinPrice)
	}
	if price.Cmp(gpc.MaxPrice) > 0 {
		price.Set(gpc.MaxPrice)
	}
	
	return price
}

// UpdateNetworkLoad updates the current network congestion level
func (gpc *GasPriceCalculator) UpdateNetworkLoad(load float64) {
	if load < 0.0 {
		load = 0.0
	}
	if load > 1.0 {
		load = 1.0
	}
	gpc.CurrentLoad = load
}