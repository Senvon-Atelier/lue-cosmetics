package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
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
	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			sessionCookie = c
			break
		}
	}
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
	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			sessionCookie = c
			break
		}
	}

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
	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			sessionCookie = c
			break
		}
	}

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
