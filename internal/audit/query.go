package audit

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"sort"
	"time"
)

type Query struct{ Store *sqlite.Store }

func NewQuery(s *sqlite.Store) *Query { return &Query{Store: s} }
func (q *Query) Events(ctx context.Context, portID string, from, to time.Time) ([]domain.AuditEvent, error) {
	return q.Store.ListAudit(ctx, portID, from, to)
}
func SortEvents(events []domain.AuditEvent) []domain.AuditEvent {
	out := append([]domain.AuditEvent(nil), events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}
func GroupByAction(events []domain.AuditEvent) map[string]int {
	out := map[string]int{}
	for _, e := range events {
		out[e.Action]++
	}
	return out
}
