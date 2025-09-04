package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// The transaction hash from the recent test
	txHash := "0x4e4d8983b1f42131ec4ee7975c090f183f3314be0b4aa5f34d3b9ebf00130552"

	log.Printf("🔍 Checking receipt for transaction: %s", txHash)

	// Try to get the receipt multiple times
	for i := 0; i < 10; i++ {
		receipt, err := getTransactionReceipt(txHash)
		if err != nil {
			log.Printf("⏳ Attempt %d: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("🎉 Transaction receipt found!")
		log.Printf("📋 Receipt details:")
		prettyPrint(receipt)
		return
	}

	log.Printf("❌ Transaction receipt not found after 10 attempts")
	log.Printf("This likely means the transaction wasn't mined due to insufficient balance")

	// Let's also check if the node is still running
	log.Printf("🔍 Checking if node is still responding...")

	chainId, err := getChainId()
	if err != nil {
		log.Printf("❌ Node not responding: %v", err)
	} else {
		log.Printf("✅ Node is responding. Chain ID: %s", chainId)
	}

	// Check current block number
	blockNum, err := getBlockNumber()
	if err != nil {
		log.Printf("❌ Could not get block number: %v", err)
	} else {
		log.Printf("📊 Current block number: %s", blockNum)
	}
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

func getChainId() (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_chainId",
		"params":  []interface{}{},
		"id":      1,
	}

	result, err := makeRPCCall(payload)
	if err != nil {
		return "", err
	}

	return result["result"].(string), nil
}

func getBlockNumber() (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	result, err := makeRPCCall(payload)
	if err != nil {
		return "", err
	}

	return result["result"].(string), nil
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

	if result["result"] == nil {
		return nil, fmt.Errorf("result is null")
	}

	return result, nil
}

func prettyPrint(data interface{}) {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}
