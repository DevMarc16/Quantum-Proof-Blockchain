package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
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
	fmt.Println("🚀 Deploying QTM Token Contract with Quantum Signatures")
	fmt.Println("=====================================================")

	// Wait for blockchain
	time.Sleep(3 * time.Second)

	// Generate quantum keys for deployment
	fmt.Println("🔐 Generating Dilithium keys for deployment...")
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	privateKeyBytes := privKey.Bytes()
	publicKeyBytes := pubKey.Bytes()

	fmt.Printf("✅ Generated keys - Public: %d bytes, Private: %d bytes\n", len(publicKeyBytes), len(privateKeyBytes))

	// Get nonce for deployment account
	deployerAddr := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

	resp, err := callRPC("eth_getTransactionCount", []interface{}{deployerAddr, "latest"})
	if err != nil {
		log.Fatal("Failed to get nonce:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error getting nonce:", resp.Error.Message)
	}

	nonce := resp.Result.(string)
	fmt.Printf("📊 Current nonce: %s\n", nonce)

	// Simple ERC-20 token bytecode (no constructor parameters)
	qtmBytecode := "608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506040518060400160405280600d81526020017f5175616e74756d20546f6b656e000000000000000000000000000000000000008152506001908161009c9190610275565b506040518060400160405280600381526020017f51544d000000000000000000000000000000000000000000000000000000000081525060029081610d8190610275565b50693a9c12e3f5ede0b90000006003819055507f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e05f8080546108009190610275565b506003548061031c5734801561004857600080fd5b506040518060400160405280600d8152602001600c5175616e74756d20546f6b656e00000000000000000000815250600190805190602001906100f5929190610175565b50620f42406003819055506003548060026000330173ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505061041d56"

	// Convert nonce from hex to uint64
	nonceInt, err := strconv.ParseUint(nonce[2:], 16, 64)
	if err != nil {
		log.Fatal("Failed to parse nonce:", err)
	}

	// Create deployment data
	deploymentDataBytes, err := hex.DecodeString(qtmBytecode)
	if err != nil {
		log.Fatal("Failed to decode deployment data:", err)
	}

	// Create a proper QuantumTransaction struct
	tx := &types.QuantumTransaction{
		ChainID:  big.NewInt(8888), // 0x22b8
		Nonce:    nonceInt,
		GasPrice: big.NewInt(1000000000), // 1 gwei
		Gas:      2000000,                // 2M gas
		To:       nil,                    // Contract creation
		Value:    big.NewInt(0),
		Data:     deploymentDataBytes,
		SigAlg:   crypto.SigAlgDilithium,
	}

	fmt.Printf("🖊️  Signing transaction with Dilithium...\n")

	// Sign the transaction using the proper method
	err = tx.SignTransaction(privateKeyBytes, crypto.SigAlgDilithium)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	fmt.Printf("✅ Transaction signed with %d byte signature\n", len(tx.Signature))

	// Marshall the transaction to JSON for RPC
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		log.Fatal("Failed to marshal transaction:", err)
	}

	fmt.Println("📋 Creating contract deployment transaction...")

	// Hex encode for RPC
	hexTx := fmt.Sprintf("0x%x", txJSON)

	fmt.Printf("📤 Sending deployment transaction (%d bytes)...\n", len(hexTx))

	// Send deployment transaction
	resp, err = callRPC("eth_sendRawTransaction", []interface{}{hexTx})
	if err != nil {
		log.Fatal("Failed to send transaction:", err)
	}

	if resp.Error != nil {
		log.Printf("❌ Deployment failed: %s\n", resp.Error.Message)
		log.Printf("Transaction JSON: %s\n", string(txJSON))
		return
	}

	txHash := resp.Result.(string)
	fmt.Printf("✅ Transaction sent! Hash: %s\n", txHash)

	// Wait for transaction to be mined
	fmt.Print("⏳ Waiting for transaction to be mined")
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		fmt.Print(".")

		resp, err := callRPC("eth_getTransactionReceipt", []interface{}{txHash})
		if err != nil {
			continue
		}

		if resp.Result != nil {
			fmt.Println(" ✅ Mined!")

			receipt := resp.Result.(map[string]interface{})
			if contractAddr, ok := receipt["contractAddress"]; ok && contractAddr != nil {
				fmt.Printf("🎉 QTM Token deployed to: %s\n", contractAddr)

				// Create deployment config
				config := map[string]interface{}{
					"contracts": map[string]interface{}{
						"QTMToken": contractAddr,
					},
					"deployer": deployerAddr,
					"chainId":  "0x22b8",
					"deployment": map[string]interface{}{
						"timestamp": time.Now().Format(time.RFC3339),
						"status":    "deployed-successfully",
						"network":   "quantum-testnet",
					},
				}

				configJSON, _ := json.MarshalIndent(config, "", "  ")
				fmt.Printf("\n📄 Deployment Configuration:\n%s\n", string(configJSON))
				return
			}
		}
	}

	fmt.Println(" ❌ Timeout waiting for transaction")
}
