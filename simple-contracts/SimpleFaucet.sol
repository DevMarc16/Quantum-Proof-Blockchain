// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

interface IERC20 {
    function transfer(address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

contract SimpleFaucet {
    IERC20 public token;
    uint256 public dripAmount = 100 * 10**18; // 100 QTM
    uint256 public validatorAmount = 100000 * 10**18; // 100K QTM
    uint256 public interval = 24 hours;
    
    mapping(address => uint256) public lastRequest;
    mapping(address => bool) public isValidator;
    
    event TokensRequested(address indexed user, uint256 amount);
    event ValidatorRegistered(address indexed validator);
    
    constructor(address _token) {
        token = IERC20(_token);
    }
    
    function requestTokens() external {
        require(block.timestamp >= lastRequest[msg.sender] + interval, "Too soon");
        
        uint256 amount = isValidator[msg.sender] ? validatorAmount : dripAmount;
        require(token.balanceOf(address(this)) >= amount, "Faucet empty");
        
        lastRequest[msg.sender] = block.timestamp;
        require(token.transfer(msg.sender, amount), "Transfer failed");
        
        emit TokensRequested(msg.sender, amount);
    }
    
    function registerValidator() external {
        isValidator[msg.sender] = true;
        emit ValidatorRegistered(msg.sender);
    }
    
    function canRequest(address user) external view returns (bool) {
        return block.timestamp >= lastRequest[user] + interval;
    }
}