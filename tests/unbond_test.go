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
	require.NoError(t, err)
	return tx
}

func TestUnbondTxFails(t *testing.T) {

}

func TestUnbondTx(t *testing.T) {
	setPermissions(t, "alice", permission.Bond)

	stake1 := getValidatorByName(t, "val_1").Stake()
	tx1 := makeUnbondTx(t, "val_1", "bob", 9999, _fee)
	signAndExecute(t, e.ErrNone, tx1, "val_1")
	stake2 := getValidatorByName(t, "val_1").Stake()
	assert.Equal(t, stake2, stake1-(9999+_fee))
}
