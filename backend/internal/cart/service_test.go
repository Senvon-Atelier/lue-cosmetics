package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// testShippingConfig mirrors backend/config/shipping_config.json so the
// integration tests don't need filesystem access.
func testShippingConfig() shipping.Config {
	return shipping.Config{FlatRateGhsMinor: 2500, FreeOverGhsMinor: 50000}
}

func newCartService(t *testing.T, pool db.Pool) *Service {
	t.Helper()
	return NewService(
		NewRepository(pool),
		catalog.NewRepository(pool),
		shipping.New(testShippingConfig()),
		zap.NewNop(),
	)
}

func seedTestUser(t *testing.T, ctx context.Context, pool db.Pool, email string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, "INSERT INTO users (id, email, name) VALUES ($1, $2, 'Test')", id, email); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

// updateProductPrice rewrites a product's price in place so tests can verify
// the at-add-time snapshot survives a price change.
func updateProductPrice(t *testing.T, ctx context.Context, pool db.Pool, productID uuid.UUID, price int64) {
	t.Helper()
	if _, err := pool.Exec(ctx, "UPDATE products SET price_ghs_minor = $1 WHERE id = $2", price, productID); err != nil {
		t.Fatalf("update price: %v", err)
	}
}

func TestService_GetOrCreate_EmptyIdentity_MintsGuestCart(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	view, token, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	if token == "" {
		t.Error("expected minted guest token, got empty")
	}
	if view.GuestToken != token {
		t.Errorf("view.GuestToken = %q, want %q", view.GuestToken, token)
	}
	if len(view.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(view.Items))
	}
	if view.SubtotalGhsMinor != 0 || view.TotalGhsMinor != 0 {
		t.Errorf("expected zero totals, got sub=%d total=%d", view.SubtotalGhsMinor, view.TotalGhsMinor)
	}
	if view.FreeShippingRemainderGhsMinor != 50000 {
		t.Errorf("FreeShippingRemainderGhsMinor = %d, want 50000",
			view.FreeShippingRemainderGhsMinor)
	}
}

func TestService_GetOrCreate_GuestToken_ReusesCart(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	v1, token, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	v2, minted, err := svc.GetOrCreate(ctx, CartIdentity{GuestToken: token})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if minted != "" {
		t.Errorf("expected no minting on reuse, got %q", minted)
	}
	if v1.CartID != v2.CartID {
		t.Errorf("cart ids differ: %s vs %s", v1.CartID, v2.CartID)
	}
}

func TestService_GetOrCreate_UserID_MintsThenReuses(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)
	userID := seedTestUser(t, ctx, pool, "user@cart.test")

	v1, token, err := svc.GetOrCreate(ctx, CartIdentity{UserID: userID})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if token != "" {
		t.Errorf("user cart shouldn't mint guest token, got %q", token)
	}
	v2, _, err := svc.GetOrCreate(ctx, CartIdentity{UserID: userID})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if v1.CartID != v2.CartID {
		t.Errorf("cart ids differ on second user GetOrCreate")
	}
}

func TestService_AddItem_SnapshotsPriceAtAddTime(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool) // seeded at 10000
	_, token, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}
	id := CartIdentity{GuestToken: token}

	view, err := svc.AddItem(ctx, id, productID, 1)
	if err != nil {
		t.Fatalf("first add: %v", err)
	}
	if got := view.Items[0].UnitPriceGhsMinor; got != 10000 {
		t.Errorf("unit price after first add = %d, want 10000", got)
	}

	// Change the catalog price.
	updateProductPrice(t, ctx, pool, productID, 99999)

	view, err = svc.AddItem(ctx, id, productID, 1)
	if err != nil {
		t.Fatalf("second add: %v", err)
	}
	if got := view.Items[0].UnitPriceGhsMinor; got != 10000 {
		t.Errorf("unit price after second add = %d, want 10000 (snapshot preserved)", got)
	}
	if got := view.Items[0].Qty; got != 2 {
		t.Errorf("qty = %d, want 2", got)
	}
}

func TestService_AddItem_UpsertSumsQty(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool)
	_, token, _ := svc.GetOrCreate(ctx, CartIdentity{})
	id := CartIdentity{GuestToken: token}

	if _, err := svc.AddItem(ctx, id, productID, 1); err != nil {
		t.Fatalf("add 1: %v", err)
	}
	view, err := svc.AddItem(ctx, id, productID, 2)
	if err != nil {
		t.Fatalf("add 2: %v", err)
	}
	if got := view.Items[0].Qty; got != 3 {
		t.Errorf("qty = %d, want 3", got)
	}
}

func TestService_AddItem_UnknownProduct_ReturnsErrUnknownProduct(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	_, err := svc.AddItem(ctx, CartIdentity{}, uuid.New(), 1)
	if !errors.Is(err, ErrUnknownProduct) {
		t.Errorf("err = %v, want ErrUnknownProduct", err)
	}
}

func TestService_UpdateQty_Happy(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool)
	_, token, _ := svc.GetOrCreate(ctx, CartIdentity{})
	id := CartIdentity{GuestToken: token}
	view, _ := svc.AddItem(ctx, id, productID, 1)
	itemID := view.Items[0].ID

	view, err := svc.UpdateQty(ctx, id, itemID, 5)
	if err != nil {
		t.Fatalf("UpdateQty: %v", err)
	}
	if view.Items[0].Qty != 5 {
		t.Errorf("qty = %d, want 5", view.Items[0].Qty)
	}
}

func TestService_UpdateQty_CrossCart_ReturnsErrItemNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool)

	// Cart A holds the item.
	_, tokenA, _ := svc.GetOrCreate(ctx, CartIdentity{})
	idA := CartIdentity{GuestToken: tokenA}
	viewA, _ := svc.AddItem(ctx, idA, productID, 2)
	itemID := viewA.Items[0].ID

	// Cart B exists but doesn't own that item.
	_, tokenB, _ := svc.GetOrCreate(ctx, CartIdentity{})
	idB := CartIdentity{GuestToken: tokenB}

	_, err := svc.UpdateQty(ctx, idB, itemID, 9)
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("err = %v, want ErrItemNotFound", err)
	}
}

func TestService_RemoveItem_CrossCart_ReturnsErrItemNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool)
	_, tokenA, _ := svc.GetOrCreate(ctx, CartIdentity{})
	idA := CartIdentity{GuestToken: tokenA}
	viewA, _ := svc.AddItem(ctx, idA, productID, 2)
	itemID := viewA.Items[0].ID

	_, tokenB, _ := svc.GetOrCreate(ctx, CartIdentity{})
	idB := CartIdentity{GuestToken: tokenB}

	_, err := svc.RemoveItem(ctx, idB, itemID)
	if !errors.Is(err, ErrItemNotFound) {
		t.Errorf("err = %v, want ErrItemNotFound", err)
	}
}

func TestService_AddItem_InvalidQty(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	_, err := svc.AddItem(ctx, CartIdentity{}, uuid.New(), 0)
	if !errors.Is(err, ErrInvalidQty) {
		t.Errorf("err = %v, want ErrInvalidQty", err)
	}
}

func TestService_View_IncludesShippingFromService(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool) // price 10000
	_, token, _ := svc.GetOrCreate(ctx, CartIdentity{})
	id := CartIdentity{GuestToken: token}

	view, err := svc.AddItem(ctx, id, productID, 1)
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	// subtotal 10000 < free_over 50000 → flat rate 2500 applies.
	expected := shipping.New(testShippingConfig()).Quote(10000)
	if view.SubtotalGhsMinor != 10000 {
		t.Errorf("subtotal = %d, want 10000", view.SubtotalGhsMinor)
	}
	if view.ShippingCostGhsMinor != expected.AppliedCostGhsMinor {
		t.Errorf("shipping = %d, want %d", view.ShippingCostGhsMinor, expected.AppliedCostGhsMinor)
	}
	if view.FreeShippingRemainderGhsMinor != expected.FreeShippingRemainderGhsMinor {
		t.Errorf("free remainder = %d, want %d", view.FreeShippingRemainderGhsMinor, expected.FreeShippingRemainderGhsMinor)
	}
	if view.TotalGhsMinor != view.SubtotalGhsMinor+view.ShippingCostGhsMinor {
		t.Errorf("total mismatch: %d != %d+%d", view.TotalGhsMinor, view.SubtotalGhsMinor, view.ShippingCostGhsMinor)
	}
}

func TestService_RemoveItem_Happy_EmptiesCart(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)

	productID := seedTestProduct(t, ctx, pool)
	_, token, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	id := CartIdentity{GuestToken: token}

	addView, err := svc.AddItem(ctx, id, productID, 2)
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if len(addView.Items) != 1 {
		t.Fatalf("after AddItem: len(Items) = %d, want 1", len(addView.Items))
	}
	itemID := addView.Items[0].ID

	removeView, err := svc.RemoveItem(ctx, id, itemID)
	if err != nil {
		t.Fatalf("RemoveItem: %v", err)
	}
	if len(removeView.Items) != 0 {
		t.Errorf("after RemoveItem: len(Items) = %d, want 0", len(removeView.Items))
	}
	if removeView.SubtotalGhsMinor != 0 {
		t.Errorf("subtotal = %d, want 0", removeView.SubtotalGhsMinor)
	}
	if removeView.TotalGhsMinor != 0 {
		t.Errorf("total = %d, want 0", removeView.TotalGhsMinor)
	}
	if removeView.FreeShippingRemainderGhsMinor != 50000 {
		t.Errorf("FreeShippingRemainderGhsMinor = %d, want 50000",
			removeView.FreeShippingRemainderGhsMinor)
	}
}

// ── MergeGuestCart ──────────────────────────────────────────────────────────

func TestService_MergeGuestCart_EmptyToken_ReturnsUserCart(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)
	userID := seedTestUser(t, ctx, pool, "empty-token@merge.test")
	productID := seedTestProduct(t, ctx, pool)

	if _, err := svc.AddItem(ctx, CartIdentity{UserID: userID}, productID, 2); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	view, err := svc.MergeGuestCart(ctx, userID, "")
	if err != nil {
		t.Fatalf("MergeGuestCart: %v", err)
	}
	if len(view.Items) != 1 || view.Items[0].Qty != 2 {
		t.Errorf("user cart should be unchanged; got %+v", view.Items)
	}
}

func TestService_MergeGuestCart_UnknownToken_IsNoop(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)
	userID := seedTestUser(t, ctx, pool, "unknown-token@merge.test")
	productID := seedTestProduct(t, ctx, pool)

	if _, err := svc.AddItem(ctx, CartIdentity{UserID: userID}, productID, 3); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	view, err := svc.MergeGuestCart(ctx, userID, uuid.NewString())
	if err != nil {
		t.Fatalf("MergeGuestCart: %v", err)
	}
	if len(view.Items) != 1 || view.Items[0].Qty != 3 {
		t.Errorf("user cart should be unchanged on unknown token; got %+v", view.Items)
	}
}

func TestService_MergeGuestCart_AddsNewProducts(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)
	userID := seedTestUser(t, ctx, pool, "addsnew@merge.test")
	p1 := seedTestProduct(t, ctx, pool)
	p2 := seedTestProduct(t, ctx, pool)

	// Guest cart with both products.
	_, gtoken, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("guest GetOrCreate: %v", err)
	}
	gid := CartIdentity{GuestToken: gtoken}
	if _, err := svc.AddItem(ctx, gid, p1, 1); err != nil {
		t.Fatalf("guest add p1: %v", err)
	}
	if _, err := svc.AddItem(ctx, gid, p2, 4); err != nil {
		t.Fatalf("guest add p2: %v", err)
	}

	view, err := svc.MergeGuestCart(ctx, userID, gtoken)
	if err != nil {
		t.Fatalf("MergeGuestCart: %v", err)
	}
	if len(view.Items) != 2 {
		t.Fatalf("user cart len = %d, want 2; items=%+v", len(view.Items), view.Items)
	}
	got := map[uuid.UUID]int32{}
	for _, it := range view.Items {
		got[it.ProductID] = it.Qty
	}
	if got[p1] != 1 || got[p2] != 4 {
		t.Errorf("qty map = %v, want p1=1 p2=4", got)
	}
}

func TestService_MergeGuestCart_SumsQtyAndKeepsUserPrice(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)
	userID := seedTestUser(t, ctx, pool, "sumprice@merge.test")
	productID := seedTestProduct(t, ctx, pool) // price 10000

	// User cart: add 1 at price 10000 (snapshot).
	if _, err := svc.AddItem(ctx, CartIdentity{UserID: userID}, productID, 1); err != nil {
		t.Fatalf("user AddItem: %v", err)
	}

	// Bump price to 99999 so guest snapshot will be different.
	updateProductPrice(t, ctx, pool, productID, 99999)

	_, gtoken, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("guest GetOrCreate: %v", err)
	}
	gid := CartIdentity{GuestToken: gtoken}
	if _, err := svc.AddItem(ctx, gid, productID, 2); err != nil {
		t.Fatalf("guest AddItem: %v", err)
	}

	view, err := svc.MergeGuestCart(ctx, userID, gtoken)
	if err != nil {
		t.Fatalf("MergeGuestCart: %v", err)
	}
	if len(view.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(view.Items))
	}
	if view.Items[0].Qty != 3 {
		t.Errorf("qty = %d, want 3 (1 user + 2 guest)", view.Items[0].Qty)
	}
	if view.Items[0].UnitPriceGhsMinor != 10000 {
		t.Errorf("unit price = %d, want 10000 (user's price wins)", view.Items[0].UnitPriceGhsMinor)
	}
}

func TestService_MergeGuestCart_DeletesGuestCart(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	svc := newCartService(t, pool)
	repo := NewRepository(pool)
	userID := seedTestUser(t, ctx, pool, "delguest@merge.test")
	productID := seedTestProduct(t, ctx, pool)

	_, gtoken, err := svc.GetOrCreate(ctx, CartIdentity{})
	if err != nil {
		t.Fatalf("guest GetOrCreate: %v", err)
	}
	if _, err := svc.AddItem(ctx, CartIdentity{GuestToken: gtoken}, productID, 1); err != nil {
		t.Fatalf("guest AddItem: %v", err)
	}

	if _, err := svc.MergeGuestCart(ctx, userID, gtoken); err != nil {
		t.Fatalf("MergeGuestCart: %v", err)
	}
	if _, err := repo.GetCartByGuestToken(ctx, gtoken); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected guest cart deleted (ErrNotFound), got %v", err)
	}
}
