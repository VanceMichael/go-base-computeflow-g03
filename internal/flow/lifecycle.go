package flow

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"time"
)

type Lifecycle struct {
	State     domain.RunState
	Version   int
	UpdatedAt time.Time
}

func (l Lifecycle) Start(now time.Time) (Lifecycle, error) {
	if err := domain.TransitionRun(l.State, domain.RunRunning); err != nil {
		return l, err
	}
	l.State = domain.RunRunning
	l.Version++
	l.UpdatedAt = now
	return l, nil
}
func (l Lifecycle) Pause(now time.Time) (Lifecycle, error) {
	if err := domain.TransitionRun(l.State, domain.RunPaused); err != nil {
		return l, err
	}
	l.State = domain.RunPaused
	l.Version++
	l.UpdatedAt = now
	return l, nil
}
func (l Lifecycle) Resume(now time.Time) (Lifecycle, error) {
	if err := domain.TransitionRun(l.State, domain.RunRunning); err != nil {
		return l, err
	}
	l.State = domain.RunRunning
	l.Version++
	l.UpdatedAt = now
	return l, nil
}
func (l Lifecycle) Complete(now time.Time) (Lifecycle, error) {
	if err := domain.TransitionRun(l.State, domain.RunCompleted); err != nil {
		return l, err
	}
	l.State = domain.RunCompleted
	l.Version++
	l.UpdatedAt = now
	return l, nil
}
func (l Lifecycle) AssertVersion(version int) error {
	if l.Version != version {
		return fmt.Errorf("%w: expected version %d got %d", domain.ErrConflict, version, l.Version)
	}
	return nil
}
