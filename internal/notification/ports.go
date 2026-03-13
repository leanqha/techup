package notification

import "context"

// Publisher is an output port for event publishing.
type Publisher interface {
	Publish(ctx context.Context, routingKey string, body []byte) error
	Close() error
}

// Consumer is an input port for subscribing to broker messages.
type Consumer interface {
	Consume(ctx context.Context, queue string, handler ConsumeHandler) error
	Close() error
}

// ConsumeHandler processes one raw broker message.
type ConsumeHandler func(ctx context.Context, message RawMessage) error

// Router dispatches already-decoded events to business handlers.
type Router interface {
	Route(ctx context.Context, event Envelope) error
}

// Codec is responsible for message serialization/deserialization.
type Codec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}
