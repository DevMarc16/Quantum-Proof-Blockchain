package evm

import (
	"errors"
	"math/big"

	"quantum-blockchain/chain/types"
	
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
)

// SimpleEVM represents a simplified EVM implementation for smart contracts
type SimpleEVM struct {
	stateDB StateInterface
	chainID *big.Int
}

// StateInterface defines the interface for state management
type StateInterface interface {
	GetBalance(addr types.Address) *big.Int
	SetBalance(addr types.Address, balance *big.Int)
	GetNonce(addr types.Address) uint64
	SetNonce(addr types.Address, nonce uint64)
	GetCode(addr types.Address) []byte
	SetCode(addr types.Address, code []byte)
	GetState(addr types.Address, hash types.Hash) types.Hash
	SetState(addr types.Address, hash types.Hash, value types.Hash)
	Exist(addr types.Address) bool
	Empty(addr types.Address) bool
}

// ExecutionResult represents the result of contract execution
type ExecutionResult struct {
	ReturnData      []byte
	GasUsed         uint64
	Err             error
	ContractAddress *types.Address
	Logs            []*etypes.Log
}

// ContractLog represents an event log from contract execution
type ContractLog struct {
	Address types.Address   `json:"address"`
	Topics  []types.Hash    `json:"topics"`
	Data    []byte          `json:"data"`
}

// NewSimpleEVM creates a new simplified EVM
func NewSimpleEVM(stateDB StateInterface, chainID *big.Int) *SimpleEVM {
	return &SimpleEVM{
		stateDB: stateDB,
		chainID: chainID,
	}
}

// ExecuteTransaction executes a quantum transaction using simplified EVM
func (evm *SimpleEVM) ExecuteTransaction(
	tx *types.QuantumTransaction,
	block *types.Block,
	gasLimit uint64,
) (*ExecutionResult, error) {
	from := tx.From()
	
	if tx.IsContractCreation() {
		return evm.executeContractCreation(tx, from, gasLimit)
	} else if tx.GetTo() != nil {
		return evm.executeContractCall(tx, from, *tx.GetTo(), gasLimit)
	} else {
		return nil, errors.New("invalid transaction: no recipient")
	}
}

func (evm *SimpleEVM) executeContractCreation(
	tx *types.QuantumTransaction,
	from types.Address,
	gasLimit uint64,
) (*ExecutionResult, error) {
	// Calculate contract address using deterministic Ethereum-compatible method
	nonce := evm.stateDB.GetNonce(from)
	contractAddr := types.CreateContractAddress(from, nonce)
	
	// Enhanced gas calculation for quantum-resistant contract creation
	gasUsed := uint64(21000) // Base transaction cost
	
	// Data cost: 4 gas per zero byte, 16 gas per non-zero byte (EIP-2028)
	for _, b := range tx.GetData() {
		if b == 0 {
			gasUsed += 4
		} else {
			gasUsed += 16
		}
	}
	
	// Code storage cost: 200 gas per byte stored
	gasUsed += uint64(200) * uint64(len(tx.GetData()))
	
	// Quantum signature verification cost (already paid during validation)
	// No additional cost here since signature was verified in transaction pool
	
	// Additional cost for quantum-resistant contract initialization
	gasUsed += uint64(5000) // Quantum contract setup overhead
	
	if gasUsed > gasLimit {
		return &ExecutionResult{
			GasUsed: gasLimit,
			Err:     errors.New("out of gas during contract creation"),
		}, nil
	}
	
	// Check if sender has sufficient balance for value transfer
	if tx.GetValue().Sign() > 0 {
		fromBalance := evm.stateDB.GetBalance(from)
		if fromBalance.Cmp(tx.GetValue()) < 0 {
			return &ExecutionResult{
				GasUsed: gasUsed,
				Err:     errors.New("insufficient balance for contract creation"),
			}, nil
		}
	}
	
	// Check if contract address already exists
	if evm.stateDB.Exist(contractAddr) && !evm.stateDB.Empty(contractAddr) {
		return &ExecutionResult{
			GasUsed: gasUsed,
			Err:     errors.New("contract address already exists"),
		}, nil
	}
	
	// Increment sender nonce BEFORE storing contract code
	evm.stateDB.SetNonce(from, nonce+1)
	
	// Store the contract code and mark contract as created
	evm.stateDB.SetCode(contractAddr, tx.GetData())
	
	// Initialize contract with zero balance if it doesn't exist
	if !evm.stateDB.Exist(contractAddr) {
		evm.stateDB.SetBalance(contractAddr, big.NewInt(0))
	}
	
	// Transfer value if any
	if tx.GetValue().Sign() > 0 {
		fromBalance := evm.stateDB.GetBalance(from)
		contractBalance := evm.stateDB.GetBalance(contractAddr)
		
		fromBalance.Sub(fromBalance, tx.GetValue())
		contractBalance.Add(contractBalance, tx.GetValue())
		
		evm.stateDB.SetBalance(from, fromBalance)
		evm.stateDB.SetBalance(contractAddr, contractBalance)
	}
	
	// Execute contract constructor and check for quantum precompile calls
	logs, constructorGas := evm.executeContractConstructor(tx.GetData(), contractAddr, from)
	gasUsed += constructorGas
	
	// Final gas check
	if gasUsed > gasLimit {
		return &ExecutionResult{
			GasUsed: gasLimit,
			Err:     errors.New("out of gas during constructor execution"),
		}, nil
	}
	
	return &ExecutionResult{
		ReturnData:      contractAddr.Bytes(),
		GasUsed:         gasUsed,
		ContractAddress: &contractAddr,
		Logs:            logs,
	}, nil
}

func (evm *SimpleEVM) executeContractCall(
	tx *types.QuantumTransaction,
	from types.Address,
	to types.Address,
	gasLimit uint64,
) (*ExecutionResult, error) {
	// Check if target is a contract
	code := evm.stateDB.GetCode(to)
	
	// Basic gas calculation
	gasUsed := uint64(21000) // Base transaction cost
	if len(code) > 0 {
		gasUsed += uint64(len(tx.GetData())) * 4 // Data cost for contract call
	}
	
	if gasUsed > gasLimit {
		return &ExecutionResult{
			GasUsed: gasLimit,
			Err:     errors.New("out of gas"),
		}, nil
	}
	
	// Transfer value if any
	if tx.GetValue().Sign() > 0 {
		fromBalance := evm.stateDB.GetBalance(from)
		toBalance := evm.stateDB.GetBalance(to)
		
		if fromBalance.Cmp(tx.GetValue()) < 0 {
			return &ExecutionResult{
				GasUsed: gasUsed,
				Err:     errors.New("insufficient balance"),
			}, nil
		}
		
		fromBalance.Sub(fromBalance, tx.GetValue())
		toBalance.Add(toBalance, tx.GetValue())
		
		evm.stateDB.SetBalance(from, fromBalance)
		evm.stateDB.SetBalance(to, toBalance)
	}
	
	var returnData []byte
	var logs []*etypes.Log
	
	if len(code) > 0 {
		// Execute contract code (simplified)
		returnData, logs = evm.executeContract(tx.GetData(), to, from)
		
		// Additional gas for contract execution
		gasUsed += uint64(len(code)) / 10 // Simplified gas calculation
	}
	
	return &ExecutionResult{
		ReturnData: returnData,
		GasUsed:    gasUsed,
		Logs:       logs,
	}, nil
}

func (evm *SimpleEVM) executeContract(
	input []byte,
	contract types.Address,
	caller types.Address,
) ([]byte, []*etypes.Log) {
	// Simplified contract execution
	// In a real implementation, this would parse and execute EVM bytecode
	
	var logs []*etypes.Log
	
	// Check for quantum precompile calls
	quantumLogs := evm.executeQuantumPrecompiles(input, contract)
	logs = append(logs, quantumLogs...)
	
	// Return success for now
	return []byte("success"), logs
}

func (evm *SimpleEVM) executeContractConstructor(
	code []byte,
	contract types.Address,
	from types.Address,
) ([]*etypes.Log, uint64) {
	var logs []*etypes.Log
	gasUsed := uint64(0)
	
	// Constructor execution gas - simplified model
	gasUsed += uint64(len(code)) * 2 // Gas for processing constructor code
	
	// Check for quantum precompile calls in constructor
	quantumLogs := evm.executeQuantumPrecompiles(code, contract)
	logs = append(logs, quantumLogs...)
	
	// Add gas cost for any quantum precompile calls
	gasUsed += uint64(len(quantumLogs)) * 1000 // 1000 gas per quantum operation log
	
	return logs, gasUsed
}

func (evm *SimpleEVM) executeQuantumPrecompiles(
	input []byte,
	contract types.Address,
) []*etypes.Log {
	var logs []*etypes.Log
	
	// Check if the input contains calls to quantum precompiles
	// This is a simplified check - in a real implementation,
	// we would parse the EVM bytecode and handle CALL operations
	
	if len(input) >= 4 {
		// Check for quantum precompile signatures
		signature := input[:4]
		
		// Detect quantum precompile calls in bytecode
		if evm.containsQuantumPrecompileCalls(signature) {
			// Create log for quantum precompile usage
			log := &etypes.Log{
				Address: common.BytesToAddress(contract.Bytes()),
				Topics: []common.Hash{
					common.HexToHash("0x" + "quantum_precompile_call"), // Event signature
				},
				Data: signature, // The precompile signature called
			}
			logs = append(logs, log)
		}
		
		// For now, return simplified logs - in production would have full bytecode analysis
		_ = signature
		_ = contract
	}
	
	return logs
}

func (evm *SimpleEVM) containsQuantumPrecompileCalls(bytecode []byte) bool {
	// Simplified detection of quantum precompile calls
	// In production, this would parse EVM bytecode for CALL operations to addresses 0x0a-0x11
	
	if len(bytecode) < 4 {
		return false
	}
	
	// Check for common patterns that might indicate quantum precompile usage
	// This is a simplified heuristic - real implementation would use full bytecode analysis
	for i := 0; i < len(bytecode)-3; i++ {
		// Look for CALL operations to quantum precompile addresses (0x0a - 0x11)
		if bytecode[i] == 0xf1 { // CALL opcode
			// Simplified check - in real implementation would properly decode stack operations
			if i+1 < len(bytecode) && bytecode[i+1] >= 0x0a && bytecode[i+1] <= 0x11 {
				return true
			}
		}
	}
	
	return false
}

func isQuantumSignature(signature []byte, algorithm string) bool {
	// Simplified signature detection
	// In a real implementation, this would check for actual function signatures
	switch algorithm {
	case "dilithium":
		return signature[0] == 0x0a
	case "falcon":
		return signature[0] == 0x0b
	case "kyber":
		return signature[0] == 0x0c
	}
	return false
}

// GetPrecompileAddress returns the address for quantum precompiles
func GetPrecompileAddress(precompile string) types.Address {
	switch precompile {
	case "dilithium":
		return types.BytesToAddress([]byte{0x0a})
	case "falcon":
		return types.BytesToAddress([]byte{0x0b})
	case "kyber":
		return types.BytesToAddress([]byte{0x0c})
	case "sphincs":
		return types.BytesToAddress([]byte{0x0d})
	}
	return types.Address{}
}

// IsPrecompileAddress checks if an address is a quantum precompile
func IsPrecompileAddress(addr types.Address) bool {
	addrBytes := addr.Bytes()
	if len(addrBytes) == 20 {
		// Check if it's one of our quantum precompiles (0x0a - 0x11)
		return addrBytes[19] >= 0x0a && addrBytes[19] <= 0x11
	}
	return false
}