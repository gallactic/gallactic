package rpc

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type Codec interface {
	EncodeBytes(interface{}) ([]byte, error)
	Encode(interface{}, io.Writer) error
	DecodeBytes(interface{}, []byte) error
	Decode(interface{}, io.Reader) error
}

// Codec that uses tendermints 'binary' package for JSON.
type TCodec struct {
}

// Get a new codec.
func NewTCodec() Codec {
	return &TCodec{}
}

// Encode to an io.Writer.
func (codec *TCodec) Encode(v interface{}, w io.Writer) error {
	bs, err := codec.EncodeBytes(v)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}

// Encode to a byte array.
func (codec *TCodec) EncodeBytes(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode from an io.Reader.
func (codec *TCodec) Decode(v interface{}, r io.Reader) error {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return codec.DecodeBytes(v, bs)
}

// Decode from a byte array.
func (codec *TCodec) DecodeBytes(v interface{}, bs []byte) error {
	return json.Unmarshal(bs, v)
}
