package node

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"quantum-blockchain/chain/config"
	"quantum-blockchain/chain/evm"
	"quantum-blockchain/chain/types"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// Blockchain represents the quantum-resistant blockchain
type Blockchain struct {
	db          *leveldb.DB
	currentBlock *types.Block
	genesis     *types.Block
	genesisConfig *config.GenesisConfig
	mu          sync.RWMutex
	
	// State management
	stateDB *StateDB
	
	// EVM execution engine
	evm *evm.SimpleEVM
	
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

// Log represents an event log (alias for Ethereum log)
type Log = etypes.Log

// StateDB represents the state database with full EVM support
type StateDB struct {
	db          *leveldb.DB
	balances    map[types.Address]*big.Int
	nonces      map[types.Address]uint64
	storage     map[types.Address]map[types.Hash]types.Hash
	code        map[types.Address][]byte
	codeHashes  map[types.Address]types.Hash
	suicides    map[types.Address]bool
	mu          sync.RWMutex
}

// StateDBAdapter adapts StateDB to types.StateDBInterface
type StateDBAdapter struct {
	stateDB *StateDB
}

func NewStateDBAdapter(stateDB *StateDB) *StateDBAdapter {
	return &StateDBAdapter{stateDB: stateDB}
}

func (adapter *StateDBAdapter) GetBalance(addr types.Address) *big.Int {
	return adapter.stateDB.GetBalance(addr)
}

func (adapter *StateDBAdapter) SetBalance(addr types.Address, balance *big.Int) {
	adapter.stateDB.SetBalance(addr, balance)
}

// NewStateDB creates a new state database
func NewStateDB(db *leveldb.DB) *StateDB {
	return &StateDB{
		db:         db,
		balances:   make(map[types.Address]*big.Int),
		nonces:     make(map[types.Address]uint64),
		storage:    make(map[types.Address]map[types.Hash]types.Hash),
		code:       make(map[types.Address][]byte),
		codeHashes: make(map[types.Address]types.Hash),
		suicides:   make(map[types.Address]bool),
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

// GetState returns contract storage value
func (s *StateDB) GetState(addr types.Address, hash types.Hash) types.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if storage, exists := s.storage[addr]; exists {
		if value, exists := storage[hash]; exists {
			return value
		}
	}
	
	// Try to load from persistent storage
	key := append(append([]byte("storage-"), addr.Bytes()...), hash.Bytes()...)
	data, err := s.db.Get(key, nil)
	if err != nil {
		return types.Hash{}
	}
	
	value := types.BytesToHash(data)
	
	// Cache it
	if s.storage[addr] == nil {
		s.storage[addr] = make(map[types.Hash]types.Hash)
	}
	s.storage[addr][hash] = value
	
	return value
}

// SetState sets contract storage value
func (s *StateDB) SetState(addr types.Address, hash types.Hash, value types.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.storage[addr] == nil {
		s.storage[addr] = make(map[types.Hash]types.Hash)
	}
	s.storage[addr][hash] = value
	
	// Persist to storage
	key := append(append([]byte("storage-"), addr.Bytes()...), hash.Bytes()...)
	s.db.Put(key, value.Bytes(), nil)
}

// GetCode returns contract code
func (s *StateDB) GetCode(addr types.Address) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if code, exists := s.code[addr]; exists {
		return code
	}
	
	// Try to load from persistent storage
	key := append([]byte("code-"), addr.Bytes()...)
	data, err := s.db.Get(key, nil)
	if err != nil {
		return nil
	}
	
	s.code[addr] = data
	return data
}

// SetCode sets contract code
func (s *StateDB) SetCode(addr types.Address, code []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.code[addr] = code
	
	// Calculate code hash
	codeHash := types.Keccak256Hash(code)
	s.codeHashes[addr] = codeHash
	
	// Persist to storage
	codeKey := append([]byte("code-"), addr.Bytes()...)
	s.db.Put(codeKey, code, nil)
	
	hashKey := append([]byte("codehash-"), addr.Bytes()...)
	s.db.Put(hashKey, codeHash.Bytes(), nil)
}

// GetCodeHash returns contract code hash
func (s *StateDB) GetCodeHash(addr types.Address) types.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if hash, exists := s.codeHashes[addr]; exists {
		return hash
	}
	
	// Try to load from persistent storage
	key := append([]byte("codehash-"), addr.Bytes()...)
	data, err := s.db.Get(key, nil)
	if err != nil {
		// Calculate hash from code if available
		code := s.GetCode(addr)
		if len(code) == 0 {
			return types.Hash{}
		}
		hash := types.Keccak256Hash(code)
		s.codeHashes[addr] = hash
		return hash
	}
	
	hash := types.BytesToHash(data)
	s.codeHashes[addr] = hash
	return hash
}

// GetCodeSize returns contract code size
func (s *StateDB) GetCodeSize(addr types.Address) int {
	code := s.GetCode(addr)
	return len(code)
}

// Exist checks if account exists
func (s *StateDB) Exist(addr types.Address) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Account exists if it has balance, nonce, or code
	if _, exists := s.balances[addr]; exists {
		return true
	}
	if _, exists := s.nonces[addr]; exists {
		return true
	}
	if _, exists := s.code[addr]; exists {
		return true
	}
	
	// Check persistent storage
	balanceKey := append([]byte("balance-"), addr.Bytes()...)
	if _, err := s.db.Get(balanceKey, nil); err == nil {
		return true
	}
	
	nonceKey := append([]byte("nonce-"), addr.Bytes()...)
	if _, err := s.db.Get(nonceKey, nil); err == nil {
		return true
	}
	
	codeKey := append([]byte("code-"), addr.Bytes()...)
	if _, err := s.db.Get(codeKey, nil); err == nil {
		return true
	}
	
	return false
}

// Empty checks if account is empty (no balance, nonce, or code)
func (s *StateDB) Empty(addr types.Address) bool {
	if !s.Exist(addr) {
		return true
	}
	
	balance := s.GetBalance(addr)
	nonce := s.GetNonce(addr)
	code := s.GetCode(addr)
	
	return balance.Sign() == 0 && nonce == 0 && len(code) == 0
}

// Suicide marks an account for deletion
func (s *StateDB) Suicide(addr types.Address) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.Exist(addr) {
		return false
	}
	
	s.suicides[addr] = true
	return true
}

// HasSuicided checks if account is marked for deletion
func (s *StateDB) HasSuicided(addr types.Address) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.suicides[addr]
}

// NewBlockchain creates a new blockchain instance
func NewBlockchain(dataDir string, genesisConfigPath string) (*Blockchain, error) {
	// Load genesis configuration
	var genesisConfig *config.GenesisConfig
	var err error
	
	if genesisConfigPath != "" {
		genesisConfig, err = config.LoadGenesisConfig(genesisConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load genesis config: %w", err)
		}
	} else {
		// Use default genesis config
		genesisConfig = config.DefaultGenesisConfig()
	}
	
	// Ensure data directory exists
	err = os.MkdirAll(dataDir, 0755)
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
		genesisConfig:   genesisConfig,
		totalDifficulty: big.NewInt(0),
		gasUsed:         0,
	}
	
	// Initialize state database
	blockchain.stateDB = NewStateDB(db)
	
	// Initialize simplified EVM
	blockchain.evm = evm.NewSimpleEVM(blockchain.stateDB, big.NewInt(8888))
	
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
	// Load allocations from genesis config
	allocations, err := bc.genesisConfig.GetAllocations()
	if err != nil {
		// Log error but continue with default allocation
		fmt.Printf("Error loading genesis allocations: %v\n", err)
		
		// Fallback to default allocation
		testAddr, _ := types.HexToAddress("0x0000000000000000000000000000000000000001")
		initialBalance := new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18)) // 1M tokens
		bc.stateDB.SetBalance(testAddr, initialBalance)
		bc.stateDB.SetNonce(testAddr, 0)
		return
	}
	
	// Set balances from genesis configuration
	for addr, balance := range allocations {
		bc.stateDB.SetBalance(addr, balance)
		bc.stateDB.SetNonce(addr, 0)
	}
	
	// Initialize validators if specified
	validators, err := bc.genesisConfig.GetValidators()
	if err != nil {
		fmt.Printf("Error loading genesis validators: %v\n", err)
	} else {
		// Set validator stakes (for now, just ensure they have balance)
		for _, validator := range validators {
			currentBalance := bc.stateDB.GetBalance(validator.Address)
			if currentBalance.Sign() == 0 {
				// Give validator minimum balance if not already allocated
				bc.stateDB.SetBalance(validator.Address, validator.Stake)
			}
		}
	}
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
	
	// Store receipts to persistent storage
	err = bc.storeReceipts(block.Hash(), receipts)
	if err != nil {
		return fmt.Errorf("failed to store receipts: %w", err)
	}
	
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
			fmt.Printf("❌ Nonce mismatch: tx has nonce %d, expected %d for %s\n", 
				tx.GetNonce(), expectedNonce, tx.From().Hex())
			return fmt.Errorf("invalid nonce for transaction from %s", tx.From().Hex())
		}
		
		// Check balance
		balance := bc.stateDB.GetBalance(tx.From())
		cost := new(big.Int).Mul(big.NewInt(int64(tx.GetGas())), tx.GetGasPrice())
		cost.Add(cost, tx.GetValue())
		
		if balance.Cmp(cost) < 0 {
			fmt.Printf("❌ Insufficient balance for tx from %s: balance=%s, cost=%s\n", 
				tx.From().Hex(), balance.String(), cost.String())
			return fmt.Errorf("insufficient balance for transaction from %s: balance=%s, cost=%s", 
				tx.From().Hex(), balance.String(), cost.String())
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
	
	// Pre-execution validation
	balance := bc.stateDB.GetBalance(from)
	cost := new(big.Int).Mul(big.NewInt(int64(tx.GetGas())), tx.GetGasPrice())
	cost.Add(cost, tx.GetValue())
	
	if balance.Cmp(cost) < 0 {
		return nil, fmt.Errorf("insufficient balance for transaction")
	}
	
	// Deduct gas cost upfront
	balance.Sub(balance, cost)
	bc.stateDB.SetBalance(from, balance)
	
	// Increment nonce
	nonce := bc.stateDB.GetNonce(from)
	bc.stateDB.SetNonce(from, nonce+1)
	
	// Execute transaction using EVM
	result, err := bc.evm.ExecuteTransaction(tx, block, block.Header.GasLimit)
	
	var (
		gasUsed         uint64
		contractAddress *types.Address
		status          uint = 1 // Success
		logs            []*Log
	)
	
	if err != nil {
		// Transaction failed, but still consume gas
		gasUsed = tx.GetGas() // Use all gas on failure
		status = 0            // Failure
		logs = []*Log{}
	} else {
		gasUsed = result.GasUsed
		contractAddress = result.ContractAddress
		logs = result.Logs
		
		// Success - convert logs (simplified for now)
		// In a real implementation, we would properly convert the log structure
		logs = []*Log{} // Empty logs for now
	}
	
	// Refund unused gas
	if gasUsed < tx.GetGas() {
		refund := new(big.Int).Mul(big.NewInt(int64(tx.GetGas()-gasUsed)), tx.GetGasPrice())
		balance = bc.stateDB.GetBalance(from)
		balance.Add(balance, refund)
		bc.stateDB.SetBalance(from, balance)
	}
	
	// Give gas fees to block producer (coinbase)
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
		Status:            status,
		Logs:              logs,
	}, nil
}


func (bc *Blockchain) storeReceipts(blockHash types.Hash, receipts []*Receipt) error {
	receiptsData, err := json.Marshal(receipts)
	if err != nil {
		return fmt.Errorf("failed to marshal receipts: %w", err)
	}
	
	receiptsKey := append([]byte("receipts-"), blockHash.Bytes()...)
	return bc.db.Put(receiptsKey, receiptsData, nil)
}

func (bc *Blockchain) getReceiptsByBlockHash(blockHash types.Hash) ([]*Receipt, error) {
	receiptsKey := append([]byte("receipts-"), blockHash.Bytes()...)
	receiptsData, err := bc.db.Get(receiptsKey, nil)
	if err != nil {
		return nil, fmt.Errorf("receipts not found: %w", err)
	}
	
	var receipts []*Receipt
	err = json.Unmarshal(receiptsData, &receipts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal receipts: %w", err)
	}
	
	return receipts, nil
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
	
	// We need to search through all blocks to find the receipt
	// In a real implementation, we'd maintain a txhash->blockHash index
	// For now, we'll do a simple search starting from the current block
	currentHeight := bc.currentBlock.Number().Uint64()
	
	for height := currentHeight; height > 0; height-- {
		block, err := bc.GetBlockByNumber(big.NewInt(int64(height)))
		if err != nil {
			continue
		}
		
		receipts, err := bc.getReceiptsByBlockHash(block.Hash())
		if err != nil {
			continue
		}
		
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

// GetCode returns the contract code at the given address
func (bc *Blockchain) GetCode(addr types.Address) []byte {
	return bc.stateDB.GetCode(addr)
}

// GetState returns the contract storage value at the given address and key
func (bc *Blockchain) GetState(addr types.Address, key types.Hash) types.Hash {
	return bc.stateDB.GetState(addr, key)
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