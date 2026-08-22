package flow

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"strings"
)

type ReleaseCommand struct {
	RunID     string
	WaveID    string
	ActorID   string
	RequestID string
	Documents []string
}

func (c ReleaseCommand) Validate() error {
	if c.RunID == "" || c.WaveID == "" || c.ActorID == "" || c.RequestID == "" {
		return fmt.Errorf("%w: incomplete release command", domain.ErrInvalid)
	}
	if len(c.Documents) == 0 {
		return fmt.Errorf("%w: empty passenger set", domain.ErrInvalid)
	}
	seen := map[string]struct{}{}
	for _, doc := range c.Documents {
		doc = strings.TrimSpace(doc)
		if err := domain.ValidateDocumentKey(doc); err != nil {
			return err
		}
		if _, ok := seen[doc]; ok {
			return fmt.Errorf("%w: duplicate document", domain.ErrInvalid)
		}
		seen[doc] = struct{}{}
	}
	return nil
}

type PauseCommand struct {
	WaveID  string
	ActorID string
	Reason  string
}

func (c PauseCommand) Validate() error {
	if c.WaveID == "" || c.ActorID == "" || strings.TrimSpace(c.Reason) == "" {
		return fmt.Errorf("%w: pause reason is required", domain.ErrInvalid)
	}
	return nil
}
func NormalizeDocuments(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
