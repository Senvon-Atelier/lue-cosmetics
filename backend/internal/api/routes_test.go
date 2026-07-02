package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/api"
	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// TestRoutesMountsExpectedPaths boots the real router against a test DB and
// asserts representative routes exist (non-404) with correct gating.
func TestRoutesMountsExpectedPaths(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	t.Cleanup(stop)
	testsupport.Migrate(t, url, "../../migrations")

	cfg := &config.Config{
		Env:                "test",
		DatabaseURL:        url,
		LogLevel:           "error",
		ShippingConfigPath: testsupport.WriteShippingConfig(t, 2000, 30000),
		SessionCookieName:  "rue_session",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	a, err := app.New(ctx, cfg)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	t.Cleanup(a.Close)

	h := api.New(a)

	cases := []struct {
		method, path string
		wantStatus   int
	}{
		{http.MethodGet, "/healthz", http.StatusOK},
		// Public catalog: admin routes now live under /admin, so /products is no
		// longer shadowed and returns the public catalog (200).
		{http.MethodGet, "/api/v1/products", http.StatusOK},
		{http.MethodGet, "/api/v1/categories", http.StatusOK},
		{http.MethodGet, "/api/v1/cart", http.StatusOK},
		{http.MethodGet, "/api/v1/me", http.StatusUnauthorized}, // auth-gated (user session required)
		// /api/v1/orders no longer exists (admin routes moved to /api/v1/admin/...).
		{http.MethodGet, "/api/v1/orders", http.StatusNotFound},
		// Admin routes live under /api/v1/admin — gated by RequireSession + RequireRole("admin").
		{http.MethodGet, "/api/v1/admin/orders", http.StatusUnauthorized},
		{http.MethodPost, "/api/v1/webhooks/paystack", http.StatusUnauthorized}, // public webhook, missing/invalid signature → 401
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.wantStatus {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, rec.Code, tc.wantStatus)
		}
	}
}
