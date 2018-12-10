package tests

import (
	"encoding/hex"
	"runtime/debug"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/txs"
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

func executeAndWait(t *testing.T, errorCode int, tx tx.Tx, name string, addr crypto.Address) /*  *events.EventDataCall*/ {
	env := txs.Enclose(tChainID, tx)
	/// ch := make(chan *events.EventDataCall)
	/// const subscriber = "exexTxWaitEvent"

	require.NoError(t, env.Sign(tSigners[name]), "Could not sign tx in call: %s", debug.Stack())

	/// events.SubscribeAccountCall(ctx, emitter, subscriber, address, env.Tx.Hash(), -1, ch)
	/// defer emitter.UnsubscribeAll(ctx, subscriber)

	err := tCommitter.Execute(env)
	assert.Equal(t, e.Code(err), errorCode)

	commit(t)
	/*
		ticker := time.NewTicker(2 * time.Second)

		select {
		case eventDataCall := <-ch:
			fmt.Println("MSG: ", eventDataCall)
			return eventDataCall, eventDataCall.Exception

		case <-ticker.C:
			return nil, e.Error(e.ErrTimeOut)
		}
	*/
	///return nil
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
var factory = "0x6060604052341561000f57600080fd5b61011c8061001e6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806377e5484d146044575b600080fd5b3415604e57600080fd5b609c600480803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509190505060de565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008151602083016000f090509190505600a165627a7a72305820a33693c0d301f0c00332f4a0c7fa39f7270d69bd9affbdda39770a9cbe7f40390029"
var adder = "0x6060604052341561000f57600080fd5b60ba8061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063771602f7146044575b600080fd5b3415604e57600080fd5b606b60048080359060200190919080359060200190919050506081565b6040518082815260200191505060405180910390f35b60008183019050929150505600a165627a7a72305820d1b89dadbf6f1c4c376399ed75b6a5a25f968cdb879f371ab4da98125539cdf80029"
var tester = "0x6060604052341561000f57600080fd5b6040516103c03803806103c0833981016040528080519060200190919080518201919050508173ffffffffffffffffffffffffffffffffffffffff166377e5484d826000604051602001526040518263ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825283818151815260200191508051906020019080838360005b838110156100c55780820151818401526020810190506100aa565b50505050905090810190601f1680156100f25780820380516001836020036101000a031916815260200191505b5092505050602060405180830381600087803b151561011057600080fd5b6102c65a03f1151561012157600080fd5b505050604051805190506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156101af57600080fd5b5050610200806101c06000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680634881cca7146100515780639dd2d1b5146100a6575b600080fd5b341561005c57600080fd5b6100646100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34156100b157600080fd5b6100d0600480803590602001909190803590602001909190505061010f565b6040518082815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663771602f784846000604051602001526040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083815260200182815260200192505050602060405180830381600087803b15156101b157600080fd5b6102c65a03f115156101c257600080fd5b505050604051805190509050929150505600a165627a7a72305820c677522627ac4ef94ee16854c54c30cb10ccb63065b1e73e7f7c6b83a0c946e30029"

func TestCallAndCreateContractPermissions(t *testing.T) {
	setPermissions(t, "alice", permission.None)
	setPermissions(t, "bob", permission.Call)
	setPermissions(t, "vbuterin", permission.CreateContract)

	factoryBytes, _ := hex.DecodeString(factory)
	//	adderBytes, _ := hex.DecodeString(factory)
	testerBytes, _ := hex.DecodeString(factory)

	// Should fail: Alice has no permission to create or call a contract
	tx1 := makeCallTx(t, "alice", crypto.Address{}, factoryBytes, 0, _fee)
	executeAndWait(t, e.ErrPermDenied, tx1, "alice", crypto.Address{})

	// Should pass: Vitalik has permission to create a contract
	tx2 := makeCallTx(t, "vbuterin", crypto.Address{}, factoryBytes, 0, _fee)
	executeAndWait(t, e.ErrNone, tx2, "vbuterin", crypto.Address{})

	// Should fail: Tester has constructor
	tx3 := makeCallTx(t, "vbuterin", crypto.Address{}, testerBytes, 0, _fee)
	executeAndWait(t, e.ErrNone, tx3, "vbuterin", crypto.Address{})

	///0x6060604052341561000f57600080fd5b6040516103c03803806103c0833981016040528080519060200190919080518201919050508173ffffffffffffffffffffffffffffffffffffffff166377e5484d826000604051602001526040518263ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825283818151815260200191508051906020019080838360005b838110156100c55780820151818401526020810190506100aa565b50505050905090810190601f1680156100f25780820380516001836020036101000a031916815260200191505b5092505050602060405180830381600087803b151561011057600080fd5b6102c65a03f1151561012157600080fd5b505050604051805190506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156101af57600080fd5b5050610200806101c06000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680634881cca7146100515780639dd2d1b5146100a6575b600080fd5b341561005c57600080fd5b6100646100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34156100b157600080fd5b6100d0600480803590602001909190803590602001909190505061010f565b6040518082815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663771602f784846000604051602001526040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083815260200182815260200192505050602060405180830381600087803b15156101b157600080fd5b6102c65a03f115156101c257600080fd5b505050604051805190509050929150505600a165627a7a72305820c677522627ac4ef94ee16854c54c30cb10ccb63065b1e73e7f7c6b83a0c946e30029 000000000000000000000000038f160ad632409bfb18582241d9fd88c1a072ba 000000000000000000000000000000000000000000000000000000000000004 000000000000000000000000000000000000000000000000000000000000000d7 6060604052341561000f57600080fd5b60ba8061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063771602f7146044575b600080fd5b3415604e57600080fd5b606b60048080359060200190919080359060200190919050506081565b6040518082815260200191505060405180910390f35b60008183019050929150505600a165627a7a72305820d1b89dadbf6f1c4c376399ed75b6a5a25f968cdb879f371ab4da98125539cdf80029000000000000000000
	// Should pass: Vitalik has permission to create a contract
	tx4 := makeCallTx(t, "vbuterin", crypto.Address{}, testerBytes, 0 /**/, _fee)
	executeAndWait(t, e.ErrNone, tx4, "vbuterin", crypto.Address{})

	///?????
	/// when a contract created it inherits the permission from the creator???? How

	// Should fail: Vitalik has no permission to call a contract
	tx5 := makeCallTx(t, "vbuterin", crypto.Address{} /**/, testerBytes, 0, _fee)
	executeAndWait(t, e.ErrNone, tx5, "vbuterin", crypto.Address{})

	// Should pass: Bob has permission to call a contract
	tx6 := makeCallTx(t, "gheis", crypto.Address{} /**/, testerBytes, 0, _fee)
	executeAndWait(t, e.ErrNone, tx6, "gheis", crypto.Address{})
}

// Test creating a contract from futher down the call stack
func TestStackOverflow(t *testing.T) {
	setPermissions(t, "alice", permission.Call|permission.CreateAccount)

	/*
	   contract Factory {
	      address a;
	      function create() returns (address){
	          a = new PreFactory();
	          return a;
	      }
	   }

	   contract PreFactory{
	      address a;
	      function create(Factory c) returns (address) {
	      	a = c.create();
	      	return a;
	      }
	   }
	*/

	// run-time byte code for each of the above
	preFactoryCode, _ := hex.DecodeString("60606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
	factoryCode, _ := hex.DecodeString("60606040526000357C010000000000000000000000000000000000000000000000000000000090048063EFC81A8C146037576035565B005B60426004805050606E565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B6000604051610153806100E0833901809050604051809103906000F0600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905060DD565B90566060604052610141806100126000396000F360606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
	createData, _ := hex.DecodeString("9ed93318")

	_, preFactoryAddr := makeContractAccount(t, preFactoryCode, 0, permission.Call)
	assert.NotNil(t, preFactoryAddr)
	_, factoryAddr := makeContractAccount(t, factoryCode, 0, permission.Call)

	createData = append(createData, factoryAddr.Word256().Bytes()...)

	// call the pre-factory, triggering the factory to run a create
	tx1 := makeCallTx(t, "alice", preFactoryAddr, createData, 3, _fee)
	signAndExecute(t, e.ErrGeneric, tx1, "alice")
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
	signAndExecute(t, e.ErrGeneric, tx1, "alice")

	tx2 := makeCallTx(t, "alice", caller2Addr, sendData, sendAmt, _fee)
	signAndExecute(t, e.ErrGeneric, tx2, "alice")

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
