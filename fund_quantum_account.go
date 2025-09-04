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
	fmt.Println("💰 Funding Quantum Account for Contract Deployment")
	fmt.Println("==============================================")

	// Generate quantum address for deployment
	fmt.Println("🔐 Generating Dilithium keys for deployment account...")
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	privateKeyBytes := privKey.Bytes()
	publicKeyBytes := pubKey.Bytes()
	quantumAddr := types.PublicKeyToAddress(publicKeyBytes)

	fmt.Printf("✅ Generated quantum address: %s\n", quantumAddr.Hex())
	fmt.Printf("📊 Private key: %x\n", privateKeyBytes)
	fmt.Printf("📊 Public key length: %d bytes\n", len(publicKeyBytes))

	// Check current balance
	resp, err := callRPC("eth_getBalance", []interface{}{quantumAddr.Hex(), "latest"})
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error getting balance:", resp.Error.Message)
	}

	currentBalance := resp.Result.(string)
	fmt.Printf("💳 Current balance: %s\n", currentBalance)

	// Save the keys and address to a file for later use
	keyInfo := map[string]string{
		"address":    quantumAddr.Hex(),
		"privateKey": hex.EncodeToString(privateKeyBytes),
		"publicKey":  hex.EncodeToString(publicKeyBytes),
	}

	keyInfoJSON, _ := json.MarshalIndent(keyInfo, "", "  ")
	fmt.Printf("\n🔑 Deployment Account Info:\n%s\n", string(keyInfoJSON))
	
	// For now, just output the info - in a real scenario we'd need to fund this account
	// from the pre-funded account using a regular transaction (not quantum)
	
	fundingAmount := big.NewInt(0)
	fundingAmount.SetString("5000000000000000000", 10) // 5 ETH
	
	fmt.Printf("\n💸 This account needs funding of %s wei (%s ETH)\n", 
		fundingAmount.String(), 
		new(big.Float).Quo(new(big.Float).SetInt(fundingAmount), new(big.Float).SetInt64(1e18)).String())
	
	fmt.Println("\n✅ Save this information to deploy contracts with the funded quantum account!")
}