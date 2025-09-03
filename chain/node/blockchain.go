package node

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"quantum-blockchain/chain/types"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// Blockchain represents the quantum-resistant blockchain
type Blockchain struct {
	db          *leveldb.DB
	currentBlock *types.Block
	genesis     *types.Block
	mu          sync.RWMutex
	
	// State management
	stateDB    *StateDB
	receipts   map[types.Hash][]*Receipt
	
	// Chain metrics
	totalDifficulty *big.Int
	gasUsed         uint64
}

// Receipt represents a transaction receipt
type Receipt struct {
	TxHash          types.Hash    `json:"transactionHash"`
	TxIndex         uint         `json:"transactionIndex"`
	BlockHash       types.Hash    `json:"blockHash"`
	BlockNumber     *big.Int     `json:"blockNumber"`
	From            types.Address `json:"from"`
	To              *types.Address `json:"to"`
	GasUsed         uint64       `json:"gasUsed"`
	CumulativeGasUsed uint64     `json:"cumulativeGasUsed"`
	ContractAddress *types.Address `json:"contractAddress"`
	Status          uint         `json:"status"` // 1 for success, 0 for failure
	Logs            []*Log       `json:"logs"`
}

// Log represents an event log
type Log struct {
	Address     types.Address `json:"address"`
	Topics      []types.Hash  `json:"topics"`
	Data        []byte        `json:"data"`
	BlockNumber uint64        `json:"blockNumber"`
	TxHash      types.Hash    `json:"transactionHash"`
	TxIndex     uint          `json:"transactionIndex"`
	BlockHash   types.Hash    `json:"blockHash"`
	Index       uint          `json:"logIndex"`
}

// StateDB represents the state database (simplified)
type StateDB struct {
	db      *leveldb.DB
	balances map[types.Address]*big.Int
	nonces   map[types.Address]uint64
	storage  map[types.Address]map[types.Hash][]byte
	mu       sync.RWMutex
}

// NewStateDB creates a new state database
func NewStateDB(db *leveldb.DB) *StateDB {
	return &StateDB{
		db:       db,
		balances: make(map[types.Address]*big.Int),
		nonces:   make(map[types.Address]uint64),
		storage:  make(map[types.Address]map[types.Hash][]byte),
	}
}

// GetBalance returns the balance of an address
func (s *StateDB) GetBalance(addr types.Address) *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if balance, exists := s.balances[addr]; exists {
		return new(big.Int).Set(balance)
	}
	
	// Try to load from persistent storage
	key := append([]byte("balance-"), addr.Bytes()...)
	data, err := s.db.Get(key, nil)
	if err != nil {
		return big.NewInt(0)
	}
	
	balance := new(big.Int)
	balance.SetBytes(data)
	s.balances[addr] = balance
	
	return new(big.Int).Set(balance)
}

// SetBalance sets the balance of an address
func (s *StateDB) SetBalance(addr types.Address, balance *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.balances[addr] = new(big.Int).Set(balance)
	
	// Persist to storage
	key := append([]byte("balance-"), addr.Bytes()...)
	s.db.Put(key, balance.Bytes(), nil)
}

// GetNonce returns the nonce of an address
func (s *StateDB) GetNonce(addr types.Address) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if nonce, exists := s.nonces[addr]; exists {
		return nonce
	}
	
	// Try to load from persistent storage
	key := append([]byte("nonce-"), addr.Bytes()...)
	data, err := s.db.Get(key, nil)
	if err != nil {
		return 0
	}
	
	nonce := new(big.Int)
	nonce.SetBytes(data)
	s.nonces[addr] = nonce.Uint64()
	
	return nonce.Uint64()
}

// SetNonce sets the nonce of an address
func (s *StateDB) SetNonce(addr types.Address, nonce uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.nonces[addr] = nonce
	
	// Persist to storage
	key := append([]byte("nonce-"), addr.Bytes()...)
	s.db.Put(key, big.NewInt(int64(nonce)).Bytes(), nil)
}

// NewBlockchain creates a new blockchain instance
func NewBlockchain(dataDir string) (*Blockchain, error) {
	// Ensure data directory exists
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Open database
	dbPath := filepath.Join(dataDir, "blockchain.db")
	db, err := leveldb.OpenFile(dbPath, &opt.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	blockchain := &Blockchain{
		db:              db,
		receipts:        make(map[types.Hash][]*Receipt),
		totalDifficulty: big.NewInt(0),
		gasUsed:         0,
	}
	
	// Initialize state database
	blockchain.stateDB = NewStateDB(db)
	
	// Load or create genesis block
	genesis, err := blockchain.loadOrCreateGenesis()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize genesis: %w", err)
	}
	
	blockchain.genesis = genesis
	blockchain.currentBlock = genesis
	
	// Load current head
	currentHash, err := db.Get([]byte("current-head"), nil)
	if err == nil {
		hash := types.BytesToHash(currentHash)
		block, err := blockchain.getBlockByHash(hash)
		if err == nil {
			blockchain.currentBlock = block
		}
	}
	
	return blockchain, nil
}

func (bc *Blockchain) loadOrCreateGenesis() (*types.Block, error) {
	// Try to load existing genesis
	genesisHash, err := bc.db.Get([]byte("genesis"), nil)
	if err == nil {
		hash := types.BytesToHash(genesisHash)
		return bc.getBlockByHash(hash)
	}
	
	// Create new genesis block
	genesis := types.Genesis()
	
	// Initialize genesis state
	bc.initializeGenesisState(genesis)
	
	// Store genesis block
	err = bc.storeBlock(genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to store genesis block: %w", err)
	}
	
	// Mark as genesis
	bc.db.Put([]byte("genesis"), genesis.Hash().Bytes(), nil)
	bc.db.Put([]byte("current-head"), genesis.Hash().Bytes(), nil)
	
	return genesis, nil
}

func (bc *Blockchain) initializeGenesisState(genesis *types.Block) {
	// Give initial balance to a test address for demo purposes
	testAddr, _ := types.HexToAddress("0x0000000000000000000000000000000000000001")
	initialBalance := new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18)) // 1M tokens
	
	bc.stateDB.SetBalance(testAddr, initialBalance)
	bc.stateDB.SetNonce(testAddr, 0)
}

// AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(block *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	// Validate block
	err := bc.validateBlock(block)
	if err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}
	
	// Execute transactions
	receipts, err := bc.executeTransactions(block)
	if err != nil {
		return fmt.Errorf("transaction execution failed: %w", err)
	}
	
	// Store receipts
	bc.receipts[block.Hash()] = receipts
	
	// Store block
	err = bc.storeBlock(block)
	if err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}
	
	// Update current head
	bc.currentBlock = block
	bc.db.Put([]byte("current-head"), block.Hash().Bytes(), nil)
	
	return nil
}

func (bc *Blockchain) validateBlock(block *types.Block) error {
	// Check parent hash
	if !block.ParentHash().Equal(bc.currentBlock.Hash()) {
		return fmt.Errorf("invalid parent hash")
	}
	
	// Check block number
	expectedNumber := new(big.Int).Add(bc.currentBlock.Number(), big.NewInt(1))
	if block.Number().Cmp(expectedNumber) != 0 {
		return fmt.Errorf("invalid block number")
	}
	
	// Check timestamp
	if block.Time() <= bc.currentBlock.Time() {
		return fmt.Errorf("block timestamp must be greater than parent")
	}
	
	// Validate transactions
	for _, tx := range block.Transactions {
		valid, err := tx.VerifySignature()
		if err != nil {
			return fmt.Errorf("transaction verification failed: %w", err)
		}
		if !valid {
			return fmt.Errorf("invalid transaction signature")
		}
		
		// Check nonce
		expectedNonce := bc.stateDB.GetNonce(tx.From())
		if tx.GetNonce() != expectedNonce {
			return fmt.Errorf("invalid nonce for transaction from %s", tx.From().Hex())
		}
		
		// Check balance
		balance := bc.stateDB.GetBalance(tx.From())
		cost := new(big.Int).Mul(big.NewInt(int64(tx.GetGas())), tx.GetGasPrice())
		cost.Add(cost, tx.GetValue())
		
		if balance.Cmp(cost) < 0 {
			return fmt.Errorf("insufficient balance for transaction from %s", tx.From().Hex())
		}
	}
	
	return nil
}

func (bc *Blockchain) executeTransactions(block *types.Block) ([]*Receipt, error) {
	receipts := make([]*Receipt, 0, len(block.Transactions))
	cumulativeGasUsed := uint64(0)
	
	for i, tx := range block.Transactions {
		receipt, err := bc.executeTransaction(tx, block, uint(i), cumulativeGasUsed)
		if err != nil {
			return nil, fmt.Errorf("failed to execute transaction %s: %w", tx.Hash().Hex(), err)
		}
		
		cumulativeGasUsed += receipt.GasUsed
		receipt.CumulativeGasUsed = cumulativeGasUsed
		receipts = append(receipts, receipt)
	}
	
	// Update block gas used
	block.Header.GasUsed = cumulativeGasUsed
	
	return receipts, nil
}

func (bc *Blockchain) executeTransaction(tx *types.QuantumTransaction, block *types.Block, txIndex uint, cumulativeGasUsed uint64) (*Receipt, error) {
	from := tx.From()
	
	// Deduct gas cost and value
	balance := bc.stateDB.GetBalance(from)
	cost := new(big.Int).Mul(big.NewInt(int64(tx.GetGas())), tx.GetGasPrice())
	cost.Add(cost, tx.GetValue())
	
	balance.Sub(balance, cost)
	bc.stateDB.SetBalance(from, balance)
	
	// Increment nonce
	nonce := bc.stateDB.GetNonce(from)
	bc.stateDB.SetNonce(from, nonce+1)
	
	// Transfer value if not contract creation
	var contractAddress *types.Address
	gasUsed := uint64(21000) // Base transaction cost
	
	if tx.IsContractCreation() {
		// Contract creation (simplified)
		// In a real implementation, this would deploy the contract code
		addr := types.PublicKeyToAddress(append(from.Bytes(), byte(nonce)))
		contractAddress = &addr
		gasUsed += uint64(len(tx.GetData())) * 4 // 4 gas per byte of data
	} else {
		// Value transfer
		if tx.GetTo() != nil {
			toBalance := bc.stateDB.GetBalance(*tx.GetTo())
			toBalance.Add(toBalance, tx.GetValue())
			bc.stateDB.SetBalance(*tx.GetTo(), toBalance)
		}
		
		// Add gas for data
		gasUsed += uint64(len(tx.GetData())) * 4
	}
	
	// Refund unused gas
	if gasUsed < tx.GetGas() {
		refund := new(big.Int).Mul(big.NewInt(int64(tx.GetGas()-gasUsed)), tx.GetGasPrice())
		balance = bc.stateDB.GetBalance(from)
		balance.Add(balance, refund)
		bc.stateDB.SetBalance(from, balance)
	}
	
	// Give gas fees to block producer
	gasFees := new(big.Int).Mul(big.NewInt(int64(gasUsed)), tx.GetGasPrice())
	producerBalance := bc.stateDB.GetBalance(block.Coinbase())
	producerBalance.Add(producerBalance, gasFees)
	bc.stateDB.SetBalance(block.Coinbase(), producerBalance)
	
	return &Receipt{
		TxHash:            tx.Hash(),
		TxIndex:           txIndex,
		BlockHash:         block.Hash(),
		BlockNumber:       block.Number(),
		From:              from,
		To:                tx.GetTo(),
		GasUsed:           gasUsed,
		CumulativeGasUsed: cumulativeGasUsed + gasUsed,
		ContractAddress:   contractAddress,
		Status:            1, // Success
		Logs:              []*Log{},
	}, nil
}

func (bc *Blockchain) storeBlock(block *types.Block) error {
	// Store block
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}
	
	blockKey := append([]byte("block-"), block.Hash().Bytes()...)
	err = bc.db.Put(blockKey, blockData, nil)
	if err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}
	
	// Store height->hash mapping
	heightKey := append([]byte("height-"), big.NewInt(0).SetUint64(block.Number().Uint64()).Bytes()...)
	err = bc.db.Put(heightKey, block.Hash().Bytes(), nil)
	if err != nil {
		return fmt.Errorf("failed to store height mapping: %w", err)
	}
	
	return nil
}

func (bc *Blockchain) getBlockByHash(hash types.Hash) (*types.Block, error) {
	blockKey := append([]byte("block-"), hash.Bytes()...)
	blockData, err := bc.db.Get(blockKey, nil)
	if err != nil {
		return nil, fmt.Errorf("block not found: %w", err)
	}
	
	var block types.Block
	err = json.Unmarshal(blockData, &block)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}
	
	return &block, nil
}

// GetCurrentBlock returns the current head block
func (bc *Blockchain) GetCurrentBlock() *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	return bc.currentBlock
}

// GetBlockByNumber returns a block by number
func (bc *Blockchain) GetBlockByNumber(number *big.Int) (*types.Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	heightKey := append([]byte("height-"), number.Bytes()...)
	hashData, err := bc.db.Get(heightKey, nil)
	if err != nil {
		return nil, fmt.Errorf("block not found at height %s", number.String())
	}
	
	hash := types.BytesToHash(hashData)
	return bc.getBlockByHash(hash)
}

// GetBlockByHash returns a block by hash
func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*types.Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	return bc.getBlockByHash(hash)
}

// GetTransactionReceipt returns a transaction receipt
func (bc *Blockchain) GetTransactionReceipt(txHash types.Hash) (*Receipt, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	// Find the receipt in memory first
	for _, receipts := range bc.receipts {
		for _, receipt := range receipts {
			if receipt.TxHash.Equal(txHash) {
				return receipt, nil
			}
		}
	}
	
	return nil, fmt.Errorf("transaction receipt not found")
}

// GetBalance returns the balance of an address
func (bc *Blockchain) GetBalance(addr types.Address) *big.Int {
	return bc.stateDB.GetBalance(addr)
}

// GetNonce returns the nonce of an address
func (bc *Blockchain) GetNonce(addr types.Address) uint64 {
	return bc.stateDB.GetNonce(addr)
}

// Close closes the blockchain database
func (bc *Blockchain) Close() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if bc.db != nil {
		return bc.db.Close()
	}
	
	return nil
}