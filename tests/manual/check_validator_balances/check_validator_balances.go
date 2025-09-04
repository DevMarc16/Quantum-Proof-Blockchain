package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	ID      int         `json:"id"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func makeRPCCall(url string, method string, params []interface{}) (interface{}, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response RPCResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return response.Result, nil
}

func main() {
	// Actual validator addresses from key files
	validators := map[string]string{
		"Validator 1": "0x7648c1a9d82a6e259e334a3f0bf556df5418c370",
		"Validator 2": "0x0db8c6713ad3305d74014fc0491dc5401851fb77", 
		"Validator 3": "0x3df7edf63d8042d9e79e537e4cf4dbfd2830fe7d",
	}
	
	// RPC endpoints
	endpoints := map[string]string{
		"Validator 1": "http://localhost:8545",
		"Validator 2": "http://localhost:8547",
		"Validator 3": "http://localhost:8549",
	}

	fmt.Println("🔍 Checking Validator Balances and Block Heights")
	fmt.Println("================================================")

	for name, endpoint := range endpoints {
		fmt.Printf("\n%s (%s):\n", name, endpoint)
		
		// Get block height
		blockResult, err := makeRPCCall(endpoint, "eth_blockNumber", []interface{}{})
		if err != nil {
			fmt.Printf("  ❌ Failed to get block height: %v\n", err)
			continue
		}
		
		// Get validator balance
		validatorAddr := validators[name]
		balanceResult, err := makeRPCCall(endpoint, "eth_getBalance", []interface{}{validatorAddr, "latest"})
		if err != nil {
			fmt.Printf("  ❌ Failed to get validator balance: %v\n", err)
			continue
		}
		
		fmt.Printf("  📊 Block Height: %s\n", blockResult)
		fmt.Printf("  💰 Validator Balance: %s QTM (hex)\n", balanceResult)
		fmt.Printf("  📍 Validator Address: %s\n", validatorAddr)
	}
	
	// Also check the old fixed address for comparison
	fmt.Printf("\n🔍 Old Fixed Address (should remain at 0):\n")
	fixedAddr := "0x0911ee379271364e5902be7dc0cc72cd97294ade"
	balanceResult, err := makeRPCCall("http://localhost:8545", "eth_getBalance", []interface{}{fixedAddr, "latest"})
	if err != nil {
		fmt.Printf("  ❌ Failed to get fixed address balance: %v\n", err)
	} else {
		fmt.Printf("  💰 Fixed Address Balance: %s QTM (hex)\n", balanceResult)
		fmt.Printf("  📍 Fixed Address: %s\n", fixedAddr)
	}

	fmt.Println("\n✅ Balance check complete!")
}