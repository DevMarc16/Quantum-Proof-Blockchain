package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"quantum-blockchain/chain/consensus"
	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/governance"
	"quantum-blockchain/chain/monitoring"
	"quantum-blockchain/chain/network"
	"quantum-blockchain/chain/types"
)

// Config represents node configuration
type Config struct {
	DataDir        string   `json:"dataDir"`
	NetworkID      uint64   `json:"networkId"`
	ListenAddr     string   `json:"listenAddr"`
	HTTPPort       int      `json:"httpPort"`
	WSPort         int      `json:"wsPort"`
	BootstrapPeers []string `json:"bootstrapPeers"`
	ValidatorKey   string   `json:"validatorKey,omitempty"`
	ValidatorAlg   string   `json:"validatorAlg,omitempty"`
	GenesisConfig  string   `json:"genesisConfig,omitempty"`
	Mining         bool     `json:"mining"`
	GasLimit       uint64   `json:"gasLimit"`
	GasPrice       *big.Int `json:"gasPrice"`
}

// DefaultConfig returns default node configuration
func DefaultConfig() *Config {
	return &Config{
		DataDir:        "./data",
		NetworkID:      8888,
		ListenAddr:     "0.0.0.0:30303",
		HTTPPort:       8545,
		WSPort:         8546,
		BootstrapPeers: []string{},
		Mining:         true,                // Enable mining by default for fast block production
		GasLimit:       50000000,            // Increased for high throughput
		GasPrice:       big.NewInt(1000000), // Lower gas price for cheap transactions
	}
}

// Node represents a quantum-resistant blockchain node with multi-validator consensus
type Node struct {
	config         *Config
	multiConsensus *consensus.MultiValidatorConsensus // Production multi-validator consensus
	blockchain     *Blockchain
	txPool         *TxPool
	p2p            *P2PNetwork
	enhancedP2P    *network.EnhancedP2PNetwork // Enhanced P2P networking
	rpc            *RPCServer
	tokenSupply    *types.TokenSupply           // Native QTM token management
	gasPricing     *types.GasPriceCalculator    // Dynamic gas pricing
	governance     *governance.GovernanceSystem // On-chain governance
	monitoring     *monitoring.MetricsServer    // Monitoring and metrics

	// Validator info
	validatorPrivKey []byte
	validatorAlg     crypto.SignatureAlgorithm
	validatorAddr    types.Address

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// State
	running bool
	mining  bool
}

// NewNode creates a new node instance with QTM token and fast consensus
func NewNode(config *Config) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize native QTM token supply
	tokenSupply := types.NewTokenSupply()

	// Initialize dynamic gas pricing
	gasPricing := types.NewGasPriceCalculator()

	node := &Node{
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
		tokenSupply: tokenSupply,
		gasPricing:  gasPricing,
	}

	// Initialize validator if configured
	if config.ValidatorKey != "" {
		err := node.initValidator()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize validator: %w", err)
		}
	} else {
		log.Printf("⚠️ No validator private key found - mining will not be enabled")
	}

	// Initialize blockchain
	blockchain, err := NewBlockchain(config.DataDir, config.GenesisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize blockchain: %w", err)
	}
	node.blockchain = blockchain

	// Connect TokenSupply to StateDB for balance synchronization
	stateDBAdapter := NewStateDBAdapter(blockchain.stateDB)
	tokenSupply.SetStateDB(stateDBAdapter)

	// Initialize transaction pool with larger capacity for higher throughput
	node.txPool = NewTxPool(5000) // Max 5000 pending transactions for fast blocks

	// Initialize multi-validator consensus system
	chainID := big.NewInt(int64(config.NetworkID))
	node.multiConsensus = consensus.NewMultiValidatorConsensus(chainID)

	// Initialize governance system (will connect validator set later)
	node.governance = governance.NewGovernanceSystem(chainID, nil)

	// Initialize monitoring system
	node.monitoring = monitoring.NewMetricsServer(&monitoring.MetricsConfig{
		ListenAddr:  ":8080",
		MetricsPath: "/metrics",
		HealthPath:  "/health",
	})

	// Register validator if configured
	if node.validatorPrivKey != nil {
		log.Printf("🔑 Registering validator: %s", node.validatorAddr.Hex())

		// Initialize validator with significant stake (minimum 100K QTM)
		initialStake := new(big.Int)
		initialStake.SetString("100000000000000000000000", 10) // 100K QTM with 18 decimals

		// Set initial balance in BOTH TokenSupply AND StateDB for proper synchronization
		tokenSupply.SetBalance(node.validatorAddr, initialStake)
		blockchain.stateDB.SetBalance(node.validatorAddr, initialStake)

		// Initialize nonce to 0 if it's a new validator
		if blockchain.stateDB.GetNonce(node.validatorAddr) == 0 {
			// Nonce already 0, no need to set
			log.Printf("🔢 Validator nonce initialized: 0")
		}

		log.Printf("💰 Validator initial balance set: %s QTM",
			new(big.Int).Div(initialStake, big.NewInt(1e18)).String())

		// Register as validator in multi-validator consensus
		err = node.multiConsensus.RegisterValidator(
			node.validatorAddr,
			node.getPublicKey(),
			initialStake,
			node.validatorAlg,
			0.05, // 5% commission
		)
		if err != nil {
			return nil, fmt.Errorf("failed to register validator: %w", err)
		}

		// Stake tokens for consensus participation
		err = tokenSupply.Stake(node.validatorAddr, initialStake)
		if err != nil {
			return nil, fmt.Errorf("failed to stake tokens: %w", err)
		}

		log.Printf("✅ Validator registered successfully with stake: %s QTM", initialStake.String())
	} else {
		log.Printf("⚠️ No validator private key found - validator mode disabled")
	}

	// Initialize enhanced P2P network with security features
	node.enhancedP2P = network.NewEnhancedP2PNetwork(&network.NetworkConfig{
		ListenAddr: config.ListenAddr,
		MaxPeers:   50,
		NetworkID:  uint64(config.NetworkID),
	})

	// Initialize legacy P2P for compatibility
	node.p2p = NewP2PNetwork(config.ListenAddr, config.BootstrapPeers)

	// Initialize RPC server
	node.rpc = NewRPCServer(node, config.HTTPPort, config.WSPort)

	return node, nil
}

func (n *Node) initValidator() error {
	if n.config.ValidatorKey == "auto" {
		// Auto-generate validator key and persist it
		return n.generateAndSaveValidator()
	} else {
		// Load existing validator key
		return n.loadValidator(n.config.ValidatorKey)
	}
}

func (n *Node) generateAndSaveValidator() error {
	// Try to load existing validator key first
	keyPath := n.config.DataDir + "/validator.key"
	if existingKey, err := n.loadValidatorFromFile(keyPath); err == nil {
		log.Printf("🔑 Loaded existing validator key from %s", keyPath)
		n.validatorPrivKey = existingKey
		n.validatorAlg = crypto.SigAlgDilithium

		// Derive address from private key
		privKey, err := crypto.DilithiumPrivateKeyFromBytes(existingKey)
		if err != nil {
			return fmt.Errorf("failed to parse existing validator key: %w", err)
		}
		pubKey := privKey.Public()
		n.validatorAddr = types.PublicKeyToAddress(pubKey.Bytes())

		log.Printf("🔑 Validator address: %s", n.validatorAddr.Hex())
		return nil
	}

	// Generate new validator key
	log.Printf("🔑 Generating new validator key...")
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate validator key: %w", err)
	}

	n.validatorPrivKey = privKey.Bytes()
	n.validatorAlg = crypto.SigAlgDilithium
	n.validatorAddr = types.PublicKeyToAddress(pubKey.Bytes())

	// Save the key for persistence
	err = n.saveValidatorToFile(keyPath, n.validatorPrivKey)
	if err != nil {
		log.Printf("⚠️ Failed to save validator key: %v", err)
		// Continue anyway - validator will work but key won't persist
	} else {
		log.Printf("💾 Saved validator key to %s", keyPath)
	}

	log.Printf("🔑 Generated validator address: %s", n.validatorAddr.Hex())
	return nil
}

func (n *Node) loadValidator(keyHex string) error {
	// Load validator from hex string
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid validator key hex: %w", err)
	}

	n.validatorPrivKey = keyBytes
	n.validatorAlg = crypto.SigAlgDilithium

	// Derive address
	privKey, err := crypto.DilithiumPrivateKeyFromBytes(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse validator key: %w", err)
	}
	pubKey := privKey.Public()
	n.validatorAddr = types.PublicKeyToAddress(pubKey.Bytes())

	return nil
}

func (n *Node) loadValidatorFromFile(keyPath string) ([]byte, error) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("validator key file not found: %s", keyPath)
	}

	hexData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read validator key file: %w", err)
	}

	keyBytes, err := hex.DecodeString(string(hexData))
	if err != nil {
		return nil, fmt.Errorf("invalid hex in validator key file: %w", err)
	}

	return keyBytes, nil
}

func (n *Node) saveValidatorToFile(keyPath string, keyBytes []byte) error {
	// Create directory if it doesn't exist
	err := os.MkdirAll(n.config.DataDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	hexData := hex.EncodeToString(keyBytes)
	err = ioutil.WriteFile(keyPath, []byte(hexData), 0600) // Secure permissions
	if err != nil {
		return fmt.Errorf("failed to write validator key file: %w", err)
	}

	return nil
}

func (n *Node) getPublicKey() []byte {
	if n.validatorAlg == crypto.SigAlgDilithium {
		// Derive public key from private key
		privKey, _ := crypto.DilithiumPrivateKeyFromBytes(n.validatorPrivKey)
		return privKey.Bytes()[:crypto.DilithiumPublicKeySize] // Extract public portion
	}
	return nil
}

// Start starts the node
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("node already running")
	}

	log.Printf("Starting Quantum-Resistant Blockchain Node...")
	log.Printf("Network ID: %d", n.config.NetworkID)
	log.Printf("Listen Address: %s", n.config.ListenAddr)

	// Start enhanced P2P network with security features (temporarily disabled to avoid port conflicts)
	// TODO: Configure enhanced P2P to use different port than legacy P2P
	// if err := n.enhancedP2P.Start(); err != nil {
	//	return fmt.Errorf("failed to start enhanced P2P network: %w", err)
	// }

	// Start legacy P2P for compatibility
	if err := n.p2p.Start(n.ctx); err != nil {
		return fmt.Errorf("failed to start P2P network: %w", err)
	}

	// Start RPC server
	if err := n.rpc.Start(); err != nil {
		return fmt.Errorf("failed to start RPC server: %w", err)
	}

	// Start monitoring and metrics collection
	if err := n.monitoring.Start(); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// Start governance system
	// Governance system is now active and ready for proposals
	log.Printf("🏛️ Governance system initialized")

	// Start block production if validator and mining enabled
	log.Printf("🔍 Checking mining conditions: multiConsensus=%v, Mining=%v, validatorPrivKey=%v",
		n.multiConsensus != nil, n.config.Mining, n.validatorPrivKey != nil)

	if n.multiConsensus != nil && n.config.Mining && n.validatorPrivKey != nil {
		log.Printf("🔧 Setting mining to true...")
		n.mining = true // Set directly to avoid deadlock (we already hold the mutex)
		log.Printf("🚀 Starting multi-validator consensus...")
		n.startMultiValidatorConsensus()
		log.Printf("✅ Fast block production started (2-second blocks)")
		log.Printf("Started mining")
	} else {
		log.Printf("⚠️ Mining not started - missing requirements")
	}

	n.running = true
	log.Printf("Node started successfully")

	return nil
}

// Stop stops the node
func (n *Node) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		return
	}

	log.Printf("Stopping node...")

	n.cancel()
	n.wg.Wait()

	if n.rpc != nil {
		n.rpc.Stop()
	}

	if n.p2p != nil {
		n.p2p.Stop()
	}

	if n.blockchain != nil {
		n.blockchain.Close()
	}

	n.running = false
	log.Printf("Node stopped")
}

// AddTransaction adds a transaction to the pool
func (n *Node) AddTransaction(tx *types.QuantumTransaction) error {
	// Validate transaction
	valid, err := tx.VerifySignature()
	if err != nil {
		return fmt.Errorf("transaction verification failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid transaction signature")
	}

	// Add to pool
	return n.txPool.AddTransaction(tx)
}

// GetBlockchain returns the blockchain
func (n *Node) GetBlockchain() *Blockchain {
	return n.blockchain
}

// GetTxPool returns the transaction pool
func (n *Node) GetTxPool() *TxPool {
	return n.txPool
}

func (n *Node) startBlockProduction() {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		// Fast consensus: 2-second block time for Flare-like performance
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		log.Printf("Started fast block production (2-second blocks)")

		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				n.produceFastBlock()
			}
		}
	}()
}

// startMultiValidatorConsensus starts the multi-validator consensus engine
func (n *Node) startMultiValidatorConsensus() {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		// Multi-validator consensus: 2-second block time with validator rotation
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		log.Printf("Started multi-validator consensus (2-second blocks)")

		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				n.produceConsensusBlock()
			}
		}
	}()
}

// produceFastBlock produces blocks using fast consensus with QTM rewards
func (n *Node) produceFastBlock() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.mining {
		return
	}

	// Get current head
	currentBlock := n.blockchain.GetCurrentBlock()
	blockHeight := new(big.Int).Add(currentBlock.Number(), big.NewInt(1))

	// Check if this validator should propose next block
	nextProposer, err := n.multiConsensus.GetNextProposer(blockHeight.Uint64())
	if err != nil {
		log.Printf("Failed to get next proposer: %v", err)
		return
	}

	if nextProposer != n.validatorAddr {
		// Not our turn to propose
		return
	}

	// Update network load for dynamic gas pricing
	pendingCount := n.txPool.Size()
	networkLoad := float64(pendingCount) / 5000.0 // 5000 is max pool size
	n.gasPricing.UpdateNetworkLoad(networkLoad)

	// Get pending transactions with higher limit for throughput
	transactions := n.txPool.GetPendingTransactions(500) // Up to 500 tx per 2-second block!
	if len(transactions) > 0 {
		log.Printf("📦 Including %d transactions in block", len(transactions))
	}

	// Create block with optimized gas limit
	blockGasLimit := uint64(types.DefaultBlockGasLimit) // 50M gas for high throughput

	block := types.NewBlock(&types.BlockHeader{
		ParentHash:  currentBlock.Hash(),
		UncleHash:   types.ZeroHash,    // No uncles in quantum blockchain
		Coinbase:    n.validatorAddr,   // Set validator as coinbase
		Root:        types.ZeroHash,    // State root - simplified for now
		TxHash:      types.ZeroHash,    // Transaction root - will be calculated
		ReceiptHash: types.ZeroHash,    // Receipt root - simplified for now
		Bloom:       make([]byte, 256), // Empty bloom filter
		Difficulty:  big.NewInt(1),     // Fixed difficulty for PoS
		Number:      blockHeight,
		GasLimit:    blockGasLimit,
		GasUsed:     n.calculateGasUsed(transactions),
		Time:        uint64(time.Now().Unix()), // Add current timestamp
		Extra:       []byte("Quantum-Fast"),    // Extra data
		MixDigest:   types.ZeroHash,            // Not used in PoS
		Nonce:       0,                         // Not used in PoS
	}, transactions, nil)

	// Sign the block with validator signature
	err = block.Header.SignBlock(n.validatorPrivKey, n.validatorAlg, n.validatorAddr)
	if err != nil {
		log.Printf("Failed to sign block: %v", err)
		return
	}

	// Check if this validator should propose the next block
	proposer, err := n.multiConsensus.GetNextProposer(blockHeight.Uint64())
	if err != nil {
		log.Printf("Failed to get next proposer: %v", err)
		return
	}

	if proposer != n.validatorAddr {
		log.Printf("Not our turn to propose (proposer: %s)", proposer.Hex())
		return
	}

	// Add block to blockchain
	err = n.blockchain.AddBlock(block)
	if err != nil {
		log.Printf("Failed to add block: %v", err)
		return
	}

	// Remove transactions from pool
	for _, tx := range transactions {
		n.txPool.RemoveTransaction(tx.Hash())
	}

	// Calculate transaction fees from included transactions
	transactionFees := big.NewInt(0)
	for _, tx := range transactions {
		fee := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(tx.Gas)))
		transactionFees.Add(transactionFees, fee)
	}

	// Calculate block reward using tokenomics engine
	blockReward := new(big.Int)
	blockReward.SetString(types.BlockReward, 10)

	// Distribute rewards to the actual block proposer (determined by consensus)
	err = n.multiConsensus.DistributeBlockReward(nextProposer, blockReward, transactionFees, n.tokenSupply)
	if err != nil {
		log.Printf("Failed to distribute block reward: %v", err)
	}

	log.Printf("🚀 Fast block #%d: %d tx, %.1f%% load, proposer: %s, reward: %s QTM (+ fees: %s QTM)",
		blockHeight.Uint64(), len(transactions), networkLoad*100,
		nextProposer.Hex()[:10]+"...",
		new(big.Int).Div(blockReward, big.NewInt(1e18)).String(),
		new(big.Int).Div(transactionFees, big.NewInt(1e18)).String())

	// Broadcast block to peers
	n.p2p.BroadcastBlock(block)
}

// produceConsensusBlock produces blocks using multi-validator consensus with advanced features
func (n *Node) produceConsensusBlock() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.mining {
		return
	}

	// Get current head
	currentBlock := n.blockchain.GetCurrentBlock()
	blockHeight := new(big.Int).Add(currentBlock.Number(), big.NewInt(1))

	// Check if this validator should propose next block using multi-validator consensus
	nextProposer, err := n.multiConsensus.GetNextProposer(blockHeight.Uint64())
	if err != nil {
		log.Printf("Failed to get next proposer: %v", err)
		return
	}

	if nextProposer != n.validatorAddr {
		// Not our turn to propose, but validate incoming blocks
		return
	}

	// Update metrics
	// Monitor block proposal (metrics implementation pending)
	log.Printf("📊 Block proposed by validator: %s", n.validatorAddr.Hex())

	// Update network load for dynamic gas pricing
	pendingCount := n.txPool.Size()
	networkLoad := float64(pendingCount) / 5000.0 // 5000 is max pool size
	n.gasPricing.UpdateNetworkLoad(networkLoad)

	// Get pending transactions with higher limit for throughput
	transactions := n.txPool.GetPendingTransactions(500) // Up to 500 tx per 2-second block!
	if len(transactions) > 0 {
		log.Printf("📦 Including %d transactions in block", len(transactions))
	}

	// Create block with optimized gas limit
	blockGasLimit := uint64(types.DefaultBlockGasLimit) // 50M gas for high throughput

	block := types.NewBlock(&types.BlockHeader{
		ParentHash:  currentBlock.Hash(),
		UncleHash:   types.ZeroHash,    // No uncles in quantum blockchain
		Coinbase:    n.validatorAddr,   // Set validator as coinbase
		Root:        types.ZeroHash,    // State root - simplified for now
		TxHash:      types.ZeroHash,    // Transaction root - will be calculated
		ReceiptHash: types.ZeroHash,    // Receipt root - simplified for now
		Bloom:       make([]byte, 256), // Empty bloom filter
		Difficulty:  big.NewInt(1),     // Fixed difficulty for PoS
		Number:      blockHeight,
		GasLimit:    blockGasLimit,
		GasUsed:     n.calculateGasUsed(transactions),
		Time:        uint64(time.Now().Unix()), // Add current timestamp
		Extra:       []byte("Quantum-Multi"),   // Extra data
		MixDigest:   types.ZeroHash,            // Not used in PoS
		Nonce:       0,                         // Not used in PoS
	}, transactions, nil)

	// Sign block with validator's quantum-resistant key
	blockHash := block.Hash()
	signature, err := crypto.SignMessage(blockHash.Bytes(), n.validatorAlg, n.validatorPrivKey)
	if err != nil {
		log.Printf("Failed to sign block: %v", err)
		return
	}

	// Set signature in block
	// TODO: Implement block signature storage in Block type
	// block.SetSignature(signature, n.getPublicKey(), n.validatorAlg)
	log.Printf("Block signed with quantum signature (len: %d bytes)", len(signature.Signature))

	// Add block to blockchain
	err = n.blockchain.AddBlock(block)
	if err != nil {
		log.Printf("Failed to add block: %v", err)
		return
	}

	// Calculate transaction fees from included transactions
	transactionFees := big.NewInt(0)
	for _, tx := range transactions {
		fee := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(tx.Gas)))
		transactionFees.Add(transactionFees, fee)
	}

	// Calculate block reward using standard tokenomics
	blockReward := big.NewInt(1000000000000000000) // 1 QTM reward per block

	// Distribute rewards to the actual block proposer (determined by consensus)
	err = n.multiConsensus.DistributeBlockReward(nextProposer, blockReward, transactionFees, n.tokenSupply)
	if err != nil {
		log.Printf("Failed to distribute block reward: %v", err)
	}

	// Remove included transactions from pool (method name may differ)
	for _, tx := range transactions {
		n.txPool.RemoveTransaction(tx.Hash())
	}

	// Update monitoring metrics (metrics implementation pending)
	log.Printf("📊 Block accepted with %d transactions", len(transactions))

	log.Printf("🏛️ Multi-validator block #%d: %d tx, %.1f%% load, proposer: %s, reward: %s QTM (+ fees: %s QTM)",
		blockHeight.Uint64(), len(transactions), networkLoad*100,
		nextProposer.Hex()[:10]+"...",
		new(big.Int).Div(blockReward, big.NewInt(1e18)).String(),
		new(big.Int).Div(transactionFees, big.NewInt(1e18)).String())

	// Broadcast block to enhanced P2P network
	// TODO: Implement BroadcastBlock method in EnhancedP2PNetwork
	// n.enhancedP2P.BroadcastBlock(block)
	log.Printf("🌐 Block broadcast via P2P network (enhanced implementation pending)")

	// Also broadcast to legacy P2P for compatibility
	n.p2p.BroadcastBlock(block)
}

// calculateGasUsed calculates total gas used by transactions with optimized quantum costs
func (n *Node) calculateGasUsed(transactions []*types.QuantumTransaction) uint64 {
	totalGas := uint64(0)
	networkLoad := n.gasPricing.CurrentLoad

	for _, tx := range transactions {
		// Base transaction cost (much lower than Ethereum's 21000)
		gasUsed := uint64(5000) // Reduced base cost

		// Add data cost (reduced rate)
		gasUsed += uint64(len(tx.Data)) * 2 // 2 gas per byte vs Ethereum's 16

		// Add quantum signature verification cost (heavily optimized)
		switch tx.SigAlg {
		case crypto.SigAlgDilithium:
			gasUsed += 800 // Reduced from 50000!
		case crypto.SigAlgFalcon:
			gasUsed += 600 // Reduced from 30000!
		default:
			gasUsed += 800
		}

		// Dynamic adjustment based on network load (minimal impact)
		loadMultiplier := 1.0 + (networkLoad * 0.2) // Max 1.2x increase
		gasUsed = uint64(float64(gasUsed) * loadMultiplier)

		totalGas += gasUsed
	}

	return totalGas
}

// SetMining starts or stops mining
func (n *Node) SetMining(mining bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.mining = mining
	if mining {
		log.Printf("Started mining")
	} else {
		log.Printf("Stopped mining")
	}
}

// IsMining returns whether the node is mining
func (n *Node) IsMining() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.mining
}

// GetValidatorAddress returns the validator address
func (n *Node) GetValidatorAddress() types.Address {
	return n.validatorAddr
}

// GetConfig returns the node configuration
func (n *Node) GetConfig() *Config {
	return n.config
}

// GetTokenSupply returns the native QTM token supply manager
func (n *Node) GetTokenSupply() *types.TokenSupply {
	return n.tokenSupply
}

// GetMultiConsensus returns the multi-validator consensus mechanism
func (n *Node) GetMultiConsensus() *consensus.MultiValidatorConsensus {
	return n.multiConsensus
}

// GetGasPricing returns the dynamic gas price calculator
func (n *Node) GetGasPricing() *types.GasPriceCalculator {
	return n.gasPricing
}

// TransferQTM transfers QTM tokens between addresses
func (n *Node) TransferQTM(from, to types.Address, amount *big.Int) error {
	return n.tokenSupply.Transfer(from, to, amount)
}

// GetQTMBalance returns QTM balance for an address
func (n *Node) GetQTMBalance(addr types.Address) *big.Int {
	return n.tokenSupply.GetBalance(addr)
}

// GetTokenInfo returns information about the native QTM token
func (n *Node) GetTokenInfo() *types.TokenInfo {
	return n.tokenSupply.GetTokenInfo()
}

// GetValidators returns the current active validator set
func (n *Node) GetValidators() []*consensus.ValidatorState {
	return n.multiConsensus.GetValidatorSet()
}

// GetConsensusInfo returns consensus mechanism information
func (n *Node) GetConsensusInfo() map[string]interface{} {
	return n.multiConsensus.GetConsensusInfo()
}
