package unit

import (
	"math/big"
	"testing"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func TestAddressCreation(t *testing.T) {
	// Test address from bytes
	testBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	addr := types.BytesToAddress(testBytes)
	
	if len(addr.Bytes()) != types.AddressLength {
		t.Errorf("Expected address length %d, got %d", types.AddressLength, len(addr.Bytes()))
	}

	// Test address from hex
	hexAddr := "0x1234567890123456789012345678901234567890"
	addr2, err := types.HexToAddress(hexAddr)
	if err != nil {
		t.Fatalf("Failed to create address from hex: %v", err)
	}

	expectedHex := "0x1234567890123456789012345678901234567890"
	if addr2.Hex() != expectedHex {
		t.Errorf("Expected hex %s, got %s", expectedHex, addr2.Hex())
	}

	// Test invalid hex
	_, err = types.HexToAddress("invalid")
	if err == nil {
		t.Error("Should have failed with invalid hex")
	}
}

func TestHashCreation(t *testing.T) {
	// Test hash from bytes
	testBytes := make([]byte, 32)
	for i := range testBytes {
		testBytes[i] = byte(i)
	}
	hash := types.BytesToHash(testBytes)
	
	if len(hash.Bytes()) != types.HashLength {
		t.Errorf("Expected hash length %d, got %d", types.HashLength, len(hash.Bytes()))
	}

	// Test hash from hex
	hexHash := "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	hash2, err := types.HexToHash(hexHash)
	if err != nil {
		t.Fatalf("Failed to create hash from hex: %v", err)
	}

	if hash2.Hex() != hexHash {
		t.Errorf("Expected hex %s, got %s", hexHash, hash2.Hex())
	}
}

func TestPublicKeyToAddress(t *testing.T) {
	// Generate a quantum key pair
	_, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Derive address from public key
	addr := types.PublicKeyToAddress(pubKey.Bytes())
	
	if addr.IsZero() {
		t.Error("Address should not be zero")
	}

	// Test that same public key generates same address
	addr2 := types.PublicKeyToAddress(pubKey.Bytes())
	if !addr.Equal(addr2) {
		t.Error("Same public key should generate same address")
	}

	// Test with different public key
	_, pubKey3, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	addr3 := types.PublicKeyToAddress(pubKey3.Bytes())
	if addr.Equal(addr3) {
		t.Error("Different public keys should generate different addresses")
	}
}

func TestQuantumTransaction(t *testing.T) {
	// Create transaction parameters
	chainID := big.NewInt(8888)
	nonce := uint64(1)
	to := types.BytesToAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei
	data := []byte("Hello, World!")

	// Create transaction
	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)
	
	// Test getters
	if tx.GetChainID().Cmp(chainID) != 0 {
		t.Error("Chain ID mismatch")
	}
	
	if tx.GetNonce() != nonce {
		t.Error("Nonce mismatch")
	}
	
	if tx.GetTo() == nil || !tx.GetTo().Equal(to) {
		t.Error("To address mismatch")
	}
	
	if tx.GetValue().Cmp(value) != 0 {
		t.Error("Value mismatch")
	}
	
	if tx.GetGas() != gasLimit {
		t.Error("Gas limit mismatch")
	}
	
	if tx.GetGasPrice().Cmp(gasPrice) != 0 {
		t.Error("Gas price mismatch")
	}
	
	if string(tx.GetData()) != string(data) {
		t.Error("Data mismatch")
	}

	// Test contract creation
	contractTx := types.NewQuantumTransaction(chainID, nonce, nil, value, gasLimit, gasPrice, data)
	if !contractTx.IsContractCreation() {
		t.Error("Should be a contract creation transaction")
	}
}

func TestTransactionSigning(t *testing.T) {
	// Generate key pair
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create transaction
	chainID := big.NewInt(8888)
	nonce := uint64(1)
	to := types.BytesToAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	value := big.NewInt(1000000000000000000)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000)
	data := []byte{}

	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)
	
	// Sign transaction
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify signature
	valid, err := tx.VerifySignature()
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	
	if !valid {
		t.Error("Signature verification failed")
	}

	// Check that from address is set
	from := tx.From()
	if from.IsZero() {
		t.Error("From address should not be zero after signing")
	}

	// Check that hash is set
	hash := tx.Hash()
	if hash.IsZero() {
		t.Error("Transaction hash should not be zero after signing")
	}

	// Test signing hash consistency
	sigHash1 := tx.SigningHash()
	sigHash2 := tx.SigningHash()
	if !sigHash1.Equal(sigHash2) {
		t.Error("Signing hash should be consistent")
	}
}

func TestTransactionJSON(t *testing.T) {
	// Create and sign transaction
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	chainID := big.NewInt(8888)
	nonce := uint64(1)
	to := types.BytesToAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	value := big.NewInt(1000000000000000000)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000)
	data := []byte("test data")

	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Test JSON marshaling
	jsonData, err := tx.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal transaction to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty")
	}

	// Basic check that it contains expected fields
	jsonStr := string(jsonData)
	expectedFields := []string{"hash", "chainId", "nonce", "gasPrice", "gas", "to", "value", "input", "sigAlg", "signature", "from"}
	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON should contain field: %s", field)
		}
	}
}

func TestBlockCreation(t *testing.T) {
	// Create block header
	parentHash := types.BytesToHash([]byte("parent"))
	coinbase := types.BytesToAddress([]byte("coinbase"))
	root := types.BytesToHash([]byte("state root"))
	number := big.NewInt(1)
	gasLimit := uint64(15000000)
	timestamp := uint64(1234567890)

	header := types.NewBlockHeader(parentHash, coinbase, root, number, gasLimit, timestamp)
	
	// Verify header fields
	if !header.ParentHash.Equal(parentHash) {
		t.Error("Parent hash mismatch")
	}
	
	if !header.Coinbase.Equal(coinbase) {
		t.Error("Coinbase mismatch")
	}
	
	if !header.Root.Equal(root) {
		t.Error("Root mismatch")
	}
	
	if header.Number.Cmp(number) != 0 {
		t.Error("Number mismatch")
	}
	
	if header.GasLimit != gasLimit {
		t.Error("Gas limit mismatch")
	}
	
	if header.Time != timestamp {
		t.Error("Timestamp mismatch")
	}

	// Create block
	transactions := []*types.QuantumTransaction{}
	uncles := []*types.BlockHeader{}
	
	block := types.NewBlock(header, transactions, uncles)
	
	// Verify block
	if block.Number().Cmp(number) != 0 {
		t.Error("Block number mismatch")
	}
	
	if block.Time() != timestamp {
		t.Error("Block timestamp mismatch")
	}
	
	if block.GasLimit() != gasLimit {
		t.Error("Block gas limit mismatch")
	}
	
	if !block.Coinbase().Equal(coinbase) {
		t.Error("Block coinbase mismatch")
	}
	
	if !block.ParentHash().Equal(parentHash) {
		t.Error("Block parent hash mismatch")
	}
}

func TestBlockSigning(t *testing.T) {
	// Generate validator key
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate validator key: %v", err)
	}
	
	validatorAddr := types.BytesToAddress([]byte("validator"))

	// Create block header
	parentHash := types.BytesToHash([]byte("parent"))
	coinbase := validatorAddr
	root := types.BytesToHash([]byte("state root"))
	number := big.NewInt(1)
	gasLimit := uint64(15000000)
	timestamp := uint64(1234567890)

	header := types.NewBlockHeader(parentHash, coinbase, root, number, gasLimit, timestamp)
	
	// Sign block
	err = header.SignBlock(privKey.Bytes(), crypto.SigAlgDilithium, validatorAddr)
	if err != nil {
		t.Fatalf("Failed to sign block: %v", err)
	}

	// Verify signature
	valid, err := header.VerifyValidatorSignature()
	if err != nil {
		t.Fatalf("Failed to verify validator signature: %v", err)
	}
	
	if !valid {
		t.Error("Validator signature verification failed")
	}

	// Check validator address
	if !header.ValidatorAddr.Equal(validatorAddr) {
		t.Error("Validator address mismatch")
	}

	// Check that signature is present
	if header.ValidatorSig == nil {
		t.Error("Validator signature should not be nil")
	}
}

func TestGenesisBlock(t *testing.T) {
	genesis := types.Genesis()
	
	// Verify genesis properties
	if genesis.Number().Cmp(big.NewInt(0)) != 0 {
		t.Error("Genesis block number should be 0")
	}
	
	if !genesis.ParentHash().IsZero() {
		t.Error("Genesis parent hash should be zero")
	}
	
	if genesis.Time() == 0 {
		t.Error("Genesis timestamp should not be zero")
	}
	
	if len(genesis.Transactions) != 0 {
		t.Error("Genesis should have no transactions")
	}
	
	if genesis.GasLimit() == 0 {
		t.Error("Genesis gas limit should not be zero")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsHelper(s[1:len(s)-1], substr))))
}

func containsHelper(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}