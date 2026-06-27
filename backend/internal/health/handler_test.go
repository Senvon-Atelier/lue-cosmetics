package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// writeShippingConfig writes a minimal shipping config to t.TempDir and returns the path.
func writeShippingConfig(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "shipping_config.json")
	if err := os.WriteFile(p, []byte(`{"flat_rate_ghs_minor":2500,"free_over_ghs_minor":50000}`), 0644); err != nil {
		t.Fatalf("write shipping config: %v", err)
	}
	return p
}

func TestHealthOK(t *testing.T) {
	ctx := context.Background()
	url, stop := testsupport.StartPostgres(t)
	defer stop()
	cfg := &config.Config{
		Env: "development", DatabaseURL: url,
		CORSOrigins:        []string{"http://localhost:5173"},
		LogLevel:           "debug",
		ShippingConfigPath: writeShippingConfig(t),
	}
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a, err := app.New(c, cfg)
	if err != nil {
		t.Fatalf("app: %v", err)
	}
	defer a.Close()

	rec := httptest.NewRecorder()
	health.Handler(a)(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct{ Status, DB string }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Status != "ok" || body.DB != "ok" {
		t.Errorf("body = %+v", body)
	}
}

func TestHealthDownReturns503(t *testing.T) {
	// closed pool → ping fails
	ctx := context.Background()
	url, stop := testsupport.StartPostgres(t)
	defer stop()
	cfg := &config.Config{
		Env: "development", DatabaseURL: url,
		CORSOrigins:        []string{"http://localhost:5173"},
		LogLevel:           "debug",
		ShippingConfigPath: writeShippingConfig(t),
	}
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	a, err := app.New(c, cfg)
	if err != nil {
		t.Fatalf("app: %v", err)
	}
	a.Pool.Close()

	rec := httptest.NewRecorder()
	health.Handler(a)(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want 503", rec.Code)
	}
}
