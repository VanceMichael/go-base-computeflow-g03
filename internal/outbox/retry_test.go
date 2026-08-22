package outbox_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/outbox"
	"testing"
	"time"
)

func TestRetryDelayCapsAtMaximum(t *testing.T) {
	p := outbox.RetryPolicy{Initial: time.Second, MaxDelay: 4 * time.Second}
	if p.Delay(4) != 4*time.Second {
		t.Fatal(p.Delay(4))
	}
}
func TestRetryServiceRejectsDeliveredMessage(t *testing.T) {
	s := outbox.NewRetryService(nil, outbox.RetryPolicy{MaxAttempts: 3})
	m := domain.OutboxMessage{State: string(domain.OutboxDelivered), Attempts: 1}
	if err := s.ShouldRetry(m); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
func TestSenderPropagatesCancellation(t *testing.T) {
	sender := outbox.Sender{Client: fakeClient{err: context.Canceled}}
	err := sender.Send(context.Background(), domain.OutboxMessage{EventKey: "e", Payload: "p", IdempotencyKey: "k"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

type fakeClient struct{ err error }

func (f fakeClient) Deliver(context.Context, string, string, string) error { return f.err }
