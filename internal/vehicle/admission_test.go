package vehicle_test

import (
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/vehicle"
	"testing"
)

func TestAdmissionHoldsVehicleWithOpenIncident(t *testing.T) {
	a, err := vehicle.DecideAdmission(vehicle.RiskInput{VehicleID: "v", IncidentOpen: true}, func(string, float64) (string, error) { t.Fatal("policy should not run"); return "", nil })
	if err != nil || vehicle.AllowedToEnter(a) {
		t.Fatalf("%+v %v", a, err)
	}
}
func TestAdmissionRejectsMissingVehicle(t *testing.T) {
	if _, err := vehicle.DecideAdmission(vehicle.RiskInput{}, func(string, float64) (string, error) { return "admit", nil }); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestQueuePrioritizesHighPriorityAndEarlierItem(t *testing.T) {
	got := vehicle.OrderQueue([]vehicle.QueueItem{{VehicleID: "late", Priority: 1, EnqueuedAt: 2}, {VehicleID: "high", Priority: 2, EnqueuedAt: 3}, {VehicleID: "early", Priority: 1, EnqueuedAt: 1}})
	if got[0].VehicleID != "high" || got[1].VehicleID != "early" {
		t.Fatalf("%v", got)
	}
}
