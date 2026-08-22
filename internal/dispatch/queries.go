package dispatch

import (
	"github.com/VanceMichael/harborflow/internal/domain"
	"sort"
)

type Workload struct {
	ResponderID string
	State       string
	Assignments int
}

func SortByWorkload(items []Workload) []Workload {
	out := append([]Workload(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Assignments == out[j].Assignments {
			return out[i].ResponderID < out[j].ResponderID
		}
		return out[i].Assignments < out[j].Assignments
	})
	return out
}
func Available(items []Workload) []Workload {
	out := make([]Workload, 0)
	for _, item := range items {
		if item.State == "available" {
			out = append(out, item)
		}
	}
	return out
}
func ValidateWorkload(item Workload) error {
	if item.ResponderID == "" || item.Assignments < 0 {
		return domain.ErrInvalid
	}
	return nil
}
