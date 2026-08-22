package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"time"
)

type RetryService struct {
	Store  *sqlite.Store
	Policy RetryPolicy
}
type RetryPolicy struct {
	MaxAttempts int
	Initial     time.Duration
	MaxDelay    time.Duration
}

func (r RetryPolicy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}
	d := r.Initial
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= r.MaxDelay {
			return r.MaxDelay
		}
	}
	if d > r.MaxDelay {
		return r.MaxDelay
	}
	return d
}
func NewRetryService(s *sqlite.Store, p RetryPolicy) *RetryService {
	return &RetryService{Store: s, Policy: p}
}
func (r *RetryService) Failure(ctx context.Context, id, owner, message string) error {
	return r.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := r.Store.SetOutboxError(ctx, tx, id, owner, message); err != nil {
			return err
		}
		return nil
	})
}
func (r *RetryService) ShouldRetry(m domain.OutboxMessage) error {
	if m.Attempts >= r.Policy.MaxAttempts {
		return fmt.Errorf("%w: attempts exhausted", domain.ErrConflict)
	}
	if m.State != string(domain.OutboxPending) {
		return fmt.Errorf("%w: state %s", domain.ErrConflict, m.State)
	}
	return nil
}
