package notification

import "time"

// MessageType identifies a cross-service event name (for example: account.password_reset.requested).
type MessageType string

// Envelope is the transport-agnostic message contract shared by producers and workers.
type Envelope struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Version   int                    `json:"version"`
	CreatedAt time.Time              `json:"created_at"`
	Payload   []byte                 `json:"payload"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// RawMessage is the broker-level payload received by consumers before decoding.
type RawMessage struct {
	RoutingKey  string
	ContentType string
	Body        []byte
	Headers     map[string]interface{}
}
