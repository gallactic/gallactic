package tests

import (
	"bytes"
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

func makeCallTx(t *testing.T, from string, addr crypto.Address, data []byte, amt, fee uint64) *tx.CallTx {
	acc := getAccountByName(t, from)
	tx, err := tx.NewCallTx(acc.Address(), addr, acc.Sequence()+1, data, 210000, amt, fee)
	require.Equal(t, amt, tx.Amount())
	require.Equal(t, fee, tx.Fee())
	assert.NoError(t, err)

	return tx
}

func execTxWaitAccountCall(t *testing.T, tx tx.Tx, name string, addr crypto.Address) ( /* *events.EventDataCall*/ error, error) {
	env := txs.Enclose(tChainID, tx)
	/// ch := make(chan *events.EventDataCall)
	/// const subscriber = "exexTxWaitEvent"

	require.NoError(t, env.Sign(tSigners[name]), "Could not sign tx in call: %s", debug.Stack())

	/// events.SubscribeAccountCall(ctx, emitter, subscriber, address, env.Tx.Hash(), -1, ch)
	/// defer emitter.UnsubscribeAll(ctx, subscriber)

	err := tCommitter.Execute(env)
	assert.NoError(t, err)

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
	return err, err
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

func TestCallPermission(t *testing.T) {
	setPermissions(t, "alice", permission.Call)

	// A single input, having the permission, should succeed
	// 	create simple contract
	_, simpleContractAddr := makeContractAccount(t, []byte{0x60}, 0, 0)

	tx1 := makeCallTx(t, "alice", simpleContractAddr, nil, 100, _fee)
	_, err := execTxWaitAccountCall(t, tx1, "alice", simpleContractAddr)
	require.NoError(t, err)

	//----------------------------------------------------------
	// call to contract that calls simple contract - without perm

	// create contract that calls the simple contract
	contractCode1 := callContractCode(simpleContractAddr, 0)
	caller1Acc, caller1Addr := makeContractAccount(t, contractCode1, 1000, permission.ZeroPermissions)

	// A single input, having the permission, but the contract doesn't have permission
	// we need to subscribe to the Call event to detect the exception
	tx2 := makeCallTx(t, "alice", caller1Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx2, "alice", caller1Addr)
	require.Equal(t, e.Code(err), e.ErrPermDenied)

	//----------------------------------------------------------
	// call to contract that calls simple contract - with perm
	// A single input, having the permission, and the contract has permission
	caller1Acc.SetPermissions(permission.Call)
	updateAccount(t, caller1Acc)
	tx3 := makeCallTx(t, "alice", caller1Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx3, "alice", caller1Addr)
	require.NoError(t, err)

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// caller1Contract does not have call perms, but caller2Contract does.
	contractCode2 := callContractCode(caller1Addr, 0)
	caller2Acc, caller2Addr := makeContractAccount(t, contractCode2, 1000, 0)

	caller1Acc.UnsetPermissions(permission.Call)
	caller2Acc.SetPermissions(permission.Call)
	updateAccount(t, caller1Acc)
	updateAccount(t, caller2Acc)

	tx4 := makeCallTx(t, "alice", caller2Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx4, "alice", caller1Addr)
	require.Error(t, err)

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// both caller1 and caller2 have permission
	caller1Acc.SetPermissions(permission.Call)
	updateAccount(t, caller1Acc)

	tx5 := makeCallTx(t, "alice", caller2Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx5, "alice", caller1Addr)
	require.NoError(t, err)
}

func TestCreateContractPermission(t *testing.T) {
	setPermissions(t, "alice", permission.Call|permission.CreateContract)

	//------------------------------
	// create a simple contract
	contractCode := []byte{0x60}
	createCode := wrapContractForCreateCode(contractCode)

	// A single input, having the permission, should succeed
	tx1 := makeCallTx(t, "alice", crypto.Address{}, createCode, 100, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	// ensure the contract is there
	contractAddr := crypto.DeriveContractAddress(tx1.Caller().Address, tx1.Caller().Sequence)
	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	if !bytes.Equal(contractAcc.Code(), contractCode) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), contractCode)
	}

	//------------------------------
	// create contract that uses the CREATE op
	factoryCode := createContractCode()
	createFactoryCode := wrapContractForCreateCode(factoryCode)

	// A single input, having the permission, should succeed
	tx2 := makeCallTx(t, "alice", crypto.Address{}, createFactoryCode, 100, _fee)
	signAndExecute(t, e.ErrNone, tx2, "alice")

	// ensure the contract is there
	contractAddr = crypto.DeriveContractAddress(tx2.Caller().Address, tx2.Caller().Sequence)
	contractAcc = getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	if !bytes.Equal(contractAcc.Code(), factoryCode) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), factoryCode)
	}

	//------------------------------
	// call the contract (should FAIL)
	tx3 := makeCallTx(t, "alice", contractAddr, createCode, 100, _fee)
	_, err := execTxWaitAccountCall(t, tx3, "alice", contractAddr)
	require.Error(t, err)

	//------------------------------
	// call the contract (should PASS)
	contractAcc.SetPermissions(permission.CreateContract)
	updateAccount(t, contractAcc)

	tx4 := makeCallTx(t, "alice", contractAddr, createCode, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx4, "alice", contractAddr)
	require.NoError(t, err)

	//--------------------------------
	// call the empty address
	code := callContractCode(crypto.Address{}, 0)

	_, contractAddr2 := makeContractAccount(t, code, 1000, permission.Call|permission.CreateContract)

	// this should call the 0 address but not create ...
	tx5 := makeCallTx(t, "alice", contractAddr2, createCode, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx5, "alice", crypto.Address{})
	require.NoError(t, err)

	zeroAcc := getAccount(t, crypto.Address{})
	require.NotNil(t, zeroAcc)
	if len(zeroAcc.Code()) != 0 {
		t.Fatal("the zero account was given code from a CALL!")
	}
}

func TestCreateContractPermission2(t *testing.T) {
	setPermissions(t, "alice", permission.Send|permission.CreateAccount)
	setPermissions(t, "bob", permission.Send)
	setPermissions(t, "carol", permission.Call)

	aliceBalance := getBalance(t, "alice")
	bobBalance := getBalance(t, "bob")
	carolBalance := getBalance(t, "carol")

	//----------------------------------------------------------
	// CALL to unknown account

	// call to contract that calls unknown account - without create_account perm
	// create contract that calls the simple contract
	newAddress := newAccountAddress(t)
	contractCode := callContractCode(newAddress, 3)
	_, caller1Addr := makeContractAccount(t, contractCode, 0, 0)

	// A single input, having the call permission, but the contract doesn't have any permission
	tx7 := makeCallTx(t, "carol", caller1Addr, nil, 100, _fee)
	_, err := execTxWaitAccountCall(t, tx7, "carol", caller1Addr)
	require.Equal(t, e.Code(err), e.ErrPermDenied)

	// A single input, having the call permission, but the contract doesn't have only call permission
	_, caller2Addr := makeContractAccount(t, contractCode, 0, permission.Call)

	tx8 := makeCallTx(t, "carol", caller2Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx8, "carol", caller2Addr)
	require.Equal(t, e.Code(err), e.ErrPermDenied)

	// A single input, having the call permission, but the contract doesn't have call and create account permissions
	_, caller3Addr := makeContractAccount(t, contractCode, 0, permission.Call|permission.CreateAccount)
	tx9 := makeCallTx(t, "carol", caller3Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx9, "carol", caller3Addr)
	require.Equal(t, e.Code(err), e.ErrPermDenied)

	// Both input and contract have call and create account permissions
	setPermissions(t, "carol", permission.Call|permission.CreateContract)
	_, caller4Addr := makeContractAccount(t, contractCode, 0, permission.Call|permission.CreateAccount)
	tx10 := makeCallTx(t, "carol", caller4Addr, nil, 100, _fee)
	_, err = execTxWaitAccountCall(t, tx10, "carol", caller4Addr)
	require.NoError(t, err)

	checkBalance(t, "alice", aliceBalance-(4*(5+_fee)))
	checkBalance(t, "bob", bobBalance-(3*(5+_fee)))
	checkBalance(t, "carol", carolBalance-(100+(4*_fee))) /// 1 successfull transaction + 3 failed transactions
	checkBalanceByAddress(t, newAddress, 3)
	checkBalanceByAddress(t, caller4Addr, 97)
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
	_, err := execTxWaitAccountCall(t, tx1, "alice", preFactoryAddr)
	require.Error(t, err)
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
	_, err := execTxWaitAccountCall(t, tx1, "alice", caller1Addr)
	require.Error(t, err)

	tx2 := makeCallTx(t, "alice", caller2Addr, sendData, sendAmt, _fee)
	_, err = execTxWaitAccountCall(t, tx2, "alice", caller2Addr)
	require.NoError(t, err)

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
	_, err := execTxWaitAccountCall(t, tx1, "alice", contractAddr)
	require.NoError(t, err)

	// if we do it again, the caller should lose fee
	tx2 := makeCallTx(t, "alice", contractAddr, nil, sendAmt, _fee)
	_, err = execTxWaitAccountCall(t, tx2, "alice", contractAddr)
	require.Equal(t, e.Code(err), e.ErrTimeOut)

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
