package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

func main() {
	fmt.Println("🔍 Deriving Actual Validator Addresses from Key Files")
	fmt.Println("=====================================================")

	validators := []struct {
		name    string
		keyFile string
	}{
		{"Validator 1", "validator-1-data/validator.key"},
		{"Validator 2", "validator-2-data/validator.key"},
		{"Validator 3", "validator-3-data/validator.key"},
	}

	for _, v := range validators {
		fmt.Printf("\n%s:\n", v.name)
		
		// Read the key file
		hexData, err := ioutil.ReadFile(v.keyFile)
		if err != nil {
			fmt.Printf("  ❌ Failed to read key file %s: %v\n", v.keyFile, err)
			continue
		}

		// Decode hex
		keyBytes, err := hex.DecodeString(string(hexData))
		if err != nil {
			fmt.Printf("  ❌ Failed to decode hex from %s: %v\n", v.keyFile, err)
			continue
		}

		// Parse private key
		privKey, err := crypto.DilithiumPrivateKeyFromBytes(keyBytes)
		if err != nil {
			fmt.Printf("  ❌ Failed to parse private key from %s: %v\n", v.keyFile, err)
			continue
		}

		// Get public key
		pubKey := privKey.Public()
		
		// Derive address
		address := types.PublicKeyToAddress(pubKey.Bytes())
		
		fmt.Printf("  📍 Address: %s\n", address.Hex())
		fmt.Printf("  🔑 Key File: %s\n", v.keyFile)
	}

	fmt.Println("\n✅ Address derivation complete!")
}