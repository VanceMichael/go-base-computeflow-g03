package audit

import (
	"context"
	"encoding/json"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

type Export struct {
	Events []domain.AuditEvent
	From   time.Time
	To     time.Time
	PortID string
}

func (e Export) Validate() error {
	if e.PortID == "" || e.From.IsZero() || e.To.IsZero() || !e.From.Before(e.To) {
		return domain.ErrInvalid
	}
	return nil
}
func (e Export) JSON() ([]byte, error) {
	return json.Marshal(struct {
		PortID string              `json:"port_id"`
		From   time.Time           `json:"from"`
		To     time.Time           `json:"to"`
		Events []domain.AuditEvent `json:"events"`
	}{e.PortID, e.From, e.To, e.Events})
}

type Exporter struct{ Service *Service }

func (x Exporter) Build(ctx context.Context, port string, from, to time.Time) (Export, error) {
	events, err := x.Service.Timeline(ctx, port, from, to)
	return Export{Events: events, From: from, To: to, PortID: port}, err
}
