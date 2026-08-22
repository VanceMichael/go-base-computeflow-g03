package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

func (s *Store) InsertBatch(ctx context.Context, tx *sql.Tx, b domain.VehicleBatch) error {
	return exec(tx, ctx, `INSERT INTO vehicle_batches(id,run_id,manifest_key,state,version,created_at) VALUES(?,?,?,?,?,?)`, b.ID, b.RunID, b.ManifestKey, b.State, b.Version, stamp(b.CreatedAt))
}
func (s *Store) InsertVehicle(ctx context.Context, tx *sql.Tx, v domain.Vehicle) error {
	return exec(tx, ctx, `INSERT INTO vehicles(id,batch_id,plate_key,state,version,created_at) VALUES(?,?,?,?,?,?)`, v.ID, v.BatchID, v.PlateKey, string(v.State), v.Version, stamp(v.CreatedAt))
}
func (s *Store) InsertLane(ctx context.Context, l domain.Lane) error {
	return exec(s.DB, ctx, `INSERT INTO lanes(id,port_id,lane_no,state,version) VALUES(?,?,?,?,?)`, l.ID, l.PortID, l.LaneNo, string(l.State), l.Version)
}
func (s *Store) GetLane(ctx context.Context, id, portID string) (domain.Lane, error) {
	var l domain.Lane
	err := s.DB.QueryRowContext(ctx, `SELECT id,port_id,lane_no,state,version FROM lanes WHERE id=? AND port_id=?`, id, portID).Scan(&l.ID, &l.PortID, &l.LaneNo, &l.State, &l.Version)
	return l, err
}
func (s *Store) AssignLane(ctx context.Context, tx *sql.Tx, aID, laneID, vehicleID, owner string, until time.Time) (bool, error) {
	n, err := rowsAffected(tx, ctx, `INSERT INTO lane_assignments(id,lane_id,vehicle_id,owner,state,lease_until,version) SELECT ?,?,?,?,?,?,1 WHERE EXISTS(SELECT 1 FROM lanes WHERE id=? AND state='open')`, aID, laneID, vehicleID, owner, "active", stamp(until), laneID)
	return n == 1, err
}
func (s *Store) UpdateVehicleState(ctx context.Context, tx *sql.Tx, id, from, to string, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE vehicles SET state=?,version=version+1 WHERE id=? AND state=? AND version=?`, to, id, from, version)
	return n == 1, err
}
func (s *Store) UpdateBatchState(ctx context.Context, tx *sql.Tx, id, from, to string, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE vehicle_batches SET state=?,version=version+1 WHERE id=? AND state=? AND version=?`, to, id, from, version)
	return n == 1, err
}
func (s *Store) ReleaseLaneAssignments(ctx context.Context, tx *sql.Tx, laneID string) error {
	return exec(tx, ctx, `UPDATE lane_assignments SET state='released' WHERE lane_id=? AND state='active'`, laneID)
}
func (s *Store) CountActiveVehicles(ctx context.Context, laneID string) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM lane_assignments WHERE lane_id=? AND state='active'`, laneID).Scan(&n)
	return n, err
}
