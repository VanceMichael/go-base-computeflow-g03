package incident

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"strings"
)

type Triage struct {
	Severity               string
	RequiresMedical        bool
	RequiresTrafficControl bool
	Notes                  string
}

func Classify(description string) Triage {
	d := strings.ToLower(description)
	out := Triage{Severity: "low"}
	if strings.Contains(d, "unwell") || strings.Contains(d, "medical") {
		out.Severity = "high"
		out.RequiresMedical = true
	}
	if strings.Contains(d, "lane") || strings.Contains(d, "vehicle") {
		out.Severity = "medium"
		out.RequiresTrafficControl = true
	}
	out.Notes = description
	return out
}
func ValidateResolution(t Triage, resolution string) error {
	if strings.TrimSpace(resolution) == "" {
		return fmt.Errorf("%w: resolution", domain.ErrInvalid)
	}
	if t.RequiresMedical && !strings.Contains(strings.ToLower(resolution), "medical") {
		return fmt.Errorf("%w: medical evidence", domain.ErrInvalid)
	}
	return nil
}
