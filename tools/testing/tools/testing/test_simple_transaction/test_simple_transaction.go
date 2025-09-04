package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Connect to blockchain
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Test simple balance query first
	addr, _ := types.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	// Convert our types.Address to ethereum common.Address
	ethAddr := common.BytesToAddress(addr.Bytes())
	balance, err := client.BalanceAt(context.Background(), ethAddr, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}

	fmt.Printf("Account balance: %s QTM\n", balance.String())

	// Test raw transaction creation using our quantum transaction format
	qtx := &types.QuantumTransaction{
		ChainID:   big.NewInt(8888),
		Nonce:     0,
		To:        &addr,
		Value:     big.NewInt(1000000000000000000), // 1 QTM
		Gas:       21000,
		GasPrice:  big.NewInt(1000000000),
		Data:      []byte{},
		SigAlg:    crypto.SigAlgDilithium,
		PublicKey: []byte("test-public-key"),
		Signature: []byte("test-signature"),
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
