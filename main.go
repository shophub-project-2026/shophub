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

	"github.com/shophub-project-2026/shophub/internal/auth"
	"github.com/shophub-project-2026/shophub/internal/config"
	"github.com/shophub-project-2026/shophub/internal/db"
	"github.com/shophub-project-2026/shophub/internal/server"
	"github.com/shophub-project-2026/shophub/internal/server/middleware"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		slog.Error("connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("run migrations", "err", err)
		os.Exit(1)
	}

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	jwtMiddleware := middleware.JWT(authSvc.TokenParserFn())

	srv := server.New(cfg.HTTPAddr, cfg.HTTPPort)

	srv.HandleFunc("POST /auth/register", authHandler.Register)
	srv.HandleFunc("POST /auth/login", authHandler.Login)

	_ = jwtMiddleware

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.HTTPAddr, cfg.HTTPPort)
		slog.Info("server starting", "addr", addr)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
