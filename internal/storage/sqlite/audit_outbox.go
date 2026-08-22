package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

func (s *Store) InsertAudit(ctx context.Context, tx *sql.Tx, e domain.AuditEvent) error {
	return exec(tx, ctx, `INSERT INTO audit_events(id,port_id,actor_id,action,subject_type,subject_id,outcome,request_id,details,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, e.ID, e.PortID, e.ActorID, e.Action, e.SubjectType, e.SubjectID, e.Outcome, e.RequestID, e.Details, stamp(e.CreatedAt))
}
func (s *Store) ListAudit(ctx context.Context, portID string, from, to time.Time) ([]domain.AuditEvent, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,port_id,COALESCE(actor_id,''),action,subject_type,subject_id,outcome,request_id,details,created_at FROM audit_events WHERE port_id=? AND created_at>=? AND created_at<? ORDER BY created_at,id`, portID, stamp(from), stamp(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		var created string
		if err := rows.Scan(&e.ID, &e.PortID, &e.ActorID, &e.Action, &e.SubjectType, &e.SubjectID, &e.Outcome, &e.RequestID, &e.Details, &created); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = scanTime(created)
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s *Store) InsertOutbox(ctx context.Context, tx *sql.Tx, m domain.OutboxMessage, payload string) error {
	return exec(tx, ctx, `INSERT INTO outbox_messages(id,port_id,event_key,payload,state,attempts,idempotency_key,created_at) VALUES(?,?,?,?,?,?,?,?)`, m.ID, m.PortID, m.EventKey, payload, string(m.State), m.Attempts, m.IdempotencyKey, stamp(m.CreatedAt))
}
func (s *Store) ClaimOutbox(ctx context.Context, tx *sql.Tx, id, owner string, until time.Time) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE outbox_messages SET state='leased',owner=?,lease_until=?,attempts=attempts+1 WHERE id=? AND state='pending'`, owner, stamp(until), id)
	return n == 1, err
}
func (s *Store) MarkOutbox(ctx context.Context, tx *sql.Tx, id, owner string, now time.Time) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE outbox_messages SET state='delivered',owner=NULL,lease_until=NULL,delivered_at=? WHERE id=? AND state='leased' AND owner=?`, stamp(now), id, owner)
	return n == 1, err
}
func (s *Store) RequeueExpiredOutbox(ctx context.Context, tx *sql.Tx, now time.Time) (int64, error) {
	return rowsAffected(tx, ctx, `UPDATE outbox_messages SET state='pending',owner=NULL,lease_until=NULL WHERE state='leased' AND lease_until<?`, stamp(now))
}
func (s *Store) ReadIdempotency(ctx context.Context, portID, operation, key string) (int, string, error) {
	var code int
	var body string
	err := s.DB.QueryRowContext(ctx, `SELECT response_code,response_body FROM idempotency_records WHERE port_id=? AND operation=? AND request_key=?`, portID, operation, key).Scan(&code, &body)
	return code, body, err
}
func (s *Store) WriteIdempotency(ctx context.Context, tx *sql.Tx, portID, operation, key string, code int, body string, now time.Time) error {
	return exec(tx, ctx, `INSERT INTO idempotency_records(id,port_id,operation,request_key,response_code,response_body,created_at) VALUES(?,?,?,?,?,?,?)`, portID+operation+key, portID, operation, key, code, body, stamp(now))
}
