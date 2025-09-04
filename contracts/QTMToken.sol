// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./interfaces/IQuantumToken.sol";

/**
 * @title QTMToken
 * @notice Quantum Token (QTM) - Native token for the quantum-resistant blockchain
 * @dev ERC-20 compatible token with minting capabilities for validator rewards
 */
contract QTMToken is IQuantumToken {
    string public name;
    string public symbol;
    uint8 public decimals = 18;
    uint256 private _totalSupply;
    
    address public owner;
    mapping(address => bool) public minters;
    
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event MinterAdded(address indexed minter);
    event MinterRemoved(address indexed minter);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }
    
    modifier onlyMinter() {
        require(minters[msg.sender] || msg.sender == owner, "Not minter");
        _;
    }
    
    constructor(string memory _name, string memory _symbol, uint256 _initialSupply) {
        name = _name;
        symbol = _symbol;
        owner = msg.sender;
        minters[msg.sender] = true;
        
        if (_initialSupply > 0) {
            _totalSupply = _initialSupply;
            _balances[msg.sender] = _initialSupply;
            emit Transfer(address(0), msg.sender, _initialSupply);
        }
    }
    
    function totalSupply() external view override returns (uint256) {
        return _totalSupply;
    }
    
    function balanceOf(address account) external view override returns (uint256) {
        return _balances[account];
    }
    
    function transfer(address to, uint256 amount) external override returns (bool) {
        address from = msg.sender;
        _transfer(from, to, amount);
        return true;
    }
    
    function allowance(address tokenOwner, address spender) external view override returns (uint256) {
        return _allowances[tokenOwner][spender];
    }
    
    function approve(address spender, uint256 amount) external override returns (bool) {
        address tokenOwner = msg.sender;
        _approve(tokenOwner, spender, amount);
        return true;
    }
    
    function transferFrom(address from, address to, uint256 amount) external override returns (bool) {
        address spender = msg.sender;
        uint256 currentAllowance = _allowances[from][spender];
        
        if (currentAllowance != type(uint256).max) {
            require(currentAllowance >= amount, "Insufficient allowance");
            _approve(from, spender, currentAllowance - amount);
        }
        
        _transfer(from, to, amount);
        return true;
    }
    
    function mint(address to, uint256 amount) external onlyMinter returns (bool) {
        require(to != address(0), "Invalid address");
        
        _totalSupply += amount;
        _balances[to] += amount;
        
        emit Transfer(address(0), to, amount);
        return true;
    }
    
    function burn(uint256 amount) external returns (bool) {
        address from = msg.sender;
        require(_balances[from] >= amount, "Insufficient balance");
        
        _balances[from] -= amount;
        _totalSupply -= amount;
        
        emit Transfer(from, address(0), amount);
        return true;
    }
    
    function addMinter(address minter) external onlyOwner {
        require(minter != address(0), "Invalid address");
        minters[minter] = true;
        emit MinterAdded(minter);
    }
    
    function removeMinter(address minter) external onlyOwner {
        minters[minter] = false;
        emit MinterRemoved(minter);
    }
    
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "Invalid address");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }
    
    function _transfer(address from, address to, uint256 amount) internal {
        require(from != address(0), "Invalid sender");
        require(to != address(0), "Invalid recipient");
        require(_balances[from] >= amount, "Insufficient balance");
        
        _balances[from] -= amount;
        _balances[to] += amount;
        
        emit Transfer(from, to, amount);
    }
    
    function _approve(address tokenOwner, address spender, uint256 amount) internal {
        require(tokenOwner != address(0), "Invalid owner");
        require(spender != address(0), "Invalid spender");
        
        _allowances[tokenOwner][spender] = amount;
        emit Approval(tokenOwner, spender, amount);
    }
}