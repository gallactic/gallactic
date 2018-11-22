package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type BondTx struct {
	data bondData
}

type bondData struct {
	From      TxInput          `json:"from"`
	To        TxOutput         `json:"to"`         // Validator
	PublicKey crypto.PublicKey `json:"public_key"` // Validator
}

func NewBondTx(from crypto.Address, to crypto.PublicKey, amount, sequence, fee uint64) (*BondTx, error) {
	return &BondTx{
		data: bondData{
			From: TxInput{
				Address:  from,
				Sequence: sequence,
				Amount:   amount + fee,
			},
			To: TxOutput{
				Address: to.ValidatorAddress(),
				Amount:  amount,
			},
			PublicKey: to,
		},
	}, nil
}

func (tx *BondTx) Type() Type                  { return TypeBond }
func (tx *BondTx) From() TxInput               { return tx.data.From }
func (tx *BondTx) To() TxOutput                { return tx.data.To }
func (tx *BondTx) PublicKey() crypto.PublicKey { return tx.data.PublicKey }

func (tx *BondTx) Signers() []TxInput {
	return []TxInput{tx.data.From}
}

func (tx *BondTx) Amount() uint64 {
	return tx.data.To.Amount
}

func (tx *BondTx) Fee() uint64 {
	return tx.data.From.Amount - tx.data.To.Amount
}

func (tx *BondTx) EnsureValid() error {
	if tx.data.To.Amount > tx.data.From.Amount {
		return e.Error(e.ErrInsufficientFunds)
	}

	if err := tx.data.From.ensureValid(); err != nil {
		return err
	}

	if err := tx.data.To.ensureValid(); err != nil {
		return err
	}

	if !tx.data.To.Address.Verify(tx.data.PublicKey) {
		return e.Error(e.ErrInvalidPublicKey)
	}

	if !tx.data.From.Address.IsAccountAddress() {
		return e.Error(e.ErrInvalidAddress)
	}

	if !tx.data.To.Address.IsValidatorAddress() {
		return e.Error(e.ErrInvalidAddress)
	}

	return nil
}

/// ----------
/// MARSHALING

func (tx BondTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(tx.data)
}

func (tx *BondTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, &tx.data)
}

func (tx BondTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *BondTx) UnmarshalJSON(bs []byte) error {
	return json.Unmarshal(bs, &tx.data)
}
