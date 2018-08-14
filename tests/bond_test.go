package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/require"
)

func makeBondTx(t *testing.T, from, to string, amt, fee uint64) *tx.BondTx {
	acc := getAccountByName(t, from)
	var toPb crypto.PublicKey
	if to != "" {
		toPb = tValidators[to].PublicKey()
	} else {
		toPb, _ = crypto.GenerateKey(nil)
	}

	tx, err := tx.NewBondTx(acc.Address(), toPb, amt, acc.Sequence()+1, fee)
	require.Equal(t, amt, tx.Amount())
	require.Equal(t, fee, tx.Fee())
	require.NoError(t, err)
	return tx
}

func TestBondTxFails(t *testing.T) {
	setPermissions(t, "alice", permission.Bond)
	setPermissions(t, "bob", permission.Send)

	tx1 := makeBondTx(t, "alice", "", 9999, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	tx2 := makeBondTx(t, "alice", "val_1", 9999, _fee)
	signAndExecute(t, e.ErrNone, tx2, "alice")

	tx3 := makeBondTx(t, "bob", "", 9999, _fee)
	signAndExecute(t, e.ErrPermDenied, tx3, "bob")
}

func TestBondTx(t *testing.T) {
	setPermissions(t, "alice", permission.Bond)
	setPermissions(t, "bob", permission.Bond)

	stake1 := getValidatorByName(t, "val_1").Stake()
	aliceBalance := getBalance(t, "alice")
	bobBalance := getBalance(t, "bob")

	tx1 := makeBondTx(t, "alice", "val_1", 9999, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	tx2 := makeBondTx(t, "bob", "val_1", 9999, _fee)
	signAndExecute(t, e.ErrNone, tx2, "bob")

	stake2 := getValidatorByName(t, "val_1").Stake()
	assert.Equal(t, stake2, stake1+(2*9999))

	checkBalance(t, "alice", aliceBalance-(9999+_fee))
	checkBalance(t, "bob", bobBalance-(9999+_fee))
}
