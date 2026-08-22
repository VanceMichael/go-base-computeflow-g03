package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
)

func (s *Store) InsertIncident(ctx context.Context, i domain.Incident) error {
	var runID any
	if i.RunID != "" {
		runID = i.RunID
	}
	return exec(s.DB, ctx, `INSERT INTO incidents(id,port_id,run_id,subject_type,subject_id,state,severity,description,version,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, i.ID, i.PortID, runID, i.SubjectType, i.SubjectID, string(i.State), i.Severity, i.Description, i.Version, stamp(i.CreatedAt))
}
func (s *Store) GetIncident(ctx context.Context, id, portID string) (domain.Incident, error) {
	var i domain.Incident
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT id,port_id,COALESCE(run_id,''),subject_type,subject_id,state,severity,description,version,created_at FROM incidents WHERE id=? AND port_id=?`, id, portID).Scan(&i.ID, &i.PortID, &i.RunID, &i.SubjectType, &i.SubjectID, &i.State, &i.Severity, &i.Description, &i.Version, &created)
	if err != nil {
		return i, err
	}
	i.CreatedAt, _ = scanTime(created)
	return i, nil
}
func (s *Store) UpdateIncidentState(ctx context.Context, tx *sql.Tx, id, portID, from, to string, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE incidents SET state=?,version=version+1 WHERE id=? AND port_id=? AND state=? AND version=?`, to, id, portID, from, version)
	return n == 1, err
}
func (s *Store) InsertResponder(ctx context.Context, r domain.Responder) error {
	return exec(s.DB, ctx, `INSERT INTO responders(id,port_id,name,state,version) VALUES(?,?,?,?,?)`, r.ID, r.PortID, r.Name, r.State, r.Version)
}
func (s *Store) AssignResponder(ctx context.Context, tx *sql.Tx, id, incidentID, responderID, owner string) (bool, error) {
	n, err := rowsAffected(tx, ctx, `INSERT INTO dispatch_assignments(id,incident_id,responder_id,owner,state,version) SELECT ?,?,?,?,'active',1 WHERE EXISTS(SELECT 1 FROM responders WHERE id=? AND state='available')`, id, incidentID, responderID, owner, responderID)
	return n == 1, err
}
func (s *Store) SetResponderState(ctx context.Context, tx *sql.Tx, id, from, to string, version int) (bool, error) {
	n, err := rowsAffected(tx, ctx, `UPDATE responders SET state=?,version=version+1 WHERE id=? AND state=? AND version=?`, to, id, from, version)
	return n == 1, err
}
