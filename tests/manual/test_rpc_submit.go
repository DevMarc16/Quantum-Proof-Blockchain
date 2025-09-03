package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func main() {
	// Generate keys
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	// Create transaction
	chainID := big.NewInt(8888)
	nonce := uint64(0)
	to, _ := types.HexToAddress("0x742d35Cc6671C0532925a3b8D581C027d2b3d07f")
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei
	data := []byte{}

	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)

	// Sign transaction
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	fmt.Printf("Transaction signed successfully!\n")
	fmt.Printf("Transaction Hash: %s\n", tx.Hash().Hex())
	fmt.Printf("From Address: %s\n", tx.From().Hex())
	
	// Verify signature locally
	valid, err := tx.VerifySignature()
	if err != nil {
		log.Fatal("Signature verification error:", err)
	}
	fmt.Printf("Local signature verification: %t\n", valid)

	// Marshal to JSON (this is what gets sent as raw transaction)
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatal("JSON marshal error:", err)
	}

	// Submit via RPC
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []string{string(txJSON)},
		ID:      1,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		log.Fatal("Request marshal error:", err)
	}

	fmt.Printf("Submitting transaction to RPC...\n")
	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		log.Fatal("HTTP request error:", err)
	}
	defer resp.Body.Close()

	var rpcResp JSONRPCResponse
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		log.Fatal("Response decode error:", err)
	}

	if rpcResp.Error != nil {
		fmt.Printf("RPC Error: %v\n", rpcResp.Error)
	} else {
		fmt.Printf("Transaction submitted successfully!\n")
		fmt.Printf("Result: %v\n", rpcResp.Result)
	}
}