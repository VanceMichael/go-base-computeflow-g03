package flow

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
)

type RunReport struct {
	RunID          string
	PassengerTotal int
	Cleared        int
	Held           int
	Vehicles       int
	OpenIncidents  int
}
type Reporter struct{ Store *sqlite.Store }

func NewReporter(s *sqlite.Store) *Reporter { return &Reporter{Store: s} }
func (r *Reporter) Build(ctx context.Context, run domain.StressRun) (RunReport, error) {
	if run.ID == "" {
		return RunReport{}, fmt.Errorf("%w: run id", domain.ErrInvalid)
	}
	total, cleared, held, err := r.Store.CountRunPassengers(ctx, run.ID)
	if err != nil {
		return RunReport{}, err
	}
	vehicles, err := r.Store.CountRunVehicles(ctx, run.ID)
	if err != nil {
		return RunReport{}, err
	}
	items, totalInc, err := r.Store.ListOpenIncidents(ctx, run.PortID, 200, 0)
	if err != nil {
		return RunReport{}, err
	}
	_ = items
	return RunReport{RunID: run.ID, PassengerTotal: total, Cleared: cleared, Held: held, Vehicles: vehicles, OpenIncidents: totalInc}, nil
}
func (r RunReport) CompletionReady() bool {
	return r.PassengerTotal == r.Cleared && r.Held == 0 && r.OpenIncidents == 0
}
