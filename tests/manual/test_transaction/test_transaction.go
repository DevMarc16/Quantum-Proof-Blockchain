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
	privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatal("Failed to generate keys:", err)
	}

	// Create transaction
	chainID := big.NewInt(8888)
	nonce := uint64(0)
	toAddr := types.Address{0x74, 0x2d, 0x35, 0xCc, 0x66, 0x71, 0xC0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xD5, 0x81, 0xC0, 0x27, 0xd2, 0xb3, 0xd0, 0x7f}
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei
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
		log.Fatal("Failed to sign transaction:", err)
	}
	tx.Signature = qrSig.Signature

	fmt.Printf("Transaction Hash: %s\n", tx.Hash().Hex())
	fmt.Printf("From Address: %s\n", tx.From().Hex())
	fmt.Printf("Signature Algorithm: %d\n", tx.SigAlg)
	fmt.Printf("Public Key Length: %d\n", len(tx.PublicKey))
	fmt.Printf("Signature Length: %d\n", len(tx.Signature))

	// Verify signature works locally
	valid, err := crypto.VerifySignature(sigHash[:], qrSig)
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

func main() {
	runTestTransaction()
}