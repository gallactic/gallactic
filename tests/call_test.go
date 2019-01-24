package tests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
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

func TestCallAndCreateContractPermissions(t *testing.T) {
	/*
	   pragma solidity ^0.4.18;

	   contract Factory {
	       function Create(bytes code) public returns (address addr) {
	           assembly {
	               addr := create(0,add(code,0x20), mload(code))
	           }
	       }
	   }

	   contract Adder {
	       function add(uint a, uint b) public pure returns (uint){
	           return a+b;
	       }
	   }

	   contract Tester {
	       Adder a;

	       function Tester(address factory, bytes code) public {
	           a = Adder(Factory(factory).Create(code));
	           if(address(a) == 0) throw;
	       }

	       function AdderAddr() public constant returns (address){
	           return a;
	       }

	       function AdderAdd(uint x, uint y) public view returns (uint){
	           return a.add(x,y);
	       }
	   }
	*/

	// bytecodes
	var factory = "6060604052341561000f57600080fd5b61011c8061001e6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806377e5484d146044575b600080fd5b3415604e57600080fd5b609c600480803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509190505060de565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008151602083016000f090509190505600a165627a7a7230582013fa62d6f3101082c5e00400929901b7354193b6db053056b4d67b84f143cdec0029"
	var adder = "6060604052341561000f57600080fd5b60ba8061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063771602f7146044575b600080fd5b3415604e57600080fd5b606b60048080359060200190919080359060200190919050506081565b6040518082815260200191505060405180910390f35b60008183019050929150505600a165627a7a723058208a946f6d7a6a422f86f3645f6e41512625187c0713b3718c4f670a65d245d3230029"
	var tester = "6060604052341561000f57600080fd5b6040516103c03803806103c0833981016040528080519060200190919080518201919050508173ffffffffffffffffffffffffffffffffffffffff166377e5484d826000604051602001526040518263ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825283818151815260200191508051906020019080838360005b838110156100c55780820151818401526020810190506100aa565b50505050905090810190601f1680156100f25780820380516001836020036101000a031916815260200191505b5092505050602060405180830381600087803b151561011057600080fd5b6102c65a03f1151561012157600080fd5b505050604051805190506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156101af57600080fd5b5050610200806101c06000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680634881cca7146100515780639dd2d1b5146100a6575b600080fd5b341561005c57600080fd5b6100646100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34156100b157600080fd5b6100d0600480803590602001909190803590602001909190505061010f565b6040518082815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663771602f784846000604051602001526040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083815260200182815260200192505050602060405180830381600087803b15156101b157600080fd5b6102c65a03f115156101c257600080fd5b505050604051805190509050929150505600a165627a7a72305820b77914363827571526736ed47dffabcc02813731c54c3c1a0269352c84305bd30029"

	setPermissions(t, "alice", permission.None)
	setPermissions(t, "bob", permission.Call)
	setPermissions(t, "carol", permission.CreateContract)
	setPermissions(t, "vbuterin", permission.Call|permission.CreateContract)

	factoryBytes, _ := hex.DecodeString(factory)
	adderBytes, _ := hex.DecodeString(adder)
	testerBytes, _ := hex.DecodeString(tester)

	// Should fail: Alice has no permission to create or call a contract
	tx1 := makeCallTx(t, "alice", crypto.Address{}, factoryBytes, 0, _fee)
	signAndExecute(t, e.ErrPermDenied, tx1, "alice")

	// Should fail: Bob has call permission but create contract
	tx2 := makeCallTx(t, "bob", crypto.Address{}, factoryBytes, 0, _fee)
	signAndExecute(t, e.ErrPermDenied, tx2, "bob")

	// Should fail: Carol has create contract permission but not call
	tx3 := makeCallTx(t, "carol", crypto.Address{}, factoryBytes, 0, _fee)
	signAndExecute(t, e.ErrPermDenied, tx3, "carol")

	// Should pass: Vitalik has permission to create and call a contract
	tx4 := makeCallTx(t, "vbuterin", crypto.Address{}, factoryBytes, 0, _fee)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "vbuterin")
	assert.Equal(t, rec4.Failed, false)

	// Should pass: Vitalik has permission to create and call a contract
	tx5 := makeCallTx(t, "vbuterin", crypto.Address{}, adderBytes, 0, _fee)
	_, rec5 := signAndExecute(t, e.ErrNone, tx5, "vbuterin")
	assert.Equal(t, rec5.Failed, false)

	// Should fail: Tester has constructor
	tx6 := makeCallTx(t, "vbuterin", crypto.Address{}, testerBytes, 0, _fee)
	_, rec6 := signAndExecute(t, e.ErrNone, tx6, "vbuterin")
	assert.Equal(t, rec6.Failed, true)

}

// Test creating a contract from futher down the call stack
func TestStackOverflow(t *testing.T) {
	setPermissions(t, "vbuterin", permission.Call|permission.CreateContract)

	/*
		pragma solidity ^0.4.0;

		contract A {
			function overflow() public constant returns (int) {
				return overflow();
			}
		}
	*/

	contractA, _ := hex.DecodeString("608060405234801561001057600080fd5b5060a48061001f6000396000f300608060405260043610603e576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680624264c3146043575b600080fd5b348015604e57600080fd5b506055606b565b6040518082815260200191505060405180910390f35b60006073606b565b9050905600a165627a7a723058203f79e7e64c0023b9c9a103344607f633e35a89ffb907155178e6549c16016fad0029")
	overflow, _ := hex.DecodeString("004264c3")

	tx1 := makeCallTx(t, "vbuterin", crypto.Address{}, contractA, 0, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "vbuterin")
	require.Equal(t, rec1.Failed, false)
	fmt.Println(rec1)

	tx2 := makeCallTx(t, "vbuterin", *rec1.ContractAddress, overflow, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "vbuterin")
	require.Equal(t, rec2.Failed, true)
	fmt.Println(rec2)

	require.Equal(t, rec2.UsedGas, 0)
	require.Equal(t, rec2.ContractAddress, 0)
	require.Equal(t, rec2.Height, 0)
	require.Equal(t, rec2.Status, "fff")
}

func TestContractSend(t *testing.T) {
	setPermissions(t, "alice", permission.Call)
	/*
	   contract Caller {
	      function send(address x){
	          x.send(msg.value);
	      }
	   }
	*/
	callerCode, _ := hex.DecodeString("60606040526000357c0100000000000000000000000000000000000000000000000000000000900480633e58c58c146037576035565b005b604b6004808035906020019091905050604d565b005b8073ffffffffffffffffffffffffffffffffffffffff16600034604051809050600060405180830381858888f19350505050505b5056")
	sendData, _ := hex.DecodeString("3e58c58c")

	_, caller1Addr := makeContractAccount(t, callerCode, 0, 0)
	_, caller2Addr := makeContractAccount(t, callerCode, 0, permission.Call)

	sendData = append(sendData, tAccounts["bob"].Address().Word256().Bytes()...)
	sendAmt := uint64(10)

	aliceBalance := getBalance(t, "alice")
	bobBalance := getBalance(t, "bob")

	tx1 := makeCallTx(t, "alice", caller1Addr, sendData, sendAmt, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	fmt.Println(rec1)

	tx2 := makeCallTx(t, "alice", caller2Addr, sendData, sendAmt, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	fmt.Println(rec2)

	checkBalance(t, "alice", aliceBalance-sendAmt-_fee-_fee)
	checkBalance(t, "bob", bobBalance+sendAmt)
	checkBalanceByAddress(t, caller1Addr, 0)
	checkBalanceByAddress(t, caller2Addr, 0)
}

func TestSelfDestruct(t *testing.T) {
	setPermissions(t, "alice", permission.Send|permission.Call|permission.CreateAccount)

	aliceBalance := getBalance(t, "alice")
	bobBalance := getBalance(t, "bob")
	sendAmt := uint64(1)
	refundedBalance := uint64(100)

	// store 0x1 at 0x1, push an address, then self-destruct:)
	contractCode := []byte{0x60, 0x01, 0x60, 0x01, 0x55, 0x73}
	contractCode = append(contractCode, tAccounts["bob"].Address().RawBytes()...)
	contractCode = append(contractCode, 0xff)

	_, contractAddr := makeContractAccount(t, contractCode, refundedBalance, 0)

	// send call tx with no data, cause self-destruct
	tx1 := makeCallTx(t, "alice", contractAddr, nil, sendAmt, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	// if we do it again, the caller should lose fee
	tx2 := makeCallTx(t, "alice", contractAddr, nil, sendAmt, _fee)
	signAndExecute(t, e.ErrNone, tx2, "alice")

	contractAcc := getAccount(t, contractAddr)
	require.Nil(t, contractAcc, "Expected account to be removed")

	checkBalance(t, "alice", aliceBalance-sendAmt-_fee-_fee)
	checkBalance(t, "bob", bobBalance+refundedBalance+sendAmt)
}

func TestTxSequence(t *testing.T) {
	setPermissions(t, "alice", permission.Send)

	sequence1 := getAccountByName(t, "alice").Sequence()
	sequence2 := getAccountByName(t, "bob").Sequence()
	for i := 0; i < 100; i++ {
		panic("dddddddddddddddddddddddddddddd")
		tx := makeSendTx(t, "alice", "bob", 1, _fee)
		signAndExecute(t, e.ErrNone, tx, "alice")
	}

	require.Equal(t, sequence1+100, getAccountByName(t, "alice").Sequence())
	require.Equal(t, sequence2, getAccountByName(t, "bob").Sequence())
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
	code, _ := hex.DecodeString("608060405234801561001057600080fd5b5060cc8061001f6000396000f300608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c146044575b600080fd5b348015604f57600080fd5b5060566098565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050905600a165627a7a7230582051d64d7178179a4ce4a151a49fe75742727f92a2becc39861f5a812c9e9bc00b0029")

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Failed, false)

	// check sequence
	seq2 := getAccountByName(t, "alice").Sequence()
	assert.Equal(t, seq2, seq1+1)

	contractAddr := *rec1.ContractAddress
	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	if !bytes.Equal(contractAcc.Code(), rec1.Output) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), rec1.Output)
	}

	// Input is the function hash of `get()`
	input, _ := hex.DecodeString("6d4ce63c")
	tx2 := makeCallTx(t, "alice", contractAddr, input, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Failed, false)
	addr1, _ := crypto.AccountAddress(rec2.Output[12:])
	addr2 := tx2.Caller().Address
	assert.Equal(t, addr1.String(), addr2.String())
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

		}
	*/

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, _ := hex.DecodeString("608060405234801561001057600080fd5b5060df8061001f6000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680635362f8a214604e578063e1cb0e52146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820d95c60259bab2980d08fef23190fd2160ac19d7c103250a94e427f8d4d01b2030029")

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Failed, false)
	seq2 := getAccountByName(t, "alice").Sequence()
	assert.Equal(t, seq2, seq1+1)

	contractAddr := *rec1.ContractAddress
	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	// empty storage
	input1, _ := hex.DecodeString("e1cb0e52")
	tx11 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	_, rec11 := signAndExecute(t, e.ErrNone, tx11, "alice")
	assert.Equal(t, rec11.Output, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	// Input is the function hash of `setVal()`: 100
	input2, _ := hex.DecodeString("5362f8a20000000000000000000000000000000000000000000000000000000000000064")
	tx2 := makeCallTx(t, "alice", contractAddr, input2, 0, 100000)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Failed, false)

	// Input is the function hash of `getVal()`
	tx3 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	_, rec3 := signAndExecute(t, e.ErrNone, tx3, "alice")
	assert.Equal(t, rec3.Output[31:], []byte{100})

	// Input is the function hash of `setVal()`: aaaa...
	input3, _ := hex.DecodeString("5362f8a2aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tx4 := makeCallTx(t, "alice", contractAddr, input3, 0, 100000)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "alice")
	assert.Equal(t, rec4.Failed, false)

	// Input is the function hash of `getVal()`
	tx5 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	_, rec5 := signAndExecute(t, e.ErrNone, tx5, "alice")
	out, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Equal(t, rec5.Output, out)
}

func TestStorage2(t *testing.T) {
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	/*
		pragma solidity ^0.4.18;

		contract EvmTest2{
			int value;

			function EvmTest2(int val) public {
				value = val;
			}

			function setVal(int val) public {
				value = val;
			}

			function getVal() public returns(int) {
				return value;
			}

		}
	*/

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, _ := hex.DecodeString(`608060405234801561001057600080fd5b5060405160208061012883398101806040528101908080519060200190929190505050806000819055505060df806100496000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680635362f8a214604e578063e1cb0e52146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820e6ff3a1bc432f3b22cb4f663e7a9f70e9c12e701dd92c2021609fd8481a0998f002900000000000000000000000000000000000000000000000000000000000000aa`)
	data, _ := hex.DecodeString("00000000000000000000000000000000000000000000000000000000000000aa")
	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Failed, false)
	seq2 := getAccountByName(t, "alice").Sequence()
	assert.Equal(t, seq2, seq1+1)

	contractAddr := *rec1.ContractAddress
	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	// Input is the function hash of `getVal()`
	input1, _ := hex.DecodeString("e1cb0e52")
	tx2 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Failed, false)
	assert.Equal(t, rec2.Output, data)

	// Input is the function hash of `setVal()`: 100
	input2, _ := hex.DecodeString("5362f8a20000000000000000000000000000000000000000000000000000000000000064")
	tx3 := makeCallTx(t, "alice", contractAddr, input2, 0, 100000)
	_, rec3 := signAndExecute(t, e.ErrNone, tx3, "alice")
	assert.Equal(t, rec3.Failed, false)

	// Input is the function hash of `getVal()`
	tx4 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "alice")
	assert.Equal(t, rec4.Failed, false)
	assert.Equal(t, rec4.Output[31:], []byte{100})

	// Input is the function hash of `setVal()`: aaaa...
	input3, _ := hex.DecodeString("5362f8a2aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tx5 := makeCallTx(t, "alice", contractAddr, input3, 0, 100000)
	_, rec5 := signAndExecute(t, e.ErrNone, tx5, "alice")
	assert.Equal(t, rec5.Failed, false)

	// Input is the function hash of `getVal()`
	tx6 := makeCallTx(t, "alice", contractAddr, input1, 0, 100000)
	_, rec6 := signAndExecute(t, e.ErrNone, tx6, "alice")
	out, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Equal(t, rec6.Failed, false)
	assert.Equal(t, rec6.Output, out)
}

func TestCallContract2(t *testing.T) {
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
