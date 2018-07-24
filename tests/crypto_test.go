package tests

import (
	"testing"

	"github.com/gallactic/gallactic/crypto"
)

func generateNewAccountAddress(t *testing.T) crypto.Address {
	return generateNewPublicKey(t).AccountAddress()
}

func generateNewPublicKey(t *testing.T) crypto.PublicKey {
	privateKey := crypto.GeneratePrivateKey(nil)
	return privateKey.PublicKey()
}
