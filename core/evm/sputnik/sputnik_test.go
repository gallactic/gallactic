package sputnik

import (
	"encoding/hex"
	"fmt"
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
	pk, _ := crypto.GenerateKey(nil)
	val1, _ := validator.NewValidator(pk, 0)
	vals := []*validator.Validator{val1}

	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	gen := proposal.MakeGenesis("bar", time.Now().Truncate(0), gAcc, nil, nil, vals)
	db := dbm.NewMemDB()
	bc, err := blockchain.LoadOrNewBlockchain(db, gen, nil, logging.NewNoopLogger())

	require.NoError(t, err)

	callerAddr := convertEthAddress("a54fc84e16b4af78e7d1288114e7dcb9397daac8")
	var callerStr string
	callerStr = callerAddr.String()
	fmt.Println("Caller:", callerAddr.RawBytes())
	fmt.Println("Caller:", callerStr)


	tmpAddr := convertEthAddress("4acb57e88f38dceecf8d2ac1b13ec2a397d88491")
	calleeAddr:=crypto.DeriveContractAddress(tmpAddr,0)
	var calleeStr string
	calleeStr = calleeAddr.String()
	fmt.Println("Callee:", calleeAddr.RawBytes())
	fmt.Println("Callee:", calleeStr)

	caller, _ := account.NewAccount(callerAddr)
	callee, _ := account.NewContractAccount(calleeAddr)

	testCode := createContractCode()
	fmt.Println("CODE: ", testCode)

	txDep, _ := tx.NewCallTx(callerAddr, calleeAddr, 1, testCode, 10000000, 0, 1)

	var gas uint64

	st := state.NewState(db, logging.NewNoopLogger())
	cache := state.NewCache(st)

	caller.AddToBalance(1000000)
	callee.AddToBalance(2000000)
	caller.SetCode([]byte{})
	callee.SetCode(testCode)

	cache.UpdateAccount(caller)

	//Execute a non-exist Account
	/*
		fmt.Println("\n========================================\nExecute a non-exist Account...\n========================================")
		outC, _ := Execute(bc, cache, caller, callee, txDep, &gas,false)
		fmt.Println("\n\nUSED GAS: ", gas)
		fmt.Println("OUTPUT: ", outC)
	*/

	cache.UpdateAccount(callee)

	//Deploy
	fmt.Println("\n========================================\nDeploy Contract...\n========================================")
	outD, _ := Execute(bc, cache, caller, callee, txDep, &gas, true)
	fmt.Println("\n\nUSED GAS: ", gas)
	fmt.Println("OUTPUT: ", outD)

	//Execute Set Method
	fmt.Println("\n========================================\nCall Set Method by 1234567...\n========================================")
	setMethod, _ := hex.DecodeString("60fe47b100000000000000000000000000000000000000000000000000000000000001c8")
	txSet, _ := tx.NewCallTx(caller.Address(), callee.Address(), 1, setMethod, 10000000, 10, 1)
	outS, _ := Execute(bc, cache, caller, callee, txSet, &gas, false)
	fmt.Println("\n\nUSED GAS: ", gas)
	fmt.Println("OUTPUT: ", outS)

	//Execute Get Method
	fmt.Println("\n========================================\nCall Get Method...\n========================================")
	getMethod, _ := hex.DecodeString("6d4ce63c")
	txGet, _ := tx.NewCallTx(caller.Address(), callee.Address(), 1, getMethod, 10000000, 0, 1)
	outG, _ := Execute(bc, cache, caller, callee, txGet, &gas, false)
	fmt.Println("\n\nUSED GAS: ", gas)
	fmt.Println("OUTPUT: ", outG)
}

func createContractCode() []byte {
	deployCode, _ := hex.DecodeString("60806040526212d6876001556216e36060005560e9806100206000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146058575b600080fd5b6056600435607c565b005b348015606357600080fd5b50606a60b7565b60408051918252519081900360200190f35b60018190556040805182815290517f23f9887eb044d32dba99d7b0b753c61c3c3b72d70ff0addb9a843542fd7642129181900360200190a150565b600154905600a165627a7a7230582013452558cc58a514b8056c0b45a3f1ab8c5f736b2e087c65e615650b562415ff0029")
	return deployCode
}

func convertEthAddress(ethAddr string) crypto.Address {

	var addr common.Address
	addr.SetString(ethAddr)
	sputnikAddr, err := crypto.AccountAddress(addr.Bytes())
	if err != nil {
		//panic("cannot convert to burrow address")
		return sputnikAddr
	}

	return sputnikAddr
}




