package app

import (
	"context"
	"log/slog"

	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

type Application struct {
	Config   *config.Config
	Pool     db.Pool
	Logger   *slog.Logger
	Shipping *shipping.Service
}

func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	logger := NewLogger(cfg.LogLevel, cfg.Env)
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	shipCfg, err := shipping.LoadConfig(cfg.ShippingConfigPath)
	if err != nil {
		pool.Close()
		return nil, err
	}
	ship := shipping.New(shipCfg)
	return &Application{Config: cfg, Pool: pool, Logger: logger, Shipping: ship}, nil
}

func (a *Application) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
}
