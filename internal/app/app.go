package app

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/audit"
	"github.com/VanceMichael/computeflow/internal/capacity"
	"github.com/VanceMichael/computeflow/internal/config"
	"github.com/VanceMichael/computeflow/internal/dispatch"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/flow"
	"github.com/VanceMichael/computeflow/internal/gate"
	"github.com/VanceMichael/computeflow/internal/httpapi"
	"github.com/VanceMichael/computeflow/internal/identity"
	"github.com/VanceMichael/computeflow/internal/incident"
	"github.com/VanceMichael/computeflow/internal/middleware"
	"github.com/VanceMichael/computeflow/internal/outbox"
	"github.com/VanceMichael/computeflow/internal/risk"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"github.com/VanceMichael/computeflow/internal/vehicle"
	"github.com/VanceMichael/computeflow/internal/worker"
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
	u := domain.User{ID: uuid.NewString(), PortID: p.ID, Email: "coordinator@computeflow.local", DisplayName: "Compute Market Coordinator", Role: domain.RoleCoordinator, Active: true, CreatedAt: s.Now()}
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
	go func() { a.logger.Info("computeflow listening", "addr", addr); errCh <- server.ListenAndServe() }()
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
		cancel()
		a.Workers.Wait()
		if closeErr := a.Store.Close(); closeErr != nil {
			return fmt.Errorf("serve: %v; close store: %w", err, closeErr)
		}
		return err
	}
}
