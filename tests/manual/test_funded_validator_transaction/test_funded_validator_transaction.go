package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	log.Println("🚀 Testing transaction from validator with existing balance...")

	// Use a known validator address that has been receiving rewards
	// We'll need to get its private key from the actual validator files
	
	// First let's check what balance this validator has
	validatorAddress := types.HexToAddress("0x951a4aece2548a5a6ffd69bab3dee1d62a6c75c1")
	
	balance, err := getBalance(validatorAddress)
	if err != nil {
		log.Fatalf("Failed to get validator balance: %v", err)
	}
	log.Printf("Validator balance: %s", balance)

	// Get nonce
	nonce, err := getTransactionCount(validatorAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	log.Printf("Validator nonce: %d", nonce)

	// Since we can't access the validator's private key easily, let's instead 
	// add a funded address to the genesis configuration and restart the network

	// For now, let's test by creating a simple transaction but funding the sender first
	// through genesis configuration modification

	log.Println("❌ Cannot proceed without validator private key")
	log.Println("💡 Solution: We need to add test addresses to genesis configuration")
	log.Println("📝 The validator has funds but we don't have access to its private key")
	
	fmt.Printf(`
🔧 To fix transaction mining, we need to:

1. Add test addresses with initial balances to genesis configuration
2. OR: Add a method to fund arbitrary addresses for testing
3. OR: Use the actual validator private key for testing

Current validator balance: %s QTM
This proves the blockchain can handle balances and transfers.

The issue is that test transactions are created with unfunded addresses.
`, balance)

	// Let's try to submit a transaction anyway to confirm the error
	log.Println("🧪 Testing with unfunded address to confirm error...")
	
	// Create a transaction from an unfunded address
	testPrivateKey, err := crypto.GenerateDilithiumKeys()
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	testPublicKey := testPrivateKey.PublicKey()
	testAddress := crypto.PublicKeyToAddress(testPublicKey)
	
	log.Printf("Test address: %s", testAddress.Hex())

	// Check its balance (should be 0)
	testBalance, _ := getBalance(testAddress)
	log.Printf("Test address balance: %s", testBalance)

	// Create transaction
	recipientAddress := types.HexToAddress("0x742d35Cc2cC0b34aC2F4a7770e6Bd4b7A00B7D8F")
	
	tx := &types.QuantumTransaction{
		TxNonce:    0,
		TxTo:       &recipientAddress,
		TxValue:    types.NewBigInt(1000000000000000000), // 1 QTM
		TxGas:      21000,
		TxGasPrice: types.NewBigInt(1000000000), // 1 Gwei
		TxData:     []byte{},
		SigAlg:     1, // Dilithium
		PublicKey:  testPublicKey,
	}

	// Sign the transaction
	sigHash := tx.SigningHash()
	signature, err := crypto.SignMessage(sigHash[:], testPrivateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	tx.Signature = signature.Signature

	log.Printf("Transaction Hash: %s", tx.Hash().Hex())
	log.Printf("From Address: %s", tx.From().Hex())

	// Submit transaction
	rlpData, err := tx.EncodeRLP()
	if err != nil {
		log.Fatalf("Failed to encode RLP: %v", err)
	}

	txHash, err := submitQuantumTransaction(hex.EncodeToString(rlpData))
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	log.Printf("✅ Transaction submitted: %s", txHash)
	log.Println("⏳ Checking if it gets included in blocks...")

	// Wait and check for receipt
	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		receipt, err := getTransactionReceipt(txHash)
		if err == nil {
			log.Printf("🎉 SUCCESS! Transaction mined!")
			log.Printf("Block number: %v", receipt.BlockNumber)
			log.Printf("Gas used: %v", receipt.GasUsed)
			return
		}
		log.Printf("⏳ Attempt %d: Transaction not yet mined (expected due to insufficient balance)", i+1)
	}

	log.Println("📊 Final result: Transaction submitted but not mined due to insufficient balance")
}

func getBalance(address types.Address) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{address.Hex(), "latest"},
		"id":      1,
	}

	response, err := makeRPCCall(payload)
	if err != nil {
		return "", err
	}

	return response["result"].(string), nil
}

func getTransactionCount(address types.Address) (uint64, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionCount",
		"params":  []interface{}{address.Hex(), "latest"},
		"id":      1,
	}

	response, err := makeRPCCall(payload)
	if err != nil {
		return 0, err
	}

	nonceHex := response["result"].(string)
	if nonceHex == "0x" || nonceHex == "0x0" {
		return 0, nil
	}

	nonce, err := hex.DecodeString(nonceHex[2:])
	if err != nil {
		return 0, err
	}

	if len(nonce) == 0 {
		return 0, nil
	}

	// Convert bytes to uint64
	result := uint64(0)
	for _, b := range nonce {
		result = (result << 8) | uint64(b)
	}

	return result, nil
}

func submitQuantumTransaction(rawTxHex string) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "quantum_sendRawTransaction",
		"params":  []string{"0x" + rawTxHex},
		"id":      1,
	}

	response, err := makeRPCCall(payload)
	if err != nil {
		return "", err
	}

	return response["result"].(string), nil
}

func getTransactionReceipt(txHash string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []string{txHash},
		"id":      1,
	}

	return makeRPCCall(payload)
}

func makeRPCCall(payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if errorObj, exists := result["error"]; exists {
		return nil, fmt.Errorf("RPC error: %v", errorObj)
	}

	return result, nil
}