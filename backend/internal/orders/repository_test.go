package orders

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func seedUser(t *testing.T, ctx context.Context, pool db.Pool, email string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(ctx,
		"INSERT INTO users (id, email, name) VALUES ($1, $2, 'Test User')",
		id, email)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func seedProduct(t *testing.T, ctx context.Context, pool db.Pool) uuid.UUID {
	t.Helper()
	brandID := uuid.New()
	categoryID := uuid.New()
	productID := uuid.New()

	if _, err := pool.Exec(ctx,
		"INSERT INTO brands (id, slug, name) VALUES ($1, $2, $3)",
		brandID, "test-brand-"+brandID.String()[:8], "Test Brand"); err != nil {
		t.Fatalf("seed brand: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO categories (id, slug, label) VALUES ($1, $2, $3)",
		categoryID, "test-cat-"+categoryID.String()[:8], "Test Category"); err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO products (id, slug, name, brand_id, category_id, price_ghs_minor) VALUES ($1, $2, $3, $4, $5, $6)",
		productID, "test-prod-"+productID.String()[:8], "Test Product", brandID, categoryID, int64(10000)); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	return productID
}

func mustReference(t *testing.T) string {
	t.Helper()
	ref, err := GenerateReference()
	if err != nil {
		t.Fatalf("generate reference: %v", err)
	}
	return ref
}

func createOrderParams(userID uuid.UUID, ref string) sqlcq.CreateOrderParams {
	return sqlcq.CreateOrderParams{
		UserID:            userID,
		SubtotalGhsMinor:  10000,
		ShippingGhsMinor:  2500,
		TotalGhsMinor:     12500,
		PaystackReference: ref,
		ShippingAddress:   []byte(`{"line1":"123 Main","city":"Accra","country":"GH"}`),
	}
}

func TestRepository_CreateOrder_Happy(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer1@example.com")
	ref := mustReference(t)

	order, err := repo.CreateOrder(ctx, createOrderParams(userID, ref))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.ID == uuid.Nil {
		t.Error("expected order ID to be set")
	}
	if order.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", order.Status)
	}
	if order.PaystackReference != ref {
		t.Errorf("expected reference %q, got %q", ref, order.PaystackReference)
	}
	if order.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, order.UserID)
	}
	if order.TotalGhsMinor != 12500 {
		t.Errorf("expected total 12500, got %d", order.TotalGhsMinor)
	}
}

func TestRepository_GetOrderByReference_HappyAndNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer2@example.com")
	ref := mustReference(t)

	created, err := repo.CreateOrder(ctx, createOrderParams(userID, ref))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	got, err := repo.GetOrderByReference(ctx, ref)
	if err != nil {
		t.Fatalf("get by reference: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, got.ID)
	}

	_, err = repo.GetOrderByReference(ctx, "RUE-DEADBEEF")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_GetOrderByID_HappyAndNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer3@example.com")
	ref := mustReference(t)

	created, err := repo.CreateOrder(ctx, createOrderParams(userID, ref))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	got, err := repo.GetOrderByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.PaystackReference != ref {
		t.Errorf("expected reference %q, got %q", ref, got.PaystackReference)
	}

	_, err = repo.GetOrderByID(ctx, uuid.New())
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_CreateOrderItem_AndList(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer4@example.com")
	ref := mustReference(t)
	order, err := repo.CreateOrder(ctx, createOrderParams(userID, ref))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	productID1 := seedProduct(t, ctx, pool)
	productID2 := seedProduct(t, ctx, pool)

	items := []sqlcq.CreateOrderItemParams{
		{
			OrderID:              order.ID,
			ProductID:            productID1,
			Qty:                  2,
			UnitPriceGhsMinor:    10000,
			ProductNameSnapshot:  "Test Product A",
			ProductBrandSnapshot: "Brand A",
			ProductImageSnapshot: "/img/a.png",
		},
		{
			OrderID:              order.ID,
			ProductID:            productID2,
			Qty:                  1,
			UnitPriceGhsMinor:    20000,
			ProductNameSnapshot:  "Test Product B",
			ProductBrandSnapshot: "Brand B",
			ProductImageSnapshot: "/img/b.png",
		},
	}
	for _, p := range items {
		if _, err := repo.CreateOrderItem(ctx, p); err != nil {
			t.Fatalf("create order item: %v", err)
		}
	}

	listed, err := repo.ListOrderItems(ctx, order.ID)
	if err != nil {
		t.Fatalf("list order items: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 items, got %d", len(listed))
	}
	if listed[0].ProductNameSnapshot != "Test Product A" {
		t.Errorf("expected first item name 'Test Product A', got %q", listed[0].ProductNameSnapshot)
	}
}

func TestRepository_PaystackReference_UniqueViolation(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer5@example.com")
	ref := mustReference(t)

	if _, err := repo.CreateOrder(ctx, createOrderParams(userID, ref)); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.CreateOrder(ctx, createOrderParams(userID, ref))
	if err == nil {
		t.Fatalf("expected unique violation, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") &&
		!strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Errorf("expected unique/duplicate error, got %v", err)
	}
}

func TestRepository_OrderItems_FK_Cascade_OnOrderDelete(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer6@example.com")
	ref := mustReference(t)
	order, err := repo.CreateOrder(ctx, createOrderParams(userID, ref))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	productID := seedProduct(t, ctx, pool)
	if _, err := repo.CreateOrderItem(ctx, sqlcq.CreateOrderItemParams{
		OrderID:              order.ID,
		ProductID:            productID,
		Qty:                  1,
		UnitPriceGhsMinor:    10000,
		ProductNameSnapshot:  "Test Product",
		ProductBrandSnapshot: "Brand",
		ProductImageSnapshot: "/img/x.png",
	}); err != nil {
		t.Fatalf("create order item: %v", err)
	}

	if _, err := pool.Exec(ctx, "DELETE FROM orders WHERE id = $1", order.ID); err != nil {
		t.Fatalf("delete order: %v", err)
	}

	items, err := repo.ListOrderItems(ctx, order.ID)
	if err != nil {
		t.Fatalf("list order items: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items after cascade delete, got %d", len(items))
	}
}

func TestRepository_CountOrdersByStatus(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "buyer7@example.com")

	for i := 0; i < 3; i++ {
		if _, err := repo.CreateOrder(ctx, createOrderParams(userID, mustReference(t))); err != nil {
			t.Fatalf("create order: %v", err)
		}
	}

	n, err := repo.CountOrdersByStatus(ctx, "pending")
	if err != nil {
		t.Fatalf("count pending: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 pending, got %d", n)
	}

	paid, err := repo.CountOrdersByStatus(ctx, "paid")
	if err != nil {
		t.Fatalf("count paid: %v", err)
	}
	if paid != 0 {
		t.Errorf("expected 0 paid, got %d", paid)
	}
}

func TestRepository_Pool_Exposed(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	if repo.Pool() == nil {
		t.Error("expected Pool() to return non-nil")
	}
	// Sanity: pool actually works.
	if _, err := repo.Pool().Exec(ctx, "SELECT 1"); err != nil {
		t.Errorf("pool exec: %v", err)
	}
}
