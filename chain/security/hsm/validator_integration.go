package hsm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"quantum-blockchain/chain/types"
)

// ValidatorHSMService integrates HSM with validator operations
type ValidatorHSMService struct {
	hsmManager  HSMManager
	provider    HSMProvider
	validatorID string
	keyID       string
	handle      *HSMKeyHandle
	mu          sync.RWMutex
	lastUsed    time.Time
	signCount   int64
	initialized bool
}

// ValidatorHSMConfig contains configuration for validator HSM integration
type ValidatorHSMConfig struct {
	ValidatorID     string `json:"validator_id"`
	HSMProvider     string `json:"hsm_provider"`
	KeyRotationDays int    `json:"key_rotation_days"`
	MaxSignatures   int64  `json:"max_signatures"`
	BackupEnabled   bool   `json:"backup_enabled"`
}

// NewValidatorHSMService creates a new validator HSM service
func NewValidatorHSMService(manager HSMManager, config ValidatorHSMConfig) *ValidatorHSMService {
	return &ValidatorHSMService{
		hsmManager:  manager,
		validatorID: config.ValidatorID,
		signCount:   0,
	}
}

// Initialize sets up HSM integration for validator
func (v *ValidatorHSMService) Initialize(ctx context.Context, config ValidatorHSMConfig) error {
	// Get HSM provider
	provider, err := v.hsmManager.GetProvider(config.HSMProvider)
	if err != nil {
		return fmt.Errorf("failed to get HSM provider %s: %v", config.HSMProvider, err)
	}
	v.provider = provider

	// Validate HSM provider
	validation, err := v.hsmManager.ValidateProvider(ctx, config.HSMProvider)
	if err != nil || !validation.Valid {
		return fmt.Errorf("HSM provider validation failed: %v", err)
	}

	// Create or retrieve validator key
	keyID := fmt.Sprintf("validator-%s", config.ValidatorID)
	handle, err := v.getOrCreateValidatorKey(ctx, keyID, config.HSMProvider)
	if err != nil {
		return fmt.Errorf("failed to setup validator key: %v", err)
	}

	v.keyID = keyID
	v.handle = handle
	v.initialized = true
	v.lastUsed = time.Now()

	log.Printf("✅ Validator HSM Service initialized for %s (Key: %s, Algorithm: %v)",
		config.ValidatorID, keyID, handle.Algorithm)

	// Start background key rotation monitoring
	go v.monitorKeyRotation(ctx)

	return nil
}

// SignBlock signs a block using HSM-stored validator key
func (v *ValidatorHSMService) SignBlock(ctx context.Context, block *types.Block) ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.initialized {
		return nil, fmt.Errorf("validator HSM service not initialized")
	}

	// Prepare signing data (block hash)
	signingData := block.Header.Hash().Bytes()

	// Check if key rotation is needed before signing
	if v.needsRotation() {
		log.Printf("⚠️ Key rotation needed for validator %s, but proceeding with current key", v.validatorID)
		// In production: queue key rotation after this signature
	}

	// Sign using HSM
	signature, err := v.provider.Sign(ctx, v.keyID, signingData)
	if err != nil {
		return nil, fmt.Errorf("HSM signing failed: %v", err)
	}

	// Update usage statistics
	v.signCount++
	v.lastUsed = time.Now()

	log.Printf("✍️ Block signed by validator %s using HSM (signature count: %d)",
		v.validatorID, v.signCount)

	return signature, nil
}

// SignTransaction signs a transaction using HSM-stored key
func (v *ValidatorHSMService) SignTransaction(ctx context.Context, tx *types.QuantumTransaction) ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.initialized {
		return nil, fmt.Errorf("validator HSM service not initialized")
	}

	// Get signing hash (excludes signature field)
	signingData := tx.SigningHash().Bytes()

	// Sign using HSM
	signature, err := v.provider.Sign(ctx, v.keyID, signingData)
	if err != nil {
		return nil, fmt.Errorf("HSM transaction signing failed: %v", err)
	}

	// Update usage statistics
	v.signCount++
	v.lastUsed = time.Now()

	log.Printf("✍️ Transaction signed by validator %s using HSM", v.validatorID)
	return signature, nil
}

// GetPublicKey returns the validator's public key from HSM
func (v *ValidatorHSMService) GetPublicKey(ctx context.Context) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.initialized {
		return nil, fmt.Errorf("validator HSM service not initialized")
	}

	return v.provider.GetPublicKey(ctx, v.keyID)
}

// RotateKey performs validator key rotation
func (v *ValidatorHSMService) RotateKey(ctx context.Context, newProvider string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	log.Printf("🔄 Starting key rotation for validator %s", v.validatorID)

	// Perform key rotation through HSM manager
	newHandle, err := v.hsmManager.RotateKey(ctx, v.keyID, newProvider)
	if err != nil {
		return fmt.Errorf("key rotation failed: %v", err)
	}

	// Update service with new key
	oldKeyID := v.keyID
	v.keyID = newHandle.ID
	v.handle = newHandle
	v.signCount = 0 // Reset counter
	v.lastUsed = time.Now()

	log.Printf("🔑 Key rotated for validator %s: %s -> %s",
		v.validatorID, oldKeyID, newHandle.ID)

	return nil
}

// GetKeyInfo returns current key information
func (v *ValidatorHSMService) GetKeyInfo() *HSMKeyHandle {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.initialized {
		return nil
	}

	// Return copy to prevent external modification
	info := *v.handle
	return &info
}

// GetUsageStats returns key usage statistics
func (v *ValidatorHSMService) GetUsageStats() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return map[string]interface{}{
		"validator_id": v.validatorID,
		"key_id":       v.keyID,
		"sign_count":   v.signCount,
		"last_used":    v.lastUsed,
		"key_age":      time.Since(v.handle.CreatedAt),
		"algorithm":    v.handle.Algorithm,
		"initialized":  v.initialized,
	}
}

// Health checks HSM service health
func (v *ValidatorHSMService) Health(ctx context.Context) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.initialized {
		return fmt.Errorf("service not initialized")
	}

	// Check HSM connectivity
	if err := v.provider.Health(ctx); err != nil {
		return fmt.Errorf("HSM health check failed: %v", err)
	}

	// Test key access
	_, err := v.provider.GetKey(ctx, v.keyID)
	if err != nil {
		return fmt.Errorf("key access test failed: %v", err)
	}

	return nil
}

// Close cleanly shuts down HSM service
func (v *ValidatorHSMService) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.provider != nil {
		if err := v.provider.Close(); err != nil {
			return err
		}
	}

	v.initialized = false
	log.Printf("🔌 Validator HSM service closed for %s", v.validatorID)
	return nil
}

// Private helper methods

// getOrCreateValidatorKey gets existing key or creates new one
func (v *ValidatorHSMService) getOrCreateValidatorKey(ctx context.Context, keyID string, providerName string) (*HSMKeyHandle, error) {
	// Try to get existing key first
	provider, err := v.hsmManager.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	handle, err := provider.GetKey(ctx, keyID)
	if err == nil {
		log.Printf("🔑 Using existing validator key: %s", keyID)
		return handle, nil
	}

	// Create new validator key
	log.Printf("🔑 Creating new validator key: %s", keyID)
	return v.hsmManager.CreateValidatorKey(ctx, v.validatorID, providerName)
}

// needsRotation checks if key rotation is needed
func (v *ValidatorHSMService) needsRotation() bool {
	if v.handle == nil {
		return false
	}

	// Check age-based rotation (90 days)
	maxAge := 90 * 24 * time.Hour
	if time.Since(v.handle.CreatedAt) > maxAge {
		return true
	}

	// Check usage-based rotation (1M signatures)
	maxSignatures := int64(1000000)
	if v.signCount > maxSignatures {
		return true
	}

	return false
}

// monitorKeyRotation monitors and triggers automatic key rotation
func (v *ValidatorHSMService) monitorKeyRotation(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.mu.RLock()
			needsRotation := v.needsRotation()
			v.mu.RUnlock()

			if needsRotation {
				log.Printf("⏰ Automatic key rotation triggered for validator %s", v.validatorID)

				// Use same provider for rotation (can be configurable)
				providerName := "aws-cloudhsm" // Default provider
				if err := v.RotateKey(ctx, providerName); err != nil {
					log.Printf("❌ Automatic key rotation failed for validator %s: %v", v.validatorID, err)
				} else {
					log.Printf("✅ Automatic key rotation completed for validator %s", v.validatorID)
				}
			}
		}
	}
}

// ValidatorHSMFactory creates HSM services for validators
type ValidatorHSMFactory struct {
	manager HSMManager
	config  HSMManagerConfig
}

// NewValidatorHSMFactory creates a new validator HSM factory
func NewValidatorHSMFactory(manager HSMManager, config HSMManagerConfig) *ValidatorHSMFactory {
	return &ValidatorHSMFactory{
		manager: manager,
		config:  config,
	}
}

// CreateValidatorHSM creates HSM service for a validator
func (f *ValidatorHSMFactory) CreateValidatorHSM(validatorID string) (*ValidatorHSMService, error) {
	config := ValidatorHSMConfig{
		ValidatorID:     validatorID,
		HSMProvider:     f.config.DefaultProvider,
		KeyRotationDays: 90,
		MaxSignatures:   1000000,
		BackupEnabled:   f.config.BackupEnabled,
	}

	service := NewValidatorHSMService(f.manager, config)
	return service, nil
}
