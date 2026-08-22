package worker

import (
	"context"
	"sync"
	"time"
)

type Job func(context.Context) error
type Runner struct {
	jobs []Job
	wg   sync.WaitGroup
	once sync.Once
}

func New(jobs ...Job) *Runner { return &Runner{jobs: jobs} }
func (r *Runner) Run(ctx context.Context) {
	r.once.Do(func() {
		for _, job := range r.jobs {
			j := job
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				ticker := time.NewTicker(200 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						_ = j(ctx)
					}
				}
			}()
		}
	})
}
func (r *Runner) Wait()                                      { r.wg.Wait() }
func (r *Runner) RunOnce(ctx context.Context, job Job) error { return job(ctx) }
