package gate

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type Protocol struct {
	Stages             int
	RequireDocument    bool
	RequireFace        bool
	RequireFingerprint bool
}

func DefaultProtocol() Protocol {
	return Protocol{Stages: 3, RequireDocument: true, RequireFace: true, RequireFingerprint: true}
}
func (p Protocol) Validate() error {
	if p.Stages != 3 || !p.RequireDocument || !p.RequireFace || !p.RequireFingerprint {
		return fmt.Errorf("%w: joint protocol requires three checks", domain.ErrInvalid)
	}
	return nil
}
func (p Protocol) Next(stage int, accepted bool) (domain.ScanState, error) {
	if stage < 1 || stage > p.Stages {
		return domain.ScanRejected, domain.ErrInvalid
	}
	if !accepted {
		return domain.ScanRejected, nil
	}
	if stage == p.Stages {
		return domain.ScanCleared, nil
	}
	return domain.ScanPending, nil
}
