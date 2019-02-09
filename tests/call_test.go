package tests

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultGas uint64 = 21000000

func makeCallTx(t *testing.T, from string, addr crypto.Address, data []byte, amt, fee uint64) *tx.CallTx {
	acc := getAccountByName(t, from)
	tx, err := tx.NewCallTx(acc.Address(), addr, acc.Sequence()+1, data, defaultGas, amt, fee)
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
	signAndExecute(t, e.ErrPermissionDenied, tx1, "alice")

	// simple call tx with send permission should fail
	tx2 := makeCallTx(t, "bob", simpleContractAddr, nil, 100, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx2, "bob")

	// simple call tx with create permission should fail
	tx3 := makeCallTx(t, "dan", simpleContractAddr, nil, 100, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx3, "dan")

	//-------------------
	// create txs

	// simple call create tx should fail
	tx4 := makeCallTx(t, "alice", crypto.Address{}, nil, 100, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx4, "alice")

	// simple call create tx with send perm should fail
	tx5 := makeCallTx(t, "bob", crypto.Address{}, nil, 100, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx5, "bob")

	// simple call create tx with call perm should fail
	tx6 := makeCallTx(t, "carol", crypto.Address{}, nil, 100, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx6, "carol")
}

func TestCreateContractNew(t *testing.T) {
	/*
		pragma solidity ^0.4.18;

		contract Adder {
		    function add(uint a, uint b) public pure returns (uint){
		        return a+b;
		    }
		}

		contract Tester {
		    Adder a;

		    function Tester() public {
		        a = new Adder();
		    }

		    function AdderAddr() public constant returns (address){
		        return a;
		    }

		    function AdderAdd(uint x, uint y) public view returns (uint){
		        return a.add(x,y);
		    }
		}
	*/

	setPermissions(t, "alice", permission.None)
	setPermissions(t, "bob", permission.Call)
	setPermissions(t, "carol", permission.CreateContract)
	setPermissions(t, "vbuterin", permission.Call|permission.CreateContract)

	// bytecodes
	adderBytes, _ := hex.DecodeString("6060604052341561000f57600080fd5b60ba8061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063771602f7146044575b600080fd5b3415604e57600080fd5b606b60048080359060200190919080359060200190919050506081565b6040518082815260200191505060405180910390f35b60008183019050929150505600a165627a7a72305820ded37d1d3fa3f6ea3df33b652f4e40ee519dfcadf898054d77acd193533066580029")
	testerBytes, _ := hex.DecodeString("6060604052341561000f57600080fd5b610017610071565b604051809103906000f080151561002d57600080fd5b6000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610080565b60405160d78061028f83390190565b6102008061008f6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680634881cca7146100515780639dd2d1b5146100a6575b600080fd5b341561005c57600080fd5b6100646100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34156100b157600080fd5b6100d0600480803590602001909190803590602001909190505061010f565b6040518082815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663771602f784846000604051602001526040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083815260200182815260200192505050602060405180830381600087803b15156101b157600080fd5b6102c65a03f115156101c257600080fd5b505050604051805190509050929150505600a165627a7a723058206b26d71f7a4612ff6dbc2eda4c2105cdfa2dbb2011c39869c6aa4d40b702871b00296060604052341561000f57600080fd5b60ba8061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063771602f7146044575b600080fd5b3415604e57600080fd5b606b60048080359060200190919080359060200190919050506081565b6040518082815260200191505060405180910390f35b60008183019050929150505600a165627a7a72305820e8305c2bcda499d6ec2b571dad1eaaeb4cbdc10ac93ca798d574816cf2fbb0ae0029")
	adderAddFunc, _ := hex.DecodeString("771602f7")
	testerAddFunc, _ := hex.DecodeString("9dd2d1b5")
	testerAddrFunc, _ := hex.DecodeString("4881cca7")

	// Should fail: Alice has no permission to create or call a contract
	tx1 := makeCallTx(t, "alice", crypto.Address{}, adderBytes, 0, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx1, "alice")

	// Should fail: Bob has call permission but create contract
	tx2 := makeCallTx(t, "bob", crypto.Address{}, adderBytes, 0, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx2, "bob")

	// Should fail: Carol has create contract permission but not call
	tx3 := makeCallTx(t, "carol", crypto.Address{}, adderBytes, 0, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx3, "carol")

	// Adder: Should pass: vbuterin has permission to call and create a contract
	tx4 := makeCallTx(t, "vbuterin", crypto.Address{}, adderBytes, 0, _fee)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "vbuterin")
	require.Equal(t, rec4.Status, txs.Ok)
	require.Equal(t, rec4.GasWanted, defaultGas)
	require.NotZero(t, rec4.GasUsed)

	// Should pass: result is 5
	adderAddData1 := addParams_2(adderAddFunc, 1, 4)
	returnValue1, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000005")
	tx44 := makeCallTx(t, "vbuterin", *rec4.ContractAddress, adderAddData1, 0, _fee)
	_, rec44 := signAndExecute(t, e.ErrNone, tx44, "vbuterin")
	assert.Equal(t, rec44.Status, txs.Ok)
	assert.Equal(t, rec44.Output, returnValue1)

	// Tester: Should pass: vbuterin has permission to call and create a contract
	tx5 := makeCallTx(t, "vbuterin", crypto.Address{}, testerBytes, 0, _fee)
	_, rec5 := signAndExecute(t, e.ErrNone, tx5, "vbuterin")
	require.Equal(t, rec5.Status, txs.Ok)
	fmt.Printf("Tester address: %s\n", rec5.ContractAddress) // -> Tester address

	// should fail: wrong function hash
	adderAddFuncWrong, _ := hex.DecodeString("9dd2d1b6") // actual is 9dd2d1b5
	adderAddDataWrong := addParams_2(adderAddFuncWrong, 1, 4)
	tx6 := makeCallTx(t, "vbuterin", *rec5.ContractAddress, adderAddDataWrong, 0, _fee)
	_, rec6 := signAndExecute(t, e.ErrNone, tx6, "vbuterin")
	assert.Equal(t, rec6.Status, txs.Failed)
	assert.Empty(t, rec6.Output)

	// Should pass: call tester_add function, result is 5
	testerAddData2 := addParams_2(testerAddFunc, 1, 4)
	returnValue2, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000005")
	tx7 := makeCallTx(t, "vbuterin", *rec5.ContractAddress, testerAddData2, 0, _fee)
	_, rec7 := signAndExecute(t, e.ErrNone, tx7, "vbuterin")
	assert.Equal(t, rec7.Status, txs.Ok)
	assert.Equal(t, rec7.Output, returnValue2)

	// Should pass: get the address of deployed adder contract
	tx8 := makeCallTx(t, "vbuterin", *rec5.ContractAddress, testerAddrFunc, 0, _fee)
	_, rec8 := signAndExecute(t, e.ErrNone, tx8, "vbuterin")
	assert.Equal(t, rec8.Status, txs.Ok)

	// Should pass: call adder_add function, result is 5
	addr, err := crypto.ContractAddress(rec8.Output[12:])
	require.NoError(t, err)
	tx9 := makeCallTx(t, "vbuterin", addr, adderAddData1, 0, _fee)
	_, rec9 := signAndExecute(t, e.ErrNone, tx9, "vbuterin")
	assert.Equal(t, rec9.Status, txs.Ok)
	assert.Equal(t, rec9.Output, returnValue1)
}

func TestCreateContract(t *testing.T) {
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

	setPermissions(t, "alice", permission.None)
	setPermissions(t, "bob", permission.Call)
	setPermissions(t, "carol", permission.CreateContract)
	setPermissions(t, "vbuterin", permission.Call|permission.CreateContract)

	factoryBytes, _ := hex.DecodeString("6060604052341561000f57600080fd5b61011c8061001e6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806377e5484d146044575b600080fd5b3415604e57600080fd5b609c600480803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509190505060de565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008151602083016000f090509190505600a165627a7a723058205b838433b9d5bb3dd706fd8c6d5e8d52c57cc7bf89c101b58d4f38f61bde0ab30029")
	adderBytes, _ := hex.DecodeString("6060604052341561000f57600080fd5b60ba8061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063771602f7146044575b600080fd5b3415604e57600080fd5b606b60048080359060200190919080359060200190919050506081565b6040518082815260200191505060405180910390f35b60008183019050929150505600a165627a7a72305820757b3773ced4707dc6dd8b9e09c9de0803d141c1533af46967ff3d89a035b3970029")
	testerBytes, _ := hex.DecodeString("6060604052341561000f57600080fd5b6040516103c03803806103c0833981016040528080519060200190919080518201919050508173ffffffffffffffffffffffffffffffffffffffff166377e5484d826000604051602001526040518263ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825283818151815260200191508051906020019080838360005b838110156100c55780820151818401526020810190506100aa565b50505050905090810190601f1680156100f25780820380516001836020036101000a031916815260200191505b5092505050602060405180830381600087803b151561011057600080fd5b6102c65a03f1151561012157600080fd5b505050604051805190506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156101af57600080fd5b5050610200806101c06000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680634881cca7146100515780639dd2d1b5146100a6575b600080fd5b341561005c57600080fd5b6100646100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34156100b157600080fd5b6100d0600480803590602001909190803590602001909190505061010f565b6040518082815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663771602f784846000604051602001526040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083815260200182815260200192505050602060405180830381600087803b15156101b157600080fd5b6102c65a03f115156101c257600080fd5b505050604051805190509050929150505600a165627a7a7230582095edf6d06ae1624aef1797f7eff2920c697db4cfa947becdf49d3435b286963a0029")
	testerAddFunc, _ := hex.DecodeString("9dd2d1b5")
	testerAddrFunc, _ := hex.DecodeString("4881cca7")

	// Should fail: Alice has no permission to create or call a contract
	tx1 := makeCallTx(t, "alice", crypto.Address{}, factoryBytes, 0, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx1, "alice")

	// Should fail: Bob has call permission but create contract
	tx2 := makeCallTx(t, "bob", crypto.Address{}, factoryBytes, 0, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx2, "bob")

	// Should fail: Carol has create contract permission but not call
	tx3 := makeCallTx(t, "carol", crypto.Address{}, factoryBytes, 0, _fee)
	signAndExecute(t, e.ErrPermissionDenied, tx3, "carol")

	seq1 := getAccountByName(t, "vbuterin").Sequence()

	// Factory: Should pass: Vitalik has permission to create and call a contract
	tx4 := makeCallTx(t, "vbuterin", crypto.Address{}, factoryBytes, 0, _fee)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "vbuterin")
	assert.Equal(t, rec4.Status, txs.Ok)
	factoryAddr := *rec4.ContractAddress

	/*
		// Adder: Should pass: Vitalik has permission to create and call a contract
		tx5 := makeCallTx(t, "vbuterin", crypto.Address{}, adderBytes, 0, _fee)
		_, rec5 := signAndExecute(t, e.ErrNone, tx5, "vbuterin")
		assert.Equal(t, rec5.Status, txs.Ok)
	*/

	// Should fail: Tester has constructor
	tx6 := makeCallTx(t, "vbuterin", crypto.Address{}, testerBytes, 0, _fee)
	_, rec6 := signAndExecute(t, e.ErrNone, tx6, "vbuterin")
	assert.Equal(t, rec6.Status, txs.Failed)

	// Tester: Should pass: Tester has constructor with proper vales
	testerBytes = addParams_1(testerBytes, factoryAddr, adderBytes)

	tx7 := makeCallTx(t, "vbuterin", crypto.Address{}, testerBytes, 0, _fee)
	_, rec7 := signAndExecute(t, e.ErrNone, tx7, "vbuterin")
	assert.Equal(t, rec7.Status, txs.Ok)

	// should pass: get adder address
	tx8 := makeCallTx(t, "vbuterin", *rec7.ContractAddress, testerAddrFunc, 0, _fee)
	_, rec8 := signAndExecute(t, e.ErrNone, tx8, "vbuterin")
	assert.Equal(t, rec8.Status, txs.Ok)
	assert.NotEmpty(t, rec8.Output)
	testerAddr, _ := crypto.ContractAddress(rec8.Output[12:])

	// Should pass: add 1+4=5
	testerAddData1 := addParams_2(testerAddFunc, 1, 4)
	returnValue1, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000005")
	tx9 := makeCallTx(t, "vbuterin", *rec7.ContractAddress, testerAddData1, 0, _fee)
	_, rec9 := signAndExecute(t, e.ErrNone, tx9, "vbuterin")
	assert.Equal(t, rec9.Status, txs.Ok)
	assert.Equal(t, rec9.Output, returnValue1)

	seq2 := getAccountByName(t, "vbuterin").Sequence()
	seq3 := getAccount(t, testerAddr).Sequence()
	seq4 := getAccount(t, factoryAddr).Sequence()
	assert.Equal(t, seq1+5, seq2) // vbuterin sent 5 tx
	assert.Equal(t, seq3, uint64(0))
	assert.Equal(t, seq4, uint64(1))
}

func TestStackOverflow(t *testing.T) {
	/*
		pragma solidity ^0.4.0;

		contract A {
			function overflow() public constant returns (int) {
				return overflow();
			}
		}
	*/
	setPermissions(t, "vbuterin", permission.Call|permission.CreateContract)

	contractA, _ := hex.DecodeString("608060405234801561001057600080fd5b5060a48061001f6000396000f300608060405260043610603e576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680624264c3146043575b600080fd5b348015604e57600080fd5b506055606b565b6040518082815260200191505060405180910390f35b60006073606b565b9050905600a165627a7a723058203f79e7e64c0023b9c9a103344607f633e35a89ffb907155178e6549c16016fad0029")
	overflow, _ := hex.DecodeString("004264c3")

	tx1 := makeCallTx(t, "vbuterin", crypto.Address{}, contractA, 0, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "vbuterin")
	require.Equal(t, rec1.Status, txs.Ok)
	fmt.Println(rec1)

	tx2 := makeCallTx(t, "vbuterin", *rec1.ContractAddress, overflow, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "vbuterin")
	require.Equal(t, rec2.Status, txs.Failed)
	require.NotZero(t, rec2.GasUsed)
}

func TestContractSend(t *testing.T) {
	/*
		pragma solidity ^0.4.18;

		contract Send {
			function send(address x) public payable{
				x.send(msg.value);
			}
			function sendBalance(address x) public payable{
				x.send(balance(this));
			}
			function stake() public payable{

			}
			function stake2() public {

			}
			function balance(address x) public view returns (uint256) {
				return address(x).balance;
			}
		}
	*/

	//code, _ := hex.DecodeString("6060604052341561000f57600080fd5b6102058061001e6000396000f30060606040526004361061006d576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806305ea3cf1146100725780633a4b66f1146100875780633e58c58c146100915780635292af1f146100bf578063e3d670d7146100ed575b600080fd5b341561007d57600080fd5b61008561013a565b005b61008f61013c565b005b6100bd600480803573ffffffffffffffffffffffffffffffffffffffff1690602001909190505061013e565b005b6100eb600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091905050610177565b005b34156100f857600080fd5b610124600480803573ffffffffffffffffffffffffffffffffffffffff169060200190919050506101b8565b6040518082815260200191505060405180910390f35b565b565b8073ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f193505050505050565b8073ffffffffffffffffffffffffffffffffffffffff166108fc61019a306101b8565b9081150290604051600060405180830381858888f193505050505050565b60008173ffffffffffffffffffffffffffffffffffffffff163190509190505600a165627a7a723058204131f76eeba980361855afa74e3005f73e68936a33d2a855aea79a7c8b9665450029")

	/// Case 1: deploy: successful
	/// case 2: alice calls Send (amount 10) to Bob address
	/// 	     check: alice balance = 10-fee
	///			 check: contract balance: 0
	///			 check: bob balance += 10
	///
	/// case 3: alice calls Stake (amount 20)
	/// 	     check: alice balance = 20-fee
	///			 check: contract balance: 20
	///
	/// case 3: alice calls stake2 (amount 20)
	/// 	     check: Error. should fail
	///
	/// case 3: alice calls sendBalance
	/// 	     check: alice balance += 20
	///			 check: contract balance: 0

}

func TestTxSequence(t *testing.T) {
	/*
		pragma solidity ^0.4.18;
		contract empty {}
	*/
	setPermissions(t, "alice", permission.Call|permission.CreateContract)

	code, _ := hex.DecodeString("60606040523415600e57600080fd5b603580601b6000396000f3006060604052600080fd00a165627a7a7230582010a4ef31d2df6c96a7b2f027cfceb25025ee26593bae4fffc346e661d6b4200c0029")

	sequence1 := getAccountByName(t, "alice").Sequence()
	for i := 0; i < 100; i++ {
		tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)
		_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
		require.Equal(t, rec1.Status, txs.Ok)

		tx2 := makeCallTx(t, "alice", crypto.Address{}, []byte{}, 0, _fee)
		_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
		require.Equal(t, rec2.Status, txs.Ok)

		tx3 := makeCallTx(t, "alice", crypto.Address{}, []byte{0x1}, 0, _fee)
		_, rec3 := signAndExecute(t, e.ErrNone, tx3, "alice")
		require.Equal(t, rec3.Status, txs.Failed)

		tx4 := makeCallTx(t, "alice", crypto.Address{}, code, getBalance(t, "alice")+1, _fee)
		_, rec4 := signAndExecute(t, e.ErrInsufficientFunds, tx4, "alice")
		require.Equal(t, rec4.Status, txs.Failed)
	}

	require.Equal(t, sequence1+300, getAccountByName(t, "alice").Sequence())
}

func TestCallContract(t *testing.T) {
	/*
		pragma solidity ^0.4.0;

		contract SimpleStorage {
			function get() public constant returns (address) {
				return msg.sender;
			}
		}
	*/

	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, _ := hex.DecodeString("608060405234801561001057600080fd5b5060cc8061001f6000396000f300608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c146044575b600080fd5b348015604f57600080fd5b5060566098565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050905600a165627a7a7230582051d64d7178179a4ce4a151a49fe75742727f92a2becc39861f5a812c9e9bc00b0029")
	getFunc, _ := hex.DecodeString("6d4ce63c")

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Status, txs.Ok)

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
	tx2 := makeCallTx(t, "alice", contractAddr, getFunc, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Status, txs.Ok)
	addr1, _ := crypto.AccountAddress(rec2.Output[12:])
	addr2 := tx2.Caller().Address
	assert.Equal(t, addr1.String(), addr2.String())
}

func TestStorage(t *testing.T) {
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
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, _ := hex.DecodeString("608060405234801561001057600080fd5b5060df8061001f6000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680635362f8a214604e578063e1cb0e52146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820d95c60259bab2980d08fef23190fd2160ac19d7c103250a94e427f8d4d01b2030029")
	getFunc, _ := hex.DecodeString("e1cb0e52")
	setFunc, _ := hex.DecodeString("5362f8a2")

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Status, txs.Ok)
	seq2 := getAccountByName(t, "alice").Sequence()
	assert.Equal(t, seq2, seq1+1)

	contractAddr := *rec1.ContractAddress
	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	// empty storage
	tx11 := makeCallTx(t, "alice", contractAddr, getFunc, 0, _fee)
	_, rec11 := signAndExecute(t, e.ErrNone, tx11, "alice")
	assert.Equal(t, rec11.Output, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	// Input is the function hash of `setVal()`: 100
	retVal1, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000064")
	setData1 := addParams_3(setFunc, 100)
	tx2 := makeCallTx(t, "alice", contractAddr, setData1, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Status, txs.Ok)

	// Input is the function hash of `getVal()`
	tx3 := makeCallTx(t, "alice", contractAddr, getFunc, 0, _fee)
	_, rec3 := signAndExecute(t, e.ErrNone, tx3, "alice")
	assert.Equal(t, rec3.Output, retVal1)

	// Input is the function hash of `setVal()`: ccb4...
	retVal2, _ := hex.DecodeString("ccb49089f0f3c8339bef0ff8af2351740aefb9701c0c490f1b5528d8173c5de4")
	setData2 := setFunc
	setData2 = append(setData2, retVal2...)
	tx4 := makeCallTx(t, "alice", contractAddr, setData2, 0, _fee)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "alice")
	assert.Equal(t, rec4.Status, txs.Ok)

	// Input is the function hash of `getVal()`
	tx5 := makeCallTx(t, "alice", contractAddr, getFunc, 0, _fee)
	_, rec5 := signAndExecute(t, e.ErrNone, tx5, "alice")
	assert.Equal(t, rec5.Output, retVal2)
}

func TestStorage2(t *testing.T) {
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
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, _ := hex.DecodeString(`608060405234801561001057600080fd5b5060405160208061012883398101806040528101908080519060200190929190505050806000819055505060df806100496000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680635362f8a214604e578063e1cb0e52146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820e6ff3a1bc432f3b22cb4f663e7a9f70e9c12e701dd92c2021609fd8481a0998f002900000000000000000000000000000000000000000000000000000000000000aa`)
	data, _ := hex.DecodeString("00000000000000000000000000000000000000000000000000000000000000aa")
	getValueFunc, _ := hex.DecodeString("e1cb0e52")
	setValueFunc, _ := hex.DecodeString("5362f8a2")

	// A single input, having the permission, should succeed
	seq1 := getAccountByName(t, "alice").Sequence()
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)

	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Status, txs.Ok)
	seq2 := getAccountByName(t, "alice").Sequence()
	assert.Equal(t, seq2, seq1+1)

	contractAddr := *rec1.ContractAddress
	contractAcc := getAccount(t, contractAddr)
	require.NotNil(t, contractAcc, "failed to create contract %s", contractAddr)

	// Input is the function hash of `getVal()`
	tx2 := makeCallTx(t, "alice", contractAddr, getValueFunc, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Status, txs.Ok)
	assert.Equal(t, rec2.Output, data)

	// Input is the function hash of `setVal()`: 100
	retVal1, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000064")
	setData1 := addParams_3(setValueFunc, 100)
	tx3 := makeCallTx(t, "alice", contractAddr, setData1, 0, _fee)
	_, rec3 := signAndExecute(t, e.ErrNone, tx3, "alice")
	assert.Equal(t, rec3.Status, txs.Ok)

	// Input is the function hash of `getVal()`
	tx4 := makeCallTx(t, "alice", contractAddr, getValueFunc, 0, _fee)
	_, rec4 := signAndExecute(t, e.ErrNone, tx4, "alice")
	assert.Equal(t, rec4.Status, txs.Ok)
	assert.Equal(t, rec4.Output, retVal1)
}

func TestSelfDestruct(t *testing.T) {
	/*
		pragma solidity ^0.4.18;

		contract SelfDestruct {
		    address private owner;


			function SelfDestruct() public {
				owner = msg.sender;
			}

			function kill() public{
				selfdestruct(owner);
			}

			function hello() public pure returns (string) {
			    return "hello";
			}
		}
	*/
	setPermissions(t, "alice", permission.CreateContract|permission.Call)

	code, _ := hex.DecodeString("6060604052341561000f57600080fd5b336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506101b18061005e6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806319ff1d211461005157806341c0e1b5146100df575b600080fd5b341561005c57600080fd5b6100646100f4565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100a4578082015181840152602081019050610089565b50505050905090810190601f1680156100d15780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34156100ea57600080fd5b6100f2610137565b005b6100fc610171565b6040805190810160405280600581526020017f68656c6c6f000000000000000000000000000000000000000000000000000000815250905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16ff5b6020604051908101604052806000815250905600a165627a7a723058204d604789a2ea6d0dfd5624a6b0b1069ada540935b2766d0681abdb7219d59f0a0029")
	killFunc, _ := hex.DecodeString("41c0e1b5")
	helloFunc, _ := hex.DecodeString("19ff1d21")

	// A single input, having the permission, should succeed
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, rec1.Status, txs.Ok)

	// Should succeed, calling hello
	tx2 := makeCallTx(t, "alice", *rec1.ContractAddress, helloFunc, 0, _fee)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, rec2.Status, txs.Ok)

	// Should succeed, calling kill!
	tx3 := makeCallTx(t, "alice", *rec1.ContractAddress, killFunc, 0, _fee)
	_, rec3 := signAndExecute(t, e.ErrNone, tx3, "alice")
	assert.Equal(t, rec3.Status, txs.Ok)

	// Should fail, calling hello again
	tx4 := makeCallTx(t, "alice", *rec1.ContractAddress, helloFunc, 0, _fee)
	_, rec4 := signAndExecute(t, e.ErrInvalidAddress, tx4, "alice")
	assert.Equal(t, rec4.Status, txs.Failed)
}

func addParams_1(code []byte, addr crypto.Address, data []byte) []byte {
	// add first argument: address
	ethAddr := addr.RawBytes()[2:22]
	padding := make([]byte, 32-(len(ethAddr)%32))

	code = append(code, padding...)
	code = append(code, ethAddr...)

	// add second argument: data
	// offset of byte array
	offset, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000040")
	code = append(code, offset...)

	// length of data
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(len(data)))
	padding = make([]byte, 28)
	code = append(code, padding...)
	code = append(code, bs...)

	// data
	padding = make([]byte, 32-(len(data)%32))
	code = append(code, data...)
	code = append(code, padding...)

	return code
}

func addParams_2(code []byte, val1, val2 uint32) []byte {
	bs := make([]byte, 4)
	padding := make([]byte, 28)
	binary.BigEndian.PutUint32(bs, uint32(val1))
	code = append(code, padding...)
	code = append(code, bs...)

	binary.BigEndian.PutUint32(bs, uint32(val2))
	code = append(code, padding...)
	code = append(code, bs...)

	return code
}

func addParams_3(code []byte, val1 uint32) []byte {
	bs := make([]byte, 4)
	padding := make([]byte, 28)
	binary.BigEndian.PutUint32(bs, uint32(val1))
	code = append(code, padding...)
	code = append(code, bs...)

	return code
}

func addParams_4(code []byte, addr crypto.Address) []byte {
	ethAddr := addr.RawBytes()[2:22]
	padding := make([]byte, 32-(len(ethAddr)%32))

	code = append(code, padding...)
	code = append(code, ethAddr...)

	return code
}
