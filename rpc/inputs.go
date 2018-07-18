package rpc

import (
	"github.com/gallactic/gallactic/binary"
	"github.com/gallactic/gallactic/crypto"
)

type (
	// Used to send an address. The address should be hex and properly formatted.

	AddressParam struct {
		Address crypto.Address `json:"address"`
	}

	// Used to send an address
	FilterListParam struct {
		Filters []*FilterData `json:"filters"`
	}

	StorageAtParam struct {
		Address crypto.Address  `json:"address"`
		Key     binary.HexBytes `json:"key"`
	}

	HeightParam struct {
		Height uint64 `json:"height"`
	}

	BlocksParam struct {
		MinHeight uint64 `json:"minHeight"`
		MaxHeight uint64 `json:"maxHeight"`
	}

	PeerParam struct {
		Address crypto.Address `json:"address"`
	}
)
