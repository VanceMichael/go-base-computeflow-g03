package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
	"time"
)

func (s *Store) InsertRun(ctx context.Context, r domain.StressRun) error {
	return exec(s.DB, ctx, `INSERT INTO stress_runs(id,port_id,name,state,version,starts_at,created_by,created_at) VALUES(?,?,?,?,?,?,?,?)`, r.ID, r.PortID, r.Name, string(r.State), r.Version, stamp(r.StartsAt), r.CreatedBy, stamp(r.CreatedAt))
}
func (s *Store) GetRun(ctx context.Context, id, portID string) (domain.StressRun, error) {
	var r domain.StressRun
	var starts, created string
	err := s.DB.QueryRowContext(ctx, `SELECT id,port_id,name,state,version,starts_at,created_by,created_at FROM stress_runs WHERE id=? AND port_id=?`, id, portID).Scan(&r.ID, &r.PortID, &r.Name, &r.State, &r.Version, &starts, &r.CreatedBy, &created)
	if err != nil {
		return r, err
	}
	r.StartsAt, _ = scanTime(starts)
	r.CreatedAt, _ = scanTime(created)
	return r, nil
}
func (s *Store) UpdateRunState(ctx context.Context, tx *sql.Tx, id, from, to string, version int, ended *time.Time) (bool, error) {
	var end any
	if ended != nil {
		end = stamp(*ended)
	}
	n, err := rowsAffected(tx, ctx, `UPDATE stress_runs SET state=?,version=version+1,ends_at=? WHERE id=? AND state=? AND version=?`, to, end, id, from, version)
	return n == 1, err
}
func (s *Store) InsertWave(ctx context.Context, tx *sql.Tx, w domain.PassengerWave) error {
	return exec(tx, ctx, `INSERT INTO passenger_waves(id,run_id,sequence_no,state,version,planned_at) VALUES(?,?,?,?,?,?)`, w.ID, w.RunID, w.SequenceNo, string(w.State), w.Version, stamp(w.PlannedAt))
}
func (s *Store) InsertPassenger(ctx context.Context, tx *sql.Tx, p domain.Passenger) error {
	return exec(tx, ctx, `INSERT INTO passengers(id,wave_id,document_key,state,version,created_at) VALUES(?,?,?,?,?,?)`, p.ID, p.WaveID, p.DocumentKey, string(p.State), p.Version, stamp(p.CreatedAt))
}
func (s *Store) ListPassengers(ctx context.Context, waveID string) ([]domain.Passenger, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,wave_id,document_key,state,version,created_at FROM passengers WHERE wave_id=? ORDER BY created_at,id`, waveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Passenger
	for rows.Next() {
		var p domain.Passenger
		var created string
		if err := rows.Scan(&p.ID, &p.WaveID, &p.DocumentKey, &p.State, &p.Version, &created); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = scanTime(created)
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *Store) UpdateWaveState(ctx context.Context, tx *sql.Tx, id, from, to string, version int, when *time.Time) (bool, error) {
	var released any
	if when != nil {
		released = stamp(*when)
	}
	n, err := rowsAffected(tx, ctx, `UPDATE passenger_waves SET state=?,version=version+1,released_at=? WHERE id=? AND state=? AND version=?`, to, released, id, from, version)
	return n == 1, err
}
func (s *Store) UpdatePassengerState(ctx context.Context, tx *sql.Tx, id, from, to string, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE passengers SET state=?,version=version+1 WHERE id=? AND state=? AND version=?`, to, id, from, version)
	return n == 1, err
}
