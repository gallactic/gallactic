package events

import (
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs"
	tmCommon "github.com/tendermint/tendermint/libs/common"
)

/*
type EventDataTx interface {
	// empty interface
}

func RegisterEventTx(cdc *amino.Codec) {
	cdc.RegisterConcrete(EventDataCallTx{}, "gallactic/event/tx/call", nil)
}

type EventDataCallTx struct {
	hash []byte
	addr crypto.Address
}
*/

type EventDataTx struct {
	Addr crypto.Address
	Tx   txs.Envelope
	Tags []tmCommon.KVPair
}
