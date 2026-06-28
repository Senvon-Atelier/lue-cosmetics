package cart

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// newCartHandlerStack spins up auth.Service + cart.Service + a chi router
// roughly mirroring cmd/api/main.go but minimal. Returns the router and a
// cleanup func.
func newCartHandlerStack(t *testing.T) (*Handlers, http.Handler, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams
	authHandlers := auth.NewHandlers(authSvc, "rue_session", "", false)

	cartSvc := newCartService(t, pool)
	cartHandlers := NewHandlers(cartSvc, authSvc, "rue_session", "", false)

	r := chi.NewRouter()
	authHandlers.Mount(r)
	cartHandlers.Mount(r)

	return cartHandlers, r, cleanup
}

func doJSON(t *testing.T, router http.Handler, method, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func TestHandler_GetCart_NoCookies_MintsGuestAndSetsCookie(t *testing.T) {
	_, router, cleanup := newCartHandlerStack(t)
	defer cleanup()

	rr := doJSON(t, router, http.MethodGet, "/cart", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	if c := testsupport.FindCookie(rr.Result(), GuestCookieName); c == nil || c.Value == "" {
		t.Fatalf("expected %s cookie to be set; headers: %v", GuestCookieName, rr.Header())
	}
	setCookie := rr.Header().Get("Set-Cookie")
	if strings.Contains(setCookie, "HttpOnly") {
		t.Errorf("guest cookie must NOT be HttpOnly: %s", setCookie)
	}
	var body cartResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 0 {
		t.Errorf("expected empty items, got %d", len(body.Items))
	}
}

func TestHandler_GetCart_WithGuestCookie_NoNewCookie(t *testing.T) {
	_, router, cleanup := newCartHandlerStack(t)
	defer cleanup()

	// First call mints.
	rr1 := doJSON(t, router, http.MethodGet, "/cart", nil)
	guestCookie := testsupport.FindCookie(rr1.Result(), GuestCookieName)
	if guestCookie == nil {
		t.Fatal("no guest cookie minted on first call")
	}

	// Second call with that cookie shouldn't mint again.
	rr2 := doJSON(t, router, http.MethodGet, "/cart", nil, guestCookie)
	if rr2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr2.Code)
	}
	if c := testsupport.FindCookie(rr2.Result(), GuestCookieName); c != nil {
		t.Errorf("expected no new guest cookie on reuse; got %q", c.Value)
	}
}

func TestHandler_PostCartItems_AddsAndMintsCookieIfNeeded(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	logger := zap.NewNop()

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams
	authHandlers := auth.NewHandlers(authSvc, "rue_session", "", false)

	cartSvc := newCartService(t, pool)
	cartHandlers := NewHandlers(cartSvc, authSvc, "rue_session", "", false)

	r := chi.NewRouter()
	authHandlers.Mount(r)
	cartHandlers.Mount(r)

	productID := seedTestProduct(t, ctx, pool)

	rr := doJSON(t, r, http.MethodPost, "/cart/items", map[string]any{
		"product_id": productID.String(),
		"qty":        2,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	if c := testsupport.FindCookie(rr.Result(), GuestCookieName); c == nil {
		t.Errorf("expected guest cookie minted on first POST")
	}
	var body cartResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].Qty != 2 {
		t.Errorf("expected 1 item qty=2, got %+v", body.Items)
	}
}

func TestHandler_PatchCartItems_BadQty_400(t *testing.T) {
	_, router, cleanup := newCartHandlerStack(t)
	defer cleanup()

	// Need a real cart item id for the URL parse to succeed past chi.
	// Use a random uuid; we don't expect to reach the service (qty <1 is
	// caught after parse but before service call).
	rr := doJSON(t, router, http.MethodPatch, "/cart/items/00000000-0000-0000-0000-000000000001", map[string]any{
		"qty": 0,
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_PatchCartItems_CrossCart_404(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	logger := zap.NewNop()

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams
	authHandlers := auth.NewHandlers(authSvc, "rue_session", "", false)
	cartSvc := newCartService(t, pool)
	cartHandlers := NewHandlers(cartSvc, authSvc, "rue_session", "", false)

	r := chi.NewRouter()
	authHandlers.Mount(r)
	cartHandlers.Mount(r)

	productID := seedTestProduct(t, ctx, pool)

	// Cart A: add an item, capture its id and cookie.
	rrA := doJSON(t, r, http.MethodPost, "/cart/items", map[string]any{
		"product_id": productID.String(), "qty": 1,
	})
	if rrA.Code != http.StatusOK {
		t.Fatalf("A post = %d; %s", rrA.Code, rrA.Body.String())
	}
	var bodyA cartResponse
	if err := json.NewDecoder(rrA.Body).Decode(&bodyA); err != nil {
		t.Fatalf("decode A: %v", err)
	}
	itemID := bodyA.Items[0].ID

	// Cart B: a separate guest session.
	rrB := doJSON(t, r, http.MethodGet, "/cart", nil)
	cookieB := testsupport.FindCookie(rrB.Result(), GuestCookieName)
	if cookieB == nil {
		t.Fatal("no cookie for B")
	}

	// B tries to PATCH A's item → 404.
	rr := doJSON(t, r, http.MethodPatch, "/cart/items/"+itemID, map[string]any{
		"qty": 9,
	}, cookieB)
	if rr.Code != http.StatusNotFound {
		t.Errorf("cross-cart PATCH = %d, want 404; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DeleteCartItems_Happy_204(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	logger := zap.NewNop()

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams
	authHandlers := auth.NewHandlers(authSvc, "rue_session", "", false)
	cartSvc := newCartService(t, pool)
	cartHandlers := NewHandlers(cartSvc, authSvc, "rue_session", "", false)

	r := chi.NewRouter()
	authHandlers.Mount(r)
	cartHandlers.Mount(r)

	productID := seedTestProduct(t, ctx, pool)

	rrAdd := doJSON(t, r, http.MethodPost, "/cart/items", map[string]any{
		"product_id": productID.String(), "qty": 1,
	})
	cookie := testsupport.FindCookie(rrAdd.Result(), GuestCookieName)
	var body cartResponse
	if err := json.NewDecoder(rrAdd.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	itemID := body.Items[0].ID

	rr := doJSON(t, r, http.MethodDelete, "/cart/items/"+itemID, nil, cookie)
	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
}
