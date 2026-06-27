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
	ShippingConfigPath string   `envconfig:"SHIPPING_CONFIG_PATH" default:"seed/config/shipping_config.json"`
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
