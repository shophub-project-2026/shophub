package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shophub-project-2026/shophub/internal/config"
	"github.com/shophub-project-2026/shophub/internal/server"
)

func main() {
	cfg := config.Load()

	srv := server.New(cfg.HTTPAddr, cfg.HTTPPort)

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.HTTPAddr, cfg.HTTPPort)
		slog.Info("server starting", "addr", addr)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
