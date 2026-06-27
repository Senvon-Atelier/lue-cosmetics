package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/me"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

// @title           Rue Cosmetics API
// @version         0.1.0
// @description     E-commerce backend for the Rue Cosmetics case study.
// @BasePath        /api/v1
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer a.Close()

	r := chi.NewRouter()
	r.Use(httpx.Recovery(a.Logger))
	r.Use(httpx.RequestID)
	r.Use(httpx.CORS(cfg.CORSOrigins))

	// /healthz stays at the root for uptime monitoring.
	r.Get("/healthz", health.Handler(a))

	// All public + future protected APIs mount under /api/v1.
	catalogHandlers := catalog.NewHandlers(catalog.NewRepository(a.Pool))
	shippingHandlers := shipping.NewHandlers(a.Shipping)
	r.Route("/api/v1", func(api chi.Router) {
		catalogHandlers.Mount(api)
		shippingHandlers.Mount(api)

		secure := cfg.Env != "development"
		authHandlers := auth.NewHandlers(a.Auth, cfg.SessionCookieName, cfg.SessionCookieDomain, secure)
		authHandlers.Mount(api) // public: /auth/signup, /auth/login, /auth/logout, /auth/session

		// Auth-gated routes (one Group with RequireSession middleware)
		api.Group(func(r chi.Router) {
			r.Use(authHandlers.RequireSession)
			me.NewHandlers().Mount(r) // GET /me
		})
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		a.Logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		a.Logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, scancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer scancel()
	return srv.Shutdown(shutdownCtx)
}
