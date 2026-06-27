package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func TestApplicationNewWiresPool(t *testing.T) {
	ctx := context.Background()
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	// Write a valid shipping config to a temp dir so app.New doesn't need the
	// file to exist at the relative default path (which doesn't resolve from
	// the test's working directory).
	tmpDir := t.TempDir()
	shipPath := filepath.Join(tmpDir, "shipping_config.json")
	if err := os.WriteFile(shipPath, []byte(`{"flat_rate_ghs_minor":2500,"free_over_ghs_minor":50000}`), 0644); err != nil {
		t.Fatalf("write shipping config: %v", err)
	}

	cfg := &config.Config{
		Port:               0,
		Env:                "development",
		DatabaseURL:        url,
		CORSOrigins:        []string{"http://localhost:5173"},
		LogLevel:           "debug",
		ShippingConfigPath: shipPath,
	}
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a, err := app.New(c, cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Close()
	if a.Pool == nil || a.Logger == nil || a.Config == nil {
		t.Errorf("nil field on Application")
	}
	if a.Shipping == nil {
		t.Error("nil Shipping")
	}
	if a.Auth == nil {
		t.Error("nil Auth")
	}
	if a.Email == nil {
		t.Error("nil Email")
	}
}
