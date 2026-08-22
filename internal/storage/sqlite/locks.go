package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

type LockResult struct {
	Acquired  bool
	Version   int
	ExpiresAt time.Time
}

func (s *Store) AcquireRunLock(ctx context.Context, tx *sql.Tx, runID, portID, owner string, version int, until time.Time) (LockResult, error) {
	var current int
	if err := tx.QueryRowContext(ctx, `SELECT version FROM stress_runs WHERE id=? AND port_id=? AND state IN ('running','paused')`, runID, portID).Scan(&current); err != nil {
		return LockResult{}, err
	}
	if current != version {
		return LockResult{Version: current}, domain.ErrConflict
	}
	n, err := rowsAffected(tx, ctx, `UPDATE stress_runs SET version=version+1 WHERE id=? AND port_id=? AND version=?`, runID, portID, version)
	return LockResult{Acquired: n == 1, Version: version + 1, ExpiresAt: until}, err
}
func (s *Store) ReleaseRunLock(ctx context.Context, tx *sql.Tx, runID, portID string, version int) error {
	n, err := rowsAffected(tx, ctx, `UPDATE stress_runs SET version=version+1 WHERE id=? AND port_id=? AND version=?`, runID, portID, version)
	if err != nil {
		return err
	}
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s *Store) LeaseIsValid(ctx context.Context, until *time.Time, now time.Time) bool {
	return until != nil && now.Before(*until)
}
