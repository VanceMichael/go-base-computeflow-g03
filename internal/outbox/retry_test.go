package outbox_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/outbox"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestRetryFailureUsesConfiguredDeadThreshold(t *testing.T) {
	f := testsupport.New(t)
	m := domain.OutboxMessage{ID: uuid.NewString(), PortID: f.Port.ID, EventKey: "event", State: string(domain.OutboxPending), IdempotencyKey: "retry", CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertOutbox(ctx, tx, m, "{}"); err != nil {
			t.Fatal(err)
		}
	})
	if err := outbox.New(f.Store).Claim(context.Background(), m.ID, "worker", f.Now); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Store.DB.Exec(`UPDATE outbox_messages SET attempts=3 WHERE id=?`, m.ID); err != nil {
		t.Fatal(err)
	}
	if err := outbox.NewRetryService(f.Store, outbox.RetryPolicy{MaxAttempts: 3}).Failure(context.Background(), m.ID, "worker", "failed"); err != nil {
		t.Fatal(err)
	}
	got, err := f.Store.FindOutbox(context.Background(), m.ID, f.Port.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != string(domain.OutboxDead) {
		t.Fatalf("state=%s", got.State)
	}
}

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
