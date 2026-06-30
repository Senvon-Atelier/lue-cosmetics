package addresses

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func newAddressesRouter(t *testing.T) (http.Handler, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()

	// Setup auth
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams

	authH := auth.NewHandlers(authSvc, "rue_session", "", false)

	// Setup addresses
	repo := NewRepository(pool)
	svc := NewService(repo, pool, logger)
	addrH := NewHandlers(svc, logger)

	r := chi.NewRouter()
	authH.Mount(r) // /auth/signup, /auth/login etc.
	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(g chi.Router) {
			g.Use(authH.RequireSession)
			g.Route("/me", func(meRouter chi.Router) {
				addrH.Mount(meRouter) // /me/addresses*
			})
		})
	})

	return r, func() {
		cleanup()
	}
}

func TestMount_AttachesRelativeToMeSubrouter(t *testing.T) {
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	logger := zap.NewNop()
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams
	authH := auth.NewHandlers(authSvc, "rue_session", "", false)

	addrH := NewHandlers(NewService(NewRepository(pool), pool, logger), logger)
	r := chi.NewRouter()
	authH.Mount(r)
	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(g chi.Router) {
			g.Use(authH.RequireSession)
			g.Route("/me", func(meRouter chi.Router) {
				addrH.Mount(meRouter)
			})
		})
	})

	cookie := signupAndGetCookie(t, r)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/addresses", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected relative /me subrouter mount to serve addresses list with 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func signupAndGetCookie(t *testing.T, router http.Handler) *http.Cookie {
	t.Helper()
	body := strings.NewReader(`{"email":"addr@handler.test","password":"hunter22","name":"AddrUser"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}

	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			return c
		}
	}
	t.Fatal("no session cookie from signup")
	return nil
}

func TestCreate_201_HappyPath(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	body := strings.NewReader(`{
		"label": "Home",
		"line1": "123 Main St",
		"line2": "Apt 4B",
		"city": "Accra",
		"region": "Greater Accra",
		"phone": "0201234567"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses", body)
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp addressResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Label != "Home" {
		t.Errorf("expected label 'Home', got %s", resp.Label)
	}
	if !resp.IsDefault {
		t.Error("expected first address to be default")
	}
}

func TestCreate_400_MissingLine1(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	body := strings.NewReader(`{
		"label": "Home",
		"line1": "",
		"line2": "",
		"city": "Accra",
		"region": "Greater Accra",
		"phone": "0201234567"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses", body)
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreate_400_LimitReached(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	// Create max addresses
	for i := 0; i < MaxAddressesPerUser; i++ {
		body := strings.NewReader(`{
			"label": "Addr",
			"line1": "123 St",
			"line2": "",
			"city": "Accra",
			"region": "Greater Accra",
			"phone": "0201234567"
		}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses", body)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated && i < MaxAddressesPerUser-1 {
			t.Fatalf("create %d: %d %s", i, rr.Code, rr.Body.String())
		}
	}

	// Try to create one more
	body := strings.NewReader(`{
		"label": "Overflow",
		"line1": "456 St",
		"line2": "",
		"city": "Accra",
		"region": "Greater Accra",
		"phone": "0209876543"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses", body)
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	// Check for address_limit_reached code
	var errResp struct {
		Error struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if errResp.Error.Fields["code"] != "address_limit_reached" {
		t.Errorf("expected fields.code=address_limit_reached, got %v", errResp.Error.Fields)
	}
}

func TestList_200_OrdersDefaultFirst(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	// Create first address (will be default)
	addr1 := createAddress(t, router, cookie, "Home1", "1 St")
	// Create second address (not default)
	createAddress(t, router, cookie, "Work", "2 St")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/addresses", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp listAddressesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(resp.Addresses))
	}
	if !resp.Addresses[0].IsDefault {
		t.Error("expected first address to be default")
	}
	if resp.Addresses[0].ID != addr1 {
		t.Error("expected first address (Home1) to be default and first")
	}
}

func TestUpdate_200_HappyPath(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	addrID := createAddress(t, router, cookie, "Original", "123 Main St")

	body := strings.NewReader(`{"label": "Work"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/addresses/"+addrID, body)
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp addressResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Label != "Work" {
		t.Errorf("expected label 'Work', got %s", resp.Label)
	}
}

func TestUpdate_404_NotOwned(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	// User A
	cookieA := signupAndGetCookieWithName(t, router, "usera@handler.test")
	addrIDA := createAddress(t, router, cookieA, "UserA Home", "1 St")

	// User B tries to update User A's address
	cookieB := signupAndGetCookieWithName(t, router, "userb@handler.test")

	body := strings.NewReader(`{"label": "Hacked"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/addresses/"+addrIDA, body)
	req.AddCookie(cookieB)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdate_404_DoesNotExist(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	body := strings.NewReader(`{"label": "X"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/addresses/"+uuid.New().String(), body)
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDelete_204_HappyPath_NonDefault(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	// Create default first
	createAddress(t, router, cookie, "Home", "1 Home St")
	// Create non-default
	addrID := createAddress(t, router, cookie, "Work", "2 Work St")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/addresses/"+addrID, nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDelete_204_HappyPath_DefaultPromotesNextOldest(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	// Create three addresses
	home := createAddress(t, router, cookie, "Home", "1 Home St")
	work := createAddress(t, router, cookie, "Work", "2 Work St")
	gym := createAddress(t, router, cookie, "Gym", "3 Gym St")

	// Set Work as default
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses/"+work+"/default", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("set default: %d %s", rr.Code, rr.Body.String())
	}

	// Delete default (Work)
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/me/addresses/"+work, nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete default: %d %s", rr.Code, rr.Body.String())
	}

	// Verify Home is now default (oldest remaining)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me/addresses", nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var resp listAddressesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(resp.Addresses))
	}
	if !resp.Addresses[0].IsDefault || resp.Addresses[0].ID != home {
		t.Error("expected Home (oldest remaining) to be promoted to default")
	}
	if resp.Addresses[1].ID == gym && resp.Addresses[1].IsDefault {
		t.Error("expected Gym to not be default")
	}
}

func TestDelete_404_NotOwned(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookieA := signupAndGetCookieWithName(t, router, "usera@handler.test")
	addrIDA := createAddress(t, router, cookieA, "UserA Home", "1 St")

	cookieB := signupAndGetCookieWithName(t, router, "userb@handler.test")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/addresses/"+addrIDA, nil)
	req.AddCookie(cookieB)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSetDefault_200_FlipsAtomically(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	// Create two addresses
	home := createAddress(t, router, cookie, "Home", "1 Home St")
	work := createAddress(t, router, cookie, "Work", "2 Work St")

	// Set Work as default
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses/"+work+"/default", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp addressResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.IsDefault || resp.ID != work {
		t.Error("expected Work to be default")
	}

	// Verify list order
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me/addresses", nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var listResp listAddressesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if listResp.Addresses[0].ID != work || !listResp.Addresses[0].IsDefault {
		t.Error("expected Work to be default and first in list")
	}
	if listResp.Addresses[1].ID == home && listResp.Addresses[1].IsDefault {
		t.Error("expected Home to not be default")
	}
}

func TestSetDefault_200_IdempotentOnAlreadyDefault(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookie := signupAndGetCookie(t, router)

	// First address is default by creation
	addrID := createAddress(t, router, cookie, "Home", "1 Home St")

	// Set default on already-default (should be no-op)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses/"+addrID+"/default", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp addressResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.IsDefault {
		t.Error("expected address to still be default")
	}
}

func TestSetDefault_404_NotOwned(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	cookieA := signupAndGetCookieWithName(t, router, "usera@handler.test")
	addrIDA := createAddress(t, router, cookieA, "UserA Home", "1 St")

	cookieB := signupAndGetCookieWithName(t, router, "userb@handler.test")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses/"+addrIDA+"/default", nil)
	req.AddCookie(cookieB)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAllEndpoints_401_WithoutSession(t *testing.T) {
	router, cleanup := newAddressesRouter(t)
	defer cleanup()

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/v1/me/addresses", `{"label":"Home","line1":"1 St","city":"Accra","region":"Greater Accra","phone":"0201111111"}`},
		{"GET", "/api/v1/me/addresses", ""},
		{"PATCH", "/api/v1/me/addresses/" + uuid.New().String(), `{"label":"X"}`},
		{"DELETE", "/api/v1/me/addresses/" + uuid.New().String(), ""},
		{"POST", "/api/v1/me/addresses/" + uuid.New().String() + "/default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.method == "POST" || tt.method == "PATCH" {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d for %s %s", rr.Code, tt.method, tt.path)
			}
		})
	}
}

// Helper: create an address and return its ID
func createAddress(t *testing.T, router http.Handler, cookie *http.Cookie, label, line1 string) string {
	t.Helper()
	body := strings.NewReader(`{
		"label": "` + label + `",
		"line1": "` + line1 + `",
		"line2": "",
		"city": "Accra",
		"region": "Greater Accra",
		"phone": "0201234567"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/addresses", body)
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create address: %d %s", rr.Code, rr.Body.String())
	}

	var resp addressResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp.ID
}

// Helper: signup with a specific email
func signupAndGetCookieWithName(t *testing.T, router http.Handler, email string) *http.Cookie {
	t.Helper()
	body := strings.NewReader(`{"email":"` + email + `","password":"hunter22","name":"TestUser"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}

	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			return c
		}
	}
	t.Fatal("no session cookie from signup")
	return nil
}
