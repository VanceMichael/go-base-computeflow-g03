package query

import (
	"fmt"
	"strings"
	"time"
)

type AuditFilter struct {
	PortID      string
	RunID       string
	SubjectType string
	From        time.Time
	To          time.Time
	PageLimit   int
	PageOffset  int
}

func (f AuditFilter) Validate() error {
	if f.PortID == "" {
		return fmt.Errorf("port_id is required")
	}
	if !f.From.IsZero() && !f.To.IsZero() && !f.From.Before(f.To) {
		return fmt.Errorf("from must precede to")
	}
	if f.PageLimit < 0 || f.PageLimit > 200 {
		return fmt.Errorf("invalid page limit")
	}
	if f.PageOffset < 0 {
		return fmt.Errorf("invalid page offset")
	}
	return nil
}
func (f AuditFilter) Where(start int) (string, []any) {
	clauses := []string{"port_id=?"}
	args := []any{f.PortID}
	if f.RunID != "" {
		clauses = append(clauses, "subject_id IN (SELECT id FROM stress_runs WHERE id=? AND port_id=?)")
		args = append(args, f.RunID, f.PortID)
	}
	if f.SubjectType != "" {
		clauses = append(clauses, "subject_type=?")
		args = append(args, f.SubjectType)
	}
	if !f.From.IsZero() {
		clauses = append(clauses, "created_at>=?")
		args = append(args, f.From.UTC().Format(time.RFC3339Nano))
	}
	if !f.To.IsZero() {
		clauses = append(clauses, "created_at<?")
		args = append(args, f.To.UTC().Format(time.RFC3339Nano))
	}
	return strings.Join(clauses, " AND "), args
}

type VehicleFilter struct {
	PortID string
	State  string
	Since  time.Time
}

func (f VehicleFilter) Validate() error {
	if f.PortID == "" {
		return fmt.Errorf("port_id is required")
	}
	return nil
}
