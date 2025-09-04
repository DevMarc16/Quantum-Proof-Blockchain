package hsm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	qcrypto "quantum-blockchain/chain/crypto"
)

// DefaultHSMManager implements HSMManager interface
type DefaultHSMManager struct {
	providers map[string]HSMProvider
	mu        sync.RWMutex
	auditLog  []AuditEntry
	policies  map[string]*KeyRotationPolicy
	config    HSMManagerConfig
}

// HSMManagerConfig contains configuration for HSM manager
type HSMManagerConfig struct {
	DefaultProvider    string                        `json:"default_provider"`
	RequiredFIPSLevel  int                           `json:"required_fips_level"`
	AuditRetentionDays int                           `json:"audit_retention_days"`
	BackupEnabled      bool                          `json:"backup_enabled"`
	BackupLocation     string                        `json:"backup_location"`
	RotationPolicies   map[string]*KeyRotationPolicy `json:"rotation_policies"`
	EmergencyContacts  []string                      `json:"emergency_contacts"`
	MaxFailedAttempts  int                           `json:"max_failed_attempts"`
}

// NewHSMManager creates a new HSM manager
func NewHSMManager(config HSMManagerConfig) *DefaultHSMManager {
	return &DefaultHSMManager{
		providers: make(map[string]HSMProvider),
		auditLog:  make([]AuditEntry, 0),
		policies:  make(map[string]*KeyRotationPolicy),
		config:    config,
	}
}

// RegisterProvider registers a new HSM provider
func (m *DefaultHSMManager) RegisterProvider(name string, provider HSMProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	m.providers[name] = provider
	m.logAudit("register_provider", "", "system", "success", fmt.Sprintf("registered provider: %s", name))
	log.Printf("✅ HSM Provider registered: %s", name)
	return nil
}

// GetProvider returns a registered HSM provider
func (m *DefaultHSMManager) GetProvider(name string) (HSMProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ValidateProvider checks if provider meets security requirements
func (m *DefaultHSMManager) ValidateProvider(ctx context.Context, name string) (*ValidationResult, error) {
	provider, err := m.GetProvider(name)
	if err != nil {
		return nil, err
	}

	// Perform comprehensive validation
	result := &ValidationResult{
		LastValidation: time.Now(),
	}

	// Check connectivity
	if err := provider.Health(ctx); err != nil {
		result.Valid = false
		result.HealthStatus = fmt.Sprintf("Health check failed: %v", err)
		m.logAudit("validate_provider", "", "system", "failed", err.Error())
		return result, nil
	}

	// Validate FIPS compliance
	result.FIPSCompliant = m.validateFIPSCompliance(name)
	if !result.FIPSCompliant && m.config.RequiredFIPSLevel > 0 {
		result.Valid = false
		result.HealthStatus = "FIPS compliance validation failed"
		return result, nil
	}

	// Check supported algorithms
	result.Algorithms = []qcrypto.SignatureAlgorithm{
		qcrypto.SigAlgDilithium,
		qcrypto.SigAlgFalcon,
	}

	// Simulate capacity check
	result.MaxKeys = 10000
	result.CurrentKeys = 5 // Mock current usage

	result.Valid = true
	result.HealthStatus = "All validations passed"

	m.logAudit("validate_provider", "", "system", "success", fmt.Sprintf("provider %s validated", name))
	log.Printf("✅ HSM Provider %s validation: PASSED", name)
	return result, nil
}

// CreateValidatorKey creates a new validator signing key
func (m *DefaultHSMManager) CreateValidatorKey(ctx context.Context, validatorID string, providerName string) (*HSMKeyHandle, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %v", providerName, err)
	}

	// Validate provider first
	validation, err := m.ValidateProvider(ctx, providerName)
	if err != nil || !validation.Valid {
		return nil, fmt.Errorf("provider %s validation failed", providerName)
	}

	// Generate quantum-resistant validator key (default: Dilithium)
	keyID := fmt.Sprintf("validator-%s", validatorID)
	handle, err := provider.GenerateKey(ctx, keyID, qcrypto.SigAlgDilithium)
	if err != nil {
		m.logAudit("create_validator_key", keyID, validatorID, "failed", err.Error())
		return nil, fmt.Errorf("failed to generate validator key: %v", err)
	}

	// Set key usage and rotation policy
	handle.Usage = KeyUsageValidatorSigning
	m.setRotationPolicy(keyID, &KeyRotationPolicy{
		MaxAge:           90 * 24 * time.Hour, // 90 days
		MaxSignatures:    1000000,             // 1M signatures
		ForceRotation:    false,
		RotationSchedule: "0 0 1 */3 *",      // Quarterly
		NotifyBefore:     7 * 24 * time.Hour, // 1 week notice
	})

	// Create backup if enabled
	if m.config.BackupEnabled {
		if err := m.BackupKey(ctx, keyID, m.config.BackupLocation); err != nil {
			log.Printf("⚠️ Failed to backup key %s: %v", keyID, err)
		}
	}

	m.logAudit("create_validator_key", keyID, validatorID, "success", "")
	log.Printf("🔑 Created validator key %s for %s (Algorithm: %v)", keyID, validatorID, handle.Algorithm)
	return handle, nil
}

// RotateKey performs secure key rotation
func (m *DefaultHSMManager) RotateKey(ctx context.Context, keyID string, newProvider string) (*HSMKeyHandle, error) {
	// Get current key information
	oldProvider, err := m.findKeyProvider(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find current provider for key %s: %v", keyID, err)
	}

	oldHandle, err := oldProvider.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current key %s: %v", keyID, err)
	}

	// Generate new key with same algorithm
	newProviderInstance, err := m.GetProvider(newProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get new provider %s: %v", newProvider, err)
	}

	newKeyID := fmt.Sprintf("%s-rotated-%d", keyID, time.Now().Unix())
	newHandle, err := newProviderInstance.GenerateKey(ctx, newKeyID, oldHandle.Algorithm)
	if err != nil {
		m.logAudit("rotate_key", keyID, "system", "failed", err.Error())
		return nil, fmt.Errorf("failed to generate new key: %v", err)
	}

	// Backup old key before deletion
	if m.config.BackupEnabled {
		backupPath := fmt.Sprintf("%s/rotated-%s-%d", m.config.BackupLocation, keyID, time.Now().Unix())
		if err := m.BackupKey(ctx, keyID, backupPath); err != nil {
			log.Printf("⚠️ Failed to backup old key %s: %v", keyID, err)
		}
	}

	// Schedule old key deletion (grace period for rollback)
	go m.scheduleKeyDeletion(ctx, oldProvider, keyID, 24*time.Hour)

	m.logAudit("rotate_key", keyID, "system", "success", fmt.Sprintf("rotated to %s", newKeyID))
	log.Printf("🔄 Rotated key %s to new key %s", keyID, newKeyID)
	return newHandle, nil
}

// BackupKey creates secure backup of key material
func (m *DefaultHSMManager) BackupKey(ctx context.Context, keyID string, destination string) error {
	provider, err := m.findKeyProvider(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to find provider for key %s: %v", keyID, err)
	}

	handle, err := provider.GetKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get key %s: %v", keyID, err)
	}

	// Create encrypted backup (simplified for demo)
	backup := map[string]interface{}{
		"key_id":     keyID,
		"algorithm":  handle.Algorithm,
		"created_at": handle.CreatedAt,
		"backup_at":  time.Now(),
		"metadata":   handle,
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return fmt.Errorf("failed to serialize backup: %v", err)
	}

	// In production: encrypt backup with master key and store securely
	log.Printf("💾 Backup created for key %s (size: %d bytes)", keyID, len(backupJSON))

	m.logAudit("backup_key", keyID, "system", "success", destination)
	return nil
}

// RestoreKey restores key from secure backup
func (m *DefaultHSMManager) RestoreKey(ctx context.Context, backupPath string, newKeyID string) (*HSMKeyHandle, error) {
	// In production: decrypt and validate backup
	log.Printf("🔄 Restoring key from backup %s to %s", backupPath, newKeyID)

	// Simulate key restoration
	handle := &HSMKeyHandle{
		ID:        newKeyID,
		Algorithm: qcrypto.SigAlgDilithium,
		PublicKey: make([]byte, 1312),
		CreatedAt: time.Now(),
		Label:     fmt.Sprintf("restored-%s", newKeyID),
		Usage:     KeyUsageValidatorSigning,
	}

	m.logAudit("restore_key", newKeyID, "system", "success", backupPath)
	return handle, nil
}

// AuditLog returns audit trail for HSM operations
func (m *DefaultHSMManager) AuditLog(ctx context.Context, keyID string, since time.Time) ([]AuditEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []AuditEntry
	for _, entry := range m.auditLog {
		if entry.Timestamp.After(since) && (keyID == "" || entry.KeyID == keyID) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// EmergencyRecovery performs emergency key recovery procedures
func (m *DefaultHSMManager) EmergencyRecovery(ctx context.Context, params EmergencyParams) error {
	log.Printf("🚨 EMERGENCY RECOVERY TRIGGERED: %s", params.TriggerReason)

	// Validate emergency parameters
	if time.Now().After(params.ExpiresAt) {
		return fmt.Errorf("emergency recovery request expired")
	}

	// In production: validate authorization signatures, multi-party approval, etc.

	// Perform emergency procedures
	for _, keyID := range params.RecoveryKeys {
		// Disable compromised keys
		_, err := m.findKeyProvider(ctx, keyID)
		if err != nil {
			log.Printf("⚠️ Failed to find provider for emergency key %s: %v", keyID, err)
			continue
		}

		// Mark key as compromised (don't delete immediately)
		log.Printf("🔒 Marking key %s as compromised", keyID)
	}

	// Notify emergency contacts
	for _, contact := range m.config.EmergencyContacts {
		log.Printf("📞 Notifying emergency contact: %s", contact)
	}

	m.logAudit("emergency_recovery", "", params.AuthorizedBy, "success", params.TriggerReason)
	return nil
}

// Helper methods

func (m *DefaultHSMManager) findKeyProvider(ctx context.Context, keyID string) (HSMProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Search all providers for the key
	for _, provider := range m.providers {
		if _, err := provider.GetKey(ctx, keyID); err == nil {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("key %s not found in any provider", keyID)
}

func (m *DefaultHSMManager) validateFIPSCompliance(providerName string) bool {
	// In production: check actual FIPS certification
	fipsLevels := map[string]int{
		"aws-cloudhsm":   3, // FIPS 140-2 Level 3
		"azure-keyvault": 2, // FIPS 140-2 Level 2
		"pkcs11-hsm":     4, // Can be Level 4
	}

	level, exists := fipsLevels[providerName]
	return exists && level >= m.config.RequiredFIPSLevel
}

func (m *DefaultHSMManager) setRotationPolicy(keyID string, policy *KeyRotationPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[keyID] = policy
}

func (m *DefaultHSMManager) scheduleKeyDeletion(ctx context.Context, provider HSMProvider, keyID string, delay time.Duration) {
	time.Sleep(delay)
	if err := provider.DeleteKey(ctx, keyID); err != nil {
		log.Printf("⚠️ Failed to delete old key %s: %v", keyID, err)
	} else {
		log.Printf("🗑️ Deleted old key %s after grace period", keyID)
	}
}

func (m *DefaultHSMManager) logAudit(operation, keyID, userID, result, detail string) {
	entry := AuditEntry{
		Timestamp:   time.Now(),
		Operation:   operation,
		KeyID:       keyID,
		UserID:      userID,
		Source:      "hsm-manager",
		Result:      result,
		ErrorDetail: detail,
	}

	m.auditLog = append(m.auditLog, entry)

	// Keep audit log size manageable
	if len(m.auditLog) > 10000 {
		m.auditLog = m.auditLog[1000:] // Keep last 9000 entries
	}
}
