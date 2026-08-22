package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"time"
)

type Service struct{ Store *sqlite.Store }

func New(s *sqlite.Store) *Service { return &Service{Store: s} }
func (s *Service) Enqueue(ctx context.Context, tx *sql.Tx, portID, eventKey, payload, idempotency string, now time.Time) (domain.OutboxMessage, error) {
	m := domain.OutboxMessage{ID: uuid.NewString(), PortID: portID, EventKey: eventKey, Payload: payload, State: string(domain.OutboxPending), IdempotencyKey: idempotency, CreatedAt: now}
	return m, s.Store.InsertOutbox(ctx, tx, m, payload)
}
func (s *Service) Claim(ctx context.Context, id, owner string, now time.Time) error {
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.ClaimOutbox(ctx, tx, id, owner, now.Add(2*time.Minute))
		return e
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: outbox lease", domain.ErrConflict)
	}
	return nil
}
func (s *Service) MarkDelivered(ctx context.Context, id, owner string, now time.Time) error {
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.MarkOutbox(ctx, tx, id, owner, now)
		return e
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: outbox owner", domain.ErrConflict)
	}
	return nil
}
func (s *Service) RequeueExpired(ctx context.Context, now time.Time) (int64, error) {
	var n int64
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		n, e = s.Store.RequeueExpiredOutbox(ctx, tx, now)
		return e
	})
	return n, err
}
