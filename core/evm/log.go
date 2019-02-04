package evm

import (
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/crypto"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

type Log struct {
	Address crypto.Address
	Topics  []binary.HexBytes
	Data    binary.HexBytes
}

type Logs []Log

// MarshalBinary - Marshal Log into []byte
func (logs Logs) MarshalBinary() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(logs)
}

// UnmarshalBinary - Unmarshal []byte into Log
func (logs *Logs) UnmarshalBinary(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, &logs)
}
