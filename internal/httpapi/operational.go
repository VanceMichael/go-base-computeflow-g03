package httpapi

import (
	"context"
	"encoding/json"
	"github.com/VanceMichael/harborflow/internal/capacity"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/identity"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"net/http"
	"time"
)

type Operational struct {
	Store    *sqlite.Store
	Identity *identity.Service
	Capacity *capacity.Service
}

func NewOperational(s *sqlite.Store, i *identity.Service, c *capacity.Service) *Operational {
	return &Operational{Store: s, Identity: i, Capacity: c}
}
func (o *Operational) Metrics(ctx context.Context, portID string) (sqlite.PortMetrics, error) {
	return o.Store.PortMetrics(ctx, portID)
}
func (o *Operational) CanAudit(ctx context.Context, token, portID string) bool {
	u, err := o.Identity.Authenticate(ctx, token, time.Now().UTC())
	return err == nil && u.PortID == portID && (u.Role == domain.RoleAuditor || u.Role == domain.RoleCoordinator)
}
func (o *Operational) SnapshotJSON(x domain.CapacitySnapshot) ([]byte, error) {
	return json.Marshal(struct {
		RunID      string `json:"run_id"`
		Passengers int    `json:"passengers"`
		Cleared    int    `json:"cleared"`
		Held       int    `json:"held"`
		Vehicles   int    `json:"vehicles"`
		State      string `json:"state"`
	}{x.RunID, x.Passengers, x.Cleared, x.Held, x.Vehicles, x.State})
}
func noContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }
