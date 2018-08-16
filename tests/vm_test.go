package tests

import (
	"encoding/hex"
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"testing"

	. "github.com/hyperledger/burrow/execution/evm/asm"
	. "github.com/hyperledger/burrow/execution/evm/asm/bc"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

var defaultGas uint64 = 100000

// convenience function for contract that calls a given address
func callContractCode(contractAddr crypto.Address, amt byte) []byte {
	// calldatacopy into mem and use as input to call
	memOff, inputOff := byte(0x0), byte(0x0)
	value := amt /// amount to transfer
	inOff := byte(0x0)
	retOff, retSize := byte(0x0), byte(0x20)

	// this is the code we want to run (call a contract and return)
	return MustSplice(CALLDATASIZE, PUSH1, inputOff, PUSH1, memOff,
		CALLDATACOPY, PUSH1, retSize, PUSH1, retOff, CALLDATASIZE, PUSH1, inOff,
		PUSH1, value, PUSH20, contractAddr.RawBytes(),
		// Zeno loves us - call with half of the available gas each time we CALL
		PUSH1, 2, GAS, DIV, CALL,
		PUSH1, 32, PUSH1, 0, RETURN)
}

// wrap a contract in create code
func wrapContractForCreateCode(contractCode []byte) []byte {
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	code := append([]byte{0x7f}, binary.RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code
}

// convenience function for contract that is a factory for the code that comes as call data
func createContractCode() []byte {
	// TODO: gas ...

	// calldatacopy the calldatasize
	memOff, inputOff := byte(0x0), byte(0x0)
	contractCode := []byte{0x60, memOff, 0x60, inputOff, 0x36, 0x37}

	// create
	value := byte(0x1)
	contractCode = append(contractCode, []byte{0x60, value, 0x36, 0x60, memOff, 0xf0}...)
	return contractCode
}

func TestCallContract(t *testing.T) {
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	/*
			pragma solidity ^0.4.0;

			contract SimpleStorage {
		                function get() public constant returns (address) {
		        	        return msg.sender;
		    	        }
			}
	*/

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, err := hex.DecodeString("6060604052341561000f57600080fd5b60ca8061001d6000396000f30060606040526004361060" +
		"3f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c14604457" +
		"5b600080fd5b3415604e57600080fd5b60546096565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ff" +
		"ffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050905600a165627a" +
		"7a72305820b9ebf49535372094ae88f56d9ad18f2a79c146c8f56e7ef33b9402924045071e0029")
	require.NoError(t, err)
	require.NoError(t, err)

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	signAndExecute(t, e.ErrNone, tx1, "alice")
	seq2 := getAccountByName(t, "alice").Sequence()
	// ensure the contract is there
	assert.Equal(t, seq2, seq1+1)
	contractAddr := crypto.DeriveContractAddress(tx1.Caller().Address, seq1)

	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	/*if !bytes.Equal(contractAcc.Code(), code) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), code)
	}*/

	// Input is the function hash of `get()`
	input, err := hex.DecodeString("6d4ce63c")
	tx2 := makeCallTx(t, "alice", contractAddr, input, 0, _fee)
	signAndExecute(t, e.ErrNone, tx2, "alice")

}

func TestStorage(t *testing.T) {
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	/*
	pragma solidity ^0.4.18;

	contract EvmTest1{

	    int value;

	    function setVal(int val) public {
	        value = val;
	    }

	    function getVal() public returns(int) {
	        return value;
	    }

	}}
	*/

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, err := hex.DecodeString("608060405234801561001057600080fd5b5060df8061001f6000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680635362f8a214604e578063e1cb0e52146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820eab96fcffb7eb443ef582b370ecc90eee103a7934c43bed47a0ae2b93f9164910029")
	require.NoError(t, err)
	require.NoError(t, err)

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	signAndExecute(t, e.ErrNone, tx1, "alice")
	seq2 := getAccountByName(t, "alice").Sequence()
	// ensure the contract is there
	assert.Equal(t, seq2, seq1+1)
	contractAddr := crypto.DeriveContractAddress(tx1.Caller().Address, seq1)

	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	// Input is the function hash of `setVal()`
	input1, err := hex.DecodeString("5362f8a20000000000000000000000000000000000000000000000000000000000000064")
	tx2 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	signAndExecute(t, e.ErrNone, tx2, "alice")

	// Input is the function hash of `getVal()`
	input2, err := hex.DecodeString("e1cb0e52")
	tx3 := makeCallTx(t, "alice", contractAddr, input2, 0, 100000)
	signAndExecute(t, e.ErrNone, tx3, "alice")
}
