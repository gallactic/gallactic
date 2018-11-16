package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type SortitionTx struct {
	data sortitionData
}

type sortitionData struct {
	Validator TxInput `json:"validator"`
	Height    uint64  `json:"height"`
	Index     uint64  `json:"index"`
	Proof     []byte  `json:"proof"`
}

func NewSortitionTx(validator crypto.Address, height, seq, fee, index uint64, proof []byte) (*SortitionTx, error) {
	return &SortitionTx{
		data: sortitionData{
			Validator: TxInput{
				Address:  validator,
				Sequence: seq,
				Amount:   fee,
			},
			Height: height,
			Index:  index,
			Proof:  proof,
		},
	}, nil
}

func (tx *SortitionTx) Type() Type         { return TypeSortition }
func (tx *SortitionTx) Validator() TxInput { return tx.data.Validator }
func (tx *SortitionTx) Height() uint64     { return tx.data.Height }
func (tx *SortitionTx) Index() uint64      { return tx.data.Index }
func (tx *SortitionTx) Proof() []byte      { return tx.data.Proof }

func (tx *SortitionTx) Signers() []TxInput {
	return []TxInput{tx.data.Validator}
}

func (tx *SortitionTx) Amount() uint64 {
	return 0
}

func (tx *SortitionTx) Fee() uint64 {
	return tx.data.Validator.Amount
}

func (tx *SortitionTx) EnsureValid() error {
	if err := tx.data.Validator.ensureValid(); err != nil {
		return err
	}

	if !tx.data.Validator.Address.IsValidatorAddress() {
		return e.Error(e.ErrInvalidAddress)
	}

	return nil
}

/// ----------
/// MARSHALING

func (tx SortitionTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(tx.data)
}

func (tx *SortitionTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, &tx.data)
}

func (tx SortitionTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *SortitionTx) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &tx.data)
	if err != nil {
		return err
	}
	return nil
}
