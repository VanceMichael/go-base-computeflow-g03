package gate

import (
	"context"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
)

type Summary struct {
	PassengerID string
	Stages      int
	Cleared     int
	Rejected    int
	Pending     int
}

func (s Summary) Complete() bool {
	return s.Stages > 0 && s.Cleared == s.Stages && s.Rejected == 0 && s.Pending == 0
}

type Summarizer struct{ Store *sqlite.Store }

func NewSummarizer(s *sqlite.Store) *Summarizer { return &Summarizer{Store: s} }
func (s *Summarizer) ForPassenger(ctx context.Context, id string) (Summary, error) {
	rows, err := s.Store.DB.QueryContext(ctx, `SELECT state,COUNT(*) FROM gate_scans WHERE passenger_id=? GROUP BY state`, id)
	if err != nil {
		return Summary{}, err
	}
	defer rows.Close()
	out := Summary{PassengerID: id}
	for rows.Next() {
		var state string
		var n int
		if err := rows.Scan(&state, &n); err != nil {
			return out, err
		}
		out.Stages += n
		switch domain.ScanState(state) {
		case domain.ScanCleared:
			out.Cleared += n
		case domain.ScanRejected:
			out.Rejected += n
		default:
			out.Pending += n
		}
	}
	if out.Stages == 0 {
		return out, fmt.Errorf("%w: no scans", domain.ErrNotFound)
	}
	return out, rows.Err()
}
