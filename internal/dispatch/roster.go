package dispatch

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"sort"
)

type RosterEntry struct {
	ResponderID        string
	Skills             []string
	Active             bool
	CurrentAssignments int
}

func Eligible(entries []RosterEntry, skill string) []RosterEntry {
	out := make([]RosterEntry, 0)
	for _, e := range entries {
		if !e.Active || e.CurrentAssignments > 0 {
			continue
		}
		for _, s := range e.Skills {
			if s == skill {
				out = append(out, e)
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ResponderID < out[j].ResponderID })
	return out
}
func RequireResponder(entry RosterEntry) error {
	if !entry.Active || entry.CurrentAssignments > 0 {
		return fmt.Errorf("%w: responder unavailable", domain.ErrConflict)
	}
	return nil
}
