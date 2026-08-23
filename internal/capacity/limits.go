package capacity

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type Limits struct {
	MaxPassengers    int
	MaxVehicles      int
	MaxOpenIncidents int
}

func (l Limits) Validate() error {
	if l.MaxPassengers <= 0 || l.MaxVehicles <= 0 || l.MaxOpenIncidents < 0 {
		return fmt.Errorf("%w: invalid limits", domain.ErrInvalid)
	}
	return nil
}

type Usage struct {
	Passengers    int
	Vehicles      int
	OpenIncidents int
}

func (l Limits) Allows(u Usage) bool {
	return u.Passengers <= l.MaxPassengers && u.Vehicles <= l.MaxVehicles && u.OpenIncidents <= l.MaxOpenIncidents
}
func (l Limits) Remaining(u Usage) Usage {
	return Usage{Passengers: max(0, l.MaxPassengers-u.Passengers), Vehicles: max(0, l.MaxVehicles-u.Vehicles), OpenIncidents: max(0, l.MaxOpenIncidents-u.OpenIncidents)}
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
