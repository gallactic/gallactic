package tests

import (
	"runtime/debug"
	"testing"

	"github.com/gallactic/gallactic/core/events"

	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var _fee uint64 = 10

func setupBlockchain(m *testing.M) {
	tDB = dbm.NewMemDB()
	tBC, _ = blockchain.LoadOrNewBlockchain(tDB, tGenesis, nil, tLogger)
	tEventBus = events.NewEventBus(tLogger)
	tChecker = execution.NewBatchChecker(tBC, tLogger)
	tCommitter = execution.NewBatchCommitter(tBC, tEventBus, tLogger)
	tState = tBC.State()

	tEventBus.Start()
}

func commit(t *testing.T) {
	err := tCommitter.Commit()

	assert.NoError(t, err)
	// commit and clear caches
	assert.NoError(t, tCommitter.Reset())
	assert.NoError(t, tChecker.Reset())
}

func signAndExecute(t *testing.T, errorCode int, tx tx.Tx, names ...string) (*txs.Envelope, *txs.Receipt) {
	signers := make([]crypto.Signer, len(names))
	for i, name := range names {
		signers[i] = tSigners[name]
	}

	ins := tx.Signers()
	seq := make([]uint64, len(ins))
	totalBalance1 := uint64(0)
	totalBalance2 := uint64(0)

	for i, in := range ins {
		if in.Address.IsAccountAddress() {
			acc := getAccount(t, in.Address)
			seq[i] = acc.Sequence()
			totalBalance1 += acc.Balance()
		} else {
			val := getValidator(t, in.Address)
			seq[i] = val.Sequence()
			totalBalance1 += val.Stake()
		}
	}

	env := txs.Enclose(tChainID, tx)
	require.NoError(t, env.Sign(signers...), "Could not sign tx in call: %s", debug.Stack())
	rec := env.GenerateReceipt()

	if errorCode != e.ErrNone {
		require.Equal(t, e.Code(tChecker.Execute(env, rec)), errorCode, "Tx should fail: %s", debug.Stack())
		require.Equal(t, e.Code(tCommitter.Execute(env, rec)), errorCode, "Tx should fail: %s", debug.Stack())

		/// check total balance and sequence, should not change
		for i, in := range ins {
			if in.Address.IsAccountAddress() {
				acc := getAccount(t, in.Address)
				if seq[i] != acc.Sequence() {
					assert.Failf(t, "Invalid sequence", "Account: %v. Got: %v, Expected: %v", in.Address.String(), in.Sequence, seq[i])
				}
				totalBalance2 += acc.Balance()
			} else {
				val := getValidator(t, in.Address)
				if seq[i] != val.Sequence() {
					assert.Failf(t, "Invalid sequence", "Validator: %v. Got: %v, Expected: %v", in.Address.String(), in.Sequence, seq[i])
				}
				totalBalance2 += val.Stake()
			}
		}

		assert.Equal(t, totalBalance2, totalBalance1, "Unexpected total balance")
	} else {
		require.NoError(t, tChecker.Execute(env, rec), "Tx should not fail: %s", debug.Stack())
		require.NoError(t, tCommitter.Execute(env, rec), "Tx should not fail: %s", debug.Stack())
		commit(t)

		/// check total balance and sequence, should change
		for i, in := range ins {
			if in.Address.IsAccountAddress() {
				acc := getAccount(t, in.Address)
				if seq[i]+1 != acc.Sequence() {
					assert.Failf(t, "Invalid sequence", "Account: %v. Got: %v, Expected: %v", in.Address.String(), acc.Sequence(), seq[i]+1)
				}
				totalBalance2 += acc.Balance()
			} else {
				val := getValidator(t, in.Address)
				if seq[i]+1 != val.Sequence() {
					assert.Failf(t, "Invalid sequence", "Validator: %v. Got: %v, Expected: %v", in.Address.String(), val.Sequence(), seq[i]+1)
				}
				totalBalance2 += val.Stake()
			}
		}

		if rec.Status == txs.Ok {
			assert.Equal(t, totalBalance2, totalBalance1-tx.Amount()-tx.Fee(), "Unexpected total balance")
		} else {
			assert.Equal(t, totalBalance2, totalBalance1-tx.Fee(), "Unexpected total balance")
		}
	}

	require.Equal(t, rec.Hash.Bytes(), env.Hash())
	require.Equal(t, rec.Type, env.Type)

	return env, rec
}
