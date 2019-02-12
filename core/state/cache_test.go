package state

import (
	"testing"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func TestAccountChange(t *testing.T) {
	st := newState()
	cache := NewCache(st)
	pb1, _ := crypto.GenerateKeyFromSecret("secret1")
	pb2, _ := crypto.GenerateKeyFromSecret("secret2")
	pb3, _ := crypto.GenerateKeyFromSecret("secret3")
	addr1 := pb1.AccountAddress()
	addr2 := pb2.AccountAddress()
	addr3 := pb3.AccountAddress()

	acc, err := cache.GetAccount(addr1)
	assert.Error(t, err) // return an error
	assert.Nil(t, acc, nil)

	// update cache
	acc1, err := account.NewAccount(addr1)
	acc1.AddToBalance(10)
	cache.UpdateAccount(acc1)
	acc11, err := cache.GetAccount(addr1)
	assert.NoError(t, err)
	assert.Equal(t, acc1, acc11)

	// update state
	acc2, err := account.NewAccount(addr2)
	st.updateAccount(acc2)
	acc22, err := cache.GetAccount(addr2)
	assert.NoError(t, err)
	assert.Equal(t, acc2, acc22)

	/// update storages
	val, err := cache.GetStorage(addr1, binary.Uint64ToWord256(1))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(0))
	cache.SetStorage(addr1, binary.Uint64ToWord256(1), binary.Uint64ToWord256(2))
	val, err = cache.GetStorage(addr1, binary.Uint64ToWord256(2))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(0)) // wrong storage key
	val, err = cache.GetStorage(addr1, binary.Uint64ToWord256(1))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(2))

	/// Update storage then account
	acc3, err := account.NewAccount(addr3)
	st.updateAccount(acc3)
	cache.SetStorage(addr3, binary.Uint64ToWord256(1), binary.Uint64ToWord256(2))
	acc3.AddToBalance(10)
	cache.UpdateAccount(acc3)
	acc33, err := cache.GetAccount(addr3)
	assert.NoError(t, err)
	assert.Equal(t, acc3, acc33)
	val, err = cache.GetStorage(addr3, binary.Uint64ToWord256(1))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(2))

	/// accounts should be untouched while changing storages
	acc11, err = cache.GetAccount(addr1)
	assert.NoError(t, err)
	assert.Equal(t, acc1, acc11)

	acc22, err = cache.GetAccount(addr2)
	assert.NoError(t, err)
	assert.Equal(t, acc2, acc22)

	cache.Reset()
	assert.Equal(t, cache.accChanges.Len(), 0)
	assert.Equal(t, cache.valChanges.Len(), 0)
}
