package worker

import (
	"context"
	"time"
)

type Check func(context.Context) error

func RunChecks(ctx context.Context, checks ...Check) error {
	for _, check := range checks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if check != nil {
			if err := check(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
func DeadlineContext(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, d)
}
