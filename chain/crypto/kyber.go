package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/kem/kyber/kyber512"
)

const (
	KyberPublicKeySize    = kyber512.PublicKeySize
	KyberPrivateKeySize   = kyber512.PrivateKeySize
	KyberCiphertextSize   = kyber512.CiphertextSize
	KyberSharedSecretSize = kyber512.SharedKeySize
)

type KyberPrivateKey struct {
	privateKey [KyberPrivateKeySize]byte
}

type KyberPublicKey struct {
	publicKey [KyberPublicKeySize]byte
}

// GenerateKyberKeyPair generates a new Kyber key pair using real CRYSTALS-Kyber
func GenerateKyberKeyPair() (*KyberPrivateKey, *KyberPublicKey, error) {
	publicKey, privateKey, err := kyber512.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Kyber key pair: %w", err)
	}

	// Pack keys into arrays
	var privKey KyberPrivateKey
	var pubKey KyberPublicKey

	privateKey.Pack(privKey.privateKey[:])
	publicKey.Pack(pubKey.publicKey[:])

	return &privKey, &pubKey, nil
}

// Encapsulate generates a shared secret and encapsulates it using Kyber KEM
func (pub *KyberPublicKey) Encapsulate() ([]byte, []byte, error) {
	// Unpack public key
	var publicKey kyber512.PublicKey
	publicKey.Unpack(pub.publicKey[:])

	// Generate encapsulated shared secret
	ciphertext := make([]byte, KyberCiphertextSize)
	sharedSecret := make([]byte, KyberSharedSecretSize)

	// Use the Encapsulate function
	publicKey.EncapsulateTo(ciphertext, sharedSecret, nil)

	return ciphertext, sharedSecret, nil
}

// Decapsulate recovers the shared secret from the ciphertext using Kyber KEM
func (priv *KyberPrivateKey) Decapsulate(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) != KyberCiphertextSize {
		return nil, errors.New("invalid ciphertext size")
	}

	// Unpack private key
	var privateKey kyber512.PrivateKey
	privateKey.Unpack(priv.privateKey[:])

	// Decapsulate the shared secret
	sharedSecret := make([]byte, KyberSharedSecretSize)
	privateKey.DecapsulateTo(sharedSecret, ciphertext)

	return sharedSecret, nil
}

// Bytes returns the public key as bytes
func (pub *KyberPublicKey) Bytes() []byte {
	return pub.publicKey[:]
}

// Bytes returns the private key as bytes
func (priv *KyberPrivateKey) Bytes() []byte {
	return priv.privateKey[:]
}

// KyberPublicKeyFromBytes creates a public key from bytes
func KyberPublicKeyFromBytes(data []byte) (*KyberPublicKey, error) {
	if len(data) != KyberPublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	var pubKey KyberPublicKey
	copy(pubKey.publicKey[:], data)

	// Validate by unpacking
	var publicKey kyber512.PublicKey
	publicKey.Unpack(pubKey.publicKey[:])

	return &pubKey, nil
}

// KyberPrivateKeyFromBytes creates a private key from bytes
func KyberPrivateKeyFromBytes(data []byte) (*KyberPrivateKey, error) {
	if len(data) != KyberPrivateKeySize {
		return nil, errors.New("invalid private key size")
	}

	var privKey KyberPrivateKey
	copy(privKey.privateKey[:], data)

	// Validate by unpacking
	var privateKey kyber512.PrivateKey
	privateKey.Unpack(privKey.privateKey[:])

	return &privKey, nil
}

// KyberDecapsulate performs KEM decapsulation given raw bytes
func KyberDecapsulate(ciphertext, privateKeyBytes []byte) ([]byte, error) {
	if len(privateKeyBytes) != KyberPrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	if len(ciphertext) != KyberCiphertextSize {
		return nil, errors.New("invalid ciphertext size")
	}

	// Create private key array
	var privKeyArray [KyberPrivateKeySize]byte
	copy(privKeyArray[:], privateKeyBytes)

	// Unpack private key
	var privateKey kyber512.PrivateKey
	privateKey.Unpack(privKeyArray[:])

	// Decapsulate
	sharedSecret := make([]byte, KyberSharedSecretSize)
	privateKey.DecapsulateTo(sharedSecret, ciphertext)

	return sharedSecret, nil
}
