package incident

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"io"
)

type EvidenceWriter interface {
	Write([]byte) (int, error)
	Close() error
}
type Archiver struct {
	Open func(context.Context, string) (EvidenceWriter, error)
}

func (a Archiver) ArchiveEvidence(ctx context.Context, keys []string, payload []byte) error {
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		w, err := a.Open(ctx, key)
		if err != nil {
			return err
		}
		_, writeErr := w.Write(payload)
		closeErr := w.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
func ValidateEvidenceState(state domain.IncidentState) error {
	if state != domain.IncidentResolved {
		return fmt.Errorf("%w: evidence needs resolved incident", domain.ErrInvalid)
	}
	return nil
}

var _ io.Writer
