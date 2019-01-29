package executors

import (
	"fmt"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
)

type PermissionContext struct {
	Committing bool
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *PermissionContext) Execute(txEnv *txs.Envelope, txRec *txs.Receipt) error {
	tx, ok := txEnv.Tx.(*tx.PermissionsTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	modifier, err := getInputAccount(ctx.Cache, tx.Modifier(), permission.ModifyPermission)
	if err != nil {
		return err
	}

	if tx.Modified().Address == crypto.GlobalAddress {
		return fmt.Errorf("Modifying global account is not allowed")
	}

	modified, err := getOutputAccount(ctx.Cache, tx.Modified())
	if err != nil {
		return err
	}

	if modified == nil {
		return fmt.Errorf("Try to modify non-existing account")
	}

	if !permission.EnsureValid(tx.Permissions()) {
		return e.Error(e.ErrInvalidPermission)
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

	/// Update state cache
	ctx.Cache.UpdateAccount(modified)
	ctx.Cache.UpdateAccount(modifier)

	return nil
}
