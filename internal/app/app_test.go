package app_test

import (
	"context"
	"github.com/VanceMichael/harborflow/internal/app"
	"github.com/VanceMichael/harborflow/internal/config"
	"log/slog"
	"testing"
	"time"
)

func TestAppSeedsOperationalPortAndReadiness(t *testing.T) {
	cfg := config.Config{Port: 0, DatabasePath: t.TempDir() + "/app.db", BusinessZone: time.UTC, SessionTTL: time.Hour, ShutdownGrace: time.Second}
	a, err := app.New(context.Background(), cfg, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Store.Close()
	if err := a.Store.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := a.Store.DB.QueryRow(`SELECT COUNT(*) FROM ports`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("%d %v", n, err)
	}
}
