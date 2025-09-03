package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/dilithium/mode2"
)

const (
	DilithiumPublicKeySize  = mode2.PublicKeySize
	DilithiumPrivateKeySize = mode2.PrivateKeySize
	DilithiumSignatureSize  = mode2.SignatureSize
)

type DilithiumPrivateKey struct {
	privateKey [DilithiumPrivateKeySize]byte
}

type DilithiumPublicKey struct {
	publicKey [DilithiumPublicKeySize]byte
}

// GenerateDilithiumKeyPair generates a new Dilithium key pair using real CRYSTALS-Dilithium
func GenerateDilithiumKeyPair() (*DilithiumPrivateKey, *DilithiumPublicKey, error) {
	publicKey, privateKey, err := mode2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Dilithium key pair: %w", err)
	}

	// Pack keys into arrays
	var privKey DilithiumPrivateKey
	var pubKey DilithiumPublicKey
	
	privateKey.Pack(&privKey.privateKey)
	publicKey.Pack(&pubKey.publicKey)

	return &privKey, &pubKey, nil
}

// Sign signs a message using Dilithium
func (priv *DilithiumPrivateKey) Sign(message []byte) ([]byte, error) {
	// Unpack private key
	var privateKey mode2.PrivateKey
	privateKey.Unpack(&priv.privateKey)

	// Sign the message
	var signature [DilithiumSignatureSize]byte
	mode2.SignTo(&privateKey, message, signature[:])
	
	return signature[:], nil
}

// Verify verifies a Dilithium signature
func (pub *DilithiumPublicKey) Verify(message, signature []byte) bool {
	if len(signature) != DilithiumSignatureSize {
		return false
	}
	
	// Unpack public key
	var publicKey mode2.PublicKey
	publicKey.Unpack(&pub.publicKey)

	// Verify signature
	return mode2.Verify(&publicKey, message, signature)
}

// Bytes returns the public key as bytes
func (pub *DilithiumPublicKey) Bytes() []byte {
	return pub.publicKey[:]
}

// Bytes returns the private key as bytes
func (priv *DilithiumPrivateKey) Bytes() []byte {
	return priv.privateKey[:]
}

// Public returns the corresponding public key
func (priv *DilithiumPrivateKey) Public() *DilithiumPublicKey {
	var privateKey mode2.PrivateKey
	privateKey.Unpack(&priv.privateKey)
	
	var publicKeyBytes [DilithiumPublicKeySize]byte
	publicKey := privateKey.Public().(*mode2.PublicKey)
	publicKey.Pack(&publicKeyBytes)
	
	return &DilithiumPublicKey{
		publicKey: publicKeyBytes,
	}
}

// DilithiumPublicKeyFromBytes creates a public key from bytes
func DilithiumPublicKeyFromBytes(data []byte) (*DilithiumPublicKey, error) {
	if len(data) != DilithiumPublicKeySize {
		return nil, errors.New("invalid public key size")
	}
	
	var pubKey DilithiumPublicKey
	copy(pubKey.publicKey[:], data)
	
	// Validate by unpacking
	var publicKey mode2.PublicKey
	publicKey.Unpack(&pubKey.publicKey)
	
	return &pubKey, nil
}

// DilithiumPrivateKeyFromBytes creates a private key from bytes
func DilithiumPrivateKeyFromBytes(data []byte) (*DilithiumPrivateKey, error) {
	if len(data) != DilithiumPrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	
	var privKey DilithiumPrivateKey
	copy(privKey.privateKey[:], data)
	
	// Validate by unpacking
	var privateKey mode2.PrivateKey
	privateKey.Unpack(&privKey.privateKey)
	
	return &privKey, nil
}

// VerifyDilithium verifies a Dilithium signature given raw bytes
func VerifyDilithium(message, signature, publicKeyBytes []byte) bool {
	if len(publicKeyBytes) != DilithiumPublicKeySize {
		return false
	}
	if len(signature) != DilithiumSignatureSize {
		return false
	}
	
	// Create public key array
	var pubKeyArray [DilithiumPublicKeySize]byte
	copy(pubKeyArray[:], publicKeyBytes)
	
	// Unpack public key and verify
	var publicKey mode2.PublicKey
	publicKey.Unpack(&pubKeyArray)
	
	return mode2.Verify(&publicKey, message, signature)
}