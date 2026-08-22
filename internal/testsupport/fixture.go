package testsupport

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"testing"
	"time"
)

type Fixture struct {
	T     *testing.T
	Store *sqlite.Store
	Port  domain.Port
	User  domain.User
	Now   time.Time
}

func New(t *testing.T) *Fixture {
	t.Helper()
	s, err := sqlite.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	s.SetClock(func() time.Time { return now })
	p := domain.Port{ID: uuid.NewString(), Code: "TEST", Name: "Test Port", Timezone: "Asia/Shanghai", CreatedAt: now}
	if err := s.InsertPort(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	u := domain.User{ID: uuid.NewString(), PortID: p.ID, Email: "coordinator@test.local", DisplayName: "Test Coordinator", Role: domain.RoleCoordinator, Active: true, CreatedAt: now}
	if err := s.InsertUser(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	return &Fixture{T: t, Store: s, Port: p, User: u, Now: now}
}
func (f *Fixture) Tx(fn func(context.Context, *sql.Tx)) {
	f.T.Helper()
	if err := f.Store.WithTx(context.Background(), func(ctx context.Context, tx *sql.Tx) error { fn(ctx, tx); return nil }); err != nil {
		f.T.Fatal(err)
	}
}
func (f *Fixture) Gate(no int) domain.Gate {
	g := domain.Gate{ID: uuid.NewString(), PortID: f.Port.ID, GateNo: no, Mode: "cooperative", Active: true}
	if err := f.Store.InsertGate(context.Background(), g); err != nil {
		f.T.Fatal(err)
	}
	return g
}
func (f *Fixture) Lane(no int) domain.Lane {
	l := domain.Lane{ID: uuid.NewString(), PortID: f.Port.ID, LaneNo: no, State: domain.LaneOpen, Version: 1}
	if err := f.Store.InsertLane(context.Background(), l); err != nil {
		f.T.Fatal(err)
	}
	return l
}
func (f *Fixture) Responder(name string) domain.Responder {
	r := domain.Responder{ID: uuid.NewString(), PortID: f.Port.ID, Name: name, State: "available", Version: 1}
	if err := f.Store.InsertResponder(context.Background(), r); err != nil {
		f.T.Fatal(err)
	}
	return r
}

func (f *Fixture) Passenger(key string) domain.Passenger {
	run := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, Name: "fixture", State: domain.RunRunning, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), run); err != nil {
		f.T.Fatal(err)
	}
	wave := domain.PassengerWave{ID: uuid.NewString(), RunID: run.ID, SequenceNo: 1, State: domain.WavePlanned, Version: 1, PlannedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertWave(ctx, tx, wave); err != nil {
			f.T.Fatal(err)
		}
	})
	p := domain.Passenger{ID: uuid.NewString(), WaveID: wave.ID, DocumentKey: key, State: domain.PassengerChecking, Version: 1, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertPassenger(ctx, tx, p); err != nil {
			f.T.Fatal(err)
		}
	})
	return p
}

func (f *Fixture) Vehicle(key string) domain.Vehicle {
	run := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, Name: "fixture", State: domain.RunRunning, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), run); err != nil {
		f.T.Fatal(err)
	}
	batch := domain.VehicleBatch{ID: uuid.NewString(), RunID: run.ID, ManifestKey: uuid.NewString(), State: string(domain.VehicleQueued), Version: 1, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertBatch(ctx, tx, batch); err != nil {
			f.T.Fatal(err)
		}
	})
	v := domain.Vehicle{ID: uuid.NewString(), BatchID: batch.ID, PlateKey: key, State: domain.VehicleQueued, Version: 1, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertVehicle(ctx, tx, v); err != nil {
			f.T.Fatal(err)
		}
	})
	return v
}
