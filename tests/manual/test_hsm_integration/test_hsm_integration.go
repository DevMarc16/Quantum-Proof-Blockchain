package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"quantum-blockchain/chain/security/hsm"
)

func main() {
	fmt.Println("🔐 Testing HSM Integration...")

	// Create HSM manager with enterprise configuration
	config := hsm.HSMManagerConfig{
		DefaultProvider:    "aws-cloudhsm",
		RequiredFIPSLevel:  3,
		AuditRetentionDays: 365,
		BackupEnabled:      true,
		BackupLocation:     "/secure/backups",
		EmergencyContacts:  []string{"security@quantum-blockchain.org"},
		MaxFailedAttempts:  3,
	}

	manager := hsm.NewHSMManager(config)

	// Register AWS CloudHSM provider
	cloudHsmProvider := hsm.NewAWSCloudHSMProvider()

	err := manager.RegisterProvider("aws-cloudhsm", cloudHsmProvider)
	if err != nil {
		log.Fatalf("Failed to register HSM provider: %v", err)
	}

	ctx := context.Background()

	// Validate HSM provider
	validation, err := manager.ValidateProvider(ctx, "aws-cloudhsm")
	if err != nil {
		log.Fatalf("HSM validation failed: %v", err)
	}

	fmt.Printf("✅ HSM Validation Result:\n")
	fmt.Printf("   Status: %s\n", validation.HealthStatus)
	fmt.Printf("   FIPS Compliant: %v\n", validation.FIPSCompliant)
	fmt.Printf("   Supported Algorithms: %v\n", validation.Algorithms)
	fmt.Printf("   Key Capacity: %d/%d\n", validation.CurrentKeys, validation.MaxKeys)

	// Create validator keys for multi-validator network
	validators := []string{"validator-001", "validator-002", "validator-003"}

	for _, validatorID := range validators {
		fmt.Printf("🔑 Creating HSM key for %s...\n", validatorID)

		handle, err := manager.CreateValidatorKey(ctx, validatorID, "aws-cloudhsm")
		if err != nil {
			log.Printf("❌ Failed to create key for %s: %v", validatorID, err)
			continue
		}

		fmt.Printf("   ✅ Key ID: %s\n", handle.ID)
		fmt.Printf("   📊 Algorithm: %v\n", handle.Algorithm)
		fmt.Printf("   🏷️ Usage: %v\n", handle.Usage)
		fmt.Printf("   📅 Created: %v\n", handle.CreatedAt.Format(time.RFC3339))
	}

	// Test key rotation
	fmt.Println("🔄 Testing key rotation...")
	newHandle, err := manager.RotateKey(ctx, "validator-validator-001", "aws-cloudhsm")
	if err != nil {
		log.Printf("❌ Key rotation failed: %v", err)
	} else {
		fmt.Printf("   ✅ New key: %s\n", newHandle.ID)
	}

	// Test backup functionality
	fmt.Println("💾 Testing key backup...")
	err = manager.BackupKey(ctx, "validator-validator-002", "/secure/backups/validator-002")
	if err != nil {
		log.Printf("❌ Backup failed: %v", err)
	}

	// Display audit log
	fmt.Println("📋 Recent HSM operations:")
	auditEntries, err := manager.AuditLog(ctx, "", time.Now().Add(-1*time.Hour))
	if err != nil {
		log.Printf("❌ Failed to get audit log: %v", err)
	} else {
		for _, entry := range auditEntries {
			fmt.Printf("   %s: %s (%s) - %s\n",
				entry.Timestamp.Format("15:04:05"),
				entry.Operation,
				entry.Result,
				entry.KeyID)
		}
	}

	fmt.Println("🎉 HSM Integration Test Complete!")
}
