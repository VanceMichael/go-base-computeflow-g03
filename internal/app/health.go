package app

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/domain"
	"time"
)

type Health struct {
	Status    string
	Database  string
	CheckedAt time.Time
}

func (a *App) Health(ctx context.Context) Health {
	h := Health{Status: "ok", Database: "ok", CheckedAt: time.Now().UTC()}
	if err := a.Store.Ping(ctx); err != nil {
		h.Status = "degraded"
		h.Database = err.Error()
	}
	return h
}
func (a *App) DemoPort(ctx context.Context) (domain.Port, error) { return a.Store.EnsureDemoPort(ctx) }
