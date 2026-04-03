package notification

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageDeduplicator interface {
	Seen(ctx context.Context, messageID string) (bool, error)
	MarkProcessed(ctx context.Context, messageID string) error
}

type MessageHandler func(context.Context, amqp.Delivery) error

type NoopDeduplicator struct{}

func (NoopDeduplicator) Seen(context.Context, string) (bool, error) {
	return false, nil
}

func (NoopDeduplicator) MarkProcessed(context.Context, string) error {
	return nil
}

type InMemoryDeduplicator struct {
	mu   sync.RWMutex
	seen map[string]struct{}
}

func NewInMemoryDeduplicator() *InMemoryDeduplicator {
	return &InMemoryDeduplicator{
		seen: map[string]struct{}{},
	}
}

func (d *InMemoryDeduplicator) Seen(_ context.Context, messageID string) (bool, error) {
	if messageID == "" {
		return false, nil
	}

	d.mu.RLock()
	_, ok := d.seen[messageID]
	d.mu.RUnlock()
	return ok, nil
}

func (d *InMemoryDeduplicator) MarkProcessed(_ context.Context, messageID string) error {
	if messageID == "" {
		return nil
	}

	d.mu.Lock()
	d.seen[messageID] = struct{}{}
	d.mu.Unlock()
	return nil
}

func HandleWithDedup(ctx context.Context, deduplicator MessageDeduplicator, msg amqp.Delivery, handler MessageHandler) error {
	if deduplicator == nil {
		deduplicator = NoopDeduplicator{}
	}

	if handler == nil {
		return nil
	}

	if msg.MessageId != "" {
		seen, err := deduplicator.Seen(ctx, msg.MessageId)
		if err != nil {
			return err
		}
		if seen {
			return nil
		}
	}

	if err := handler(ctx, msg); err != nil {
		return err
	}

	if msg.MessageId == "" {
		return nil
	}

	return deduplicator.MarkProcessed(ctx, msg.MessageId)
}
