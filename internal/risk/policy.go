package risk

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"strings"
)

type Policy struct {
	BlockedInstitutions map[string]bool
	RequireManualReview bool
}

func (p Policy) Evaluate(institution string, confidence float64) (string, string, error) {
	institution = strings.TrimSpace(institution)
	if institution == "" {
		return "", "", fmt.Errorf("%w: institution", domain.ErrInvalid)
	}
	if p.BlockedInstitutions[institution] {
		return "held", "blocked institution", nil
	}
	if confidence < 0 || confidence > 1 {
		return "", "", fmt.Errorf("%w: confidence", domain.ErrInvalid)
	}
	if p.RequireManualReview && confidence < 0.9 {
		return "held", "manual review", nil
	}
	return "cleared", "accepted", nil
}
func (p Policy) Merge(other Policy) Policy {
	out := Policy{BlockedInstitutions: map[string]bool{}, RequireManualReview: p.RequireManualReview || other.RequireManualReview}
	for k, v := range p.BlockedInstitutions {
		out.BlockedInstitutions[k] = v
	}
	for k, v := range other.BlockedInstitutions {
		out.BlockedInstitutions[k] = v
	}
	return out
}
