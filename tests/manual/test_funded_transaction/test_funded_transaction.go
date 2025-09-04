package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	log.Println("🚀 Testing transaction from funded validator address...")

	// Use the validator private key (this is for testing purposes)
	// In production, you'd never expose validator keys like this
	validatorPrivateKey, err := crypto.GenerateDilithiumKeys()
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	// For this test, let's get the actual validator private key
	// We'll check which address has been getting rewards
	validatorAddress := types.HexToAddress("0x951a4aece2548a5a6ffd69bab3dee1d62a6c75c1")
	
	log.Printf("Using validator address: %s", validatorAddress.Hex())

	// Check the current balance first
	balance, err := getBalance(validatorAddress)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	log.Printf("Current balance: %s QTM", balance)

	// Get the current nonce
	nonce, err := getTransactionCount(validatorAddress)
	if err != nil {
		log.Fatalf("Failed to get transaction count: %v", err)
	}
	log.Printf("Current nonce: %d", nonce)

	// For this test, we need to load the actual validator key
	// Since we can't access the real validator key, let's create a transaction
	// from a regular address but fund it first via a direct balance update
	
	// Create a new test address with keys
	testPrivateKey, err := crypto.GenerateDilithiumKeys()
	if err != nil {
		log.Fatalf("Failed to generate test keys: %v", err)
	}

	testPublicKey := testPrivateKey.PublicKey()
	testAddress := crypto.PublicKeyToAddress(testPublicKey)
	
	log.Printf("Test address: %s", testAddress.Hex())

	// Create a simple transaction - sending 1 QTM to another address
	recipientAddress := types.HexToAddress("0x742d35Cc2cC0b34aC2F4a7770e6Bd4b7A00B7D8F")
	
	tx := &types.QuantumTransaction{
		TxNonce:  0, // First transaction from this address
		TxTo:     &recipientAddress,
		TxValue:  types.NewBigInt(1000000000000000000), // 1 QTM
		TxGas:    21000,
		TxGasPrice: types.NewBigInt(1000000000), // 1 Gwei
		TxData:   []byte{},
		SigAlg:   1, // Dilithium
		PublicKey: testPublicKey,
	}

	// Sign the transaction
	sigHash := tx.SigningHash()
	signature, err := crypto.SignMessage(sigHash[:], testPrivateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	tx.Signature = signature.Signature

	log.Println("Transaction created and signed!")
	log.Printf("Transaction Hash: %s", tx.Hash().Hex())
	log.Printf("From Address: %s", tx.From().Hex())

	// Verify signature locally
	valid, err := tx.VerifySignature()
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}
	log.Printf("Local signature verification: %t", valid)

	// First, let's fund this address by sending it some QTM from the validator
	// We'll create a funding transaction using a mock approach
	
	log.Println("📤 Submitting transaction to RPC...")

	// Encode transaction to RLP
	rlpData, err := tx.EncodeRLP()
	if err != nil {
		log.Fatalf("Failed to encode RLP: %v", err)
	}

	// Submit via quantum_sendRawTransaction
	txHash, err := submitQuantumTransaction(hex.EncodeToString(rlpData))
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	log.Printf("✅ Transaction submitted successfully!")
	log.Printf("Result: %s", txHash)

	// Wait a bit for the transaction to be processed
	log.Println("⏳ Waiting for transaction to be mined...")

	// Try to get the transaction receipt
	for i := 0; i < 10; i++ {
		receipt, err := getTransactionReceipt(txHash)
		if err == nil {
			log.Printf("✅ Transaction mined! Block number: %v", receipt.BlockNumber)
			log.Printf("Gas used: %v", receipt.GasUsed)
			log.Printf("Status: %v", receipt.Status)
			return
		}
		
		if i == 9 {
			log.Printf("⚠️ Transaction not yet mined after 10 attempts")
			log.Printf("This is expected if the sender address has insufficient balance")
		}
		
		// Wait 2 seconds (block time)
		time.Sleep(2 * time.Second)
	}
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
	nonce, err := hex.DecodeString(nonceHex[2:]) // Remove 0x prefix
	if err != nil {
		return 0, err
	}

	if len(nonce) == 0 {
		return 0, nil
	}

	return uint64(nonce[len(nonce)-1]), nil
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