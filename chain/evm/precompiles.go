package evm

import (
	"errors"
	"math/big"

	"quantum-blockchain/chain/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// Quantum-resistant precompile addresses
var (
	DilithiumVerifyAddress = common.BytesToAddress([]byte{10})  // 0x0a
	FalconVerifyAddress    = common.BytesToAddress([]byte{11})  // 0x0b
	KyberDecapsAddress     = common.BytesToAddress([]byte{12})  // 0x0c
	SPHINCSVerifyAddress   = common.BytesToAddress([]byte{13})  // 0x0d
)

// Gas costs for quantum-resistant operations
const (
	DilithiumVerifyGas = uint64(50000)  // Higher cost due to large signature
	FalconVerifyGas    = uint64(30000)  // Lower cost, smaller signature
	KyberDecapsGas     = uint64(20000)  // KEM decapsulation cost
	SPHINCSVerifyGas   = uint64(100000) // Highest cost for hash-based signatures
)

// QuantumPrecompiles returns the quantum-resistant precompiled contracts
func QuantumPrecompiles() map[common.Address]vm.PrecompiledContract {
	return map[common.Address]vm.PrecompiledContract{
		DilithiumVerifyAddress: &DilithiumVerify{},
		FalconVerifyAddress:    &FalconVerify{},
		KyberDecapsAddress:     &KyberDecaps{},
		SPHINCSVerifyAddress:   &SPHINCSVerify{},
	}
}

// DilithiumVerify precompiled contract
type DilithiumVerify struct{}

func (c *DilithiumVerify) RequiredGas(input []byte) uint64 {
	return DilithiumVerifyGas
}

func (c *DilithiumVerify) Run(input []byte) ([]byte, error) {
	// Input format: [32 bytes message hash][1312 bytes public key][2420 bytes signature]
	const (
		messageOffset = 0
		messageSize   = 32
		pubkeyOffset  = messageOffset + messageSize
		pubkeySize    = crypto.DilithiumPublicKeySize
		sigOffset     = pubkeyOffset + pubkeySize
		sigSize       = crypto.DilithiumSignatureSize
		totalSize     = messageSize + pubkeySize + sigSize
	)

	if len(input) < totalSize {
		return nil, errors.New("insufficient input data for Dilithium verification")
	}

	message := input[messageOffset : messageOffset+messageSize]
	publicKey := input[pubkeyOffset : pubkeyOffset+pubkeySize]
	signature := input[sigOffset : sigOffset+sigSize]

	valid := crypto.VerifyDilithium(message, signature, publicKey)
	
	result := make([]byte, 32)
	if valid {
		result[31] = 1 // Return 1 if valid, 0 if invalid
	}
	
	return result, nil
}

// FalconVerify precompiled contract
type FalconVerify struct{}

func (c *FalconVerify) RequiredGas(input []byte) uint64 {
	return FalconVerifyGas
}

func (c *FalconVerify) Run(input []byte) ([]byte, error) {
	// Input format: [32 bytes message hash][897 bytes public key][variable signature]
	const (
		messageOffset = 0
		messageSize   = 32
		pubkeyOffset  = messageOffset + messageSize
		pubkeySize    = crypto.FalconPublicKeySize
		sigOffset     = pubkeyOffset + pubkeySize
		minTotalSize  = messageSize + pubkeySize + 1 // At least 1 byte signature
	)

	if len(input) < minTotalSize {
		return nil, errors.New("insufficient input data for Falcon verification")
	}

	message := input[messageOffset : messageOffset+messageSize]
	publicKey := input[pubkeyOffset : pubkeyOffset+pubkeySize]
	signature := input[sigOffset:]

	if len(signature) > crypto.FalconSignatureSize {
		return nil, errors.New("signature too large")
	}

	valid := crypto.VerifyFalcon(message, signature, publicKey)
	
	result := make([]byte, 32)
	if valid {
		result[31] = 1 // Return 1 if valid, 0 if invalid
	}
	
	return result, nil
}

// KyberDecaps precompiled contract (restricted to system use)
type KyberDecaps struct{}

func (c *KyberDecaps) RequiredGas(input []byte) uint64 {
	return KyberDecapsGas
}

func (c *KyberDecaps) Run(input []byte) ([]byte, error) {
	// This precompile is restricted and should only be used by system contracts
	// Input format: [768 bytes ciphertext][1632 bytes private key]
	const (
		ctOffset      = 0
		ctSize        = crypto.KyberCiphertextSize
		privkeyOffset = ctOffset + ctSize
		privkeySize   = crypto.KyberPrivateKeySize
		totalSize     = ctSize + privkeySize
	)

	if len(input) < totalSize {
		return nil, errors.New("insufficient input data for Kyber decapsulation")
	}

	ciphertext := input[ctOffset : ctOffset+ctSize]
	privateKey := input[privkeyOffset : privkeyOffset+privkeySize]

	sharedSecret, err := crypto.KyberDecapsulate(ciphertext, privateKey)
	if err != nil {
		return nil, err
	}

	// Pad to 32 bytes for EVM compatibility
	result := make([]byte, 32)
	copy(result[32-len(sharedSecret):], sharedSecret)
	
	return result, nil
}

// SPHINCSVerify precompiled contract (placeholder for future implementation)
type SPHINCSVerify struct{}

func (c *SPHINCSVerify) RequiredGas(input []byte) uint64 {
	return SPHINCSVerifyGas
}

func (c *SPHINCSVerify) Run(input []byte) ([]byte, error) {
	// SPHINCS+ implementation would go here
	// For now, always return false
	return make([]byte, 32), nil
}

// UpdateQuantumPrecompiles adds quantum precompiles to the existing precompile map
func UpdateQuantumPrecompiles(precompiles map[common.Address]vm.PrecompiledContract) {
	quantumPrecompiles := QuantumPrecompiles()
	for addr, contract := range quantumPrecompiles {
		precompiles[addr] = contract
	}
}

// QuantumChainConfig extends Ethereum's chain config for quantum resistance
type QuantumChainConfig struct {
	*params.ChainConfig
	QuantumBlock *big.Int `json:"quantumBlock,omitempty"` // Block number when quantum precompiles activate
}

// NewQuantumChainConfig creates a new quantum chain configuration
func NewQuantumChainConfig() *QuantumChainConfig {
	return &QuantumChainConfig{
		ChainConfig: &params.ChainConfig{
			ChainID:                 big.NewInt(8888), // Quantum chain ID
			HomesteadBlock:          big.NewInt(0),
			EIP150Block:             big.NewInt(0),
			EIP155Block:             big.NewInt(0),
			EIP158Block:             big.NewInt(0),
			ByzantiumBlock:          big.NewInt(0),
			ConstantinopleBlock:     big.NewInt(0),
			PetersburgBlock:         big.NewInt(0),
			IstanbulBlock:           big.NewInt(0),
			MuirGlacierBlock:        big.NewInt(0),
			BerlinBlock:             big.NewInt(0),
			LondonBlock:             big.NewInt(0),
			ArrowGlacierBlock:       big.NewInt(0),
			GrayGlacierBlock:        big.NewInt(0),
			MergeNetsplitBlock:      big.NewInt(0),
			ShanghaiTime:            new(uint64),
			CancunTime:              new(uint64),
			PragueTime:              new(uint64),
		},
		QuantumBlock: big.NewInt(0), // Activate quantum precompiles from genesis
	}
}