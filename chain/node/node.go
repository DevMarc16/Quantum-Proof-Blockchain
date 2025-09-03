package node

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"quantum-blockchain/chain/consensus"
	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// Config represents node configuration
type Config struct {
	DataDir         string            `json:"dataDir"`
	NetworkID       uint64            `json:"networkId"`
	ListenAddr      string            `json:"listenAddr"`
	HTTPPort        int               `json:"httpPort"`
	WSPort          int               `json:"wsPort"`
	BootstrapPeers  []string          `json:"bootstrapPeers"`
	ValidatorKey    string            `json:"validatorKey,omitempty"`
	ValidatorAlg    string            `json:"validatorAlg,omitempty"`
	Mining          bool              `json:"mining"`
	GasLimit        uint64            `json:"gasLimit"`
	GasPrice        *big.Int          `json:"gasPrice"`
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
		Mining:         true,  // Enable mining by default for fast block production
		GasLimit:       50000000, // Increased for high throughput
		GasPrice:       big.NewInt(1000000), // Lower gas price for cheap transactions
	}
}

// Node represents a quantum-resistant blockchain node with fast consensus and QTM token
type Node struct {
	config        *Config
	consensus     *consensus.QuantumPoSConsensus  // Legacy consensus
	fastConsensus *consensus.FastConsensus        // New fast consensus (Flare-like)
	blockchain    *Blockchain
	txPool        *TxPool
	p2p           *P2PNetwork
	rpc           *RPCServer
	tokenSupply   *types.TokenSupply              // Native QTM token management
	gasPricing    *types.GasPriceCalculator       // Dynamic gas pricing
	
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
	}
	
	// Initialize blockchain
	blockchain, err := NewBlockchain(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize blockchain: %w", err)
	}
	node.blockchain = blockchain
	
	// Initialize transaction pool with larger capacity for higher throughput
	node.txPool = NewTxPool(5000) // Max 5000 pending transactions for fast blocks
	
	// Initialize fast consensus (replaces legacy QuantumPoS)
	chainID := big.NewInt(int64(config.NetworkID))
	node.fastConsensus = consensus.NewFastConsensus(chainID, tokenSupply)
	
	// Register validator if configured
	if node.validatorPrivKey != nil {
		log.Printf("🔑 Registering validator: %s", node.validatorAddr.Hex())
		
		// Initialize validator with significant stake
		initialStake := new(big.Int)
		initialStake.SetString("1000000000000000000000000", 10) // 1M QTM with 18 decimals
		
		// Give initial QTM tokens to validator
		tokenSupply.SetBalance(node.validatorAddr, initialStake)
		
		// Register as validator in fast consensus
		err = node.fastConsensus.RegisterValidator(
			node.validatorAddr,
			node.getPublicKey(),
			initialStake,
			node.validatorAlg,
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
		log.Printf("⚠️ No validator private key found - mining will not be enabled")
	}
	
	// Initialize P2P network
	node.p2p = NewP2PNetwork(config.ListenAddr, config.BootstrapPeers)
	
	// Initialize RPC server
	node.rpc = NewRPCServer(node, config.HTTPPort, config.WSPort)
	
	return node, nil
}

func (n *Node) initValidator() error {
	// For demo purposes, generate a new key if none provided
	// In production, this should load from secure storage
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		return err
	}
	
	n.validatorPrivKey = privKey.Bytes()
	n.validatorAlg = crypto.SigAlgDilithium
	n.validatorAddr = types.PublicKeyToAddress(pubKey.Bytes())
	
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
	
	// Start P2P network
	if err := n.p2p.Start(n.ctx); err != nil {
		return fmt.Errorf("failed to start P2P network: %w", err)
	}
	
	// Start RPC server
	if err := n.rpc.Start(); err != nil {
		return fmt.Errorf("failed to start RPC server: %w", err)
	}
	
	// Start block production if validator and mining enabled  
	log.Printf("🔍 Checking mining conditions: fastConsensus=%v, Mining=%v, validatorPrivKey=%v", 
		n.fastConsensus != nil, n.config.Mining, n.validatorPrivKey != nil)
	
	if n.fastConsensus != nil && n.config.Mining && n.validatorPrivKey != nil {
		log.Printf("🔧 Setting mining to true...")
		n.mining = true  // Set directly to avoid deadlock (we already hold the mutex)
		log.Printf("🚀 Starting block production...")
		n.startBlockProduction()
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
	nextProposer, err := n.fastConsensus.GetNextProposer(blockHeight.Uint64())
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
	
	// Create block with optimized gas limit
	blockGasLimit := uint64(types.DefaultBlockGasLimit) // 50M gas for high throughput
	
	block := types.NewBlock(&types.BlockHeader{
		ParentHash:  currentBlock.Hash(),
		UncleHash:   types.ZeroHash,           // No uncles in quantum blockchain
		Coinbase:    n.validatorAddr,          // Set validator as coinbase
		Root:        types.ZeroHash,           // State root - simplified for now
		TxHash:      types.ZeroHash,           // Transaction root - will be calculated
		ReceiptHash: types.ZeroHash,           // Receipt root - simplified for now
		Bloom:       make([]byte, 256),        // Empty bloom filter
		Difficulty:  big.NewInt(1),            // Fixed difficulty for PoS
		Number:      blockHeight,
		GasLimit:    blockGasLimit,
		GasUsed:     n.calculateGasUsed(transactions),
		Time:        uint64(time.Now().Unix()), // Add current timestamp
		Extra:       []byte("Quantum-Fast"),    // Extra data
		MixDigest:   types.ZeroHash,           // Not used in PoS
		Nonce:       0,                        // Not used in PoS
	}, transactions, nil)
	
	// Validate block using fast consensus
	err = n.fastConsensus.ValidateBlock(block, n.validatorAddr)
	if err != nil {
		log.Printf("Failed to validate block: %v", err)
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
	
	// Mint block reward in QTM
	blockReward := new(big.Int)
	blockReward.SetString(types.BlockReward, 10)
	err = n.tokenSupply.Mint(n.validatorAddr, blockReward)
	if err != nil {
		log.Printf("Failed to mint block reward: %v", err)
	}
	
	log.Printf("🚀 Fast block #%d: %d tx, %.1f%% load, reward: %s QTM", 
		blockHeight.Uint64(), len(transactions), networkLoad*100, 
		new(big.Int).Div(blockReward, big.NewInt(1e18)).String())
	
	// Broadcast block to peers
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

// GetFastConsensus returns the fast consensus mechanism
func (n *Node) GetFastConsensus() *consensus.FastConsensus {
	return n.fastConsensus
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
func (n *Node) GetValidators() []*consensus.Validator {
	return n.fastConsensus.GetActiveValidators()
}

// GetConsensusInfo returns consensus mechanism information
func (n *Node) GetConsensusInfo() map[string]interface{} {
	return n.fastConsensus.GetConsensusInfo()
}