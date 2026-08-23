package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
	"time"
)

func (s *Store) InsertGatesAndLanes(ctx context.Context, tx *sql.Tx, portID string, gates []domain.Gate, lanes []domain.Lane) error {
	for _, g := range gates {
		if err := exec(tx, ctx, `INSERT INTO gates(id,port_id,gate_no,mode,active) VALUES(?,?,?,?,?)`, g.ID, portID, g.GateNo, g.Mode, boolInt(g.Active)); err != nil {
			return err
		}
	}
	for _, l := range lanes {
		if err := exec(tx, ctx, `INSERT INTO lanes(id,port_id,lane_no,state,version) VALUES(?,?,?,?,?)`, l.ID, portID, l.LaneNo, string(l.State), l.Version); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) SetGateActive(ctx context.Context, id string, active bool) error {
	return exec(s.DB, ctx, `UPDATE gates SET active=? WHERE id=?`, boolInt(active), id)
}
func (s *Store) OpenLaneCount(ctx context.Context, portID string) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM lanes WHERE port_id=? AND state='open'`, portID).Scan(&n)
	return n, err
}
func (s *Store) ExpireLeases(ctx context.Context, now time.Time) (int64, error) {
	r, err := s.DB.ExecContext(ctx, `UPDATE lane_assignments SET state='released',owner='' WHERE state='active' AND lease_until<?`, stamp(now))
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
