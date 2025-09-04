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

func callRPC(url string, method string, params interface{}) (interface{}, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response JSONRPCResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", response.Error)
	}

	return response.Result, nil
}

func main() {
	fmt.Println("🧪 Testing Transaction Receipt Lookup Functionality")
	fmt.Println("==================================================")

	// Generate quantum keys for transaction
	fmt.Println("🔑 Generating quantum keys...")
	dilithiumPrivKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate Dilithium key:", err)
	}

	// Use the pre-funded quantum address
	fromAddr, _ := types.HexToAddress("0x7889e2f42d63650635ad2987bd3582f7a183e6e9")
	fmt.Printf("📍 Using pre-funded address: %s\n", fromAddr.Hex())

	// Check initial balance
	balance, err := callRPC("http://localhost:8545", "eth_getBalance", []interface{}{fromAddr.Hex(), "latest"})
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}
	fmt.Printf("💰 Current balance: %s\n", balance)

	// Get nonce
	nonceResult, err := callRPC("http://localhost:8545", "eth_getTransactionCount", []interface{}{fromAddr.Hex(), "latest"})
	if err != nil {
		log.Fatal("Failed to get nonce:", err)
	}
	nonceHex, ok := nonceResult.(string)
	if !ok {
		log.Fatal("Invalid nonce format")
	}
	nonce := new(big.Int)
	nonce.SetString(nonceHex[2:], 16)
	fmt.Printf("🔢 Current nonce: %d\n", nonce.Uint64())

	// Create a simple contract deployment transaction
	// Simple contract bytecode that just stores a value
	contractBytecode := "608060405234801561001057600080fd5b506040516020806101168339810180604052810190808051906020019092919050505080600081905550506100c7806100496000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146072575b600080fd5b348015605857600080fd5b50607060048036038101908080359060200190929190505050609a565b005b348015607d57600080fd5b506084609d565b6040518082815260200191505060405180910390f35b80600081905550565b600080549050905600a165627a7a7230582012345678901234567890123456789012345678901234567890123456789012340029"
	bytecode, _ := hex.DecodeString(contractBytecode)

	// Create quantum transaction
	tx := &types.QuantumTransaction{
		Nonce:    nonce.Uint64(),
		GasPrice: big.NewInt(1000000000), // 1 Gwei
		Gas:      2000000,                // High gas limit for contract deployment
		Value:    big.NewInt(0),
		Data:     bytecode,
		ChainID:  big.NewInt(8888),
		SigAlg:   crypto.SigAlgDilithium,
	}

	// Sign the transaction
	sigHash := tx.SigningHash()
	fmt.Printf("🔐 Signing transaction with hash: %s\n", hex.EncodeToString(sigHash[:]))

	qrSig, err := crypto.SignMessage(sigHash[:], crypto.SigAlgDilithium, dilithiumPrivKey.Bytes())
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}
	tx.Signature = qrSig.Signature
	tx.PublicKey = qrSig.PublicKey

	// Encode transaction to JSON
	txBytes, err := tx.MarshalJSON()
	if err != nil {
		log.Fatal("Failed to encode transaction:", err)
	}

	// Submit transaction
	fmt.Println("📤 Submitting contract deployment transaction...")
	txHash, err := callRPC("http://localhost:8545", "eth_sendRawTransaction", []string{"0x" + hex.EncodeToString(txBytes)})
	if err != nil {
		log.Fatal("Failed to submit transaction:", err)
	}

	txHashStr, ok := txHash.(string)
	if !ok {
		log.Fatal("Invalid transaction hash format")
	}

	fmt.Printf("✅ Transaction submitted with hash: %s\n", txHashStr)

	// Wait for transaction to be mined
	fmt.Println("⏳ Waiting for transaction to be mined...")
	for i := 0; i < 30; i++ {
		receipt, err := callRPC("http://localhost:8545", "eth_getTransactionReceipt", []string{txHashStr})
		if err == nil && receipt != nil {
			fmt.Println("🎉 Transaction receipt found!")
			receiptJSON, _ := json.MarshalIndent(receipt, "", "  ")
			fmt.Printf("📋 Receipt details:\n%s\n", string(receiptJSON))

			// Verify receipt contains expected data
			receiptMap, ok := receipt.(map[string]interface{})
			if ok {
				if receiptMap["status"] == "0x1" {
					fmt.Println("✅ Transaction executed successfully!")
				} else {
					fmt.Println("❌ Transaction failed!")
				}

				if contractAddr := receiptMap["contractAddress"]; contractAddr != nil {
					fmt.Printf("🏗️ Contract deployed at: %s\n", contractAddr)
				}

				if gasUsed := receiptMap["gasUsed"]; gasUsed != nil {
					fmt.Printf("⛽ Gas used: %s\n", gasUsed)
				}
			}

			fmt.Println("🎯 Receipt lookup test PASSED!")
			return
		}

		time.Sleep(2 * time.Second)
		fmt.Printf("⏳ Waiting... (attempt %d/30)\n", i+1)
	}

	fmt.Println("❌ Receipt lookup test FAILED - transaction receipt not found after 60 seconds")
}
