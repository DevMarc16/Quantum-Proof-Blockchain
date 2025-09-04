package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	fmt.Println("🔍 Debugging Quantum Transaction")
	
	// Recreate the same transaction as before
	nonce := "0x0"
	nonceInt, _ := strconv.ParseUint(nonce[2:], 16, 64)
	
	// Generate same type of keys
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	privateKeyBytes := privKey.Bytes()
	_ = pubKey.Bytes() // publicKeyBytes not used in this debug

	// Simple ERC-20 token bytecode
	qtmBytecode := "608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506040518060400160405280600d81526020017f5175616e74756d20546f6b656e000000000000000000000000000000000000008152506001908161009c9190610275565b50"

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

	fmt.Printf("Transaction before signing:\n")
	fmt.Printf("  ChainID: %s\n", tx.ChainID.String())
	fmt.Printf("  Nonce: %d\n", tx.Nonce)
	fmt.Printf("  To: %v\n", tx.To)
	fmt.Printf("  Value: %s\n", tx.Value.String())
	fmt.Printf("  Gas: %d\n", tx.Gas)
	fmt.Printf("  GasPrice: %s\n", tx.GasPrice.String())
	
	// Sign the transaction
	err = tx.SignTransaction(privateKeyBytes, crypto.SigAlgDilithium)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	// Marshall to JSON and check
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		log.Fatal("Failed to marshal transaction:", err)
	}

	fmt.Printf("\nTransaction JSON:\n%s\n", string(txJSON))
	
	// Try to unmarshal it back to see if there are any issues
	var parsedTx types.QuantumTransaction
	err = parsedTx.UnmarshalJSON(txJSON)
	if err != nil {
		log.Printf("❌ Failed to unmarshal: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully unmarshaled\n")
	}
	
	// Check individual address components
	fmt.Printf("\nAddress Analysis:\n")
	if tx.To != nil {
		fmt.Printf("  To address: %s (length: %d)\n", tx.To.Hex(), len(tx.To.Hex()))
	} else {
		fmt.Printf("  To address: nil (contract creation)\n")
	}
	
	fromAddr := tx.From()
	fmt.Printf("  From address: %s (length: %d)\n", fromAddr.Hex(), len(fromAddr.Hex()))
}