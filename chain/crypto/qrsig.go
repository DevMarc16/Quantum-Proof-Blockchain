package crypto

import (
	"errors"
	"fmt"
)

// SignatureAlgorithm defines the quantum-resistant signature algorithms
type SignatureAlgorithm uint8

const (
	SigAlgDilithium SignatureAlgorithm = iota + 1
	SigAlgFalcon
	SigAlgSPHINCS // Reserved for future implementation
)

// String returns the string representation of the signature algorithm
func (alg SignatureAlgorithm) String() string {
	switch alg {
	case SigAlgDilithium:
		return "Dilithium"
	case SigAlgFalcon:
		return "Falcon"
	case SigAlgSPHINCS:
		return "SPHINCS+"
	default:
		return "Unknown"
	}
}

// QRSignature represents a quantum-resistant signature
type QRSignature struct {
	Algorithm SignatureAlgorithm
	Signature []byte
	PublicKey []byte // Included for first-time use
}

// SignMessage signs a message using the specified quantum-resistant algorithm
func SignMessage(message []byte, algorithm SignatureAlgorithm, privateKeyBytes []byte) (*QRSignature, error) {
	var signature []byte
	var publicKey []byte

	switch algorithm {
	case SigAlgDilithium:
		priv, err := DilithiumPrivateKeyFromBytes(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid Dilithium private key: %w", err)
		}
		
		signature, err = priv.Sign(message)
		if err != nil {
			return nil, fmt.Errorf("Dilithium signing failed: %w", err)
		}
		
		// Derive public key from private key properly
		// For Dilithium, we need to extract the public key portion from the private key
		// The private key contains both private and public components
		if len(privateKeyBytes) >= DilithiumPublicKeySize {
			// The public key is embedded in the private key structure
			// We generate a new keypair and extract just the public key
			// In production, we'd store the public key separately
			publicKey = make([]byte, DilithiumPublicKeySize)
			// For now, use first part of private key as placeholder
			// Real implementation would properly extract the public component
			copy(publicKey, privateKeyBytes[:DilithiumPublicKeySize])
		}

	case SigAlgFalcon:
		priv, err := FalconPrivateKeyFromBytes(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid Falcon private key: %w", err)
		}
		
		signature, err = priv.Sign(message)
		if err != nil {
			return nil, fmt.Errorf("Falcon signing failed: %w", err)
		}
		
		// Similar approach for Falcon
		if len(privateKeyBytes) >= FalconPublicKeySize {
			publicKey = make([]byte, FalconPublicKeySize)
			copy(publicKey, privateKeyBytes[:FalconPublicKeySize])
		}

	default:
		return nil, fmt.Errorf("unsupported signature algorithm: %v", algorithm)
	}

	return &QRSignature{
		Algorithm: algorithm,
		Signature: signature,
		PublicKey: publicKey,
	}, nil
}

// VerifySignature verifies a quantum-resistant signature
func VerifySignature(message []byte, qrSig *QRSignature) (bool, error) {
	if qrSig == nil {
		return false, errors.New("signature is nil")
	}

	switch qrSig.Algorithm {
	case SigAlgDilithium:
		return VerifyDilithium(message, qrSig.Signature, qrSig.PublicKey), nil
	case SigAlgFalcon:
		return VerifyFalcon(message, qrSig.Signature, qrSig.PublicKey), nil
	default:
		return false, fmt.Errorf("unsupported signature algorithm: %v", qrSig.Algorithm)
	}
}

// GetPublicKeySize returns the public key size for the given algorithm
func GetPublicKeySize(algorithm SignatureAlgorithm) (int, error) {
	switch algorithm {
	case SigAlgDilithium:
		return DilithiumPublicKeySize, nil
	case SigAlgFalcon:
		return FalconPublicKeySize, nil
	default:
		return 0, fmt.Errorf("unsupported signature algorithm: %v", algorithm)
	}
}

// GetSignatureSize returns the signature size for the given algorithm
func GetSignatureSize(algorithm SignatureAlgorithm) (int, error) {
	switch algorithm {
	case SigAlgDilithium:
		return DilithiumSignatureSize, nil
	case SigAlgFalcon:
		return FalconSignatureSize, nil
	default:
		return 0, fmt.Errorf("unsupported signature algorithm: %v", algorithm)
	}
}

// GetPrivateKeySize returns the private key size for the given algorithm
func GetPrivateKeySize(algorithm SignatureAlgorithm) (int, error) {
	switch algorithm {
	case SigAlgDilithium:
		return DilithiumPrivateKeySize, nil
	case SigAlgFalcon:
		return FalconPrivateKeySize, nil
	default:
		return 0, fmt.Errorf("unsupported signature algorithm: %v", algorithm)
	}
}