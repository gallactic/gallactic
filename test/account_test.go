package test

import (
	"math/rand"
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
)

func setupAccountPool(m *testing.M) {
	names := []string{"alice", "bob", "carol", "dan", "eve", "satoshi", "vbuterin", "finterran", "b00f", "pouladzade", "benjaminbollen", "silasdavis", "ebuchman", "zramsay", "seanyoung", "VoR0220",
		"smblucker", "shuangjj", "compleatang", "prestonjbyrne", "ietv", "bryant1410", "jaekwon", "ratranqu", "dennismckinnon"}

	accountPool = make(map[string]*account.Account)
	//signerPool = make(map[string]crypto.Signer)

	for i, name := range names {
		bal := rand.New(rand.NewSource(int64(i))).Uint64()
		signer := crypto.PrivateKeyFromSecret(name)
		acc, _ := account.NewAccount(signer.PublicKey().AccountAddress())

		acc.AddToBalance(bal)

		accountPool[name] = acc
		//signerPool[name] = signer
	}
}

/*
func makeAccount(t *testing.T, bal uint64, perm account.Permissions) (*account.Account, crypto.Address) {
	acc := account.NewAccount(generateNewAddress(t))
	acc.SetPermissions(perm)
	acc.AddToBalance(bal)
	updateAccount(t, account)
	commit(t)

	return account, acc.Address()
}

func makeContractAccount(t *testing.T, code []byte, bal uint64, perm permission.Permissions) (*account.Account, crypto.Address) {
	deriveFrom := getAccount(t, "b00f")
	contractAcc := evm.DeriveNewAccount(deriveFrom)
	contractAcc.SetCode(code)
	contractAcc.SetPermissions(perm)
	contractAcc.AddToBalance(bal)
	updateAccount(t, contractAcc)
	updateAccount(t, deriveFrom)
	commit(t)

	return contractAcc, contractAcc.Address()
}
*/
