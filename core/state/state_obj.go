package state

import "github.com/gallactic/gallactic/crypto"

type StateObj interface {
	Address() crypto.Address
	Balance() uint64
	SubtractFromBalance(amount uint64) error
	AddToBalance(amount uint64) error
	Sequence() uint64
	IncSequence()
}
