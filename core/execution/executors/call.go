package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/evm/sputnik"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
)

type CallContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *CallContext) Execute(txEnv *txs.Envelope) error {
	tx, ok := txEnv.Tx.(*tx.CallTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	caller, err := getInputAccount(ctx.Cache, tx.Caller(), permission.Call)
	if err != nil {
		return err
	}

	var callee *account.Account
	if tx.CreateContract() {
		if !ctx.Cache.HasPermissions(caller, permission.CreateContract) {
			return e.Errorf(e.ErrPermDenied, "%s has %s but needs %s", caller.Address(), caller.Permissions(), permission.CreateContract)
		}

		callee, err = deriveNewAccount(caller)
		if err != nil {
			return err
		}
		callee.SetCode(tx.Data())
		ctx.Logger.TraceMsg("Creating new contract",
			"contract_address", callee.Address(),
			"init_code", tx.Data())
	} else {
		/// TODO : write test for this case: create and call in same block
		callee, err = ctx.Cache.GetAccount(tx.Callee().Address)
		if err != nil {
			return err
		}
	}

	if ctx.Committing {
		err := ctx.Deliver(tx, caller, callee)
		if err != nil {
			return err
		}
	}

	err = caller.SubtractFromBalance(tx.Fee())
	if err != nil {
		return err
	}

	/// Update state cache
	ctx.Cache.UpdateAccount(caller)
	ctx.Cache.UpdateAccount(callee)

	return nil
}

func (ctx *CallContext) Deliver(tx *tx.CallTx, caller, callee *account.Account) error {

	if callee == nil || len(callee.Code()) == 0 {
		// if you call an account that doesn't exist
		// or an account with no code then we take fees (sorry pal)
		// NOTE: it's fine to create a contract and call it within one
		// block (sequence number will prevent re-ordering of those txs)
		// but to create with one contract and call with another
		// you have to wait a block to avoid a re-ordering attack
		// that will take your fees
		if callee == nil {
			panic("panic_test")
			ctx.Logger.InfoMsg("Execute to address that does not exist",
				"caller_address", tx.Caller(),
				"callee_address", tx.Callee())
		} else {
			ctx.Logger.InfoMsg("Execute to address that holds no code",
				"caller_address", tx.Caller(),
				"callee_address", tx.Callee())
		}

		return nil
	}

	var gas uint64
	ret, err := sputnik.Execute(ctx.BC, ctx.Cache, caller, callee, tx, &gas, false)

	if err != nil {
		return err
	}
	if tx.CreateContract() {
		callee.SetCode(ret)
	}
	code := callee.Code()
	ctx.Logger.TraceMsg("Calling existing contract",
		"contract_address", callee.Address(),
		"input", tx.Data(),
		"contract_code", code)

	ctx.Logger.Trace.Log("callee", callee.Address().String())
	// Create a receipt from the ret and whether it erred.
	ctx.Logger.TraceMsg("VM call complete",
		"caller", caller,
		"callee", callee,
		"return", ret,
		structure.ErrorKey, err)

	return nil
}

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func deriveNewAccount(creator *account.Account) (*account.Account, error) {
	// Generate an address
	seq := creator.Sequence()
	creator.IncSequence()

	addr := crypto.DeriveContractAddress(creator.Address(), seq)

	// Create account from address.
	acc, err := account.NewAccount(addr)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
