package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port               int      `envconfig:"PORT" default:"8080"`
	Env                string   `envconfig:"ENV" default:"development"`
	DatabaseURL        string   `envconfig:"DATABASE_URL" required:"true"`
	CORSOrigins        []string `envconfig:"CORS_ORIGINS" default:"http://localhost:5173"`
	LogLevel           string   `envconfig:"LOG_LEVEL" default:"info"`
	ShippingConfigPath   string   `envconfig:"SHIPPING_CONFIG_PATH" default:"seed/config/shipping_config.json"`
	SessionCookieName    string   `envconfig:"SESSION_COOKIE_NAME" default:"rue_session"`
	SessionCookieDomain  string   `envconfig:"SESSION_COOKIE_DOMAIN" default:""`
	GoogleClientID       string   `envconfig:"GOOGLE_CLIENT_ID" default:""`
	GoogleClientSecret   string   `envconfig:"GOOGLE_CLIENT_SECRET" default:""`
	GoogleRedirectURL    string   `envconfig:"GOOGLE_REDIRECT_URL" default:"http://localhost:8080/api/v1/auth/google/callback"`
	EmailAllowlist       []string `envconfig:"EMAIL_ALLOWLIST" default:""`
	FrontendBaseURL      string   `envconfig:"FRONTEND_BASE_URL" default:"http://localhost:5173"`
	PaystackSecretKey    string   `envconfig:"PAYSTACK_SECRET_KEY" default:""`
	PaystackBaseURL      string   `envconfig:"PAYSTACK_BASE_URL" default:"https://api.paystack.co"`
	PaystackCallbackURL  string   `envconfig:"PAYSTACK_CALLBACK_URL" default:"http://localhost:5173/checkout/return"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}
	return &cfg, nil
}
