package sputnik

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestSputnikVM(t *testing.T) {
	//Create blockchain
	pk, _ := crypto.GenerateKey(nil)
	val1, _ := validator.NewValidator(pk, 0)
	vals := []*validator.Validator{val1}

	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	gen := proposal.MakeGenesis("bar", time.Now().Truncate(0), gAcc, nil, nil, vals)
	db := dbm.NewMemDB()
	bc, err := blockchain.LoadOrNewBlockchain(db, gen, nil, logging.NewNoopLogger())

	require.NoError(t, err)

	callerAddr := toGallecticAddress("6ac7ea33f8831ea9dcc53393aaa88b25a785dbf0")
	caller, _ := account.NewAccount(callerAddr)
	caller.AddToBalance(1000000)
	caller.SetCode([]byte{})

	st := state.NewState(db, logging.NewNoopLogger())
	cache := state.NewCache(st)
	cache.UpdateAccount(caller)

	adapter1 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: nil, GasLimit: 1000000, Amount: 0, Data: []byte{}, Nonce: 1}
	_, errDeployZ := Execute(&adapter1)
	require.Error(t, errDeployZ)

	//Deploy a random contract.
	adapter2 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: nil, GasLimit: 1000000, Amount: 0, Data: []byte{60, 80, 120, 48, 22, 8, 0, 0, 34}, Nonce: 2}
	_, errDeployE := Execute(&adapter2)
	require.Error(t, errDeployE)

	//Deploy a valid contract
	testCode := createContractCode()
	adapter3 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: nil, GasLimit: 1000000, Amount: 0, Data: testCode, Nonce: 3}
	outD, errDeploy := Execute(&adapter3)
	require.NoError(t, errDeploy)
	require.Equal(t, getContractCodeAfterDeploy(), hex.EncodeToString(outD.Output))

	callee, _ := cache.GetAccount(adapter3.Callee.Address())

	//Call none exist method
	noneMethod, _ := hex.DecodeString("c0ae47d2")
	adapter4 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: callee, GasLimit: 1000000, Amount: 0, Data: noneMethod, Nonce: 4}
	outN, errNone := Execute(&adapter4)

	require.Error(t, errNone)
	require.Equal(t, 0, len(outN.Output))

	//Call SetMethod() by 1234567 as parameter
	setMethod, _ := hex.DecodeString("60fe47b100000000000000000000000000000000000000000000000000000000000001c8")
	adapter5 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: callee, GasLimit: 1000000, Amount: 0, Data: setMethod, Nonce: 5}
	outS, errSet := Execute(&adapter5)
	require.NoError(t, errSet)
	require.Equal(t, 0, len(outS.Output))

	//Call Get() Method...
	getMethod, _ := hex.DecodeString("6d4ce63c")
	adapter6 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: callee, GasLimit: 1000000, Amount: 0, Data: getMethod, Nonce: 6}
	outG, errGetG := Execute(&adapter6)
	require.NoError(t, errGetG)
	require.Equal(t, "00000000000000000000000000000000000000000000000000000000000001c8", hex.EncodeToString(outG.Output))

	//Call GetOwner() Method...
	getOwnerMethod, _ := hex.DecodeString("893d20e8")
	adapter7 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: callee, GasLimit: 1000000, Amount: 0, Data: getOwnerMethod, Nonce: 7}
	outW, errGetW := Execute(&adapter7)
	require.NoError(t, errGetW)
	require.Equal(t, ToEthAddress(caller.Address()).Bytes(), outW.Output[12:])

	//Call kill() Method...
	killMethod, _ := hex.DecodeString("41c0e1b5")
	adapter8 := GallacticAdapter{BlockChain: bc, Cache: cache, Caller: caller,
		Callee: callee, GasLimit: 1000000, Amount: 0, Data: killMethod, Nonce: 7}
	_, errK := Execute(&adapter8)
	require.NoError(t, errK)

}

func createContractCode() []byte {
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
	deployCode, _ := hex.DecodeString("608060405234801561001057600080fd5b5033600260006101000a815481600160a060020a030219169083600160a060020a031602179055506101be806100476000396000f3006080604052600436106100615763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166341c0e1b5811461006657806360fe47b11461007d5780636d4ce63c14610088578063893d20e8146100b0575b600080fd5b34801561007257600080fd5b5061007b610107565b005b61007b60043561012d565b34801561009457600080fd5b5061009d610168565b6040805191825251602090910181900390f35b3480156100bc57600080fd5b506100c561016e565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60025473ffffffffffffffffffffffffffffffffffffffff60006101000a909104811616ff5b60018190556040805182815290517f23f9887eb044d32dba99d7b0b753c61c3c3b72d70ff0addb9a843542fd7642129160200181900390a150565b60015490565b60025460006101000a900473ffffffffffffffffffffffffffffffffffffffff16905600a165627a7a7230582001a5bb7dbc53c4e0e7acc1b23010f4dd1415e0b440e8784ac8ce8d0696c841720029")
	return deployCode
}
func getContractCodeAfterDeploy() string {
	return "6080604052600436106100615763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166341c0e1b5811461006657806360fe47b11461007d5780636d4ce63c14610088578063893d20e8146100b0575b600080fd5b34801561007257600080fd5b5061007b610107565b005b61007b60043561012d565b34801561009457600080fd5b5061009d610168565b6040805191825251602090910181900390f35b3480156100bc57600080fd5b506100c561016e565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60025473ffffffffffffffffffffffffffffffffffffffff60006101000a909104811616ff5b60018190556040805182815290517f23f9887eb044d32dba99d7b0b753c61c3c3b72d70ff0addb9a843542fd7642129160200181900390a150565b60015490565b60025460006101000a900473ffffffffffffffffffffffffffffffffffffffff16905600a165627a7a7230582001a5bb7dbc53c4e0e7acc1b23010f4dd1415e0b440e8784ac8ce8d0696c841720029"
}

func toGallecticAddress(ethAddr string) crypto.Address {

	var addr common.Address
	sss, _ := hex.DecodeString(ethAddr)
	addr.SetBytes(sss) //SetString(ethAddr)
	sputnikAddr, err := crypto.AccountAddress(addr.Bytes())
	if err != nil {
		return sputnikAddr
	}

	return sputnikAddr
}
