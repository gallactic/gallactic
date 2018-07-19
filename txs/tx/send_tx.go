package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
)

type SendTx struct {
	data sendData
}
type sendData struct {
	Senders   []TxInput  `json:"sender"`
	Receivers []TxOutput `json:"receiver"`
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

func (tx *SendTx) Type() Type          { return TypeSend }
func (tx *SendTx) Inputs() []TxInput   { return tx.data.Senders }
func (tx *SendTx) Outputs() []TxOutput { return tx.data.Receivers }

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

/// ----------
/// MARSHALING

func (tx SendTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinary(tx.data)
}

func (tx *SendTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinary(bs, &tx.data)
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
