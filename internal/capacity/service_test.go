package capacity_test

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/computeflow/internal/capacity"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestSnapshotCountsPassengerStates(t *testing.T) {
	f := testsupport.New(t)
	run := domain.StressRun{ID: uuid.NewString(), PortID: f.Port.ID, State: domain.RunRunning, Version: 1, StartsAt: f.Now, CreatedBy: f.User.ID, CreatedAt: f.Now}
	if err := f.Store.InsertRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	f.Tx(func(ctx context.Context, tx *sql.Tx) {
		w := domain.PassengerWave{ID: uuid.NewString(), RunID: run.ID, SequenceNo: 1, State: domain.WaveReleased, Version: 1, PlannedAt: f.Now}
		if err := f.Store.InsertWave(ctx, tx, w); err != nil {
			f.T.Fatal(err)
		}
		for n, state := range []domain.PassengerState{domain.PassengerCleared, domain.PassengerHeld, domain.PassengerWaiting} {
			if err := f.Store.InsertPassenger(ctx, tx, domain.Passenger{ID: uuid.NewString(), WaveID: w.ID, DocumentKey: string(rune('A' + n)), State: state, Version: 1, CreatedAt: f.Now}); err != nil {
				f.T.Fatal(err)
			}
		}
	})
	x, err := capacity.New(f.Store, time.UTC).Snapshot(context.Background(), run.ID, f.Now.Add(-time.Hour), f.Now, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if x.Passengers != 3 || x.Cleared != 1 || x.Held != 1 {
		t.Fatalf("%+v", x)
	}
}
func TestWindowUsesPortLocalMidnight(t *testing.T) {
	zone, _ := time.LoadLocation("Asia/Shanghai")
	from, to := capacity.New(nil, zone).Window(time.Date(2026, 8, 23, 23, 0, 0, 0, time.UTC))
	if from.Hour() != 0 || to.Sub(from) != 24*time.Hour {
		t.Fatalf("%v %v", from, to)
	}
}
