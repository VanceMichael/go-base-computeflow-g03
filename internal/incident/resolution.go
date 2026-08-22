package incident

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"strings"
)

type Resolution struct {
	ResponderID  string
	Summary      string
	EvidenceKeys []string
	ResolvedAt   string
}

func (r Resolution) Validate() error {
	if r.ResponderID == "" || strings.TrimSpace(r.Summary) == "" || r.ResolvedAt == "" {
		return fmt.Errorf("%w: incomplete resolution", domain.ErrInvalid)
	}
	if len(r.EvidenceKeys) == 0 {
		return fmt.Errorf("%w: evidence required", domain.ErrInvalid)
	}
	return nil
}
func CanClose(state domain.IncidentState, resolution Resolution) error {
	if state != domain.IncidentResolved {
		return fmt.Errorf("%w: incident must be resolved", domain.ErrInvalid)
	}
	return resolution.Validate()
}
func NormalizeSummary(v string) string { return strings.Join(strings.Fields(v), " ") }
