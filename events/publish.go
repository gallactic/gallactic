package events

import (
	"context"
	"fmt"

	tmPubSub "github.com/tendermint/tendermint/libs/pubsub"
)

// PublishEventTx publishes tx event with tags from Result. Note it will add
// predefined tags (EventTypeKey, TxHashKey). Existing tags with the same names
// will be overwritten.
func (b *EventBus) PublishEventTx(data EventDataTx) error {
	// no explicit deadline for publishing events
	ctx := context.Background()

	tags := b.validateAndStringifyTags(data.Tags, b.Logger.With("tx", data.Tx))

	// add predefined tags
	logIfTagExists(EventTypeKey, tags, b.Logger)
	tags[EventTypeKey] = EventTx

	logIfTagExists(TxHashKey, tags, b.Logger)
	tags[TxHashKey] = fmt.Sprintf("%X", data.Tx.Hash())

	//// TODO:::
	////logIfTagExists(TxHeightKey, tags, b.Logger)
	////tags[TxHeightKey] = fmt.Sprintf("%d", data.Height)

	b.pubsub.PublishWithTags(ctx, data, tmPubSub.NewTagMap(tags))
	return nil
}
