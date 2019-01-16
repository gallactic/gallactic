package events

import (
	"crypto/rand"
	"fmt"

	"github.com/gallactic/gallactic/txs"
	amino "github.com/tendermint/go-amino"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	hex "github.com/tmthrgd/go-hex"
)

// Reserved event types (alphabetically sorted).
const (
	EventTx = "Tx"
)

///////////////////////////////////////////////////////////////////////////////
// ENCODING / DECODING
///////////////////////////////////////////////////////////////////////////////

// TMEventData implements events.EventData.
type EventData interface {
	// empty interface
}

func RegisterEventDatas(cdc *amino.Codec) {
	cdc.RegisterConcrete(EventDataTx{}, "tendermint/event/tx", nil)
	//cdc.RegisterInterface((*EventDataTx)(nil), nil)
}

///////////////////////////////////////////////////////////////////////////////
// PUBSUB
///////////////////////////////////////////////////////////////////////////////

const (
	// EventTypeKey is a reserved key, used to specify event type in tags.
	EventTypeKey = "tm.event"
	// TxHashKey is a reserved key, used to specify transaction's hash.
	// see EventBus#PublishEventTx
	TxHashKey = "tx.hash"
	// TxHeightKey is a reserved key, used to specify transaction block's height.
	// see EventBus#PublishEventTx
	TxHeightKey = "tx.height"
)

var (
	EventQueryTx = QueryForEvent(EventTx)
)

func EventQueryTxFor(tx txs.Envelope) tmpubsub.Query {
	return tmquery.MustParse(fmt.Sprintf("%s='%s' AND %s='%X'", EventTypeKey, EventTx, TxHashKey, tx.Hash()))
}

func QueryForEvent(eventType string) tmpubsub.Query {
	return tmquery.MustParse(fmt.Sprintf("%s='%s'", EventTypeKey, eventType))
}

// BlockEventPublisher publishes all block related events
type BlockEventPublisher interface {
	//PublishEventNewBlock(block EventDataNewBlock) error
	//PublishEventNewBlockHeader(header EventDataNewBlockHeader) error
	PublishEventTx(EventDataTx) error
	//PublishEventValidatorSetUpdates(EventDataValidatorSetUpdates) error
}

type TxEventPublisher interface {
	PublishEventTx(EventDataTx) error
}

func GenSubID() string {
	bs := make([]byte, 32)
	rand.Read(bs)
	return hex.EncodeUpperToString(bs)
}
