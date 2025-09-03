package node

import (
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"

	"quantum-blockchain/chain/types"
)

// TxPool manages pending transactions
type TxPool struct {
	transactions map[types.Hash]*types.QuantumTransaction
	byNonce      map[types.Address][]*types.QuantumTransaction
	maxSize      int
	mu           sync.RWMutex
}

// NewTxPool creates a new transaction pool
func NewTxPool(maxSize int) *TxPool {
	return &TxPool{
		transactions: make(map[types.Hash]*types.QuantumTransaction),
		byNonce:      make(map[types.Address][]*types.QuantumTransaction),
		maxSize:      maxSize,
	}
}

// AddTransaction adds a transaction to the pool
func (pool *TxPool) AddTransaction(tx *types.QuantumTransaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	
	txHash := tx.Hash()
	
	// Check if transaction already exists
	if _, exists := pool.transactions[txHash]; exists {
		return errors.New("transaction already exists in pool")
	}
	
	// Check pool size
	if len(pool.transactions) >= pool.maxSize {
		return errors.New("transaction pool is full")
	}
	
	// Add to main map
	pool.transactions[txHash] = tx
	
	// Add to nonce-ordered list
	from := tx.From()
	pool.byNonce[from] = append(pool.byNonce[from], tx)
	
	// Sort by nonce
	sort.Slice(pool.byNonce[from], func(i, j int) bool {
		return pool.byNonce[from][i].GetNonce() < pool.byNonce[from][j].GetNonce()
	})
	
	return nil
}

// RemoveTransaction removes a transaction from the pool
func (pool *TxPool) RemoveTransaction(txHash types.Hash) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	
	tx, exists := pool.transactions[txHash]
	if !exists {
		return errors.New("transaction not found in pool")
	}
	
	// Remove from main map
	delete(pool.transactions, txHash)
	
	// Remove from nonce list
	from := tx.From()
	txs := pool.byNonce[from]
	for i, poolTx := range txs {
		if poolTx.Hash().Equal(txHash) {
			pool.byNonce[from] = append(txs[:i], txs[i+1:]...)
			break
		}
	}
	
	// Clean up empty address entries
	if len(pool.byNonce[from]) == 0 {
		delete(pool.byNonce, from)
	}
	
	return nil
}

// GetTransaction returns a transaction by hash
func (pool *TxPool) GetTransaction(txHash types.Hash) (*types.QuantumTransaction, bool) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	tx, exists := pool.transactions[txHash]
	return tx, exists
}

// GetPendingTransactions returns up to maxCount pending transactions
func (pool *TxPool) GetPendingTransactions(maxCount int) []*types.QuantumTransaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	var result []*types.QuantumTransaction
	count := 0
	
	// Collect transactions in nonce order from all addresses
	for _, txs := range pool.byNonce {
		for _, tx := range txs {
			if count >= maxCount {
				return result
			}
			result = append(result, tx)
			count++
		}
	}
	
	return result
}

// GetTransactionsByAddress returns transactions for a specific address
func (pool *TxPool) GetTransactionsByAddress(addr types.Address) []*types.QuantumTransaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	txs, exists := pool.byNonce[addr]
	if !exists {
		return []*types.QuantumTransaction{}
	}
	
	// Return a copy
	result := make([]*types.QuantumTransaction, len(txs))
	copy(result, txs)
	return result
}

// GetNextNonceForAddress returns the next expected nonce for an address
func (pool *TxPool) GetNextNonceForAddress(addr types.Address) uint64 {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	txs, exists := pool.byNonce[addr]
	if !exists || len(txs) == 0 {
		return 0 // Will be validated against blockchain state
	}
	
	// Find the highest nonce
	highestNonce := txs[len(txs)-1].GetNonce()
	return highestNonce + 1
}

// Size returns the number of transactions in the pool
func (pool *TxPool) Size() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	return len(pool.transactions)
}

// Clear removes all transactions from the pool
func (pool *TxPool) Clear() {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	
	pool.transactions = make(map[types.Hash]*types.QuantumTransaction)
	pool.byNonce = make(map[types.Address][]*types.QuantumTransaction)
}

// ValidateTransaction validates a transaction before adding to pool
func (pool *TxPool) ValidateTransaction(tx *types.QuantumTransaction) error {
	// Verify signature
	valid, err := tx.VerifySignature()
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid signature")
	}
	
	// Check transaction size
	if tx.Size() > 32*1024 { // Max 32KB transaction
		return errors.New("transaction too large")
	}
	
	// Check gas limit
	if tx.GetGas() > 15000000 { // Max block gas limit
		return errors.New("gas limit too high")
	}
	
	// Check gas price (minimum 1 Gwei)
	if tx.GetGasPrice().Cmp(big.NewInt(1)) < 0 {
		return errors.New("gas price too low")
	}
	
	return nil
}

// PruneTransactions removes expired or invalid transactions
func (pool *TxPool) PruneTransactions() {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	
	// Remove transactions older than 1 hour (simplified pruning)
	// In a real implementation, you'd track transaction timestamps
	currentTime := time.Now().Unix()
	
	for hash, tx := range pool.transactions {
		// Simple age check based on nonce being too old
		// This is a simplified approach - real implementation would be more sophisticated
		if currentTime-int64(tx.GetNonce()) > 3600 { // 1 hour approximation
			delete(pool.transactions, hash)
			
			// Also remove from nonce list
			from := tx.From()
			if txs, exists := pool.byNonce[from]; exists {
				for i, poolTx := range txs {
					if poolTx.Hash().Equal(hash) {
						pool.byNonce[from] = append(txs[:i], txs[i+1:]...)
						break
					}
				}
				
				if len(pool.byNonce[from]) == 0 {
					delete(pool.byNonce, from)
				}
			}
		}
	}
}

// GetStats returns transaction pool statistics
func (pool *TxPool) GetStats() map[string]interface{} {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	return map[string]interface{}{
		"pending": len(pool.transactions),
		"queued":  0, // Simplified - all transactions are considered pending
		"maxSize": pool.maxSize,
		"addresses": len(pool.byNonce),
	}
}

// Helper function for big.NewInt(1)
var Big1 = func() *types.BigInt {
	return (*types.BigInt)(types.NewBigInt(1))
}()

// Add BigInt type to types package
func init() {
	// This will be handled by the types package
}