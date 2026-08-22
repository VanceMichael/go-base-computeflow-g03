package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"strings"
	"time"
)

type PassengerRow struct {
	Passenger domain.Passenger
	WaveState domain.WaveState
	RunID     string
	RunState  domain.RunState
}
type VehicleRow struct {
	Vehicle    domain.Vehicle
	BatchState string
	RunID      string
	RunState   domain.RunState
}
type QueueFilter struct {
	PortID string
	State  string
	Limit  int
	Offset int
}

func (f QueueFilter) Validate() error {
	if f.PortID == "" {
		return domain.ErrInvalid
	}
	if f.Limit < 1 || f.Limit > 200 || f.Offset < 0 {
		return domain.ErrInvalid
	}
	return nil
}
func (s *Store) PassengerQueue(ctx context.Context, f QueueFilter) ([]PassengerRow, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	clauses := []string{"r.port_id=?"}
	args := []any{f.PortID}
	if f.State != "" {
		clauses = append(clauses, "p.state=?")
		args = append(args, f.State)
	}
	query := `SELECT p.id,p.wave_id,p.document_key,p.state,p.version,p.created_at,w.state,r.id,r.state FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE ` + strings.Join(clauses, " AND ") + ` ORDER BY p.created_at,p.id LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PassengerRow
	for rows.Next() {
		var x PassengerRow
		var created string
		if err := rows.Scan(&x.Passenger.ID, &x.Passenger.WaveID, &x.Passenger.DocumentKey, &x.Passenger.State, &x.Passenger.Version, &created, &x.WaveState, &x.RunID, &x.RunState); err != nil {
			return nil, err
		}
		x.Passenger.CreatedAt, _ = scanTime(created)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) CountPassengerQueue(ctx context.Context, portID, state string) (int, error) {
	query := `SELECT COUNT(*) FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=?`
	args := []any{portID}
	if state != "" {
		query += ` AND p.state=?`
		args = append(args, state)
	}
	var n int
	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&n)
	return n, err
}
func (s *Store) VehicleQueue(ctx context.Context, f QueueFilter) ([]VehicleRow, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	query := `SELECT v.id,v.batch_id,v.plate_key,v.state,v.version,v.created_at,b.state,r.id,r.state FROM vehicles v JOIN vehicle_batches b ON b.id=v.batch_id JOIN stress_runs r ON r.id=b.run_id WHERE r.port_id=?`
	args := []any{f.PortID}
	if f.State != "" {
		query += ` AND v.state=?`
		args = append(args, f.State)
	}
	query += ` ORDER BY v.created_at,v.id LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VehicleRow
	for rows.Next() {
		var x VehicleRow
		var created string
		if err := rows.Scan(&x.Vehicle.ID, &x.Vehicle.BatchID, &x.Vehicle.PlateKey, &x.Vehicle.State, &x.Vehicle.Version, &created, &x.BatchState, &x.RunID, &x.RunState); err != nil {
			return nil, err
		}
		x.Vehicle.CreatedAt, _ = scanTime(created)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) IncidentCountBySeverity(ctx context.Context, portID string) map[string]int {
	out := map[string]int{}
	rows, err := s.DB.QueryContext(ctx, `SELECT severity,COUNT(*) FROM incidents WHERE port_id=? AND state<>? GROUP BY severity`, portID, string(domain.IncidentClosed))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var n int
		if rows.Scan(&k, &n) == nil {
			out[k] = n
		}
	}
	return out
}
func (s *Store) GateUtilization(ctx context.Context, portID string, from, to time.Time) map[string]int {
	out := map[string]int{}
	rows, err := s.DB.QueryContext(ctx, `SELECT g.gate_no,COUNT(*) FROM gate_scans s JOIN gates g ON g.id=s.gate_id WHERE g.port_id=? AND s.created_at>=? AND s.created_at<? GROUP BY g.gate_no`, portID, stamp(from), stamp(to))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var gateNo, n int
		if rows.Scan(&gateNo, &n) == nil {
			out[fmt.Sprint(gateNo)] = n
		}
	}
	return out
}
func (s *Store) WriteIdempotentTransition(ctx context.Context, portID, operation, key string, fn func(context.Context, *sql.Tx) (int, string, error), now time.Time) (int, string, error) {
	code, body, err := s.ReadIdempotency(ctx, portID, operation, key)
	if err == nil {
		return code, body, nil
	}
	var resultCode int
	var resultBody string
	err = s.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		resultCode, resultBody, e = fn(ctx, tx)
		if e != nil {
			return e
		}
		return s.WriteIdempotency(ctx, tx, portID, operation, key, resultCode, resultBody, now)
	})
	return resultCode, resultBody, err
}
func (s *Store) VerifyOpenRun(ctx context.Context, runID, portID string) error {
	var state string
	if err := s.DB.QueryRowContext(ctx, `SELECT state FROM stress_runs WHERE id=? AND port_id=?`, runID, portID).Scan(&state); err != nil {
		return err
	}
	if state != "running" && state != "paused" {
		return fmt.Errorf("%w: run %s", domain.ErrConflict, state)
	}
	return nil
}
