package config_test

import (
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/config"
)

func TestLoadParsesEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/z")
	t.Setenv("CORS_ORIGINS", "https://a.com,https://b.com")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("SHIPPING_CONFIG_PATH", "/custom/path/shipping.json")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want production", cfg.Env)
	}
	if cfg.DatabaseURL != "postgres://x:y@localhost:5432/z" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[0] != "https://a.com" || cfg.CORSOrigins[1] != "https://b.com" {
		t.Errorf("CORSOrigins = %v", cfg.CORSOrigins)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q", cfg.LogLevel)
	}
	if cfg.ShippingConfigPath != "/custom/path/shipping.json" {
		t.Errorf("ShippingConfigPath = %q, want /custom/path/shipping.json", cfg.ShippingConfigPath)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	// envconfig treats empty as zero-value unless `required:"true"` is set.
	t.Setenv("DATABASE_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty")
	}
}

func TestLoadParsesAuthEnvVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/z")
	t.Setenv("SESSION_COOKIE_NAME", "my_session")
	t.Setenv("SESSION_COOKIE_DOMAIN", "example.com")
	t.Setenv("GOOGLE_CLIENT_ID", "gid123")
	t.Setenv("GOOGLE_CLIENT_SECRET", "gsecret456")
	t.Setenv("GOOGLE_REDIRECT_URL", "https://example.com/callback")
	t.Setenv("EMAIL_ALLOWLIST", "alice@example.com,bob@example.com")
	t.Setenv("FRONTEND_BASE_URL", "https://app.example.com")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SessionCookieName != "my_session" {
		t.Errorf("SessionCookieName = %q, want my_session", cfg.SessionCookieName)
	}
	if cfg.SessionCookieDomain != "example.com" {
		t.Errorf("SessionCookieDomain = %q, want example.com", cfg.SessionCookieDomain)
	}
	if cfg.GoogleClientID != "gid123" {
		t.Errorf("GoogleClientID = %q, want gid123", cfg.GoogleClientID)
	}
	if cfg.GoogleClientSecret != "gsecret456" {
		t.Errorf("GoogleClientSecret = %q, want gsecret456", cfg.GoogleClientSecret)
	}
	if cfg.GoogleRedirectURL != "https://example.com/callback" {
		t.Errorf("GoogleRedirectURL = %q, want https://example.com/callback", cfg.GoogleRedirectURL)
	}
	if len(cfg.EmailAllowlist) != 2 || cfg.EmailAllowlist[0] != "alice@example.com" || cfg.EmailAllowlist[1] != "bob@example.com" {
		t.Errorf("EmailAllowlist = %v", cfg.EmailAllowlist)
	}
	if cfg.FrontendBaseURL != "https://app.example.com" {
		t.Errorf("FrontendBaseURL = %q, want https://app.example.com", cfg.FrontendBaseURL)
	}
}
