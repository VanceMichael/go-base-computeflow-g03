package gate

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type ProtocolState struct {
	Stage       int
	Document    bool
	Face        bool
	Fingerprint bool
}

func (s ProtocolState) Accepts(stage int) bool {
	switch stage {
	case 1:
		return s.Document
	case 2:
		return s.Document && s.Face
	case 3:
		return s.Document && s.Face && s.Fingerprint
	default:
		return false
	}
}
func (s ProtocolState) Advance(stage int) ProtocolState {
	out := s
	switch stage {
	case 1:
		out.Document = true
	case 2:
		out.Face = true
	case 3:
		out.Fingerprint = true
	}
	if stage > out.Stage {
		out.Stage = stage
	}
	return out
}
func (s ProtocolState) Validate() error {
	if s.Stage < 0 || s.Stage > 3 {
		return fmt.Errorf("%w: stage", domain.ErrInvalid)
	}
	if s.Stage >= 1 && !s.Document {
		return fmt.Errorf("%w: document missing", domain.ErrInvalid)
	}
	if s.Stage >= 2 && !s.Face {
		return fmt.Errorf("%w: face missing", domain.ErrInvalid)
	}
	if s.Stage >= 3 && !s.Fingerprint {
		return fmt.Errorf("%w: fingerprint missing", domain.ErrInvalid)
	}
	return nil
}
