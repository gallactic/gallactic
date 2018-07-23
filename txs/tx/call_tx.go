package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
)

type CallTx struct {
	data callData
}

type callData struct {
	Caller   TxInput  `json:"caller"`
	Callee   TxOutput `json:"callee,omitempty"`
	GasLimit uint64   `json:"gas_limit"`
	Data     []byte   `json:"data,omitempty"`
}

func NewCallTx(caller, callee crypto.Address, sequence uint64, data []byte, gasLimit, amount, fee uint64) (*CallTx, error) {
	return &CallTx{
		data: callData{
			Caller: TxInput{
				Address:  caller,
				Sequence: sequence,
				Amount:   amount + fee,
			},
			Callee: TxOutput{
				Address: callee,
				Amount:  amount,
			},
			GasLimit: gasLimit,
			Data:     data,
		},
	}, nil
}

func (tx *CallTx) Type() Type           { return TypeCall }
func (tx *CallTx) Signers() []TxInput   { return []TxInput{tx.data.Caller} }
func (tx *CallTx) Caller() TxInput      { return tx.data.Caller }
func (tx *CallTx) Callee() TxOutput     { return tx.data.Callee }
func (tx *CallTx) Amount() uint64       { return tx.data.Callee.Amount }
func (tx *CallTx) Sequence() uint64     { return tx.data.Caller.Sequence }
func (tx *CallTx) GasLimit() uint64     { return tx.data.GasLimit }
func (tx *CallTx) Fee() uint64          { return tx.data.Caller.Amount - tx.data.Callee.Amount }
func (tx *CallTx) Data() []byte         { return tx.data.Data }
func (tx *CallTx) CreateContract() bool { return true }

/// ----------
/// MARSHALING

func (tx CallTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinary(tx.data)
}

func (tx *CallTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinary(bs, &tx.data)
}

func (tx CallTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *CallTx) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &tx.data)
	if err != nil {
		return err
	}
	return nil
}
