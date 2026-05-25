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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shophub-project-2026/shophub/internal/auth"
	"github.com/shophub-project-2026/shophub/internal/config"
	"github.com/shophub-project-2026/shophub/internal/db"
	"github.com/shophub-project-2026/shophub/internal/server"
	"github.com/shophub-project-2026/shophub/internal/server/middleware"
	"github.com/shophub-project-2026/shophub/internal/shops"
	"github.com/shophub-project-2026/shophub/internal/ui"
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

	scheme := runtime.NewScheme()
	if err := shops.AddToScheme(scheme); err != nil {
		slog.Error("register k8s scheme", "err", err)
		os.Exit(1)
	}

	var restCfg *rest.Config
	if cfg.KubeConfig != "" {
		restCfg, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	} else {
		restCfg, err = ctrl.GetConfig()
	}
	if err != nil {
		slog.Error("build k8s config", "err", err)
		os.Exit(1)
	}

	k8sClient, err := client.New(restCfg, client.Options{Scheme: scheme})
	if err != nil {
		slog.Error("create k8s client", "err", err)
		os.Exit(1)
	}

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	jwtMiddleware := middleware.JWT(authSvc.TokenParserFn())

	shopsRepo := shops.NewRepository(k8sClient, pool)
	shopsHandler := shops.NewHandler(shopsRepo)

	srv := server.New(cfg.HTTPAddr, cfg.HTTPPort)

	srv.HandleFunc("POST /auth/register", authHandler.Register)
	srv.HandleFunc("POST /auth/login", authHandler.Login)

	uiHandler := ui.NewHandler(authSvc, shopsRepo)
	srv.HandleFunc("GET /login", uiHandler.LoginPage)
	srv.HandleFunc("POST /login", uiHandler.LoginPost)
	srv.HandleFunc("GET /register", uiHandler.RegisterPage)
	srv.HandleFunc("POST /register", uiHandler.RegisterPost)
	srv.HandleFunc("GET /", uiHandler.Root)
	srv.HandleFunc("GET /logout", uiHandler.Logout)

	uiJwt := middleware.JWTOrRedirect(authSvc.TokenParserFn(), "/login")
	srv.Handle("GET /dashboard", uiJwt(http.HandlerFunc(uiHandler.Dashboard)))
	srv.Handle("GET /shops/new", uiJwt(http.HandlerFunc(uiHandler.ShopNew)))
	srv.Handle("POST /shops/new", uiJwt(http.HandlerFunc(uiHandler.ShopNewPost)))
	srv.Handle("GET /shops/{name}", uiJwt(http.HandlerFunc(uiHandler.ShopDetail)))
	srv.Handle("GET /shops/{name}/edit", uiJwt(http.HandlerFunc(uiHandler.ShopEdit)))
	srv.Handle("POST /shops/{name}/edit", uiJwt(http.HandlerFunc(uiHandler.ShopEditPost)))

	srv.Handle("GET /shops", jwtMiddleware(http.HandlerFunc(shopsHandler.List)))
	srv.Handle("POST /shops", jwtMiddleware(http.HandlerFunc(shopsHandler.Create)))
	srv.Handle("PUT /shops/{name}", jwtMiddleware(http.HandlerFunc(shopsHandler.Update)))
	srv.Handle("DELETE /shops/{name}", jwtMiddleware(http.HandlerFunc(shopsHandler.Delete)))

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
