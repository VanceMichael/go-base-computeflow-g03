package capacity

import (
	"context"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"time"
)

type Query struct {
	Store *sqlite.Store
	Zone  *time.Location
}

func NewQuery(s *sqlite.Store, z *time.Location) *Query { return &Query{Store: s, Zone: z} }
func (q *Query) DayWindow(day time.Time) (time.Time, time.Time) {
	return New(q.Store, q.Zone).Window(day)
}
func (q *Query) StateCounts(ctx context.Context, portID string) map[string]int {
	return q.Store.CountArtifactsByState(ctx, portID)
}
func qRound(v float64) int { return int(v + 0.5) }
func Project(current, delta int) int {
	if current+delta < 0 {
		return 0
	}
	return qRound(float64(current + delta))
}

var _ domain.PageInfo
