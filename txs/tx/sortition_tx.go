package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
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

func NewSortitionTx(validator crypto.Address, height, index, sequence, fee uint64, proof []byte) *SortitionTx {
	return &SortitionTx{
		data: sortitionData{
			Validator: TxInput{
				Address:  validator,
				Sequence: sequence,
				Amount:   fee,
			},
			Height: height,
			Index:  index,
			Proof:  proof,
		},
	}
}

func (tx *SortitionTx) Type() Type { return TypeSortition }

func (tx *SortitionTx) Inputs() []TxInput {
	return []TxInput{tx.data.Validator}
}

func (tx *SortitionTx) Outputs() []TxOutput {
	return nil
}

/// ----------
/// MARSHALING

func (tx *SortitionTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinary(tx.data)
}

func (tx *SortitionTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinary(bs, &tx.data)
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
