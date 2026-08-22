package sqlite

import (
	"context"
	"time"
)

func (s *Store) DeleteOldAudit(ctx context.Context, portID string, before time.Time) (int64, error) {
	r, err := s.DB.ExecContext(ctx, `DELETE FROM audit_events WHERE port_id=? AND created_at<?`, portID, stamp(before))
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
func (s *Store) DeleteDeliveredOutbox(ctx context.Context, portID string, before time.Time) (int64, error) {
	r, err := s.DB.ExecContext(ctx, `DELETE FROM outbox_messages WHERE port_id=? AND state='delivered' AND delivered_at<?`, portID, stamp(before))
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
