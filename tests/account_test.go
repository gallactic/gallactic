package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
)

func setupAccountPool(m *testing.M) {
	names := []string{"alice", "bob", "carol", "dan", "eve", "satoshi", "vbuterin", "finterran", "b00f", "pouladzade", "benjaminbollen", "silasdavis", "ebuchman", "zramsay", "seanyoung", "VoR0220",
		"smblucker", "shuangjj", "compleatang", "prestonjbyrne", "ietv", "bryant1410", "jaekwon", "ratranqu", "dennismckinnon"}

	tAccounts = make(map[string]*account.Account)
	tValidators = make(map[string]*validator.Validator)
	tSigners = make(map[string]crypto.Signer)

	for i, name := range names {
		bal := rand.New(rand.NewSource(int64(i))).Uint64()
		pb, pv := crypto.GenerateKeyFromSecret(name)
		acc, _ := account.NewAccount(pb.AccountAddress())
		signer := crypto.NewAccountSigner(pv)
		acc.AddToBalance(bal)

		tAccounts[name] = acc
		tSigners[name] = signer
	}

	for i := 0; i < 80; i++ {
		stake := rand.New(rand.NewSource(int64(i))).Uint64()
		name := fmt.Sprintf("val_%d", i+1)
		pb, pv := crypto.GenerateKeyFromSecret(name)
		val, _ := validator.NewValidator(pb, 0)
		signer := crypto.NewValidatorSigner(pv)
		val.AddToStake(stake)

		tValidators[name] = val
		tSigners[name] = signer
	}
}

func newAccountAddress(t *testing.T) crypto.Address {
	pb, _ := crypto.GenerateKey(nil)
	return pb.AccountAddress()
}

func makeAccount(t *testing.T, bal uint64, perm account.Permissions) (*account.Account, crypto.Address) {
	acc, err := account.NewAccount(newAccountAddress(t))
	require.NoError(t, err)
	acc.SetPermissions(perm)
	acc.AddToBalance(bal)
	updateAccount(t, acc)
	commit(t)

	return acc, acc.Address()
}

func makeContractAccount(t *testing.T, code []byte, bal uint64, perm account.Permissions) (*account.Account, crypto.Address) {
	deriveFrom := getAccountByName(t, "b00f")
	ctrAddr := crypto.DeriveContractAddress(deriveFrom.Address(), deriveFrom.Sequence())
	acc, err := account.NewAccount(ctrAddr)
	require.NoError(t, err)
	acc.SetCode(code)
	acc.SetPermissions(perm)
	acc.AddToBalance(bal)
	updateAccount(t, acc)
	updateAccount(t, deriveFrom)
	commit(t)

	return acc, acc.Address()
}
