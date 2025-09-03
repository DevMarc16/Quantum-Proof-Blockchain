package main

import (
	"bytes"
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

func main() {
	fmt.Println("🚀 Testing Fast Quantum Blockchain Performance")
	fmt.Println("=============================================")
	
	// Test 1: Check chain ID
	fmt.Println("\n1. Testing Chain ID...")
	chainID, err := getChainID()
	if err != nil {
		log.Fatal("Failed to get chain ID:", err)
	}
	fmt.Printf("✅ Chain ID: %s (0x%s)\n", chainID, chainID[2:])
	
	// Test 2: Create and submit quantum transaction
	fmt.Println("\n2. Creating quantum transaction...")
	start := time.Now()
	
	// Generate keys
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	// Create transaction with optimized values
	chainIDInt := big.NewInt(8888)
	nonce := uint64(0)
	to, _ := types.HexToAddress("0x742d35Cc6671C0532925a3b8D581C027d2b3d07f")
	value := big.NewInt(1000000000000000000) // 1 QTM (not ETH!)
	gasLimit := uint64(5800)                  // Optimized gas limit
	gasPrice := big.NewInt(1000000)          // Low gas price (1 micro-QTM)
	data := []byte{}

	tx := types.NewQuantumTransaction(chainIDInt, nonce, &to, value, gasLimit, gasPrice, data)

	// Sign transaction
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}
	
	signingTime := time.Since(start)
	fmt.Printf("✅ Transaction signed in: %v\n", signingTime)
	
	// Verify signature locally
	valid, err := tx.VerifySignature()
	if err != nil {
		log.Fatal("Signature verification error:", err)
	}
	if !valid {
		log.Fatal("Invalid signature")
	}
	
	verificationTime := time.Since(start) - signingTime
	fmt.Printf("✅ Signature verified in: %v\n", verificationTime)
	
	// Test 3: Submit transaction
	fmt.Printf("\n3. Submitting to fast blockchain (RPC: 8546)...\n")
	
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatal("JSON marshal error:", err)
	}
	
	submissionStart := time.Now()
	txHash, err := submitTransaction(string(txJSON))
	if err != nil {
		log.Fatal("Failed to submit transaction:", err)
	}
	submissionTime := time.Since(submissionStart)
	
	fmt.Printf("✅ Transaction submitted in: %v\n", submissionTime)
	fmt.Printf("📝 Transaction hash: %s\n", txHash)
	
	// Test 4: Query transaction back
	fmt.Println("\n4. Querying transaction...")
	queryStart := time.Now()
	
	result, err := getTransaction(txHash)
	if err != nil {
		log.Fatal("Failed to query transaction:", err)
	}
	queryTime := time.Since(queryStart)
	
	fmt.Printf("✅ Transaction queried in: %v\n", queryTime)
	
	// Test 5: Performance summary
	fmt.Println("\n🏁 Performance Summary")
	fmt.Println("=====================")
	totalTime := time.Since(start)
	fmt.Printf("⚡ Signing:      %v\n", signingTime)
	fmt.Printf("🔍 Verification: %v\n", verificationTime)
	fmt.Printf("📤 Submission:   %v\n", submissionTime)
	fmt.Printf("📥 Query:        %v\n", queryTime)
	fmt.Printf("🕒 Total:        %v\n", totalTime)
	
	// Show transaction details
	if result != nil {
		fmt.Println("\n📊 Transaction Details:")
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("%s\n", resultJSON)
		
		// Extract key metrics
		if resultMap, ok := result.(map[string]interface{}); ok {
			if gas, ok := resultMap["gas"].(string); ok {
				fmt.Printf("⛽ Gas Limit: %s\n", gas)
			}
			if gasPrice, ok := resultMap["gasPrice"].(string); ok {
				fmt.Printf("💰 Gas Price: %s\n", gasPrice)
			}
			if sigAlg, ok := resultMap["sigAlg"].(float64); ok {
				algName := "Unknown"
				if sigAlg == 1 {
					algName = "Dilithium"
				}
				fmt.Printf("🔐 Signature: %s (Algorithm %v)\n", algName, sigAlg)
			}
		}
	}
	
	fmt.Println("\n✨ Fast quantum blockchain test complete!")
}

func getChainID() (string, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_chainId",
		Params:  []interface{}{},
		ID:      1,
	}
	
	resp, err := makeRPCRequest(req)
	if err != nil {
		return "", err
	}
	
	if resp.Error != nil {
		return "", fmt.Errorf("RPC error: %v", resp.Error)
	}
	
	if result, ok := resp.Result.(string); ok {
		return result, nil
	}
	
	return "", fmt.Errorf("unexpected result type")
}

func submitTransaction(txJSON string) (string, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []string{txJSON},
		ID:      1,
	}
	
	resp, err := makeRPCRequest(req)
	if err != nil {
		return "", err
	}
	
	if resp.Error != nil {
		return "", fmt.Errorf("RPC error: %v", resp.Error)
	}
	
	if result, ok := resp.Result.(string); ok {
		return result, nil
	}
	
	return "", fmt.Errorf("unexpected result type")
}

func getTransaction(txHash string) (interface{}, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionByHash",
		Params:  []string{txHash},
		ID:      1,
	}
	
	resp, err := makeRPCRequest(req)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", resp.Error)
	}
	
	return resp.Result, nil
}

func makeRPCRequest(req JSONRPCRequest) (*JSONRPCResponse, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	// Use port 8547 for the fast blockchain
	resp, err := http.Post("http://localhost:8547", "application/json", bytes.NewBuffer(reqData))
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