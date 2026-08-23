package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type ConstraintReport struct {
	ForeignKeys        int
	OpenLeasedScans    int
	ActiveAssignments  int
	DuplicateDocuments int
}

func (s *Store) ConstraintReport(ctx context.Context, portID string) (ConstraintReport, error) {
	var r ConstraintReport
	queries := []struct {
		dest *int
		sql  string
	}{{&r.ForeignKeys, `PRAGMA foreign_key_check`}, {&r.OpenLeasedScans, `SELECT COUNT(*) FROM gate_scans g JOIN passengers p ON p.id=g.passenger_id JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=? AND g.state='leased'`}, {&r.ActiveAssignments, `SELECT COUNT(*) FROM lane_assignments a JOIN lanes l ON l.id=a.lane_id WHERE l.port_id=? AND a.state='active'`}, {&r.DuplicateDocuments, `SELECT COUNT(*) FROM (SELECT p.document_key FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=? GROUP BY p.document_key HAVING COUNT(*)>1)`}}
	for i, item := range queries {
		var err error
		if i == 0 {
			err = s.DB.QueryRowContext(ctx, item.sql).Scan(item.dest)
		} else {
			err = s.DB.QueryRowContext(ctx, item.sql, portID).Scan(item.dest)
		}
		if err != nil {
			return r, err
		}
	}
	return r, nil
}
func (r ConstraintReport) Healthy() bool { return r.ForeignKeys == 0 && r.DuplicateDocuments == 0 }
func (s *Store) EnsureForeignKeys(ctx context.Context) error {
	var value int
	if err := s.DB.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&value); err != nil {
		return err
	}
	if value != 1 {
		return fmt.Errorf("%w: foreign keys disabled", domain.ErrConflict)
	}
	return nil
}
func (s *Store) ExecTx(ctx context.Context, fn func(*sql.Tx) error) error {
	return s.WithTx(ctx, func(_ context.Context, tx *sql.Tx) error { return fn(tx) })
}
