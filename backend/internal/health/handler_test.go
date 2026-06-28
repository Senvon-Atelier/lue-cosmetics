package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func TestHealthOK(t *testing.T) {
	ctx := context.Background()
	url, stop := testsupport.StartPostgres(t)
	defer stop()
	cfg := &config.Config{
		Env: "development", DatabaseURL: url,
		CORSOrigins:        []string{"http://localhost:5173"},
		LogLevel:           "debug",
		ShippingConfigPath: testsupport.WriteShippingConfig(t, 2500, 50000),
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
		ShippingConfigPath: testsupport.WriteShippingConfig(t, 2500, 50000),
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
