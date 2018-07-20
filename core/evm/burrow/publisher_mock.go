package burrow

import (
	"context"

	burrowEvent "github.com/hyperledger/burrow/event"
)

type eventPublisher struct{}

func (pf eventPublisher) Publish(ctx context.Context, message interface{}, tags burrowEvent.Tags) error {
	return nil
}
