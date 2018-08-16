package evm

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
)

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func DeriveNewAccount(creator *account.Account) (*account.Account, error) {
	// Generate an address
	seq := creator.Sequence()
	creator.IncSequence()

	addr := crypto.DeriveContractAddress(creator.Address(), seq)

	// Create account from address.
	acc, err := account.NewAccount(addr)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
