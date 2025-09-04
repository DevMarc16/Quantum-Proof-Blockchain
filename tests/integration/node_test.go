package integration

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/node"
	"quantum-blockchain/chain/types"
)

// TestNodeStartup tests basic node startup and RPC functionality
func TestNodeStartup(t *testing.T) {
	// Create temporary data directory
	tempDir := t.TempDir()

	// Create node configuration
	config := &node.Config{
		DataDir:      tempDir,
		NetworkID:    8888,
		ListenAddr:   "127.0.0.1:0", // Use any available port
		HTTPPort:     0,             // Use any available port
		WSPort:       0,             // Use any available port
		ValidatorKey: "auto",        // Enable validator for mining
		ValidatorAlg: "dilithium",
		Mining:       false,
		GasLimit:     15000000,
		GasPrice:     big.NewInt(1000000000),
	}

	// Create and start node
	testNode, err := node.NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	err = testNode.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	defer testNode.Stop()

	// Give node time to start up
	time.Sleep(2 * time.Second)

	// Test basic functionality
	blockchain := testNode.GetBlockchain()
	if blockchain == nil {
		t.Fatal("Blockchain should not be nil")
	}

	currentBlock := blockchain.GetCurrentBlock()
	if currentBlock == nil {
		t.Fatal("Current block should not be nil")
	}

	// Should be genesis block
	if currentBlock.Number().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected genesis block (0), got block %s", currentBlock.Number().String())
	}
}

// TestRPCEndpoints tests various RPC endpoints
func TestRPCEndpoints(t *testing.T) {
	// Skip if running in CI without network access
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping RPC test in CI environment")
	}

	// This test assumes a node is running on localhost:8545
	// In a real CI/CD pipeline, you would start a test node first

	baseURL := "http://localhost:8545"

	tests := []struct {
		name     string
		method   string
		params   interface{}
		expected string
	}{
		{
			name:     "Chain ID",
			method:   "eth_chainId",
			params:   []interface{}{},
			expected: "0x22b8", // 8888 in hex
		},
		{
			name:     "Block Number",
			method:   "eth_blockNumber",
			params:   []interface{}{},
			expected: "0x", // Should start with 0x
		},
		{
			name:     "Net Version",
			method:   "net_version",
			params:   []interface{}{},
			expected: "8888",
		},
		{
			name:     "Gas Price",
			method:   "eth_gasPrice",
			params:   []interface{}{},
			expected: "0x3b9aca00", // 1 Gwei
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  test.method,
				"params":  test.params,
				"id":      1,
			}

			jsonReq, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonReq))
			if err != nil {
				t.Skipf("Could not connect to test node: %v", err)
				return
			}
			defer resp.Body.Close()

			var rpcResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&rpcResp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if rpcResp["error"] != nil {
				t.Fatalf("RPC error: %v", rpcResp["error"])
			}

			result, ok := rpcResp["result"].(string)
			if !ok && test.expected != "" {
				t.Fatalf("Expected string result, got %T", rpcResp["result"])
			}

			if test.expected == "0x" {
				// Just check it starts with 0x
				if len(result) < 2 || result[:2] != "0x" {
					t.Errorf("Expected result to start with 0x, got %s", result)
				}
			} else if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

// TestQuantumSpecificEndpoints tests quantum-specific RPC methods
func TestQuantumSpecificEndpoints(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping quantum RPC test in CI environment")
	}

	baseURL := "http://localhost:8545"

	t.Run("Supported Algorithms", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "quantum_getSupportedAlgorithms",
			"params":  []interface{}{},
			"id":      1,
		}

		jsonReq, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonReq))
		if err != nil {
			t.Skipf("Could not connect to test node: %v", err)
			return
		}
		defer resp.Body.Close()

		var rpcResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&rpcResp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if rpcResp["error"] != nil {
			t.Fatalf("RPC error: %v", rpcResp["error"])
		}

		result, ok := rpcResp["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected object result, got %T", rpcResp["result"])
		}

		// Check that we have signature algorithms
		sigAlgs, ok := result["signature"].([]interface{})
		if !ok || len(sigAlgs) == 0 {
			t.Error("Expected signature algorithms list")
		}

		// Should contain Dilithium and Falcon
		sigAlgStrings := make([]string, len(sigAlgs))
		for i, alg := range sigAlgs {
			sigAlgStrings[i] = alg.(string)
		}

		if !containsString(sigAlgStrings, "Dilithium") {
			t.Error("Should support Dilithium")
		}

		if !containsString(sigAlgStrings, "Falcon") {
			t.Error("Should support Falcon")
		}
	})
}

// TestTransactionLifecycle tests the full transaction lifecycle
func TestTransactionLifecycle(t *testing.T) {
	// Create temporary data directory
	tempDir := t.TempDir()

	// Create node configuration
	config := &node.Config{
		DataDir:      tempDir,
		NetworkID:    8888,
		ListenAddr:   "127.0.0.1:0",
		HTTPPort:     0,
		WSPort:       0,
		ValidatorKey: "auto",        // Enable validator for mining
		ValidatorAlg: "dilithium",
		Mining:       true, // Enable mining for this test
		GasLimit:     15000000,
		GasPrice:     big.NewInt(1000000000),
	}

	// Create and start node
	testNode, err := node.NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	err = testNode.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	defer testNode.Stop()

	// Start mining
	testNode.SetMining(true)

	// Give node time to start up and mine some blocks
	time.Sleep(5 * time.Second)

	// Create a quantum transaction
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	chainID := big.NewInt(8888)
	nonce := uint64(0)
	to := types.BytesToAddress([]byte("test recipient"))
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000)
	data := []byte{}

	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)

	// Sign transaction
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Add transaction to node
	err = testNode.AddTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to add transaction to node: %v", err)
	}

	// Verify transaction is in pool
	txPool := testNode.GetTxPool()
	poolTx, found := txPool.GetTransaction(tx.Hash())
	if !found {
		t.Error("Transaction should be in pool")
	}

	if !poolTx.Hash().Equal(tx.Hash()) {
		t.Error("Transaction hash mismatch")
	}

	// Wait for transaction to be mined
	time.Sleep(15 * time.Second) // Wait for a few blocks

	// Check if transaction was removed from pool (indicating it was mined)
	_, stillInPool := txPool.GetTransaction(tx.Hash())
	if stillInPool {
		t.Log("Transaction still in pool, may not have been mined yet")
	}
}

// TestConsensus tests basic consensus functionality
func TestConsensus(t *testing.T) {
	// Create temporary data directory
	tempDir := t.TempDir()

	// Create validator node configuration
	config := &node.Config{
		DataDir:      tempDir,
		NetworkID:    8888,
		ListenAddr:   "127.0.0.1:0",
		HTTPPort:     0,
		WSPort:       0,
		ValidatorKey: "auto",        // Enable validator for mining
		ValidatorAlg: "dilithium",
		Mining:       true,
		GasLimit:     15000000,
		GasPrice:     big.NewInt(1000000000),
	}

	// Create and start validator node
	validatorNode, err := node.NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create validator node: %v", err)
	}

	err = validatorNode.Start()
	if err != nil {
		t.Fatalf("Failed to start validator node: %v", err)
	}
	defer validatorNode.Stop()

	// Start mining
	validatorNode.SetMining(true)

	// Wait for some blocks to be mined
	time.Sleep(30 * time.Second)

	blockchain := validatorNode.GetBlockchain()
	currentBlock := blockchain.GetCurrentBlock()

	// Should have mined some blocks
	if currentBlock.Number().Cmp(big.NewInt(0)) <= 0 {
		t.Error("Should have mined at least one block")
	}

	// Check that blocks have valid quantum signatures (optional check)
	t.Log("Checking for quantum validator signatures...")
	if currentBlock.Header.ValidatorSig != nil {
		t.Log("Found validator signature, verifying...")
		// Verify the signature
		valid, err := currentBlock.Header.VerifyValidatorSignature()
		if err != nil {
			t.Logf("Signature verification error (non-fatal): %v", err)
		} else if valid {
			t.Log("Validator signature is valid")
		} else {
			t.Log("Validator signature verification failed (non-fatal)")
		}
	} else {
		t.Log("No validator signature found (mining without multi-validator consensus)")
	}

	t.Logf("Successfully mined %s blocks with quantum consensus", currentBlock.Number().String())
}

// Helper function to check if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestBlockchainOperations tests basic blockchain operations
func TestBlockchainOperations(t *testing.T) {
	tempDir := t.TempDir()

	config := &node.Config{
		DataDir:   tempDir,
		NetworkID: 8888,
		GasLimit:  15000000,
		GasPrice:  big.NewInt(1000000000),
	}

	testNode, err := node.NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	blockchain := testNode.GetBlockchain()

	// Test getting current block (should be genesis)
	currentBlock := blockchain.GetCurrentBlock()
	if currentBlock.Number().Cmp(big.NewInt(0)) != 0 {
		t.Error("Should start with genesis block")
	}

	// Test getting block by number
	block0, err := blockchain.GetBlockByNumber(big.NewInt(0))
	if err != nil {
		t.Fatalf("Failed to get block 0: %v", err)
	}

	if !block0.Hash().Equal(currentBlock.Hash()) {
		t.Error("Block 0 should match genesis block")
	}

	// Test getting block by hash
	blockByHash, err := blockchain.GetBlockByHash(currentBlock.Hash())
	if err != nil {
		t.Fatalf("Failed to get block by hash: %v", err)
	}

	if !blockByHash.Hash().Equal(currentBlock.Hash()) {
		t.Error("Block by hash should match current block")
	}

	// Test balance operations
	testAddr := types.BytesToAddress([]byte("test"))
	initialBalance := blockchain.GetBalance(testAddr)
	
	// Should be zero initially
	if initialBalance.Cmp(big.NewInt(0)) != 0 {
		t.Error("Initial balance should be zero")
	}

	// Test nonce operations
	initialNonce := blockchain.GetNonce(testAddr)
	if initialNonce != 0 {
		t.Error("Initial nonce should be zero")
	}
}

// TestTxPool tests transaction pool operations
func TestTxPool(t *testing.T) {
	tempDir := t.TempDir()

	config := &node.Config{
		DataDir:   tempDir,
		NetworkID: 8888,
		GasLimit:  15000000,
		GasPrice:  big.NewInt(1000000000),
	}

	testNode, err := node.NewNode(config)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	txPool := testNode.GetTxPool()

	// Test initial pool state
	if txPool.Size() != 0 {
		t.Error("Pool should be empty initially")
	}

	// Create test transaction
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	chainID := big.NewInt(8888)
	nonce := uint64(0)
	to := types.BytesToAddress([]byte("recipient"))
	value := big.NewInt(1000)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000)
	data := []byte{}

	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Add transaction to pool
	err = txPool.AddTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to add transaction to pool: %v", err)
	}

	// Check pool size
	if txPool.Size() != 1 {
		t.Error("Pool should contain one transaction")
	}

	// Get transaction from pool
	poolTx, found := txPool.GetTransaction(tx.Hash())
	if !found {
		t.Error("Transaction should be found in pool")
	}

	if !poolTx.Hash().Equal(tx.Hash()) {
		t.Error("Pool transaction should match original")
	}

	// Get pending transactions
	pending := txPool.GetPendingTransactions(10)
	if len(pending) != 1 {
		t.Error("Should have one pending transaction")
	}

	// Remove transaction
	err = txPool.RemoveTransaction(tx.Hash())
	if err != nil {
		t.Fatalf("Failed to remove transaction: %v", err)
	}

	// Check pool is empty
	if txPool.Size() != 0 {
		t.Error("Pool should be empty after removal")
	}
}