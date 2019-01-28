package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/evm/sputnikvm"
	"github.com/gallactic/gallactic/core/state"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
)

type CallContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *CallContext) Execute(txEnv *txs.Envelope, txRec *txs.Receipt) error {
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
			return e.Errorf(e.ErrPermissionDenied, "%s has %s but needs %s", caller.Address(), caller.Permissions(), permission.CreateContract)
		}

		// In case of create contract we must pass nil as callee
		// sputnik vm will create the account and returns the code
		callee = nil

		ctx.Logger.TraceMsg("Creating new contract", "init_code", tx.Data())
	} else {
		callee, _ = ctx.Cache.GetAccount(tx.Callee().Address)
		if callee == nil {
			return e.Errorf(e.ErrInvalidAddress, "attempt to call a non-existing account: %s", tx.Callee().Address)
		}
	}

	if ctx.Committing {
		ret := ctx.Deliver(tx, caller, callee)

		caller.IncSequence()

		//Here we can acquire sputnik VM result
		if ret.Failed {
			txRec.Status = txs.Failed
		} else {
			txRec.Status = txs.Ok
		}
		txRec.GasUsed = ret.UsedGas
		txRec.GasWanted = tx.GasLimit()
		txRec.Output = ret.Output
		txRec.ContractAddress = ret.ContractAddress
	}

	err = caller.SubtractFromBalance(tx.Fee())
	if err != nil {
		return err
	}

	/// Update state cache
	ctx.Cache.UpdateAccount(caller)

	return nil
}

func (ctx *CallContext) Deliver(tx *tx.CallTx, caller, callee *account.Account) sputnikvm.Output {
	defer func() {
		/* TODO:::: better crash now for testnet
		// NOTE: SputnikVM should never crash, report it as error message to the caller
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", txEnv.String(), r,
				debug.Stack())
		}
		*/
	}()

	adapter := sputnikvm.GallacticAdapter{ctx.BC, ctx.Cache, caller,
		callee, tx.GasLimit(), tx.Amount(), tx.Data(), caller.Sequence()}

	ret := sputnikvm.Execute(&adapter)

	// Create a receipt from the ret and whether it erred.
	ctx.Logger.TraceMsg("VM call complete",
		"caller", caller,
		"callee", callee,
		"return", ret)

	return ret
}
