package worker

import (
	"sync/atomic"
	"time"
)

type Heartbeat struct {
	last     atomic.Int64
	failures atomic.Int64
}

func NewHeartbeat() *Heartbeat          { return &Heartbeat{} }
func (h *Heartbeat) Beat(now time.Time) { h.last.Store(now.UnixNano()) }
func (h *Heartbeat) Fail()              { h.failures.Add(1) }
func (h *Heartbeat) Healthy(now time.Time, ttl time.Duration) bool {
	last := h.last.Load()
	return last > 0 && now.Sub(time.Unix(0, last)) <= ttl
}
func (h *Heartbeat) Failures() uint64 { return uint64(h.failures.Load()) }
