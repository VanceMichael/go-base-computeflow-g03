package sqlite

import (
	"context"
	"time"
)

type PortMetrics struct {
	OpenRuns       int
	ActiveWaves    int
	QueuedVehicles int
	OpenIncidents  int
	PendingOutbox  int
}

func (s *Store) PortMetrics(ctx context.Context, portID string) (PortMetrics, error) {
	var m PortMetrics
	queries := []struct {
		dest  *int
		query string
	}{{&m.OpenRuns, `SELECT COUNT(*) FROM stress_runs WHERE port_id=? AND state IN ('running','paused')`}, {&m.ActiveWaves, `SELECT COUNT(*) FROM passenger_waves w JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=? AND w.state IN ('releasing','released')`}, {&m.QueuedVehicles, `SELECT COUNT(*) FROM vehicles v JOIN vehicle_batches b ON b.id=v.batch_id JOIN stress_runs r ON r.id=b.run_id WHERE r.port_id=? AND v.state IN ('queued','assessing')`}, {&m.OpenIncidents, `SELECT COUNT(*) FROM incidents WHERE port_id=? AND state IN ('open','acknowledged','resolved')`}, {&m.PendingOutbox, `SELECT COUNT(*) FROM outbox_messages WHERE port_id=? AND state IN ('pending','leased')`}}
	for _, item := range queries {
		if err := s.DB.QueryRowContext(ctx, item.query, portID).Scan(item.dest); err != nil {
			return m, err
		}
	}
	return m, nil
}
func (s *Store) HasRecentActivity(ctx context.Context, portID string, since time.Time) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_events WHERE port_id=? AND created_at>=?`, portID, stamp(since)).Scan(&n)
	return n > 0, err
}
