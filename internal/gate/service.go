package gate

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"time"
)

type Service struct{ Store *sqlite.Store }

func New(s *sqlite.Store) *Service { return &Service{Store: s} }
func (s *Service) Prepare(ctx context.Context, p domain.Passenger, gates []domain.Gate, now time.Time) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return s.Store.CreateJointScans(ctx, tx, p.ID, gates, now)
	})
}
func (s *Service) Claim(ctx context.Context, scanID, owner string, version int, now time.Time) (string, error) {
	token := uuid.NewString()
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.ClaimScan(ctx, tx, scanID, owner, token, now.Add(2*time.Minute), version)
		return e
	})
	if err != nil {
		return "", err
	}
	if !ok {
		return "", domain.ErrConflict
	}
	return token, nil
}
func (s *Service) Complete(ctx context.Context, scanID, owner, token string, version int, cleared bool) error {
	state := string(domain.ScanCleared)
	if !cleared {
		state = string(domain.ScanRejected)
	}
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.CompleteScan(ctx, tx, scanID, owner, token, state, version)
		return e
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: scan ownership", domain.ErrConflict)
	}
	return nil
}
