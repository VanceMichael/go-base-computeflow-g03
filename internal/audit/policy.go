package audit

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type Policy struct {
	RequireRequestID bool
	AllowedActions   map[string]bool
}

func (p Policy) Validate(action, requestID string) error {
	if action == "" {
		return fmt.Errorf("%w: action", domain.ErrInvalid)
	}
	if p.RequireRequestID && requestID == "" {
		return fmt.Errorf("%w: request id", domain.ErrInvalid)
	}
	if len(p.AllowedActions) > 0 && !p.AllowedActions[action] {
		return fmt.Errorf("%w: action not allowed", domain.ErrUnauthorized)
	}
	return nil
}
func DefaultPolicy() Policy {
	return Policy{RequireRequestID: true, AllowedActions: map[string]bool{"run.created": true, "wave.released": true, "incident.opened": true, "outbox.delivered": true}}
}
