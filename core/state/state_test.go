package state

import (
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func loadState(t *testing.T, db dbm.DB, hash []byte) *State {
	s, err := LoadState(db, hash, logging.NewNoopLogger())
	require.NoError(t, err)
	require.NotNil(t, s)

	return s
}

func saveState(t *testing.T, state *State) []byte {
	hash, err := state.SaveState()
	require.NoError(t, err)

	fmt.Printf("hash:%v\n", hash)

	return hash
}

func updateAccount(t *testing.T, state *State, acc *account.Account) {
	err := state.UpdateAccount(acc)
	require.NoError(t, err)
}

func getAccount(t *testing.T, state *State, addr crypto.Address) *account.Account {
	account := state.GetAccount(addr)
	return account
}

func TestState_LoadingWrongHash(t *testing.T) {
	db := dbm.NewMemDB()
	s0, err := LoadState(db, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, logging.NewNoopLogger())
	require.Error(t, err)
	require.Nil(t, s0)
}

func TestState_Loading(t *testing.T) {
	db := dbm.NewMemDB()
	state := NewState(db, logging.NewNoopLogger())

	foo := account.NewAccountFromSecret("Foo")
	bar := account.NewAccountFromSecret("Bar")

	foo.AddToBalance(1)
	bar.AddToBalance(1)

	updateAccount(t, state, foo)
	updateAccount(t, state, bar)

	hash1 := saveState(t, state)
	hash2 := saveState(t, state)
	require.Equal(t, hash1, hash2)

	foo.AddToBalance(1)
	updateAccount(t, state, foo)
	hash3 := saveState(t, state)

	require.NotEqual(t, hash1, hash3)

	/// --- Immutable saved state
	state1 := loadState(t, db, hash1)
	foo.AddToBalance(1)
	_, err := state1.SaveState()
	require.Error(t, err)
	/// ---

	state2 := loadState(t, db, hash2)
	state3 := loadState(t, db, hash3)

	foo2 := getAccount(t, state2, foo.Address())
	foo3 := getAccount(t, state3, foo.Address())

	require.Equal(t, uint64(1), foo2.Balance())
	require.Equal(t, uint64(2), foo3.Balance())
}

func TestState_Loading2(t *testing.T) {
	db := dbm.NewMemDB()
	state := NewState(db, logging.NewNoopLogger())

	foo := account.NewAccountFromSecret("Foo")
	bar := account.NewAccountFromSecret("Bar")

	foo.AddToBalance(1)
	bar.AddToBalance(1)

	updateAccount(t, state, foo)
	hash1 := saveState(t, state)

	updateAccount(t, state, bar)
	hash2 := saveState(t, state)

	require.NotEqual(t, hash1, hash2)

	foo2 := getAccount(t, state, foo.Address())
	bar2 := getAccount(t, state, bar.Address())

	require.Equal(t, uint64(1), foo2.Balance())
	require.Equal(t, uint64(1), bar2.Balance())

	state2 := loadState(t, db, hash2)

	foo3 := getAccount(t, state2, foo.Address())
	bar3 := getAccount(t, state2, bar.Address())

	require.Equal(t, uint64(1), foo3.Balance())
	require.Equal(t, uint64(1), bar3.Balance())
}

func TestState_UpdateAccount(t *testing.T) {
	state := NewState(dbm.NewMemDB(), logging.NewNoopLogger())
	foo1 := account.NewAccountFromSecret("Foo")

	foo1.AddToBalance(1)
	foo1.SetCode([]byte{0x60})
	updateAccount(t, state, foo1)

	foo2 := getAccount(t, state, foo1.Address())
	assert.Equal(t, foo1, foo2)
}
