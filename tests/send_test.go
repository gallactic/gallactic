package tests

import (
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSendTx(t *testing.T, from, to string, amt, fee uint64) *tx.SendTx {
	tx, err := tx.EmptySendTx()
	require.NoError(t, err)

	addSender(t, tx, from, amt, fee)
	addReceiver(t, tx, to, amt)

	require.Equal(t, amt, tx.Amount())
	require.Equal(t, fee, tx.Fee())

	return tx
}

func addSender(t *testing.T, tx *tx.SendTx, from string, amt, fee uint64) *tx.SendTx {
	acc := getAccountByName(t, from)
	tx.AddSender(acc.Address(), acc.Sequence()+1, amt+fee)
	return tx
}

func addReceiver(t *testing.T, tx *tx.SendTx, to string, amt uint64) *tx.SendTx {
	var toAddress crypto.Address
	if to != "" {
		toAddress = tAccounts[to].Address()
	} else {
		toAddress = newAccountAddress(t)
	}

	tx.AddReceiver(toAddress, amt)
	return tx
}

func getBalance(t *testing.T, name string) uint64 {
	return getBalanceByAddress(t, tAccounts[name].Address())
}

func getBalanceByAddress(t *testing.T, addr crypto.Address) uint64 {
	acc := getAccount(t, addr)
	require.NotNil(t, acc)
	return acc.Balance()
}

func checkBalance(t *testing.T, name string, amt uint64) {
	checkBalanceByAddress(t, tAccounts[name].Address(), amt)
}

func checkBalanceByAddress(t *testing.T, addr crypto.Address, amt uint64) {
	acc := getAccount(t, addr)
	require.NotNil(t, acc)
	assert.Equal(t, acc.Balance(), amt)
}

func TestSendTxFails(t *testing.T) {
	setPermissions(t, "alice", permission.Send)
	setPermissions(t, "bob", permission.Call)
	setPermissions(t, "carol", permission.CreateContract)

	tx1 := makeSendTx(t, "alice", "dan", 100, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	// simple send tx with call perm should fail
	tx2 := makeSendTx(t, "bob", "dan", 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx2, "bob")

	// simple send tx with create perm should fail
	tx3 := makeSendTx(t, "carol", "dan", 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx3, "carol")

	// simple send tx to unknown account without create_account perm should fail
	tx5 := makeSendTx(t, "alice", "", 100, _fee)
	signAndExecute(t, e.ErrPermDenied, tx5, "alice")

	// Output amount can  be zero
	tx6 := makeSendTx(t, "alice", "dan", 0, _fee)
	signAndExecute(t, e.ErrNone, tx6, "alice")
}

func TestSendPermission(t *testing.T) {
	setPermissions(t, "alice", permission.Send)
	setPermissions(t, "bob", 0)

	// A single input, having the permission, should succeed
	tx1 := makeSendTx(t, "alice", "carol", 10, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	tx2 := makeSendTx(t, "alice", "carol", 10, _fee)
	addSender(t, tx2, "bob", 10, _fee)
	addReceiver(t, tx2, "carol", 10)

	// Two inputs, one with permission, one without, should fail
	signAndExecute(t, e.ErrPermDenied, tx2, "alice", "bob")
}
func TestCreateAccountPermission(t *testing.T) {
	setPermissions(t, "alice", permission.Send|permission.CreateAccount)
	setPermissions(t, "bob", permission.Send)

	aliceBalance := getBalance(t, "alice")
	bobBalance := getBalance(t, "bob")
	//----------------------------------------------------------
	// SendTx to unknown account

	// A single input, having the permission, should succeed
	tx1 := makeSendTx(t, "alice", "", 5, _fee)
	signAndExecute(t, e.ErrNone, tx1, "alice")

	// Two inputs, both with send, should succeed
	tx2 := makeSendTx(t, "alice", "eve", 5, _fee)
	addSender(t, tx2, "bob", 5, _fee)
	tx2.Receivers()[0].Amount = 10
	signAndExecute(t, e.ErrNone, tx2, "alice", "bob")

	// Two inputs, both with send, one with create, one without, should fail
	tx3 := makeSendTx(t, "alice", "", 5, _fee)
	addSender(t, tx3, "bob", 5, _fee)
	tx3.Receivers()[0].Amount = 10
	signAndExecute(t, e.ErrPermDenied, tx3, "alice", "bob")

	// Two inputs, both with send, one with create, one without, two outputs (one known, one unknown) should fail
	tx4 := makeSendTx(t, "alice", "eve", 5, _fee)
	addSender(t, tx4, "bob", 5, _fee)
	addReceiver(t, tx4, "", 5)
	signAndExecute(t, e.ErrPermDenied, tx4, "alice", "bob")

	// Two inputs, both with send, both with create, should pass
	setPermissions(t, "bob", permission.Send|permission.CreateAccount)
	tx5 := makeSendTx(t, "alice", "", 5, _fee)
	addSender(t, tx5, "bob", 5, _fee)
	tx5.Receivers()[0].Amount = 10
	signAndExecute(t, e.ErrNone, tx5, "alice", "bob")

	// Two inputs, both with send, both with create, two outputs (one known, one unknown) should pass
	tx6 := makeSendTx(t, "alice", "eve", 5, _fee)
	addSender(t, tx6, "bob", 5, _fee)
	addReceiver(t, tx6, "", 5)
	signAndExecute(t, e.ErrNone, tx6, "alice", "bob")

	checkBalance(t, "alice", aliceBalance-(4*(5+_fee)))
	checkBalance(t, "bob", bobBalance-(3*(5+_fee)))
}

func TestMultiSigs(t *testing.T) {
	tx, _ := tx.EmptySendTx()
	names := make([]string, 0)

	for n, a := range tAccounts {
		acc := getAccount(t, a.Address())
		acc.SetPermissions(permission.Send | permission.CreateAccount)
		updateAccount(t, acc) // update required permissions

		tx.AddSender(a.Address(), acc.Sequence()+1, 1000)
		tx.AddReceiver(newAccountAddress(t), 999) /// send to new address
		names = append(names, n)
	}

	signAndExecute(t, e.ErrNone, tx, names...)
}
