package sqlite

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/domain"
)

func (s *Store) ListOpenIncidents(ctx context.Context, portID string, limit, offset int) ([]domain.Incident, int, error) {
	var total int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM incidents WHERE port_id=? AND state IN ('open','acknowledged','resolved')`, portID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT id,port_id,COALESCE(run_id,''),subject_type,subject_id,state,severity,description,version,created_at FROM incidents WHERE port_id=? AND state IN ('open','acknowledged','resolved') ORDER BY created_at,id LIMIT ? OFFSET ?`, portID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []domain.Incident
	for rows.Next() {
		var i domain.Incident
		var created string
		if err := rows.Scan(&i.ID, &i.PortID, &i.RunID, &i.SubjectType, &i.SubjectID, &i.State, &i.Severity, &i.Description, &i.Version, &created); err != nil {
			return nil, 0, err
		}
		i.CreatedAt, _ = scanTime(created)
		result = append(result, i)
	}
	return result, total, rows.Err()
}
func (s *Store) CountArtifactsByState(ctx context.Context, portID string) map[string]int {
	result := map[string]int{}
	rows, err := s.DB.QueryContext(ctx, `SELECT state,COUNT(*) FROM passengers p JOIN passenger_waves w ON w.id=p.wave_id JOIN stress_runs r ON r.id=w.run_id WHERE r.port_id=? GROUP BY state`, portID)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		var n int
		if rows.Scan(&state, &n) == nil {
			result[state] = n
		}
	}
	return result
}
