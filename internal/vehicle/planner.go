package vehicle

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"sort"
)

type Plan struct {
	BatchID         string
	LaneID          string
	VehicleIDs      []string
	ExpectedVersion int
}

func (p Plan) Validate() error {
	if p.BatchID == "" || p.LaneID == "" || len(p.VehicleIDs) == 0 {
		return fmt.Errorf("%w: incomplete vehicle plan", domain.ErrInvalid)
	}
	seen := map[string]struct{}{}
	for _, id := range p.VehicleIDs {
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%w: duplicate vehicle", domain.ErrInvalid)
		}
		seen[id] = struct{}{}
	}
	return nil
}
func StableVehicleOrder(ids []string) []string {
	out := append([]string(nil), ids...)
	sort.Strings(out)
	return out
}

type Admission struct {
	VehicleID string
	Decision  string
	Reason    string
}

func Admit(decision, reason string) Admission { return Admission{Decision: decision, Reason: reason} }
