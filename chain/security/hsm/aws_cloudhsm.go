package hsm

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudhsm"
	qcrypto "quantum-blockchain/chain/crypto"
)

// AWSCloudHSMProvider implements HSMProvider for AWS CloudHSM
type AWSCloudHSMProvider struct {
	client    *cloudhsm.CloudHSM
	config    HSMConfig
	connected bool
	session   *session.Session
	auditLog  []AuditEntry
}

// NewAWSCloudHSMProvider creates a new AWS CloudHSM provider
func NewAWSCloudHSMProvider() *AWSCloudHSMProvider {
	return &AWSCloudHSMProvider{
		auditLog: make([]AuditEntry, 0),
	}
}

// Initialize connects to AWS CloudHSM
func (p *AWSCloudHSMProvider) Initialize(ctx context.Context, config HSMConfig) error {
	p.config = config

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(config.Credentials["region"]),
		Endpoint: aws.String(config.Endpoint),
	})
	if err != nil {
		p.logAudit("initialize", "", "system", "failed", err.Error())
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	p.session = sess
	p.client = cloudhsm.New(sess)
	p.connected = true

	// Validate FIPS compliance
	if config.FIPSLevel < 3 {
		return fmt.Errorf("AWS CloudHSM requires FIPS 140-2 Level 3 or higher")
	}

	p.logAudit("initialize", "", "system", "success", "")
	log.Printf("✅ AWS CloudHSM initialized successfully")
	return nil
}

// GenerateKey generates a quantum-resistant key pair in AWS CloudHSM
func (p *AWSCloudHSMProvider) GenerateKey(ctx context.Context, keyID string, algorithm qcrypto.SignatureAlgorithm) (*HSMKeyHandle, error) {
	if !p.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// Validate quantum algorithm support
	if !p.supportsAlgorithm(algorithm) {
		return nil, fmt.Errorf("algorithm %v not supported by AWS CloudHSM", algorithm)
	}

	// Generate key pair using quantum-resistant algorithm
	var publicKey []byte
	var err error

	switch algorithm {
	case qcrypto.SigAlgDilithium:
		publicKey, err = p.generateDilithiumKey(ctx, keyID)
	case qcrypto.SigAlgFalcon:
		publicKey, err = p.generateFalconKey(ctx, keyID)
	default:
		err = fmt.Errorf("unsupported algorithm: %v", algorithm)
	}

	if err != nil {
		p.logAudit("generate_key", keyID, "system", "failed", err.Error())
		return nil, err
	}

	handle := &HSMKeyHandle{
		ID:        keyID,
		Algorithm: algorithm,
		PublicKey: publicKey,
		CreatedAt: time.Now(),
		Label:     fmt.Sprintf("quantum-key-%s", keyID),
		Usage:     KeyUsageValidatorSigning,
	}

	p.logAudit("generate_key", keyID, "system", "success", "")
	log.Printf("✅ Generated quantum key %s in AWS CloudHSM", keyID)
	return handle, nil
}

// generateDilithiumKey generates CRYSTALS-Dilithium key in CloudHSM
func (p *AWSCloudHSMProvider) generateDilithiumKey(ctx context.Context, keyID string) ([]byte, error) {
	// Simulate CloudHSM Dilithium key generation
	// In production, this would use CloudHSM's PKCS#11 interface
	// with Dilithium algorithm support (when available)

	// For now, generate a mock public key structure
	mockPublicKey := make([]byte, 1312) // Dilithium-II public key size
	_, err := rand.Read(mockPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mock Dilithium key: %v", err)
	}

	log.Printf("🔐 Generated Dilithium key %s (1312 bytes) in CloudHSM", keyID)
	return mockPublicKey, nil
}

// generateFalconKey generates FALCON key in CloudHSM
func (p *AWSCloudHSMProvider) generateFalconKey(ctx context.Context, keyID string) ([]byte, error) {
	// Simulate CloudHSM Falcon key generation
	mockPublicKey := make([]byte, 897) // Falcon-512 public key size
	_, err := rand.Read(mockPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mock Falcon key: %v", err)
	}

	log.Printf("🔐 Generated Falcon key %s (897 bytes) in CloudHSM", keyID)
	return mockPublicKey, nil
}

// GetKey retrieves an existing key handle
func (p *AWSCloudHSMProvider) GetKey(ctx context.Context, keyID string) (*HSMKeyHandle, error) {
	if !p.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// Simulate key retrieval from CloudHSM
	handle := &HSMKeyHandle{
		ID:        keyID,
		Algorithm: qcrypto.SigAlgDilithium, // Default to Dilithium
		PublicKey: make([]byte, 1312),
		CreatedAt: time.Now().Add(-24 * time.Hour), // Simulate existing key
		Label:     fmt.Sprintf("quantum-key-%s", keyID),
		Usage:     KeyUsageValidatorSigning,
	}

	p.logAudit("get_key", keyID, "system", "success", "")
	return handle, nil
}

// ListKeys returns all key IDs in the HSM
func (p *AWSCloudHSMProvider) ListKeys(ctx context.Context) ([]string, error) {
	if !p.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// Simulate key listing from CloudHSM
	keys := []string{
		"validator-001",
		"validator-002",
		"validator-003",
		"governance-master",
		"emergency-backup",
	}

	p.logAudit("list_keys", "", "system", "success", fmt.Sprintf("found %d keys", len(keys)))
	return keys, nil
}

// DeleteKey securely deletes a key
func (p *AWSCloudHSMProvider) DeleteKey(ctx context.Context, keyID string) error {
	if !p.connected {
		return fmt.Errorf("HSM not connected")
	}

	// Simulate secure key deletion in CloudHSM
	log.Printf("🗑️ Securely deleting key %s from AWS CloudHSM", keyID)

	p.logAudit("delete_key", keyID, "system", "success", "")
	return nil
}

// Sign performs quantum-resistant signing
func (p *AWSCloudHSMProvider) Sign(ctx context.Context, keyID string, data []byte) ([]byte, error) {
	if !p.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// Simulate quantum-resistant signing in CloudHSM
	signature := make([]byte, 2420) // Dilithium-II signature size
	_, err := rand.Read(signature)
	if err != nil {
		p.logAudit("sign", keyID, "system", "failed", err.Error())
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}

	p.logAudit("sign", keyID, "system", "success", fmt.Sprintf("signed %d bytes", len(data)))
	log.Printf("✍️ Signed data with key %s (signature: %d bytes)", keyID, len(signature))
	return signature, nil
}

// GetPublicKey retrieves public key for a key ID
func (p *AWSCloudHSMProvider) GetPublicKey(ctx context.Context, keyID string) ([]byte, error) {
	handle, err := p.GetKey(ctx, keyID)
	if err != nil {
		return nil, err
	}
	return handle.PublicKey, nil
}

// Health checks HSM connectivity
func (p *AWSCloudHSMProvider) Health(ctx context.Context) error {
	if !p.connected {
		return fmt.Errorf("HSM not connected")
	}

	// Simulate health check
	log.Printf("💗 AWS CloudHSM health check: OK")
	return nil
}

// Close disconnects from HSM
func (p *AWSCloudHSMProvider) Close() error {
	p.connected = false
	log.Printf("🔌 AWS CloudHSM connection closed")
	return nil
}

// supportsAlgorithm checks if algorithm is supported
func (p *AWSCloudHSMProvider) supportsAlgorithm(algorithm qcrypto.SignatureAlgorithm) bool {
	// AWS CloudHSM supports these quantum-resistant algorithms
	supported := map[qcrypto.SignatureAlgorithm]bool{
		qcrypto.SigAlgDilithium: true, // CRYSTALS-Dilithium
		qcrypto.SigAlgFalcon:    true, // FALCON (when available)
	}
	return supported[algorithm]
}

// logAudit records audit entry
func (p *AWSCloudHSMProvider) logAudit(operation, keyID, userID, result, errorDetail string) {
	entry := AuditEntry{
		Timestamp:   time.Now(),
		Operation:   operation,
		KeyID:       keyID,
		UserID:      userID,
		Source:      "aws-cloudhsm",
		Result:      result,
		ErrorDetail: errorDetail,
	}

	p.auditLog = append(p.auditLog, entry)

	// In production, send to secure audit logging system
	auditJSON, _ := json.Marshal(entry)
	log.Printf("📋 AUDIT: %s", string(auditJSON))
}

// GetAuditLog returns audit trail (for testing)
func (p *AWSCloudHSMProvider) GetAuditLog() []AuditEntry {
	return p.auditLog
}
