// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract QuantumTest {
    uint256 public value;
    address public owner;
    
    event ValueChanged(uint256 oldValue, uint256 newValue);
    
    constructor(uint256 _initialValue) {
        value = _initialValue;
        owner = msg.sender;
    }
    
    function setValue(uint256 _newValue) public {
        require(msg.sender == owner, "Only owner can set value");
        uint256 oldValue = value;
        value = _newValue;
        emit ValueChanged(oldValue, _newValue);
    }
    
    function getValue() public view returns (uint256) {
        return value;
    }
}