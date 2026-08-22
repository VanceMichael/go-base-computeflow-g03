package app

import (
	"context"
	"sync"
	"time"
)

type Lifecycle struct {
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func NewLifecycle(parent context.Context) (*Lifecycle, context.Context) {
	ctx, cancel := context.WithCancel(parent)
	return &Lifecycle{cancel: cancel, done: make(chan struct{})}, ctx
}
func (l *Lifecycle) Stop() { l.once.Do(func() { l.cancel(); close(l.done) }) }
func (l *Lifecycle) Wait(timeout time.Duration) bool {
	if timeout <= 0 {
		<-l.done
		return true
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-l.done:
		return true
	case <-timer.C:
		return false
	}
}
