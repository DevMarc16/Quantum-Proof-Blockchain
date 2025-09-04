package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("🔍 Checking Genesis Balances...")
	
	// Test addresses from genesis configuration
	testAddresses := []string{
		"0x129b052af5f7858ab578c8c8f244eaac818fa504", // Test address from rpc_submit
		"0x742d35Cc2cC0b34aC2F4a7770e6Bd4b7A00B7D8F", // Common test recipient
		"0x951a4aece2548a5a6ffd69bab3dee1d62a6c75c1", // Old validator address
		"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", // Hardhat test address
		"0x0000000000000000000000000000000000000001", // Original genesis address
	}
	
	for _, addr := range testAddresses {
		balance, err := getBalance(addr)
		if err != nil {
			log.Printf("❌ Error getting balance for %s: %v", addr, err)
			continue
		}
		
		fmt.Printf("💰 %s: %s\n", addr, balance)
	}
}

func getBalance(address string) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance", 
		"params":  []interface{}{address, "latest"},
		"id":      1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if errorObj, exists := result["error"]; exists {
		return "", fmt.Errorf("RPC error: %v", errorObj)
	}
	
	if result["result"] == nil {
		return "0x0", nil
	}

	return result["result"].(string), nil
}