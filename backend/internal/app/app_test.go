package app_test

import (
	"context"
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

	cfg := &config.Config{
		Port: 0, Env: "development",
		DatabaseURL: url,
		CORSOrigins: []string{"http://localhost:5173"},
		LogLevel:    "debug",
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
}
