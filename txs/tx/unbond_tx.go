package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
)

type UnbondTx struct {
	data unbondData
}

type unbondData struct {
	From TxInput  `json:"from"` // Validator
	To   TxOutput `json:"to"`
}

func NewUnbondTx(from, to crypto.Address, amount, sequence, fee uint64) *UnbondTx {
	return &UnbondTx{
		data: unbondData{
			From: TxInput{
				Address:  from,
				Sequence: sequence,
				Amount:   amount + fee,
			},
			To: TxOutput{
				Address: to,
				Amount:  amount,
			},
		},
	}
}

func (tx *UnbondTx) Type() Type           { return TypeUnbond }
func (tx *UnbondTx) From() crypto.Address { return tx.data.From.Address }
func (tx *UnbondTx) To() crypto.Address   { return tx.data.To.Address }
func (tx *UnbondTx) Amount() uint64       { return tx.data.To.Amount }
func (tx *UnbondTx) Fee() uint64          { return tx.data.From.Amount - tx.data.To.Amount }

func (tx *UnbondTx) Inputs() []TxInput {
	return []TxInput{tx.data.From}
}

func (tx *UnbondTx) Outputs() []TxOutput {
	return []TxOutput{tx.data.To}
}

/// ----------
/// MARSHALING

func (tx *UnbondTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinary(tx.data)
}

func (tx *UnbondTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinary(bs, &tx.data)
}

func (tx UnbondTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *UnbondTx) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &tx.data)
	if err != nil {
		return err
	}
	return nil
}
