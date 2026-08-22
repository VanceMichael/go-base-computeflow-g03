package worker

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"sync"
	"time"
)

type RecoveryState struct {
	mu          sync.Mutex
	RunStates   map[string]domain.RunState
	LastAttempt map[string]time.Time
	Attempts    map[string]int
}

func NewRecoveryState() *RecoveryState {
	return &RecoveryState{RunStates: map[string]domain.RunState{}, LastAttempt: map[string]time.Time{}, Attempts: map[string]int{}}
}
func (r *RecoveryState) Begin(runID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if runID == "" {
		return fmt.Errorf("%w: run id", domain.ErrInvalid)
	}
	if r.RunStates[runID] == domain.RunCompleted {
		return fmt.Errorf("%w: run already complete", domain.ErrConflict)
	}
	r.Attempts[runID]++
	r.LastAttempt[runID] = now
	return nil
}
func (r *RecoveryState) Mark(runID string, state domain.RunState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RunStates[runID] = state
}
func (r *RecoveryState) Retryable(runID string, max int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Attempts[runID] < max && r.RunStates[runID] != domain.RunCompleted
}
