package notification

import "context"

// Service is an application-level producer use-case.
type Service struct {
	publisher Publisher
}

func NewService(publisher Publisher) *Service {
	return &Service{publisher: publisher}
}

// Publish forwards a prepared event envelope to the selected routing key.
// Encoding strategy is intentionally delegated to adapter/wiring level.
func (s *Service) Publish(ctx context.Context, routingKey string, envelopeBytes []byte) error {
	return s.publisher.Publish(ctx, routingKey, envelopeBytes)
}

func (s *Service) Close() error {
	if s == nil || s.publisher == nil {
		return nil
	}
	return s.publisher.Close()
}
