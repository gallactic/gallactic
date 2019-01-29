package txs

import (
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/evm"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"
)

const (
	Ok     = 0
	Failed = 1
)

// Transaction receipt
type Receipt struct {
	Type            tx.Type         `json:"type"`
	Hash            binary.HexBytes `json:"hash"`
	Status          int             `json:"status"`
	Height          int64           `json:"height,omitempty"`
	GasUsed         uint64          `json:"gasUsed,omitempty"`
	GasWanted       uint64          `json:"gasWanted,omitempty"`
	ContractAddress *crypto.Address `json:"contractAddress,omitempty"`
	Logs            evm.Logs        `json:"logs,omitempty"`
	Output          []byte          `json:"output,omitempty"`
}
