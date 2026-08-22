package observability

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	requests    atomic.Uint64
	failures    atomic.Uint64
	workerRuns  atomic.Uint64
	lastRequest atomic.Int64
}

func New() *Metrics { return &Metrics{} }
func (m *Metrics) Request(ok bool) {
	m.requests.Add(1)
	m.lastRequest.Store(time.Now().UnixNano())
	if !ok {
		m.failures.Add(1)
	}
}
func (m *Metrics) WorkerRun() { m.workerRuns.Add(1) }

type Snapshot struct {
	Requests, Failures, WorkerRuns uint64
	LastRequest                    time.Time
}

func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{Requests: m.requests.Load(), Failures: m.failures.Load(), WorkerRuns: m.workerRuns.Load(), LastRequest: time.Unix(0, m.lastRequest.Load())}
}
