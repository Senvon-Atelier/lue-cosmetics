package main_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func TestServerBootsAndHealthzReturnsOK(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "..", "..")
	bin := filepath.Join(t.TempDir(), "api")
	build := exec.Command("go", "build", "-o", bin, "./cmd/api")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	// Build absolute path to the shipping config so the binary can find it
	// regardless of its working directory.
	shipConfigPath, err := filepath.Abs(filepath.Join(root, "seed", "config", "shipping_config.json"))
	if err != nil {
		t.Fatalf("shipping config abs: %v", err)
	}

	// Stubbed Paystack server. Started BEFORE the binary so its URL is in env.
	const paystackSecret = "sk_test_smoke"
	paystackStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/transaction/initialize":
			var in struct {
				Reference string `json:"reference"`
			}
			_ = json.NewDecoder(r.Body).Decode(&in)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"status":true,"data":{"authorization_url":%q,"access_code":"AC","reference":%q}}`,
				"https://stub.paystack.test/checkout/"+in.Reference, in.Reference)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/transaction/verify/"):
			ref := strings.TrimPrefix(r.URL.Path, "/transaction/verify/")
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"status":true,"data":{"reference":%q,"status":"success","amount":12500,"currency":"GHS","id":1234567}}`, ref)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer paystackStub.Close()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"PORT=18080",
		"ENV=development",
		"DATABASE_URL="+url,
		"CORS_ORIGINS=http://localhost:5173",
		"LOG_LEVEL=debug",
		"SHIPPING_CONFIG_PATH="+shipConfigPath,
		"PAYSTACK_BASE_URL="+paystackStub.URL,
		"PAYSTACK_SECRET_KEY="+paystackSecret,
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	deadline := time.Now().Add(10 * time.Second)
	var resp *http.Response
	for time.Now().Before(deadline) {
		resp, err = http.Get("http://127.0.0.1:18080/healthz")
		if err == nil && resp.StatusCode == 200 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil || resp == nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("healthz code = %d", resp.StatusCode)
		}
		t.Fatalf("healthz never returned 200: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	// Verify /api/v1/categories is reachable (empty array — no seed data).
	resp, err = http.Get("http://127.0.0.1:18080/api/v1/categories")
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("/api/v1/categories code = %d", resp.StatusCode)
		}
		t.Fatalf("/api/v1/categories failed: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	// Verify /api/v1/shipping/quote returns a valid quote for a below-threshold subtotal.
	resp, err = http.Get("http://127.0.0.1:18080/api/v1/shipping/quote?subtotal=10000")
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("/shipping/quote code = %d", resp.StatusCode)
		}
		t.Fatalf("/shipping/quote failed: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	// POST /api/v1/auth/signup → 201, capture cookie
	signupBody := strings.NewReader(`{"email":"smoke@y.test","password":"hunter22","name":"Smoke"}`)
	signupReq, _ := http.NewRequest("POST", "http://127.0.0.1:18080/api/v1/auth/signup", signupBody)
	signupReq.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(signupReq)
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("signup code = %d", resp.StatusCode)
	}
	sessionCookie := testsupport.FindCookie(resp, "rue_session")
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if sessionCookie == nil {
		t.Fatal("no rue_session cookie set on signup")
	}

	// GET /api/v1/me with cookie → 200
	meReq, _ := http.NewRequest("GET", "http://127.0.0.1:18080/api/v1/me", nil)
	meReq.AddCookie(sessionCookie)
	resp, err = http.DefaultClient.Do(meReq)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("me code = %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// ---- Cart smoke: guest → seed a product → add → signup → merge → confirm in user cart ----
	serverURL := "http://127.0.0.1:18080"
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("seed pool: %v", err)
	}
	defer pool.Close()

	brandID := uuid.New()
	categoryID := uuid.New()
	productID := uuid.New()
	if _, err := pool.Exec(ctx,
		"INSERT INTO brands (id, slug, name) VALUES ($1, $2, $3)",
		brandID, "smoke-brand", "Smoke Brand"); err != nil {
		t.Fatalf("seed brand: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO categories (id, slug, label) VALUES ($1, $2, $3)",
		categoryID, "smoke-cat", "Smoke Category"); err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO products (id, slug, name, brand_id, category_id, price_ghs_minor) VALUES ($1, $2, $3, $4, $5, $6)",
		productID, "smoke-prod", "Smoke Product", brandID, categoryID, int64(12500)); err != nil {
		t.Fatalf("seed product: %v", err)
	}

	// 1. Anonymous GET /api/v1/cart → mints rue_guest_cart.
	resp, err = http.Get(serverURL + "/api/v1/cart")
	if err != nil {
		t.Fatalf("anon GET /cart: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("anon GET /cart status = %d, want 200", resp.StatusCode)
	}
	guestCookie := testsupport.FindCookie(resp, "rue_guest_cart")
	if guestCookie == nil || guestCookie.Value == "" {
		t.Fatalf("anon GET /cart: no rue_guest_cart cookie set")
	}
	guestToken := guestCookie.Value
	_ = resp.Body.Close()

	// 2. POST /api/v1/cart/items with guest cookie.
	addBody := fmt.Sprintf(`{"product_id":"%s","qty":2}`, productID)
	addReq, _ := http.NewRequest("POST", serverURL+"/api/v1/cart/items", strings.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.AddCookie(guestCookie)
	resp, err = http.DefaultClient.Do(addReq)
	if err != nil {
		t.Fatalf("POST /cart/items: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("POST /cart/items status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 3. Sign up a fresh user; capture rue_session.
	signupEmail := fmt.Sprintf("smoke-merge-%s@y.test", uuid.NewString()[:8])
	mergeSignupBody := fmt.Sprintf(`{"email":"%s","password":"hunter22","name":"Smoke"}`, signupEmail)
	mergeSignupReq, _ := http.NewRequest("POST", serverURL+"/api/v1/auth/signup", strings.NewReader(mergeSignupBody))
	mergeSignupReq.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(mergeSignupReq)
	if err != nil {
		t.Fatalf("merge-signup: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("merge-signup status = %d, want 201", resp.StatusCode)
	}
	mergeSession := testsupport.FindCookie(resp, "rue_session")
	if mergeSession == nil {
		t.Fatalf("merge-signup: no rue_session cookie")
	}
	_ = resp.Body.Close()

	// 4. POST /api/v1/cart/merge.
	mergeBody := fmt.Sprintf(`{"guest_token":"%s"}`, guestToken)
	mergeReq, _ := http.NewRequest("POST", serverURL+"/api/v1/cart/merge", strings.NewReader(mergeBody))
	mergeReq.Header.Set("Content-Type", "application/json")
	mergeReq.AddCookie(mergeSession)
	resp, err = http.DefaultClient.Do(mergeReq)
	if err != nil {
		t.Fatalf("POST /cart/merge: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("POST /cart/merge status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 5. GET /api/v1/cart with the user's session — the merged item must be there.
	getReq, _ := http.NewRequest("GET", serverURL+"/api/v1/cart", nil)
	getReq.AddCookie(mergeSession)
	resp, err = http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("GET /cart (user): %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("GET /cart (user) status = %d, want 200", resp.StatusCode)
	}
	var cartView struct {
		Items []struct {
			ProductID string `json:"product_id"`
			Qty       int    `json:"qty"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cartView); err != nil {
		t.Fatalf("decode user cart: %v", err)
	}
	_ = resp.Body.Close()
	if len(cartView.Items) != 1 {
		t.Fatalf("user cart items = %d, want 1", len(cartView.Items))
	}
	if cartView.Items[0].ProductID != productID.String() {
		t.Errorf("merged item product = %s, want %s", cartView.Items[0].ProductID, productID.String())
	}
	if cartView.Items[0].Qty != 2 {
		t.Errorf("merged item qty = %d, want 2", cartView.Items[0].Qty)
	}

	// ---- Checkout smoke: init → simulated signed webhook → DB-asserts ----

	// 6. POST /api/v1/checkout/init with a stub shipping_address.
	checkoutInitBody := `{
		"shipping_address": {
			"line1": "1 Smoke Lane",
			"city": "Accra",
			"region": "Greater Accra",
			"phone": "+233200000000",
			"label": "Home"
		},
		"shipping_method": "standard"
	}`
	initReq, _ := http.NewRequest("POST", serverURL+"/api/v1/checkout/init", strings.NewReader(checkoutInitBody))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.AddCookie(mergeSession)
	resp, err = http.DefaultClient.Do(initReq)
	if err != nil {
		t.Fatalf("POST /checkout/init: %v", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("POST /checkout/init status = %d, want 200; body: %s", resp.StatusCode, body)
	}
	var initResp struct {
		OrderID          string `json:"order_id"`
		Reference        string `json:"reference"`
		AuthorizationURL string `json:"authorization_url"`
		TotalGhsMinor    int64  `json:"total_ghs_minor"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		t.Fatalf("decode /checkout/init: %v", err)
	}
	_ = resp.Body.Close()
	if !strings.HasPrefix(initResp.AuthorizationURL, paystackStub.URL) &&
		!strings.HasPrefix(initResp.AuthorizationURL, "https://stub.paystack.test") {
		t.Errorf("authorization_url = %q, want stub-prefixed", initResp.AuthorizationURL)
	}
	if !strings.HasPrefix(initResp.Reference, "RUE-") {
		t.Errorf("reference = %q, want RUE-XXXXXXXX", initResp.Reference)
	}

	// 7. Simulate a signed Paystack webhook for that reference.
	webhookBody := []byte(fmt.Sprintf(
		`{"event":"charge.success","data":{"reference":"%s","status":"success","amount":%d,"id":1234567}}`,
		initResp.Reference, initResp.TotalGhsMinor))
	mac := hmac.New(sha512.New, []byte(paystackSecret))
	mac.Write(webhookBody)
	signature := hex.EncodeToString(mac.Sum(nil))

	whReq, _ := http.NewRequest("POST", serverURL+"/api/v1/webhooks/paystack", bytes.NewReader(webhookBody))
	whReq.Header.Set("Content-Type", "application/json")
	whReq.Header.Set("x-paystack-signature", signature)
	resp, err = http.DefaultClient.Do(whReq)
	if err != nil {
		t.Fatalf("POST /webhooks/paystack: %v", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("webhook status = %d, want 200; body: %s", resp.StatusCode, body)
	}
	_ = resp.Body.Close()

	// 8. DB-asserts via the side pool: order status = paid, cart_items empty.
	var orderStatus string
	if err := pool.QueryRow(ctx,
		"SELECT status FROM orders WHERE paystack_reference = $1", initResp.Reference).Scan(&orderStatus); err != nil {
		t.Fatalf("query order status: %v", err)
	}
	if orderStatus != "paid" {
		t.Errorf("DB order status = %q, want paid", orderStatus)
	}

	// Fetch the user_id of the order to assert cart_items cleared for them.
	var orderUserID uuid.UUID
	if err := pool.QueryRow(ctx,
		"SELECT user_id FROM orders WHERE paystack_reference = $1", initResp.Reference).Scan(&orderUserID); err != nil {
		t.Fatalf("query order user: %v", err)
	}
	var cartItemCount int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM cart_items ci JOIN carts c ON ci.cart_id = c.id WHERE c.user_id = $1",
		orderUserID).Scan(&cartItemCount); err != nil {
		t.Fatalf("count cart_items: %v", err)
	}
	if cartItemCount != 0 {
		t.Errorf("post-webhook cart_items = %d, want 0", cartItemCount)
	}
}
