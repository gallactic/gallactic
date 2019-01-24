package txs

import (
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"
)

// Transaction receipt
type Receipt struct {
	Type            tx.Type         `json:"type"`
	Hash            binary.HexBytes `json:"hash"`
	Status          string          `json:"status,omitempty"`
	Failed          bool            `json:"failed,omitempty"`
	Height          int64           `json:"height,omitempty"`
	UsedGas         uint64          `json:"usedGas,omitempty"`
	ContractAddress *crypto.Address `json:"contractAddress,omitempty"`
	Output          []byte          `json:"output,omitempty"`
}
