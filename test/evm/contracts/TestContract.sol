// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

contract TestContract {
    uint256 public value;
    address public lastCaller;
    
    event ValueSet(uint256 newValue, address caller);
    
    function setValue(uint256 _value) public {
        value = _value;
        lastCaller = msg.sender;
        emit ValueSet(_value, msg.sender);
    }
    
    function getValue() public view returns (uint256) {
        return value;
    }
    
    function getLastCaller() public view returns (address) {
        return lastCaller;
    }
    
    // Simple function to test different gas costs
    function expensiveOperation() public {
        for (uint i = 0; i < 100; i++) {
            value = value + 1;
        }
        lastCaller = msg.sender;
    }
}