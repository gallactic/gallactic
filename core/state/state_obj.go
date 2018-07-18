package state

import "github.com/gallactic/gallactic/crypto"

type StateObj interface {
	Address() crypto.Address
	Balance() uint64
	SubtractFromBalance(amt uint64) error
	AddToBalance(amt uint64) error
	Sequence() uint64
	IncSequence()
}
