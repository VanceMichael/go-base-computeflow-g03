package flow

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"time"
)

type Importer struct{ Store *sqlite.Store }
type Participant struct {
	DocumentKey string
	Institution string
}
type ImportResult struct {
	WaveID  string
	Created int
	Reused  bool
}

func NewImporter(s *sqlite.Store) *Importer { return &Importer{Store: s} }
func (i *Importer) ImportParticipants(ctx context.Context, run domain.StressRun, batchKey string, participants []Participant, now time.Time) (ImportResult, error) {
	if run.State != domain.RunRunning {
		return ImportResult{}, fmt.Errorf("%w: run state", domain.ErrInvalid)
	}
	if batchKey == "" {
		return ImportResult{}, fmt.Errorf("%w: batch key", domain.ErrInvalid)
	}
	w := domain.PassengerWave{ID: uuid.NewString(), RunID: run.ID, SequenceNo: 1, State: domain.WavePlanned, Version: 1, PlannedAt: now}
	err := i.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := i.Store.InsertWave(ctx, tx, w); err != nil {
			if sqlite.IsConstraint(err) {
				return domain.ErrConflict
			}
			return err
		}
		for _, p := range participants {
			if err := domain.ValidateDocumentKey(p.DocumentKey); err != nil {
				return err
			}
			if p.Institution == "" {
				return fmt.Errorf("%w: institution", domain.ErrInvalid)
			}
			if err := i.Store.InsertPassenger(ctx, tx, domain.Passenger{ID: uuid.NewString(), WaveID: w.ID, DocumentKey: p.DocumentKey, State: domain.PassengerWaiting, Version: 1, CreatedAt: now}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ImportResult{}, err
	}
	return ImportResult{WaveID: w.ID, Created: len(participants)}, nil
}
