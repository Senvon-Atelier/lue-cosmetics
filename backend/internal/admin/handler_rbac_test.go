package admin_test

// handler_rbac_test.go — RBAC integration matrix for admin routes.
//
// All admin routes require session + admin role. This matrix pins that gating
// for representative endpoints so a future mount refactor cannot silently
// expose them. After Task 4 the routes live under /admin/... as the OpenAPI
// contract declares.
//
// Scope: anonymous → 401, customer session → 403.
// (200-path admin cases are covered in internal/auth/rbac_test.go which
// already has a reusable admin-session helper; we stay in the 401/403 scope
// here to avoid duplicating the testcontainer cost.)

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/admin"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// buildAdminTestRouter wires auth public routes (for signup/login setup) plus
// admin routes mounted under /admin, mirroring the production layout in
// internal/api/routes.go.
func buildAdminTestRouter(t *testing.T) (http.Handler, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")

	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, zap.NewNop(), noopSender{}, nil)
	svc.Params = auth.TestParams

	authHandlers := auth.NewHandlers(svc, "rue_session", "", false)

	// admin.Service can be nil: 401/403 matrix hits middleware before handlers.
	adminHandlers := admin.NewHandlers(nil, authHandlers, zap.NewNop())

	r := chi.NewRouter()
	authHandlers.Mount(r)   // /auth/signup, /auth/login, etc.
	adminHandlers.Mount(r)  // /admin/dashboard, /admin/orders, etc.

	return r, cleanup
}

// noopSender satisfies the email.Sender interface without sending anything.
type noopSender struct{}

func (noopSender) Send(_ context.Context, _, _ string, _ map[string]any) error {
	return nil
}

// postJSONAdmin is a minimal inline helper (postJSON lives in auth_test and
// is not exported; we duplicate the 10-line pattern rather than create a new
// testsupport export).
func postJSONAdmin(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// signupAndGetAdminCookie signs up a new user and returns the session cookie.
func signupAndGetAdminCookie(t *testing.T, router http.Handler, email string) *http.Cookie {
	t.Helper()
	rr := postJSONAdmin(t, router, "/auth/signup", map[string]string{
		"email": email, "password": "hunter22", "name": "Test User",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup(%s) = %d, want 201; body: %s", email, rr.Code, rr.Body.String())
	}
	if c := testsupport.FindCookie(rr.Result(), "rue_session"); c != nil {
		return c
	}
	t.Fatalf("signup(%s): no rue_session cookie", email)
	return nil
}

// doGetAdmin sends GET <path> with an optional cookie, returns the recorder.
func doGetAdmin(router http.Handler, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// TestAdminRoutesRequireAdminRole pins the RBAC gating for the admin routes
// mounted under /admin (anonymous → 401, customer → 403) across three
// representative endpoints that the OpenAPI contract declares.
func TestAdminRoutesRequireAdminRole(t *testing.T) {
	router, cleanup := buildAdminTestRouter(t)
	defer cleanup()

	customerCookie := signupAndGetAdminCookie(t, router, "customer@admin-rbac.test")

	type testCase struct {
		name     string
		cookie   *http.Cookie
		path     string
		wantCode int
	}

	cases := []testCase{
		// /admin/dashboard
		{"anonymous → /admin/dashboard → 401", nil, "/admin/dashboard", http.StatusUnauthorized},
		{"customer → /admin/dashboard → 403", customerCookie, "/admin/dashboard", http.StatusForbidden},
		// /admin/orders
		{"anonymous → /admin/orders → 401", nil, "/admin/orders", http.StatusUnauthorized},
		{"customer → /admin/orders → 403", customerCookie, "/admin/orders", http.StatusForbidden},
		// /admin/analytics/stats
		{"anonymous → /admin/analytics/stats → 401", nil, "/admin/analytics/stats", http.StatusUnauthorized},
		{"customer → /admin/analytics/stats → 403", customerCookie, "/admin/analytics/stats", http.StatusForbidden},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := doGetAdmin(router, tc.path, tc.cookie)
			if rr.Code != tc.wantCode {
				t.Errorf("GET %s → %d, want %d; body: %s",
					tc.path, rr.Code, tc.wantCode, rr.Body.String())
			}
		})
	}
}
