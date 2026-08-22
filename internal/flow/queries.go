package flow

import (
	"context"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"time"
)

type QueryService struct{ Store *sqlite.Store }

func NewQueryService(s *sqlite.Store) *QueryService { return &QueryService{Store: s} }
func (q *QueryService) Passenger(ctx context.Context, id, portID string) (domain.Passenger, error) {
	p, err := q.Store.FindPassenger(ctx, id, portID)
	if err != nil {
		return p, fmt.Errorf("passenger lookup: %w", err)
	}
	return p, nil
}
func (q *QueryService) Run(ctx context.Context, id, portID string) (sqlite.RunSummary, error) {
	summary, err := q.Store.RunSummary(ctx, id, portID)
	if err != nil {
		return summary, fmt.Errorf("run summary: %w", err)
	}
	return summary, nil
}

type WaveWindow struct {
	From time.Time
	To   time.Time
}

func (w WaveWindow) Valid() bool { return !w.From.IsZero() && !w.To.IsZero() && w.From.Before(w.To) }
func (q *QueryService) ActivePassengers(ctx context.Context, waveID string) ([]domain.Passenger, error) {
	items, err := q.Store.ListPassengers(ctx, waveID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Passenger, 0, len(items))
	for _, p := range items {
		if domain.IsActivePassenger(p.State) {
			out = append(out, p)
		}
	}
	return out, nil
}
