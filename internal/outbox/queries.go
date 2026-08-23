package outbox

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
)

type Query struct{ Store *sqlite.Store }

func NewQuery(s *sqlite.Store) *Query { return &Query{Store: s} }
func (q *Query) Get(ctx context.Context, id, portID string) (domain.OutboxMessage, error) {
	m, err := q.Store.FindOutbox(ctx, id, portID)
	if err != nil {
		return m, fmt.Errorf("outbox lookup: %w", err)
	}
	return m, nil
}
func CanRetry(m domain.OutboxMessage, max int) bool {
	return m.State == string(domain.OutboxPending) && m.Attempts < max
}
func IsTerminal(m domain.OutboxMessage) bool {
	return m.State == string(domain.OutboxDelivered) || m.State == string(domain.OutboxDead)
}
