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

type JSONRPCRequest_live struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type JSONRPCResponse_live struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func runLiveBlockchainTest() {
	fmt.Println("🚀 Live Quantum Blockchain Demo")
	fmt.Println("===============================")
	fmt.Println("⚡ 2-second blocks | 💰 QTM token | 🔐 Quantum-resistant")
	fmt.Println()

	// Monitor blockchain for 30 seconds
	fmt.Println("📊 Starting blockchain monitor...")
	go monitorBlockchain()

	// Wait a moment for monitoring to start
	time.Sleep(2 * time.Second)

	// Submit multiple transactions
	fmt.Println("\n💸 Submitting quantum transactions...")
	
	for i := 0; i < 5; i++ {
		txHash := submitTestTransaction(uint64(i))
		fmt.Printf("✅ Tx %d submitted: %s\n", i+1, txHash[:16]+"...")
		time.Sleep(500 * time.Millisecond) // 0.5s between transactions
	}

	fmt.Println("\n⏰ Waiting 30 seconds to watch block production...")
	time.Sleep(30 * time.Second)
	
	fmt.Println("\n🏁 Live demo complete!")
}

func monitorBlockchain() {
	var lastBlockNumber int64 = -1
	
	for i := 0; i < 60; i++ { // Monitor for 60 iterations (30 seconds at 0.5s intervals)
		blockNumber := getCurrentBlockNumber()
		
		if blockNumber > lastBlockNumber {
			// New block detected!
			block := getBlockByNumber(fmt.Sprintf("0x%x", blockNumber))
			
			txCount := 0
			if block != nil {
				if blockMap, ok := block.(map[string]interface{}); ok {
					if transactions, ok := blockMap["transactions"].([]interface{}); ok {
						txCount = len(transactions)
					}
				}
			}
			
			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("🎯 [%s] Block #%d mined with %d transactions\n", 
				timestamp, blockNumber, txCount)
			
			lastBlockNumber = blockNumber
		}
		
		time.Sleep(500 * time.Millisecond)
	}
}

func submitTestTransaction(nonce uint64) string {
	// Generate keys
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Printf("Failed to generate keys: %v", err)
		return ""
	}

	// Create transaction
	chainID := big.NewInt(8888)
	toAddr := types.Address{0x74, 0x2d, 0x35, 0xCc, 0x66, 0x71, 0xC0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xD5, 0x81, 0xC0, 0x27, 0xd2, 0xb3, 0xd0, 0x7f}
	value := big.NewInt(1000000000000000000) // 1 QTM
	gasLimit := uint64(5800)
	gasPrice := big.NewInt(1000000) // 0.001 QTM gas price
	data := []byte{}

	tx := &types.QuantumTransaction{
		ChainID:   types.NewBigInt(chainID.Int64()),
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
		log.Printf("Failed to sign transaction: %v", err)
		return ""
	}
	tx.Signature = qrSig.Signature

	// Submit via RPC
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return ""
	}

	req := JSONRPCRequest_live{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []string{string(txJSON)},
		ID:      1,
	}

	resp, err := makeRPCRequest_live(req)
	if err != nil {
		log.Printf("Failed to submit transaction: %v", err)
		return ""
	}

	if resp.Error != nil {
		log.Printf("RPC error: %v", resp.Error)
		return ""
	}

	if result, ok := resp.Result.(string); ok {
		return result
	}

	return ""
}

func getCurrentBlockNumber() int64 {
	req := JSONRPCRequest_live{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  []interface{}{},
		ID:      1,
	}

	resp, err := makeRPCRequest_live(req)
	if err != nil {
		return -1
	}

	if resp.Error != nil {
		return -1
	}

	if result, ok := resp.Result.(string); ok {
		var blockNum int64
		fmt.Sscanf(result, "0x%x", &blockNum)
		return blockNum
	}

	return -1
}

func getBlockByNumber(blockNumber string) interface{} {
	req := JSONRPCRequest_live{
		JSONRPC: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{blockNumber, true},
		ID:      1,
	}

	resp, err := makeRPCRequest_live(req)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		return nil
	}

	return resp.Result
}

func makeRPCRequest_live(req JSONRPCRequest_live) (*JSONRPCResponse_live, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp JSONRPCResponse_live
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, err
	}

	return &rpcResp, nil
}

func main() {
	runLiveBlockchainTest()
}