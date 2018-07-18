package tx

import (
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type TxInput struct {
	Address  crypto.Address `json:"address"`
	Amount   uint64         `json:"amount"`
	Sequence uint64         `json:"sequence"`
}

func (txIn *TxInput) ValidateBasic() error {
	if txIn.Amount == 0 {
		return e.Error(e.ErrInvalidAmount)
	}
	return nil
}
