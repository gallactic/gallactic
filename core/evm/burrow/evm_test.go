package burrow

import (
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestVM(t *testing.T) {
	val1, _ := validator.NewValidator(crypto.GeneratePrivateKey(nil).PublicKey(), 0)
	vals := []*validator.Validator{val1}

	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	gen := proposal.MakeGenesis("bar", time.Now().Truncate(0), gAcc, nil, nil, vals)
	db := dbm.NewMemDB()
	bc, err := blockchain.LoadOrNewBlockchain(db, gen, logging.NewNoopLogger())
	require.NoError(t, err)

	callerAddr, _ := crypto.AddressFromString("ac9E2cyNA5UfB8pUpqzEz4QCcBpp8sxnEaN")
	calleeAddr := crypto.GlobalAddress
	caller, _ := account.NewAccount(callerAddr)
	callee, _ := account.NewAccount(calleeAddr)
	callee.SetCode(createContractCode())
	tx, _ := tx.NewCallTx(caller.Address(), calleeAddr, 1, []byte{1}, 2100, 0, 100)
	var gas uint64

	Call(bc, caller, callee, tx, &gas)
}

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
