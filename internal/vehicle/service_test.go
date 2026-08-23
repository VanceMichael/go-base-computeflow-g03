package vehicle_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"github.com/VanceMichael/computeflow/internal/vehicle"
	"testing"
)

func TestCreateBatchPersistsAllVehicles(t *testing.T) {
	f := testsupport.New(t)
	run := domain.StressRun{ID: "run", PortID: f.Port.ID, State: domain.RunRunning}
	if err := f.Store.InsertRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	b, err := vehicle.New(f.Store).CreateBatch(context.Background(), run.ID, "manifest-1", []string{"A", "B", "C"}, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := f.Store.DB.QueryRow(`SELECT COUNT(*) FROM vehicles WHERE batch_id=?`, b.ID).Scan(&n); err != nil || n != 3 {
		t.Fatalf("%d %v", n, err)
	}
}
func TestAssignLaneRequiresOpenLane(t *testing.T) {
	f := testsupport.New(t)
	lane := f.Lane(1)
	v := f.Vehicle("vehicle")
	if _, err := f.Store.DB.ExecContext(context.Background(), `UPDATE lanes SET state='closed' WHERE id=?`, lane.ID); err != nil {
		t.Fatal(err)
	}
	if err := vehicle.New(f.Store).AssignLane(context.Background(), lane.ID, v.ID, "dispatcher", f.Now); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
func TestAssignLaneCreatesOneOwnership(t *testing.T) {
	f := testsupport.New(t)
	lane := f.Lane(1)
	vehicleA := f.Vehicle("vehicle-a")
	if err := vehicle.New(f.Store).AssignLane(context.Background(), lane.ID, vehicleA.ID, "dispatcher", f.Now); err != nil {
		t.Fatal(err)
	}
	vehicleB := f.Vehicle("vehicle-b")
	if err := vehicle.New(f.Store).AssignLane(context.Background(), lane.ID, vehicleB.ID, "dispatcher", f.Now); err == nil {
		t.Fatal("second assignment succeeded")
	}
}
func TestCloseLaneReleasesActiveAssignments(t *testing.T) {
	f := testsupport.New(t)
	lane := f.Lane(1)
	vehicleA := f.Vehicle("vehicle-a")
	if err := vehicle.New(f.Store).AssignLane(context.Background(), lane.ID, vehicleA.ID, "dispatcher", f.Now); err != nil {
		t.Fatal(err)
	}
	if err := vehicle.New(f.Store).CloseLane(context.Background(), lane.ID); err != nil {
		t.Fatal(err)
	}
	n, err := f.Store.CountActiveVehicles(context.Background(), lane.ID)
	if err != nil || n != 0 {
		t.Fatalf("%d %v", n, err)
	}
}
func TestPlannerRejectsDuplicateVehicle(t *testing.T) {
	p := vehicle.Plan{BatchID: "b", LaneID: "l", VehicleIDs: []string{"v", "v"}}
	if err := p.Validate(); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestPlannerSortsVehicleIDsDeterministically(t *testing.T) {
	got := vehicle.StableVehicleOrder([]string{"v3", "v1", "v2"})
	if got[0] != "v1" || got[2] != "v3" {
		t.Fatalf("%v", got)
	}
}
