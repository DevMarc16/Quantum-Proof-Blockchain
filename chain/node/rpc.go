package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"

	"github.com/gorilla/websocket"
)

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RPCServer handles JSON-RPC requests
type RPCServer struct {
	node       *Node
	httpServer *http.Server
	wsUpgrader websocket.Upgrader
	httpPort   int
	wsPort     int
	
	// Method handlers
	methods map[string]func(json.RawMessage) (interface{}, error)
}

// NewRPCServer creates a new RPC server
func NewRPCServer(node *Node, httpPort, wsPort int) *RPCServer {
	server := &RPCServer{
		node:     node,
		httpPort: httpPort,
		wsPort:   wsPort,
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		methods: make(map[string]func(json.RawMessage) (interface{}, error)),
	}
	
	server.registerMethods()
	return server
}

// Start starts the RPC server
func (s *RPCServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHTTP)
	mux.HandleFunc("/ws", s.handleWebSocket)
	
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: mux,
	}
	
	log.Printf("Starting RPC server on port %d", s.httpPort)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("RPC server error: %v", err)
		}
	}()
	
	return nil
}

// Stop stops the RPC server
func (s *RPCServer) Stop() {
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
}

func (s *RPCServer) registerMethods() {
	// Blockchain methods
	s.methods["eth_chainId"] = s.ethChainId
	s.methods["eth_blockNumber"] = s.ethBlockNumber
	s.methods["eth_getBalance"] = s.ethGetBalance
	s.methods["eth_getTransactionCount"] = s.ethGetTransactionCount
	s.methods["eth_getBlockByNumber"] = s.ethGetBlockByNumber
	s.methods["eth_getBlockByHash"] = s.ethGetBlockByHash
	s.methods["eth_getTransactionByHash"] = s.ethGetTransactionByHash
	s.methods["eth_getTransactionReceipt"] = s.ethGetTransactionReceipt
	s.methods["eth_sendRawTransaction"] = s.ethSendRawTransaction
	s.methods["eth_gasPrice"] = s.ethGasPrice
	s.methods["eth_estimateGas"] = s.ethEstimateGas
	s.methods["net_version"] = s.netVersion
	s.methods["net_peerCount"] = s.netPeerCount
	
	// Quantum-specific methods
	s.methods["quantum_getSupportedAlgorithms"] = s.quantumGetSupportedAlgorithms
	s.methods["quantum_validateSignature"] = s.quantumValidateSignature
	s.methods["quantum_getValidatorSet"] = s.quantumGetValidatorSet
	
	// Mining methods
	s.methods["miner_start"] = s.minerStart
	s.methods["miner_stop"] = s.minerStop
	s.methods["miner_setEtherbase"] = s.minerSetEtherbase
}

func (s *RPCServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	var req JSONRPCRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.writeError(w, &RPCError{Code: -32700, Message: "Parse error"}, nil)
		return
	}
	
	response := s.handleRequest(&req)
	json.NewEncoder(w).Encode(response)
}

func (s *RPCServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	
	for {
		var req JSONRPCRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		
		response := s.handleRequest(&req)
		err = conn.WriteJSON(response)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

func (s *RPCServer) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	method, exists := s.methods[req.Method]
	if !exists {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		}
	}
	
	result, err := method(req.Params)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32000, Message: err.Error()},
			ID:      req.ID,
		}
	}
	
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

func (s *RPCServer) writeError(w http.ResponseWriter, err *RPCError, id interface{}) {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		Error:   err,
		ID:      id,
	}
	json.NewEncoder(w).Encode(response)
}

// RPC method implementations

func (s *RPCServer) ethChainId(params json.RawMessage) (interface{}, error) {
	return "0x22b8", nil // 8888 in hex
}

func (s *RPCServer) ethBlockNumber(params json.RawMessage) (interface{}, error) {
	currentBlock := s.node.blockchain.GetCurrentBlock()
	return fmt.Sprintf("0x%x", currentBlock.Number().Uint64()), nil
}

func (s *RPCServer) ethGetBalance(params json.RawMessage) (interface{}, error) {
	var p []interface{}
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	addrStr, ok := p[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid address")
	}
	
	addr, err := types.HexToAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}
	
	balance := s.node.blockchain.GetBalance(addr)
	return fmt.Sprintf("0x%x", balance), nil
}

func (s *RPCServer) ethGetTransactionCount(params json.RawMessage) (interface{}, error) {
	var p []interface{}
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	addrStr, ok := p[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid address")
	}
	
	addr, err := types.HexToAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}
	
	nonce := s.node.blockchain.GetNonce(addr)
	return fmt.Sprintf("0x%x", nonce), nil
}

func (s *RPCServer) ethGetBlockByNumber(params json.RawMessage) (interface{}, error) {
	var p []interface{}
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	blockNumStr, ok := p[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid block number")
	}
	
	var blockNum *big.Int
	if blockNumStr == "latest" {
		blockNum = s.node.blockchain.GetCurrentBlock().Number()
	} else {
		if strings.HasPrefix(blockNumStr, "0x") {
			blockNumStr = blockNumStr[2:]
		}
		num, err := strconv.ParseUint(blockNumStr, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid block number format: %w", err)
		}
		blockNum = big.NewInt(int64(num))
	}
	
	block, err := s.node.blockchain.GetBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}
	
	return block, nil
}

func (s *RPCServer) ethGetBlockByHash(params json.RawMessage) (interface{}, error) {
	var p []interface{}
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	hashStr, ok := p[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid hash")
	}
	
	hash, err := types.HexToHash(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %w", err)
	}
	
	block, err := s.node.blockchain.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	
	return block, nil
}

func (s *RPCServer) ethGetTransactionByHash(params json.RawMessage) (interface{}, error) {
	var p []interface{}
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	hashStr, ok := p[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid hash")
	}
	
	hash, err := types.HexToHash(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %w", err)
	}
	
	// Check transaction pool first
	tx, found := s.node.txPool.GetTransaction(hash)
	if found {
		return tx, nil
	}
	
	// TODO: Search in blockchain
	return nil, fmt.Errorf("transaction not found")
}

func (s *RPCServer) ethGetTransactionReceipt(params json.RawMessage) (interface{}, error) {
	var p []interface{}
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	hashStr, ok := p[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid hash")
	}
	
	hash, err := types.HexToHash(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %w", err)
	}
	
	receipt, err := s.node.blockchain.GetTransactionReceipt(hash)
	if err != nil {
		return nil, err
	}
	
	return receipt, nil
}

func (s *RPCServer) ethSendRawTransaction(params json.RawMessage) (interface{}, error) {
	var p []string
	err := json.Unmarshal(params, &p)
	if err != nil || len(p) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	// Decode the raw transaction
	rawTx := p[0]
	tx, err := types.DecodeRLPTransaction([]byte(rawTx))
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}
	
	// Validate the transaction
	if err := s.validateQuantumTransaction(tx); err != nil {
		return nil, fmt.Errorf("invalid transaction: %w", err)
	}
	
	// Add to transaction pool
	if s.node != nil && s.node.txPool != nil {
		if err := s.node.txPool.AddTransaction(tx); err != nil {
			return nil, fmt.Errorf("failed to add transaction to pool: %w", err)
		}
	}
	
	// Return transaction hash
	return tx.Hash().Hex(), nil
}

// validateQuantumTransaction validates a quantum-resistant transaction
func (s *RPCServer) validateQuantumTransaction(tx *types.QuantumTransaction) error {
	// Verify quantum-resistant signature using the signing hash
	sigHash := tx.SigningHash()
	qrSig := &crypto.QRSignature{
		Algorithm: tx.SigAlg,
		Signature: tx.Signature,
		PublicKey: tx.PublicKey,
	}
	
	valid, err := crypto.VerifySignature(sigHash[:], qrSig)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid quantum signature")
	}
	
	// Basic transaction validation
	if tx.ChainID == nil || tx.ChainID.Uint64() != 8888 {
		return fmt.Errorf("invalid chain ID")
	}
	
	if tx.GasPrice == nil || tx.GasPrice.Sign() <= 0 {
		return fmt.Errorf("invalid gas price")
	}
	
	if tx.Gas == 0 {
		return fmt.Errorf("invalid gas limit")
	}
	
	// Validate signature algorithm
	switch tx.SigAlg {
	case crypto.SigAlgDilithium, crypto.SigAlgFalcon:
		// Valid
	default:
		return fmt.Errorf("unsupported signature algorithm: %v", tx.SigAlg)
	}
	
	return nil
}

func (s *RPCServer) ethGasPrice(params json.RawMessage) (interface{}, error) {
	return "0x3b9aca00", nil // 1 Gwei
}

func (s *RPCServer) ethEstimateGas(params json.RawMessage) (interface{}, error) {
	return "0x5208", nil // 21000 gas
}

func (s *RPCServer) netVersion(params json.RawMessage) (interface{}, error) {
	return "8888", nil
}

func (s *RPCServer) netPeerCount(params json.RawMessage) (interface{}, error) {
	peers := s.node.p2p.GetPeers()
	return fmt.Sprintf("0x%x", len(peers)), nil
}

// Quantum-specific methods

func (s *RPCServer) quantumGetSupportedAlgorithms(params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"signature": []string{"Dilithium", "Falcon"},
		"kem":       []string{"Kyber"},
		"hash":      []string{"SHA3-256", "SHA3-512"},
	}, nil
}

func (s *RPCServer) quantumValidateSignature(params json.RawMessage) (interface{}, error) {
	var p map[string]interface{}
	err := json.Unmarshal(params, &p)
	if err != nil {
		return nil, fmt.Errorf("invalid parameters")
	}
	
	// This would validate a quantum signature
	// Implementation would depend on the specific parameters
	return map[string]bool{"valid": true}, nil
}

func (s *RPCServer) quantumGetValidatorSet(params json.RawMessage) (interface{}, error) {
	if s.node.consensus == nil {
		return nil, fmt.Errorf("consensus engine not initialized")
	}
	
	validatorSet := s.node.consensus.GetValidatorSet()
	if validatorSet == nil {
		return nil, fmt.Errorf("validator set not available")
	}
	
	return validatorSet, nil
}

// Mining methods

func (s *RPCServer) minerStart(params json.RawMessage) (interface{}, error) {
	s.node.SetMining(true)
	return true, nil
}

func (s *RPCServer) minerStop(params json.RawMessage) (interface{}, error) {
	s.node.SetMining(false)
	return true, nil
}

func (s *RPCServer) minerSetEtherbase(params json.RawMessage) (interface{}, error) {
	// For now, just return success
	return true, nil
}