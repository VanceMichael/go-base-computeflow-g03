package sqlite

import (
	"context"
	"database/sql"
	"time"
)

func (s *Store) WithRunUpdate(ctx context.Context, id, portID, from, to string, version int, when time.Time) (bool, error) {
	var n int64
	err := s.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		n, e = rowsAffected(tx, ctx, `UPDATE stress_runs SET state=?,version=version+1,ends_at=CASE WHEN ?='completed' THEN ? ELSE ends_at END WHERE id=? AND port_id=? AND state=? AND version=?`, to, to, stamp(when), id, portID, from, version)
		return e
	})
	return n == 1, err
}
