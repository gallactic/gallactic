package tests

import (
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func setupGenesis(m *testing.M) {
	accounts := make([]*account.Account, len(tAccounts))
	validators := make([]*validator.Validator, len(tValidators))

	i := 0
	for _, acc := range tAccounts {
		accounts[i] = acc
		i++
	}

	i = 0
	for _, val := range tValidators {
		validators[i] = val
		i++
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
