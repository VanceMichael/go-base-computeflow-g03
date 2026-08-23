package gate

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
)

type QueryService struct {
	Store    *sqlite.Store
	Protocol Protocol
}

func NewQueryService(s *sqlite.Store) *QueryService {
	return &QueryService{Store: s, Protocol: DefaultProtocol()}
}
func (q *QueryService) CanRelease(ctx context.Context, passengerID string) (bool, error) {
	summary, err := NewSummarizer(q.Store).ForPassenger(ctx, passengerID)
	if err != nil {
		return false, err
	}
	if summary.Stages != q.Protocol.Stages {
		return false, fmt.Errorf("%w: incomplete protocol", domain.ErrConflict)
	}
	return summary.Complete(), nil
}
func (q *QueryService) ProtocolReady() error { return q.Protocol.Validate() }
