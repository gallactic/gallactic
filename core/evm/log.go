package evm

import (
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/crypto"
)

type Log struct {
	Address crypto.Address
	Topics  binary.HexBytes
	Data    binary.HexBytes
}
