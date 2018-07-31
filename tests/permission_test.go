package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/require"
)

/// TODO : test fro sequence increase. +1 after tx get successfull. 0: if not successfull. +2: for create contract

func makePermissionTx(t *testing.T, modifier, modified string, perm account.Permissions, set bool, fee uint64) *tx.PermissionsTx {
	acc1 := getAccountByName(t, modifier)
	acc2 := getAccountByName(t, modified)
	tx, err := tx.NewPermissionsTx(acc1.Address(), acc2.Address(), perm, set, acc1.Sequence()+1, fee)
	require.NoError(t, err)

	return tx
}

func setPermissions(t *testing.T, name string, perm account.Permissions) {
	acc := getAccountByName(t, name)
	/// First remove all permissions, then set new one
	acc.UnsetPermissions(acc.Permissions())
	acc.SetPermissions(perm)
	updateAccount(t, acc)

	commit(t)
}

func TestPermissionsTxFails(t *testing.T) {
	setPermissions(t, "alice", permission.ModifyPermission)
	setPermissions(t, "bob", permission.Send)

	tx1 := makePermissionTx(t, "alice", "dan", permission.Call, true, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	tx2 := makePermissionTx(t, "bob", "dan", permission.Call, true, _fee)
	signAndExecute(t, e.ErrPermDenied, tx2, "bob")
}

func TestPermissionsTx(t *testing.T) {
	setPermissions(t, "alice", permission.ModifyPermission)
	setPermissions(t, "bob", permission.Send)

	tx1 := makePermissionTx(t, "alice", "bob", permission.Call, true, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")
	assert.Equal(t, getAccountByName(t, "bob").Permissions(), permission.Send|permission.Call)

	tx2 := makePermissionTx(t, "alice", "bob", permission.Call, false, _fee)
	signAndExecute(t, e.ErrNone, tx2, "alice")
	assert.Equal(t, getAccountByName(t, "bob").Permissions(), permission.Send)
}
