package incident

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
)

type QueryService struct{ Store *sqlite.Store }

func NewQueryService(s *sqlite.Store) *QueryService { return &QueryService{Store: s} }
func (q *QueryService) Get(ctx context.Context, id, portID string) (domain.Incident, error) {
	return q.Store.GetIncident(ctx, id, portID)
}
func (q *QueryService) Active(ctx context.Context, portID string, page domain.Page) (sqlite.IncidentPage, error) {
	return q.Store.IncidentPage(ctx, portID, page)
}
func (q *QueryService) RequiresMedical(description string) bool {
	return Classify(description).RequiresMedical
}
