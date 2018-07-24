package tests

import (
	"testing"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSendTx(t *testing.T, from, to string, amount, fee uint64) *tx.SendTx {
	tx, err := tx.EmptySendTx()
	require.NoError(t, err)

	addSender(t, tx, from, amount, fee)
	addReceiver(t, tx, to, amount)
	return tx
}

func addSender(t *testing.T, tx *tx.SendTx, from string, amount, fee uint64) *tx.SendTx {
	acc := getAccountByName(t, from)
	tx.AddSender(acc.Address(), acc.Sequence()+1, amount+fee)
	return tx
}

func addReceiver(t *testing.T, tx *tx.SendTx, to string, amt uint64) *tx.SendTx {
	var toAddress crypto.Address
	if to != "" {
		toAddress = tAccounts[to].Address()
	} else {
		toAddress = generateNewAccountAddress(t)
	}

	tx.AddReceiver(toAddress, amt)
	return tx
}

func makeCallTx(t *testing.T, from string, addr *crypto.Address, data []byte, amt, fee uint64) *tx.CallTx {
	acc := getAccountByName(t, from)
	tx, err := tx.NewCallTx(acc.Address(), addr, acc.Sequence()+1, data, 210000, amt, fee)
	assert.NoError(t, err)

	return tx
}
