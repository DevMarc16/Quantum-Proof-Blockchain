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
		Mining:         false,
		GasLimit:       15000000,
		GasPrice:       big.NewInt(1000000000), // 1 Gwei
	}
}

// Node represents a quantum-resistant blockchain node
type Node struct {
	config     *Config
	consensus  *consensus.QuantumPoSConsensus
	blockchain *Blockchain
	txPool     *TxPool
	p2p        *P2PNetwork
	rpc        *RPCServer
	
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

// NewNode creates a new node instance
func NewNode(config *Config) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	node := &Node{
		config: config,
		ctx:    ctx,
		cancel: cancel,
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
	
	// Initialize transaction pool
	node.txPool = NewTxPool(1000) // Max 1000 pending transactions
	
	// Initialize consensus
	if node.validatorPrivKey != nil {
		node.consensus = consensus.NewQuantumPoSConsensus(
			node.validatorPrivKey,
			node.validatorAlg,
			node.validatorAddr,
		)
		
		// Create initial validator set with this validator
		validatorInfo := &consensus.ValidatorInfo{
			Address:   node.validatorAddr,
			PublicKey: node.getPublicKey(),
			SigAlg:    node.validatorAlg,
			Stake:     big.NewInt(1000000), // 1M tokens
			LastActive: 0,
			Slashed:   false,
		}
		
		validatorSet := consensus.NewValidatorSet([]*consensus.ValidatorInfo{validatorInfo})
		node.consensus.SetValidatorSet(validatorSet)
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
	
	// Start block production if validator
	if n.consensus != nil && n.config.Mining {
		n.startBlockProduction()
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
		
		ticker := time.NewTicker(n.consensus.GetBlockTime())
		defer ticker.Stop()
		
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				n.produceBlock()
			}
		}
	}()
}

func (n *Node) produceBlock() {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if !n.mining {
		return
	}
	
	// Get current head
	currentBlock := n.blockchain.GetCurrentBlock()
	
	// Get pending transactions
	transactions := n.txPool.GetPendingTransactions(100) // Max 100 transactions per block
	
	// Prepare block
	block, err := n.consensus.PrepareBlock(
		currentBlock,
		transactions,
		n.validatorAddr,
		n.config.GasLimit,
	)
	
	if err != nil {
		log.Printf("Failed to prepare block: %v", err)
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
	
	log.Printf("Produced block #%d with %d transactions", block.Number(), len(transactions))
	
	// Broadcast block to peers
	n.p2p.BroadcastBlock(block)
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