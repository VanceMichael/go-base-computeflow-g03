package flow

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"sync"
	"time"
)

type TransitionLog struct {
	mu      sync.RWMutex
	entries []domain.OperationalEvent
}

func NewTransitionLog() *TransitionLog {
	return &TransitionLog{entries: make([]domain.OperationalEvent, 0)}
}
func (l *TransitionLog) Append(e domain.OperationalEvent) error {
	if !e.IsScoped() {
		return fmt.Errorf("%w: event scope", domain.ErrInvalid)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, e)
	return nil
}
func (l *TransitionLog) Since(portID string, from time.Time) []domain.OperationalEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]domain.OperationalEvent, 0)
	for _, e := range l.entries {
		if e.PortID == portID && !e.OccurredAt.Before(from) {
			out = append(out, e)
		}
	}
	return out
}
func (l *TransitionLog) Len() int { l.mu.RLock(); defer l.mu.RUnlock(); return len(l.entries) }
