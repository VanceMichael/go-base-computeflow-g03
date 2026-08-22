package vehicle

import (
	"github.com/VanceMichael/harborflow/internal/domain"
	"sort"
)

type QueueItem struct {
	VehicleID  string
	Priority   int
	EnqueuedAt int64
}
type Queue []QueueItem

func (q Queue) Len() int { return len(q) }
func (q Queue) Less(i, j int) bool {
	if q[i].Priority == q[j].Priority {
		return q[i].EnqueuedAt < q[j].EnqueuedAt
	}
	return q[i].Priority > q[j].Priority
}
func (q Queue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func OrderQueue(items []QueueItem) []QueueItem {
	out := append([]QueueItem(nil), items...)
	sort.Sort(Queue(out))
	return out
}
func AdmissionFor(state domain.VehicleState) Admission {
	switch state {
	case domain.VehicleQueued:
		return Admit("queue", "awaiting risk assessment")
	case domain.VehicleAssessing:
		return Admit("hold", "risk assessment in progress")
	case domain.VehicleAdmitted:
		return Admit("admit", "risk cleared")
	default:
		return Admit("hold", "manual review required")
	}
}
