package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	fmt.Println("🚀 Deploying Quantum Validator System Contracts")
	fmt.Println("===============================================")

	// Connect to our quantum blockchain
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to blockchain:", err)
	}

	// Setup deployer account (Foundry test account #1)
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		log.Fatal("Failed to parse private key:", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	fmt.Printf("Deployer address: %s\n", fromAddress.Hex())

	// Check balance
	balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}
	fmt.Printf("Deployer balance: %s ETH\n", formatEther(balance))

	// Setup transaction parameters
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Failed to get nonce:", err)
	}

	gasPrice := big.NewInt(1000000000) // 1 gwei
	chainID := big.NewInt(8888)        // Our quantum chain ID

	fmt.Printf("Chain ID: %d\n", chainID.Int64())
	fmt.Printf("Starting nonce: %d\n", nonce)

	// Create a simple token contract first (ERC-20 basic)
	simpleTokenBytecode := "608060405234801561001057600080fd5b5060405161082838038061082883398101604081905261002f9161008e565b600080546001600160a01b03199081163390811783556001805490921617905560028190556003819055604051908152600080516020610808833981519152906020015b60405180910390a25061013e565b600060208284031215156100a057600080fd5b5051919050565b60006020828403121561010657600080fd5b81516001600160a01b038116811461011d57600080fd5b9392505050565b610628806101326000396000f3fe"

	fmt.Println("")
	fmt.Println("📄 Deploying Simple Token Contract...")

	// Create contract creation transaction
	tx := types.NewContractCreation(
		nonce,
		big.NewInt(0),                       // value
		uint64(1000000),                     // gas limit
		gasPrice,                            // gas price
		common.FromHex(simpleTokenBytecode), // data
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal("Failed to send transaction:", err)
	}

	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

	// Wait for receipt
	fmt.Print("Waiting for transaction to be mined...")
	for i := 0; i < 30; i++ { // Wait up to 30 * 2 = 60 seconds
		receipt, err := client.TransactionReceipt(context.Background(), signedTx.Hash())
		if err == nil {
			fmt.Println(" ✅ Mined!")
			fmt.Printf("Contract deployed to: %s\n", receipt.ContractAddress.Hex())
			fmt.Printf("Gas used: %d\n", receipt.GasUsed)
			fmt.Printf("Block number: %d\n", receipt.BlockNumber.Uint64())

			// Create deployment config
			fmt.Println("")
			fmt.Println("🎉 Deployment Successful!")
			fmt.Println("Contract Address:", receipt.ContractAddress.Hex())
			return
		}

		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}

	fmt.Println(" ❌ Timeout waiting for transaction")
}

func formatEther(wei *big.Int) string {
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return ether.Text('f', 6)
}
