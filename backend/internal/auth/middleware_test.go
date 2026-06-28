package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// stubHandler records whether it was called and what UserID it saw.
type stubHandler struct {
	called bool
	userID uuid.UUID
	ok     bool
}

func (s *stubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.called = true
	s.userID, s.ok = auth.GetUserID(r.Context())
	w.WriteHeader(http.StatusOK)
}

func TestRequireSession_MissingCookie_Returns401(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()

	stub := &stubHandler{}
	r := chi.NewRouter()
	r.Use(h.RequireSession)
	r.Get("/me", stub.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if stub.called {
		t.Error("inner handler should not have been called")
	}
}

func TestRequireSession_InvalidToken_Returns401(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()

	stub := &stubHandler{}
	r := chi.NewRouter()
	r.Use(h.RequireSession)
	r.Get("/me", stub.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "rue_session", Value: "bogus-token"})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireSession_ValidToken_InjectsContext(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()

	// Sign up to get a real session token.
	rr := postJSON(t, routerWith(h), "/auth/signup", map[string]string{
		"email": "mw@test.test", "password": "hunter22", "name": "MW",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup failed: %d %s", rr.Code, rr.Body.String())
	}
	// Extract cookie.
	sessionCookie := testsupport.FindCookie(rr.Result(), "rue_session")
	if sessionCookie == nil {
		t.Fatal("no session cookie from signup")
	}

	stub := &stubHandler{}
	mwRouter := chi.NewRouter()
	mwRouter.Use(h.RequireSession)
	mwRouter.Get("/me", stub.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(sessionCookie)
	rrMe := httptest.NewRecorder()
	mwRouter.ServeHTTP(rrMe, req)

	if rrMe.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrMe.Code)
	}
	if !stub.called {
		t.Error("inner handler was not called")
	}
	if stub.userID == (uuid.UUID{}) || !stub.ok {
		t.Error("UserID not injected into context")
	}
}

func TestRequireRole_WrongRole_Returns403(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()

	// Sign up (gets "customer" role).
	rr := postJSON(t, routerWith(h), "/auth/signup", map[string]string{
		"email": "cust@role.test", "password": "hunter22", "name": "Cust",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup: %d", rr.Code)
	}
	sessionCookie := testsupport.FindCookie(rr.Result(), "rue_session")

	stub := &stubHandler{}
	r := chi.NewRouter()
	r.Use(h.RequireSession)
	r.With(h.RequireRole("admin")).Get("/admin", stub.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(sessionCookie)
	rrAdmin := httptest.NewRecorder()
	r.ServeHTTP(rrAdmin, req)

	if rrAdmin.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rrAdmin.Code)
	}
	if stub.called {
		t.Error("inner handler should not have been called")
	}
}

// TestRequireSession_RolledExpiry_RefreshesCookie verifies that when GetSession
// rolls the DB expiry forward (remaining < 29 days), RequireSession re-emits
// the browser cookie with a fresh Expires/MaxAge.
func TestRequireSession_RolledExpiry_RefreshesCookie(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()

	h := auth.NewHandlers(svc, "rue_session", "", false)
	signupRR := postJSON(t, routerWith(h), "/auth/signup", map[string]string{
		"email": "roll@mw.test", "password": "hunter22", "name": "Roll",
	})
	if signupRR.Code != http.StatusCreated {
		t.Fatalf("signup: %d %s", signupRR.Code, signupRR.Body.String())
	}
	sessionCookie := testsupport.FindCookie(signupRR.Result(), "rue_session")
	if sessionCookie == nil {
		t.Fatal("no session cookie from signup")
	}

	// Advance the service clock so the session appears to have only 1 day left
	// (well below the 29-day roll threshold), which will trigger a roll.
	svc.Now = func() time.Time { return time.Now().Add(29 * 24 * time.Hour) }

	mwRouter := chi.NewRouter()
	mwRouter.Use(h.RequireSession)
	mwRouter.Get("/probe", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.AddCookie(sessionCookie)
	rrProbe := httptest.NewRecorder()
	mwRouter.ServeHTTP(rrProbe, req)

	if rrProbe.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rrProbe.Code, rrProbe.Body.String())
	}

	// The response must contain a fresh Set-Cookie header for rue_session.
	refreshed := testsupport.FindCookie(rrProbe.Result(), "rue_session")
	if refreshed == nil {
		t.Fatal("RequireSession did not re-emit the session cookie after rolling expiry")
	}
	if refreshed.MaxAge <= 0 && refreshed.Expires.IsZero() {
		t.Error("refreshed cookie has no valid MaxAge or Expires")
	}
	// Confirm the cookie value is unchanged (same token).
	if refreshed.Value != sessionCookie.Value {
		t.Errorf("refreshed cookie token changed: got %q want %q", refreshed.Value, sessionCookie.Value)
	}
	// Confirm Set-Cookie header contains "rue_session" (redundant but explicit).
	setCookie := rrProbe.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "rue_session=") {
		t.Errorf("Set-Cookie header missing rue_session: %s", setCookie)
	}
}

func TestRequireRole_MatchingRole_Passes(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()

	// Sign up (gets "customer" role).
	rr := postJSON(t, routerWith(h), "/auth/signup", map[string]string{
		"email": "pass@role.test", "password": "hunter22", "name": "Pass",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup: %d", rr.Code)
	}
	sessionCookie := testsupport.FindCookie(rr.Result(), "rue_session")

	stub := &stubHandler{}
	r := chi.NewRouter()
	r.Use(h.RequireSession)
	r.With(h.RequireRole("customer")).Get("/profile", stub.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.AddCookie(sessionCookie)
	rrProf := httptest.NewRecorder()
	r.ServeHTTP(rrProf, req)

	if rrProf.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrProf.Code)
	}
	if !stub.called {
		t.Error("inner handler should have been called")
	}
}
