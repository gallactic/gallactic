package tx

import (
	"fmt"

	"github.com/gallactic/gallactic/crypto"
)

type TxOutput struct {
	Address crypto.Address `json:"address"`
	Amount  uint64         `json:"amount"`
}

func (txOut *TxOutput) ValidateBasic() error {
	return nil
}

func (txOut *TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%s, Amount:%v}", txOut.Address, txOut.Amount)
}
