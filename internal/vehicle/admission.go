package vehicle

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"strings"
)

type RiskInput struct {
	VehicleID    string
	Institution  string
	Confidence   float64
	IncidentOpen bool
}

func DecideAdmission(input RiskInput, policy func(string, float64) (string, error)) (Admission, error) {
	if input.VehicleID == "" {
		return Admission{}, fmt.Errorf("%w: vehicle", domain.ErrInvalid)
	}
	if input.IncidentOpen {
		return Admit("hold", "open incident"), nil
	}
	decision, err := policy(input.Institution, input.Confidence)
	if err != nil {
		return Admission{}, err
	}
	if strings.TrimSpace(decision) == "" {
		return Admission{}, fmt.Errorf("%w: empty decision", domain.ErrInvalid)
	}
	return Admit(decision, "risk policy"), nil
}
func AllowedToEnter(a Admission) bool { return a.Decision == "admit" || a.Decision == "cleared" }
