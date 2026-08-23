package outbox

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"sync"
)

type DeliveryLedger struct {
	mu   sync.RWMutex
	Seen map[string]string
}

func NewLedger() *DeliveryLedger { return &DeliveryLedger{Seen: map[string]string{}} }
func (l *DeliveryLedger) Record(key, response string) error {
	if key == "" {
		return domain.ErrInvalid
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if old, ok := l.Seen[key]; ok && old != response {
		return fmt.Errorf("%w: delivery response changed", domain.ErrConflict)
	}
	l.Seen[key] = response
	return nil
}
func (l *DeliveryLedger) Replay(key string) (string, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	v, ok := l.Seen[key]
	return v, ok
}
