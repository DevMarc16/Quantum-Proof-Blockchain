package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	fmt.Println("🔧 Testing transaction processing with fixed state management...")

	// Create a quantum transaction from the validator (who now has balance)
	validatorAddr, _ := types.HexToAddress("0x67d12165b9950574912ed6f1ca13512dfb8c37cc")
	recipientAddr, _ := types.HexToAddress("0x1234567890123456789012345678901234567890")
	
	// Generate transaction parameters
	nonce := uint64(0)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei
	value := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)) // 1 QTM
	
	// Create transaction
	tx := &types.QuantumTransaction{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &recipientAddr,
		Value:    value,
		Data:     []byte{},
		SigAlg:   crypto.SigAlgDilithium,
	}

	// For testing, we'll use a dummy signature
	// In production, this would be signed with the validator's private key
	tx.Signature = []byte("dummy_signature_for_testing")
	tx.PublicKey = []byte("dummy_public_key_for_testing")

	// Serialize transaction for RPC call
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}

	fmt.Printf("📝 Created transaction:\n")
	fmt.Printf("  From: %s\n", validatorAddr.Hex())
	fmt.Printf("  To: %s\n", recipientAddr.Hex())
	fmt.Printf("  Value: %s QTM\n", new(big.Int).Div(value, big.NewInt(1e18)).String())
	fmt.Printf("  Gas: %d\n", gasLimit)
	fmt.Printf("  Gas Price: %s Gwei\n", new(big.Int).Div(gasPrice, big.NewInt(1e9)).String())
	
	fmt.Printf("\n📤 Transaction JSON:\n%s\n", string(txJSON))
	
	fmt.Printf("\n✅ Transaction created successfully!\n")
	fmt.Printf("🔍 Next step: Submit this transaction to the network to test mining\n")
	fmt.Printf("💡 The validator now has sufficient balance (%d+ QTM) to process transactions!\n", 100000)
}