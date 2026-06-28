package auth_test

// rbac_test.go — RBAC integration matrix
//
// Covers:
//   - Anonymous / Customer / Admin × GET /me + GET /admin/ping
//   - Unit tests for MustBeAdmin (no testcontainers needed)

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/me"
)

// ── router builder ───────────────────────────────────────────────────────────

// buildRBACRouter wires the same Group structure that main.go uses, plus a
// stub /admin/ping protected by RequireSession + RequireRole("admin").
func buildRBACRouter(h *auth.Handlers) http.Handler {
	r := chi.NewRouter()

	// Public auth routes.
	h.Mount(r)

	// Auth-gated group.
	r.Group(func(r chi.Router) {
		r.Use(h.RequireSession)

		// GET /me
		me.NewHandlers().Mount(r)
		// POST /auth/verify-email/resend
		h.MountAuthGated(r)

		// Admin-only sub-group.
		r.Group(func(r chi.Router) {
			r.Use(h.RequireRole("admin"))
			r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
				// Belt-and-suspenders: MustBeAdmin even though middleware already checked.
				if !auth.MustBeAdmin(w, r) {
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			})
		})
	})

	return r
}

// ── helpers ──────────────────────────────────────────────────────────────────

// signupAndGetCookie signs up a new user and returns the session cookie.
func signupAndGetCookie(t *testing.T, router http.Handler, email string) *http.Cookie {
	t.Helper()
	rr := postJSON(t, router, "/auth/signup", map[string]string{
		"email": email, "password": "hunter22", "name": "Test User",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup(%s) = %d, want 201; body: %s", email, rr.Code, rr.Body.String())
	}
	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			return c
		}
	}
	t.Fatalf("signup(%s): no rue_session cookie", email)
	return nil
}

// doGet sends GET <path> with an optional cookie, returns the recorder.
func doGet(router http.Handler, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// ── RBAC matrix test ─────────────────────────────────────────────────────────

func TestRBACMatrix(t *testing.T) {
	svc, pool, cleanup := newService(t)
	defer cleanup()

	h := auth.NewHandlers(svc, "rue_session", "", false)
	router := buildRBACRouter(h)

	// --- set up identities ---

	customerCookie := signupAndGetCookie(t, router, "customer@rbac.test")

	// For admin we need the pool to insert the role.
	// db.Pool is *pgxpool.Pool which has Exec(ctx, sql, args...) returning (pgconn.CommandTag, error).
	// We can't use the generic interface above with pgxpool directly; use the pool directly.
	adminCookie := func() *http.Cookie {
		rr := postJSON(t, router, "/auth/signup", map[string]string{
			"email": "admin@rbac.test", "password": "hunter22", "name": "Admin User",
		})
		if rr.Code != http.StatusCreated {
			t.Fatalf("admin signup = %d; body: %s", rr.Code, rr.Body.String())
		}
		var body struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode admin signup body: %v", err)
		}
		if _, err := pool.Exec(
			context.Background(),
			`INSERT INTO user_roles (user_id, role) VALUES ($1, 'admin') ON CONFLICT DO NOTHING`,
			body.UserID,
		); err != nil {
			t.Fatalf("insert admin role: %v", err)
		}
		// Login fresh to mint a session that reflects the admin role.
		lrr := postJSON(t, router, "/auth/login", map[string]string{
			"email": "admin@rbac.test", "password": "hunter22",
		})
		if lrr.Code != http.StatusOK {
			t.Fatalf("admin login = %d; body: %s", lrr.Code, lrr.Body.String())
		}
		for _, c := range lrr.Result().Cookies() {
			if c.Name == "rue_session" {
				return c
			}
		}
		t.Fatal("admin login: no session cookie")
		return nil
	}()

	// --- matrix ---

	type testCase struct {
		name     string
		cookie   *http.Cookie
		path     string
		wantCode int
	}

	cases := []testCase{
		// Anonymous
		{"anonymous → /me → 401", nil, "/me", http.StatusUnauthorized},
		{"anonymous → /admin/ping → 401", nil, "/admin/ping", http.StatusUnauthorized},
		// Customer
		{"customer → /me → 200", customerCookie, "/me", http.StatusOK},
		{"customer → /admin/ping → 403", customerCookie, "/admin/ping", http.StatusForbidden},
		// Admin
		{"admin → /me → 200", adminCookie, "/me", http.StatusOK},
		{"admin → /admin/ping → 200", adminCookie, "/admin/ping", http.StatusOK},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := doGet(router, tc.path, tc.cookie)
			if rr.Code != tc.wantCode {
				t.Errorf("GET %s → %d, want %d; body: %s",
					tc.path, rr.Code, tc.wantCode, rr.Body.String())
			}
		})
	}

	// Verify admin /me response includes role:"admin"
	t.Run("admin /me role field", func(t *testing.T) {
		rr := doGet(router, "/me", adminCookie)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET /me (admin) = %d", rr.Code)
		}
		var resp struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("decode /me body: %v", err)
		}
		if resp.Role != "admin" {
			t.Errorf("role = %q, want admin", resp.Role)
		}
	})
}

// ── MustBeAdmin unit tests (no testcontainers) ───────────────────────────────

// TestMustBeAdmin_AdminContext tests that MustBeAdmin returns true when the
// request context carries role "admin".
func TestMustBeAdmin_AdminContext(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()

	h := auth.NewHandlers(svc, "rue_session", "", false)

	// Build a minimal router: RequireSession wraps a handler that calls MustBeAdmin.
	var gotResult bool
	innerRouter := chi.NewRouter()
	innerRouter.Use(h.RequireSession)
	innerRouter.Get("/probe", func(w http.ResponseWriter, r *http.Request) {
		gotResult = auth.MustBeAdmin(w, r)
		if gotResult {
			w.WriteHeader(http.StatusOK)
		}
		// If false, MustBeAdmin already wrote 403.
	})

	signupRouter := chi.NewRouter()
	h.Mount(signupRouter)

	// Sign up admin user.
	rr := postJSON(t, signupRouter, "/auth/signup", map[string]string{
		"email": "mustbeadmin-ok@test.test", "password": "hunter22", "name": "A",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup = %d", rr.Code)
	}
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Get the pool from the service via the repository.
	dbPool := svc.Repo.Pool()
	if _, err := dbPool.Exec(
		context.Background(),
		`INSERT INTO user_roles (user_id, role) VALUES ($1, 'admin') ON CONFLICT DO NOTHING`,
		body.UserID,
	); err != nil {
		t.Fatalf("insert admin role: %v", err)
	}

	// Login to get fresh session with admin role.
	lrr := postJSON(t, signupRouter, "/auth/login", map[string]string{
		"email": "mustbeadmin-ok@test.test", "password": "hunter22",
	})
	if lrr.Code != http.StatusOK {
		t.Fatalf("login = %d", lrr.Code)
	}
	var adminCookie *http.Cookie
	for _, c := range lrr.Result().Cookies() {
		if c.Name == "rue_session" {
			adminCookie = c
			break
		}
	}
	if adminCookie == nil {
		t.Fatal("no session cookie")
	}

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(adminCookie)
	probeRR := httptest.NewRecorder()
	innerRouter.ServeHTTP(probeRR, req)

	if probeRR.Code != http.StatusOK {
		t.Errorf("MustBeAdmin with admin role: got %d, want 200", probeRR.Code)
	}
	if !gotResult {
		t.Error("MustBeAdmin should return true for admin")
	}
}

// TestMustBeAdmin_NonAdminContext tests that MustBeAdmin returns false and
// writes 403 when the request context carries a non-admin role.
func TestMustBeAdmin_NonAdminContext(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()

	h := auth.NewHandlers(svc, "rue_session", "", false)

	var gotResult bool
	innerRouter := chi.NewRouter()
	innerRouter.Use(h.RequireSession)
	innerRouter.Get("/probe", func(w http.ResponseWriter, r *http.Request) {
		gotResult = auth.MustBeAdmin(w, r)
		if gotResult {
			w.WriteHeader(http.StatusOK)
		}
		// If false, MustBeAdmin already wrote 403.
	})

	signupRouter := chi.NewRouter()
	h.Mount(signupRouter)

	rr := postJSON(t, signupRouter, "/auth/signup", map[string]string{
		"email": "mustbeadmin-no@test.test", "password": "hunter22", "name": "B",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup = %d", rr.Code)
	}

	var customerCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			customerCookie = c
			break
		}
	}
	if customerCookie == nil {
		t.Fatal("no session cookie")
	}

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(customerCookie)
	probeRR := httptest.NewRecorder()
	innerRouter.ServeHTTP(probeRR, req)

	if probeRR.Code != http.StatusForbidden {
		t.Errorf("MustBeAdmin with customer role: got %d, want 403", probeRR.Code)
	}
	if gotResult {
		t.Error("MustBeAdmin should return false for non-admin")
	}
}
