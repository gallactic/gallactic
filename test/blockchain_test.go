package test

import (
	"testing"

	"github.com/gallactic/gallactic/core/blockchain"
	dbm "github.com/tendermint/tmlibs/db"
)

func setupBlockchain(m *testing.M) {
	bc1Db = dbm.NewMemDB()
	bc1, _ = blockchain.LoadOrNewBlockchain(bc1Db, genesisDoc, nopLogger)
	bc1State = bc1.State()
}

/*
func updateAccount(t *testing.T, acc *account.Account) {
	_, err := bc1State.Update(func(ws execution.Updatable) error {
		return ws.UpdateAccount(account)
	})
	require.NoError(t, err)
}
*/
