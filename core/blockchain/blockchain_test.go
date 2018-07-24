package blockchain

import (
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestPersistedState(t *testing.T) {
	val := []*validator.Validator{
		validator.NewValidator(crypto.GeneratePrivateKey(nil).PublicKey(), 1000, 0)}

	/// To strip monotonics from time use time.Truncate(0)
	gen := genesis.MakeGenesisDoc("bar", time.Now().Truncate(0), permission.ZeroPermissions, nil, val)
	db := dbm.NewMemDB()
	bc1, err := newBlockchain(db, gen, logging.NewNoopLogger())
	require.NoError(t, err)

	hash1, err := bc1.CommitBlock(time.Now().Truncate(0), []byte{1, 2})
	require.NoError(t, err)

	// The hash should not change
	hash2, err := bc1.CommitBlock(time.Now().Truncate(0), []byte{3, 4})
	require.NoError(t, err)

	// update state, the hash should change
	addr, _ := crypto.AddressFromString("ac9E2cyNA5UfB8pUpqzEz4QCcBpp8sxnEaN")
	acc, _ := account.NewAccount(addr)
	assert.NoError(t, bc1.state.UpdateAccount(acc))
	hash3, err := bc1.CommitBlock(time.Now().Truncate(0), []byte{5, 6})
	require.NoError(t, err)

	require.Equal(t, hash1, hash2)
	require.NotEqual(t, hash2, hash3)
	bc1.save() /// save last state

	/// load blockchain
	bc2, err2 := loadBlockchain(db, logging.NewNoopLogger())
	require.NoError(t, err2)

	assert.Equal(t, bc1.data, bc2.data)
}
