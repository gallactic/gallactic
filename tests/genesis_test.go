package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func setupGenesis(m *testing.M) {
	numValidators := 80
	accounts := make([]*account.Account, len(tAccounts))
	validators := make([]*validator.Validator, numValidators)

	i := 0
	for _, acc := range tAccounts {
		accounts[i] = acc
		i++
	}

	for i := 0; i < numValidators; i++ {
		stake := rand.New(rand.NewSource(int64(i))).Uint64()
		privateKey := crypto.GeneratePrivateKey(nil)
		publicKey := privateKey.PublicKey()

		validator := validator.NewValidator(publicKey, stake, 0)
		validators[i] = validator
	}

	gAcc, _ := account.NewAccount(crypto.GlobalAddress)

	tGenesis = proposal.MakeGenesis("test-chain", time.Now(), gAcc, accounts, nil, validators)
	tChainID = tGenesis.ChainID()
}

func TestGenesisDocFromJSON(t *testing.T) {
	bs, err := tGenesis.MarshalJSON()
	assert.NoError(t, err)

	gen2 := new(proposal.Genesis)
	err = gen2.UnmarshalJSON(bs)
	assert.NoError(t, err)

	bsOut, err := gen2.MarshalJSON()
	assert.NoError(t, err)

	assert.Equal(t, bs, bsOut)
	assert.Equal(t, tGenesis.Hash(), gen2.Hash())
}
