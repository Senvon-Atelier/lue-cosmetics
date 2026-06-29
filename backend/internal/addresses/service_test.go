package addresses

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func TestCreate_FirstAddressIsDefault(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user1@example.com")

	addr, err := svc.Create(ctx, userID, AddressInput{
		Label:  "Home",
		Line1:  "123 Main St",
		Line2:  "",
		City:   "Accra",
		Region: "Greater Accra",
		Phone:  "0201234567",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}
	if !addr.IsDefault {
		t.Error("expected first address to be default")
	}
}

func TestCreate_SecondAddressIsNotDefault(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user2@example.com")

	// First address - should be default
	_, err := svc.Create(ctx, userID, AddressInput{
		Label:  "Home",
		Line1:  "1 Home St",
		Line2:  "",
		City:   "Accra",
		Region: "Greater Accra",
		Phone:  "0201111111",
	})
	if err != nil {
		t.Fatalf("create first address: %v", err)
	}

	// Second address - should not be default
	second, err := svc.Create(ctx, userID, AddressInput{
		Label:  "Work",
		Line1:  "2 Work St",
		Line2:  "",
		City:   "Accra",
		Region: "Greater Accra",
		Phone:  "0202222222",
	})
	if err != nil {
		t.Fatalf("create second address: %v", err)
	}
	if second.IsDefault {
		t.Error("expected second address to not be default")
	}

	// Verify first is still default
	listed, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("list addresses: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(listed))
	}
	if !listed[0].IsDefault {
		t.Error("expected first address to still be default")
	}
	if listed[1].IsDefault {
		t.Error("expected second address to not be default")
	}
}

func TestCreate_TrimsAndDefaultsLabel(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user3@example.com")

	addr, err := svc.Create(ctx, userID, AddressInput{
		Label:  "  ",
		Line1:  "123 Main St",
		Line2:  "",
		City:   "Accra",
		Region: "Greater Accra",
		Phone:  "0201234567",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}
	if addr.Label != "Home" {
		t.Errorf("expected label to default to 'Home', got %q", addr.Label)
	}
}

func TestCreate_ValidationFailures(t *testing.T) {
	tests := []struct {
		name    string
		input   AddressInput
		wantErr error
	}{
		{
			name: "empty line1",
			input: AddressInput{
				Label: "Home", Line1: "", Line2: "", City: "Accra",
				Region: "Greater Accra", Phone: "0201234567",
			},
			wantErr: ErrInvalidAddress,
		},
		{
			name: "empty city",
			input: AddressInput{
				Label: "Home", Line1: "123 St", Line2: "", City: "",
				Region: "Greater Accra", Phone: "0201234567",
			},
			wantErr: ErrInvalidAddress,
		},
		{
			name: "overlong label",
			input: AddressInput{
				Label: string(make([]byte, 51)), Line1: "123 St", Line2: "",
				City: "Accra", Region: "Greater Accra", Phone: "0201234567",
			},
			wantErr: ErrInvalidAddress,
		},
		{
			name: "empty phone",
			input: AddressInput{
				Label: "Home", Line1: "123 St", Line2: "", City: "Accra",
				Region: "Greater Accra", Phone: "",
			},
			wantErr: ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
			defer cleanup()

			repo := NewRepository(pool)
			svc := NewService(repo, pool, zap.NewNop())
			userID := seedUser(t, ctx, pool, "user4a@example.com")

			_, err := svc.Create(ctx, userID, tt.input)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreate_CapReached_ReturnsErrAddressLimitReached(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user5@example.com")

	// Create max addresses
	for i := 0; i < MaxAddressesPerUser; i++ {
		_, err := svc.Create(ctx, userID, AddressInput{
			Label:  "Addr",
			Line1:  "123 St",
			Line2:  "",
			City:   "Accra",
			Region: "Greater Accra",
			Phone:  "0201234567",
		})
		if err != nil {
			t.Fatalf("create address %d: %v", i, err)
		}
	}

	// Try to create one more
	_, err := svc.Create(ctx, userID, AddressInput{
		Label:  "Overflow",
		Line1:  "456 St",
		Line2:  "",
		City:   "Accra",
		Region: "Greater Accra",
		Phone:  "0209876543",
	})
	if err != ErrAddressLimitReached {
		t.Errorf("expected ErrAddressLimitReached, got %v", err)
	}
}

func TestUpdate_HappyPath_MergesPatch(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user6@example.com")

	created, err := svc.Create(ctx, userID, AddressInput{
		Label:  "Original",
		Line1:  "123 Main St",
		Line2:  "Apt 1",
		City:   "Accra",
		Region: "Greater Accra",
		Phone:  "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// Patch only label and phone
	newLabel := "Work"
	newPhone := "0509999999"
	updated, err := svc.Update(ctx, userID, created.ID, AddressPatch{
		Label: &newLabel,
		Phone: &newPhone,
	})
	if err != nil {
		t.Fatalf("update address: %v", err)
	}

	if updated.Label != "Work" {
		t.Errorf("expected label 'Work', got %q", updated.Label)
	}
	if updated.Phone != "0509999999" {
		t.Errorf("expected phone '0509999999', got %q", updated.Phone)
	}
	// Other fields unchanged
	if updated.Line1 != "123 Main St" {
		t.Errorf("expected line1 unchanged, got %q", updated.Line1)
	}
	if updated.City != "Accra" {
		t.Errorf("expected city unchanged, got %q", updated.City)
	}
}

func TestUpdate_NotOwned_ReturnsErrNotOwned(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	user1 := seedUser(t, ctx, pool, "user7a@example.com")
	user2 := seedUser(t, ctx, pool, "user7b@example.com")

	addr, err := svc.Create(ctx, user1, AddressInput{
		Label: "User1 Home", Line1: "1 St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// User2 tries to update User1's address
	_, err = svc.Update(ctx, user2, addr.ID, AddressPatch{
		Label: func(s string) *string { return &s }("Hacked"),
	})
	if err != ErrNotOwned {
		t.Errorf("expected ErrNotOwned, got %v", err)
	}
}

func TestUpdate_ValidationOnMergedResult(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user8@example.com")

	created, err := svc.Create(ctx, userID, AddressInput{
		Label: "Home", Line1: "123 St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// Try to set line1 to empty
	emptyLine1 := ""
	_, err = svc.Update(ctx, userID, created.ID, AddressPatch{
		Line1: &emptyLine1,
	})
	if err != ErrInvalidAddress {
		t.Errorf("expected ErrInvalidAddress, got %v", err)
	}
}

func TestDelete_NonDefault_JustRemoves(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user9@example.com")

	// Create default first
	_, err := svc.Create(ctx, userID, AddressInput{
		Label: "Home", Line1: "1 Home St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create default: %v", err)
	}

	// Create non-default
	nonDefault, err := svc.Create(ctx, userID, AddressInput{
		Label: "Work", Line1: "2 Work St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0202222222",
	})
	if err != nil {
		t.Fatalf("create non-default: %v", err)
	}

	// Delete non-default
	if err := svc.Delete(ctx, userID, nonDefault.ID); err != nil {
		t.Fatalf("delete non-default: %v", err)
	}

	// Verify only default remains
	listed, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 address, got %d", len(listed))
	}
	if !listed[0].IsDefault || listed[0].Label != "Home" {
		t.Error("default address should remain")
	}
}

func TestDelete_Default_PromotesNextOldest(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user10@example.com")

	// Create three addresses in order
	home, err := svc.Create(ctx, userID, AddressInput{
		Label: "Home", Line1: "1 Home St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create home: %v", err)
	}

	work, err := svc.Create(ctx, userID, AddressInput{
		Label: "Work", Line1: "2 Work St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0202222222",
	})
	if err != nil {
		t.Fatalf("create work: %v", err)
	}

	_, err = svc.Create(ctx, userID, AddressInput{
		Label: "Gym", Line1: "3 Gym St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0203333333",
	})
	if err != nil {
		t.Fatalf("create gym: %v", err)
	}

	// Set Work as default (Home was default by being first)
	work, err = svc.SetDefault(ctx, userID, work.ID)
	if err != nil {
		t.Fatalf("set work as default: %v", err)
	}

	// Delete the default (Work)
	if err := svc.Delete(ctx, userID, work.ID); err != nil {
		t.Fatalf("delete default: %v", err)
	}

	// Oldest remaining (Home) should now be default
	listed, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(listed))
	}
	if !listed[0].IsDefault {
		t.Error("expected oldest remaining (Home) to be promoted to default")
	}
	if listed[0].ID != home.ID {
		t.Errorf("expected Home to be default, got %s", listed[0].Label)
	}
	// Gym should not be default
	if listed[1].IsDefault {
		t.Error("expected Gym to not be default")
	}
}

func TestDelete_Default_NoOthers_LeavesUserWithZero(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user11@example.com")

	// Create only address (default by being first)
	addr, err := svc.Create(ctx, userID, AddressInput{
		Label: "Home", Line1: "1 Home St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// Delete the only (default) address
	if err := svc.Delete(ctx, userID, addr.ID); err != nil {
		t.Fatalf("delete only address: %v", err)
	}

	// User should have zero addresses
	listed, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected 0 addresses, got %d", len(listed))
	}
}

func TestDelete_NotOwned_ReturnsErrNotOwned(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	user1 := seedUser(t, ctx, pool, "user12a@example.com")
	user2 := seedUser(t, ctx, pool, "user12b@example.com")

	addr, err := svc.Create(ctx, user1, AddressInput{
		Label: "User1 Home", Line1: "1 St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// User2 tries to delete User1's address
	err = svc.Delete(ctx, user2, addr.ID)
	if err != ErrNotOwned {
		t.Errorf("expected ErrNotOwned, got %v", err)
	}
}

func TestSetDefault_HappyPath_FlipsAtomically(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user13@example.com")

	// Create two addresses
	_, err := svc.Create(ctx, userID, AddressInput{
		Label: "Home", Line1: "1 Home St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create home: %v", err)
	}

	work, err := svc.Create(ctx, userID, AddressInput{
		Label: "Work", Line1: "2 Work St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0202222222",
	})
	if err != nil {
		t.Fatalf("create work: %v", err)
	}

	// Home is default (first), set Work as default
	work, err = svc.SetDefault(ctx, userID, work.ID)
	if err != nil {
		t.Fatalf("set work as default: %v", err)
	}

	// Verify flip
	listed, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("list addresses: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(listed))
	}

	// Work should be default and first
	if listed[0].ID != work.ID || !listed[0].IsDefault {
		t.Error("expected Work to be default and first")
	}

	// Home should not be default
	if listed[1].IsDefault {
		t.Error("expected Home to not be default")
	}
}

func TestSetDefault_Idempotent_AlreadyDefault(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user14@example.com")

	// Create address (default by being first)
	addr, err := svc.Create(ctx, userID, AddressInput{
		Label: "Home", Line1: "1 Home St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// Set default on already-default address (should be no-op)
	result, err := svc.SetDefault(ctx, userID, addr.ID)
	if err != nil {
		t.Fatalf("set default on already default: %v", err)
	}

	if !result.IsDefault {
		t.Error("expected address to still be default")
	}
	if result.ID != addr.ID {
		t.Error("expected same address returned")
	}
}

func TestSetDefault_NotOwned_ReturnsErrNotOwned(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	user1 := seedUser(t, ctx, pool, "user15a@example.com")
	user2 := seedUser(t, ctx, pool, "user15b@example.com")

	addr, err := svc.Create(ctx, user1, AddressInput{
		Label: "User1 Home", Line1: "1 St", Line2: "", City: "Accra",
		Region: "Greater Accra", Phone: "0201111111",
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}

	// User2 tries to set User1's address as default
	_, err = svc.SetDefault(ctx, user2, addr.ID)
	if err != ErrNotOwned {
		t.Errorf("expected ErrNotOwned, got %v", err)
	}
}

func TestSetDefault_Nonexistent_ReturnsErrNotFound(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()

	repo := NewRepository(pool)
	svc := NewService(repo, pool, zap.NewNop())
	userID := seedUser(t, ctx, pool, "user16@example.com")

	_, err := svc.SetDefault(ctx, userID, uuid.New())
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
