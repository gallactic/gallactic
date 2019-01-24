package evm

import (
	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/crypto"
)

type Log struct {
	Address crypto.Address
	Topics  []common.Hash
	Data    []byte
}
