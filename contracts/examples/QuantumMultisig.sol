// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../lib/QuantumVerifier.sol";

/**
 * @title QuantumMultisig
 * @dev A quantum-resistant multisignature wallet
 */
contract QuantumMultisig {
    using QuantumVerifier for bytes32;
    
    // Events
    event Deposit(address indexed sender, uint256 amount, uint256 balance);
    event SubmitTransaction(
        address indexed owner,
        uint256 indexed txIndex,
        address indexed to,
        uint256 value,
        bytes data
    );
    event ConfirmTransaction(address indexed owner, uint256 indexed txIndex);
    event RevokeConfirmation(address indexed owner, uint256 indexed txIndex);
    event ExecuteTransaction(address indexed owner, uint256 indexed txIndex);
    event OwnerAdded(address indexed owner, uint8 algorithm);
    event OwnerRemoved(address indexed owner);
    event RequiredConfirmationsChanged(uint256 required);
    
    // Structures
    struct Owner {
        uint8 algorithm;
        bytes publicKey;
        bool active;
    }
    
    struct Transaction {
        address to;
        uint256 value;
        bytes data;
        bool executed;
        uint256 numConfirmations;
    }
    
    // State variables
    mapping(address => Owner) public owners;
    address[] public ownerList;
    mapping(uint256 => mapping(address => bool)) public isConfirmed;
    
    Transaction[] public transactions;
    uint256 public numConfirmationsRequired;
    
    modifier onlyOwner() {
        require(owners[msg.sender].active, "Not an active owner");
        _;
    }
    
    modifier txExists(uint256 _txIndex) {
        require(_txIndex < transactions.length, "Transaction does not exist");
        _;
    }
    
    modifier notExecuted(uint256 _txIndex) {
        require(!transactions[_txIndex].executed, "Transaction already executed");
        _;
    }
    
    modifier notConfirmed(uint256 _txIndex) {
        require(!isConfirmed[_txIndex][msg.sender], "Transaction already confirmed");
        _;
    }
    
    /**
     * @dev Constructor to initialize the multisig wallet
     * @param _owners Array of owner addresses
     * @param _algorithms Array of signature algorithms for each owner
     * @param _publicKeys Array of public keys for each owner
     * @param _numConfirmationsRequired Number of confirmations required
     */
    constructor(
        address[] memory _owners,
        uint8[] memory _algorithms,
        bytes[] memory _publicKeys,
        uint256 _numConfirmationsRequired
    ) {
        require(_owners.length > 0, "Owners required");
        require(
            _owners.length == _algorithms.length && 
            _algorithms.length == _publicKeys.length,
            "Arrays length mismatch"
        );
        require(
            _numConfirmationsRequired > 0 && 
            _numConfirmationsRequired <= _owners.length,
            "Invalid number of required confirmations"
        );
        
        for (uint256 i = 0; i < _owners.length; i++) {
            address owner = _owners[i];
            uint8 algorithm = _algorithms[i];
            bytes memory publicKey = _publicKeys[i];
            
            require(owner != address(0), "Invalid owner");
            require(QuantumVerifier.isAlgorithmSupported(algorithm), "Unsupported algorithm");
            require(!owners[owner].active, "Owner not unique");
            require(
                publicKey.length == QuantumVerifier.getPublicKeyLength(algorithm),
                "Invalid public key length"
            );
            
            // Verify that the address matches the public key
            address derivedAddress = address(uint160(uint256(keccak256(publicKey))));
            require(derivedAddress == owner, "Address does not match public key");
            
            owners[owner] = Owner({
                algorithm: algorithm,
                publicKey: publicKey,
                active: true
            });
            ownerList.push(owner);
            
            emit OwnerAdded(owner, algorithm);
        }
        
        numConfirmationsRequired = _numConfirmationsRequired;
        emit RequiredConfirmationsChanged(_numConfirmationsRequired);
    }
    
    /**
     * @dev Fallback function to receive ETH
     */
    receive() external payable {
        emit Deposit(msg.sender, msg.value, address(this).balance);
    }
    
    /**
     * @dev Submit a transaction for confirmation
     * @param _to Destination address
     * @param _value Amount to send
     * @param _data Transaction data
     * @param _messageHash Hash of the transaction message
     * @param _signature Quantum signature
     */
    function submitTransaction(
        address _to,
        uint256 _value,
        bytes memory _data,
        bytes32 _messageHash,
        bytes memory _signature
    ) public {
        // Verify quantum signature
        Owner memory owner = owners[msg.sender];
        require(owner.active, "Not an active owner");
        
        bool isValid = QuantumVerifier.verifySignature(
            owner.algorithm,
            _messageHash,
            _signature,
            owner.publicKey
        );
        require(isValid, "Invalid quantum signature");
        
        // Verify message hash matches transaction data
        bytes32 expectedHash = keccak256(abi.encodePacked(_to, _value, _data));
        require(_messageHash == expectedHash, "Message hash mismatch");
        
        uint256 txIndex = transactions.length;
        
        transactions.push(Transaction({
            to: _to,
            value: _value,
            data: _data,
            executed: false,
            numConfirmations: 0
        }));
        
        emit SubmitTransaction(msg.sender, txIndex, _to, _value, _data);
    }
    
    /**
     * @dev Confirm a transaction using quantum signature
     * @param _txIndex Transaction index
     * @param _messageHash Hash of the confirmation message
     * @param _signature Quantum signature
     */
    function confirmTransaction(
        uint256 _txIndex,
        bytes32 _messageHash,
        bytes memory _signature
    ) public onlyOwner txExists(_txIndex) notExecuted(_txIndex) notConfirmed(_txIndex) {
        // Verify quantum signature
        Owner memory owner = owners[msg.sender];
        
        bool isValid = QuantumVerifier.verifySignature(
            owner.algorithm,
            _messageHash,
            _signature,
            owner.publicKey
        );
        require(isValid, "Invalid quantum signature");
        
        // Verify message hash includes transaction index and sender
        bytes32 expectedHash = keccak256(abi.encodePacked(_txIndex, msg.sender, "confirm"));
        require(_messageHash == expectedHash, "Message hash mismatch");
        
        Transaction storage transaction = transactions[_txIndex];
        transaction.numConfirmations += 1;
        isConfirmed[_txIndex][msg.sender] = true;
        
        emit ConfirmTransaction(msg.sender, _txIndex);
    }
    
    /**
     * @dev Execute a confirmed transaction
     * @param _txIndex Transaction index
     */
    function executeTransaction(uint256 _txIndex) 
        public onlyOwner txExists(_txIndex) notExecuted(_txIndex) {
        Transaction storage transaction = transactions[_txIndex];
        
        require(
            transaction.numConfirmations >= numConfirmationsRequired,
            "Cannot execute transaction"
        );
        
        transaction.executed = true;
        
        (bool success, ) = transaction.to.call{value: transaction.value}(
            transaction.data
        );
        require(success, "Transaction failed");
        
        emit ExecuteTransaction(msg.sender, _txIndex);
    }
    
    /**
     * @dev Revoke confirmation for a transaction
     * @param _txIndex Transaction index
     * @param _messageHash Hash of the revocation message
     * @param _signature Quantum signature
     */
    function revokeConfirmation(
        uint256 _txIndex,
        bytes32 _messageHash,
        bytes memory _signature
    ) public onlyOwner txExists(_txIndex) notExecuted(_txIndex) {
        require(isConfirmed[_txIndex][msg.sender], "Transaction not confirmed");
        
        // Verify quantum signature
        Owner memory owner = owners[msg.sender];
        
        bool isValid = QuantumVerifier.verifySignature(
            owner.algorithm,
            _messageHash,
            _signature,
            owner.publicKey
        );
        require(isValid, "Invalid quantum signature");
        
        // Verify message hash includes transaction index and sender
        bytes32 expectedHash = keccak256(abi.encodePacked(_txIndex, msg.sender, "revoke"));
        require(_messageHash == expectedHash, "Message hash mismatch");
        
        Transaction storage transaction = transactions[_txIndex];
        transaction.numConfirmations -= 1;
        isConfirmed[_txIndex][msg.sender] = false;
        
        emit RevokeConfirmation(msg.sender, _txIndex);
    }
    
    /**
     * @dev Add a new owner (requires multisig approval)
     * @param _owner New owner address
     * @param _algorithm Signature algorithm for new owner
     * @param _publicKey Public key for new owner
     */
    function addOwner(
        address _owner,
        uint8 _algorithm,
        bytes memory _publicKey
    ) public {
        require(msg.sender == address(this), "Can only be called via multisig");
        require(_owner != address(0), "Invalid owner");
        require(!owners[_owner].active, "Owner already exists");
        require(QuantumVerifier.isAlgorithmSupported(_algorithm), "Unsupported algorithm");
        require(
            _publicKey.length == QuantumVerifier.getPublicKeyLength(_algorithm),
            "Invalid public key length"
        );
        
        // Verify that the address matches the public key
        address derivedAddress = address(uint160(uint256(keccak256(_publicKey))));
        require(derivedAddress == _owner, "Address does not match public key");
        
        owners[_owner] = Owner({
            algorithm: _algorithm,
            publicKey: _publicKey,
            active: true
        });
        ownerList.push(_owner);
        
        emit OwnerAdded(_owner, _algorithm);
    }
    
    /**
     * @dev Remove an owner (requires multisig approval)
     * @param _owner Owner address to remove
     */
    function removeOwner(address _owner) public {
        require(msg.sender == address(this), "Can only be called via multisig");
        require(owners[_owner].active, "Owner does not exist");
        require(ownerList.length > numConfirmationsRequired, "Cannot remove owner");
        
        owners[_owner].active = false;
        
        // Remove from owner list
        for (uint256 i = 0; i < ownerList.length; i++) {
            if (ownerList[i] == _owner) {
                ownerList[i] = ownerList[ownerList.length - 1];
                ownerList.pop();
                break;
            }
        }
        
        emit OwnerRemoved(_owner);
    }
    
    /**
     * @dev Change the number of required confirmations
     * @param _numConfirmationsRequired New number of required confirmations
     */
    function changeRequirement(uint256 _numConfirmationsRequired) public {
        require(msg.sender == address(this), "Can only be called via multisig");
        require(
            _numConfirmationsRequired > 0 && 
            _numConfirmationsRequired <= ownerList.length,
            "Invalid number of required confirmations"
        );
        
        numConfirmationsRequired = _numConfirmationsRequired;
        emit RequiredConfirmationsChanged(_numConfirmationsRequired);
    }
    
    // View functions
    
    function getOwners() public view returns (address[] memory) {
        return ownerList;
    }
    
    function getTransactionCount() public view returns (uint256) {
        return transactions.length;
    }
    
    function getTransaction(uint256 _txIndex) public view returns (
        address to,
        uint256 value,
        bytes memory data,
        bool executed,
        uint256 numConfirmations
    ) {
        Transaction storage transaction = transactions[_txIndex];
        
        return (
            transaction.to,
            transaction.value,
            transaction.data,
            transaction.executed,
            transaction.numConfirmations
        );
    }
    
    function getOwnerInfo(address _owner) public view returns (
        uint8 algorithm,
        bytes memory publicKey,
        bool active
    ) {
        Owner memory owner = owners[_owner];
        return (owner.algorithm, owner.publicKey, owner.active);
    }
}