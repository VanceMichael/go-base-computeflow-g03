package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

type DependencyHealth struct {
	Database         bool
	ForeignKeys      bool
	MigrationVersion int
	CheckedAt        time.Time
}

func (s *Store) DependencyHealth(ctx context.Context) DependencyHealth {
	h := DependencyHealth{CheckedAt: s.Now()}
	if err := s.DB.PingContext(ctx); err != nil {
		return h
	}
	h.Database = true
	var fk int
	if s.DB.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&fk) == nil {
		h.ForeignKeys = fk == 1
	}
	_ = s.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&h.MigrationVersion)
	return h
}
func (s *Store) ValidateReference(ctx context.Context, table, id, portID string) error {
	allowed := map[string]string{"stress_runs": "SELECT 1 FROM stress_runs WHERE id=? AND port_id=?", "incidents": "SELECT 1 FROM incidents WHERE id=? AND port_id=?", "lanes": "SELECT 1 FROM lanes WHERE id=? AND port_id=?"}
	query, ok := allowed[table]
	if !ok {
		return domain.ErrInvalid
	}
	var one int
	if err := s.DB.QueryRowContext(ctx, query, id, portID).Scan(&one); err != nil {
		return err
	}
	return nil
}
func (s *Store) CountRows(ctx context.Context, table, portID string) (int, error) {
	queries := map[string]string{"stress_runs": `SELECT COUNT(*) FROM stress_runs WHERE port_id=?`, "incidents": `SELECT COUNT(*) FROM incidents WHERE port_id=?`, "lanes": `SELECT COUNT(*) FROM lanes WHERE port_id=?`, "gates": `SELECT COUNT(*) FROM gates WHERE port_id=?`}
	query, ok := queries[table]
	if !ok {
		return 0, domain.ErrInvalid
	}
	var n int
	err := s.DB.QueryRowContext(ctx, query, portID).Scan(&n)
	return n, err
}
func (s *Store) SetClockAndCheck(now func() time.Time) error {
	if now == nil {
		return domain.ErrInvalid
	}
	s.SetClock(now)
	return nil
}

var _ *sql.Tx
