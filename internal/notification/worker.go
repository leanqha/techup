package notification

import (
	"context"
	"encoding/json"
)

// Worker is an application-level consumer use-case.
type Worker struct {
	consumer Consumer
	router   Router
}

func NewWorker(consumer Consumer, router Router) *Worker {
	return &Worker{consumer: consumer, router: router}
}

// Run starts queue consumption and routes each decoded event envelope.
func (w *Worker) Run(ctx context.Context, queue string) error {
	return w.consumer.Consume(ctx, queue, func(ctx context.Context, message RawMessage) error {
		var event Envelope
		if err := json.Unmarshal(message.Body, &event); err != nil {
			return err
		}
		return w.router.Route(ctx, event)
	})
}

func (w *Worker) Close() error {
	if w == nil || w.consumer == nil {
		return nil
	}
	return w.consumer.Close()
}
