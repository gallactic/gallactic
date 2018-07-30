package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type CallTx struct {
	data callData
}

type callData struct {
	Caller   TxInput         `json:"caller"`
	Callee   TxOutput        `json:"callee"`
	GasLimit uint64          `json:"gasLimit"`
	Data     binary.HexBytes `json:"data,omitempty"`
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
				Amount:  amount + fee,
			},
			GasLimit: gasLimit,
			Data:     data,
		},
	}, nil
}

func (tx *CallTx) Type() Type       { return TypeCall }
func (tx *CallTx) Caller() TxInput  { return tx.data.Caller }
func (tx *CallTx) Callee() TxOutput { return tx.data.Callee }
func (tx *CallTx) GasLimit() uint64 { return tx.data.GasLimit }
func (tx *CallTx) Data() []byte     { return tx.data.Data }

func (tx *CallTx) Signers() []TxInput {
	return []TxInput{tx.data.Caller}
}

func (tx *CallTx) Amount() uint64 {
	return tx.data.Callee.Amount
}

func (tx *CallTx) Fee() uint64 {
	return tx.data.Caller.Amount - tx.data.Callee.Amount
}

func (tx *CallTx) EnsureValid() error {
	if tx.data.Callee.Amount > tx.data.Caller.Amount {
		return e.Error(e.ErrInsufficientFunds)
	}

	if err := tx.data.Caller.ensureValid(); err != nil {
		return err
	}

	if !tx.data.Caller.Address.IsAccountAddress() {
		return e.Error(e.ErrInvalidAddress)
	}

	if !tx.CreateContract() {
		if err := tx.data.Callee.ensureValid(); err != nil {
			return err
		}

		if !tx.data.Callee.Address.IsContractAddress() {
			return e.Error(e.ErrInvalidAddress)
		}
	}

	return nil
}

func (tx *CallTx) CreateContract() bool {
	return tx.data.Callee.Address == crypto.Address{}
}

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
