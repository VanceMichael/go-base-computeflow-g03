package incident

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"time"
)

type Service struct{ Store *sqlite.Store }

func New(s *sqlite.Store) *Service { return &Service{Store: s} }
func (s *Service) Open(ctx context.Context, portID, runID, subjectType, subjectID, severity, description string, now time.Time) (domain.Incident, error) {
	i := domain.Incident{ID: uuid.NewString(), PortID: portID, RunID: runID, SubjectType: subjectType, SubjectID: subjectID, State: domain.IncidentOpen, Severity: severity, Description: description, Version: 1, CreatedAt: now}
	return i, s.Store.InsertIncident(ctx, i)
}
func (s *Service) Transition(ctx context.Context, i domain.Incident, to domain.IncidentState) error {
	if err := domain.TransitionIncident(i.State, to); err != nil {
		return err
	}
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.UpdateIncidentState(ctx, tx, i.ID, i.PortID, string(i.State), string(to), i.Version)
		return e
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: incident version", domain.ErrConflict)
	}
	return nil
}
