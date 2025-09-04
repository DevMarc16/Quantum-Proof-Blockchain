package hsm

import (
	"context"
	"time"

	qcrypto "quantum-blockchain/chain/crypto"
)

// HSMProvider defines the interface for Hardware Security Module providers
type HSMProvider interface {
	// Initialize connects to the HSM and performs authentication
	Initialize(ctx context.Context, config HSMConfig) error

	// GenerateKey generates a new quantum-resistant key pair in the HSM
	GenerateKey(ctx context.Context, keyID string, algorithm qcrypto.SignatureAlgorithm) (*HSMKeyHandle, error)

	// GetKey retrieves an existing key handle from the HSM
	GetKey(ctx context.Context, keyID string) (*HSMKeyHandle, error)

	// ListKeys returns all key IDs available in the HSM
	ListKeys(ctx context.Context) ([]string, error)

	// DeleteKey securely deletes a key from the HSM
	DeleteKey(ctx context.Context, keyID string) error

	// Sign performs quantum-resistant signing operation using HSM-stored key
	Sign(ctx context.Context, keyID string, data []byte) ([]byte, error)

	// GetPublicKey retrieves the public key for a given key ID
	GetPublicKey(ctx context.Context, keyID string) ([]byte, error)

	// Health checks HSM connectivity and status
	Health(ctx context.Context) error

	// Close cleanly disconnects from the HSM
	Close() error
}

// HSMKeyHandle represents a key stored in the HSM
type HSMKeyHandle struct {
	ID        string                     `json:"id"`
	Algorithm qcrypto.SignatureAlgorithm `json:"algorithm"`
	PublicKey []byte                     `json:"public_key"`
	CreatedAt time.Time                  `json:"created_at"`
	Label     string                     `json:"label"`
	Usage     KeyUsage                   `json:"usage"`
}

// KeyUsage defines the intended use for HSM keys
type KeyUsage int

const (
	KeyUsageValidatorSigning KeyUsage = iota // For validator block signing
	KeyUsageGovernance                       // For governance proposals
	KeyUsageBridge                           // For cross-chain bridge operations
	KeyUsageEmergency                        // For emergency procedures
)

// HSMConfig contains configuration for HSM providers
type HSMConfig struct {
	Provider     string            `json:"provider"`      // "aws-cloudhsm", "azure-keyvault", "pkcs11"
	Endpoint     string            `json:"endpoint"`      // HSM endpoint URL
	Credentials  map[string]string `json:"credentials"`   // Authentication credentials
	Partition    string            `json:"partition"`     // HSM partition/slot
	Pin          string            `json:"pin"`           // HSM PIN/password
	MaxRetries   int               `json:"max_retries"`   // Connection retry limit
	Timeout      time.Duration     `json:"timeout"`       // Operation timeout
	FIPSLevel    int               `json:"fips_level"`    // Required FIPS 140-2 level
	EnableBackup bool              `json:"enable_backup"` // Enable key backup
}

// ValidationResult contains HSM validation results
type ValidationResult struct {
	Valid          bool                         `json:"valid"`
	FIPSCompliant  bool                         `json:"fips_compliant"`
	Algorithms     []qcrypto.SignatureAlgorithm `json:"supported_algorithms"`
	MaxKeys        int                          `json:"max_keys"`
	CurrentKeys    int                          `json:"current_keys"`
	HealthStatus   string                       `json:"health_status"`
	LastValidation time.Time                    `json:"last_validation"`
}

// AuditEntry represents an HSM operation audit log
type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Operation   string    `json:"operation"`
	KeyID       string    `json:"key_id,omitempty"`
	UserID      string    `json:"user_id"`
	Source      string    `json:"source"`
	Result      string    `json:"result"`
	ErrorDetail string    `json:"error_detail,omitempty"`
}

// HSMManager manages multiple HSM providers and key operations
type HSMManager interface {
	// RegisterProvider registers a new HSM provider
	RegisterProvider(name string, provider HSMProvider) error

	// GetProvider returns a registered HSM provider
	GetProvider(name string) (HSMProvider, error)

	// ValidateProvider checks if provider meets security requirements
	ValidateProvider(ctx context.Context, name string) (*ValidationResult, error)

	// CreateValidatorKey creates a new validator signing key
	CreateValidatorKey(ctx context.Context, validatorID string, provider string) (*HSMKeyHandle, error)

	// RotateKey performs secure key rotation
	RotateKey(ctx context.Context, keyID string, newProvider string) (*HSMKeyHandle, error)

	// BackupKey creates secure backup of key material
	BackupKey(ctx context.Context, keyID string, destination string) error

	// RestoreKey restores key from secure backup
	RestoreKey(ctx context.Context, backupPath string, newKeyID string) (*HSMKeyHandle, error)

	// AuditLog returns audit trail for HSM operations
	AuditLog(ctx context.Context, keyID string, since time.Time) ([]AuditEntry, error)

	// EmergencyRecovery performs emergency key recovery procedures
	EmergencyRecovery(ctx context.Context, params EmergencyParams) error
}

// EmergencyParams contains parameters for emergency recovery
type EmergencyParams struct {
	TriggerReason  string    `json:"trigger_reason"`
	RecoveryKeys   []string  `json:"recovery_keys"`
	NewProvider    string    `json:"new_provider"`
	AuthorizedBy   string    `json:"authorized_by"`
	ValidationCode string    `json:"validation_code"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// KeyRotationPolicy defines key rotation requirements
type KeyRotationPolicy struct {
	MaxAge           time.Duration `json:"max_age"`
	MaxSignatures    int64         `json:"max_signatures"`
	ForceRotation    bool          `json:"force_rotation"`
	RotationSchedule string        `json:"rotation_schedule"` // Cron expression
	NotifyBefore     time.Duration `json:"notify_before"`
}
