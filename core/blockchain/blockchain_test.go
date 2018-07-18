package blockchain

import (
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestPersistedState(t *testing.T) {
	/// To strip monotonics from time use time.Truncate(0)
	gen := genesis.MakeGenesisDoc("bar", time.Now().Truncate(0), permission.ZeroPermissions, nil, nil)
	db := dbm.NewMemDB()
	bc1, err1 := newBlockchain(db, gen, logging.NewNoopLogger())
	require.NoError(t, err1)

	bc1.CommitBlock(time.Now().Truncate(0), []byte{0x1, 0x2})
	bc1.CommitBlock(time.Now().Truncate(0), []byte{0x3, 0x4})

	bc1.save() /// save last state

	/// load blockchain
	bc2, err2 := loadBlockchain(db, logging.NewNoopLogger())
	require.NoError(t, err2)

	assert.Equal(t, bc1, bc2)
}
