package vehicle

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
func (s *Service) CreateBatch(ctx context.Context, runID, manifest string, plates []string, now time.Time) (domain.VehicleBatch, error) {
	b := domain.VehicleBatch{ID: uuid.NewString(), RunID: runID, ManifestKey: manifest, State: string(domain.VehicleQueued), Version: 1, CreatedAt: now}
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.Store.InsertBatch(ctx, tx, b); err != nil {
			return err
		}
		for _, plate := range plates {
			if err := s.Store.InsertVehicle(ctx, tx, domain.Vehicle{ID: uuid.NewString(), BatchID: b.ID, PlateKey: plate, State: domain.VehicleQueued, Version: 1, CreatedAt: now}); err != nil {
				return err
			}
		}
		return nil
	})
	return b, err
}
func (s *Service) AssignLane(ctx context.Context, laneID, vehicleID, owner string, now time.Time) error {
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.AssignLane(ctx, tx, uuid.NewString(), laneID, vehicleID, owner, now.Add(3*time.Minute))
		return e
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: lane unavailable", domain.ErrConflict)
	}
	return nil
}
func (s *Service) CloseLane(ctx context.Context, laneID string) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error { return s.Store.ReleaseLaneAssignments(ctx, tx, laneID) })
}
