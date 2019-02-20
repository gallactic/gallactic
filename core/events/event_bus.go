package events

import (
	"context"

	tmCommon "github.com/tendermint/tendermint/libs/common"
	//tmLog "github.com/tendermint/tendermint/libs/log"
	tmLogger "github.com/gallactic/gallactic/core/consensus/tendermint/logger"
	tmPubSub "github.com/tendermint/tendermint/libs/pubsub"
)

const defaultCapacity = 0

type EventBus interface {
	Start() error
	Stop() error
	Subscribe(ctx context.Context, subscriber string, query tmPubSub.Query, out chan<- interface{}) error
	Unsubscribe(ctx context.Context, subscriber string, query tmPubSub.Query) error
	UnsubscribeAll(ctx context.Context, subscriber string) error
	Publish(msg interface{}, tags tmPubSub.TagMap) error
}

// EventBus is a common bus for all events going through the system. All calls
// are proxied to underlying pubsub server. All events must be published using
// EventBus to ensure correct data types.
type eventBus struct {
	tmCommon.BaseService
	pubsub *tmPubSub.Server
}

// NewEventBus returns a new event bus.
func NewEventBus() EventBus {
	return NewEventBusWithBufferCapacity(defaultCapacity)
}

// NewEventBusWithBufferCapacity returns a new event bus with the given buffer capacity.
func NewEventBusWithBufferCapacity(cap int) EventBus {
	// capacity could be exposed later if needed
	pubsub := tmPubSub.NewServer(tmPubSub.BufferCapacity(cap))
	b := &eventBus{pubsub: pubsub}
	b.BaseService = *tmCommon.NewBaseService(nil, "EventBus", b)
	l := tmLogger.NewLogger()
	b.pubsub.SetLogger(l)
	return b
}

func (b *eventBus) Start() error {
	return b.pubsub.Start()
}

func (b *eventBus) Stop() error {
	return b.pubsub.Stop()
}

func (b *eventBus) Subscribe(ctx context.Context, subscriber string, query tmPubSub.Query, out chan<- interface{}) error {
	return b.pubsub.Subscribe(ctx, subscriber, query, out)
}

func (b *eventBus) Unsubscribe(ctx context.Context, subscriber string, query tmPubSub.Query) error {
	return b.pubsub.Unsubscribe(ctx, subscriber, query)
}

func (b *eventBus) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return b.pubsub.UnsubscribeAll(ctx, subscriber)
}

func (b *eventBus) Publish(msg interface{}, tags tmPubSub.TagMap) error {
	// no explicit deadline for publishing events
	ctx := context.Background()
	b.pubsub.PublishWithTags(ctx, msg, tags)
	return nil
}
