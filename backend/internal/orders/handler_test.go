package orders

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/cart"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/payments/paystack"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

type handlerFixture struct {
	ctx          context.Context
	pool         db.Pool
	router       http.Handler
	ordersSvc    *Service
	ordersHand   *Handlers
	authSvc      *auth.Service
	authHand     *auth.Handlers
	cartSvc      *cart.Service
	catalog      *catalog.Repository
	stub         *paystackStub
	sender       *captureSender
	secret       string
	cleanup      func()
}

func setupHandler(t *testing.T) *handlerFixture {
	t.Helper()
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()

	catRepo := catalog.NewRepository(pool)
	ship := shipping.New(shipping.Config{FlatRateGhsMinor: 2500, FreeOverGhsMinor: 100000})
	cartRepo := cart.NewRepository(pool)
	cartSvc := cart.NewService(cartRepo, catRepo, ship, logger)

	stub := newPaystackStub(t)
	const secret = "sk_test_handler"
	psClient := paystack.NewClient(stub.server.URL, secret)

	sender := &captureSender{}

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, sender, nil) // empty allowlist → auto-verify
	authSvc.Params = auth.TestParams
	authHand := auth.NewHandlers(authSvc, "rue_session", "", false)

	ordersRepo := NewRepository(pool)
	ordersSvc := NewService(ordersRepo, cartSvc, catRepo, ship, psClient, sender, pool, logger,
		"http://localhost:5173/checkout/return")
	ordersHand := NewHandlers(ordersSvc, secret, logger)

	r := chi.NewRouter()
	authHand.Mount(r)
	cartHand := cart.NewHandlers(cartSvc, authSvc, "rue_session", "", false)
	cartHand.Mount(r)
	ordersHand.MountPublic(r)
	r.Group(func(rr chi.Router) {
		rr.Use(authHand.RequireSession)
		ordersHand.MountAuthGated(rr)
	})

	return &handlerFixture{
		ctx:        ctx,
		pool:       pool,
		router:     r,
		ordersSvc:  ordersSvc,
		ordersHand: ordersHand,
		authSvc:    authSvc,
		authHand:   authHand,
		cartSvc:    cartSvc,
		catalog:    catRepo,
		stub:       stub,
		sender:     sender,
		secret:     secret,
		cleanup:    cleanup,
	}
}

func postJSONReq(t *testing.T, router http.Handler, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func getReq(t *testing.T, router http.Handler, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// signUpUser creates a user and returns their session cookie + user_id.
func signUpUser(t *testing.T, fx *handlerFixture, email, name string) (*http.Cookie, uuid.UUID) {
	t.Helper()
	rr := postJSONReq(t, fx.router, "/auth/signup", map[string]string{
		"email": email, "password": "hunter22", "name": name,
	}, nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("signup status %d: %s", rr.Code, rr.Body.String())
	}
	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "rue_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("no session cookie after signup")
	}
	var resp struct {
		UserID string `json:"user_id"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	id, err := uuid.Parse(resp.UserID)
	if err != nil {
		t.Fatalf("parse user_id from signup body: %v (body=%s)", err, rr.Body.String())
	}
	return sessionCookie, id
}

func addItemToUserCart(t *testing.T, fx *handlerFixture, cookie *http.Cookie, productID uuid.UUID, qty int32) {
	t.Helper()
	rr := postJSONReq(t, fx.router, "/cart/items",
		map[string]any{"product_id": productID.String(), "qty": qty}, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("add to cart status %d: %s", rr.Code, rr.Body.String())
	}
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestHandlerInitCheckout_Unauthenticated_Returns401(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	rr := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
	}, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerInitCheckout_HappyPath(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	cookie, _ := signUpUser(t, fx, "buyer@h.test", "Buyer")
	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addItemToUserCart(t, fx, cookie, productID, 1)

	rr := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
		"shipping_method":  "standard",
	}, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var resp initCheckoutResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.AuthorizationURL != "https://stub/auth" {
		t.Errorf("authorization_url = %q", resp.AuthorizationURL)
	}
	if !strings.HasPrefix(resp.Reference, "RUE-") {
		t.Errorf("reference = %q", resp.Reference)
	}
}

func TestHandlerInitCheckout_InvalidAddress_Returns400(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	cookie, _ := signUpUser(t, fx, "buyer-ia@h.test", "Buyer")
	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addItemToUserCart(t, fx, cookie, productID, 1)

	rr := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": ShippingAddress{Line1: "only-line1"},
	}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "validation_failed") {
		t.Errorf("expected validation_failed code; body: %s", rr.Body.String())
	}
}

func TestHandlerInitCheckout_EmptyCart_Returns400(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()
	cookie, _ := signUpUser(t, fx, "buyer-ec@h.test", "Buyer")

	rr := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
	}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerInitCheckout_PaystackNotConfigured_Returns503(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()
	// Force the client into "not configured" state.
	fx.ordersSvc.Paystack = paystack.NewClient("", "")

	cookie, _ := signUpUser(t, fx, "buyer-nc@h.test", "Buyer")
	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addItemToUserCart(t, fx, cookie, productID, 1)

	rr := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
	}, cookie)
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "upstream_error") {
		t.Errorf("expected upstream_error code; body: %s", rr.Body.String())
	}
}

func TestHandlerWebhook_InvalidSignature_Returns401(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	body := []byte(`{"event":"charge.success","data":{"reference":"RUE-FAKE0001","status":"success","id":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	req.Header.Set("x-paystack-signature", "not-a-valid-signature")
	rr := httptest.NewRecorder()
	fx.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401; body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerWebhook_ValidSignature_MarksPaid(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	cookie, _ := signUpUser(t, fx, "buyer-wh@h.test", "Buyer")
	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addItemToUserCart(t, fx, cookie, productID, 1)
	initRR := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
	}, cookie)
	if initRR.Code != http.StatusOK {
		t.Fatalf("init failed: %s", initRR.Body.String())
	}
	var initResp initCheckoutResponse
	_ = json.Unmarshal(initRR.Body.Bytes(), &initResp)

	body := []byte(fmt.Sprintf(
		`{"event":"charge.success","data":{"reference":"%s","status":"success","id":7777}}`,
		initResp.Reference))
	mac := hmac.New(sha512.New, []byte(fx.secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	req.Header.Set("x-paystack-signature", sig)
	rr := httptest.NewRecorder()
	fx.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("webhook status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	order, err := fx.ordersSvc.Repo.GetOrderByReference(fx.ctx, initResp.Reference)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.Status != "paid" {
		t.Errorf("order status = %q, want paid", order.Status)
	}
}

func TestHandlerWebhook_UnknownReference_Returns200(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	body := []byte(`{"event":"charge.success","data":{"reference":"RUE-NOEXIST","status":"success","id":1}}`)
	mac := hmac.New(sha512.New, []byte(fx.secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	req.Header.Set("x-paystack-signature", sig)
	rr := httptest.NewRecorder()
	fx.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (idempotent ack); body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerVerifyCheckout_OtherUser_Returns404(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	// User A creates an order.
	cookieA, _ := signUpUser(t, fx, "alice@h.test", "Alice")
	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addItemToUserCart(t, fx, cookieA, productID, 1)
	initRR := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
	}, cookieA)
	var initResp initCheckoutResponse
	_ = json.Unmarshal(initRR.Body.Bytes(), &initResp)

	// User B tries to verify A's reference.
	cookieB, _ := signUpUser(t, fx, "bob@h.test", "Bob")
	rr := getReq(t, fx.router, "/checkout/verify/"+initResp.Reference, cookieB)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (IDOR guard); body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerVerifyCheckout_OwnReference_Returns200(t *testing.T) {
	fx := setupHandler(t)
	defer fx.cleanup()

	cookie, _ := signUpUser(t, fx, "owner@h.test", "Owner")
	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addItemToUserCart(t, fx, cookie, productID, 1)
	initRR := postJSONReq(t, fx.router, "/checkout/init", map[string]any{
		"shipping_address": validAddress(),
	}, cookie)
	var initResp initCheckoutResponse
	_ = json.Unmarshal(initRR.Body.Bytes(), &initResp)

	rr := getReq(t, fx.router, "/checkout/verify/"+initResp.Reference, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var resp verifyCheckoutResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != "paid" {
		t.Errorf("status = %q, want paid (stub reports success)", resp.Status)
	}
}
