package outbox

import (
	"context"
	"errors"
	"github.com/VanceMichael/harborflow/internal/domain"
)

type DeliveryClient interface {
	Deliver(context.Context, string, string, string) error
}
type Sender struct{ Client DeliveryClient }

func (s Sender) Send(ctx context.Context, m domain.OutboxMessage) error {
	if m.IdempotencyKey == "" {
		return domain.ErrInvalid
	}
	if err := s.Client.Deliver(ctx, m.EventKey, m.Payload, m.IdempotencyKey); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return domain.ErrUnavailable
	}
	return nil
}
