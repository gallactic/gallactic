package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
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
func (tx *BondTx) From() crypto.Address        { return tx.data.From.Address }
func (tx *BondTx) To() crypto.Address          { return tx.data.To.Address }
func (tx *BondTx) PublicKey() crypto.PublicKey { return tx.data.PublicKey }
func (tx *BondTx) Amount() uint64              { return tx.data.To.Amount }
func (tx *BondTx) Fee() uint64                 { return tx.data.From.Amount - tx.data.To.Amount }

func (tx *BondTx) Signers() []TxInput {
	return []TxInput{tx.data.From}
}

func (tx *BondTx) Outputs() []TxOutput {
	return []TxOutput{tx.data.To}
}

/// ----------
/// MARSHALING

func (tx BondTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinary(tx.data)
}

func (tx *BondTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinary(bs, &tx.data)
}

func (tx BondTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *BondTx) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &tx.data)
	if err != nil {
		return err
	}
	return nil
}
