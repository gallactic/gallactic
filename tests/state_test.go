package tests

import (
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func loadState(t *testing.T, hash []byte) *state.State {
	s, err := state.LoadState(tDB, hash, tLogger)
	require.NoError(t, err)
	require.NotNil(t, s)

	return s
}

func saveState(t *testing.T) []byte {
	hash, err := tState.SaveState()
	require.NoError(t, err)

	fmt.Printf("hash:%v\n", hash)

	return hash
}

func updateAccount(t *testing.T, acc *account.Account) {
	err := tState.UpdateAccount(acc)
	require.NoError(t, err)
}

func getAccountByName(t *testing.T, name string) *account.Account {
	acc := getAccount(t, tAccounts[name].Address())
	require.NotNil(t, acc)
	return acc
}

func getAccount(t *testing.T, addr crypto.Address) *account.Account {
	acc, err := tState.GetAccount(addr)
	assert.NoError(t, err)
	return acc
}

func getValidatorByName(t *testing.T, name string) *validator.Validator {
	val := getValidator(t, tValidators[name].Address())
	require.NotNil(t, val)
	return val
}

func getValidator(t *testing.T, addr crypto.Address) *validator.Validator {
	val, err := tState.GetValidator(addr)
	assert.NoError(t, err)
	return val
}

func TestState_LoadingWrongHash(t *testing.T) {
	db := dbm.NewMemDB()
	s0, err := state.LoadState(db, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, tLogger)
	require.Error(t, err)
	require.Nil(t, s0)
}

func TestState_Loading(t *testing.T) {
	foo := account.NewAccountFromSecret("Foo")
	bar := account.NewAccountFromSecret("Bar")

	foo.AddToBalance(1)
	bar.AddToBalance(1)

	updateAccount(t, foo)
	updateAccount(t, bar)

	hash1 := saveState(t)
	hash2 := saveState(t)
	require.Equal(t, hash1, hash2)

	foo.AddToBalance(1)
	updateAccount(t, foo)
	hash3 := saveState(t)

	require.NotEqual(t, hash1, hash3)

	/// --- Immutable saved state
	state1 := loadState(t, hash1)
	foo.AddToBalance(1)
	_, err := state1.SaveState()
	require.Error(t, err)
	/// ---

	st2 := loadState(t, hash2)
	st3 := loadState(t, hash3)

	foo2, _ := st2.GetAccount(foo.Address())
	foo3, _ := st3.GetAccount(foo.Address())

	require.Equal(t, uint64(1), foo2.Balance())
	require.Equal(t, uint64(2), foo3.Balance())
}

func TestState_Loading2(t *testing.T) {
	foo := account.NewAccountFromSecret("Foo")
	bar := account.NewAccountFromSecret("Bar")

	foo.AddToBalance(1)
	bar.AddToBalance(1)

	updateAccount(t, foo)
	hash1 := saveState(t)

	updateAccount(t, bar)
	hash2 := saveState(t)

	require.NotEqual(t, hash1, hash2)

	foo2 := getAccount(t, foo.Address())
	bar2 := getAccount(t, bar.Address())

	require.Equal(t, uint64(1), foo2.Balance())
	require.Equal(t, uint64(1), bar2.Balance())

	st2 := loadState(t, hash2)

	foo3, _ := st2.GetAccount(foo.Address())
	bar3, _ := st2.GetAccount(bar.Address())

	require.Equal(t, uint64(1), foo3.Balance())
	require.Equal(t, uint64(1), bar3.Balance())
}

func TestState_UpdateAccount(t *testing.T) {
	foo1 := account.NewAccountFromSecret("Foo")

	foo1.AddToBalance(1)
	foo1.SetCode([]byte{0x60})
	updateAccount(t, foo1)

	foo2 := getAccount(t, foo1.Address())
	assert.Equal(t, foo1, foo2)
}
