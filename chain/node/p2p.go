package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"quantum-blockchain/chain/types"

	"github.com/gorilla/websocket"
)

// MessageType defines P2P message types
type MessageType uint8

const (
	MsgTypeHandshake MessageType = iota
	MsgTypeBlock
	MsgTypeTransaction
	MsgTypePing
	MsgTypePong
	MsgTypeGetBlocks
	MsgTypeBlocks
)

// P2PMessage represents a P2P network message
type P2PMessage struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
	From      string          `json:"from"`
}

// HandshakeData represents handshake information
type HandshakeData struct {
	Version   uint32 `json:"version"`
	NetworkID uint64 `json:"networkId"`
	NodeID    string `json:"nodeId"`
	Height    uint64 `json:"height"`
}

// Peer represents a connected peer
type Peer struct {
	ID       string          `json:"id"`
	Address  string          `json:"address"`
	Conn     *websocket.Conn `json:"-"`
	NodeInfo *HandshakeData  `json:"nodeInfo"`
	LastSeen time.Time       `json:"lastSeen"`
	mu       sync.Mutex
}

// SendMessage sends a message to the peer
func (p *Peer) SendMessage(msg *P2PMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Conn.WriteJSON(msg)
}

// P2PNetwork manages peer-to-peer networking
type P2PNetwork struct {
	listenAddr     string
	bootstrapPeers []string
	peers          map[string]*Peer
	nodeID         string
	networkID      uint64

	// Message handlers
	messageHandlers map[MessageType]func(*Peer, json.RawMessage)

	// Control
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex

	// Callbacks
	onBlock       func(*types.Block)
	onTransaction func(*types.QuantumTransaction)
}

// NewP2PNetwork creates a new P2P network
func NewP2PNetwork(listenAddr string, bootstrapPeers []string) *P2PNetwork {
	ctx, cancel := context.WithCancel(context.Background())

	p2p := &P2PNetwork{
		listenAddr:      listenAddr,
		bootstrapPeers:  bootstrapPeers,
		peers:           make(map[string]*Peer),
		nodeID:          generateNodeID(),
		networkID:       8888, // Quantum chain network ID
		messageHandlers: make(map[MessageType]func(*Peer, json.RawMessage)),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Register message handlers
	p2p.registerMessageHandlers()

	return p2p
}

func generateNodeID() string {
	return fmt.Sprintf("quantum-node-%d", time.Now().UnixNano())
}

// Start starts the P2P network
func (p2p *P2PNetwork) Start(ctx context.Context) error {
	p2p.mu.Lock()
	defer p2p.mu.Unlock()

	// Start listening for incoming connections
	listener, err := net.Listen("tcp", p2p.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	p2p.listener = listener
	log.Printf("P2P network listening on %s", p2p.listenAddr)

	// Start accepting connections
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		p2p.acceptConnections()
	}()

	// Connect to bootstrap peers
	for _, peerAddr := range p2p.bootstrapPeers {
		p2p.wg.Add(1)
		go func(addr string) {
			defer p2p.wg.Done()
			p2p.connectToPeer(addr)
		}(peerAddr)
	}

	// Start peer maintenance
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		p2p.maintainPeers()
	}()

	return nil
}

// Stop stops the P2P network
func (p2p *P2PNetwork) Stop() {
	p2p.mu.Lock()
	defer p2p.mu.Unlock()

	p2p.cancel()

	if p2p.listener != nil {
		p2p.listener.Close()
	}

	// Close all peer connections
	for _, peer := range p2p.peers {
		peer.Conn.Close()
	}

	p2p.wg.Wait()
	log.Printf("P2P network stopped")
}

func (p2p *P2PNetwork) acceptConnections() {
	for {
		select {
		case <-p2p.ctx.Done():
			return
		default:
		}

		conn, err := p2p.listener.Accept()
		if err != nil {
			if p2p.ctx.Err() != nil {
				return // Shutting down
			}
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Upgrade to WebSocket
		upgrader := websocket.Upgrader{}
		wsConn, err := upgrader.Upgrade(&httpResponseWriter{conn}, &http.Request{}, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			conn.Close()
			continue
		}

		p2p.wg.Add(1)
		go func() {
			defer p2p.wg.Done()
			p2p.handleIncomingConnection(wsConn)
		}()
	}
}

func (p2p *P2PNetwork) handleIncomingConnection(conn *websocket.Conn) {
	defer conn.Close()

	// Wait for handshake
	var msg P2PMessage
	err := conn.ReadJSON(&msg)
	if err != nil {
		log.Printf("Failed to read handshake: %v", err)
		return
	}

	if msg.Type != MsgTypeHandshake {
		log.Printf("Expected handshake message, got %v", msg.Type)
		return
	}

	var handshake HandshakeData
	err = json.Unmarshal(msg.Data, &handshake)
	if err != nil {
		log.Printf("Failed to unmarshal handshake: %v", err)
		return
	}

	// Validate handshake
	if handshake.NetworkID != p2p.networkID {
		log.Printf("Network ID mismatch: expected %d, got %d", p2p.networkID, handshake.NetworkID)
		return
	}

	// Create peer
	peer := &Peer{
		ID:       handshake.NodeID,
		Address:  conn.RemoteAddr().String(),
		Conn:     conn,
		NodeInfo: &handshake,
		LastSeen: time.Now(),
	}

	// Add to peers
	p2p.mu.Lock()
	p2p.peers[peer.ID] = peer
	p2p.mu.Unlock()

	log.Printf("New peer connected: %s", peer.ID)

	// Send our handshake
	ourHandshake := HandshakeData{
		Version:   1,
		NetworkID: p2p.networkID,
		NodeID:    p2p.nodeID,
		Height:    0, // Would get from blockchain
	}

	handshakeData, _ := json.Marshal(ourHandshake)
	responseMsg := &P2PMessage{
		Type:      MsgTypeHandshake,
		Data:      handshakeData,
		Timestamp: time.Now().Unix(),
		From:      p2p.nodeID,
	}

	peer.SendMessage(responseMsg)

	// Handle messages from this peer
	p2p.handlePeerMessages(peer)
}

func (p2p *P2PNetwork) connectToPeer(address string) {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial("ws://"+address, nil)
	if err != nil {
		log.Printf("Failed to connect to peer %s: %v", address, err)
		return
	}

	defer conn.Close()

	// Send handshake
	handshake := HandshakeData{
		Version:   1,
		NetworkID: p2p.networkID,
		NodeID:    p2p.nodeID,
		Height:    0, // Would get from blockchain
	}

	handshakeData, _ := json.Marshal(handshake)
	msg := &P2PMessage{
		Type:      MsgTypeHandshake,
		Data:      handshakeData,
		Timestamp: time.Now().Unix(),
		From:      p2p.nodeID,
	}

	err = conn.WriteJSON(msg)
	if err != nil {
		log.Printf("Failed to send handshake to %s: %v", address, err)
		return
	}

	// Wait for handshake response
	var responseMsg P2PMessage
	err = conn.ReadJSON(&responseMsg)
	if err != nil {
		log.Printf("Failed to read handshake response from %s: %v", address, err)
		return
	}

	if responseMsg.Type != MsgTypeHandshake {
		log.Printf("Expected handshake response from %s", address)
		return
	}

	var peerHandshake HandshakeData
	err = json.Unmarshal(responseMsg.Data, &peerHandshake)
	if err != nil {
		log.Printf("Failed to unmarshal handshake from %s: %v", address, err)
		return
	}

	// Create peer
	peer := &Peer{
		ID:       peerHandshake.NodeID,
		Address:  address,
		Conn:     conn,
		NodeInfo: &peerHandshake,
		LastSeen: time.Now(),
	}

	// Add to peers
	p2p.mu.Lock()
	p2p.peers[peer.ID] = peer
	p2p.mu.Unlock()

	log.Printf("Connected to peer: %s", peer.ID)

	// Handle messages from this peer
	p2p.handlePeerMessages(peer)
}

func (p2p *P2PNetwork) handlePeerMessages(peer *Peer) {
	defer func() {
		// Remove peer on disconnect
		p2p.mu.Lock()
		delete(p2p.peers, peer.ID)
		p2p.mu.Unlock()

		log.Printf("Peer disconnected: %s", peer.ID)
	}()

	for {
		select {
		case <-p2p.ctx.Done():
			return
		default:
		}

		var msg P2PMessage
		err := peer.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Failed to read message from peer %s: %v", peer.ID, err)
			return
		}

		peer.LastSeen = time.Now()

		// Handle message
		p2p.mu.RLock()
		handler, exists := p2p.messageHandlers[msg.Type]
		p2p.mu.RUnlock()

		if exists {
			handler(peer, msg.Data)
		}
	}
}

func (p2p *P2PNetwork) registerMessageHandlers() {
	p2p.messageHandlers[MsgTypePing] = p2p.handlePing
	p2p.messageHandlers[MsgTypePong] = p2p.handlePong
	p2p.messageHandlers[MsgTypeBlock] = p2p.handleBlock
	p2p.messageHandlers[MsgTypeTransaction] = p2p.handleTransaction
}

func (p2p *P2PNetwork) handlePing(peer *Peer, data json.RawMessage) {
	// Respond with pong
	pongMsg := &P2PMessage{
		Type:      MsgTypePong,
		Data:      []byte("{}"),
		Timestamp: time.Now().Unix(),
		From:      p2p.nodeID,
	}

	peer.SendMessage(pongMsg)
}

func (p2p *P2PNetwork) handlePong(peer *Peer, data json.RawMessage) {
	// Just update last seen time (already done)
}

func (p2p *P2PNetwork) handleBlock(peer *Peer, data json.RawMessage) {
	var block types.Block
	err := json.Unmarshal(data, &block)
	if err != nil {
		log.Printf("Failed to unmarshal block from peer %s: %v", peer.ID, err)
		return
	}

	// Forward to block handler if set
	if p2p.onBlock != nil {
		p2p.onBlock(&block)
	}
}

func (p2p *P2PNetwork) handleTransaction(peer *Peer, data json.RawMessage) {
	var tx types.QuantumTransaction
	err := json.Unmarshal(data, &tx)
	if err != nil {
		log.Printf("Failed to unmarshal transaction from peer %s: %v", peer.ID, err)
		return
	}

	// Forward to transaction handler if set
	if p2p.onTransaction != nil {
		p2p.onTransaction(&tx)
	}
}

func (p2p *P2PNetwork) maintainPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p2p.ctx.Done():
			return
		case <-ticker.C:
			p2p.sendPings()
		}
	}
}

func (p2p *P2PNetwork) sendPings() {
	p2p.mu.RLock()
	peers := make([]*Peer, 0, len(p2p.peers))
	for _, peer := range p2p.peers {
		peers = append(peers, peer)
	}
	p2p.mu.RUnlock()

	pingMsg := &P2PMessage{
		Type:      MsgTypePing,
		Data:      []byte("{}"),
		Timestamp: time.Now().Unix(),
		From:      p2p.nodeID,
	}

	for _, peer := range peers {
		peer.SendMessage(pingMsg)
	}
}

// BroadcastBlock broadcasts a block to all peers
func (p2p *P2PNetwork) BroadcastBlock(block *types.Block) {
	blockData, err := json.Marshal(block)
	if err != nil {
		log.Printf("Failed to marshal block for broadcast: %v", err)
		return
	}

	msg := &P2PMessage{
		Type:      MsgTypeBlock,
		Data:      blockData,
		Timestamp: time.Now().Unix(),
		From:      p2p.nodeID,
	}

	p2p.broadcast(msg)
}

// BroadcastTransaction broadcasts a transaction to all peers
func (p2p *P2PNetwork) BroadcastTransaction(tx *types.QuantumTransaction) {
	txData, err := json.Marshal(tx)
	if err != nil {
		log.Printf("Failed to marshal transaction for broadcast: %v", err)
		return
	}

	msg := &P2PMessage{
		Type:      MsgTypeTransaction,
		Data:      txData,
		Timestamp: time.Now().Unix(),
		From:      p2p.nodeID,
	}

	p2p.broadcast(msg)
}

func (p2p *P2PNetwork) broadcast(msg *P2PMessage) {
	p2p.mu.RLock()
	peers := make([]*Peer, 0, len(p2p.peers))
	for _, peer := range p2p.peers {
		peers = append(peers, peer)
	}
	p2p.mu.RUnlock()

	for _, peer := range peers {
		go func(p *Peer) {
			err := p.SendMessage(msg)
			if err != nil {
				log.Printf("Failed to send message to peer %s: %v", p.ID, err)
			}
		}(peer)
	}
}

// SetBlockHandler sets the block message handler
func (p2p *P2PNetwork) SetBlockHandler(handler func(*types.Block)) {
	p2p.onBlock = handler
}

// SetTransactionHandler sets the transaction message handler
func (p2p *P2PNetwork) SetTransactionHandler(handler func(*types.QuantumTransaction)) {
	p2p.onTransaction = handler
}

// GetPeers returns connected peers
func (p2p *P2PNetwork) GetPeers() []*Peer {
	p2p.mu.RLock()
	defer p2p.mu.RUnlock()

	peers := make([]*Peer, 0, len(p2p.peers))
	for _, peer := range p2p.peers {
		peers = append(peers, peer)
	}

	return peers
}

// Simple HTTP response writer for WebSocket upgrade
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
