package txs

import (
	"encoding/json"
)

type jsonCodec struct{}

func NewJSONCodec() *jsonCodec {
	return &jsonCodec{}
}

func (gwc *jsonCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return json.Marshal(env)
}

func (gwc *jsonCodec) DecodeTx(bs []byte) (*Envelope, error) {
	env := new(Envelope)
	err := json.Unmarshal(bs, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}
