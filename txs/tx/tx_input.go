package tx

import (
	"github.com/gallactic/gallactic/crypto"
)

type TxInput struct {
	Address  crypto.Address `json:"address"`
	Amount   uint64         `json:"amount"`
	Sequence uint64         `json:"sequence"`
}

func (in *TxInput) ensureValid() error {
	return in.Address.EnsureValid()
}
