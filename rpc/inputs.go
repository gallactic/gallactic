package rpc

import (
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/crypto"
)

type (
	AddressInput struct {
		Address crypto.Address `json:"address"`
	}

	FilterListInput struct {
		Filters []*FilterData `json:"filters"`
	}

	StorageAtInput struct {
		Address crypto.Address  `json:"address"`
		Key     binary.HexBytes `json:"key"`
	}

	BlockInput struct {
		Height uint64 `json:"height"`
	}

	BlocksInput struct {
		MinHeight uint64 `json:"minHeight"`
		MaxHeight uint64 `json:"maxHeight"`
	}

	PeersInput struct {
		Address crypto.Address `json:"address"`
	}
)
