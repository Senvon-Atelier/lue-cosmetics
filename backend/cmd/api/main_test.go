package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"PORT=18080",
		"ENV=development",
		"DATABASE_URL="+url,
		"CORS_ORIGINS=http://localhost:5173",
		"LOG_LEVEL=debug",
		"SHIPPING_CONFIG_PATH="+shipConfigPath,
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
}
