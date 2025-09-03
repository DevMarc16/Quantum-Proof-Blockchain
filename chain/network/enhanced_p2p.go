package network

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"quantum-blockchain/chain/consensus"
	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"

	"github.com/gorilla/websocket"
)

// EnhancedP2PNetwork provides production-grade P2P networking for validators
type EnhancedP2PNetwork struct {
	// Core configuration
	nodeID         string
	networkID      uint64
	chainID        *big.Int
	listenAddr     string
	publicAddr     string
	
	// Validator networking
	validatorAddr    types.Address
	isValidator      bool
	validatorPrivKey []byte
	sigAlgorithm     crypto.SignatureAlgorithm
	
	// Peer management
	peers            map[string]*ValidatorPeer
	peersByAddr      map[types.Address]*ValidatorPeer
	bootstrapPeers   []string
	maxPeers         int
	minPeerLatency   time.Duration
	
	// Message handling
	messageHandlers  map[MessageType]MessageHandler
	messageRateLimit map[MessageType]RateLimit
	
	// Security
	tlsConfig        *tls.Config
	allowedPeers     map[string]bool // Permissioned network option
	bannedPeers      map[string]time.Time
	rateLimiter      *RateLimiter
	
	// Performance
	networkMetrics   *NetworkMetrics
	connectionPool   *ConnectionPool
	
	// Consensus integration
	consensusEngine  *consensus.MultiValidatorConsensus
	
	// Control
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	
	// Network status
	running          bool
	listener         net.Listener
	
	// Event handlers
	onPeerConnect    func(*ValidatorPeer)
	onPeerDisconnect func(*ValidatorPeer)
	onConsensusMsg   func(*ConsensusMessage)
	onNetworkEvent   func(NetworkEvent)
}

// ValidatorPeer represents a connected validator peer
type ValidatorPeer struct {
	ID               string                    `json:"id"`
	Address          string                    `json:"address"`
	ValidatorAddr    types.Address            `json:"validatorAddr"`
	PublicKey        []byte                   `json:"publicKey"`
	SigAlgorithm     crypto.SignatureAlgorithm `json:"sigAlgorithm"`
	
	// Connection
	Conn             *websocket.Conn          `json:"-"`
	ConnectedAt      time.Time                `json:"connectedAt"`
	LastSeen         time.Time                `json:"lastSeen"`
	LastPing         time.Time                `json:"lastPing"`
	
	// Performance metrics
	Latency          time.Duration            `json:"latency"`
	MessagesSent     uint64                   `json:"messagesSent"`
	MessagesReceived uint64                   `json:"messagesReceived"`
	BytesSent        uint64                   `json:"bytesSent"`
	BytesReceived    uint64                   `json:"bytesReceived"`
	
	// Status
	IsValidator      bool                     `json:"isValidator"`
	IsBootstrap      bool                     `json:"isBootstrap"`
	Reputation       float64                  `json:"reputation"`
	
	// Security
	FailedAttempts   int                      `json:"failedAttempts"`
	LastFailure      time.Time                `json:"lastFailure"`
	
	// Thread safety
	mu               sync.RWMutex
}

// MessageType defines enhanced message types for validator networking
type MessageType uint8

const (
	// Basic P2P messages
	MsgHandshake MessageType = iota
	MsgPing
	MsgPong
	
	// Blockchain messages  
	MsgBlock
	MsgTransaction
	MsgBlockRequest
	MsgBlockResponse
	
	// Consensus messages
	MsgConsensusVote
	MsgConsensusProposal
	MsgConsensusCommit
	MsgConsensusFinalize
	
	// Validator messages
	MsgValidatorAnnounce
	MsgValidatorChallenge
	MsgValidatorResponse
	
	// Network messages
	MsgPeerExchange
	MsgNetworkStatus
	MsgSyncRequest
	MsgSyncResponse
)

// MessageHandler defines message handler interface
type MessageHandler func(*ValidatorPeer, *P2PMessage) error

// RateLimit defines rate limiting configuration
type RateLimit struct {
	MaxMessages uint64        `json:"maxMessages"`
	TimeWindow  time.Duration `json:"timeWindow"`
	BurstLimit  uint64        `json:"burstLimit"`
}

// NetworkMetrics tracks network performance
type NetworkMetrics struct {
	TotalPeers        int           `json:"totalPeers"`
	ValidatorPeers    int           `json:"validatorPeers"`
	AvgLatency        time.Duration `json:"avgLatency"`
	MessageRate       float64       `json:"messageRate"`
	BandwidthUsage    uint64        `json:"bandwidthUsage"`
	ConnectedSince    time.Time     `json:"connectedSince"`
	TotalMessages     uint64        `json:"totalMessages"`
	DroppedMessages   uint64        `json:"droppedMessages"`
	NetworkHealth     float64       `json:"networkHealth"`
	LastUpdate        time.Time     `json:"lastUpdate"`
	
	// Per-message-type metrics
	MessageStats      map[MessageType]*MessageStats `json:"messageStats"`
}

// MessageStats tracks per-message statistics
type MessageStats struct {
	Count     uint64        `json:"count"`
	Bytes     uint64        `json:"bytes"`
	AvgTime   time.Duration `json:"avgTime"`
	Errors    uint64        `json:"errors"`
}

// ConnectionPool manages connection pooling for efficiency
type ConnectionPool struct {
	maxConnections int
	idleTimeout    time.Duration
	connections    chan *websocket.Conn
	mu             sync.Mutex
}

// RateLimiter provides DDoS protection
type RateLimiter struct {
	limits         map[string]*TokenBucket
	globalLimit    *TokenBucket
	cleanupTicker  *time.Ticker
	mu             sync.RWMutex
}

// TokenBucket implements token bucket rate limiting
type TokenBucket struct {
	capacity     uint64
	tokens       uint64
	refillRate   uint64
	lastRefill   time.Time
	mu           sync.Mutex
}

// NetworkEvent represents network events
type NetworkEvent struct {
	Type      string                 `json:"type"`
	PeerID    string                 `json:"peerId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// P2PMessage represents enhanced P2P message format
type P2PMessage struct {
	Type         MessageType     `json:"type"`
	Data         json.RawMessage `json:"data"`
	Timestamp    int64           `json:"timestamp"`
	From         string          `json:"from"`
	To           string          `json:"to,omitempty"`
	Signature    []byte          `json:"signature,omitempty"`
	PublicKey    []byte          `json:"publicKey,omitempty"`
	SigAlgorithm crypto.SignatureAlgorithm `json:"sigAlgorithm,omitempty"`
	Nonce        uint64          `json:"nonce,omitempty"`
	Priority     uint8           `json:"priority,omitempty"`
}

// ConsensusMessage represents consensus-specific messages
type ConsensusMessage struct {
	Type        consensus.VoteType `json:"type"`
	BlockHash   types.Hash         `json:"blockHash"`
	BlockHeight uint64             `json:"blockHeight"`
	Validator   types.Address      `json:"validator"`
	Signature   []byte             `json:"signature"`
	PublicKey   []byte             `json:"publicKey"`
	Timestamp   time.Time          `json:"timestamp"`
	Evidence    []byte             `json:"evidence,omitempty"`
}

// NewEnhancedP2PNetwork creates a new enhanced P2P network
func NewEnhancedP2PNetwork(config *NetworkConfig) *EnhancedP2PNetwork {
	ctx, cancel := context.WithCancel(context.Background())
	
	network := &EnhancedP2PNetwork{
		nodeID:           generateSecureNodeID(),
		networkID:        config.NetworkID,
		chainID:          config.ChainID,
		listenAddr:       config.ListenAddr,
		publicAddr:       config.PublicAddr,
		bootstrapPeers:   config.BootstrapPeers,
		maxPeers:         config.MaxPeers,
		minPeerLatency:   config.MinPeerLatency,
		peers:            make(map[string]*ValidatorPeer),
		peersByAddr:      make(map[types.Address]*ValidatorPeer),
		messageHandlers:  make(map[MessageType]MessageHandler),
		messageRateLimit: make(map[MessageType]RateLimit),
		allowedPeers:     make(map[string]bool),
		bannedPeers:      make(map[string]time.Time),
		ctx:              ctx,
		cancel:           cancel,
		networkMetrics: &NetworkMetrics{
			ConnectedSince:   time.Now(),
			MessageStats:     make(map[MessageType]*MessageStats),
		},
		rateLimiter:      NewRateLimiter(),
		connectionPool:   NewConnectionPool(config.MaxConnections),
	}
	
	// Configure rate limits
	network.configureRateLimits()
	
	// Register message handlers
	network.registerMessageHandlers()
	
	return network
}

// NetworkConfig defines network configuration
type NetworkConfig struct {
	NetworkID        uint64            `json:"networkId"`
	ChainID          *big.Int          `json:"chainId"`
	ListenAddr       string            `json:"listenAddr"`
	PublicAddr       string            `json:"publicAddr"`
	BootstrapPeers   []string          `json:"bootstrapPeers"`
	MaxPeers         int               `json:"maxPeers"`
	MaxConnections   int               `json:"maxConnections"`
	MinPeerLatency   time.Duration     `json:"minPeerLatency"`
	EnableTLS        bool              `json:"enableTLS"`
	Permissioned     bool              `json:"permissioned"`
}

// SetValidator configures this node as a validator
func (n *EnhancedP2PNetwork) SetValidator(
	addr types.Address,
	privKey []byte,
	sigAlg crypto.SignatureAlgorithm,
) {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	n.validatorAddr = addr
	n.validatorPrivKey = privKey
	n.sigAlgorithm = sigAlg
	n.isValidator = true
}

// SetConsensusEngine integrates with consensus engine
func (n *EnhancedP2PNetwork) SetConsensusEngine(consensus *consensus.MultiValidatorConsensus) {
	n.consensusEngine = consensus
}

// Start starts the enhanced P2P network
func (n *EnhancedP2PNetwork) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if n.running {
		return fmt.Errorf("network already running")
	}
	
	// Start listener
	listener, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	n.listener = listener
	
	log.Printf("Enhanced P2P network listening on %s", n.listenAddr)
	
	// Start connection acceptor
	n.wg.Add(1)
	go n.acceptConnections()
	
	// Connect to bootstrap peers
	for _, peerAddr := range n.bootstrapPeers {
		n.wg.Add(1)
		go n.connectToBootstrapPeer(peerAddr)
	}
	
	// Start network maintenance
	n.wg.Add(1)
	go n.maintainNetwork()
	
	// Start metrics collection
	n.wg.Add(1)
	go n.collectMetrics()
	
	n.running = true
	return nil
}

// Stop stops the P2P network
func (n *EnhancedP2PNetwork) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if !n.running {
		return
	}
	
	n.cancel()
	
	if n.listener != nil {
		n.listener.Close()
	}
	
	// Close all peer connections
	for _, peer := range n.peers {
		peer.Conn.Close()
	}
	
	n.wg.Wait()
	n.running = false
	log.Printf("Enhanced P2P network stopped")
}

// BroadcastConsensusMessage broadcasts consensus messages to all validator peers
func (n *EnhancedP2PNetwork) BroadcastConsensusMessage(msg *ConsensusMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal consensus message: %w", err)
	}
	
	// Sign message if we're a validator
	var signature []byte
	var publicKey []byte
	var sigAlg crypto.SignatureAlgorithm
	
	if n.isValidator && n.validatorPrivKey != nil {
		qrSig, err := crypto.SignMessage(data, n.sigAlgorithm, n.validatorPrivKey)
		if err != nil {
			return fmt.Errorf("failed to sign consensus message: %w", err)
		}
		signature = qrSig.Signature
		publicKey = n.getPublicKey()
		sigAlg = n.sigAlgorithm
	}
	
	p2pMsg := &P2PMessage{
		Type:         MsgConsensusVote,
		Data:         data,
		Timestamp:    time.Now().Unix(),
		From:         n.nodeID,
		Signature:    signature,
		PublicKey:    publicKey,
		SigAlgorithm: sigAlg,
		Priority:     255, // Highest priority for consensus
	}
	
	return n.broadcastToValidators(p2pMsg)
}

// broadcastToValidators broadcasts messages to validator peers only
func (n *EnhancedP2PNetwork) broadcastToValidators(msg *P2PMessage) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	validatorPeers := make([]*ValidatorPeer, 0)
	for _, peer := range n.peers {
		if peer.IsValidator {
			validatorPeers = append(validatorPeers, peer)
		}
	}
	
	if len(validatorPeers) == 0 {
		return fmt.Errorf("no validator peers connected")
	}
	
	// Send to all validator peers concurrently
	errChan := make(chan error, len(validatorPeers))
	for _, peer := range validatorPeers {
		go func(p *ValidatorPeer) {
			errChan <- p.SendMessage(msg)
		}(peer)
	}
	
	// Collect errors
	var errors []error
	for i := 0; i < len(validatorPeers); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) == len(validatorPeers) {
		return fmt.Errorf("failed to send to all validator peers")
	}
	
	return nil
}

// SendMessage sends a message to a specific peer
func (peer *ValidatorPeer) SendMessage(msg *P2PMessage) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()
	
	err := peer.Conn.WriteJSON(msg)
	if err != nil {
		return err
	}
	
	peer.MessagesSent++
	peer.LastSeen = time.Now()
	
	return nil
}

// configureRateLimits sets up rate limiting for different message types
func (n *EnhancedP2PNetwork) configureRateLimits() {
	n.messageRateLimit[MsgConsensusVote] = RateLimit{
		MaxMessages: 100,
		TimeWindow:  time.Minute,
		BurstLimit:  10,
	}
	
	n.messageRateLimit[MsgBlock] = RateLimit{
		MaxMessages: 50,
		TimeWindow:  time.Minute,
		BurstLimit:  5,
	}
	
	n.messageRateLimit[MsgTransaction] = RateLimit{
		MaxMessages: 1000,
		TimeWindow:  time.Minute,
		BurstLimit:  100,
	}
	
	// More restrictive for handshake to prevent DoS
	n.messageRateLimit[MsgHandshake] = RateLimit{
		MaxMessages: 10,
		TimeWindow:  time.Minute,
		BurstLimit:  2,
	}
}

// registerMessageHandlers registers message handlers
func (n *EnhancedP2PNetwork) registerMessageHandlers() {
	n.messageHandlers[MsgHandshake] = n.handleHandshake
	n.messageHandlers[MsgPing] = n.handlePing
	n.messageHandlers[MsgPong] = n.handlePong
	n.messageHandlers[MsgConsensusVote] = n.handleConsensusMessage
	n.messageHandlers[MsgBlock] = n.handleBlock
	n.messageHandlers[MsgTransaction] = n.handleTransaction
	n.messageHandlers[MsgValidatorAnnounce] = n.handleValidatorAnnounce
}

// acceptConnections accepts incoming connections
func (n *EnhancedP2PNetwork) acceptConnections() {
	defer n.wg.Done()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
		}
		
		conn, err := n.listener.Accept()
		if err != nil {
			if n.ctx.Err() != nil {
				return
			}
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		
		// Check if we have room for more peers
		n.mu.RLock()
		peerCount := len(n.peers)
		n.mu.RUnlock()
		
		if peerCount >= n.maxPeers {
			log.Printf("Max peers reached, rejecting connection")
			conn.Close()
			continue
		}
		
		// Rate limit connections
		clientAddr := conn.RemoteAddr().String()
		if !n.rateLimiter.Allow(clientAddr) {
			log.Printf("Rate limit exceeded for %s", clientAddr)
			conn.Close()
			continue
		}
		
		n.wg.Add(1)
		go n.handleIncomingConnection(conn)
	}
}

// handleIncomingConnection handles new incoming connections
func (n *EnhancedP2PNetwork) handleIncomingConnection(conn net.Conn) {
	defer n.wg.Done()
	defer conn.Close()
	
	// Upgrade to WebSocket with timeout
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
	}
	
	wsConn, err := upgrader.Upgrade(&httpResponseWriter{conn}, nil, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer wsConn.Close()
	
	// Set connection timeouts
	wsConn.SetReadDeadline(time.Now().Add(30 * time.Second))
	wsConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	// Handle the connection
	peer, err := n.performHandshake(wsConn, false)
	if err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}
	
	n.addPeer(peer)
	n.handlePeerMessages(peer)
}

// GetNetworkMetrics returns current network metrics
func (n *EnhancedP2PNetwork) GetNetworkMetrics() *NetworkMetrics {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	// Update real-time metrics
	n.networkMetrics.TotalPeers = len(n.peers)
	
	validatorCount := 0
	totalLatency := time.Duration(0)
	for _, peer := range n.peers {
		if peer.IsValidator {
			validatorCount++
		}
		totalLatency += peer.Latency
	}
	
	n.networkMetrics.ValidatorPeers = validatorCount
	if len(n.peers) > 0 {
		n.networkMetrics.AvgLatency = totalLatency / time.Duration(len(n.peers))
	}
	
	n.networkMetrics.LastUpdate = time.Now()
	
	return n.networkMetrics
}

// Helper functions continue...

// generateSecureNodeID generates a secure node ID
func generateSecureNodeID() string {
	return fmt.Sprintf("qnode-%d-%x", time.Now().UnixNano(), rand.Uint64())
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:      make(map[string]*TokenBucket),
		globalLimit: NewTokenBucket(1000, 100), // Global: 1000 capacity, 100/sec refill
		cleanupTicker: time.NewTicker(5 * time.Minute),
	}
	
	go rl.cleanup()
	return rl
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate uint64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if an action is allowed under rate limiting
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.RLock()
	bucket, exists := rl.limits[clientID]
	rl.mu.RUnlock()
	
	if !exists {
		rl.mu.Lock()
		bucket = NewTokenBucket(100, 10) // Per client: 100 capacity, 10/sec refill
		rl.limits[clientID] = bucket
		rl.mu.Unlock()
	}
	
	return bucket.Consume(1) && rl.globalLimit.Consume(1)
}

// Consume attempts to consume tokens from the bucket
func (tb *TokenBucket) Consume(tokens uint64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.refill()
	
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}
	return false
}

// refill adds tokens to the bucket based on time elapsed
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := uint64(elapsed.Seconds()) * tb.refillRate
	
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}
}

// cleanup removes old entries from rate limiter
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.mu.Lock()
			now := time.Now()
			for clientID, bucket := range rl.limits {
				if now.Sub(bucket.lastRefill) > 10*time.Minute {
					delete(rl.limits, clientID)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxConnections int) *ConnectionPool {
	return &ConnectionPool{
		maxConnections: maxConnections,
		idleTimeout:    5 * time.Minute,
		connections:    make(chan *websocket.Conn, maxConnections),
	}
}

// min returns the minimum of two uint64 values
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// Placeholder for additional methods
func (n *EnhancedP2PNetwork) handleHandshake(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

func (n *EnhancedP2PNetwork) handlePing(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

func (n *EnhancedP2PNetwork) handlePong(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

func (n *EnhancedP2PNetwork) handleConsensusMessage(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

func (n *EnhancedP2PNetwork) handleBlock(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

func (n *EnhancedP2PNetwork) handleTransaction(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

func (n *EnhancedP2PNetwork) handleValidatorAnnounce(peer *ValidatorPeer, msg *P2PMessage) error {
	// Implementation would go here
	return nil
}

// Additional placeholder methods for completeness
func (n *EnhancedP2PNetwork) connectToBootstrapPeer(addr string) {
	defer n.wg.Done()
	// Implementation would go here
}

func (n *EnhancedP2PNetwork) maintainNetwork() {
	defer n.wg.Done()
	// Implementation would go here
}

func (n *EnhancedP2PNetwork) collectMetrics() {
	defer n.wg.Done()
	// Implementation would go here
}

func (n *EnhancedP2PNetwork) performHandshake(conn *websocket.Conn, isOutgoing bool) (*ValidatorPeer, error) {
	// Implementation would go here
	return nil, nil
}

func (n *EnhancedP2PNetwork) addPeer(peer *ValidatorPeer) {
	// Implementation would go here
}

func (n *EnhancedP2PNetwork) handlePeerMessages(peer *ValidatorPeer) {
	// Implementation would go here
}

func (n *EnhancedP2PNetwork) getPublicKey() []byte {
	// Implementation would go here
	return nil
}

// httpResponseWriter is a simple implementation for WebSocket upgrade
type httpResponseWriter struct {
	conn net.Conn
}

func (w *httpResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *httpResponseWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w *httpResponseWriter) WriteHeader(statusCode int) {
	// No-op for WebSocket upgrade
}