package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
)

func (s *Store) InsertSnapshot(ctx context.Context, tx *sql.Tx, x domain.CapacitySnapshot) error {
	return exec(tx, ctx, `INSERT INTO capacity_snapshots(id,run_id,window_start,window_end,passengers,cleared,held,vehicles,state,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, x.ID, x.RunID, stamp(x.WindowStart), stamp(x.WindowEnd), x.Passengers, x.Cleared, x.Held, x.Vehicles, x.State, stamp(x.CreatedAt))
}
func (s *Store) CountRunPassengers(ctx context.Context, runID string) (int, int, int, error) {
	var total, cleared, held int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*),SUM(CASE WHEN p.state='cleared' THEN 1 ELSE 0 END),SUM(CASE WHEN p.state='held' THEN 1 ELSE 0 END) FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id WHERE w.run_id=?`, runID).Scan(&total, &cleared, &held)
	return total, cleared, held, err
}
func (s *Store) CountRunVehicles(ctx context.Context, runID string) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicles v JOIN vehicle_batches b ON b.id=v.batch_id WHERE b.run_id=?`, runID).Scan(&n)
	return n, err
}
