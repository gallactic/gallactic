package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/hyperledger/burrow/logging"
)

type SendContext struct {
	Committer bool
	State     *state.State
	Logger    *logging.Logger
}

func (ctx *SendContext) Execute(txEnv *txs.Envelope) error {
	tx, ok := txEnv.Tx.(*tx.SendTx)
	if !ok {
		return e.Error(e.ErrTxWrongPayload)
	}

	objs, err := getStateObjs(ctx.State, tx)
	if err != nil {
		return err
	}

	for _, in := range tx.Inputs() {
		acc, ok := objs[in.Address].(*account.Account)
		if !ok {
			return e.Errorf(e.ErrInvalidAddress, "%s is not an account address", in.Address)
		}
		if !HasSendPermission(ctx.State, acc) {
			return e.Errorf(e.ErrPermDenied, "%s has %s but needs %s", in.Address, acc.Permissions(), permission.Send)
		}
	}

	for _, out := range tx.Outputs() {
		if objs[out.Address] == nil {
			/// check for CreateAccount permission
			for _, in := range tx.Inputs() {
				acc, _ := objs[in.Address].(*account.Account)
				if !HasCreateAccountPermission(ctx.State, acc) {
					return e.Errorf(e.ErrPermDenied, "%s has %s but needs %s", in.Address, acc.Permissions(), permission.CreateAccount)
				}
			}
		}
	}

	/// Create accounts
	for _, out := range tx.Outputs() {
		if objs[out.Address] == nil {
			objs[out.Address], err = account.NewAccount(out.Address)
			if err != nil {
				return e.Errorf(e.ErrInvalidAddress, "%s is not an account address", out.Address)
			}
		}
	}

	// Good! Adjust accounts
	err = adjustInputs(objs, tx.Inputs())
	if err != nil {
		return err
	}

	err = adjustOutputs(objs, tx.Outputs())
	if err != nil {
		return err
	}

	/*
		for _, obj := range objs {
			ctx.State.UpdateObj(obj)
		}
	*/

	return nil
}
