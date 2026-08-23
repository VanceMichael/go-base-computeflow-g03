package audit

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"time"
)

type Service struct{ Store *sqlite.Store }

func New(s *sqlite.Store) *Service { return &Service{Store: s} }
func (s *Service) Record(ctx context.Context, tx *sql.Tx, portID, actor, action, subjectType, subjectID, outcome, requestID, details string, now time.Time) error {
	return s.Store.InsertAudit(ctx, tx, domain.AuditEvent{ID: uuid.NewString(), PortID: portID, ActorID: actor, Action: action, SubjectType: subjectType, SubjectID: subjectID, Outcome: outcome, RequestID: requestID, Details: details, CreatedAt: now})
}
func (s *Service) Timeline(ctx context.Context, portID string, from, to time.Time) ([]domain.AuditEvent, error) {
	return s.Store.ListAudit(ctx, portID, from, to)
}
