// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title QuantumCrypto
 * @notice Library for quantum-resistant cryptographic operations
 * @dev Provides interfaces to quantum signature verification precompiles
 */
library QuantumCrypto {
    // Precompile addresses for quantum crypto operations
    address constant DILITHIUM_VERIFY = 0x000000000000000000000000000000000000000A;
    address constant FALCON_VERIFY = 0x000000000000000000000000000000000000000b;
    address constant KYBER_KEM_DECAPS = 0x000000000000000000000000000000000000000C;
    address constant SPHINCS_VERIFY = 0x000000000000000000000000000000000000000d;
    
    // Algorithm identifiers
    uint8 constant DILITHIUM_ALGORITHM = 1;
    uint8 constant FALCON_ALGORITHM = 2;
    uint8 constant SPHINCS_ALGORITHM = 3;
    
    // Key and signature size constants
    uint256 constant DILITHIUM_PUBKEY_SIZE = 1312;
    uint256 constant DILITHIUM_SIGNATURE_SIZE = 2420;
    uint256 constant FALCON_PUBKEY_SIZE = 897;
    uint256 constant FALCON_SIGNATURE_SIZE = 690;
    uint256 constant SPHINCS_PUBKEY_SIZE = 32;
    uint256 constant SPHINCS_SIGNATURE_SIZE = 17088;
    
    // Gas costs for quantum operations
    uint256 constant DILITHIUM_GAS_COST = 800;
    uint256 constant FALCON_GAS_COST = 400;
    uint256 constant SPHINCS_GAS_COST = 2000;
    uint256 constant KYBER_KEM_GAS_COST = 600;
    
    /**
     * @notice Verify a Dilithium signature
     * @param pubKey The public key (1312 bytes)
     * @param message The message that was signed
     * @param signature The signature (2420 bytes)
     * @return success True if signature is valid
     */
    function verifyDilithium(
        bytes memory pubKey,
        bytes memory message,
        bytes memory signature
    ) internal view returns (bool success) {
        require(pubKey.length == DILITHIUM_PUBKEY_SIZE, "Invalid Dilithium pubkey size");
        require(signature.length == DILITHIUM_SIGNATURE_SIZE, "Invalid Dilithium signature size");
        
        bytes memory input = abi.encodePacked(pubKey, message, signature);
        uint256 inputSize = input.length;
        
        assembly {
            success := staticcall(
                DILITHIUM_GAS_COST,
                DILITHIUM_VERIFY,
                add(input, 0x20),
                inputSize,
                0,
                0
            )
        }
    }
    
    /**
     * @notice Verify a Falcon signature
     * @param pubKey The public key (897 bytes)
     * @param message The message that was signed
     * @param signature The signature (variable size, ~690 bytes)
     * @return success True if signature is valid
     */
    function verifyFalcon(
        bytes memory pubKey,
        bytes memory message,
        bytes memory signature
    ) internal view returns (bool success) {
        require(pubKey.length == FALCON_PUBKEY_SIZE, "Invalid Falcon pubkey size");
        
        bytes memory input = abi.encodePacked(pubKey, message, signature);
        uint256 inputSize = input.length;
        
        assembly {
            success := staticcall(
                FALCON_GAS_COST,
                FALCON_VERIFY,
                add(input, 0x20),
                inputSize,
                0,
                0
            )
        }
    }
    
    /**
     * @notice Verify a SPHINCS+ signature (for long-term security)
     * @param pubKey The public key (32 bytes)
     * @param message The message that was signed
     * @param signature The signature (17088 bytes)
     * @return success True if signature is valid
     */
    function verifySPHINCS(
        bytes memory pubKey,
        bytes memory message,
        bytes memory signature
    ) internal view returns (bool success) {
        require(pubKey.length == SPHINCS_PUBKEY_SIZE, "Invalid SPHINCS pubkey size");
        require(signature.length == SPHINCS_SIGNATURE_SIZE, "Invalid SPHINCS signature size");
        
        bytes memory input = abi.encodePacked(pubKey, message, signature);
        uint256 inputSize = input.length;
        
        assembly {
            success := staticcall(
                SPHINCS_GAS_COST,
                SPHINCS_VERIFY,
                add(input, 0x20),
                inputSize,
                0,
                0
            )
        }
    }
    
    /**
     * @notice Verify quantum signature based on algorithm
     * @param algorithm Algorithm identifier (1=Dilithium, 2=Falcon, 3=SPHINCS)
     * @param pubKey The public key
     * @param message The message that was signed
     * @param signature The signature
     * @return success True if signature is valid
     */
    function verifyQuantumSignature(
        uint8 algorithm,
        bytes memory pubKey,
        bytes memory message,
        bytes memory signature
    ) internal view returns (bool success) {
        if (algorithm == DILITHIUM_ALGORITHM) {
            return verifyDilithium(pubKey, message, signature);
        } else if (algorithm == FALCON_ALGORITHM) {
            return verifyFalcon(pubKey, message, signature);
        } else if (algorithm == SPHINCS_ALGORITHM) {
            return verifySPHINCS(pubKey, message, signature);
        } else {
            revert("Unsupported quantum algorithm");
        }
    }
    
    /**
     * @notice Kyber KEM decapsulation (for key exchange)
     * @param ciphertext The KEM ciphertext
     * @param secretKey The secret key
     * @return sharedSecret The decapsulated shared secret
     */
    function kyberDecapsulate(
        bytes memory ciphertext,
        bytes memory secretKey
    ) internal view returns (bytes32 sharedSecret) {
        bytes memory input = abi.encodePacked(ciphertext, secretKey);
        uint256 inputSize = input.length;
        
        assembly {
            let success := staticcall(
                KYBER_KEM_GAS_COST,
                KYBER_KEM_DECAPS,
                add(input, 0x20),
                inputSize,
                0x00,
                0x20
            )
            
            if success {
                sharedSecret := mload(0x00)
            }
        }
    }
    
    /**
     * @notice Get gas cost for quantum operation
     * @param algorithm Algorithm identifier
     * @return gasCost Gas cost for the operation
     */
    function getQuantumGasCost(uint8 algorithm) internal pure returns (uint256 gasCost) {
        if (algorithm == DILITHIUM_ALGORITHM) {
            return DILITHIUM_GAS_COST;
        } else if (algorithm == FALCON_ALGORITHM) {
            return FALCON_GAS_COST;
        } else if (algorithm == SPHINCS_ALGORITHM) {
            return SPHINCS_GAS_COST;
        } else {
            revert("Unsupported quantum algorithm");
        }
    }
    
    /**
     * @notice Validate public key format
     * @param algorithm Algorithm identifier
     * @param pubKey Public key bytes
     * @return valid True if public key format is correct
     */
    function validatePublicKey(uint8 algorithm, bytes memory pubKey) internal pure returns (bool valid) {
        if (algorithm == DILITHIUM_ALGORITHM) {
            return pubKey.length == DILITHIUM_PUBKEY_SIZE;
        } else if (algorithm == FALCON_ALGORITHM) {
            return pubKey.length == FALCON_PUBKEY_SIZE;
        } else if (algorithm == SPHINCS_ALGORITHM) {
            return pubKey.length == SPHINCS_PUBKEY_SIZE;
        }
        return false;
    }
    
    /**
     * @notice Validate signature format
     * @param algorithm Algorithm identifier
     * @param signature Signature bytes
     * @return valid True if signature format is correct
     */
    function validateSignature(uint8 algorithm, bytes memory signature) internal pure returns (bool valid) {
        if (algorithm == DILITHIUM_ALGORITHM) {
            return signature.length == DILITHIUM_SIGNATURE_SIZE;
        } else if (algorithm == FALCON_ALGORITHM) {
            // Falcon signatures are variable size, but typically around 690 bytes
            return signature.length >= 600 && signature.length <= 800;
        } else if (algorithm == SPHINCS_ALGORITHM) {
            return signature.length == SPHINCS_SIGNATURE_SIZE;
        }
        return false;
    }
    
    /**
     * @notice Generate a deterministic validator address from quantum public key
     * @param pubKey Quantum public key
     * @return validatorAddress 20-byte address
     */
    function pubKeyToAddress(bytes memory pubKey) internal pure returns (address validatorAddress) {
        bytes32 hash = keccak256(pubKey);
        validatorAddress = address(uint160(uint256(hash)));
    }
    
    /**
     * @notice Create a quantum signature verification message
     * @param validator Validator address
     * @param blockHash Block hash being signed
     * @param blockNumber Block number
     * @return message Message bytes for signing
     */
    function createValidatorSignatureMessage(
        address validator,
        bytes32 blockHash,
        uint256 blockNumber
    ) internal pure returns (bytes memory message) {
        return abi.encodePacked(
            "QUANTUM_VALIDATOR_SIGNATURE",
            validator,
            blockHash,
            blockNumber
        );
    }
    
    /**
     * @notice Create a delegation signature verification message
     * @param delegator Delegator address
     * @param validator Validator address
     * @param amount Delegation amount
     * @param nonce Nonce for replay protection
     * @return message Message bytes for signing
     */
    function createDelegationSignatureMessage(
        address delegator,
        address validator,
        uint256 amount,
        uint256 nonce
    ) internal pure returns (bytes memory message) {
        return abi.encodePacked(
            "QUANTUM_DELEGATION_SIGNATURE",
            delegator,
            validator,
            amount,
            nonce
        );
    }
    
    /**
     * @notice Batch verify multiple quantum signatures
     * @param algorithms Array of algorithm identifiers
     * @param pubKeys Array of public keys
     * @param messages Array of messages
     * @param signatures Array of signatures
     * @return results Array of verification results
     */
    function batchVerifyQuantumSignatures(
        uint8[] memory algorithms,
        bytes[] memory pubKeys,
        bytes[] memory messages,
        bytes[] memory signatures
    ) internal view returns (bool[] memory results) {
        require(
            algorithms.length == pubKeys.length &&
            pubKeys.length == messages.length &&
            messages.length == signatures.length,
            "Array lengths must match"
        );
        
        results = new bool[](algorithms.length);
        
        for (uint256 i = 0; i < algorithms.length; i++) {
            results[i] = verifyQuantumSignature(
                algorithms[i],
                pubKeys[i],
                messages[i],
                signatures[i]
            );
        }
    }
    
    /**
     * @notice Hash function compatible with quantum security (SHA-3)
     * @param data Data to hash
     * @return hash SHA-3 hash
     */
    function quantumSecureHash(bytes memory data) internal pure returns (bytes32 hash) {
        return keccak256(data); // Using Keccak-256 as quantum-secure hash
    }
}