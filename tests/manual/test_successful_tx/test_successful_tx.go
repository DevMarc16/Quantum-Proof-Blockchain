package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

type JSONRPCRequest_success struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type JSONRPCResponse_success struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func runTestSuccessfulTx() {
	fmt.Println("🚀 Successful Quantum Transaction Test")
	fmt.Println("=====================================")
	
	// Step 1: Get validator private key from the node
	fmt.Println("🔑 Getting validator private key from node...")
	validatorKeyHex, err := getValidatorPrivateKey()
	if err != nil {
		log.Fatal("Failed to get validator key:", err)
	}
	fmt.Println("✅ Got validator private key")
	
	// Step 2: Get validator address from the node  
	validatorAddr, err := getValidatorAddress_success()
	if err != nil {
		log.Fatal("Failed to get validator address:", err)
	}
	fmt.Printf("💰 Checking validator balance: %s\n", validatorAddr)
	balance := getBalance_success(validatorAddr)
	fmt.Printf("✅ Validator balance: %s QTM\n", balance)
	
	if balance == "0x0" {
		log.Fatal("Validator has no balance. Make sure genesis funding is working.")
	}
	
	// Step 3: Convert hex key to private key
	keyBytes, err := hex.DecodeString(strings.TrimPrefix(validatorKeyHex, "0x"))
	if err != nil {
		log.Fatal("Failed to decode private key:", err)
	}
	
	privKey, err := crypto.DilithiumPrivateKeyFromBytes(keyBytes)
	if err != nil {
		log.Fatal("Failed to create private key:", err)
	}
	
	// Step 4: Get nonce for validator
	nonce := getNonce(validatorAddr)
	fmt.Printf("📝 Validator nonce: %s\n", nonce)
	
	// Convert nonce from hex
	nonceInt := big.NewInt(0)
	nonceInt.SetString(strings.TrimPrefix(nonce, "0x"), 16)
	
	// Step 5: Create transaction
	fmt.Println("\n💸 Creating funded quantum transaction...")
	
	chainID := big.NewInt(8888)
	toAddr := types.Address{0x74, 0x2d, 0x35, 0xCc, 0x66, 0x71, 0xC0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xD5, 0x81, 0xC0, 0x27, 0xd2, 0xb3, 0xd0, 0x7f}
	value := big.NewInt(1000000000000000000) // 1 QTM
	gasLimit := uint64(5800)                  // Optimized gas limit
	gasPrice := big.NewInt(1000000)          // Low gas price
	data := []byte{}
	
	tx := &types.QuantumTransaction{
		ChainID:   types.NewBigInt(chainID.Int64()),
		Nonce:     nonceInt.Uint64(),
		To:        &toAddr,
		Value:     types.NewBigInt(value.Int64()),
		Gas:       gasLimit,
		GasPrice:  types.NewBigInt(gasPrice.Int64()),
		Data:      data,
		SigAlg:    crypto.SigAlgDilithium,
	}
	
	// Step 6: Sign transaction with validator key
	fmt.Println("🔐 Signing with validator's quantum key...")
	pubKey := privKey.Public()
	tx.PublicKey = pubKey.Bytes()
	sigHash := tx.SigningHash()
	qrSig, err := crypto.SignMessage(sigHash[:], crypto.SigAlgDilithium, privKey.Bytes())
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}
	tx.Signature = qrSig.Signature
	fmt.Println("✅ Transaction signed successfully!")
	
	// Step 7: Verify signature locally
	valid, err := crypto.VerifySignature(sigHash[:], qrSig)
	if err != nil {
		log.Fatal("Signature verification error:", err)
	}
	if !valid {
		log.Fatal("Invalid signature!")
	}
	fmt.Println("✅ Signature verified locally")
	
	// Step 8: Submit transaction
	fmt.Println("\n📤 Submitting to quantum blockchain...")
	
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatal("JSON marshal error:", err)
	}
	
	txHash, err := submitTransaction_success(string(txJSON))
	if err != nil {
		log.Fatal("Failed to submit transaction:", err)
	}
	
	fmt.Printf("✅ SUCCESS! Transaction submitted: %s\n", txHash)
	
	// Step 9: Query the transaction back
	fmt.Println("\n🔍 Querying submitted transaction...")
	result, err := getTransaction(txHash)
	if err != nil {
		log.Fatal("Failed to query transaction:", err)
	}
	
	fmt.Println("✅ Transaction found in blockchain!")
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("📊 Transaction details:\n%s\n", resultJSON)
	
	fmt.Println("\n🎉 SUCCESS: Quantum transaction completed without errors!")
	fmt.Println("🔐 Your quantum blockchain is fully operational!")
}

func getValidatorPrivateKey() (string, error) {
	req := JSONRPCRequest_success{
		JSONRPC: "2.0",
		Method:  "test_getValidatorKey",
		Params:  []interface{}{},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_success(req)
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

func getValidatorAddress_success() (string, error) {
	req := JSONRPCRequest_success{
		JSONRPC: "2.0",
		Method:  "test_getValidatorAddress",
		Params:  []interface{}{},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_success(req)
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

func getBalance_success(address string) string {
	req := JSONRPCRequest_success{
		JSONRPC: "2.0",
		Method:  "eth_getBalance",
		Params:  []interface{}{address, "latest"},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_success(req)
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

func getNonce(address string) string {
	req := JSONRPCRequest_success{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionCount",
		Params:  []interface{}{address, "latest"},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_success(req)
	if err != nil {
		return "0x0"
	}
	
	if resp.Error != nil {
		return "0x0"
	}
	
	if result, ok := resp.Result.(string); ok {
		return result
	}
	
	return "0x0"
}

func submitTransaction_success(txJSON string) (string, error) {
	req := JSONRPCRequest_success{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []string{txJSON},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_success(req)
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
	req := JSONRPCRequest_success{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionByHash",
		Params:  []string{txHash},
		ID:      1,
	}
	
	resp, err := makeRPCRequest_success(req)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", resp.Error)
	}
	
	return resp.Result, nil
}

func makeRPCRequest_success(req JSONRPCRequest_success) (*JSONRPCResponse_success, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var rpcResp JSONRPCResponse_success
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, err
	}
	
	return &rpcResp, nil
}

func main() {
	runTestSuccessfulTx()
}