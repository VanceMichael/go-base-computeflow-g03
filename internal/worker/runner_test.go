package worker_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/harborflow/internal/worker"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunnerStopsJobsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	r := worker.New(func(context.Context) error { calls.Add(1); return nil })
	r.Run(ctx)
	time.Sleep(250 * time.Millisecond)
	cancel()
	r.Wait()
	if calls.Load() == 0 {
		t.Fatal("job never ran")
	}
}
func TestRunnerRunOnceReturnsJobError(t *testing.T) {
	want := errors.New("failure")
	err := worker.New().RunOnce(context.Background(), func(context.Context) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("%v", err)
	}
}
func TestBackoffGrowsAndCaps(t *testing.T) {
	b := worker.Backoff{Initial: time.Second, Max: 4 * time.Second}
	if b.Delay(1) != time.Second || b.Delay(3) != 4*time.Second || b.Delay(8) != 4*time.Second {
		t.Fatalf("%v %v %v", b.Delay(1), b.Delay(3), b.Delay(8))
	}
}
func TestRetryPolicyStopsAtMaximum(t *testing.T) {
	p := worker.RetryPolicy{MaxAttempts: 3}
	if !p.Retry(2) || p.Retry(3) {
		t.Fatal("retry boundary incorrect")
	}
}
