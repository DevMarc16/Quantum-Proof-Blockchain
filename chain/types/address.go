package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

const (
	AddressLength = 20
	HashLength    = 32
)

// Address represents a 20-byte quantum-resistant address
type Address [AddressLength]byte

// Hash represents a 32-byte hash
type Hash [HashLength]byte

// ZeroAddress represents an empty address
var ZeroAddress = Address{}

// ZeroHash represents an empty hash
var ZeroHash = Hash{}

// BytesToAddress converts bytes to an address
func BytesToAddress(b []byte) Address {
	var addr Address
	if len(b) > AddressLength {
		copy(addr[:], b[len(b)-AddressLength:])
	} else {
		copy(addr[AddressLength-len(b):], b)
	}
	return addr
}

// BytesToHash converts bytes to a hash
func BytesToHash(b []byte) Hash {
	var h Hash
	if len(b) > HashLength {
		copy(h[:], b[len(b)-HashLength:])
	} else {
		copy(h[HashLength-len(b):], b)
	}
	return h
}

// Hex returns the hex representation of the address
func (addr Address) Hex() string {
	return "0x" + hex.EncodeToString(addr[:])
}

// String returns the string representation of the address
func (addr Address) String() string {
	return addr.Hex()
}

// Bytes returns the address as a byte slice
func (addr Address) Bytes() []byte {
	return addr[:]
}

// Equal checks if two addresses are equal
func (addr Address) Equal(other Address) bool {
	return bytes.Equal(addr[:], other[:])
}

// IsZero checks if the address is zero
func (addr Address) IsZero() bool {
	return addr.Equal(ZeroAddress)
}

// Hex returns the hex representation of the hash
func (h Hash) Hex() string {
	return "0x" + hex.EncodeToString(h[:])
}

// String returns the string representation of the hash
func (h Hash) String() string {
	return h.Hex()
}

// Bytes returns the hash as a byte slice
func (h Hash) Bytes() []byte {
	return h[:]
}

// Equal checks if two hashes are equal
func (h Hash) Equal(other Hash) bool {
	return bytes.Equal(h[:], other[:])
}

// IsZero checks if the hash is zero
func (h Hash) IsZero() bool {
	return h.Equal(ZeroHash)
}

// HexToAddress converts a hex string to an address
func HexToAddress(s string) (Address, error) {
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}
	if len(s) != AddressLength*2 {
		return ZeroAddress, fmt.Errorf("invalid address length: expected %d, got %d", AddressLength*2, len(s))
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return ZeroAddress, fmt.Errorf("invalid hex string: %w", err)
	}

	return BytesToAddress(bytes), nil
}

// HexToHash converts a hex string to a hash
func HexToHash(s string) (Hash, error) {
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}
	if len(s) != HashLength*2 {
		return ZeroHash, fmt.Errorf("invalid hash length: expected %d, got %d", HashLength*2, len(s))
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return ZeroHash, fmt.Errorf("invalid hex string: %w", err)
	}

	return BytesToHash(bytes), nil
}

// PublicKeyToAddress derives an address from a quantum-resistant public key
func PublicKeyToAddress(publicKey []byte) Address {
	// Use Keccak256 for EVM compatibility
	hash := Keccak256(publicKey)
	return BytesToAddress(hash[12:]) // Take last 20 bytes
}

// CreateContractAddress creates a contract address from sender and nonce
func CreateContractAddress(sender Address, nonce uint64) Address {
	// Use RLP encoding of sender + nonce to create contract address
	// This is a simplified version of Ethereum's contract address creation
	data := append(sender.Bytes(), byte(nonce))
	return BytesToAddress(Keccak256(data)[12:])
}

// Keccak256 computes the Keccak256 hash
func Keccak256(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// Keccak256Hash computes the Keccak256 hash and returns it as a Hash
func Keccak256Hash(data []byte) Hash {
	return BytesToHash(Keccak256(data))
}

// SHA256 computes the SHA256 hash
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ParseAddress parses an address from string
func ParseAddress(s string) (Address, error) {
	if s == "" {
		return ZeroAddress, errors.New("empty address string")
	}
	return HexToAddress(s)
}

// ParseHash parses a hash from string
func ParseHash(s string) (Hash, error) {
	if s == "" {
		return ZeroHash, errors.New("empty hash string")
	}
	return HexToHash(s)
}

// Uint64ToBytes converts uint64 to bytes (added by consensus package)
var Uint64ToBytes func(uint64) []byte

// BigInt type alias
type BigInt = big.Int

// NewBigInt creates a new big integer
func NewBigInt(x int64) *big.Int {
	return big.NewInt(x)
}

// HexToBytes converts hex string to bytes (added by wallet SDK)
var HexToBytes func(string) ([]byte, error)
