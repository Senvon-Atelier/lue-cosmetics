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
	"github.com/oti-adjei/ruecosmetics/internal/orders"
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
	ordersRepo := orders.NewRepository(pool)
	meH := me.NewHandlers(ordersRepo)

	r := chi.NewRouter()
	authH.Mount(r) // /auth/signup, /auth/login etc.
	r.Group(func(g chi.Router) {
		g.Use(authH.RequireSession)
		meH.Mount(g)
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

func TestListOrders_WithValidSession_Returns200(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	// Sign up to get a session cookie
	body := strings.NewReader(`{"email":"orders@handler.test","password":"hunter22","name":"OrdersUser"}`)
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

	// List orders (should be empty for new user)
	listReq := httptest.NewRequest(http.MethodGet, "/me/orders", nil)
	listReq.AddCookie(sessionCookie)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("GET /me/orders: status %d, body: %s", listRR.Code, listRR.Body.String())
	}
	respBody := listRR.Body.String()
	if !strings.Contains(respBody, `"orders":[]`) {
		t.Errorf("expected empty orders array, got: %s", respBody)
	}
	if !strings.Contains(respBody, `"total":0`) {
		t.Errorf("expected total 0, got: %s", respBody)
	}
}

func TestListOrders_WithoutCookie_Returns401(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/me/orders", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetOrder_InvalidUUID_Returns400(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	// Sign up to get a session cookie
	body := strings.NewReader(`{"email":"getorder@handler.test","password":"hunter22","name":"GetOrderUser"}`)
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

	// Try to get order with invalid UUID
	getReq := httptest.NewRequest(http.MethodGet, "/me/orders/not-a-uuid", nil)
	getReq.AddCookie(sessionCookie)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", getRR.Code)
	}
}

func TestGetOrder_WithoutCookie_Returns401(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/me/orders/some-uuid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpdateProfile_Returns501(t *testing.T) {
	router, cleanup := newMeRouter(t)
	defer cleanup()

	// Sign up to get a session cookie
	body := strings.NewReader(`{"email":"update@handler.test","password":"hunter22","name":"UpdateUser"}`)
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

	// Try to update profile (should return 501 - not implemented)
	updateBody := strings.NewReader(`{"name":"Updated Name"}`)
	updateReq := httptest.NewRequest(http.MethodPatch, "/me", updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(sessionCookie)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for unimplemented endpoint, got %d: %s", updateRR.Code, updateRR.Body.String())
	}
}
