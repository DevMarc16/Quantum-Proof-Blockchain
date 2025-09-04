package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"quantum-blockchain/chain/types"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Connect to blockchain
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Test simple balance query first
	addr, err := types.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	if err != nil {
		log.Fatal("Failed to parse address:", err)
	}
	balance, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}

	fmt.Printf("Account balance: %s QTM\n", balance.String())

	// Test raw transaction creation using our quantum transaction format
	qtx := &types.QuantumTransaction{
		ChainID:         big.NewInt(8888),
		Nonce:          0,
		To:             &types.Address{0x74, 0x2d, 0x35, 0xCc, 0x66, 0x34, 0xC0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xD0, 0x00, 0xB1, 0xb0, 0x00, 0xd1, 0xb0, 0x00},
		Value:          big.NewInt(1000000000000000000), // 1 QTM
		GasLimit:       21000,
		GasPrice:       big.NewInt(1000000000),
		Data:           []byte{},
		SignatureAlg:   1, // Dilithium
		PublicKey:      []byte("test-public-key"),
		Signature:      []byte("test-signature"),
	}

	// Test encoding to JSON for our blockchain
	jsonData, err := qtx.MarshalJSON()
	if err != nil {
		log.Fatal("Failed to marshal transaction:", err)
	}

	fmt.Printf("Quantum transaction JSON: %s\n", string(jsonData))

	// Test hex encoding for RPC
	hexData := hex.EncodeToString(jsonData)
	fmt.Printf("Hex encoded: 0x%s\n", hexData)

	fmt.Println("✅ Transaction encoding test successful!")
}