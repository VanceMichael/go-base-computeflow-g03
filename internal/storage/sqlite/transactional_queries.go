package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
	"time"
)

type TransitionRecord struct {
	SubjectID string
	From      string
	To        string
	Version   int
	At        time.Time
}

func (s *Store) TransitionPassenger(ctx context.Context, id, portID, from, to string, version int) error {
	return s.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		n, err := rowsAffected(tx, ctx, `UPDATE passengers SET state=?,version=version+1 WHERE id=? AND state=? AND version=? AND wave_id IN (SELECT w.id FROM passenger_waves w JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=?)`, to, id, from, version, portID)
		if err != nil {
			return err
		}
		if n != 1 {
			return domain.ErrConflict
		}
		return nil
	})
}
func (s *Store) TransitionVehicle(ctx context.Context, id, portID, from, to string, version int) error {
	return s.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		n, err := rowsAffected(tx, ctx, `UPDATE vehicles SET state=?,version=version+1 WHERE id=? AND state=? AND version=? AND batch_id IN (SELECT b.id FROM vehicle_batches b JOIN stress_runs r ON r.id=b.run_id WHERE r.port_id=?)`, to, id, from, version, portID)
		if err != nil {
			return err
		}
		if n != 1 {
			return domain.ErrConflict
		}
		return nil
	})
}
func (s *Store) WriteAuditAndOutbox(ctx context.Context, portID, actor, action, subjectType, subjectID, payload, eventKey, requestID string, now time.Time) error {
	return s.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.InsertAudit(ctx, tx, domain.AuditEvent{ID: subjectID + requestID, PortID: portID, ActorID: actor, Action: action, SubjectType: subjectType, SubjectID: subjectID, Outcome: "accepted", RequestID: requestID, Details: payload, CreatedAt: now}); err != nil {
			return err
		}
		return s.InsertOutbox(ctx, tx, domain.OutboxMessage{ID: eventKey + requestID, PortID: portID, EventKey: eventKey, State: string(domain.OutboxPending), IdempotencyKey: requestID, CreatedAt: now}, payload)
	})
}
