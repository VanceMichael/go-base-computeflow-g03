package dispatch

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"github.com/google/uuid"
)

type Service struct{ Store *sqlite.Store }

func New(s *sqlite.Store) *Service { return &Service{Store: s} }
func (s *Service) Claim(ctx context.Context, incidentID, responderID, owner string) error {
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.AssignResponder(ctx, tx, uuid.NewString(), incidentID, responderID, owner)
		if e != nil {
			return e
		}
		if !ok {
			return domain.ErrConflict
		}
		ok, e = s.Store.SetResponderState(ctx, tx, responderID, "available", "busy", 1)
		if e != nil {
			return e
		}
		if !ok {
			return domain.ErrConflict
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: responder unavailable", domain.ErrConflict)
	}
	return nil
}
func (s *Service) Release(ctx context.Context, responderID string, version int) error {
	var ok bool
	err := s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var e error
		ok, e = s.Store.SetResponderState(ctx, tx, responderID, "busy", "available", version)
		return e
	})
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrConflict
	}
	return nil
}
