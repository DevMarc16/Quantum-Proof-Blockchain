package types

import (
	"encoding/json"
	"math/big"

	"quantum-blockchain/chain/crypto"

	"github.com/holiman/uint256"
)

// TransactionType defines the type of quantum-resistant transaction
type TransactionType uint8

const (
	TxTypeQuantum TransactionType = 0x42 // Quantum-resistant transaction type
)

// QuantumTransaction represents a quantum-resistant transaction
type QuantumTransaction struct {
	ChainID   *big.Int               `json:"chainId"`
	Nonce     uint64                 `json:"nonce"`
	GasPrice  *big.Int               `json:"gasPrice"`
	Gas       uint64                 `json:"gas"`
	To        *Address               `json:"to"`
	Value     *big.Int               `json:"value"`
	Data      []byte                 `json:"input"`
	SigAlg    crypto.SignatureAlgorithm `json:"sigAlg"`
	PublicKey []byte                 `json:"publicKey,omitempty"` // Only for first-time use
	Signature []byte                 `json:"signature"`
	KemCapsule []byte                `json:"kemCapsule,omitempty"` // Optional KEM encapsulation
	
	// Computed fields
	hash Hash    `json:"hash"`
	size uint64  `json:"size"`
	from Address `json:"from"`
}

// NewQuantumTransaction creates a new quantum-resistant transaction
func NewQuantumTransaction(chainID *big.Int, nonce uint64, to *Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *QuantumTransaction {
	return &QuantumTransaction{
		ChainID:  chainID,
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       to,
		Value:    value,
		Data:     data,
	}
}

// SignTransaction signs a transaction with the given private key and algorithm
func (tx *QuantumTransaction) SignTransaction(privateKey []byte, algorithm crypto.SignatureAlgorithm) error {
	// Compute transaction hash for signing
	sigHash := tx.SigningHash()
	
	// Sign the hash
	qrSig, err := crypto.SignMessage(sigHash.Bytes(), algorithm, privateKey)
	if err != nil {
		return err
	}
	
	tx.SigAlg = qrSig.Algorithm
	tx.Signature = qrSig.Signature
	tx.PublicKey = qrSig.PublicKey
	
	// Compute sender address
	tx.from = PublicKeyToAddress(tx.PublicKey)
	
	// Compute transaction hash
	tx.hash = tx.Hash()
	
	return nil
}

// VerifySignature verifies the transaction signature
func (tx *QuantumTransaction) VerifySignature() (bool, error) {
	if len(tx.Signature) == 0 || len(tx.PublicKey) == 0 {
		return false, nil
	}
	
	qrSig := &crypto.QRSignature{
		Algorithm: tx.SigAlg,
		Signature: tx.Signature,
		PublicKey: tx.PublicKey,
	}
	
	sigHash := tx.SigningHash()
	return crypto.VerifySignature(sigHash.Bytes(), qrSig)
}

// SigningHash returns the hash used for signing
func (tx *QuantumTransaction) SigningHash() Hash {
	// Create signing data
	data := []byte{}
	data = append(data, tx.ChainID.Bytes()...)
	data = append(data, uint64ToBytes(tx.Nonce)...)
	data = append(data, tx.GasPrice.Bytes()...)
	data = append(data, uint64ToBytes(tx.Gas)...)
	
	if tx.To != nil {
		data = append(data, tx.To.Bytes()...)
	}
	
	data = append(data, tx.Value.Bytes()...)
	data = append(data, tx.Data...)
	
	// Include KEM capsule if present
	if len(tx.KemCapsule) > 0 {
		data = append(data, tx.KemCapsule...)
	}
	
	return BytesToHash(Keccak256(data))
}

// Hash returns the transaction hash
func (tx *QuantumTransaction) Hash() Hash {
	if !tx.hash.IsZero() {
		return tx.hash
	}
	
	// Include signature in hash
	data := []byte{}
	data = append(data, tx.SigningHash().Bytes()...)
	data = append(data, byte(tx.SigAlg))
	data = append(data, tx.Signature...)
	
	tx.hash = BytesToHash(Keccak256(data))
	return tx.hash
}

// From returns the sender address
func (tx *QuantumTransaction) From() Address {
	if tx.from.IsZero() && len(tx.PublicKey) > 0 {
		tx.from = PublicKeyToAddress(tx.PublicKey)
	}
	return tx.from
}

// Size returns the transaction size
func (tx *QuantumTransaction) Size() uint64 {
	if tx.size == 0 {
		tx.size = tx.calculateSize()
	}
	return tx.size
}

func (tx *QuantumTransaction) calculateSize() uint64 {
	size := uint64(0)
	size += 32 // ChainID
	size += 8  // Nonce
	size += 32 // GasPrice
	size += 8  // Gas
	
	if tx.To != nil {
		size += 20 // To address
	}
	
	size += 32                     // Value
	size += uint64(len(tx.Data))   // Data
	size += 1                      // SigAlg
	size += uint64(len(tx.PublicKey)) // PublicKey
	size += uint64(len(tx.Signature)) // Signature
	size += uint64(len(tx.KemCapsule)) // KemCapsule
	
	return size
}

// GasPrice returns the gas price
func (tx *QuantumTransaction) GetGasPrice() *big.Int {
	return new(big.Int).Set(tx.GasPrice)
}

// Gas returns the gas limit
func (tx *QuantumTransaction) GetGas() uint64 {
	return tx.Gas
}

// Value returns the transaction value
func (tx *QuantumTransaction) GetValue() *big.Int {
	return new(big.Int).Set(tx.Value)
}

// To returns the recipient address
func (tx *QuantumTransaction) GetTo() *Address {
	return tx.To
}

// Data returns the transaction data
func (tx *QuantumTransaction) GetData() []byte {
	return tx.Data
}

// Nonce returns the transaction nonce
func (tx *QuantumTransaction) GetNonce() uint64 {
	return tx.Nonce
}

// ChainID returns the chain ID
func (tx *QuantumTransaction) GetChainID() *big.Int {
	return tx.ChainID
}

// IsContractCreation returns true if the transaction creates a contract
func (tx *QuantumTransaction) IsContractCreation() bool {
	return tx.To == nil
}

// MarshalJSON marshals the transaction to JSON
func (tx *QuantumTransaction) MarshalJSON() ([]byte, error) {
	type txJSON struct {
		Hash      string `json:"hash"`
		ChainID   string `json:"chainId"`
		Nonce     string `json:"nonce"`
		GasPrice  string `json:"gasPrice"`
		Gas       string `json:"gas"`
		To        string `json:"to,omitempty"`
		Value     string `json:"value"`
		Data      string `json:"input"`
		SigAlg    uint8  `json:"sigAlg"`
		PublicKey string `json:"publicKey,omitempty"`
		Signature string `json:"signature"`
		KemCapsule string `json:"kemCapsule,omitempty"`
		From      string `json:"from"`
		Size      string `json:"size"`
	}
	
	var toAddr string
	if tx.To != nil {
		toAddr = tx.To.Hex()
	}
	
	// Convert big.Int to uint256.Int for JSON marshaling
	chainID := uint256.NewInt(0)
	chainID.SetFromBig(tx.ChainID)
	nonce := uint256.NewInt(tx.Nonce)
	gasPrice := uint256.NewInt(0)
	gasPrice.SetFromBig(tx.GasPrice)
	gas := uint256.NewInt(tx.Gas)
	value := uint256.NewInt(0)
	value.SetFromBig(tx.Value)
	size := uint256.NewInt(uint64(tx.Size()))
	
	return json.Marshal(&txJSON{
		Hash:      tx.Hash().Hex(),
		ChainID:   chainID.Hex(),
		Nonce:     nonce.Hex(),
		GasPrice:  gasPrice.Hex(),
		Gas:       gas.Hex(),
		To:        toAddr,
		Value:     value.Hex(),
		Data:      "0x" + string(tx.Data),
		SigAlg:    uint8(tx.SigAlg),
		PublicKey: "0x" + string(tx.PublicKey),
		Signature: "0x" + string(tx.Signature),
		KemCapsule: "0x" + string(tx.KemCapsule),
		From:      tx.From().Hex(),
		Size:      size.Hex(),
	})
}

// Helper function to convert uint64 to bytes
func uint64ToBytes(n uint64) []byte {
	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = byte(n & 0xff)
		n >>= 8
	}
	return result
}