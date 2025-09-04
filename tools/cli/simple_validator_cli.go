package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	var (
		generate  = flag.Bool("generate", false, "Generate new quantum keys")
		algorithm = flag.String("algorithm", "dilithium", "Algorithm to use (dilithium, falcon)")
		output    = flag.String("output", "./validator-keys", "Output directory for keys")
	)
	flag.Parse()

	if *generate {
		generateKeys(*algorithm, *output)
	} else {
		showUsage()
	}
}

func generateKeys(algorithm, outputDir string) {
	fmt.Printf("🔐 Generating %s quantum keys...\n", algorithm)

	var privateKey, publicKey []byte
	var err error

	switch algorithm {
	case "dilithium":
		privKey, pubKey, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			log.Fatal("Failed to generate Dilithium keys:", err)
		}
		privateKey = privKey.Bytes()
		publicKey = pubKey.Bytes()
	case "falcon":
		privKey, pubKey, err := crypto.GenerateFalconKeyPair()
		if err != nil {
			log.Fatal("Failed to generate Falcon keys:", err)
		}
		privateKey = privKey.Bytes()
		publicKey = pubKey.Bytes()
	default:
		log.Fatal("Unsupported algorithm:", algorithm)
	}

	// Create output directory
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Save private key
	privateKeyFile := fmt.Sprintf("%s/%s_private.key", outputDir, algorithm)
	err = os.WriteFile(privateKeyFile, []byte(hex.EncodeToString(privateKey)), 0600)
	if err != nil {
		log.Fatal("Failed to write private key:", err)
	}

	// Save public key
	publicKeyFile := fmt.Sprintf("%s/%s_public.key", outputDir, algorithm)
	err = os.WriteFile(publicKeyFile, []byte(hex.EncodeToString(publicKey)), 0644)
	if err != nil {
		log.Fatal("Failed to write public key:", err)
	}

	fmt.Printf("✅ Keys generated successfully!\n")
	fmt.Printf("   Private key: %s\n", privateKeyFile)
	fmt.Printf("   Public key:  %s\n", publicKeyFile)
	fmt.Printf("   Public key length: %d bytes\n", len(publicKey))
	fmt.Printf("   Private key length: %d bytes\n", len(privateKey))

	// Calculate address from public key using utility function
	address := types.PublicKeyToAddress(publicKey)
	fmt.Printf("   Validator address: %s\n", address.Hex())

	fmt.Println("")
	fmt.Println("🚀 Next steps:")
	fmt.Println("1. Get testnet tokens: Use faucet to request 100K QTM")
	fmt.Println("2. Register validator: Submit registration transaction")
	fmt.Println("3. Start validating: Run quantum-node with --validator flag")
}

func showUsage() {
	fmt.Println("Quantum Validator CLI")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  Generate keys:  ./simple_validator_cli -generate -algorithm dilithium -output ./my-keys")
	fmt.Println("  Generate keys:  ./simple_validator_cli -generate -algorithm falcon -output ./my-keys")
	fmt.Println()
	fmt.Println("Supported algorithms:")
	fmt.Println("  - dilithium: CRYSTALS-Dilithium-II (2420-byte signatures)")
	fmt.Println("  - falcon:    Falcon hybrid ED25519+Dilithium")
}
