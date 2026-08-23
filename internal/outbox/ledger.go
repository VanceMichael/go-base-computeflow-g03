package outbox

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type DeliveryLedger struct{ Seen map[string]string }

func NewLedger() *DeliveryLedger { return &DeliveryLedger{Seen: map[string]string{}} }
func (l *DeliveryLedger) Record(key, response string) error {
	if key == "" {
		return domain.ErrInvalid
	}
	if old, ok := l.Seen[key]; ok && old != response {
		return fmt.Errorf("%w: delivery response changed", domain.ErrConflict)
	}
	l.Seen[key] = response
	return nil
}
func (l *DeliveryLedger) Replay(key string) (string, bool) { v, ok := l.Seen[key]; return v, ok }
