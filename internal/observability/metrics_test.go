package observability_test

import (
	"github.com/VanceMichael/harborflow/internal/observability"
	"testing"
)

func TestMetricsSnapshotCountsRequestsAndWorkers(t *testing.T) {
	m := observability.New()
	m.Request(true)
	m.Request(false)
	m.WorkerRun()
	s := m.Snapshot()
	if s.Requests != 2 || s.Failures != 1 || s.WorkerRuns != 1 {
		t.Fatalf("%+v", s)
	}
}
