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
	ch := NewCache(st)
	pb1, _ := crypto.GenerateKeyFromSecret("secret1")
	pb2, _ := crypto.GenerateKeyFromSecret("secret2")
	pb3, _ := crypto.GenerateKeyFromSecret("secret3")
	addr1 := pb1.AccountAddress()
	addr2 := pb2.AccountAddress()
	addr3 := pb3.AccountAddress()

	acc, err := ch.GetAccount(addr1)
	assert.Error(t, err) // return an error
	assert.Nil(t, acc, nil)

	// update cache
	acc1, err := account.NewAccount(addr1)
	acc1.AddToBalance(10)
	ch.UpdateAccount(acc1)
	acc11, err := ch.GetAccount(addr1)
	assert.NoError(t, err)
	assert.Equal(t, acc1, acc11)

	// update state
	acc2, err := account.NewAccount(addr2)
	st.updateAccount(acc2)
	acc22, err := ch.GetAccount(addr2)
	assert.NoError(t, err)
	assert.Equal(t, acc2, acc22)

	/// update storages
	val, err := ch.GetStorage(addr1, binary.Uint64ToWord256(1))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(0))
	ch.SetStorage(addr1, binary.Uint64ToWord256(1), binary.Uint64ToWord256(2))
	val, err = ch.GetStorage(addr1, binary.Uint64ToWord256(2))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(0)) // wrong storage key
	val, err = ch.GetStorage(addr1, binary.Uint64ToWord256(1))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(2))

	/// Update storage then account
	acc3, err := account.NewAccount(addr3)
	st.updateAccount(acc3)
	ch.SetStorage(addr3, binary.Uint64ToWord256(1), binary.Uint64ToWord256(2))
	acc3.AddToBalance(10)
	ch.UpdateAccount(acc3)
	acc33, err := ch.GetAccount(addr3)
	assert.NoError(t, err)
	assert.Equal(t, acc3, acc33)
	val, err = ch.GetStorage(addr3, binary.Uint64ToWord256(1))
	assert.NoError(t, err)
	assert.Equal(t, val, binary.Uint64ToWord256(2))

	/// accounts should be untouched while changing storages
	acc11, err = ch.GetAccount(addr1)
	assert.NoError(t, err)
	assert.Equal(t, acc1, acc11)

	acc22, err = ch.GetAccount(addr2)
	assert.NoError(t, err)
	assert.Equal(t, acc2, acc22)
}
