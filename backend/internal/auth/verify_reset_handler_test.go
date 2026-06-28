package auth_test

// Handler-level tests for:
//   POST /auth/verify-email
//   POST /auth/verify-email/resend  (auth-gated)
//   POST /auth/password-reset/request
//   POST /auth/password-reset/confirm

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// newHandlersWithCapture builds Handlers backed by a capturing email sender.
func newHandlersWithCapture(t *testing.T) (*auth.Handlers, *capturingSender, db.Pool, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()
	repo := auth.NewRepository(pool)
	cap := &capturingSender{}
	svc := auth.NewService(repo, logger, cap, nil)
	svc.Params = auth.TestParams
	h := auth.NewHandlers(svc, "rue_session", "", false)
	return h, cap, pool, cleanup
}

// routerWithGated builds a router that includes both public and auth-gated routes.
func routerWithGated(h *auth.Handlers) http.Handler {
	r := chi.NewRouter()
	h.Mount(r)
	r.Group(func(r chi.Router) {
		r.Use(h.RequireSession)
		h.MountAuthGated(r)
	})
	return r
}

// ── /auth/verify-email ────────────────────────────────────────────────────────

func TestHandlerVerifyEmail_HappyPath(t *testing.T) {
	h, cap, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	h.Svc.Allowlist = []string{"vip@handler.test"}
	router := routerWithGated(h)

	// Sign up to trigger the verify_email send.
	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "vip@handler.test", "password": "hunter22",
	})
	call := findCall(cap, "verify_email")
	if call == nil {
		t.Fatal("no verify_email email sent on signup")
	}
	raw := call.Data["token"].(string)

	rr := postJSON(t, router, "/auth/verify-email", map[string]string{"token": raw})
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerVerifyEmail_BadToken(t *testing.T) {
	h, _, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	rr := postJSON(t, router, "/auth/verify-email", map[string]string{"token": "notavalidtoken"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "validation_failed") {
		t.Errorf("expected validation_failed in body: %s", rr.Body.String())
	}
}

func TestHandlerVerifyEmail_ReusedToken400(t *testing.T) {
	h, cap, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	h.Svc.Allowlist = []string{"vip4@handler.test"}
	router := routerWithGated(h)

	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "vip4@handler.test", "password": "hunter22",
	})
	call := findCall(cap, "verify_email")
	raw := call.Data["token"].(string)

	postJSON(t, router, "/auth/verify-email", map[string]string{"token": raw}) // first use
	rr := postJSON(t, router, "/auth/verify-email", map[string]string{"token": raw})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("reuse: status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
}

// ── /auth/verify-email/resend ─────────────────────────────────────────────────

func TestHandlerResendVerification_RequiresAuth(t *testing.T) {
	h, _, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/verify-email/resend", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestHandlerResendVerification_AuthedReturns204(t *testing.T) {
	h, _, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	// Sign up + grab cookie.
	srr := postJSON(t, router, "/auth/signup", map[string]string{
		"email": "authed@handler.test", "password": "hunter22",
	})
	cookie := ""
	for _, c := range srr.Result().Cookies() {
		if c.Name == "rue_session" {
			cookie = c.Name + "=" + c.Value
		}
	}
	if cookie == "" {
		t.Fatal("no session cookie")
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/verify-email/resend", nil)
	req.Header.Set("Cookie", cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
}

// ── /auth/password-reset/request ─────────────────────────────────────────────

func TestHandlerPasswordResetRequest_UnknownEmail204(t *testing.T) {
	h, _, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	rr := postJSON(t, router, "/auth/password-reset/request", map[string]string{
		"email": "nobody@unknown.test",
	})
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerPasswordResetRequest_KnownEmail204(t *testing.T) {
	h, cap, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "known@handler.test", "password": "hunter22",
	})
	beforeCount := len(cap.calls)
	rr := postJSON(t, router, "/auth/password-reset/request", map[string]string{
		"email": "known@handler.test",
	})
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
	if len(cap.calls) <= beforeCount {
		t.Error("expected a password_reset email send call")
	}
}

// ── /auth/password-reset/confirm ──────────────────────────────────────────────

func TestHandlerPasswordResetConfirm_HappyPath(t *testing.T) {
	h, cap, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "prconf@handler.test", "password": "oldpassword",
	})
	postJSON(t, router, "/auth/password-reset/request", map[string]string{
		"email": "prconf@handler.test",
	})
	call := findCall(cap, "password_reset")
	if call == nil {
		t.Fatal("no password_reset email call")
	}
	raw := call.Data["token"].(string)

	rr := postJSON(t, router, "/auth/password-reset/confirm", map[string]string{
		"token": raw, "new_password": "newpassword99",
	})
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerPasswordResetConfirm_PasswordTooShort400(t *testing.T) {
	h, _, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	rr := postJSON(t, router, "/auth/password-reset/confirm", map[string]string{
		"token": "sometoken", "new_password": "short",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "validation_failed") {
		t.Errorf("expected validation_failed in body: %s", rr.Body.String())
	}
}

func TestHandlerPasswordResetConfirm_BadToken400(t *testing.T) {
	h, _, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	rr := postJSON(t, router, "/auth/password-reset/confirm", map[string]string{
		"token": "notarealtoken", "new_password": "validpassword123",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerPasswordResetConfirm_ReusedToken400(t *testing.T) {
	h, cap, _, cleanup := newHandlersWithCapture(t)
	defer cleanup()
	router := routerWithGated(h)

	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "prreuse@handler.test", "password": "oldpassword",
	})
	postJSON(t, router, "/auth/password-reset/request", map[string]string{
		"email": "prreuse@handler.test",
	})
	call := findCall(cap, "password_reset")
	raw := call.Data["token"].(string)

	postJSON(t, router, "/auth/password-reset/confirm", map[string]string{
		"token": raw, "new_password": "newpassword99",
	})
	rr := postJSON(t, router, "/auth/password-reset/confirm", map[string]string{
		"token": raw, "new_password": "newpassword99",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("reuse: status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
}
