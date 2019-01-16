package events

import (
	"context"

	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
)

type NopEventBus struct{}

func (NopEventBus) Subscribe(ctx context.Context, subscriber string, query tmpubsub.Query, out chan<- interface{}) error {
	return nil
}

func (NopEventBus) Unsubscribe(ctx context.Context, subscriber string, query tmpubsub.Query) error {
	return nil
}

func (NopEventBus) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return nil
}

func (NopEventBus) PublishEventTx(data EventDataTx) error {
	return nil
}
