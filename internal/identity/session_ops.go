package identity

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"time"
)

type SessionOps struct{ Store *sqlite.Store }

func NewSessionOps(s *sqlite.Store) *SessionOps { return &SessionOps{Store: s} }
func (o *SessionOps) Purge(ctx context.Context, now time.Time) (int64, error) {
	return o.Store.PurgeExpiredSessions(ctx, now)
}
func (o *SessionOps) UserCanOperate(u domain.User) bool {
	return u.Active && (u.Role == domain.RoleCoordinator || u.Role == domain.RoleInspector || u.Role == domain.RoleDispatcher)
}
func (o *SessionOps) UserCanAudit(u domain.User) bool {
	return u.Active && (u.Role == domain.RoleCoordinator || u.Role == domain.RoleAuditor)
}
