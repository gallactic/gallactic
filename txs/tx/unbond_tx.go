package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type UnbondTx struct {
	data unbondData
}

type unbondData struct {
	From TxInput  `json:"from"` // Validator
	To   TxOutput `json:"to"`
}

func NewUnbondTx(from, to crypto.Address, amount, sequence, fee uint64) (*UnbondTx, error) {
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
	}, nil
}

func (tx *UnbondTx) Type() Type    { return TypeUnbond }
func (tx *UnbondTx) From() TxInput { return tx.data.From }
func (tx *UnbondTx) To() TxOutput  { return tx.data.To }

func (tx *UnbondTx) Signers() []TxInput {
	return []TxInput{tx.data.From}
}

func (tx *UnbondTx) Amount() uint64 {
	return tx.data.To.Amount
}

func (tx *UnbondTx) Fee() uint64 {
	return tx.data.From.Amount - tx.data.To.Amount
}

func (tx *UnbondTx) EnsureValid() error {
	if tx.data.To.Amount > tx.data.From.Amount {
		return e.Error(e.ErrInsufficientFunds)
	}

	if err := tx.data.From.ensureValid(); err != nil {
		return err
	}

	if err := tx.data.To.ensureValid(); err != nil {
		return err
	}

	if !tx.data.From.Address.IsValidatorAddress() {
		return e.Error(e.ErrInvalidAddress)
	}

	if !tx.data.To.Address.IsAccountAddress() {
		return e.Error(e.ErrInvalidAddress)
	}

	return nil
}

/// ----------
/// MARSHALING

func (tx UnbondTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(tx.data)
}

func (tx *UnbondTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, &tx.data)
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
