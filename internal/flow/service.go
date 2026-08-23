package flow

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
func (s *Service) CreateRun(ctx context.Context, portID, name, actor string, now time.Time) (domain.StressRun, error) {
	r := domain.StressRun{ID: uuid.NewString(), PortID: portID, Name: name, State: domain.RunDraft, Version: 1, StartsAt: now, CreatedBy: actor, CreatedAt: now}
	return r, s.Store.InsertRun(ctx, r)
}
func (s *Service) StartRun(ctx context.Context, r domain.StressRun, now time.Time) error {
	if err := domain.TransitionRun(r.State, domain.RunRunning); err != nil {
		return err
	}
	ok, err := s.Store.WithRunUpdate(ctx, r.ID, r.PortID, string(r.State), string(domain.RunRunning), r.Version, now)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrConflict
	}
	return nil
}
func (s *Service) CreateWave(ctx context.Context, r domain.StressRun, sequence int, documents []string, now time.Time) (domain.PassengerWave, error) {
	if r.State != domain.RunRunning {
		return domain.PassengerWave{}, fmt.Errorf("%w: run must be running", domain.ErrInvalid)
	}
	w := domain.PassengerWave{ID: uuid.NewString(), RunID: r.ID, SequenceNo: sequence, State: domain.WavePlanned, Version: 1, PlannedAt: now}
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.Store.InsertWave(ctx, tx, w); err != nil {
			return err
		}
		for _, doc := range documents {
			p := domain.Passenger{ID: uuid.NewString(), WaveID: w.ID, DocumentKey: doc, State: domain.PassengerWaiting, Version: 1, CreatedAt: now}
			if err := s.Store.InsertPassenger(ctx, tx, p); err != nil {
				return err
			}
		}
		return nil
	})
	return w, err
}
func (s *Service) ReleaseWave(ctx context.Context, w domain.PassengerWave, passengers []domain.Passenger, now time.Time) error {
	if err := domain.TransitionWave(w.State, domain.WaveReleasing); err != nil {
		return err
	}
	return s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		ok, err := s.Store.UpdateWaveState(ctx, tx, w.ID, string(w.State), string(domain.WaveReleasing), w.Version, nil)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrConflict
		}
		for _, p := range passengers {
			ok, err := s.Store.UpdatePassengerState(ctx, tx, p.ID, string(domain.PassengerWaiting), string(domain.PassengerChecking), p.Version)
			if err != nil {
				return err
			}
			if !ok {
				return domain.ErrConflict
			}
		}
		return nil
	})
}
func (s *Service) CompleteRun(ctx context.Context, r domain.StressRun, now time.Time) error {
	if err := domain.TransitionRun(r.State, domain.RunCompleted); err != nil {
		return err
	}
	ok, err := s.Store.WithRunUpdate(ctx, r.ID, r.PortID, string(r.State), string(domain.RunCompleted), r.Version, now)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrConflict
	}
	return nil
}
