package tests

import (
	"runtime/debug"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCallTx(t *testing.T, from string, addr crypto.Address, data []byte, amount, fee uint64) *tx.CallTx {
	acc := getAccountByName(t, from)
	tx, err := tx.NewCallTx(acc.Address(), addr, acc.Sequence()+1, data, 210000, amount, fee)
	require.Equal(t, amount, tx.Amount())
	require.Equal(t, fee, tx.Fee())
	assert.NoError(t, err)

	return tx
}

func execTxWaitAccountCall(t *testing.T, tx tx.Tx, name string, addr crypto.Address) ( /* *events.EventDataCall*/ error, error) {
	env := txs.Enclose(tChainID, tx)
	/// ch := make(chan *events.EventDataCall)
	/// const subscriber = "exexTxWaitEvent"

	require.NoError(t, env.Sign(tSigners[name]), "Could not sign tx in call: %s", debug.Stack())

	/// events.SubscribeAccountCall(ctx, emitter, subscriber, address, env.Tx.Hash(), -1, ch)
	/// defer emitter.UnsubscribeAll(ctx, subscriber)

	err := tCommitter.Execute(env)
	assert.NoError(t, err)

	commit(t)
	/*
		ticker := time.NewTicker(2 * time.Second)

		select {
		case eventDataCall := <-ch:
			fmt.Println("MSG: ", eventDataCall)
			return eventDataCall, eventDataCall.Exception

		case <-ticker.C:
			return nil, e.Error(e.ErrTimeOut)
		}
	*/
	return err, err
}

func TestTxSequence(t *testing.T) {
	setPermissions(t, "alice", permission.Send)

	sequence1 := getAccountByName(t, "alice").Sequence()
	sequence2 := getAccountByName(t, "bob").Sequence()
	for i := 0; i < 100; i++ {
		tx := makeSendTx(t, "alice", "bob", 1, _fee)
		signAndExecute(t, e.ErrNone, tx, "alice")
	}

	require.Equal(t, sequence1+100, getAccountByName(t, "alice").Sequence())
	require.Equal(t, sequence2, getAccountByName(t, "bob").Sequence())
}
