package worker

import (
	"context"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"time"
)

type Maintenance struct{ Store *sqlite.Store }

func NewMaintenance(s *sqlite.Store) *Maintenance { return &Maintenance{Store: s} }
func (m *Maintenance) ExpireSessions(ctx context.Context, now time.Time) error {
	_, err := m.Store.PurgeExpiredSessions(ctx, now)
	return err
}
func (m *Maintenance) ExpireResourceLeases(ctx context.Context, now time.Time) error {
	_, err := m.Store.ExpireLeases(ctx, now)
	return err
}
func (m *Maintenance) ArchiveIncidents(ctx context.Context, portID string, before time.Time) error {
	_, err := m.Store.ArchiveClosedIncidents(ctx, portID, before)
	return err
}
