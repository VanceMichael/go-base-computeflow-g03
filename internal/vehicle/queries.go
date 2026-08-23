package vehicle

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
)

type QueryService struct{ Store *sqlite.Store }

func NewQueryService(s *sqlite.Store) *QueryService { return &QueryService{Store: s} }
func (q *QueryService) Vehicle(ctx context.Context, id, portID string) (domain.Vehicle, error) {
	v, err := q.Store.FindVehicle(ctx, id, portID)
	if err != nil {
		return v, fmt.Errorf("vehicle lookup: %w", err)
	}
	return v, nil
}
func (q *QueryService) LaneCapacity(ctx context.Context, portID string) (int, error) {
	return q.Store.OpenLaneCount(ctx, portID)
}
func (q *QueryService) SortForDispatch(items []QueueItem) []QueueItem { return OrderQueue(items) }
