package tx

import (
	"github.com/gallactic/gallactic/crypto"
)

type TxOutput struct {
	Address crypto.Address `json:"address"`
	Amount  uint64         `json:"amount"`
}

func (out *TxOutput) ensureValid() error {
	return out.Address.EnsureValid()
}
