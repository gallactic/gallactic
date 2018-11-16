package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type SendTx struct {
	data sendData
}
type sendData struct {
	Senders   []TxInput  `json:"senders"`
	Receivers []TxOutput `json:"receivers"`
}

func EmptySendTx() (*SendTx, error) {
	return &SendTx{}, nil
}

func NewSendTx(from, to crypto.Address, seq, amt, fee uint64) (*SendTx, error) {
	tx := &SendTx{}
	tx.AddSender(from, seq, amt+fee)
	tx.AddReceiver(to, amt)

	return tx, nil
}

func (tx *SendTx) Type() Type            { return TypeSend }
func (tx *SendTx) Senders() []TxInput    { return tx.data.Senders }
func (tx *SendTx) Receivers() []TxOutput { return tx.data.Receivers }

func (tx *SendTx) Signers() []TxInput {
	return tx.data.Senders
}

func (tx *SendTx) Amount() uint64 {
	return tx.outAmount()
}

func (tx *SendTx) Fee() uint64 {
	return tx.inAmount() - tx.outAmount()
}

func (tx *SendTx) AddSender(addr crypto.Address, seq, amt uint64) {
	tx.data.Senders = append(tx.data.Senders, TxInput{
		Address:  addr,
		Amount:   amt,
		Sequence: seq,
	})
}

func (tx *SendTx) AddReceiver(addr crypto.Address, amt uint64) {
	tx.data.Receivers = append(tx.data.Receivers, TxOutput{
		Address: addr,
		Amount:  amt,
	})
}

func (tx *SendTx) inAmount() uint64 {
	inAmt := uint64(0)
	for _, in := range tx.data.Senders {
		inAmt += in.Amount
	}
	return inAmt
}

func (tx *SendTx) outAmount() uint64 {
	outAmt := uint64(0)
	for _, out := range tx.data.Receivers {
		outAmt += out.Amount
	}
	return outAmt
}

func (tx *SendTx) EnsureValid() error {
	if tx.outAmount() > tx.inAmount() {
		return e.Error(e.ErrInsufficientFunds)
	}

	for _, in := range tx.data.Senders {
		if err := in.ensureValid(); err != nil {
			return err
		}

		if !in.Address.IsAccountAddress() {
			return e.Error(e.ErrInvalidAddress)
		}
	}

	for _, out := range tx.data.Receivers {
		if err := out.ensureValid(); err != nil {
			return err
		}

		if !out.Address.IsAccountAddress() {
			return e.Error(e.ErrInvalidAddress)
		}
	}

	return nil
}

/// ----------
/// MARSHALING

func (tx SendTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(tx.data)
}

func (tx *SendTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, &tx.data)
}

func (tx SendTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *SendTx) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &tx.data)
	if err != nil {
		return err
	}
	return nil
}
