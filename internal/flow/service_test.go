package flow_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/flow"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"testing"
)

func runFixture(t *testing.T) (*testsupport.Fixture, domain.StressRun) {
	f := testsupport.New(t)
	s := flow.New(f.Store)
	r, err := s.CreateRun(context.Background(), f.Port.ID, "pressure-test", f.User.ID, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.StartRun(context.Background(), r, f.Now); err != nil {
		t.Fatal(err)
	}
	r.State = domain.RunRunning
	r.Version = 2
	return f, r
}
func TestCreateRunStartsInDraft(t *testing.T) {
	f := testsupport.New(t)
	r, err := flow.New(f.Store).CreateRun(context.Background(), f.Port.ID, "morning", f.User.ID, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if r.State != domain.RunDraft || r.Version != 1 {
		t.Fatalf("%+v", r)
	}
}
func TestStartRunRejectsStaleVersion(t *testing.T) {
	f := testsupport.New(t)
	r, err := flow.New(f.Store).CreateRun(context.Background(), f.Port.ID, "stale", f.User.ID, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	r.Version = 9
	if err := flow.New(f.Store).StartRun(context.Background(), r, f.Now); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
func TestCreateWavePersistsParticipantsTogether(t *testing.T) {
	f, r := runFixture(t)
	w, err := flow.New(f.Store).CreateWave(context.Background(), r, 1, []string{"HK-1", "HK-2", "HK-3"}, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	items, err := f.Store.ListPassengers(context.Background(), w.ID)
	if err != nil || len(items) != 3 {
		t.Fatalf("%d %v", len(items), err)
	}
}
func TestCreateWaveRejectsPausedRun(t *testing.T) {
	f := testsupport.New(t)
	r := domain.StressRun{ID: "r", PortID: f.Port.ID, State: domain.RunPaused}
	if _, err := flow.New(f.Store).CreateWave(context.Background(), r, 1, []string{"x"}, f.Now); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestReleaseWaveTransitionsAllPassengers(t *testing.T) {
	f, r := runFixture(t)
	w, err := flow.New(f.Store).CreateWave(context.Background(), r, 1, []string{"A", "B"}, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	passengers, err := f.Store.ListPassengers(context.Background(), w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := flow.New(f.Store).ReleaseWave(context.Background(), w, passengers, f.Now); err != nil {
		t.Fatal(err)
	}
	after, err := f.Store.ListPassengers(context.Background(), w.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range after {
		if p.State != domain.PassengerChecking {
			t.Fatalf("%+v", p)
		}
	}
}
func TestReleaseWaveRejectsAlreadyReleasedVersion(t *testing.T) {
	f, r := runFixture(t)
	w, err := flow.New(f.Store).CreateWave(context.Background(), r, 1, []string{"A"}, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	passengers, _ := f.Store.ListPassengers(context.Background(), w.ID)
	if err := flow.New(f.Store).ReleaseWave(context.Background(), w, passengers, f.Now); err != nil {
		t.Fatal(err)
	}
	if err := flow.New(f.Store).ReleaseWave(context.Background(), w, passengers, f.Now); err == nil {
		t.Fatal("second release succeeded")
	}
}
func TestCompleteRunAcceptsRunningState(t *testing.T) {
	f, r := runFixture(t)
	if err := flow.New(f.Store).CompleteRun(context.Background(), r, f.Now); err != nil {
		t.Fatal(err)
	}
}
func TestCompleteRunRejectsDraftState(t *testing.T) {
	f := testsupport.New(t)
	r := domain.StressRun{ID: "r", PortID: f.Port.ID, State: domain.RunDraft, Version: 1}
	if err := flow.New(f.Store).CompleteRun(context.Background(), r, f.Now); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
