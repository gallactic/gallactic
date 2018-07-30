package executors

import (
	"fmt"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
)

type PermissionContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *PermissionContext) Execute(txEnv *txs.Envelope) error {
	tx, ok := txEnv.Tx.(*tx.PermissionsTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	modifier, err := getInputAccount(ctx.Cache, tx.Modifier(), permission.ModifyPermission)
	if err != nil {
		return err
	}

	modified, err := getOutputAccount(ctx.Cache, tx.Modified())
	if err != nil {
		return err
	}

	if modified == nil {
		return fmt.Errorf("Try to modify non-existing account")
	}

	if !permission.EnsureValid(tx.Permissions()) {
		return e.Error(e.ErrPermInvalid)
	}

	if tx.Set() {
		if err = modified.SetPermissions(tx.Permissions()); err != nil {
			return err
		}
	} else {
		if err = modified.UnsetPermissions(tx.Permissions()); err != nil {
			return err
		}
	}

	// Good! Adjust accounts
	err = adjustInputAccount(modifier, tx.Modifier())
	if err != nil {
		return err
	}

	err = adjustOutputAccount(modified, tx.Modified())
	if err != nil {
		return err
	}

	/// Update account
	ctx.Cache.UpdateAccount(modified)
	ctx.Cache.UpdateAccount(modifier)

	return nil
}
