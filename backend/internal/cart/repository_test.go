package cart

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// seedTestProduct creates a test product with associated brand and category.
// Returns the product ID for use in cart item operations.
func seedTestProduct(t *testing.T, ctx context.Context, pool db.Pool) uuid.UUID {
	t.Helper()
	brandID := uuid.New()
	categoryID := uuid.New()
	productID := uuid.New()

	_, err := pool.Exec(ctx,
		"INSERT INTO brands (id, slug, name) VALUES ($1, $2, $3)",
		brandID, "test-brand-"+brandID.String()[:8], "Test Brand")
	if err != nil {
		t.Fatalf("seed brand: %v", err)
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO categories (id, slug, label) VALUES ($1, $2, $3)",
		categoryID, "test-cat-"+categoryID.String()[:8], "Test Category")
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO products (id, slug, name, brand_id, category_id, price_ghs_minor) VALUES ($1, $2, $3, $4, $5, $6)",
		productID, "test-prod-"+productID.String()[:8], "Test Product", brandID, categoryID, int64(10000))
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	return productID
}

func TestRepository_GetCartByUserID(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	// First create a user
	userID := uuid.New()
	_, err = pool.Exec(ctx, "INSERT INTO users (id, email, name) VALUES ($1, 'test@example.com', 'Test User')", userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	repo := NewRepository(pool)

	// Cart doesn't exist yet
	_, err = repo.GetCartByUserID(ctx, userID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Create cart
	cart, err := repo.CreateCartForUser(ctx, userID)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Cart now exists
	fetched, err := repo.GetCartByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}

	if fetched.ID != cart.ID {
		t.Errorf("expected cart ID %s, got %s", cart.ID, fetched.ID)
	}

	if !fetched.UserID.Valid || fetched.UserID.Bytes != userID {
		t.Errorf("expected user_id %s, got %v", userID, fetched.UserID)
	}

	if fetched.GuestToken != nil {
		t.Errorf("expected guest_token to be nil, got %s", *fetched.GuestToken)
	}
}

func TestRepository_GetCartByGuestToken(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	token := uuid.New().String()

	// Cart doesn't exist yet
	_, err = repo.GetCartByGuestToken(ctx, token)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Create cart
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Cart now exists
	fetched, err := repo.GetCartByGuestToken(ctx, token)
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}

	if fetched.ID != cart.ID {
		t.Errorf("expected cart ID %s, got %s", cart.ID, fetched.ID)
	}

	if fetched.GuestToken == nil || *fetched.GuestToken != token {
		t.Errorf("expected guest_token %s, got %v", token, fetched.GuestToken)
	}

	if fetched.UserID.Valid {
		t.Errorf("expected user_id to be invalid, got %v", fetched.UserID.Bytes)
	}
}

func TestRepository_CreateCartForUser(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	// First create a user
	userID := uuid.New()
	_, err = pool.Exec(ctx, "INSERT INTO users (id, email, name) VALUES ($1, 'test@example.com', 'Test User')", userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	repo := NewRepository(pool)

	cart, err := repo.CreateCartForUser(ctx, userID)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	if cart.ID == uuid.Nil {
		t.Error("expected cart ID to be set")
	}

	if !cart.UserID.Valid || cart.UserID.Bytes != userID {
		t.Errorf("expected user_id %s, got %v", userID, cart.UserID)
	}

	if cart.GuestToken != nil {
		t.Errorf("expected guest_token to be nil, got %s", *cart.GuestToken)
	}
}

func TestRepository_CreateCartForGuest(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	token := uuid.New().String()

	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	if cart.ID == uuid.Nil {
		t.Error("expected cart ID to be set")
	}

	if cart.GuestToken == nil || *cart.GuestToken != token {
		t.Errorf("expected guest_token %s, got %v", token, cart.GuestToken)
	}

	if cart.UserID.Valid {
		t.Errorf("expected user_id to be invalid, got %v", cart.UserID.Bytes)
	}
}

func TestRepository_CHECK_constraint(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	// Test 1: Insert with both user_id and guest_token NULL should fail
	t.Run("both null fails", func(t *testing.T) {
		_, err := pool.Exec(ctx, "INSERT INTO carts (user_id, guest_token) VALUES (NULL, NULL)")
		if err == nil {
			t.Error("expected CHECK constraint error, got nil")
		}
	})

	// Test 2: Insert with both user_id and guest_token set should fail
	t.Run("both set fails", func(t *testing.T) {
		userID := uuid.New()
		token := uuid.New().String()
		_, err := pool.Exec(ctx, "INSERT INTO carts (user_id, guest_token) VALUES ($1, $2)", userID, token)
		if err == nil {
			t.Error("expected CHECK constraint error, got nil")
		}
	})

	// Test 3: Insert with only user_id should succeed (after creating a user)
	t.Run("only user_id succeeds", func(t *testing.T) {
		// First create a user
		userID := uuid.New()
		_, err := pool.Exec(ctx, "INSERT INTO users (id, email, name) VALUES ($1, 'test@example.com', 'Test User')", userID)
		if err != nil {
			t.Fatalf("create user: %v", err)
		}

		// Then insert a cart with that user_id
		_, err = pool.Exec(ctx, "INSERT INTO carts (user_id, guest_token) VALUES ($1, NULL)", userID)
		if err != nil {
			t.Errorf("expected success, got %v", err)
		}
	})

	// Test 4: Insert with only guest_token should succeed
	t.Run("only guest_token succeeds", func(t *testing.T) {
		token := uuid.New().String()
		_, err := pool.Exec(ctx, "INSERT INTO carts (user_id, guest_token) VALUES (NULL, $1)", token)
		if err != nil {
			t.Errorf("expected success, got %v", err)
		}
	})
}

func TestRepository_ListCartItems(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create a guest cart
	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// List items on empty cart
	items, err := repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}

	// Create some products first
	productID1 := seedTestProduct(t, ctx, pool)
	productID2 := seedTestProduct(t, ctx, pool)

	item1, err := repo.UpsertCartItemAddQty(ctx, cart.ID, productID1, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item1: %v", err)
	}

	item2, err := repo.UpsertCartItemAddQty(ctx, cart.ID, productID2, 1, 20000)
	if err != nil {
		t.Fatalf("upsert item2: %v", err)
	}

	// List items
	items, err = repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	if items[0].ID != item1.ID {
		t.Errorf("expected first item ID %s, got %s", item1.ID, items[0].ID)
	}

	if items[1].ID != item2.ID {
		t.Errorf("expected second item ID %s, got %s", item2.ID, items[1].ID)
	}
}

func TestRepository_UpsertCartItemAddQty(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create a guest cart
	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Create a product first
	productID := seedTestProduct(t, ctx, pool)

	// Insert new item
	item, err := repo.UpsertCartItemAddQty(ctx, cart.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	if item.Qty != 2 {
		t.Errorf("expected qty 2, got %d", item.Qty)
	}

	if item.UnitPriceGhsMinor != 10000 {
		t.Errorf("expected unit_price_ghs_minor 10000, got %d", item.UnitPriceGhsMinor)
	}

	// Upsert same item - should add qty
	item, err = repo.UpsertCartItemAddQty(ctx, cart.ID, productID, 3, 15000)
	if err != nil {
		t.Fatalf("upsert item again: %v", err)
	}

	if item.Qty != 5 {
		t.Errorf("expected qty 5 (2+3), got %d", item.Qty)
	}

	// Original price should be preserved
	if item.UnitPriceGhsMinor != 10000 {
		t.Errorf("expected unit_price_ghs_minor 10000 (original), got %d", item.UnitPriceGhsMinor)
	}
}

func TestRepository_SetCartItemQty(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create a guest cart
	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Create a product first
	productID := seedTestProduct(t, ctx, pool)

	// Insert an item
	item, err := repo.UpsertCartItemAddQty(ctx, cart.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	// Set qty
	err = repo.SetCartItemQty(ctx, item.ID, cart.ID, 5)
	if err != nil {
		t.Fatalf("set qty: %v", err)
	}

	// Verify
	fetched, err := repo.GetCartItemByID(ctx, item.ID, cart.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}

	if fetched.Qty != 5 {
		t.Errorf("expected qty 5, got %d", fetched.Qty)
	}
}

func TestRepository_DeleteCartItem(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create a guest cart
	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Create a product first
	productID := seedTestProduct(t, ctx, pool)

	// Insert an item
	item, err := repo.UpsertCartItemAddQty(ctx, cart.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	// Delete item
	err = repo.DeleteCartItem(ctx, item.ID, cart.ID)
	if err != nil {
		t.Fatalf("delete item: %v", err)
	}

	// Verify deleted
	_, err = repo.GetCartItemByID(ctx, item.ID, cart.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// List items should be empty
	items, err := repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}

func TestRepository_GetCartItemByID(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create two guest carts
	token1 := uuid.New().String()
	cart1, err := repo.CreateCartForGuest(ctx, token1)
	if err != nil {
		t.Fatalf("create cart1: %v", err)
	}

	token2 := uuid.New().String()
	cart2, err := repo.CreateCartForGuest(ctx, token2)
	if err != nil {
		t.Fatalf("create cart2: %v", err)
	}

	// Create a product first
	productID := seedTestProduct(t, ctx, pool)

	// Add item to cart1
	item1, err := repo.UpsertCartItemAddQty(ctx, cart1.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item1: %v", err)
	}

	// Get item from cart1 should succeed
	_, err = repo.GetCartItemByID(ctx, item1.ID, cart1.ID)
	if err != nil {
		t.Errorf("expected success, got %v", err)
	}

	// Get item from cart2 should fail (IDOR protection)
	_, err = repo.GetCartItemByID(ctx, item1.ID, cart2.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_TouchCart(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Touch cart
	err = repo.TouchCart(ctx, cart.ID)
	if err != nil {
		t.Errorf("touch cart: %v", err)
	}
}

func TestRepository_DeleteCart(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Delete cart
	err = repo.DeleteCart(ctx, cart.ID)
	if err != nil {
		t.Errorf("delete cart: %v", err)
	}

	// Verify deleted
	_, err = repo.GetCartByGuestToken(ctx, token)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_GetCartItemByProduct(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	token := uuid.New().String()
	cart, err := repo.CreateCartForGuest(ctx, token)
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	// Create a product first
	productID := seedTestProduct(t, ctx, pool)

	// Item doesn't exist yet
	_, err = repo.GetCartItemByProduct(ctx, cart.ID, productID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Add item
	_, err = repo.UpsertCartItemAddQty(ctx, cart.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	// Item now exists
	item, err := repo.GetCartItemByProduct(ctx, cart.ID, productID)
	if err != nil {
		t.Errorf("expected success, got %v", err)
	}

	if item.ProductID != productID {
		t.Errorf("expected product_id %s, got %s", productID, item.ProductID)
	}
}

func TestRepository_SetCartItemQty_WrongCartID_ReturnsErrNotFound(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create two guest carts
	token1 := uuid.New().String()
	cart1, err := repo.CreateCartForGuest(ctx, token1)
	if err != nil {
		t.Fatalf("create cart1: %v", err)
	}

	token2 := uuid.New().String()
	cart2, err := repo.CreateCartForGuest(ctx, token2)
	if err != nil {
		t.Fatalf("create cart2: %v", err)
	}

	// Create a product and add it to cart1
	productID := seedTestProduct(t, ctx, pool)
	item1, err := repo.UpsertCartItemAddQty(ctx, cart1.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item1: %v", err)
	}

	// Try to set qty on item1 using cart2's ID (IDOR scenario)
	err = repo.SetCartItemQty(ctx, item1.ID, cart2.ID, 5)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound when setting qty on item from different cart, got %v", err)
	}

	// Verify the actual item in cart1 still has its original qty (unchanged)
	fetched, err := repo.GetCartItemByID(ctx, item1.ID, cart1.ID)
	if err != nil {
		t.Fatalf("get item1: %v", err)
	}

	if fetched.Qty != 2 {
		t.Errorf("expected original qty 2 after cross-cart attempt, got %d", fetched.Qty)
	}
}

func TestRepository_DeleteCartItem_WrongCartID_ReturnsErrNotFound(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create two guest carts
	token1 := uuid.New().String()
	cart1, err := repo.CreateCartForGuest(ctx, token1)
	if err != nil {
		t.Fatalf("create cart1: %v", err)
	}

	token2 := uuid.New().String()
	cart2, err := repo.CreateCartForGuest(ctx, token2)
	if err != nil {
		t.Fatalf("create cart2: %v", err)
	}

	// Create a product and add it to cart1
	productID := seedTestProduct(t, ctx, pool)
	item1, err := repo.UpsertCartItemAddQty(ctx, cart1.ID, productID, 2, 10000)
	if err != nil {
		t.Fatalf("upsert item1: %v", err)
	}

	// Try to delete item1 using cart2's ID (IDOR scenario)
	err = repo.DeleteCartItem(ctx, item1.ID, cart2.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound when deleting item from different cart, got %v", err)
	}

	// Verify the actual item still exists in cart1
	_, err = repo.GetCartItemByID(ctx, item1.ID, cart1.ID)
	if err != nil {
		t.Errorf("expected item to still exist after cross-cart delete attempt, got %v", err)
	}
}
