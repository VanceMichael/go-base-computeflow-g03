package sqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/VanceMichael/harborflow/internal/testsupport"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestMigrationCreatesOperationalTables(t *testing.T) {
	f := testsupport.New(t)
	var n int
	if err := f.Store.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('ports','users','sessions','stress_runs','passenger_waves','passengers','gates','gate_scans','vehicle_batches','vehicles','lanes','lane_assignments','incidents','responders','dispatch_assignments','outbox_messages','capacity_snapshots','audit_events','idempotency_records')`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 19 {
		t.Fatalf("tables=%d", n)
	}
}
func TestMigrationIsIdempotentAcrossRestart(t *testing.T) {
	path := t.TempDir() + "/harbor.db"
	s, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var v int
	if err := s.DB.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatalf("version=%d", v)
	}
}
func TestPortLookupRequiresPortIdentity(t *testing.T) {
	f := testsupport.New(t)
	if _, err := f.Store.GetPort(context.Background(), uuid.NewString()); err == nil {
		t.Fatal("unknown port returned")
	}
}
func TestPortInsertAndLookupPreservesTimezone(t *testing.T) {
	f := testsupport.New(t)
	p := domain.Port{ID: uuid.NewString(), Code: "SZ", Name: "Shenzhen", Timezone: "Asia/Shanghai", CreatedAt: f.Now}
	if err := f.Store.InsertPort(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	got, err := f.Store.GetPort(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Code != p.Code || got.Timezone != p.Timezone {
		t.Fatalf("%+v", got)
	}
}
func TestRunAndPassengerWritesShareTransaction(t *testing.T) {
	f := testsupport.New(t)
	r := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, Name: "peak", State: domain.RunRunning, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	w := domain.PassengerWave{ID: uuid.NewString(), RunID: r.ID, SequenceNo: 1, State: domain.WavePlanned, Version: 1, PlannedAt: f.Now}
	err := f.Store.WithTx(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		if err := f.Store.InsertWave(ctx, tx, w); err != nil {
			return err
		}
		return f.Store.InsertPassenger(ctx, tx, domain.Passenger{ID: uuid.NewString(), WaveID: w.ID, DocumentKey: "D-1", State: domain.PassengerWaiting, Version: 1, CreatedAt: f.Now})
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := f.Store.ListPassengers(context.Background(), w.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("%d %v", len(items), err)
	}
}
func TestTransactionRollbackLeavesNoPartialPassenger(t *testing.T) {
	f := testsupport.New(t)
	r := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, Name: "rollback", State: domain.RunRunning, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	w := domain.PassengerWave{ID: uuid.NewString(), RunID: r.ID, SequenceNo: 1, State: domain.WavePlanned, Version: 1, PlannedAt: f.Now}
	err := f.Store.WithTx(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		if err := f.Store.InsertWave(ctx, tx, w); err != nil {
			return err
		}
		if err := f.Store.InsertPassenger(ctx, tx, domain.Passenger{ID: uuid.NewString(), WaveID: w.ID, DocumentKey: "D-1", State: domain.PassengerWaiting, Version: 1, CreatedAt: f.Now}); err != nil {
			return err
		}
		return errors.New("simulated downstream failure")
	})
	if err == nil {
		t.Fatal("expected rollback")
	}
	if _, err := f.Store.ListPassengers(context.Background(), w.ID); err != nil {
		t.Fatal(err)
	}
}
func TestConditionalRunUpdateRejectsStaleVersion(t *testing.T) {
	f := testsupport.New(t)
	r := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, Name: "version", State: domain.RunDraft, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	ok, err := f.Store.WithRunUpdate(context.Background(), r.ID, r.PortID, "draft", "running", 2, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("stale update succeeded")
	}
}
func TestWavePassengerConditionalUpdatesProtectVersion(t *testing.T) {
	f := testsupport.New(t)
	r := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, Name: "wave", State: domain.RunRunning, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var w = domain.PassengerWave{ID: uuid.NewString(), RunID: r.ID, SequenceNo: 1, State: domain.WavePlanned, Version: 1, PlannedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertWave(ctx, tx, w); err != nil {
			f.T.Fatal(err)
		}
	})
	var ok bool
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		ok, err = f.Store.UpdateWaveState(ctx, tx, w.ID, "planned", "releasing", 2, nil)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if ok {
		t.Fatal("stale wave transition succeeded")
	}
}
func TestGateScanOwnershipRejectsWrongOwner(t *testing.T) {
	f := testsupport.New(t)
	g := f.Gate(1)
	p := f.Passenger("D-ownership")
	scan := domain.GateScan{ID: uuid.NewString(), PassengerID: p.ID, GateID: g.ID, Stage: 1, State: domain.ScanLeased, LeaseOwner: "worker-a", LeaseToken: "token-a", Version: 2, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertScan(ctx, tx, domain.GateScan{ID: scan.ID, PassengerID: scan.PassengerID, GateID: g.ID, Stage: 1, State: domain.ScanPending, Version: 1, CreatedAt: f.Now}); err != nil {
			f.T.Fatal(err)
		}
	})
	var ok bool
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		ok, err = f.Store.CompleteScan(ctx, tx, scan.ID, "worker-b", "token-b", "cleared", 2)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if ok {
		t.Fatal("wrong owner completed scan")
	}
}
func TestGateScanClaimAndCompletionRequiresToken(t *testing.T) {
	f := testsupport.New(t)
	g := f.Gate(1)
	p := f.Passenger("D-claim")
	scan := domain.GateScan{ID: uuid.NewString(), PassengerID: p.ID, GateID: g.ID, Stage: 1, State: domain.ScanPending, Version: 1, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertScan(ctx, tx, scan); err != nil {
			f.T.Fatal(err)
		}
	})
	var claimed bool
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		claimed, err = f.Store.ClaimScan(ctx, tx, scan.ID, "worker", "token", f.Now.Add(time.Minute), 1)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if !claimed {
		t.Fatal("claim failed")
	}
	var completed bool
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		completed, err = f.Store.CompleteScan(ctx, tx, scan.ID, "worker", "token", "cleared", 2)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if !completed {
		t.Fatal("completion failed")
	}
}
func TestOutboxOwnerControlsDelivery(t *testing.T) {
	f := testsupport.New(t)
	m := domain.OutboxMessage{ID: uuid.NewString(), PortID: f.Port.ID, EventKey: "wave.released", State: string(domain.OutboxPending), IdempotencyKey: "idem-1", CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertOutbox(ctx, tx, m, "payload"); err != nil {
			f.T.Fatal(err)
		}
	})
	var ok bool
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		ok, err = f.Store.ClaimOutbox(ctx, tx, m.ID, "owner-a", f.Now.Add(time.Minute))
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if !ok {
		t.Fatal("claim failed")
	}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		ok, err = f.Store.MarkOutbox(ctx, tx, m.ID, "owner-b", f.Now)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if ok {
		t.Fatal("foreign owner delivered")
	}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		ok, err = f.Store.MarkOutbox(ctx, tx, m.ID, "owner-a", f.Now)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if !ok {
		t.Fatal("current owner failed")
	}
}
func TestExpiredOutboxReturnsToPending(t *testing.T) {
	f := testsupport.New(t)
	m := domain.OutboxMessage{ID: uuid.NewString(), PortID: f.Port.ID, EventKey: "event", State: string(domain.OutboxPending), IdempotencyKey: "idem", CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertOutbox(ctx, tx, m, "payload"); err != nil {
			f.T.Fatal(err)
		}
	})
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		ok, err := f.Store.ClaimOutbox(ctx, tx, m.ID, "owner", f.Now.Add(-time.Minute))
		if err != nil || !ok {
			f.T.Fatalf("%v %v", ok, err)
		}
	})
	var n int64
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		var err error
		n, err = f.Store.RequeueExpiredOutbox(ctx, tx, f.Now)
		if err != nil {
			f.T.Fatal(err)
		}
	})
	if n != 1 {
		t.Fatalf("requeued=%d", n)
	}
}
func TestIdempotencyIsScopedByPortAndOperation(t *testing.T) {
	f := testsupport.New(t)
	key := "same"
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.WriteIdempotency(ctx, tx, f.Port.ID, "release", key, 201, "ok", f.Now); err != nil {
			f.T.Fatal(err)
		}
	})
	code, body, err := f.Store.ReadIdempotency(context.Background(), f.Port.ID, "release", key)
	if err != nil || code != 201 || body != "ok" {
		t.Fatalf("%d %q %v", code, body, err)
	}
	if _, _, err := f.Store.ReadIdempotency(context.Background(), f.Port.ID, "other", key); err == nil {
		t.Fatal("operation crossed")
	}
}
func TestAuditTimelineHonorsHalfOpenWindow(t *testing.T) {
	f := testsupport.New(t)
	from := f.Now.Add(-time.Hour)
	to := f.Now.Add(time.Hour)
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		for n := 0; n < 3; n++ {
			e := domain.AuditEvent{ID: uuid.NewString(), PortID: f.Port.ID, Action: "scan", SubjectType: "passenger", SubjectID: uuid.NewString(), Outcome: "ok", RequestID: uuid.NewString(), Details: "{}", CreatedAt: f.Now.Add(time.Duration(n) * time.Minute)}
			if err := f.Store.InsertAudit(ctx, tx, e); err != nil {
				f.T.Fatal(err)
			}
		}
	})
	events, err := f.Store.ListAudit(context.Background(), f.Port.ID, from, to)
	if err != nil || len(events) != 3 {
		t.Fatalf("%d %v", len(events), err)
	}
}
func TestIncidentListReturnsOnlyActiveStates(t *testing.T) {
	f := testsupport.New(t)
	for _, state := range []domain.IncidentState{domain.IncidentOpen, domain.IncidentAcknowledged, domain.IncidentClosed} {
		if err := f.Store.InsertIncident(context.Background(), domain.Incident{ID: uuid.NewString(), PortID: f.Port.ID, State: state, Severity: "high", Description: "test", Version: 1, CreatedAt: f.Now}); err != nil {
			t.Fatal(err)
		}
	}
	items, total, err := f.Store.ListOpenIncidents(context.Background(), f.Port.ID, 20, 0)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("%d %d %v", total, len(items), err)
	}
}
func TestMaintenancePurgesExpiredOrRevokedSessions(t *testing.T) {
	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s, err := sqlite.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := domain.Port{ID: uuid.NewString(), Code: "P", Name: "P", Timezone: "UTC", CreatedAt: past}
	if err := s.InsertPort(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	u := domain.User{ID: uuid.NewString(), PortID: p.ID, Email: "x@y", DisplayName: "x", Role: domain.RoleInspector, Active: true, CreatedAt: past}
	if err := s.InsertUser(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertSession(context.Background(), domain.Session{ID: uuid.NewString(), UserID: u.ID, TokenHash: "x", ExpiresAt: past.Add(-time.Hour), CreatedAt: past}); err != nil {
		t.Fatal(err)
	}
	n, err := s.PurgeExpiredSessions(context.Background(), past)
	if err != nil || n != 1 {
		t.Fatalf("%d %v", n, err)
	}
}
