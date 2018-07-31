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
	tBC, _ = blockchain.LoadOrNewBlockchain(tDB, tGenesis, tLogger)
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

	env := txs.Enclose(tChainID, tx)
	require.NoError(t, env.Sign(signers...), "Could not sign tx in call: %s", debug.Stack())

	if errorCode != e.ErrNone {
		require.Equal(t, e.Code(tChecker.Execute(env)), errorCode, "Tx should fail: %s", debug.Stack())
		require.Equal(t, e.Code(tCommitter.Execute(env)), errorCode, "Tx should fail: %s", debug.Stack())
	} else {
		require.NoError(t, tChecker.Execute(env), "Tx should not fail: %s", debug.Stack())
		require.NoError(t, tCommitter.Execute(env), "Tx should not fail: %s", debug.Stack())
		commit(t)
	}

	return env
}
