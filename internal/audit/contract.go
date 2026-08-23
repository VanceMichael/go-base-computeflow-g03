package audit

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"strings"
)

type Contract struct {
	Action      string
	SubjectType string
	SubjectID   string
	Outcome     string
	RequestID   string
}

func (c Contract) Validate() error {
	if strings.TrimSpace(c.Action) == "" || strings.TrimSpace(c.SubjectType) == "" || strings.TrimSpace(c.SubjectID) == "" || strings.TrimSpace(c.RequestID) == "" {
		return fmt.Errorf("%w: incomplete audit contract", domain.ErrInvalid)
	}
	switch c.Outcome {
	case "accepted", "rejected", "failed":
	default:
		return fmt.Errorf("%w: audit outcome", domain.ErrInvalid)
	}
	return nil
}
func (c Contract) Accepted() bool { return c.Outcome == "accepted" }
