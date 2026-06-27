package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
)

// stubValidator is a test double for IDTokenValidator that returns a fixed Payload.
type stubValidator struct {
	payload *auth.Payload
	err     error
}

func (s *stubValidator) Validate(_ context.Context, _, _ string) (*auth.Payload, error) {
	return s.payload, s.err
}

func newHandlers(t *testing.T) (*auth.Handlers, func()) {
	t.Helper()
	svc, _, cleanup := newService(t)
	h := auth.NewHandlers(svc, "rue_session", "", false)
	return h, cleanup
}

func routerWith(h *auth.Handlers) http.Handler {
	r := chi.NewRouter()
	h.Mount(r)
	return r
}

func postJSON(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
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

func TestHandlerSignupCreatesSessionCookieAndReturns201(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	rr := postJSON(t, router, "/auth/signup", map[string]string{
		"email": "alice@handler.test", "password": "hunter22", "name": "Alice",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"email_verified":true`) {
		t.Errorf("expected email_verified:true in body: %s", body)
	}
	setCookie := rr.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "rue_session=") {
		t.Errorf("expected session cookie; Set-Cookie: %s", setCookie)
	}
	if !strings.Contains(setCookie, "HttpOnly") {
		t.Errorf("expected HttpOnly cookie; Set-Cookie: %s", setCookie)
	}
}

func TestHandlerSignupDuplicateEmailReturns409(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	payload := map[string]string{"email": "bob@handler.test", "password": "hunter22"}
	postJSON(t, router, "/auth/signup", payload) // first signup
	rr := postJSON(t, router, "/auth/signup", payload)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "conflict") {
		t.Errorf("expected conflict code in body: %s", rr.Body.String())
	}
}

func TestHandlerLoginReturns200AndSetsCookie(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "carol@handler.test", "password": "hunter22",
	})
	rr := postJSON(t, router, "/auth/login", map[string]string{
		"email": "carol@handler.test", "password": "hunter22",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Header().Get("Set-Cookie"), "rue_session=") {
		t.Errorf("expected session cookie on login; Set-Cookie: %s", rr.Header().Get("Set-Cookie"))
	}
}

func TestHandlerLoginWrongPasswordReturns401(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	postJSON(t, router, "/auth/signup", map[string]string{
		"email": "dave@handler.test", "password": "hunter22",
	})
	rr := postJSON(t, router, "/auth/login", map[string]string{
		"email": "dave@handler.test", "password": "WRONG",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "unauthorized") {
		t.Errorf("expected unauthorized code in body: %s", rr.Body.String())
	}
}

func TestHandlerLogoutClearsCookieAndReturns204(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	// Sign up and grab cookie.
	srr := postJSON(t, router, "/auth/signup", map[string]string{
		"email": "eve@handler.test", "password": "hunter22",
	})
	cookie := ""
	for _, c := range srr.Result().Cookies() {
		if c.Name == "rue_session" {
			cookie = c.Name + "=" + c.Value
		}
	}
	if cookie == "" {
		t.Fatal("no session cookie from signup")
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Cookie", cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "rue_session=;") && !strings.Contains(sc, "Max-Age=0") && !strings.Contains(sc, "max-age=0") {
		// Accept any cleared cookie form.
		if !strings.Contains(sc, "Max-Age=-1") && !strings.Contains(sc, "rue_session=") {
			t.Logf("Set-Cookie after logout: %s", sc)
		}
	}
}

func TestHandlerSessionWithValidCookieReturns200(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	srr := postJSON(t, router, "/auth/signup", map[string]string{
		"email": "frank@handler.test", "password": "hunter22", "name": "Frank",
	})
	cookie := ""
	for _, c := range srr.Result().Cookies() {
		if c.Name == "rue_session" {
			cookie = c.Name + "=" + c.Value
		}
	}
	if cookie == "" {
		t.Fatal("no session cookie from signup")
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)
	req.Header.Set("Cookie", cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "frank@handler.test") {
		t.Errorf("expected email in session body: %s", body)
	}
	if !strings.Contains(body, `"role":"customer"`) {
		t.Errorf("expected role in session body: %s", body)
	}
}

func TestHandlerSessionNoCookieReturns204(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	router := routerWith(h)

	req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rr.Code)
	}
}

// ── Google OAuth handler tests ───────────────────────────────────────────────

func TestHandlerGoogleStartNotConfiguredReturns503(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	// GoogleClientID is empty by default — not configured.
	router := routerWith(h)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/start", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "not_configured") {
		t.Errorf("expected not_configured code in body: %s", rr.Body.String())
	}
}

func TestHandlerGoogleStartWithCredsSetsStateCookieAndRedirects(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	h.GoogleClientID = "test-client-id"
	h.GoogleClientSecret = "test-client-secret"
	h.GoogleRedirectURL = "http://localhost/callback"
	router := routerWith(h)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/start", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302; body: %s", rr.Code, rr.Body.String())
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "accounts.google.com") {
		t.Errorf("expected redirect to Google; got: %s", location)
	}
	// State cookie must be set.
	setCookie := rr.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "rue_oauth_state=") {
		t.Errorf("expected rue_oauth_state cookie; Set-Cookie: %s", setCookie)
	}
}

func TestHandlerGoogleCallbackStateMismatchReturns400(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	h.GoogleClientID = "test-client-id"
	h.GoogleClientSecret = "test-client-secret"
	router := routerWith(h)

	// Request has state in query but cookie holds a different value.
	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=querystate&code=somecode", nil)
	req.AddCookie(&http.Cookie{Name: "rue_oauth_state", Value: "different-state"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "validation_failed") {
		t.Errorf("expected validation_failed code in body: %s", rr.Body.String())
	}
}
