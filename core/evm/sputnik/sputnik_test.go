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
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestSputnikVM(t *testing.T) {
	//create block chain
	pk, _ := crypto.GenerateKey(nil)
	val1, _ := validator.NewValidator(pk, 0)
	vals := []*validator.Validator{val1}

	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	gen := proposal.MakeGenesis("bar", time.Now().Truncate(0), gAcc, nil, nil, vals)
	db := dbm.NewMemDB()
	bc, err := blockchain.LoadOrNewBlockchain(db, gen, nil, logging.NewNoopLogger())

	require.NoError(t, err)

	//create caller address
	callerAddr := convertEthAddress("a54fc84e16b4af78e7d1288114e7dcb9397daac8")
	//create callee address
	tmpAddr := convertEthAddress("4acb57e88f38dceecf8d2ac1b13ec2a397d88491")
	calleeAddr := crypto.DeriveContractAddress(tmpAddr, 0)

	//create caller and callee test accounts
	caller, _ := account.NewAccount(callerAddr)
	callee, _ := account.NewContractAccount(calleeAddr)

	//create sample smart contract code
	testCode := createContractCode()

	//create transaction structure
	txDep, _ := tx.NewCallTx(callerAddr, calleeAddr, 1, testCode, 10000000, 0, 1)

	var gas uint64

	//create new state and cache
	st := state.NewState(db, logging.NewNoopLogger())
	cache := state.NewCache(st)

	//update test accounts
	caller.AddToBalance(1000000)
	callee.AddToBalance(2000000)
	caller.SetCode([]byte{})
	callee.SetCode(testCode)

	cache.UpdateAccount(caller)

	//Execute a non-exist Account...
	outC, err := Execute(bc, cache, caller, callee, txDep, &gas, false)
	require.Error(t, err)
	require.Equal(t, outC, []byte{})

	//Now we add callee address and execute vm again
	cache.UpdateAccount(callee)

	//Deploy Contract...
	outD, errDeploy := Execute(bc, cache, caller, callee, txDep, &gas, true)
	require.NoError(t, errDeploy)
	require.Equal(t, hex.EncodeToString(outD), getContractCodeAfterDeploy())

	//Call Set Method by 1234567...
	setMethod, _ := hex.DecodeString("60fe47b100000000000000000000000000000000000000000000000000000000000001c8")
	txSet, _ := tx.NewCallTx(caller.Address(), callee.Address(), 1, setMethod, 10000000, 10, 1)
	outS, errSet := Execute(bc, cache, caller, callee, txSet, &gas, false)
	require.NoError(t, errSet)
	require.Equal(t, outS, []byte{})

	//Call Get Method...
	getMethod, _ := hex.DecodeString("6d4ce63c")
	txGet, _ := tx.NewCallTx(caller.Address(), callee.Address(), 1, getMethod, 10000000, 0, 1)
	outG, errGet := Execute(bc, cache, caller, callee, txGet, &gas, false)
	require.NoError(t, errGet)
	require.Equal(t, hex.EncodeToString(outG), "00000000000000000000000000000000000000000000000000000000000001c8")
}

func createContractCode() []byte {
	//Test Smart Contract
	/*
		pragma solidity ^0.4.24;
		contract SimpleStorage {
			uint private _balance;
			uint private _storedData;
			event notifyStorage(uint x);
			constructor() public payable {
				_storedData = 0x12d687;
				_balance = 1500000;
			}
			function set(uint x) public payable {
				_storedData = x;
				emit notifyStorage(x);
			}
			function get() public view returns (uint) {
				return _storedData;
			}
		}
	*/
	deployCode, _ := hex.DecodeString("60806040526212d6876001556216e36060005560e9806100206000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146058575b600080fd5b6056600435607c565b005b348015606357600080fd5b50606a60b7565b60408051918252519081900360200190f35b60018190556040805182815290517f23f9887eb044d32dba99d7b0b753c61c3c3b72d70ff0addb9a843542fd7642129181900360200190a150565b600154905600a165627a7a7230582013452558cc58a514b8056c0b45a3f1ab8c5f736b2e087c65e615650b562415ff0029")
	return deployCode
}
func getContractCodeAfterDeploy() string {
	return "60806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146058575b600080fd5b6056600435607c565b005b348015606357600080fd5b50606a60b7565b60408051918252519081900360200190f35b60018190556040805182815290517f23f9887eb044d32dba99d7b0b753c61c3c3b72d70ff0addb9a843542fd7642129181900360200190a150565b600154905600a165627a7a7230582013452558cc58a514b8056c0b45a3f1ab8c5f736b2e087c65e615650b562415ff0029"
}

func convertEthAddress(ethAddr string) crypto.Address {

	var addr common.Address
	addr.SetString(ethAddr)
	sputnikAddr, err := crypto.AccountAddress(addr.Bytes())
	if err != nil {
		return sputnikAddr
	}

	return sputnikAddr
}
