package sqlite

import (
	"context"
	"github.com/VanceMichael/harborflow/internal/domain"
	"sort"
)

type StateCount struct {
	State string
	Count int
}

func (s *Store) PassengerStateCounts(ctx context.Context, portID string) []StateCount {
	rows, err := s.DB.QueryContext(ctx, `SELECT p.state,COUNT(*) FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=? GROUP BY p.state`, portID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []StateCount
	for rows.Next() {
		var x StateCount
		if rows.Scan(&x.State, &x.Count) == nil {
			out = append(out, x)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].State < out[j].State })
	return out
}
func (s *Store) IncidentStateCounts(ctx context.Context, portID string) []StateCount {
	rows, err := s.DB.QueryContext(ctx, `SELECT state,COUNT(*) FROM incidents WHERE port_id=? GROUP BY state`, portID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []StateCount
	for rows.Next() {
		var x StateCount
		if rows.Scan(&x.State, &x.Count) == nil {
			out = append(out, x)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].State < out[j].State })
	return out
}
func (s *Store) HasActiveWork(ctx context.Context, portID string) bool {
	var n int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM stress_runs WHERE port_id=? AND state IN (?,?)`, portID, string(domain.RunRunning), string(domain.RunPaused)).Scan(&n); err != nil {
		return false
	}
	return n > 0
}
