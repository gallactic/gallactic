package blockchain

import (
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/types/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestPersistedState(t *testing.T) {
	/// To strip monotonics from time use time.Truncate(0)
	gen := genesis.MakeGenesisDoc("bar", time.Now().Truncate(0), permission.ZeroPermissions, nil, nil)
	db := dbm.NewMemDB()
	bc1, err1 := newBlockchain(db, gen)
	require.NoError(t, err1)

	bc1.CommitBlock(time.Now().Truncate(0), []byte{0x1, 0x2}, []byte{0x4, 0x5})
	bc1.CommitBlock(time.Now().Truncate(0), []byte{0x6, 0x7}, []byte{0x8, 0x9})

	bc1.save() /// save last state

	/// load blockchain
	bc2, err2 := loadBlockchain(db)
	require.NoError(t, err2)

	assert.Equal(t, bc1, bc2)
}
