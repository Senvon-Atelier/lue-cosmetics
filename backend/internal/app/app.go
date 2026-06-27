package app

import (
	"context"
	"log/slog"

	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/db"
)

type Application struct {
	Config *config.Config
	Pool   db.Pool
	Logger *slog.Logger
}

func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	logger := NewLogger(cfg.LogLevel, cfg.Env)
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &Application{Config: cfg, Pool: pool, Logger: logger}, nil
}

func (a *Application) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
}
