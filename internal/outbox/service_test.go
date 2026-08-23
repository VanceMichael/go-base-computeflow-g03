package outbox_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/outbox"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"github.com/google/uuid"
	"sync"
	"testing"
	"time"
)

func TestDeliveryLedgerIsSafeForConcurrentReplay(t *testing.T) {
	ledger := outbox.NewLedger()
	var wg sync.WaitGroup
	for n := 0; n < 16; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ledger.Record("delivery", "accepted"); err != nil {
				t.Errorf("record: %v", err)
			}
			if got, ok := ledger.Replay("delivery"); !ok || got != "accepted" {
				t.Errorf("replay=%q ok=%v", got, ok)
			}
		}()
	}
	wg.Wait()
}

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
