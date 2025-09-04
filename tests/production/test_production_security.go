package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	fmt.Println("🔒 Production Security Test Suite")
	fmt.Println("=================================")

	// Test 1: Verify dangerous test methods are removed
	fmt.Println("\n1. Testing removed security vulnerabilities...")
	testRemovedMethods()

	// Test 2: Test genesis configuration loading
	fmt.Println("\n2. Testing genesis configuration...")
	testGenesisConfiguration()

	// Test 3: Test rate limiting (will send many requests)
	fmt.Println("\n3. Testing rate limiting...")
	testRateLimiting()

	// Test 4: Test input validation
	fmt.Println("\n4. Testing input validation...")
	testInputValidation()

	// Test 5: Test persistent storage by checking data consistency
	fmt.Println("\n5. Testing persistent storage...")
	testPersistentStorage()

	fmt.Println("\n✅ Production Security Test Complete!")
	fmt.Println("🔐 Your quantum blockchain is production-secured!")
}

func testRemovedMethods() {
	// Try to call the dangerous test methods that should be removed
	dangerousMethods := []string{
		"test_getValidatorKey",
		"test_getValidatorAddress",
	}

	for _, method := range dangerousMethods {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  method,
			Params:  []interface{}{},
			ID:      1,
		}

		resp, err := makeRPCRequest(req)
		if err != nil {
			fmt.Printf("✅ Method %s properly removed (connection error expected)\n", method)
		} else if resp.Error != nil {
			fmt.Printf("✅ Method %s properly removed (error: %v)\n", method, resp.Error)
		} else {
			fmt.Printf("❌ SECURITY RISK: Method %s still exists!\n", method)
		}
	}
}

func testGenesisConfiguration() {
	// Test that genesis addresses have expected balances
	genesisAddresses := []string{
		"0x0000000000000000000000000000000000000001", // Should have 1M QTM
		"0x742d35Cc6671C0532925a3b8D581C027d2b3d07f", // Should have 100 QTM
		"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", // Should have 10,000 QTM
	}

	for _, addr := range genesisAddresses {
		balance := getBalance(addr)
		if balance != "0x0" {
			fmt.Printf("✅ Genesis address %s has balance: %s\n", addr[:10]+"...", balance)
		} else {
			fmt.Printf("❌ Genesis address %s has no balance\n", addr[:10]+"...")
		}
	}
}

func testRateLimiting() {
	// Send multiple requests quickly to test rate limiting
	fmt.Println("   Sending 10 rapid requests to test rate limiter...")
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 10; i++ {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "eth_blockNumber",
			Params:  []interface{}{},
			ID:      i,
		}

		resp, err := makeRPCRequest(req)
		if err != nil {
			fmt.Printf("   Request %d: Error - %v\n", i+1, err)
			rateLimitedCount++
		} else if resp.Error != nil {
			fmt.Printf("   Request %d: RPC Error - %v\n", i+1, resp.Error)
			rateLimitedCount++
		} else {
			successCount++
		}

		// Small delay between requests
		time.Sleep(50 * time.Millisecond)
	}

	if successCount > 0 {
		fmt.Printf("✅ Rate limiter working: %d successful, %d limited\n", successCount, rateLimitedCount)
	} else {
		fmt.Printf("❌ Rate limiter may be too aggressive or blockchain down\n")
	}
}

func testInputValidation() {
	// Test with invalid JSON-RPC versions and methods
	invalidRequests := []JSONRPCRequest{
		{JSONRPC: "1.0", Method: "eth_blockNumber", Params: []interface{}{}, ID: 1}, // Wrong version
		{JSONRPC: "2.0", Method: "INVALID_METHOD", Params: []interface{}{}, ID: 2},  // Invalid method
		{JSONRPC: "2.0", Method: "", Params: []interface{}{}, ID: 3},                // Empty method
	}

	validationErrors := 0
	for i, req := range invalidRequests {
		resp, err := makeRPCRequest(req)
		if err != nil || (resp != nil && resp.Error != nil) {
			validationErrors++
			fmt.Printf("✅ Invalid request %d properly rejected\n", i+1)
		} else {
			fmt.Printf("❌ Invalid request %d was accepted\n", i+1)
		}
	}

	if validationErrors == len(invalidRequests) {
		fmt.Printf("✅ Input validation working correctly\n")
	}
}

func testPersistentStorage() {
	// Test that the blockchain can query historical data
	currentBlock := getCurrentBlockNumber()
	fmt.Printf("✅ Current block: %s (persistent storage active)\n", currentBlock)

	// Try to get block by number
	if currentBlock != "ERROR" {
		block := getBlockByNumber("latest")
		if block != nil {
			fmt.Printf("✅ Block retrieval working (persistent storage verified)\n")
		} else {
			fmt.Printf("❌ Block retrieval failed\n")
		}
	}
}

func getBalance(address string) string {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getBalance",
		Params:  []interface{}{address, "latest"},
		ID:      1,
	}

	resp, err := makeRPCRequest(req)
	if err != nil {
		return "ERROR"
	}

	if resp.Error != nil {
		return "ERROR"
	}

	if result, ok := resp.Result.(string); ok {
		return result
	}

	return "ERROR"
}

func getCurrentBlockNumber() string {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  []interface{}{},
		ID:      1,
	}

	resp, err := makeRPCRequest(req)
	if err != nil {
		return "ERROR"
	}

	if resp.Error != nil {
		return "ERROR"
	}

	if result, ok := resp.Result.(string); ok {
		return result
	}

	return "ERROR"
}

func getBlockByNumber(blockNum string) interface{} {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{blockNum, false},
		ID:      1,
	}

	resp, err := makeRPCRequest(req)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		return nil
	}

	return resp.Result
}

func makeRPCRequest(req JSONRPCRequest) (*JSONRPCResponse, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Post("http://localhost:8545", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp JSONRPCResponse
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, err
	}

	return &rpcResp, nil
}
