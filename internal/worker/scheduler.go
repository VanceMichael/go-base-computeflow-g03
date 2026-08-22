package worker

import (
	"context"
	"sync"
	"time"
)

type Schedule struct {
	Name     string
	Interval time.Duration
	Job      Job
}
type Scheduler struct {
	schedules []Schedule
	wg        sync.WaitGroup
}

func NewScheduler(schedules ...Schedule) *Scheduler { return &Scheduler{schedules: schedules} }
func (s *Scheduler) Start(ctx context.Context) {
	for _, spec := range s.schedules {
		spec := spec
		if spec.Interval <= 0 || spec.Job == nil {
			continue
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(spec.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					_ = spec.Job(ctx)
				}
			}
		}()
	}
}
func (s *Scheduler) Wait() { s.wg.Wait() }
func (s *Scheduler) RunNow(ctx context.Context, name string) error {
	for _, spec := range s.schedules {
		if spec.Name == name && spec.Job != nil {
			return spec.Job(ctx)
		}
	}
	return context.Canceled
}
