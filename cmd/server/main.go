package main

import (
	"context"
	"github.com/VanceMichael/harborflow/internal/app"
	"github.com/VanceMichael/harborflow/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config failed", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	a, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("app failed", "error", err)
		os.Exit(1)
	}
	if err := a.Serve(ctx); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
