package capacity_test

import (
	"github.com/VanceMichael/computeflow/internal/capacity"
	"github.com/VanceMichael/computeflow/internal/domain"
	"testing"
)

func TestLimitsRejectInvalidConfiguration(t *testing.T) {
	if err := (capacity.Limits{}).Validate(); err == nil {
		t.Fatal("invalid limits accepted")
	}
}
func TestLimitsComputeRemainingCapacity(t *testing.T) {
	l := capacity.Limits{MaxPassengers: 100, MaxVehicles: 10, MaxOpenIncidents: 2}
	r := l.Remaining(capacity.Usage{Passengers: 30, Vehicles: 4, OpenIncidents: 1})
	if r.Passengers != 70 || r.Vehicles != 6 || r.OpenIncidents != 1 {
		t.Fatalf("%+v", r)
	}
}
func TestForecastDetectsOverCapacity(t *testing.T) {
	f := capacity.Forecast{WindowMinutes: 1, ExpectedPassengers: 100, Gates: 1, ServiceSeconds: 60}
	if !f.PeakExceeded() {
		t.Fatal("peak not detected")
	}
}
func TestSnapshotStateCountsRemainDomainIndependent(t *testing.T) {
	if domain.PassengerCleared == domain.PassengerHeld {
		t.Fatal("states must differ")
	}
}
