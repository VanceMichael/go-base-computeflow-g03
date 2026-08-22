package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
)

func (s *Store) CheckConsistency(ctx context.Context, portID string) error {
	checks := []struct{ name, query string }{{"orphan passengers", `SELECT COUNT(*) FROM passengers p LEFT JOIN passenger_waves w ON w.id=p.wave_id WHERE w.id IS NULL`}, {"orphan vehicles", `SELECT COUNT(*) FROM vehicles v LEFT JOIN vehicle_batches b ON b.id=v.batch_id WHERE b.id IS NULL`}, {"duplicate lane ownership", `SELECT COUNT(*) FROM (SELECT lane_id,COUNT(*) n FROM lane_assignments WHERE state='active' GROUP BY lane_id HAVING n>1)`}}
	for _, check := range checks {
		var n int
		if err := s.DB.QueryRowContext(ctx, check.query).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: %s", domain.ErrConflict, check.name)
		}
	}
	return nil
}
func (s *Store) InTx(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	return s.WithTx(ctx, fn)
}
