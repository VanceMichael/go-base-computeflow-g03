package capacity

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"time"
)

type Service struct {
	Store *sqlite.Store
	Zone  *time.Location
}

func New(s *sqlite.Store, z *time.Location) *Service { return &Service{Store: s, Zone: z} }
func (s *Service) Snapshot(ctx context.Context, runID string, from, to time.Time, now time.Time) (domain.CapacitySnapshot, error) {
	pass, clear, held, err := s.Store.CountRunPassengers(ctx, runID)
	if err != nil {
		return domain.CapacitySnapshot{}, err
	}
	vehicles, err := s.Store.CountRunVehicles(ctx, runID)
	if err != nil {
		return domain.CapacitySnapshot{}, err
	}
	x := domain.CapacitySnapshot{ID: uuid.NewString(), RunID: runID, WindowStart: from, WindowEnd: to, Passengers: pass, Cleared: clear, Held: held, Vehicles: vehicles, State: "complete", CreatedAt: now}
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error { return s.Store.InsertSnapshot(ctx, tx, x) })
	return x, err
}
func (s *Service) Window(day time.Time) (time.Time, time.Time) {
	local := day.In(s.Zone)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, s.Zone)
	return start, start.Add(24 * time.Hour)
}
