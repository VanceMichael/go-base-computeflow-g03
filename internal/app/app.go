package app

import (
	"context"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/audit"
	"github.com/VanceMichael/harborflow/internal/capacity"
	"github.com/VanceMichael/harborflow/internal/config"
	"github.com/VanceMichael/harborflow/internal/dispatch"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/flow"
	"github.com/VanceMichael/harborflow/internal/gate"
	"github.com/VanceMichael/harborflow/internal/httpapi"
	"github.com/VanceMichael/harborflow/internal/identity"
	"github.com/VanceMichael/harborflow/internal/incident"
	"github.com/VanceMichael/harborflow/internal/middleware"
	"github.com/VanceMichael/harborflow/internal/outbox"
	"github.com/VanceMichael/harborflow/internal/risk"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"github.com/VanceMichael/harborflow/internal/vehicle"
	"github.com/VanceMichael/harborflow/internal/worker"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	Config  config.Config
	Store   *sqlite.Store
	HTTP    http.Handler
	Workers *worker.Runner
	logger  *slog.Logger
}

func New(ctx context.Context, c config.Config, logger *slog.Logger) (*App, error) {
	s, err := sqlite.Open(c.DatabasePath)
	if err != nil {
		return nil, err
	}
	if err := seed(ctx, s); err != nil {
		s.Close()
		return nil, err
	}
	o := outbox.New(s)
	_ = audit.New(s)
	_ = dispatch.New(s)
	_ = incident.New(s)
	_ = gate.New(s)
	_ = vehicle.New(s)
	_ = capacity.New(s, c.BusinessZone)
	_ = risk.New(risk.StaticVerifier{})
	f := flow.New(s)
	i := identity.New(s, c.SessionTTL)
	api := httpapi.New(s, i, f)
	recovery := worker.NewRecovery(o)
	w := worker.New(func(ctx context.Context) error {
		_, err := recovery.RequeueExpiredOutbox(ctx, time.Now().UTC())
		return err
	})
	return &App{Config: c, Store: s, HTTP: middleware.Chain(api.Handler(), logger), Workers: w, logger: logger}, nil
}
func seed(ctx context.Context, s *sqlite.Store) error {
	p, err := s.EnsureDemoPort(ctx)
	if err != nil {
		return err
	}
	u := domain.User{ID: uuid.NewString(), PortID: p.ID, Email: "coordinator@harborflow.local", DisplayName: "Joint Operations Coordinator", Role: domain.RoleCoordinator, Active: true, CreatedAt: s.Now()}
	if _, err = s.DB.ExecContext(ctx, `INSERT OR IGNORE INTO users(id,port_id,email,display_name,role,active,created_at) VALUES(?,?,?,?,?,?,?)`, u.ID, u.PortID, u.Email, u.DisplayName, string(u.Role), 1, stamp(u.CreatedAt)); err != nil {
		return err
	}
	for n := 1; n <= 3; n++ {
		_, err = s.DB.ExecContext(ctx, `INSERT OR IGNORE INTO gates(id,port_id,gate_no,mode,active) VALUES(?,?,?,?,1)`, uuid.NewString(), p.ID, n, "cooperative")
		if err != nil {
			return err
		}
	}
	for n := 1; n <= 4; n++ {
		_, err = s.DB.ExecContext(ctx, `INSERT OR IGNORE INTO lanes(id,port_id,lane_no,state,version) VALUES(?,?,?,?,1)`, uuid.NewString(), p.ID, n, "open")
		if err != nil {
			return err
		}
	}
	return nil
}
func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
func (a *App) Serve(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", a.Config.Port)
	server := &http.Server{Addr: addr, Handler: a.HTTP}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	a.Workers.Run(workerCtx)
	errCh := make(chan error, 1)
	go func() { a.logger.Info("harborflow listening", "addr", addr); errCh <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, stop := context.WithTimeout(context.Background(), a.Config.ShutdownGrace)
		defer stop()
		_ = server.Shutdown(shutdownCtx)
		cancel()
		a.Workers.Wait()
		return a.Store.Close()
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
