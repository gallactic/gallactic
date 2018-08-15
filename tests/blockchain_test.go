package tests

import (
	"runtime/debug"
	"testing"

	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
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
	tChecker = execution.NewBatchChecker(tBC, tLogger)
	tCommitter = execution.NewBatchCommitter(tBC, tLogger)
	tState = tBC.State()
}

func commit(t *testing.T) {
	err := tCommitter.Commit()

	assert.NoError(t, err)
	// commit and clear caches
	assert.NoError(t, tCommitter.Reset())
	assert.NoError(t, tChecker.Reset())
}

func signAndExecute(t *testing.T, errorCode int, tx tx.Tx, names ...string) *txs.Envelope {
	signers := make([]crypto.Signer, len(names))
	for i, name := range names {
		signers[i] = tSigners[name]
	}

	ins := tx.Signers()
	seq := make([]uint64, len(ins))
	totalBalance1 := uint64(0)
	totalBalance2 := uint64(0)

	for i, in := range ins {
		acc := getAccount(t, in.Address)
		seq[i] = acc.Sequence()
		totalBalance1 += acc.Balance()
	}

	env := txs.Enclose(tChainID, tx)
	require.NoError(t, env.Sign(signers...), "Could not sign tx in call: %s", debug.Stack())

	if errorCode != e.ErrNone {
		require.Equal(t, e.Code(tChecker.Execute(env)), errorCode, "Tx should fail: %s", debug.Stack())
		require.Equal(t, e.Code(tCommitter.Execute(env)), errorCode, "Tx should fail: %s", debug.Stack())

		/// check total balance and sequence
		for i, in := range ins {
			acc := getAccount(t, in.Address)
			if seq[i] != acc.Sequence() {
				assert.Failf(t, "Invalid sequence", "Account: %v. Got: %v, Expected: %v", in.Address.String(), in.Sequence, seq[i]+1)
			}

			totalBalance2 += acc.Balance()
		}

		assert.Equal(t, totalBalance2, totalBalance1, "Unexpected total balance")

	} else {
		require.NoError(t, tChecker.Execute(env), "Tx should not fail: %s", debug.Stack())
		require.NoError(t, tCommitter.Execute(env), "Tx should not fail: %s", debug.Stack())
		commit(t)

		/// check total balance and sequence
		for i, in := range ins {
			acc := getAccount(t, in.Address)
			if seq[i]+1 != acc.Sequence() {
				assert.Failf(t, "Invalid sequence", "Account: %v. Got: %v, Expected: %v", in.Address.String(), in.Sequence, seq[i]+1)
			}

			totalBalance2 += acc.Balance()
		}

		assert.Equal(t, totalBalance2, totalBalance1-tx.Amount()-tx.Fee(), "Unexpected total balance")
	}

	return env
}
