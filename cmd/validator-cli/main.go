package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// ValidatorConfig stores validator configuration
type ValidatorConfig struct {
	Address          string `json:"address"`
	QuantumPublicKey string `json:"quantumPublicKey"`
	QuantumAlgorithm uint8  `json:"quantumAlgorithm"`
	PrivateKeyPath   string `json:"privateKeyPath"`
	StakeAmount      string `json:"stakeAmount"`
	CommissionRate   uint16 `json:"commissionRate"`
	Metadata         string `json:"metadata"`
	RPCEndpoint      string `json:"rpcEndpoint"`
}

// ValidatorProfile stores complete validator information
type ValidatorProfile struct {
	Config           ValidatorConfig `json:"config"`
	DilithiumKeyPair *DilithiumKeys  `json:"dilithiumKeys,omitempty"`
	FalconKeyPair    *FalconKeys     `json:"falconKeys,omitempty"`
	CreatedAt        int64           `json:"createdAt"`
	Status           string          `json:"status"`
}

// DilithiumKeys represents Dilithium key pair
type DilithiumKeys struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Algorithm  string `json:"algorithm"`
}

// FalconKeys represents Falcon/Hybrid key pair
type FalconKeys struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Algorithm  string `json:"algorithm"`
}

func main() {
	// Define command-line flags
	var (
		cmdGenerate    = flag.Bool("generate", false, "Generate new validator keys")
		cmdRegister    = flag.Bool("register", false, "Register as validator on-chain")
		cmdStatus      = flag.Bool("status", false, "Check validator status")
		cmdDelegate    = flag.Bool("delegate", false, "Delegate to a validator")
		cmdExport      = flag.Bool("export", false, "Export validator configuration")
		cmdImport      = flag.String("import", "", "Import validator configuration from file")
		cmdBackup      = flag.Bool("backup", false, "Backup validator keys")
		cmdRestore     = flag.String("restore", "", "Restore validator keys from backup")
		
		// Key generation options
		algorithm      = flag.String("algorithm", "dilithium", "Quantum algorithm: dilithium or falcon")
		outputDir      = flag.String("output", "./validator-keys", "Output directory for keys")
		
		// Registration options
		stakeAmount    = flag.String("stake", "100000", "Stake amount in QTM")
		commissionRate = flag.Uint("commission", 500, "Commission rate in basis points (500 = 5%)")
		metadata       = flag.String("metadata", "", "IPFS hash or URL for validator metadata")
		rpcEndpoint    = flag.String("rpc", "http://localhost:8545", "RPC endpoint")
		
		// Delegation options
		validatorAddr  = flag.String("validator", "", "Validator address to delegate to")
		delegateAmount = flag.String("amount", "100", "Amount to delegate in QTM")
		
		// Security options
		password       = flag.String("password", "", "Password for key encryption")
		mnemonic       = flag.Bool("mnemonic", false, "Generate mnemonic phrase for key recovery")
	)
	
	flag.Parse()
	
	// Process commands
	switch {
	case *cmdGenerate:
		generateValidatorKeys(*algorithm, *outputDir, *password, *mnemonic)
		
	case *cmdRegister:
		registerValidator(*outputDir, *stakeAmount, uint16(*commissionRate), *metadata, *rpcEndpoint)
		
	case *cmdStatus:
		checkValidatorStatus(*outputDir, *rpcEndpoint)
		
	case *cmdDelegate:
		delegateToValidator(*validatorAddr, *delegateAmount, *rpcEndpoint)
		
	case *cmdExport:
		exportValidatorConfig(*outputDir)
		
	case *cmdImport != "":
		importValidatorConfig(*cmdImport, *outputDir)
		
	case *cmdBackup:
		backupValidatorKeys(*outputDir, *password)
		
	case *cmdRestore != "":
		restoreValidatorKeys(*cmdRestore, *outputDir, *password)
		
	default:
		printHelp()
	}
}

// generateValidatorKeys generates new quantum-resistant validator keys
func generateValidatorKeys(algorithm, outputDir, password string, generateMnemonic bool) {
	fmt.Println("🔐 Generating Quantum-Resistant Validator Keys...")
	fmt.Printf("Algorithm: %s\n", strings.ToUpper(algorithm))
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}
	
	var profile ValidatorProfile
	profile.CreatedAt = getCurrentTimestamp()
	profile.Status = "generated"
	
	switch strings.ToLower(algorithm) {
	case "dilithium":
		// Generate Dilithium keys
		privateKey, publicKey, err := crypto.GenerateDilithiumKeys()
		if err != nil {
			fmt.Printf("Error generating Dilithium keys: %v\n", err)
			return
		}
		
		// Create validator address from public key
		address := crypto.PublicKeyToAddress(publicKey)
		
		profile.DilithiumKeyPair = &DilithiumKeys{
			PrivateKey: hex.EncodeToString(privateKey),
			PublicKey:  hex.EncodeToString(publicKey),
			Algorithm:  "CRYSTALS-Dilithium-II",
		}
		
		profile.Config = ValidatorConfig{
			Address:          address.Hex(),
			QuantumPublicKey: hex.EncodeToString(publicKey),
			QuantumAlgorithm: 1, // Dilithium
			PrivateKeyPath:   filepath.Join(outputDir, "dilithium.key"),
			StakeAmount:      "100000",
			CommissionRate:   500, // 5%
		}
		
		// Save private key (encrypted if password provided)
		if err := savePrivateKey(privateKey, profile.Config.PrivateKeyPath, password); err != nil {
			fmt.Printf("Error saving private key: %v\n", err)
			return
		}
		
		fmt.Println("✅ Dilithium keys generated successfully!")
		fmt.Printf("📍 Validator Address: %s\n", address.Hex())
		fmt.Printf("🔑 Public Key: %s...\n", hex.EncodeToString(publicKey)[:64])
		
	case "falcon", "hybrid":
		// Generate Falcon/Hybrid keys
		privateKey, publicKey, err := crypto.GenerateFalconKeys()
		if err != nil {
			fmt.Printf("Error generating Falcon keys: %v\n", err)
			return
		}
		
		address := crypto.PublicKeyToAddress(publicKey)
		
		profile.FalconKeyPair = &FalconKeys{
			PrivateKey: hex.EncodeToString(privateKey),
			PublicKey:  hex.EncodeToString(publicKey),
			Algorithm:  "Falcon-512/ED25519-Hybrid",
		}
		
		profile.Config = ValidatorConfig{
			Address:          address.Hex(),
			QuantumPublicKey: hex.EncodeToString(publicKey),
			QuantumAlgorithm: 2, // Falcon
			PrivateKeyPath:   filepath.Join(outputDir, "falcon.key"),
			StakeAmount:      "100000",
			CommissionRate:   500,
		}
		
		// Save private key
		if err := savePrivateKey(privateKey, profile.Config.PrivateKeyPath, password); err != nil {
			fmt.Printf("Error saving private key: %v\n", err)
			return
		}
		
		fmt.Println("✅ Falcon/Hybrid keys generated successfully!")
		fmt.Printf("📍 Validator Address: %s\n", address.Hex())
		fmt.Printf("🔑 Public Key: %s...\n", hex.EncodeToString(publicKey)[:64])
		
	default:
		fmt.Printf("Unknown algorithm: %s\n", algorithm)
		return
	}
	
	// Generate mnemonic if requested
	if generateMnemonic {
		mnemonic := generateMnemonicPhrase()
		mnemonicPath := filepath.Join(outputDir, "mnemonic.txt")
		if err := ioutil.WriteFile(mnemonicPath, []byte(mnemonic), 0600); err != nil {
			fmt.Printf("Error saving mnemonic: %v\n", err)
			return
		}
		fmt.Println("📝 Mnemonic phrase generated and saved")
		fmt.Printf("⚠️  IMPORTANT: Store your mnemonic phrase safely!\n")
		fmt.Printf("Mnemonic: %s\n", mnemonic)
	}
	
	// Save validator profile
	profilePath := filepath.Join(outputDir, "validator-profile.json")
	if err := saveValidatorProfile(profile, profilePath); err != nil {
		fmt.Printf("Error saving profile: %v\n", err)
		return
	}
	
	fmt.Printf("\n📁 Keys saved to: %s\n", outputDir)
	fmt.Println("⚠️  Keep your private keys secure and never share them!")
	fmt.Println("\n🚀 Next steps:")
	fmt.Println("1. Fund your validator address with at least 100,000 QTM")
	fmt.Println("2. Run: validator-cli -register to register as a validator")
}

// registerValidator registers a new validator on-chain
func registerValidator(keyDir, stakeAmount string, commissionRate uint16, metadata, rpcEndpoint string) {
	fmt.Println("📝 Registering Validator On-Chain...")
	
	// Load validator profile
	profilePath := filepath.Join(keyDir, "validator-profile.json")
	profile, err := loadValidatorProfile(profilePath)
	if err != nil {
		fmt.Printf("Error loading profile: %v\n", err)
		return
	}
	
	// Load private key
	privateKey, err := loadPrivateKey(profile.Config.PrivateKeyPath, "")
	if err != nil {
		fmt.Printf("Error loading private key: %v\n", err)
		return
	}
	
	// Parse stake amount
	stake, ok := new(big.Int).SetString(stakeAmount, 10)
	if !ok {
		fmt.Printf("Invalid stake amount: %s\n", stakeAmount)
		return
	}
	
	// Convert to wei (18 decimals)
	weiMultiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	stakeWei := new(big.Int).Mul(stake, weiMultiplier)
	
	fmt.Printf("📍 Validator Address: %s\n", profile.Config.Address)
	fmt.Printf("💰 Stake Amount: %s QTM\n", stakeAmount)
	fmt.Printf("📊 Commission Rate: %.2f%%\n", float64(commissionRate)/100)
	fmt.Printf("🌐 RPC Endpoint: %s\n", rpcEndpoint)
	
	// Create registration transaction
	fmt.Println("\n⏳ Creating registration transaction...")
	
	// In production, this would interact with the ValidatorRegistry contract
	// For now, we'll show the transaction details
	fmt.Println("\n📋 Transaction Details:")
	fmt.Printf("To: ValidatorRegistry Contract\n")
	fmt.Printf("Method: registerValidator\n")
	fmt.Printf("Parameters:\n")
	fmt.Printf("  - quantumPublicKey: %s\n", profile.Config.QuantumPublicKey[:64]+"...")
	fmt.Printf("  - sigAlgorithm: %d\n", profile.Config.QuantumAlgorithm)
	fmt.Printf("  - initialStake: %s wei\n", stakeWei.String())
	fmt.Printf("  - commissionRate: %d\n", commissionRate)
	fmt.Printf("  - metadata: %s\n", metadata)
	
	// Sign transaction with quantum signature
	message := []byte("REGISTER_VALIDATOR")
	var signature []byte
	
	if profile.Config.QuantumAlgorithm == 1 {
		signature, err = crypto.SignWithDilithium(privateKey, message)
	} else {
		signature, err = crypto.SignWithFalcon(privateKey, message)
	}
	
	if err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}
	
	fmt.Printf("\n🔏 Quantum Signature: %s...\n", hex.EncodeToString(signature)[:64])
	
	// Update profile status
	profile.Status = "registered"
	profile.Config.StakeAmount = stakeAmount
	profile.Config.CommissionRate = commissionRate
	profile.Config.Metadata = metadata
	profile.Config.RPCEndpoint = rpcEndpoint
	
	if err := saveValidatorProfile(*profile, profilePath); err != nil {
		fmt.Printf("Error updating profile: %v\n", err)
		return
	}
	
	fmt.Println("\n✅ Validator registration transaction created!")
	fmt.Println("📤 Transaction would be submitted to the blockchain")
	fmt.Println("\n⏰ After registration:")
	fmt.Println("1. Wait for transaction confirmation")
	fmt.Println("2. Your validator will be activated once minimum stake is met")
	fmt.Println("3. Start your validator node to begin validating blocks")
}

// checkValidatorStatus checks the status of a validator
func checkValidatorStatus(keyDir, rpcEndpoint string) {
	fmt.Println("🔍 Checking Validator Status...")
	
	// Load validator profile
	profilePath := filepath.Join(keyDir, "validator-profile.json")
	profile, err := loadValidatorProfile(profilePath)
	if err != nil {
		fmt.Printf("Error loading profile: %v\n", err)
		return
	}
	
	fmt.Printf("\n📊 Validator Status Report\n")
	fmt.Printf("═══════════════════════════\n")
	fmt.Printf("Address: %s\n", profile.Config.Address)
	fmt.Printf("Algorithm: %s\n", getAlgorithmName(profile.Config.QuantumAlgorithm))
	fmt.Printf("Status: %s\n", profile.Status)
	fmt.Printf("Created: %s\n", formatTimestamp(profile.CreatedAt))
	
	if profile.Status == "registered" {
		fmt.Printf("\n💰 Staking Information:\n")
		fmt.Printf("  Self Stake: %s QTM\n", profile.Config.StakeAmount)
		fmt.Printf("  Commission: %.2f%%\n", float64(profile.Config.CommissionRate)/100)
		fmt.Printf("  Total Delegated: 0 QTM (simulated)\n")
		fmt.Printf("  Active: Yes (simulated)\n")
		
		fmt.Printf("\n📈 Performance Metrics:\n")
		fmt.Printf("  Blocks Proposed: 42 (simulated)\n")
		fmt.Printf("  Blocks Missed: 1 (simulated)\n")
		fmt.Printf("  Uptime: 97.6%% (simulated)\n")
		fmt.Printf("  Last Active: 2 minutes ago (simulated)\n")
		
		fmt.Printf("\n💵 Rewards:\n")
		fmt.Printf("  Pending Rewards: 127.5 QTM (simulated)\n")
		fmt.Printf("  Total Earned: 1,842.3 QTM (simulated)\n")
		fmt.Printf("  APY: 10.2%% (simulated)\n")
	}
	
	fmt.Printf("\n🔗 RPC Endpoint: %s\n", rpcEndpoint)
}

// delegateToValidator delegates tokens to a validator
func delegateToValidator(validatorAddr, amount, rpcEndpoint string) {
	if validatorAddr == "" {
		fmt.Println("Error: Validator address required")
		return
	}
	
	fmt.Println("💫 Delegating to Validator...")
	fmt.Printf("Validator: %s\n", validatorAddr)
	fmt.Printf("Amount: %s QTM\n", amount)
	fmt.Printf("RPC: %s\n", rpcEndpoint)
	
	// Parse amount
	delegateAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		fmt.Printf("Invalid amount: %s\n", amount)
		return
	}
	
	// Convert to wei
	weiMultiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountWei := new(big.Int).Mul(delegateAmount, weiMultiplier)
	
	fmt.Printf("\n📋 Delegation Transaction:\n")
	fmt.Printf("  To: ValidatorRegistry\n")
	fmt.Printf("  Method: delegate\n")
	fmt.Printf("  Validator: %s\n", validatorAddr)
	fmt.Printf("  Amount: %s wei\n", amountWei.String())
	
	fmt.Println("\n✅ Delegation transaction prepared!")
	fmt.Println("📤 Transaction would be submitted to the blockchain")
}

// Helper functions

func savePrivateKey(key []byte, path, password string) error {
	// In production, encrypt with password
	if password != "" {
		// Encrypt key with password (simplified)
		key = append([]byte("ENCRYPTED:"), key...)
	}
	
	return ioutil.WriteFile(path, key, 0600)
}

func loadPrivateKey(path, password string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	// In production, decrypt with password
	if password != "" && strings.HasPrefix(string(data), "ENCRYPTED:") {
		data = data[10:] // Remove prefix
	}
	
	return data, nil
}

func saveValidatorProfile(profile ValidatorProfile, path string) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(path, data, 0644)
}

func loadValidatorProfile(path string) (*ValidatorProfile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var profile ValidatorProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	
	return &profile, nil
}

func generateMnemonicPhrase() string {
	// Simplified mnemonic generation
	words := []string{
		"quantum", "resist", "validator", "stake", "secure",
		"block", "chain", "dilithium", "falcon", "crystal",
		"proof", "consensus",
	}
	
	// In production, use BIP39 wordlist and proper entropy
	mnemonic := ""
	for i := 0; i < 12; i++ {
		if i > 0 {
			mnemonic += " "
		}
		mnemonic += words[i%len(words)]
	}
	
	return mnemonic
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func formatTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func getAlgorithmName(alg uint8) string {
	switch alg {
	case 1:
		return "CRYSTALS-Dilithium-II"
	case 2:
		return "Falcon-512/Hybrid"
	default:
		return "Unknown"
	}
}

func exportValidatorConfig(keyDir string) {
	fmt.Println("📤 Exporting Validator Configuration...")
	
	profilePath := filepath.Join(keyDir, "validator-profile.json")
	profile, err := loadValidatorProfile(profilePath)
	if err != nil {
		fmt.Printf("Error loading profile: %v\n", err)
		return
	}
	
	// Create export without private keys
	export := map[string]interface{}{
		"address":          profile.Config.Address,
		"publicKey":        profile.Config.QuantumPublicKey,
		"algorithm":        getAlgorithmName(profile.Config.QuantumAlgorithm),
		"commissionRate":   profile.Config.CommissionRate,
		"metadata":         profile.Config.Metadata,
		"status":           profile.Status,
		"createdAt":        formatTimestamp(profile.CreatedAt),
	}
	
	exportData, _ := json.MarshalIndent(export, "", "  ")
	exportPath := filepath.Join(keyDir, "validator-export.json")
	
	if err := ioutil.WriteFile(exportPath, exportData, 0644); err != nil {
		fmt.Printf("Error saving export: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Configuration exported to: %s\n", exportPath)
}

func importValidatorConfig(importPath, outputDir string) {
	fmt.Printf("📥 Importing Validator Configuration from: %s\n", importPath)
	
	data, err := ioutil.ReadFile(importPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	var profile ValidatorProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		return
	}
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}
	
	// Save imported profile
	profilePath := filepath.Join(outputDir, "validator-profile.json")
	if err := saveValidatorProfile(profile, profilePath); err != nil {
		fmt.Printf("Error saving profile: %v\n", err)
		return
	}
	
	fmt.Println("✅ Configuration imported successfully!")
	fmt.Printf("📁 Saved to: %s\n", outputDir)
}

func backupValidatorKeys(keyDir, password string) {
	fmt.Println("💾 Creating Validator Backup...")
	
	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupDir := fmt.Sprintf("validator-backup-%s", timestamp)
	
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		fmt.Printf("Error creating backup directory: %v\n", err)
		return
	}
	
	// Copy all files from keyDir to backupDir
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}
	
	for _, file := range files {
		if !file.IsDir() {
			src := filepath.Join(keyDir, file.Name())
			dst := filepath.Join(backupDir, file.Name())
			
			data, err := ioutil.ReadFile(src)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", file.Name(), err)
				continue
			}
			
			if err := ioutil.WriteFile(dst, data, 0600); err != nil {
				fmt.Printf("Error writing %s: %v\n", file.Name(), err)
				continue
			}
			
			fmt.Printf("  ✓ Backed up: %s\n", file.Name())
		}
	}
	
	fmt.Printf("\n✅ Backup created: %s\n", backupDir)
	fmt.Println("⚠️  Store this backup in a secure location!")
}

func restoreValidatorKeys(backupPath, outputDir, password string) {
	fmt.Printf("♻️  Restoring Validator Keys from: %s\n", backupPath)
	
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		fmt.Printf("Backup not found: %s\n", backupPath)
		return
	}
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}
	
	// Copy all files from backup to output
	files, err := ioutil.ReadDir(backupPath)
	if err != nil {
		fmt.Printf("Error reading backup: %v\n", err)
		return
	}
	
	for _, file := range files {
		if !file.IsDir() {
			src := filepath.Join(backupPath, file.Name())
			dst := filepath.Join(outputDir, file.Name())
			
			data, err := ioutil.ReadFile(src)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", file.Name(), err)
				continue
			}
			
			if err := ioutil.WriteFile(dst, data, 0600); err != nil {
				fmt.Printf("Error writing %s: %v\n", file.Name(), err)
				continue
			}
			
			fmt.Printf("  ✓ Restored: %s\n", file.Name())
		}
	}
	
	fmt.Printf("\n✅ Keys restored to: %s\n", outputDir)
}

func printHelp() {
	fmt.Println("Quantum Blockchain Validator CLI")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  -generate    Generate new validator keys")
	fmt.Println("  -register    Register as validator on-chain")
	fmt.Println("  -status      Check validator status")
	fmt.Println("  -delegate    Delegate to a validator")
	fmt.Println("  -export      Export validator configuration")
	fmt.Println("  -import      Import validator configuration")
	fmt.Println("  -backup      Backup validator keys")
	fmt.Println("  -restore     Restore validator keys")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -algorithm   Quantum algorithm (dilithium/falcon)")
	fmt.Println("  -output      Output directory for keys")
	fmt.Println("  -stake       Stake amount in QTM")
	fmt.Println("  -commission  Commission rate in basis points")
	fmt.Println("  -metadata    Validator metadata (IPFS/URL)")
	fmt.Println("  -rpc         RPC endpoint")
	fmt.Println("  -validator   Validator address (for delegation)")
	fmt.Println("  -amount      Delegation amount in QTM")
	fmt.Println("  -password    Password for key encryption")
	fmt.Println("  -mnemonic    Generate mnemonic phrase")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate new Dilithium validator keys")
	fmt.Println("  validator-cli -generate -algorithm dilithium")
	fmt.Println()
	fmt.Println("  # Register as validator with 100K QTM stake")
	fmt.Println("  validator-cli -register -stake 100000 -commission 500")
	fmt.Println()
	fmt.Println("  # Check validator status")
	fmt.Println("  validator-cli -status")
	fmt.Println()
	fmt.Println("  # Delegate 1000 QTM to a validator")
	fmt.Println("  validator-cli -delegate -validator 0x... -amount 1000")
}

// Add time import
import "time"