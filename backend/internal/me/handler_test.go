package me_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/me"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func newMeRouter(t *testing.T) (http.Handler, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()
	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, logger, email.LogSender{Log: logger}, nil)
	svc.Params = auth.TestParams

	authH := auth.NewHandlers(svc, "rue_session", "", false)

	r := chi.NewRouter()
	authH.Mount(r) // /auth/signup, /auth/login etc.
	r.Group(func(g chi.Router) {
		g.Use(authH.RequireSession)
		me.NewHandlers().Mount(g)
	})

	return r, cleanup
}

func TestGetMe_WithValidCookie_Returns200(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	// Sign up to get a session cookie.
	body := strings.NewReader(`{"email":"me@handler.test","password":"hunter22","name":"MeUser"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}

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

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.AddCookie(sessionCookie)
	meRR := httptest.NewRecorder()
	router.ServeHTTP(meRR, meReq)

	if meRR.Code != http.StatusOK {
		t.Fatalf("GET /me: status %d, body: %s", meRR.Code, meRR.Body.String())
	}
	respBody := meRR.Body.String()
	if !strings.Contains(respBody, `"email":"me@handler.test"`) {
		t.Errorf("unexpected body: %s", respBody)
	}
	if !strings.Contains(respBody, `"role":"customer"`) {
		t.Errorf("expected role:customer in body: %s", respBody)
	}
}

func TestGetMe_WithoutCookie_Returns401(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
