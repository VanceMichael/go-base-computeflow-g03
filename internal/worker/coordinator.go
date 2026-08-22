package worker

import (
	"context"
	"sync"
	"time"
)

type Coordinator struct {
	scheduler *Scheduler
	cancel    context.CancelFunc
	once      sync.Once
}

func NewCoordinator(s *Scheduler) *Coordinator { return &Coordinator{scheduler: s} }
func (c *Coordinator) Start(parent context.Context) {
	c.once.Do(func() { ctx, cancel := context.WithCancel(parent); c.cancel = cancel; c.scheduler.Start(ctx) })
}
func (c *Coordinator) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.scheduler != nil {
		c.scheduler.Wait()
	}
}
func DefaultSchedules(jobs map[string]Job) []Schedule {
	out := make([]Schedule, 0, len(jobs))
	for name, job := range jobs {
		out = append(out, Schedule{Name: name, Interval: time.Second, Job: job})
	}
	return out
}
