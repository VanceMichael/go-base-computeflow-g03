package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
)

type IncidentPage struct {
	Items []domain.Incident
	Info  domain.PageInfo
}

func (s *Store) IncidentPage(ctx context.Context, portID string, page domain.Page) (IncidentPage, error) {
	var total int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM incidents WHERE port_id=?`, portID).Scan(&total); err != nil {
		return IncidentPage{}, err
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT id,port_id,COALESCE(run_id,''),subject_type,subject_id,state,severity,description,version,created_at FROM incidents WHERE port_id=? ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`, portID, page.Limit, page.Offset)
	if err != nil {
		return IncidentPage{}, err
	}
	defer rows.Close()
	items, err := scanIncidents(rows)
	return IncidentPage{Items: items, Info: domain.NewPageInfo(total, page)}, err
}
func scanIncidents(rows *sql.Rows) ([]domain.Incident, error) {
	var out []domain.Incident
	for rows.Next() {
		var i domain.Incident
		var created string
		if err := rows.Scan(&i.ID, &i.PortID, &i.RunID, &i.SubjectType, &i.SubjectID, &i.State, &i.Severity, &i.Description, &i.Version, &created); err != nil {
			return nil, err
		}
		i.CreatedAt, _ = scanTime(created)
		out = append(out, i)
	}
	return out, rows.Err()
}
