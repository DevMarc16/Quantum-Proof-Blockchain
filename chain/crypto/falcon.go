package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
)

// Using ED25519 + Dilithium hybrid approach as Falcon alternative
// This provides both classical and quantum resistance

const (
	FalconPublicKeySize  = ed25519.PublicKeySize + DilithiumPublicKeySize   // Hybrid key
	FalconPrivateKeySize = ed25519.PrivateKeySize + DilithiumPrivateKeySize // Hybrid key
	FalconSignatureSize  = ed25519.SignatureSize + DilithiumSignatureSize   // Hybrid signature
)

type FalconPrivateKey struct {
	ed25519Key   ed25519.PrivateKey
	dilithiumKey *DilithiumPrivateKey
}

type FalconPublicKey struct {
	ed25519Key   ed25519.PublicKey
	dilithiumKey *DilithiumPublicKey
}

// GenerateFalconKeyPair generates a hybrid ED25519+Dilithium key pair for dual security
func GenerateFalconKeyPair() (*FalconPrivateKey, *FalconPublicKey, error) {
	// Generate ED25519 keypair
	ed25519Pub, ed25519Priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ED25519 key: %w", err)
	}
	
	// Generate Dilithium keypair
	dilithiumPriv, dilithiumPub, err := GenerateDilithiumKeyPair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Dilithium key: %w", err)
	}
	
	return &FalconPrivateKey{
			ed25519Key:   ed25519Priv,
			dilithiumKey: dilithiumPriv,
		}, &FalconPublicKey{
			ed25519Key:   ed25519Pub,
			dilithiumKey: dilithiumPub,
		}, nil
}

// Sign creates a hybrid signature using both ED25519 and Dilithium
func (priv *FalconPrivateKey) Sign(message []byte) ([]byte, error) {
	// Create ED25519 signature
	ed25519Sig := ed25519.Sign(priv.ed25519Key, message)
	
	// Create Dilithium signature
	dilithiumSig, err := priv.dilithiumKey.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("dilithium signing failed: %w", err)
	}
	
	// Combine signatures
	signature := make([]byte, 0, FalconSignatureSize)
	signature = append(signature, ed25519Sig...)
	signature = append(signature, dilithiumSig...)
	
	return signature, nil
}

// Verify verifies a hybrid signature
func (pub *FalconPublicKey) Verify(message, signature []byte) bool {
	if len(signature) != FalconSignatureSize {
		return false
	}
	
	// Split signature
	ed25519Sig := signature[:ed25519.SignatureSize]
	dilithiumSig := signature[ed25519.SignatureSize:]
	
	// Verify ED25519 signature
	if !ed25519.Verify(pub.ed25519Key, message, ed25519Sig) {
		return false
	}
	
	// Verify Dilithium signature
	return pub.dilithiumKey.Verify(message, dilithiumSig)
}

// Bytes returns the public key as bytes
func (pub *FalconPublicKey) Bytes() []byte {
	result := make([]byte, 0, FalconPublicKeySize)
	result = append(result, pub.ed25519Key...)
	result = append(result, pub.dilithiumKey.Bytes()...)
	return result
}

// Bytes returns the private key as bytes
func (priv *FalconPrivateKey) Bytes() []byte {
	result := make([]byte, 0, FalconPrivateKeySize)
	result = append(result, priv.ed25519Key...)
	result = append(result, priv.dilithiumKey.Bytes()...)
	return result
}

// FalconPublicKeyFromBytes creates a public key from bytes
func FalconPublicKeyFromBytes(data []byte) (*FalconPublicKey, error) {
	if len(data) != FalconPublicKeySize {
		return nil, errors.New("invalid public key size")
	}
	
	// Split the data
	ed25519Key := data[:ed25519.PublicKeySize]
	dilithiumData := data[ed25519.PublicKeySize:]
	
	// Parse Dilithium key
	dilithiumKey, err := DilithiumPublicKeyFromBytes(dilithiumData)
	if err != nil {
		return nil, fmt.Errorf("invalid dilithium public key: %w", err)
	}
	
	return &FalconPublicKey{
		ed25519Key:   ed25519.PublicKey(ed25519Key),
		dilithiumKey: dilithiumKey,
	}, nil
}

// FalconPrivateKeyFromBytes creates a private key from bytes
func FalconPrivateKeyFromBytes(data []byte) (*FalconPrivateKey, error) {
	if len(data) != FalconPrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	
	// Split the data
	ed25519Key := data[:ed25519.PrivateKeySize]
	dilithiumData := data[ed25519.PrivateKeySize:]
	
	// Parse Dilithium key
	dilithiumKey, err := DilithiumPrivateKeyFromBytes(dilithiumData)
	if err != nil {
		return nil, fmt.Errorf("invalid dilithium private key: %w", err)
	}
	
	return &FalconPrivateKey{
		ed25519Key:   ed25519.PrivateKey(ed25519Key),
		dilithiumKey: dilithiumKey,
	}, nil
}

// VerifyFalcon verifies a hybrid signature given raw bytes
func VerifyFalcon(message, signature, publicKeyBytes []byte) bool {
	pubKey, err := FalconPublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		return false
	}
	
	return pubKey.Verify(message, signature)
}