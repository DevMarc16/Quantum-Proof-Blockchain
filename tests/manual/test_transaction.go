package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func runTestTransaction() {
	// Generate keys
	privKey, _, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	// Create transaction
	chainID := big.NewInt(8888)
	nonce := uint64(0)
	to, _ := types.HexToAddress("0x742d35Cc6671C0532925a3b8D581C027d2b3d07f")
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei
	data := []byte{}

	tx := types.NewQuantumTransaction(chainID, nonce, &to, value, gasLimit, gasPrice, data)

	// Sign transaction
	err = tx.SignTransaction(privKey.Bytes(), crypto.SigAlgDilithium)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	fmt.Printf("Transaction Hash: %s\n", tx.Hash().Hex())
	fmt.Printf("From Address: %s\n", tx.From().Hex())
	fmt.Printf("Signature Algorithm: %d\n", tx.SigAlg)
	fmt.Printf("Public Key Length: %d\n", len(tx.PublicKey))
	fmt.Printf("Signature Length: %d\n", len(tx.Signature))

	// Verify signature works locally
	valid, err := tx.VerifySignature()
	if err != nil {
		log.Fatal("Signature verification error:", err)
	}
	fmt.Printf("Local signature verification: %t\n", valid)

	// Marshal to JSON to see the format
	jsonData, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		log.Fatal("JSON marshal error:", err)
	}
	fmt.Printf("Transaction JSON:\n%s\n", jsonData)
}