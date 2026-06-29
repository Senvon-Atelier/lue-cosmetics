package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/cart"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/orders"
	"github.com/oti-adjei/ruecosmetics/internal/payments/paystack"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

type Application struct {
	Config   *config.Config
	Pool     db.Pool
	Logger   *zap.Logger
	Shipping *shipping.Service
	Auth     *auth.Service
	Cart     *cart.Service
	Email    email.Sender
	Orders   *orders.Service
	Paystack *paystack.Client
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

	// Email Sender chain: LogSender (default) → ResendSender (if configured) →
	// AllowlistSender (always the outermost wrapper).
	renderer, rerr := email.NewRenderer()
	if rerr != nil {
		pool.Close()
		return nil, rerr
	}
	var inner email.Sender = email.LogSender{Log: logger}
	if resendSender, sendErr := email.NewResendSender(cfg.ResendAPIKey, cfg.ResendFromEmail, renderer, logger); sendErr == nil {
		inner = resendSender
	}
	mailSender := email.NewAllowlistSender(inner, cfg.EmailAllowlist, logger)

	catalogRepo := catalog.NewRepository(pool)

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, mailSender, cfg.EmailAllowlist)
	cartSvc := cart.NewService(cart.NewRepository(pool), catalogRepo, ship, logger)

	paystackClient := paystack.NewClient(cfg.PaystackBaseURL, cfg.PaystackSecretKey)
	ordersSvc := orders.NewService(
		orders.NewRepository(pool),
		cartSvc, catalogRepo, ship, paystackClient, mailSender, pool, logger,
		cfg.PaystackCallbackURL,
	)

	return &Application{
		Config:   cfg,
		Pool:     pool,
		Logger:   logger,
		Shipping: ship,
		Auth:     authSvc,
		Cart:     cartSvc,
		Email:    mailSender,
		Orders:   ordersSvc,
		Paystack: paystackClient,
	}, nil
}

func (a *Application) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
}
