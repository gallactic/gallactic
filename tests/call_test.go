package tests

import (
	"encoding/hex"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultGas uint64 = 100000

/// TODO : test frofor sequence increase. +1 after tx gets successfull. 0: if not successfull. +2: for create contract

func makeCallTx(t *testing.T, from string, addr crypto.Address, data []byte, amt, fee uint64) *tx.CallTx {
	acc := getAccountByName(t, from)
	tx, err := tx.NewCallTx(acc.Address(), addr, acc.Sequence()+1, data, 210000, amt, fee)
	assert.NoError(t, err)

	return tx
}

func TestCallFails(t *testing.T) {
	setPermissions(t, "alice", 0)
	setPermissions(t, "bob", permission.Send)
	setPermissions(t, "carol", permission.Call)
	setPermissions(t, "dan", permission.CreateContract)

	//-------------------
	// call txs
	_, simpleContractAddr := makeContractAccount(t, []byte{0x60}, 0, 0)

	// simple call tx should fail
	tx1 := makeCallTx(t, "alice", simpleContractAddr, nil, 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx1, "alice")

	// simple call tx with send permission should fail
	tx2 := makeCallTx(t, "bob", simpleContractAddr, nil, 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx2, "bob")

	// simple call tx with create permission should fail
	tx3 := makeCallTx(t, "dan", simpleContractAddr, nil, 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx3, "dan")

	//-------------------
	// create txs

	// simple call create tx should fail
	tx4 := makeCallTx(t, "alice", crypto.Address{}, nil, 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx4, "alice")

	// simple call create tx with send perm should fail
	tx5 := makeCallTx(t, "bob", crypto.Address{}, nil, 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx5, "bob")

	// simple call create tx with call perm should fail
	tx6 := makeCallTx(t, "carol", crypto.Address{}, nil, 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx6, "carol")
}

func TestTxSequence(t *testing.T) {
	setPermissions(t, "alice", permission.Send)

	sequence1 := getAccountByName(t, "alice").Sequence()
	sequence2 := getAccountByName(t, "bob").Sequence()
	for i := 0; i < 100; i++ {
		tx := makeSendTx(t, "alice", "bob", 1, _fee)
		signAndExecute(t, e.ErrNone, tx, "alice")
	}

	require.Equal(t, sequence1+100, getAccountByName(t, "alice").Sequence())
	require.Equal(t, sequence2, getAccountByName(t, "bob").Sequence())
}

func TestCallContract(t *testing.T) {
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	//Test Smart Contract
	/*
		pragma solidity ^0.4.24;

		contract SimpleStorage {
			uint private _balance;
			uint private _storedData;
			address private owner;

			event notifyStorage(uint x);

			constructor() public {
				owner = msg.sender;
			}

			function set(uint x) public payable {
				_storedData = x;
				emit notifyStorage(x);
			}

			function get() public view returns (uint) {
				return _storedData;
			}

			function getOwner() public view returns (address){
				return owner;
			}

			function kill() public{
				selfdestruct(owner);
			}
		}
	*/
	// This byte code is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, err := hex.DecodeString("608060405234801561001057600080fd5b5033600260006101000a815481600160a060020a030219169083600160a060020a031602179055506101be806100476000396000f3006080604052600436106100615763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166341c0e1b5811461006657806360fe47b11461007d5780636d4ce63c14610088578063893d20e8146100b0575b600080fd5b34801561007257600080fd5b5061007b610107565b005b61007b60043561012d565b34801561009457600080fd5b5061009d610168565b6040805191825251602090910181900390f35b3480156100bc57600080fd5b506100c561016e565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60025473ffffffffffffffffffffffffffffffffffffffff60006101000a909104811616ff5b60018190556040805182815290517f23f9887eb044d32dba99d7b0b753c61c3c3b72d70ff0addb9a843542fd7642129160200181900390a150565b60015490565b60025460006101000a900473ffffffffffffffffffffffffffffffffffffffff16905600a165627a7a7230582001a5bb7dbc53c4e0e7acc1b23010f4dd1415e0b440e8784ac8ce8d0696c841720029")

	require.NoError(t, err)

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	signAndExecute(t, e.ErrNone, tx1, "alice")

	seq2 := getAccountByName(t, "alice").Sequence()
	// ensure the contract is there
	assert.Equal(t, seq2, seq1+1)

	//TODO: Get Contract Addr from SputnikVM Output
	//contractAddr := ...
	//contractAcc := getAccount(t, contractAddr)
	//require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	/*if !bytes.Equal(contractAcc.Code(), code) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), code)
	}*/
}