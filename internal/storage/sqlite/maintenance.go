package sqlite

import (
	"context"
	"time"
)

func (s *Store) PurgeExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	r, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at<? OR revoked_at<?`, stamp(now), stamp(now))
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
func (s *Store) ArchiveClosedIncidents(ctx context.Context, portID string, before time.Time) (int64, error) {
	r, err := s.DB.ExecContext(ctx, `UPDATE incidents SET state='closed' WHERE port_id=? AND state='resolved' AND created_at<?`, portID, stamp(before))
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
