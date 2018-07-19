package test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func setupGenesis(m *testing.M) {
	numValidators := 80
	accounts := make([]*account.Account, len(accountPool))
	validators := make([]*validator.Validator, numValidators)

	i := 0
	for _, account := range accountPool {
		accounts[i] = account
		i++
	}

	for i := 0; i < numValidators; i++ {
		stake := rand.New(rand.NewSource(int64(i))).Uint64()
		privateKey := crypto.GeneratePrivateKey(nil)
		publicKey := privateKey.PublicKey()

		validator := validator.NewValidator(publicKey, stake, 0)
		validators[i] = validator
	}
	genesisDoc = genesis.MakeGenesisDoc("test-chain", time.Now(), permission.ZeroPermissions, accounts, validators)
	chainID = genesisDoc.ChainID()
}

func TestGenesisDocFromJSON(t *testing.T) {
	bs, err := genesisDoc.MarshalJSON()
	assert.NoError(t, err)

	gen2 := new(genesis.Genesis)
	err = gen2.UnmarshalJSON(bs)
	assert.NoError(t, err)

	bsOut, err := gen2.MarshalJSON()
	assert.NoError(t, err)

	assert.Equal(t, bs, bsOut)
	assert.Equal(t, genesisDoc.Hash(), gen2.Hash())
}
