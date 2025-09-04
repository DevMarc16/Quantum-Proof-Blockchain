package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Connect to blockchain
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to blockchain:", err)
	}

	// Set up deployer account
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		log.Fatal("Failed to create private key:", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Deploying from address: %s\n", fromAddress.Hex())

	chainID := big.NewInt(8888) // Our quantum chain ID
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal("Failed to create auth:", err)
	}

	// Set legacy transaction parameters
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(1000000000) // 1 gwei fallback
	}

	auth.GasPrice = gasPrice
	auth.GasLimit = uint64(2000000)

	fmt.Println("🚀 Deploying QTM Token Contract...")

	// QTM Token bytecode (from compiled contract)
	qtmBytecode := "0x6080604052601260025f6101000a81548160ff021916908360ff16021790555034801561002a575f5ffd5b506040516121c53803806121c5833981810160405281019061004c919061033f565b825f908161005a91906105ce565b50816001908161006a91906105ce565b503360045f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600160055f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055505f8111156101b757806003819055508060065f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055503373ffffffffffffffffffffffffffffffffffffffff165f73ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040516101ae91906106ac565b60405180910390a35b5050506106c5565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61021e826101d8565b810181811067ffffffffffffffff8211171561023d5761023c6101e8565b5b80604052505050565b5f61024f6101bf565b905061025b8282610215565b919050565b5f67ffffffffffffffff82111561027a576102796101e8565b5b610283826101d8565b9050602081019050919050565b8281835e5f83830152505050565b5f6102b06102ab84610260565b610246565b9050828152602081018484840111156102cc576102cb6101d4565b5b6102d7848285610290565b509392505050565b5f82601f8301126102f3576102f26101d0565b5b815161030384826020860161029e565b91505092915050565b5f819050919050565b61031e8161030c565b8114610328575f5ffd5b50565b5f8151905061033981610315565b92915050565b5f5f5f60608486031215610356576103556101c8565b5b5f84015167ffffffffffffffff811115610373576103726101cc565b5b61037f868287016102df565b935050602084015167ffffffffffffffff8111156103a05761039f6101cc565b5b6103ac868287016102df565b92505060406103bd8682870161032b565b9150509250925092565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061041557607f821691505b602082108103610428576104276103d1565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f6008830261048a7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8261044f565b610494868361044f565b95508019841693508086168417925050509392505050565b5f819050919050565b5f6104cf6104ca6104c58461030c565b6104ac565b61030c565b9050919050565b5f819050919050565b6104e8836104b5565b6104fc6104f4826104d6565b84845461045b565b825550505050565b5f5f905090565b610513610504565b61051e8184846104df565b505050565b5b81811015610541576105365f8261050b565b600181019050610524565b5050565b601f821115610586576105578161042e565b61056084610440565b8101602085101561056f578190505b61058361057b85610440 830182610523565b50505b505050565b5f82821c905092915050565b5f6105a65f198460080261058b565b1980831691505092915050565b5f6105be8383610597565b9150826002028217905092915050565b6105d7826103c7565b67ffffffffffffffff8111156105f0576105ef6101e8565b5b6105fa82546103fe565b610605828285610545565b5f60209050601f831160018114610636575f8415610624578287015190505b61062e85826105b3565b865550610695565b601f1984166106448661042e565b5f5b8281101561066b57848901518255600182019150602085019450602081019050610646565b868310156106885784890151610684601f891682610597565b8355505b6001600288020188555050505b505050505000000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000033b2e3c9fd0803ce8000000000000000000000000000000000000000000000000000000000000000000000d5175616e74756d20546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000351544d0000000000000000000000000000000000000000000000000000000000"

	// Deploy QTM Token
	qtmAddress, tx, _, err := deployContract(client, auth, qtmBytecode)
	if err != nil {
		log.Fatal("Failed to deploy QTM Token:", err)
	}

	fmt.Printf("✅ QTM Token deployed to: %s\n", qtmAddress.Hex())
	fmt.Printf("   Transaction hash: %s\n", tx.Hash().Hex())

	// Wait for deployment
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal("Failed to wait for QTM deployment:", err)
	}
	fmt.Println("   Transaction mined successfully!")

	fmt.Println("")
	fmt.Println("🎉 Contract Deployment Complete!")
	fmt.Println("=================================")
	fmt.Printf("QTM Token: %s\n", qtmAddress.Hex())
}

func deployContract(client *ethclient.Client, auth *bind.TransactOpts, bytecode string) (common.Address, *types.Transaction, interface{}, error) {
	// Remove 0x prefix if present
	if strings.HasPrefix(bytecode, "0x") {
		bytecode = bytecode[2:]
	}

	// Convert hex string to bytes
	data := common.FromHex(bytecode)

	// Create contract deployment transaction
	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	gasLimit := auth.GasLimit
	gasPrice := auth.GasPrice
	value := big.NewInt(0)

	tx := types.NewContractCreation(nonce, value, gasLimit, gasPrice, data)

	// Sign transaction
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	// Calculate contract address
	contractAddress := crypto.CreateAddress(auth.From, nonce)

	return contractAddress, signedTx, nil, nil
}
