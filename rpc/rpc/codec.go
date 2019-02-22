package rpcc

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// NewTmCodec will initialize Tendermint Codec
func NewTmCodec() Codec {
	return &TmCodec{}
}

// Encode to an io.Writer
func (codec *TmCodec) Encode(v interface{}, w io.Writer) error {
	bs, err := codec.EncodeBytes(v)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}

// EncodeBytes will encode to byte array
func (codec *TmCodec) EncodeBytes(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode from an io.Reader
func (codec *TmCodec) Decode(v interface{}, r io.Reader) error {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return codec.DecodeBytes(v, bs)
}

// DecodeBytes from byte array
func (codec *TmCodec) DecodeBytes(v interface{}, bs []byte) error {
	return json.Unmarshal(bs, v)
}
