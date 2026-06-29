package addresses

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

func TestCreateAddress_HappyPath(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user1@example.com")

	params := sqlcq.CreateAddressParams{
		UserID:    userID,
		Label:     "Home",
		Line1:     "123 Main Street",
		Line2:     "Apt 4B",
		City:      "Accra",
		Region:    "Greater Accra",
		Phone:     "0201234567",
		IsDefault: false,
	}

	addr, err := repo.CreateAddress(ctx, params)
	if err != nil {
		t.Fatalf("create address: %v", err)
	}
	if addr.ID == uuid.Nil {
		t.Error("expected address ID to be set")
	}
	if addr.Label != "Home" {
		t.Errorf("expected label 'Home', got %q", addr.Label)
	}
	if addr.Line1 != "123 Main Street" {
		t.Errorf("expected line1 '123 Main Street', got %q", addr.Line1)
	}
	if addr.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, addr.UserID)
	}
	if addr.IsDefault {
		t.Error("expected is_default false, got true")
	}
}

func TestGetAddressByID_NotFound_ReturnsErrNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)

	_, err := repo.GetAddressByID(ctx, uuid.New())
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListAddressesByUserID_OrdersDefaultFirstThenNewest(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user2@example.com")

	// Create three addresses: second one is default
	addr1, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Work", Line1: "1 Work St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0201111111", IsDefault: false,
	})
	if err != nil {
		t.Fatalf("create addr1: %v", err)
	}

	addr2, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Home", Line1: "2 Home St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0202222222", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("create addr2: %v", err)
	}

	addr3, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Gym", Line1: "3 Gym St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0203333333", IsDefault: false,
	})
	if err != nil {
		t.Fatalf("create addr3: %v", err)
	 }

	listed, err := repo.ListAddressesByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("list addresses: %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("expected 3 addresses, got %d", len(listed))
	}
	// Default (addr2) should be first
	if listed[0].ID != addr2.ID || !listed[0].IsDefault {
		t.Error("expected default address to be first")
	}
	// Non-defaults should be ordered by created_at DESC (newest first)
	if listed[1].ID != addr3.ID {
		t.Errorf("expected second address to be addr3 (newest non-default), got %s", listed[1].ID)
	}
	if listed[2].ID != addr1.ID {
		t.Errorf("expected third address to be addr1 (oldest non-default), got %s", listed[2].ID)
	}
}

func TestListAddressesByUserID_EmptyForUserWithNone(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user3@example.com")

	listed, err := repo.ListAddressesByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("list addresses: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("expected 0 addresses, got %d", len(listed))
	}
}

func TestUpdateAddress_HappyPath(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user4@example.com")

	created, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Old Label", Line1: "123 Old St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0201234567", IsDefault: false,
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	updated, err := repo.UpdateAddress(ctx, sqlcq.UpdateAddressParams{
		ID:     created.ID,
		Label:  "New Label",
		Line1:  "456 New St",
		Line2:  "Apt 1",
		City:   "Kumasi",
		Region: "Ashanti",
		Phone:  "0509876543",
	})
	if err != nil {
		t.Fatalf("update address: %v", err)
	}
	if updated.Label != "New Label" {
		t.Errorf("expected label 'New Label', got %q", updated.Label)
	}
	if updated.Line1 != "456 New St" {
		t.Errorf("expected line1 '456 New St', got %q", updated.Line1)
	}
	if updated.City != "Kumasi" {
		t.Errorf("expected city 'Kumasi', got %q", updated.City)
	}
}

func TestUpdateAddress_NotFound_ReturnsErrNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)

	_, err := repo.UpdateAddress(ctx, sqlcq.UpdateAddressParams{
		ID: uuid.New(), Label: "X", Line1: "Y", City: "Z", Region: "R", Phone: "P",
	})
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteAddress_HappyPath_AndCascadesNothing(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user5@example.com")

	addr, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Home", Line1: "123 Main St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0201234567", IsDefault: false,
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	if err := repo.DeleteAddress(ctx, addr.ID); err != nil {
		t.Fatalf("delete address: %v", err)
	}

	// Verify it's gone
	_, err = repo.GetAddressByID(ctx, addr.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestCountAddressesByUserID(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user6@example.com")

	// Initially zero
	count, err := repo.CountAddressesByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("count addresses: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	// Add two
	for i := 0; i < 2; i++ {
		if _, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
			UserID: userID, Label: "Addr", Line1: "123 St", Line2: "",
			City: "Accra", Region: "Greater Accra", Phone: "0201234567", IsDefault: false,
		}); err != nil {
			t.Fatalf("create address: %v", err)
		}
	}

	count, err = repo.CountAddressesByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("count addresses: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestPartialUniqueIndex_RejectsTwoDefaultsForSameUser(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user7@example.com")

	// First default succeeds
	_, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Home1", Line1: "1 St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0201111111", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("first default address: %v", err)
	}

	// Second default fails
	_, err = repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Home2", Line1: "2 St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0202222222", IsDefault: true,
	})
	if err == nil {
		t.Fatal("expected unique violation, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") &&
		!strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Errorf("expected unique/duplicate error, got %v", err)
	}
}

func TestPartialUniqueIndex_AllowsTwoDefaultsForDifferentUsers(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	user1 := seedUser(t, ctx, pool, "user8a@example.com")
	user2 := seedUser(t, ctx, pool, "user8b@example.com")

	// Both users can have a default
	_, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: user1, Label: "Home1", Line1: "1 St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0201111111", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("user1 default: %v", err)
	}

	_, err = repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: user2, Label: "Home2", Line1: "2 St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0202222222", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("user2 default: %v", err)
	}

	// Verify both exist
	list1, err := repo.ListAddressesByUserID(ctx, user1)
	if err != nil {
		t.Fatalf("list user1: %v", err)
	}
	if len(list1) != 1 || !list1[0].IsDefault {
		t.Error("user1 should have one default address")
	}

	list2, err := repo.ListAddressesByUserID(ctx, user2)
	if err != nil {
		t.Fatalf("list user2: %v", err)
	}
	if len(list2) != 1 || !list2[0].IsDefault {
		t.Error("user2 should have one default address")
	}
}

func TestCascade_OnUserDelete_AddressesGone(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	userID := seedUser(t, ctx, pool, "user9@example.com")

	// Create two addresses for the user
	addr1, err := repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Home", Line1: "1 Home St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0201111111", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("create addr1: %v", err)
	}

	_, err = repo.CreateAddress(ctx, sqlcq.CreateAddressParams{
		UserID: userID, Label: "Work", Line1: "2 Work St", Line2: "",
		City: "Accra", Region: "Greater Accra", Phone: "0202222222", IsDefault: false,
	})
	if err != nil {
		t.Fatalf("create addr2: %v", err)
	}

	// Delete the user
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	// Verify addresses are gone (cascade)
	_, err = repo.GetAddressByID(ctx, addr1.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after user delete (cascade), got %v", err)
	}

	listed, err := repo.ListAddressesByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("list after user delete: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("expected 0 addresses after cascade, got %d", len(listed))
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
