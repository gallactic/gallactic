package state

import (
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func newState() *State {
	db := dbm.NewMemDB()
	return NewState(db)
}

func TestStateChanges(t *testing.T) {
	st := newState()
	pb, _ := crypto.GenerateKeyFromSecret("secret1")
	addr := pb.AccountAddress()
	acc, err := st.GetAccount(addr)
	assert.Error(t, err) // return an error
	assert.Nil(t, acc, nil)
	assert.Equal(t, st.AccountCount(), 0)

	acc1, _ := account.NewAccount(addr)
	acc1.AddToBalance(10)
	st.updateAccount(acc1)
	acc2, err := st.GetAccount(addr)
	assert.NoError(t, err)
	assert.Equal(t, acc1, acc2)
	assert.Equal(t, st.AccountCount(), 1)
}
