package events

import (
	"context"

	tmCommon "github.com/tendermint/tendermint/libs/common"
	tmLog "github.com/tendermint/tendermint/libs/log"
	tmPubSub "github.com/tendermint/tendermint/libs/pubsub"
)

const defaultCapacity = 0

type EventBusSubscriber interface {
	Subscribe(ctx context.Context, subscriber string, query tmPubSub.Query, out chan<- interface{}) error
	Unsubscribe(ctx context.Context, subscriber string, query tmPubSub.Query) error
	UnsubscribeAll(ctx context.Context, subscriber string) error
}

// EventBus is a common bus for all events going through the system. All calls
// are proxied to underlying pubsub server. All events must be published using
// EventBus to ensure correct data types.
type EventBus struct {
	tmCommon.BaseService
	pubsub *tmPubSub.Server
}

// NewEventBus returns a new event bus.
func NewEventBus(logger tmLog.Logger) *EventBus {
	return NewEventBusWithBufferCapacity(defaultCapacity, logger)
}

// NewEventBusWithBufferCapacity returns a new event bus with the given buffer capacity.
func NewEventBusWithBufferCapacity(cap int, logger tmLog.Logger) *EventBus {
	// capacity could be exposed later if needed
	pubsub := tmPubSub.NewServer(tmPubSub.BufferCapacity(cap))
	b := &EventBus{pubsub: pubsub}
	b.BaseService = *tmCommon.NewBaseService(nil, "EventBus", b)
	b.SetLogger(tmLog.NewNopLogger()) /// TODO::
	return b
}

func (b *EventBus) SetLogger(l tmLog.Logger) {
	b.BaseService.SetLogger(l)
	b.pubsub.SetLogger(l.With("module", "pubsub"))
}

func (b *EventBus) OnStart() error {
	return b.pubsub.Start()
}

func (b *EventBus) OnStop() {
	b.pubsub.Stop()
}

func (b *EventBus) Subscribe(ctx context.Context, subscriber string, query tmPubSub.Query, out chan<- interface{}) error {
	return b.pubsub.Subscribe(ctx, subscriber, query, out)
}

func (b *EventBus) Unsubscribe(ctx context.Context, subscriber string, query tmPubSub.Query) error {
	return b.pubsub.Unsubscribe(ctx, subscriber, query)
}

func (b *EventBus) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return b.pubsub.UnsubscribeAll(ctx, subscriber)
}

func (b *EventBus) Publish(eventType string, eventData EventData) error {
	// no explicit deadline for publishing events
	ctx := context.Background()
	b.pubsub.PublishWithTags(ctx, eventData, tmPubSub.NewTagMap(map[string]string{EventTypeKey: eventType}))
	return nil
}

func (b *EventBus) validateAndStringifyTags(tags []tmCommon.KVPair, logger tmLog.Logger) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		// basic validation
		if len(tag.Key) == 0 {
			logger.Debug("Got tag with an empty key (skipping)", "tag", tag)
			continue
		}
		result[string(tag.Key)] = string(tag.Value)
	}
	return result
}

func logIfTagExists(tag string, tags map[string]string, logger tmLog.Logger) {
	if value, ok := tags[tag]; ok {
		logger.Error("Found predefined tag (value will be overwritten)", "tag", tag, "value", value)
	}
}
