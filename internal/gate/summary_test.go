package gate_test

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/gate"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"testing"
)

func TestGateSummaryRequiresEveryStage(t *testing.T) {
	f := testsupport.New(t)
	p := f.Passenger("D")
	g := f.Gate(1)
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		if err := f.Store.InsertScan(ctx, tx, domain.GateScan{ID: "scan", PassengerID: p.ID, GateID: g.ID, Stage: 1, State: "cleared", Version: 1, CreatedAt: f.Now}); err != nil {
			f.T.Fatal(err)
		}
	})
	if _, err := gate.NewSummarizer(f.Store).ForPassenger(context.Background(), p.ID); err != nil {
		t.Fatal(err)
	}
}
func TestDefaultProtocolRequiresThreeChecks(t *testing.T) {
	if err := gate.DefaultProtocol().Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestProtocolRejectsUnacceptedStage(t *testing.T) {
	p := gate.DefaultProtocol()
	if got, _ := p.Next(2, false); got != "rejected" {
		t.Fatalf("%s", got)
	}
}
