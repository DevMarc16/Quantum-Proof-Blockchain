package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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
	Error   *RPCError   `json:"error"`
	ID      int         `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func callRPC(method string, params []interface{}) (*RPCResponse, error) {
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp RPCResponse
	err = json.Unmarshal(body, &rpcResp)
	if err != nil {
		return nil, err
	}

	return &rpcResp, nil
}

func main() {
	fmt.Println("🧪 Testing quantum blockchain RPC directly")

	// Wait for blockchain to be ready
	time.Sleep(5 * time.Second)

	// Test chain ID
	resp, err := callRPC("eth_chainId", []interface{}{})
	if err != nil {
		log.Fatal("Failed to call eth_chainId:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error:", resp.Error.Message)
	}
	fmt.Printf("Chain ID: %v\n", resp.Result)

	// Test block number
	resp, err = callRPC("eth_blockNumber", []interface{}{})
	if err != nil {
		log.Fatal("Failed to call eth_blockNumber:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error:", resp.Error.Message)
	}
	fmt.Printf("Block number: %v\n", resp.Result)

	// Test balance of deployer account
	resp, err = callRPC("eth_getBalance", []interface{}{"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "latest"})
	if err != nil {
		log.Fatal("Failed to call eth_getBalance:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error:", resp.Error.Message)
	}
	fmt.Printf("Deployer balance: %v\n", resp.Result)

	// Test a simple quantum transaction (JSON format)
	quantumTx := map[string]interface{}{
		"chainId":   "0x22b8", // 8888
		"nonce":     "0x0",
		"to":        "0x742d35Cc6634C0532925a3b8D000B1b000d1b000C",
		"value":     "0xDE0B6B3A7640000", // 1 QTM
		"gasLimit":  "0x5208",            // 21000
		"gasPrice":  "0x3B9ACA00",        // 1 gwei
		"data":      "0x",
		"algorithm": 1,            // Dilithium
		"publicKey": "0x74657374", // "test" in hex
		"signature": "0x74657374", // "test" in hex
	}

	txJSON, err := json.Marshal(quantumTx)
	if err != nil {
		log.Fatal("Failed to marshal transaction:", err)
	}

	fmt.Printf("Quantum transaction JSON: %s\n", string(txJSON))

	// Hex encode for RPC
	hexTx := fmt.Sprintf("0x%x", txJSON)
	fmt.Printf("Hex encoded transaction: %s\n", hexTx)

	// Test sending the transaction
	resp, err = callRPC("eth_sendRawTransaction", []interface{}{hexTx})
	if err != nil {
		log.Fatal("Failed to call eth_sendRawTransaction:", err)
	}
	if resp.Error != nil {
		fmt.Printf("Expected RPC error (no real signature): %s\n", resp.Error.Message)
	} else {
		fmt.Printf("Transaction hash: %v\n", resp.Result)
	}

	fmt.Println("✅ RPC test complete!")
}
