package gate_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/gate"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"github.com/google/uuid"
	"testing"
)

func TestPrepareCreatesThreeCooperativeScans(t *testing.T) {
	f := testsupport.New(t)
	gates := []domain.Gate{f.Gate(1), f.Gate(2), f.Gate(3)}
	p := f.Passenger("D")
	if err := gate.New(f.Store).Prepare(context.Background(), p, gates, f.Now); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := f.Store.DB.QueryRow(`SELECT COUNT(*) FROM gate_scans WHERE passenger_id=?`, p.ID).Scan(&n); err != nil || n != 3 {
		t.Fatalf("%d %v", n, err)
	}
}
func TestClaimAndCompleteScanUsesLeaseOwner(t *testing.T) {
	f := testsupport.New(t)
	g := f.Gate(1)
	p := f.Passenger("D")
	scan := domain.GateScan{ID: uuid.NewString(), PassengerID: p.ID, GateID: g.ID, Stage: 1, State: domain.ScanPending, Version: 1, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertScan(ctx, tx, scan); err != nil {
			f.T.Fatal(err)
		}
	})
	s := gate.New(f.Store)
	token, err := s.Claim(context.Background(), scan.ID, "worker-a", 1, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Complete(context.Background(), scan.ID, "worker-b", token, 2, true); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("wrong owner result=%v", err)
	}
	if err := s.Complete(context.Background(), scan.ID, "worker-a", token, 2, true); err != nil {
		t.Fatal(err)
	}
}
func TestCompleteRejectedScanIsVisibleAsRejected(t *testing.T) {
	f := testsupport.New(t)
	g := f.Gate(1)
	p := f.Passenger("D")
	scan := domain.GateScan{ID: uuid.NewString(), PassengerID: p.ID, GateID: g.ID, Stage: 1, State: domain.ScanPending, Version: 1, CreatedAt: f.Now}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertScan(ctx, tx, scan); err != nil {
			f.T.Fatal(err)
		}
	})
	s := gate.New(f.Store)
	token, err := s.Claim(context.Background(), scan.ID, "worker", 1, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Complete(context.Background(), scan.ID, "worker", token, 2, false); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := f.Store.DB.QueryRow(`SELECT state FROM gate_scans WHERE id=?`, scan.ID).Scan(&state); err != nil || state != "rejected" {
		t.Fatalf("%s %v", state, err)
	}
}
