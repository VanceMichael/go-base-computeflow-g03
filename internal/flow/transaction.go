package flow

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
)

type Transaction struct {
	Store     *sqlite.Store
	Actor     string
	RequestID string
}

func NewTransaction(s *sqlite.Store, actor, requestID string) *Transaction {
	return &Transaction{Store: s, Actor: actor, RequestID: requestID}
}
func (t *Transaction) ValidateRelease(cmd ReleaseCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}
	if cmd.ActorID != t.Actor {
		return fmt.Errorf("%w: actor mismatch", domain.ErrUnauthorized)
	}
	return nil
}
func (t *Transaction) Run(ctx context.Context, cmd ReleaseCommand, w domain.PassengerWave, passengers []domain.Passenger) error {
	if err := t.ValidateRelease(cmd); err != nil {
		return err
	}
	return t.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		ok, err := t.Store.UpdateWaveState(ctx, tx, w.ID, string(w.State), string(domain.WaveReleasing), w.Version, nil)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrConflict
		}
		if len(passengers) != len(cmd.Documents) {
			return fmt.Errorf("%w: passenger set changed", domain.ErrConflict)
		}
		for _, p := range passengers {
			ok, err = t.Store.UpdatePassengerState(ctx, tx, p.ID, string(domain.PassengerWaiting), string(domain.PassengerChecking), p.Version)
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
func (t *Transaction) WithSavepoint(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	return t.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SAVEPOINT computeflow_operation`); err != nil {
			return err
		}
		if err := fn(ctx, tx); err != nil {
			_, _ = tx.ExecContext(ctx, `ROLLBACK TO computeflow_operation`)
			return err
		}
		_, err := tx.ExecContext(ctx, `RELEASE computeflow_operation`)
		return err
	})
}
