package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

func (s *Store) InsertGate(ctx context.Context, g domain.Gate) error {
	return exec(s.DB, ctx, `INSERT INTO gates(id,port_id,gate_no,mode,active) VALUES(?,?,?,?,?)`, g.ID, g.PortID, g.GateNo, g.Mode, boolInt(g.Active))
}
func (s *Store) InsertScan(ctx context.Context, tx *sql.Tx, x domain.GateScan) error {
	return exec(tx, ctx, `INSERT INTO gate_scans(id,passenger_id,gate_id,stage,state,version,created_at) VALUES(?,?,?,?,?,?,?)`, x.ID, x.PassengerID, x.GateID, x.Stage, string(x.State), x.Version, stamp(x.CreatedAt))
}
func (s *Store) ClaimScan(ctx context.Context, tx *sql.Tx, id, owner, token string, until time.Time, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE gate_scans SET state='leased',lease_owner=?,lease_token=?,lease_until=?,version=version+1 WHERE id=? AND state='pending' AND version=?`, owner, token, stamp(until), id, version)
	return n == 1, err
}
func (s *Store) CompleteScan(ctx context.Context, tx *sql.Tx, id, owner, token, to string, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE gate_scans SET state=?,lease_owner=NULL,lease_token=NULL,lease_until=NULL,version=version+1 WHERE id=? AND state='leased' AND lease_owner=? AND lease_token=? AND version=?`, to, id, owner, token, version)
	return n == 1, err
}
func (s *Store) CreateJointScans(ctx context.Context, tx *sql.Tx, passengerID string, gates []domain.Gate, now time.Time) error {
	for i, g := range gates {
		if err := s.InsertScan(ctx, tx, domain.GateScan{ID: passengerID + "-" + g.ID, PassengerID: passengerID, GateID: g.ID, Stage: i + 1, State: domain.ScanPending, Version: 1, CreatedAt: now}); err != nil {
			return err
		}
	}
	return nil
}
