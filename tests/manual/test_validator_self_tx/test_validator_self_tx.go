package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	fmt.Println("Testing validator self-transaction...")

	// Read the actual validator key from the running validator
	keyHex, err := ioutil.ReadFile("test-validator/validator.key")
	if err != nil {
		log.Fatalf("Failed to read validator key: %v", err)
	}

	// Decode hex string
	keyData := make([]byte, len(keyHex)/2)
	for i := 0; i < len(keyHex)/2; i++ {
		fmt.Sscanf(string(keyHex[i*2:i*2+2]), "%02x", &keyData[i])
	}

	// Parse the key
	privKey, err := crypto.DilithiumPrivateKeyFromBytes(keyData)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Get public key and address
	pubKey := privKey.Public()
	validatorAddr := types.PublicKeyToAddress(pubKey.Bytes())

	fmt.Printf("Validator address: %s\n", validatorAddr.Hex())

	// Check current balance
	balance := getBalance(validatorAddr.Hex())
	fmt.Printf("Current balance: %s wei\n", balance)

	// Create transaction - validator sends to itself
	chainID := big.NewInt(8888)
	nonce := uint64(0)
	recipientAddr := validatorAddr // Send to self
	value := big.NewInt(1e18)      // 1 QTM
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei

	tx := &types.QuantumTransaction{
		ChainID:   types.NewBigInt(chainID.Int64()),
		Nonce:     nonce,
		To:        &recipientAddr,
		Value:     types.NewBigInt(value.Int64()),
		Gas:       gasLimit,
		GasPrice:  types.NewBigInt(gasPrice.Int64()),
		Data:      []byte{},
		SigAlg:    crypto.SigAlgDilithium,
		PublicKey: pubKey.Bytes(),
	}

	// Sign the transaction
	sigHash := tx.SigningHash()
	qrSig, err := crypto.SignMessage(sigHash[:], crypto.SigAlgDilithium, privKey.Bytes())
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}
	tx.Signature = qrSig.Signature

	fmt.Printf("Transaction created:\n")
	fmt.Printf("  From: %s\n", validatorAddr.Hex())
	fmt.Printf("  To: %s\n", recipientAddr.Hex())
	fmt.Printf("  Value: 1 QTM\n")
	fmt.Printf("  Nonce: %d\n", nonce)
	fmt.Printf("  Hash: %s\n", tx.Hash().Hex())

	// Marshal to JSON for RPC submission
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}

	// Convert to hex string for eth_sendRawTransaction
	txHex := fmt.Sprintf("0x%x", txJSON)

	// Submit using eth_sendRawTransaction
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_sendRawTransaction",
		"params":  []string{txHex},
		"id":      1,
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Response: %s\n", string(body))
	
	// Parse response to get transaction hash
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if resultData, ok := result["result"].(string); ok {
		fmt.Printf("\nTransaction submitted successfully!\n")
		fmt.Printf("Transaction hash: %s\n", resultData)
		fmt.Printf("\nWait a few seconds and check the transaction receipt.\n")
	}
}

func getBalance(address string) string {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []string{address, "latest"},
		"id":      1,
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	if balance, ok := result["result"].(string); ok {
		// Convert hex to decimal
		b := new(big.Int)
		b.SetString(balance[2:], 16)
		return b.String()
	}
	return "0"
}