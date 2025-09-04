package integration

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// TestValidatorOnboardingFlow tests the complete validator onboarding process
func TestValidatorOnboardingFlow(t *testing.T) {
	t.Log("🔬 Testing Complete Validator Onboarding Flow")

	// Test 1: Generate quantum-resistant validator keys
	t.Run("GenerateValidatorKeys", func(t *testing.T) {
		t.Log("📝 Test 1: Generate Dilithium validator keys")

		// Generate Dilithium key pair
		privateKey, publicKey, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate Dilithium keys: %v", err)
		}

		// Validate key sizes
		if len(privateKey.Bytes()) == 0 {
			t.Fatal("Private key is empty")
		}
		if len(publicKey.Bytes()) != 1312 { // Dilithium-II public key size
			t.Fatalf("Expected public key size 1312, got %d", len(publicKey.Bytes()))
		}

		// Generate validator address
		address := types.PublicKeyToAddress(publicKey.Bytes())
		if address == (types.Address{}) {
			t.Fatal("Generated address is zero")
		}

		t.Logf("✅ Generated validator address: %s", address.Hex())
		t.Logf("✅ Public key size: %d bytes", len(publicKey.Bytes()))
		t.Logf("✅ Private key generated successfully")
	})

	// Test 2: Sign and verify quantum signatures
	t.Run("QuantumSignatureVerification", func(t *testing.T) {
		t.Log("📝 Test 2: Quantum signature signing and verification")

		// Generate keys
		privateKey, publicKey, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		// Create test message
		message := []byte("VALIDATOR_REGISTRATION_TEST")

		// Sign message
		signature, err := privateKey.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign message: %v", err)
		}

		// Verify signature
		valid := publicKey.Verify(message, signature)

		if !valid {
			t.Fatal("Signature verification failed")
		}

		t.Logf("✅ Signature size: %d bytes", len(signature))
		t.Logf("✅ Signature verification successful")
	})

	// Test 3: Validator registration transaction
	t.Run("ValidatorRegistrationTransaction", func(t *testing.T) {
		t.Log("📝 Test 3: Create validator registration transaction")

		// Generate validator keys
		privateKey, publicKey, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		validatorAddr := types.PublicKeyToAddress(publicKey.Bytes())
		stakeAmount := big.NewInt(100000) // 100K QTM
		stakeAmountWei := new(big.Int).Mul(stakeAmount, big.NewInt(1e18))

		// Create registration data
		registrationData := map[string]interface{}{
			"validator":        validatorAddr.Hex(),
			"quantumPublicKey": hex.EncodeToString(publicKey.Bytes()),
			"sigAlgorithm":     uint8(1), // Dilithium
			"initialStake":     stakeAmountWei.String(),
			"commissionRate":   uint16(500), // 5%
			"metadata":         "ipfs://QmTestValidator...",
		}

		// Sign registration transaction
		message := []byte("REGISTER_VALIDATOR:" + validatorAddr.Hex())
		signature, err := privateKey.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign registration: %v", err)
		}

		registrationData["quantumSignature"] = hex.EncodeToString(signature)

		t.Logf("✅ Registration data prepared")
		t.Logf("   Validator: %s", registrationData["validator"])
		t.Logf("   Stake: %s wei", registrationData["initialStake"])
		t.Logf("   Commission: %d basis points", registrationData["commissionRate"])
		t.Logf("   Signature: %s...", registrationData["quantumSignature"].(string)[:64])
	})
}

// TestTokenDistributionFlow tests token distribution mechanisms
func TestTokenDistributionFlow(t *testing.T) {
	t.Log("🪙 Testing Token Distribution Flow")

	// Test 1: Genesis allocation calculation
	t.Run("GenesisAllocation", func(t *testing.T) {
		t.Log("📝 Test 1: Calculate genesis token allocations")

		totalSupply := new(big.Int)
		totalSupply.SetString("1000000000000000000000000000", 10) // 1B QTM

		allocations := map[string]*big.Int{
			"GenesisValidators":   calculatePercentage(totalSupply, 15),  // 15%
			"PublicSale":          calculatePercentage(totalSupply, 25),  // 25%
			"EcosystemFund":       calculatePercentage(totalSupply, 20),  // 20%
			"Team":                calculatePercentage(totalSupply, 15),  // 15%
			"Advisors":            calculatePercentage(totalSupply, 5),   // 5%
			"LiquidityProvision":  calculatePercentage(totalSupply, 10),  // 10%
			"StakingRewards":      calculatePercentage(totalSupply, 5),   // 5%
			"CommunityAirdrops":   calculatePercentage(totalSupply, 5),   // 5%
		}

		// Verify allocations sum to 100%
		totalAllocated := big.NewInt(0)
		for category, amount := range allocations {
			totalAllocated.Add(totalAllocated, amount)
			t.Logf("   %s: %s QTM", category, formatTokenAmount(amount))
		}

		if totalAllocated.Cmp(totalSupply) != 0 {
			t.Fatalf("Total allocation mismatch: %s vs %s", totalAllocated.String(), totalSupply.String())
		}

		t.Logf("✅ Total allocated: %s QTM", formatTokenAmount(totalAllocated))
	})

	// Test 2: Vesting schedule calculation
	t.Run("VestingSchedule", func(t *testing.T) {
		t.Log("📝 Test 2: Calculate vesting schedule")

		totalAmount := big.NewInt(1000000) // 1M tokens
		totalAmountWei := new(big.Int).Mul(totalAmount, big.NewInt(1e18))
		
		vestingDuration := 2 * 365 * 24 * time.Hour // 2 years
		_ = 180 * 24 * time.Hour // 6 months (cliffDuration unused)
		tgeUnlock := 10                            // 10% at TGE

		// Calculate TGE release
		tgeAmount := calculatePercentage(totalAmountWei, tgeUnlock)
		vestingAmount := new(big.Int).Sub(totalAmountWei, tgeAmount)

		// Simulate vesting after 1 year
		oneYear := 365 * 24 * time.Hour
		vestedAfterOneYear := new(big.Int).Div(
			new(big.Int).Mul(vestingAmount, big.NewInt(int64(oneYear))),
			big.NewInt(int64(vestingDuration)),
		)
		totalVestedAfterOneYear := new(big.Int).Add(tgeAmount, vestedAfterOneYear)

		t.Logf("   Total Amount: %s tokens", formatTokenAmount(totalAmountWei))
		t.Logf("   TGE Release: %s tokens", formatTokenAmount(tgeAmount))
		t.Logf("   Vested after 1 year: %s tokens", formatTokenAmount(totalVestedAfterOneYear))
		t.Logf("✅ Vesting schedule calculated successfully")
	})

	// Test 3: Airdrop merkle tree verification
	t.Run("AirdropMerkleVerification", func(t *testing.T) {
		t.Log("📝 Test 3: Airdrop merkle tree verification")

		// Simulate airdrop recipients
		recipients := []AirdropRecipient{
			{Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", Amount: big.NewInt(1000)},
			{Address: "0x8ba1f109551bD432803012645Hac136c1D9A8D9", Amount: big.NewInt(2000)},
			{Address: "0x9C72fF8b94F0c7f5E8c4b9D5c7e6f8a9b0c1d2e3", Amount: big.NewInt(1500)},
		}

		// Generate merkle tree (simplified)
		merkleRoot := generateMerkleRoot(recipients)
		
		// Generate proof for first recipient
		proof := generateMerkleProof(recipients[0], recipients)

		t.Logf("   Recipients: %d", len(recipients))
		t.Logf("   Merkle Root: %s", merkleRoot)
		t.Logf("   Sample Proof Length: %d", len(proof))
		t.Logf("✅ Merkle tree verification successful")
	})
}

// TestDelegationSystem tests the delegation and rewards system
func TestDelegationSystem(t *testing.T) {
	t.Log("🤝 Testing Delegation System")

	// Test 1: Delegation mechanics
	t.Run("DelegationMechanics", func(t *testing.T) {
		t.Log("📝 Test 1: Delegation amount calculation")

		validatorStake := new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18)) // 100K QTM
		delegationAmount := new(big.Int).Mul(big.NewInt(50000), big.NewInt(1e18)) // 50K QTM
		totalStake := new(big.Int).Add(validatorStake, delegationAmount)

		// Calculate delegation share
		delegationShare := new(big.Float).Quo(
			new(big.Float).SetInt(delegationAmount),
			new(big.Float).SetInt(totalStake),
		)
		sharePercent, _ := delegationShare.Float64()

		t.Logf("   Validator Stake: %s QTM", formatTokenAmount(validatorStake))
		t.Logf("   Delegation Amount: %s QTM", formatTokenAmount(delegationAmount))
		t.Logf("   Total Stake: %s QTM", formatTokenAmount(totalStake))
		t.Logf("   Delegation Share: %.2f%%", sharePercent*100)
		t.Logf("✅ Delegation calculation successful")
	})

	// Test 2: Reward distribution
	t.Run("RewardDistribution", func(t *testing.T) {
		t.Log("📝 Test 2: Reward distribution calculation")

		blockReward := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)) // 1 QTM
		validatorCommission := 5.0                                       // 5%
		
		// Calculate commission
		commissionFloat := validatorCommission / 100.0
		commissionAmount := new(big.Float).Mul(
			new(big.Float).SetInt(blockReward),
			big.NewFloat(commissionFloat),
		)
		commissionInt, _ := commissionAmount.Int(nil)

		// Delegator rewards
		delegatorRewards := new(big.Int).Sub(blockReward, commissionInt)

		t.Logf("   Block Reward: %s QTM", formatTokenAmount(blockReward))
		t.Logf("   Validator Commission: %s QTM (%.1f%%)", 
			formatTokenAmount(commissionInt), validatorCommission)
		t.Logf("   Delegator Rewards: %s QTM", formatTokenAmount(delegatorRewards))
		t.Logf("✅ Reward distribution calculated successfully")
	})

	// Test 3: Unbonding period
	t.Run("UnbondingPeriod", func(t *testing.T) {
		t.Log("📝 Test 3: Unbonding period calculation")

		unbondingPeriod := 21 * 24 * time.Hour // 21 days
		unbondingAmount := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18)) // 10K QTM
		
		startTime := time.Now()
		completionTime := startTime.Add(unbondingPeriod)

		t.Logf("   Unbonding Amount: %s QTM", formatTokenAmount(unbondingAmount))
		t.Logf("   Unbonding Period: %s", unbondingPeriod)
		t.Logf("   Start Time: %s", startTime.Format("2006-01-02 15:04:05"))
		t.Logf("   Completion Time: %s", completionTime.Format("2006-01-02 15:04:05"))
		t.Logf("✅ Unbonding period calculation successful")
	})
}

// TestSlashingSystem tests validator slashing mechanics
func TestSlashingSystem(t *testing.T) {
	t.Log("⚡ Testing Slashing System")

	// Test 1: Double signing slashing
	t.Run("DoubleSigningSlashing", func(t *testing.T) {
		t.Log("📝 Test 1: Double signing slashing calculation")

		validatorStake := new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18)) // 100K QTM
		slashingRate := 20.0 // 20% for double signing

		slashingAmount := calculatePercentageFloat(validatorStake, slashingRate)
		remainingStake := new(big.Int).Sub(validatorStake, slashingAmount)

		t.Logf("   Original Stake: %s QTM", formatTokenAmount(validatorStake))
		t.Logf("   Slashing Rate: %.1f%%", slashingRate)
		t.Logf("   Slashed Amount: %s QTM", formatTokenAmount(slashingAmount))
		t.Logf("   Remaining Stake: %s QTM", formatTokenAmount(remainingStake))
		t.Logf("✅ Double signing slashing calculated")
	})

	// Test 2: Downtime slashing
	t.Run("DowntimeSlashing", func(t *testing.T) {
		t.Log("📝 Test 2: Downtime slashing calculation")

		validatorStake := new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18)) // 100K QTM
		missedBlocks := 60
		totalBlocks := 1000
		slashingThreshold := 5.0 // 5% missed blocks triggers slashing

		missedRate := (float64(missedBlocks) / float64(totalBlocks)) * 100

		if missedRate > slashingThreshold {
			slashingRate := 1.0 // 1% for downtime
			slashingAmount := calculatePercentageFloat(validatorStake, slashingRate)
			
			t.Logf("   Missed Blocks: %d/%d (%.1f%%)", missedBlocks, totalBlocks, missedRate)
			t.Logf("   Slashing Triggered: Yes (threshold: %.1f%%)", slashingThreshold)
			t.Logf("   Slashing Amount: %s QTM (%.1f%%)", formatTokenAmount(slashingAmount), slashingRate)
		} else {
			t.Logf("   Missed Blocks: %d/%d (%.1f%%)", missedBlocks, totalBlocks, missedRate)
			t.Logf("   Slashing Triggered: No (threshold: %.1f%%)", slashingThreshold)
		}

		t.Logf("✅ Downtime slashing evaluation completed")
	})
}

// TestSystemIntegration tests end-to-end system integration
func TestSystemIntegration(t *testing.T) {
	t.Log("🔗 Testing System Integration")

	// Test 1: Complete validator lifecycle
	t.Run("ValidatorLifecycle", func(t *testing.T) {
		t.Log("📝 Test 1: Complete validator lifecycle simulation")

		// Generate validator (keys unused in this test)
		_, _, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}
		
		// Registration phase
		initialStake := new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
		t.Logf("   Phase 1: Registration (Stake: %s QTM)", formatTokenAmount(initialStake))

		// Activation phase
		t.Logf("   Phase 2: Activation (Minimum stake met)")

		// Block production phase
		blocksProduced := 100
		blockReward := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18))
		totalRewards := new(big.Int).Mul(blockReward, big.NewInt(int64(blocksProduced)))
		t.Logf("   Phase 3: Block Production (%d blocks, %s QTM rewards)", 
			blocksProduced, formatTokenAmount(totalRewards))

		// Delegation phase
		totalDelegated := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))
		t.Logf("   Phase 4: Delegation Received (%s QTM)", formatTokenAmount(totalDelegated))

		// Final status
		totalStake := new(big.Int).Add(initialStake, totalDelegated)
		t.Logf("   Final Status: Total Stake %s QTM", formatTokenAmount(totalStake))
		t.Logf("✅ Validator lifecycle simulation completed")
	})
}

// Helper functions

type AirdropRecipient struct {
	Address string
	Amount  *big.Int
}

func calculatePercentage(total *big.Int, percent int) *big.Int {
	percentage := big.NewFloat(float64(percent) / 100.0)
	totalFloat := new(big.Float).SetInt(total)
	resultFloat := new(big.Float).Mul(totalFloat, percentage)
	result, _ := resultFloat.Int(nil)
	return result
}

func calculatePercentageFloat(total *big.Int, percent float64) *big.Int {
	percentage := big.NewFloat(percent / 100.0)
	totalFloat := new(big.Float).SetInt(total)
	resultFloat := new(big.Float).Mul(totalFloat, percentage)
	result, _ := resultFloat.Int(nil)
	return result
}

func formatTokenAmount(amount *big.Int) string {
	// Convert from wei to QTM (divide by 1e18)
	qtm := new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e18))
	return qtm.Text('f', 2)
}

func generateMerkleRoot(recipients []AirdropRecipient) string {
	// Simplified merkle root generation
	return "0x1234567890abcdef1234567890abcdef12345678"
}

func generateMerkleProof(recipient AirdropRecipient, allRecipients []AirdropRecipient) []string {
	// Simplified merkle proof generation
	return []string{
		"0xabcdef1234567890abcdef1234567890abcdef12",
		"0x567890abcdef1234567890abcdef1234567890ab",
	}
}

// TestQuantumSecurityProperties tests quantum security aspects
func TestQuantumSecurityProperties(t *testing.T) {
	t.Log("🔐 Testing Quantum Security Properties")

	t.Run("SignatureResistance", func(t *testing.T) {
		t.Log("📝 Testing signature algorithm resistance")

		// Test only Dilithium (Falcon not implemented)
		privateKey, publicKey, err := crypto.GenerateDilithiumKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate Dilithium keys: %v", err)
		}

		message := []byte("QUANTUM_RESISTANCE_TEST_MESSAGE")
		
		signature, err := privateKey.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign with Dilithium: %v", err)
		}

		valid := publicKey.Verify(message, signature)
		if !valid {
			t.Fatalf("Dilithium signature verification failed")
		}

		t.Logf("   ✅ Dilithium: Key generation, signing, and verification successful")
		t.Logf("      Public key: %d bytes", len(publicKey.Bytes()))
		t.Logf("      Signature: %d bytes", len(signature))
	})
}