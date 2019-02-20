package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
)

type SendContext struct {
	Committing bool
	Cache      *state.Cache
}

func (ctx *SendContext) Execute(txEnv *txs.Envelope, txRec *txs.Receipt) error {
	tx, ok := txEnv.Tx.(*tx.SendTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	accs := make(map[crypto.Address]*account.Account)
	err := getInputAccounts(ctx.Cache, tx.Senders(), permission.Send, accs)
	if err != nil {
		return err
	}

	err = getOutputAccounts(ctx.Cache, tx.Receivers(), accs)
	if err != nil {
		return err
	}

	for _, r := range tx.Receivers() {
		if accs[r.Address] == nil {
			/// check for CreateAccount permission
			for _, s := range tx.Senders() {
				acc := accs[s.Address]
				if !ctx.Cache.HasPermissions(acc, permission.CreateAccount) {
					return e.Errorf(e.ErrPermissionDenied, "%s has %s but needs %s", r.Address, acc.Permissions(), permission.CreateAccount)
				}
			}
		}
	}

	/// Create accounts
	for _, r := range tx.Receivers() {
		if accs[r.Address] == nil {
			accs[r.Address], err = account.NewAccount(r.Address)
			if err != nil {
				return e.Errorf(e.ErrInvalidAddress, "%s is not an account address", r.Address)
			}
		}
	}

	// Good! Adjust accounts
	err = adjustInputAccounts(accs, tx.Senders())
	if err != nil {
		return err
	}

	err = adjustOutputAccounts(accs, tx.Receivers())
	if err != nil {
		return err
	}

	/// Update state cache
	for _, acc := range accs {
		ctx.Cache.UpdateAccount(acc)
	}

	return nil
}
