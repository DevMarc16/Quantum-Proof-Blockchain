package types

import (
	"encoding/json"
	"math/big"
	"time"

	"quantum-blockchain/chain/crypto"
)

// BlockHeader represents the header of a block
type BlockHeader struct {
	ParentHash  Hash     `json:"parentHash"`
	UncleHash   Hash     `json:"sha3Uncles"`
	Coinbase    Address  `json:"miner"`
	Root        Hash     `json:"stateRoot"`
	TxHash      Hash     `json:"transactionsRoot"`
	ReceiptHash Hash     `json:"receiptsRoot"`
	Bloom       []byte   `json:"logsBloom"`
	Difficulty  *big.Int `json:"difficulty"`
	Number      *big.Int `json:"number"`
	GasLimit    uint64   `json:"gasLimit"`
	GasUsed     uint64   `json:"gasUsed"`
	Time        uint64   `json:"timestamp"`
	Extra       []byte   `json:"extraData"`
	MixDigest   Hash     `json:"mixHash"`
	Nonce       uint64   `json:"nonce"`

	// Quantum-specific fields
	ValidatorSig  *crypto.QRSignature `json:"validatorSignature"`
	ValidatorAddr Address             `json:"validatorAddress"`

	// Computed fields
	hash Hash `json:"hash"`
}

// Block represents a complete block
type Block struct {
	Header       *BlockHeader          `json:"header"`
	Transactions []*QuantumTransaction `json:"transactions"`
	Uncles       []*BlockHeader        `json:"uncles"`

	// Computed fields
	size uint64 `json:"size"`
	hash Hash   `json:"hash"`
}

// NewBlockHeader creates a new block header
func NewBlockHeader(parentHash Hash, coinbase Address, root Hash, number *big.Int, gasLimit uint64, time uint64) *BlockHeader {
	return &BlockHeader{
		ParentHash:  parentHash,
		UncleHash:   ZeroHash, // No uncles in PoS
		Coinbase:    coinbase,
		Root:        root,
		TxHash:      ZeroHash,
		ReceiptHash: ZeroHash,
		Bloom:       make([]byte, 256), // Empty bloom filter
		Difficulty:  big.NewInt(0),     // No mining difficulty in PoS
		Number:      number,
		GasLimit:    gasLimit,
		GasUsed:     0,
		Time:        time,
		Extra:       []byte{},
		MixDigest:   ZeroHash,
		Nonce:       0,
	}
}

// NewBlock creates a new block
func NewBlock(header *BlockHeader, transactions []*QuantumTransaction, uncles []*BlockHeader) *Block {
	block := &Block{
		Header:       header,
		Transactions: transactions,
		Uncles:       uncles,
	}

	// Update transaction root
	block.Header.TxHash = block.calculateTxRoot()

	// Calculate size
	block.size = block.calculateSize()

	return block
}

// Hash returns the block hash
func (b *Block) Hash() Hash {
	if !b.hash.IsZero() {
		return b.hash
	}
	return b.Header.Hash()
}

// Hash returns the header hash
func (h *BlockHeader) Hash() Hash {
	if !h.hash.IsZero() {
		return h.hash
	}

	// Create hash data
	data := []byte{}
	data = append(data, h.ParentHash.Bytes()...)
	data = append(data, h.UncleHash.Bytes()...)
	data = append(data, h.Coinbase.Bytes()...)
	data = append(data, h.Root.Bytes()...)
	data = append(data, h.TxHash.Bytes()...)
	data = append(data, h.ReceiptHash.Bytes()...)
	data = append(data, h.Bloom...)
	data = append(data, h.Difficulty.Bytes()...)
	data = append(data, h.Number.Bytes()...)
	data = append(data, uint64ToBytes(h.GasLimit)...)
	data = append(data, uint64ToBytes(h.GasUsed)...)
	data = append(data, uint64ToBytes(h.Time)...)
	data = append(data, h.Extra...)
	data = append(data, h.MixDigest.Bytes()...)
	data = append(data, uint64ToBytes(h.Nonce)...)

	h.hash = BytesToHash(Keccak256(data))
	return h.hash
}

// SigningHash returns the hash used for validator signing
func (h *BlockHeader) SigningHash() Hash {
	// Don't include validator signature in signing hash
	data := []byte{}
	data = append(data, h.ParentHash.Bytes()...)
	data = append(data, h.UncleHash.Bytes()...)
	data = append(data, h.Coinbase.Bytes()...)
	data = append(data, h.Root.Bytes()...)
	data = append(data, h.TxHash.Bytes()...)
	data = append(data, h.ReceiptHash.Bytes()...)
	data = append(data, h.Bloom...)
	data = append(data, h.Difficulty.Bytes()...)
	data = append(data, h.Number.Bytes()...)
	data = append(data, uint64ToBytes(h.GasLimit)...)
	data = append(data, uint64ToBytes(h.GasUsed)...)
	data = append(data, uint64ToBytes(h.Time)...)
	data = append(data, h.Extra...)
	data = append(data, h.MixDigest.Bytes()...)
	data = append(data, uint64ToBytes(h.Nonce)...)

	return BytesToHash(Keccak256(data))
}

// SignBlock signs the block with the validator's private key
func (h *BlockHeader) SignBlock(privateKey []byte, algorithm crypto.SignatureAlgorithm, validatorAddr Address) error {
	sigHash := h.SigningHash()

	qrSig, err := crypto.SignMessage(sigHash.Bytes(), algorithm, privateKey)
	if err != nil {
		return err
	}

	h.ValidatorSig = qrSig
	h.ValidatorAddr = validatorAddr

	// Recompute hash after signing
	h.hash = ZeroHash

	return nil
}

// VerifyValidatorSignature verifies the validator's signature on the block
func (h *BlockHeader) VerifyValidatorSignature() (bool, error) {
	if h.ValidatorSig == nil {
		return false, nil
	}

	sigHash := h.SigningHash()
	return crypto.VerifySignature(sigHash.Bytes(), h.ValidatorSig)
}

// Number returns the block number
func (b *Block) Number() *big.Int {
	return b.Header.Number
}

// Time returns the block timestamp
func (b *Block) Time() uint64 {
	return b.Header.Time
}

// GasLimit returns the gas limit
func (b *Block) GasLimit() uint64 {
	return b.Header.GasLimit
}

// GasUsed returns the gas used
func (b *Block) GasUsed() uint64 {
	return b.Header.GasUsed
}

// Coinbase returns the coinbase address
func (b *Block) Coinbase() Address {
	return b.Header.Coinbase
}

// ParentHash returns the parent hash
func (b *Block) ParentHash() Hash {
	return b.Header.ParentHash
}

// Size returns the block size
func (b *Block) Size() uint64 {
	if b.size == 0 {
		b.size = b.calculateSize()
	}
	return b.size
}

func (b *Block) calculateSize() uint64 {
	size := uint64(0)

	// Header size
	size += 32 * 8 // Hashes
	size += 20 * 2 // Addresses
	size += 256    // Bloom filter
	size += 32 * 2 // Big ints
	size += 8 * 5  // Uint64s
	size += uint64(len(b.Header.Extra))

	if b.Header.ValidatorSig != nil {
		size += 1 // Algorithm
		size += uint64(len(b.Header.ValidatorSig.Signature))
		size += uint64(len(b.Header.ValidatorSig.PublicKey))
	}

	// Transaction sizes
	for _, tx := range b.Transactions {
		size += tx.Size()
	}

	// Uncle headers (should be empty in PoS)
	for _, uncle := range b.Uncles {
		size += 32 * 8 // Basic header size
		size += uint64(len(uncle.Extra))
	}

	return size
}

func (b *Block) calculateTxRoot() Hash {
	if len(b.Transactions) == 0 {
		return ZeroHash
	}

	// Simple Merkle tree implementation
	hashes := make([]Hash, len(b.Transactions))
	for i, tx := range b.Transactions {
		hashes[i] = tx.Hash()
	}

	return calculateMerkleRoot(hashes)
}

// calculateMerkleRoot calculates the Merkle root of a list of hashes
func calculateMerkleRoot(hashes []Hash) Hash {
	if len(hashes) == 0 {
		return ZeroHash
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// Pair up hashes and hash them together
	nextLevel := []Hash{}
	for i := 0; i < len(hashes); i += 2 {
		if i+1 < len(hashes) {
			combined := append(hashes[i].Bytes(), hashes[i+1].Bytes()...)
			nextLevel = append(nextLevel, BytesToHash(Keccak256(combined)))
		} else {
			// Odd number of hashes, duplicate the last one
			combined := append(hashes[i].Bytes(), hashes[i].Bytes()...)
			nextLevel = append(nextLevel, BytesToHash(Keccak256(combined)))
		}
	}

	return calculateMerkleRoot(nextLevel)
}

// MarshalJSON marshals the block to JSON
func (b *Block) MarshalJSON() ([]byte, error) {
	type blockJSON struct {
		Header       *BlockHeader          `json:"header"`
		Transactions []*QuantumTransaction `json:"transactions"`
		Uncles       []*BlockHeader        `json:"uncles"`
		Hash         string                `json:"hash"`
		Size         uint64                `json:"size"`
	}

	return json.Marshal(&blockJSON{
		Header:       b.Header,
		Transactions: b.Transactions,
		Uncles:       b.Uncles,
		Hash:         b.Hash().Hex(),
		Size:         b.Size(),
	})
}

// Genesis creates the genesis block
func Genesis() *Block {
	header := &BlockHeader{
		ParentHash:    ZeroHash,
		UncleHash:     ZeroHash,
		Coinbase:      ZeroAddress,
		Root:          ZeroHash, // Will be set by state initialization
		TxHash:        ZeroHash,
		ReceiptHash:   ZeroHash,
		Bloom:         make([]byte, 256),
		Difficulty:    big.NewInt(0),
		Number:        big.NewInt(0),
		GasLimit:      15000000, // 15M gas limit
		GasUsed:       0,
		Time:          uint64(time.Now().Unix()),
		Extra:         []byte("Quantum-Resistant Blockchain Genesis"),
		MixDigest:     ZeroHash,
		Nonce:         0,
		ValidatorAddr: ZeroAddress,
	}

	return NewBlock(header, []*QuantumTransaction{}, []*BlockHeader{})
}
