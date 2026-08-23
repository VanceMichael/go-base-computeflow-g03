package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
	"time"
)

type RunSummary struct {
	RunID             string
	State             string
	Version           int
	WaveCount         int
	VehicleBatchCount int
	IncidentCount     int
}

func (s *Store) RunSummary(ctx context.Context, runID, portID string) (RunSummary, error) {
	var out RunSummary
	out.RunID = runID
	var state string
	if err := s.DB.QueryRowContext(ctx, `SELECT state,version FROM stress_runs WHERE id=? AND port_id=?`, runID, portID).Scan(&state, &out.Version); err != nil {
		return out, err
	}
	out.State = state
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM passenger_waves WHERE run_id=?`, runID).Scan(&out.WaveCount); err != nil {
		return out, err
	}
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_batches WHERE run_id=?`, runID).Scan(&out.VehicleBatchCount); err != nil {
		return out, err
	}
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM incidents WHERE run_id=? AND port_id=? AND state<>?`, runID, portID, string(domain.IncidentClosed)).Scan(&out.IncidentCount); err != nil {
		return out, err
	}
	return out, nil
}
func (s *Store) CreateRunWithResources(ctx context.Context, r domain.StressRun, w domain.PassengerWave, passengers []domain.Passenger, batch domain.VehicleBatch, vehicles []domain.Vehicle, gates []domain.Gate, lanes []domain.Lane) error {
	return s.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := exec(tx, ctx, `INSERT INTO stress_runs(id,port_id,name,state,version,starts_at,created_by,created_at) VALUES(?,?,?,?,?,?,?,?)`, r.ID, r.PortID, r.Name, string(r.State), r.Version, stamp(r.StartsAt), r.CreatedBy, stamp(r.CreatedAt)); err != nil {
			return err
		}
		if err := exec(tx, ctx, `INSERT INTO passenger_waves(id,run_id,sequence_no,state,version,planned_at) VALUES(?,?,?,?,?,?)`, w.ID, w.RunID, w.SequenceNo, string(w.State), w.Version, stamp(w.PlannedAt)); err != nil {
			return err
		}
		for _, p := range passengers {
			if err := exec(tx, ctx, `INSERT INTO passengers(id,wave_id,document_key,state,version,created_at) VALUES(?,?,?,?,?,?)`, p.ID, p.WaveID, p.DocumentKey, string(p.State), p.Version, stamp(p.CreatedAt)); err != nil {
				return err
			}
		}
		if err := exec(tx, ctx, `INSERT INTO vehicle_batches(id,run_id,manifest_key,state,version,created_at) VALUES(?,?,?,?,?,?)`, batch.ID, batch.RunID, batch.ManifestKey, batch.State, batch.Version, stamp(batch.CreatedAt)); err != nil {
			return err
		}
		for _, v := range vehicles {
			if err := exec(tx, ctx, `INSERT INTO vehicles(id,batch_id,plate_key,state,version,created_at) VALUES(?,?,?,?,?,?)`, v.ID, v.BatchID, v.PlateKey, string(v.State), v.Version, stamp(v.CreatedAt)); err != nil {
				return err
			}
		}
		return s.InsertGatesAndLanes(ctx, tx, r.PortID, gates, lanes)
	})
}
func (s *Store) SetWaveReleaseTime(ctx context.Context, tx *sql.Tx, id string, when time.Time) error {
	return exec(tx, ctx, `UPDATE passenger_waves SET released_at=? WHERE id=?`, stamp(when), id)
}
func (s *Store) FindPassenger(ctx context.Context, id, portID string) (domain.Passenger, error) {
	var p domain.Passenger
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT p.id,p.wave_id,p.document_key,p.state,p.version,p.created_at FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE p.id=? AND r.port_id=?`, id, portID).Scan(&p.ID, &p.WaveID, &p.DocumentKey, &p.State, &p.Version, &created)
	if err != nil {
		return p, err
	}
	p.CreatedAt, _ = scanTime(created)
	return p, nil
}
func (s *Store) FindVehicle(ctx context.Context, id, portID string) (domain.Vehicle, error) {
	var v domain.Vehicle
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT v.id,v.batch_id,v.plate_key,v.state,v.version,v.created_at FROM vehicles v JOIN vehicle_batches b ON b.id=v.batch_id JOIN stress_runs r ON r.id=b.run_id WHERE v.id=? AND r.port_id=?`, id, portID).Scan(&v.ID, &v.BatchID, &v.PlateKey, &v.State, &v.Version, &created)
	if err != nil {
		return v, err
	}
	v.CreatedAt, _ = scanTime(created)
	return v, nil
}
func (s *Store) FindOutbox(ctx context.Context, id, portID string) (domain.OutboxMessage, error) {
	var m domain.OutboxMessage
	var until, created, delivered sql.NullString
	err := s.DB.QueryRowContext(ctx, `SELECT id,port_id,event_key,payload,state,COALESCE(owner,''),idempotency_key,lease_until,attempts,COALESCE(last_error,''),created_at,delivered_at FROM outbox_messages WHERE id=? AND port_id=?`, id, portID).Scan(&m.ID, &m.PortID, &m.EventKey, &m.Payload, &m.State, &m.Owner, &m.IdempotencyKey, &until, &m.Attempts, &m.LastError, &created, &delivered)
	if err != nil {
		return m, err
	}
	m.LeaseUntil = nullableTime(until)
	m.CreatedAt, _ = scanTime(created.String)
	m.DeliveredAt = nullableTime(delivered)
	return m, nil
}
func (s *Store) SetOutboxError(ctx context.Context, tx *sql.Tx, id, owner, message string, maxAttempts int) error {
	return exec(tx, ctx, `UPDATE outbox_messages SET state=CASE WHEN attempts>=? THEN 'dead' ELSE 'pending' END,owner=NULL,lease_until=NULL,last_error=? WHERE id=? AND state='leased' AND owner=?`, maxAttempts, message, id, owner)
}
