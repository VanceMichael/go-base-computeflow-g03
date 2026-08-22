package outbox_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/outbox"
	"github.com/VanceMichael/harborflow/internal/testsupport"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestOutboxServiceClaimsAndDeliversOwnedMessage(t *testing.T) {
	f := testsupport.New(t)
	s := outbox.New(f.Store)
	m := domain.OutboxMessage{ID: uuid.NewString(), PortID: f.Port.ID, EventKey: "event", State: string(domain.OutboxPending), IdempotencyKey: "k", CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertOutbox(ctx, tx, m, "{}"); err != nil {
			f.T.Fatal(err)
		}
	})
	if err := s.Claim(context.Background(), m.ID, "worker", f.Now); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkDelivered(context.Background(), m.ID, "other", f.Now); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
	if err := s.MarkDelivered(context.Background(), m.ID, "worker", f.Now); err != nil {
		t.Fatal(err)
	}
}
func TestOutboxServiceRequeuesExpiredMessage(t *testing.T) {
	f := testsupport.New(t)
	s := outbox.New(f.Store)
	m := domain.OutboxMessage{ID: uuid.NewString(), PortID: f.Port.ID, EventKey: "event", State: string(domain.OutboxPending), IdempotencyKey: "k", CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertOutbox(ctx, tx, m, "{}"); err != nil {
			f.T.Fatal(err)
		}
	})
	if err := s.Claim(context.Background(), m.ID, "worker", f.Now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	n, err := s.RequeueExpired(context.Background(), f.Now)
	if err != nil || n != 1 {
		t.Fatalf("%d %v", n, err)
	}
}
