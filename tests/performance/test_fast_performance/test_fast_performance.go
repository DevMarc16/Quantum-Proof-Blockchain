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

type JSONRPCRequest_fast struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type JSONRPCResponse_fast struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func runFastPerformanceTest() {
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
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	// Create transaction with optimized values
	chainIDInt := big.NewInt(8888)
	nonce := uint64(0)
	toAddr := types.Address{0x74, 0x2d, 0x35, 0xCc, 0x66, 0x71, 0xC0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xD5, 0x81, 0xC0, 0x27, 0xd2, 0xb3, 0xd0, 0x7f}
	value := big.NewInt(1000000000000000000) // 1 QTM (not ETH!)
	gasLimit := uint64(5800)                  // Optimized gas limit
	gasPrice := big.NewInt(1000000)          // Low gas price (1 micro-QTM)
	data := []byte{}

	tx := &types.QuantumTransaction{
		ChainID:   types.NewBigInt(chainIDInt.Int64()),
		Nonce:     nonce,
		To:        &toAddr,
		Value:     types.NewBigInt(value.Int64()),
		Gas:       gasLimit,
		GasPrice:  types.NewBigInt(gasPrice.Int64()),
		Data:      data,
		SigAlg:    crypto.SigAlgDilithium,
		PublicKey: pubKey.Bytes(),
	}

	// Sign transaction
	sigHash := tx.SigningHash()
	qrSig, err := crypto.SignMessage(sigHash[:], crypto.SigAlgDilithium, privKey.Bytes())
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}
	tx.Signature = qrSig.Signature
	
	signingTime := time.Since(start)
	fmt.Printf("✅ Transaction signed in: %v\n", signingTime)
	
	// Verify signature locally
	valid, err := crypto.VerifySignature(sigHash[:], qrSig)
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
	req := JSONRPCRequest_fast{
		JSONRPC: "2.0",
		Method:  "eth_chainId",
		Params:  []interface{}{},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_fast(req)
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
	req := JSONRPCRequest_fast{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []string{txJSON},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_fast(req)
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
	req := JSONRPCRequest_fast{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionByHash",
		Params:  []string{txHash},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_fast(req)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", resp.Error)
	}
	
	return resp.Result, nil
}

func makeRPCRequest_fast(req JSONRPCRequest_fast) (*JSONRPCResponse_fast, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	// Use port 8545 for the blockchain
	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var rpcResp JSONRPCResponse_fast
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, err
	}
	
	return &rpcResp, nil
}

func main() {
	runFastPerformanceTest()
}