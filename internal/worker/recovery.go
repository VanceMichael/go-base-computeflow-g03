package worker

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/outbox"
	"time"
)

type Recovery struct{ Outbox *outbox.Service }

func NewRecovery(o *outbox.Service) *Recovery { return &Recovery{Outbox: o} }
func (r *Recovery) RequeueExpiredOutbox(ctx context.Context, now time.Time) (int64, error) {
	return r.Outbox.RequeueExpired(ctx, now)
}
