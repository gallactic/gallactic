package tests

import (
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeUnbondTx(t *testing.T, from, to string, amount, fee uint64) *tx.UnbondTx {
	val := getValidatorByName(t, from)
	acc := getAccountByName(t, to)
	tx, err := tx.NewUnbondTx(val.Address(), acc.Address(), amount, val.Sequence()+1, fee)
	require.Equal(t, amount, tx.Amount())
	require.Equal(t, fee, tx.Fee())
	require.NoError(t, err)
	return tx
}

func TestUnbondTxFails(t *testing.T) {

}

func TestUnbondTx(t *testing.T) {
	stake1 := getValidatorByName(t, "val_1").Stake()
	tx1 := makeUnbondTx(t, "val_1", "bob", 9999, _fee)
	signAndExecute(t, e.ErrNone, tx1, "val_1")
	stake2 := getValidatorByName(t, "val_1").Stake()
	assert.Equal(t, stake2, stake1-(9999+_fee))
}

func TestUnbondTxSequence(t *testing.T) {
	setPermissions(t, "alice", permission.Bond)

	val1 := getValidatorByName(t, "val_1")
	seq1 := val1.Sequence()
	seq2 := getAccountByName(t, "alice").Sequence()

	for i := 0; i < 100; i++ {
		tx := makeUnbondTx(t, "val_1", "alice", 9999, _fee)
		signAndExecute(t, e.ErrNone, tx, "val_1")

		invalidTx := makeUnbondTx(t, "val_1", "alice", val1.Stake()+1, _fee)
		signAndExecute(t, e.ErrInsufficientFunds, invalidTx, "val_1")
	}

	require.Equal(t, seq1+100, getValidatorByName(t, "val_1").Sequence())
	require.Equal(t, seq2, getAccountByName(t, "alice").Sequence())
}
