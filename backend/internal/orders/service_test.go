package orders

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/cart"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/payments/paystack"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

type captureSender struct {
	mu    sync.Mutex
	calls []sendCall
	err   error
}

type sendCall struct {
	To       string
	Template string
	Data     map[string]any
}

func (c *captureSender) Send(_ context.Context, to, template string, data map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, sendCall{To: to, Template: template, Data: data})
	return c.err
}

func (c *captureSender) snapshot() []sendCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]sendCall, len(c.calls))
	copy(out, c.calls)
	return out
}

// paystackStub returns a configurable httptest.NewServer that mimics the
// Paystack endpoints we hit.
type paystackStub struct {
	server          *httptest.Server
	initStatus      int
	initBody        string
	verifyStatus    int
	verifyBody      string
	initCalled      int
	verifyCalled    int
	lastInitRequest paystack.InitializeTransactionInput
}

func newPaystackStub(t *testing.T) *paystackStub {
	t.Helper()
	s := &paystackStub{
		initStatus:   http.StatusOK,
		initBody:     `{"status":true,"data":{"authorization_url":"https://stub/auth","access_code":"AC","reference":"%s"}}`,
		verifyStatus: http.StatusOK,
		verifyBody:   `{"status":true,"data":{"reference":"%s","status":"success","amount":12500,"currency":"GHS","id":9999}}`,
	}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/transaction/initialize":
			s.initCalled++
			_ = json.NewDecoder(r.Body).Decode(&s.lastInitRequest)
			w.WriteHeader(s.initStatus)
			_, _ = w.Write([]byte(formatBody(s.initBody, s.lastInitRequest.Reference)))
		case r.Method == http.MethodGet && len(r.URL.Path) > len("/transaction/verify/"):
			s.verifyCalled++
			ref := r.URL.Path[len("/transaction/verify/"):]
			w.WriteHeader(s.verifyStatus)
			_, _ = w.Write([]byte(formatBody(s.verifyBody, ref)))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.server.Close)
	return s
}

func formatBody(template, ref string) string {
	// crude printf-style substitution: replace %s once with ref.
	out := []byte(template)
	for i := 0; i < len(out)-1; i++ {
		if out[i] == '%' && out[i+1] == 's' {
			return string(out[:i]) + ref + string(out[i+2:])
		}
	}
	return template
}

type orderFixture struct {
	ctx        context.Context
	pool       db.Pool
	svc        *Service
	sender     *captureSender
	stub       *paystackStub
	userID     uuid.UUID
	userEmail  string
	userName   string
	cartSvc    *cart.Service
	catalog    *catalog.Repository
	cleanup    func()
}

func setup(t *testing.T) *orderFixture {
	t.Helper()
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")

	logger := zap.NewNop()
	catRepo := catalog.NewRepository(pool)
	ship := shipping.New(shipping.Config{FlatRateGhsMinor: 2500, FreeOverGhsMinor: 100000})
	cartRepo := cart.NewRepository(pool)
	cartSvc := cart.NewService(cartRepo, catRepo, ship, logger)

	stub := newPaystackStub(t)
	client := paystack.NewClient(stub.server.URL, "sk_test_smoke")

	sender := &captureSender{}

	repo := NewRepository(pool)
	svc := NewService(repo, cartSvc, catRepo, ship, client, sender, pool, logger, "http://localhost:5173/checkout/return")

	userID, email, name := seedUserRow(t, ctx, pool)

	return &orderFixture{
		ctx:       ctx,
		pool:      pool,
		svc:       svc,
		sender:    sender,
		stub:      stub,
		userID:    userID,
		userEmail: email,
		userName:  name,
		cartSvc:   cartSvc,
		catalog:   catRepo,
		cleanup:   cleanup,
	}
}

func seedUserRow(t *testing.T, ctx context.Context, pool db.Pool) (uuid.UUID, string, string) {
	t.Helper()
	id := uuid.New()
	email := "buyer-" + id.String()[:8] + "@example.com"
	name := "Test Buyer"
	if _, err := pool.Exec(ctx,
		"INSERT INTO users (id, email, name) VALUES ($1, $2, $3)",
		id, email, name); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id, email, name
}

func seedProductWithBrand(t *testing.T, ctx context.Context, pool db.Pool, price int64) (uuid.UUID, string, string) {
	t.Helper()
	brandID := uuid.New()
	categoryID := uuid.New()
	productID := uuid.New()
	brandName := "Brand-" + brandID.String()[:6]
	productName := "Product-" + productID.String()[:6]

	if _, err := pool.Exec(ctx,
		"INSERT INTO brands (id, slug, name) VALUES ($1, $2, $3)",
		brandID, "brand-"+brandID.String()[:8], brandName); err != nil {
		t.Fatalf("seed brand: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO categories (id, slug, label) VALUES ($1, $2, $3)",
		categoryID, "cat-"+categoryID.String()[:8], "Cat"); err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO products (id, slug, name, brand_id, category_id, price_ghs_minor, image_path) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		productID, "prod-"+productID.String()[:8], productName, brandID, categoryID, price, "/img/"+productID.String()[:6]+".png"); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	return productID, productName, brandName
}

func setProductPrice(t *testing.T, ctx context.Context, pool db.Pool, productID uuid.UUID, price int64) {
	t.Helper()
	if _, err := pool.Exec(ctx, "UPDATE products SET price_ghs_minor = $1 WHERE id = $2", price, productID); err != nil {
		t.Fatalf("update price: %v", err)
	}
}

func validAddress() ShippingAddress {
	return ShippingAddress{
		Line1: "123 Main", City: "Accra", Region: "Greater Accra", Phone: "+233200000000", Label: "Home",
	}
}

func addToCart(t *testing.T, fx *orderFixture, productID uuid.UUID, qty int32) {
	t.Helper()
	if _, err := fx.cartSvc.AddItem(fx.ctx, cart.CartIdentity{UserID: fx.userID}, productID, qty); err != nil {
		t.Fatalf("add item to cart: %v", err)
	}
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestInitCheckout_HappyPath(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	p1, p1Name, p1Brand := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	p2, p2Name, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 20000)
	addToCart(t, fx, p1, 2)
	addToCart(t, fx, p2, 1)

	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(), ShippingMethod: "standard",
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}
	if out.AuthorizationURL != "https://stub/auth" {
		t.Errorf("authorization_url = %q", out.AuthorizationURL)
	}
	if out.Reference == "" {
		t.Errorf("reference empty")
	}
	wantTotal := int64(2*10000 + 1*20000 + 2500)
	if out.TotalGhsMinor != wantTotal {
		t.Errorf("total = %d, want %d", out.TotalGhsMinor, wantTotal)
	}

	order, err := fx.svc.Repo.GetOrderByReference(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.Status != "pending" {
		t.Errorf("status = %q, want pending", order.Status)
	}
	items, err := fx.svc.Repo.ListOrderItems(fx.ctx, order.ID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	// Snapshots reflect seeded names/brands.
	foundP1 := false
	for _, it := range items {
		if it.ProductID == p1 {
			foundP1 = true
			if it.ProductNameSnapshot != p1Name {
				t.Errorf("p1 name snapshot = %q", it.ProductNameSnapshot)
			}
			if it.ProductBrandSnapshot != p1Brand {
				t.Errorf("p1 brand snapshot = %q", it.ProductBrandSnapshot)
			}
			if it.UnitPriceGhsMinor != 10000 || it.Qty != 2 {
				t.Errorf("p1 unit/qty = %d/%d", it.UnitPriceGhsMinor, it.Qty)
			}
		}
		if it.ProductID == p2 && it.ProductNameSnapshot != p2Name {
			t.Errorf("p2 name snapshot = %q", it.ProductNameSnapshot)
		}
	}
	if !foundP1 {
		t.Errorf("p1 not in order items")
	}
}

func TestInitCheckout_EmptyCart_ReturnsErrEmptyCart(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	_, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if !errors.Is(err, ErrEmptyCart) {
		t.Errorf("err = %v, want ErrEmptyCart", err)
	}
}

func TestInitCheckout_RePricesFromCatalog(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)

	// Bump the catalog price AFTER the cart snapshot.
	setProductPrice(t, fx.ctx, fx.pool, productID, 17500)

	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}

	order, err := fx.svc.Repo.GetOrderByReference(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.SubtotalGhsMinor != 17500 {
		t.Errorf("subtotal = %d, want 17500 (fresh price)", order.SubtotalGhsMinor)
	}
	items, _ := fx.svc.Repo.ListOrderItems(fx.ctx, order.ID)
	if len(items) != 1 || items[0].UnitPriceGhsMinor != 17500 {
		t.Errorf("order_item unit price = %d, want 17500 (fresh from catalog)", items[0].UnitPriceGhsMinor)
	}
}

func TestInitCheckout_PaystackInitFails_OrderRemainsPending(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()
	fx.stub.initStatus = http.StatusInternalServerError
	fx.stub.initBody = `{"status":false,"message":"upstream down"}`

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)

	_, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, paystack.ErrPaystackUpstream) {
		t.Errorf("err = %v, want ErrPaystackUpstream wrap", err)
	}
	// Order row should exist with status=pending despite the upstream failure.
	count, err := fx.svc.Repo.CountOrdersByStatus(fx.ctx, "pending")
	if err != nil {
		t.Fatalf("count pending: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 pending order, got %d", count)
	}
}

func TestMarkPaid_IdempotentSecondCall(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)
	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}

	if err := fx.svc.MarkPaid(fx.ctx, out.Reference, "12345"); err != nil {
		t.Fatalf("first MarkPaid: %v", err)
	}
	if err := fx.svc.MarkPaid(fx.ctx, out.Reference, "12345"); err != nil {
		t.Fatalf("second MarkPaid: %v", err)
	}

	order, err := fx.svc.Repo.GetOrderByReference(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.Status != "paid" {
		t.Errorf("status = %q, want paid", order.Status)
	}
	calls := fx.sender.snapshot()
	confirmCount := 0
	for _, c := range calls {
		if c.Template == "order_confirmation" {
			confirmCount++
		}
	}
	if confirmCount != 1 {
		t.Errorf("expected exactly 1 order_confirmation email, got %d", confirmCount)
	}
}

func TestMarkPaid_DeletesUserCartItems(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 2)
	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}

	if err := fx.svc.MarkPaid(fx.ctx, out.Reference, "12345"); err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}

	// Cart items gone.
	var itemCount int
	if err := fx.pool.QueryRow(fx.ctx,
		"SELECT count(*) FROM cart_items ci JOIN carts c ON ci.cart_id = c.id WHERE c.user_id = $1",
		fx.userID).Scan(&itemCount); err != nil {
		t.Fatalf("count cart items: %v", err)
	}
	if itemCount != 0 {
		t.Errorf("expected cart_items cleared, got %d", itemCount)
	}
	// Cart row remains.
	var cartCount int
	if err := fx.pool.QueryRow(fx.ctx,
		"SELECT count(*) FROM carts WHERE user_id = $1", fx.userID).Scan(&cartCount); err != nil {
		t.Fatalf("count carts: %v", err)
	}
	if cartCount != 1 {
		t.Errorf("expected cart row preserved, got %d", cartCount)
	}
}

func TestMarkPaid_EmailFailureDoesNotRollback(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()
	fx.sender.err = errors.New("smtp blew up")

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)
	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}

	if err := fx.svc.MarkPaid(fx.ctx, out.Reference, "12345"); err != nil {
		t.Fatalf("MarkPaid (with email failure) returned err: %v", err)
	}
	order, err := fx.svc.Repo.GetOrderByReference(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.Status != "paid" {
		t.Errorf("status = %q, want paid (commit must not roll back)", order.Status)
	}
}

func TestVerifyCheckout_OrderAlreadyPaid_ShortCircuits(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)
	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}
	if err := fx.svc.MarkPaid(fx.ctx, out.Reference, "12345"); err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}

	verifyCallsBefore := fx.stub.verifyCalled
	status, err := fx.svc.VerifyCheckout(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("VerifyCheckout: %v", err)
	}
	if status != "paid" {
		t.Errorf("status = %q, want paid", status)
	}
	if fx.stub.verifyCalled != verifyCallsBefore {
		t.Errorf("expected Paystack verify to NOT be called when order already paid")
	}
}

func TestVerifyCheckout_PaystackReportsSuccess_CallsMarkPaid(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)
	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}

	status, err := fx.svc.VerifyCheckout(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("VerifyCheckout: %v", err)
	}
	if status != "paid" {
		t.Errorf("status = %q, want paid", status)
	}
	order, _ := fx.svc.Repo.GetOrderByReference(fx.ctx, out.Reference)
	if order.Status != "paid" {
		t.Errorf("DB status = %q, want paid", order.Status)
	}
	if order.PaystackTransactionID == nil || *order.PaystackTransactionID != strconv.Itoa(9999) {
		t.Errorf("paystack_transaction_id = %v, want 9999", order.PaystackTransactionID)
	}
}

func TestVerifyCheckout_PaystackReportsFailure_StaysPending(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()
	fx.stub.verifyBody = `{"status":true,"data":{"reference":"%s","status":"failed","amount":12500,"currency":"GHS","id":1}}`

	productID, _, _ := seedProductWithBrand(t, fx.ctx, fx.pool, 10000)
	addToCart(t, fx, productID, 1)
	out, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: validAddress(),
	})
	if err != nil {
		t.Fatalf("InitCheckout: %v", err)
	}
	status, err := fx.svc.VerifyCheckout(fx.ctx, out.Reference)
	if err != nil {
		t.Fatalf("VerifyCheckout: %v", err)
	}
	if status != "pending" {
		t.Errorf("status = %q, want pending", status)
	}
	order, _ := fx.svc.Repo.GetOrderByReference(fx.ctx, out.Reference)
	if order.Status != "pending" {
		t.Errorf("DB status = %q, want pending", order.Status)
	}
}

// Address validation guard — ErrInvalidAddress branch coverage.
func TestInitCheckout_InvalidAddress_ReturnsErrInvalidAddress(t *testing.T) {
	fx := setup(t)
	defer fx.cleanup()

	_, err := fx.svc.InitCheckout(fx.ctx, InitCheckoutInput{
		UserID: fx.userID, UserEmail: fx.userEmail, UserName: fx.userName,
		ShippingAddress: ShippingAddress{Line1: "123 Main"}, // missing city/region/phone
	})
	if !errors.Is(err, ErrInvalidAddress) {
		t.Errorf("err = %v, want ErrInvalidAddress", err)
	}
}
