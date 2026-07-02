# Testing Readiness Features — Design Spec

**Date:** 2026-07-01
**Status:** Draft for review
**Goal:** Enable local testing of Rue Cosmetics by implementing seed data system, admin order status updates, and user profile updates

## Overview

This design specifies three interconnected features needed to make the Rue Cosmetics application testable locally:

1. **Seed Data System** — A `cmd/seed` binary that populates the database with representative demo data
2. **Admin Order Status Update** — Transactional status transitions with audit logging via new `order_history` table
3. **User Profile Updates** — Full profile editing (name + image) with validation and proper error handling

---

## 1. Seed Data System

### 1.1 Purpose

Enable local development and testing by populating an empty database with realistic demo data. The seed system creates products, users, orders, addresses, and wishlists that mirror real application state.

### 1.2 Architecture

**Binary structure:**
- Entry point: `backend/cmd/seed/main.go`
- Uses existing `app.Application` DI container for DB access
- Leverages existing sqlc-generated queries

**Data files:**
- Location: `backend/seed/data/*.json`
- Files: `products.json`, `users.json`, `orders.json`, `addresses.json`, `wishlists.json`

**Runtime config:**
- Reuses existing `shipping.Service` from `internal/shipping`
- Loads from `backend/seed/config/shipping_config.json` only if override needed for seed data
- Shipping config already loaded at app startup via `Config` struct

### 1.3 Seed Data Shape (Middle Ground)

**Products (~15 items):**
- 3-4 products per category (skincare, haircare, bodycare)
- Realistic GHS pricing (85-450 range)
- All four tones represented (lavender, cream, rose, ink)
- Placeholder image paths pointing to `frontend/public/products/`
- Mix of ratings (4.5-4.9) and review counts (45-200)
- Some with sale prices (was_price field)

**Users (4 demo accounts):**

| User Type | Email | Password | Orders | Addresses | Wishlist |
|-----------|-------|----------|--------|-----------|----------|
| Admin | admin@ruecosmetics.com | admin123 | 0 | 0 | 0 |
| Customer A | customer.a@demo.com | demo123 | 3 (paid, shipped, delivered) | 2 | 3 items |
| Customer B | customer.b@demo.com | demo123 | 1 (delivered) | 1 | 0 |
| Customer C | customer.c@demo.com | demo123 | 0 | 0 | 2 items |

**Orders (5-6 total):**
- Each order has 2-4 line items
- Statuses: paid, shipped, delivered
- Realistic subtotals (GHS 200-800 range)
- Shipping addresses properly formatted as JSONB

**Addresses (4-5 total):**
- Distributed across Customer A and B
- Mix of Accra, Kumasi, Tamale locations
- One marked as `is_default` per user

**Wishlists (6-8 entries):**
- Customer A: 3 products
- Customer C: 2 products
- Mix of categories

### 1.4 Conflict Resolution Strategy

**Catalog tables (categories, brands, products):**
```sql
INSERT INTO products (...) VALUES (...)
ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name, price=EXCLUDED.price, ...
```
- Re-running seed updates existing records with JSON data

**Demo users:**
```sql
INSERT INTO users (...) VALUES (...)
ON CONFLICT (email) DO UPDATE SET name=EXCLUDED.name, updated_at=NOW();
```
- Refreshes demo user data on re-run
- **Protection heuristic:** If a row with demo email exists but `provider != 'seed-marker'`, skip and log warning

**Demo orders/addresses/wishlists:**
```sql
DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE user_id = $demo_user_id);
DELETE FROM orders WHERE user_id = $demo_user_id;
-- Then insert fresh data
```
- Scoped delete then insert for demo user IDs only
- Real users' data never touched

**Marker system:**
- `oauth_accounts` rows for seed users have `provider = 'seed-marker'`
- This identifies which users are "owned" by the seed system

**Password hashing:**
- Uses same argon2id parameters as production auth system
- Parameters: Memory=64MiB, Time=3, Parallelism=4, SaltLength=16, KeyLength=32
- Ensures demo passwords work consistently across environments

### 1.5 Implementation Flow

```
1. Load JSON files from backend/seed/data/
2. Connect to DB via app.Application
3. Begin transaction for each entity type:
   a. Catalog (categories → brands → products)
   b. Users (passwords hashed via argon2id)
   c. Orders (with order_items)
   d. Addresses
   e. Wishlists
4. Print summary: X products, Y users, Z orders created/updated
5. Exit cleanly or roll back on error
```

### 1.6 Error Handling

- Invalid JSON: Fatal error with file/line context
- DB constraint violations: Log and continue with remaining entities
- Missing foreign keys: Fatal error (data dependencies must be satisfied)
- Missing product images: Log warning and continue (seed should not fail if `frontend/public/products/` files are missing)

---

## 2. Admin Order Status Update

### 2.1 Purpose

Enable admins to update order statuses through the API with full audit trail. Currently validates transitions but doesn't persist the change (TODO placeholder).

### 2.2 Architecture

**New migration:** `00007_order_history.sql`
- Creates `order_history` audit table
- Adds indexes for query performance

**Query additions:** `backend/queries/admin.sql`
- `UpdateOrderStatus` — updates orders table
- `InsertOrderHistory` — creates audit record

**Service layer:** `admin.Service.UpdateOrderStatus`
- Replaces TODO placeholder with transactional implementation
- Uses `db.WithTx` helper for atomicity

### 2.3 Schema: order_history Table

```sql
CREATE TABLE order_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  old_status TEXT NOT NULL,
  new_status TEXT NOT NULL,
  changed_by_user_id UUID REFERENCES users(id),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  note TEXT
);

CREATE INDEX idx_order_history_order_id ON order_history(order_id);
CREATE INDEX idx_order_history_changed_at ON order_history(changed_at DESC);
```

**Fields:**
- `id` — Unique identifier for the history entry
- `order_id` — Foreign key to orders table, cascades on order deletion
- `old_status` — Previous status before transition
- `new_status` — New status after transition
- `changed_by_user_id` — Admin user who made the change (nullable for system changes)
- `changed_at` — Timestamp of the status change
- `note` — Optional field for future use (reason, customer notes, etc.)

### 2.4 Transaction Flow

```
BEGIN
  1. SELECT * FROM orders WHERE id = $order_id FOR UPDATE
     → Lock the row for this transaction
  
  2. UPDATE orders
     SET status = $new_status, updated_at = NOW()
     WHERE id = $order_id
     → Apply the new status
  
  3. INSERT INTO order_history (order_id, old_status, new_status, changed_by_user_id)
     VALUES ($order_id, $old_status, $new_status, $admin_user_id)
     → Create audit trail
  
  4. COMMIT
     → Atomic completion
```

**Rollback scenarios:**
- Constraint violations → automatic rollback
- Application errors → explicit rollback via `db.WithTx`

### 2.5 Query Definitions

```sql
-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
          total_ghs_minor, paystack_reference, paystack_transaction_id,
          shipping_address, created_at, updated_at;

-- name: InsertOrderHistory :one
INSERT INTO order_history (order_id, old_status, new_status, changed_by_user_id, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, old_status, new_status, changed_by_user_id, changed_at, note;
```

### 2.6 Service Implementation

```go
func (s *Service) UpdateOrderStatus(ctx context.Context, params UpdateOrderStatusParams) error {
    var finalErr error
    
    err := db.WithTx(ctx, s.Repo.Pool(), func(tx pgx.Tx) error {
        // 1. Lock and fetch current order
        order, err := s.Repo.GetOrderByIDTx(ctx, tx, params.OrderID)
        if err != nil {
            return err
        }
        
        // 2. Validate transition (existing logic)
        if !isValidTransition(order.Status, params.Status) {
            return fmt.Errorf("invalid transition")
        }
        
        // 3. Update order status
        updatedOrder, err := s.Repo.UpdateOrderStatusTx(ctx, tx, params.OrderID, params.Status)
        if err != nil {
            return err
        }
        
        // 4. Insert audit record
        adminUserID := getAdminUserIDFromContext(ctx) // from session
        _, err = s.Repo.InsertOrderHistoryTx(ctx, tx, InsertOrderHistoryParams{
            OrderID:         params.OrderID,
            OldStatus:       order.Status,
            NewStatus:       params.Status,
            ChangedByUserID: adminUserID,
        })
        if err != nil {
            return err
        }
        
        s.Log.Info("order status updated",
            zap.String("order_id", params.OrderID.String()),
            zap.String("old_status", order.Status),
            zap.String("new_status", params.Status))
        
        return nil
    })
    
    return err
}
```

### 2.7 Validation Logic (State Machine Pattern)

**Centralized state machine:**
```go
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusPaid      OrderStatus = "paid"
    StatusFulfilled OrderStatus = "fulfilled"
    StatusShipped   OrderStatus = "shipped"
    StatusDelivered OrderStatus = "delivered"
    StatusCancelled OrderStatus = "cancelled"
    StatusRefunded  OrderStatus = "refunded"
)

type OrderStateMachine struct {
    transitions map[OrderStatus][]OrderStatus
}

func (sm *OrderStateMachine) CanTransition(from, to OrderStatus) bool {
    allowed, ok := sm.transitions[from]
    if !ok {
        return false
    }
    for _, status := range allowed {
        if status == to {
            return true
        }
    }
    return false
}

var defaultTransitions = map[OrderStatus][]OrderStatus{
    StatusPending:   {StatusPaid, StatusCancelled},
    StatusPaid:      {StatusFulfilled, StatusCancelled, StatusRefunded},
    StatusFulfilled: {StatusShipped, StatusCancelled},
    StatusShipped:   {StatusDelivered, StatusCancelled},
    StatusDelivered: {StatusRefunded},
    StatusCancelled: {}, // terminal
    StatusRefunded:  {}, // terminal
}
```

**Idempotency:**
- Same-status transitions (`paid → paid`) are allowed and return success without changes
- This supports safe retry of status update operations

**Authorization:**
- RBAC check enforced at handler level via `auth.RequireRole("admin")` middleware
- Service layer validates `auth.MustBeAdmin(ctx)` as belt-and-suspenders
- Order ownership not required (admin can update any order)

---

## 3. User Profile Updates

### 3.1 Purpose

Enable users to update their profile (name and image URL) through the API. Currently returns 501 "not yet implemented" (TODO placeholder).

### 3.2 Architecture

**Service layer addition:** Create `internal/me/service.go`
- `ProfileService` encapsulates profile update business logic
- Handler becomes thin: auth check → service call → response
- Validation and normalization moved to service layer

**Handler changes:** `me.Handlers`
- Inject `me.ProfileService` into handlers struct
- Handler delegates all logic to service
- Returns HTTP-appropriate errors from service errors

**Repository usage:** Leverages existing `auth.UpdateUser` query
```sql
-- name: UpdateUser :one
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, email, name, image, email_verified, created_at, updated_at;
```

**Request/Response:**
- Accepts: `{name: string, image: string}`
- Returns: Updated user profile with all fields

### 3.3 Validation Rules

**Name field:**
- Required if provided
- Min length: 2 characters
- Max length: 100 characters
- Trimmed of leading/trailing whitespace
- No control characters allowed
- Error message: "name must be between 2-100 characters"

**Image field:**
- Optional
- If provided, must be valid URL format
- Must use http:// or https:// scheme
- Error message: "image must be a valid URL"

**Email field:**
- Returns 501 "email updates require verification flow (not yet implemented)"
- Unchanged from existing behavior

### 3.4 Service Layer Implementation

**New file:** `internal/me/service.go`
```go
package me

type ProfileService struct {
    AuthRepo *auth.Repository
    Log      *zap.Logger
}

type UpdateProfileParams struct {
    UserID    uuid.UUID
    Name      *string
    Image     *string
}

type UpdateProfileResult struct {
    UserID        uuid.UUID
    Email         string
    Name          string
    Image         *string
    Role          string
    EmailVerified bool
}

func (s *ProfileService) UpdateProfile(ctx context.Context, params UpdateProfileParams) (UpdateProfileResult, error) {
    // Validate and normalize name if provided
    if params.Name != nil {
        if err := validateAndNormalizeName(params.Name); err != nil {
            return UpdateProfileResult{}, err
        }
    }

    // Validate image URL if provided
    if params.Image != nil {
        if err := validateImageURL(*params.Image); err != nil {
            return UpdateProfileResult{}, err
        }
    }

    // Build update params
    var namePtr, imagePtr pgtype.Text
    if params.Name != nil {
        namePtr = pgtype.Text{String: *params.Name, Valid: true}
    }
    if params.Image != nil {
        imagePtr = pgtype.Text{String: *params.Image, Valid: true}
    }

    // Update via auth repository
    user, err := s.AuthRepo.UpdateUser(ctx, sqlc.UpdateUserParams{
        ID:    params.UserID,
        Name:  namePtr,
        Image: imagePtr,
    })
    if err != nil {
        s.Log.Error("failed to update user profile", zap.Error(err), zap.String("user_id", params.UserID.String()))
        return UpdateProfileResult{}, fmt.Errorf("update profile: %w", err)
    }

    return UpdateProfileResult{
        UserID:        user.ID,
        Email:         user.Email,
        Name:          user.Name,
        Image:         user.Image,
        Role:          getRoleFromContext(ctx), // from session
        EmailVerified: user.EmailVerified,
    }, nil
}
```

**Handler (thin delegation):**
```go
type Handlers struct {
    ProfileService *ProfileService
    Log            *zap.Logger
}

func (h *Handlers) updateProfile(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    view, ok := auth.GetSessionView(ctx)
    if !ok {
        httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
        return
    }

    var req updateProfileRequest
    if err := httpx.ReadJSON(r, &req); err != nil {
        httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid request body", nil)
        return
    }

    // Email updates still require verification flow
    if req.Email != nil && *req.Email != view.Email {
        httpx.WriteError(w, http.StatusNotImplemented, httpx.CodeInternal, "email updates require verification flow (not yet implemented)", nil)
        return
    }

    result, err := h.ProfileService.UpdateProfile(ctx, me.UpdateProfileParams{
        UserID: view.UserID,
        Name:   req.Name,
        Image:  req.Image,
    })
    if err != nil {
        // Map service errors to HTTP responses
        if strings.Contains(err.Error(), "name must be") || strings.Contains(err.Error(), "image URL") {
            httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, err.Error(), nil)
        } else {
            httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to update profile", nil)
        }
        return
    }

    httpx.WriteJSON(w, http.StatusOK, meResponse{
        UserID:        result.UserID.String(),
        Email:         result.Email,
        Name:          result.Name,
        Image:         result.Image,
        Role:          result.Role,
        EmailVerified: result.EmailVerified,
    })
}
```

### 3.5 Validation Helpers

```go
func validateAndNormalizeName(name *string) error {
    if name == nil {
        return nil // name is optional
    }
    trimmed := strings.TrimSpace(*name)
    
    // Empty string after trimming: reject (cannot remove name entirely)
    if trimmed == "" {
        return fmt.Errorf("name cannot be empty")
    }
    
    if len(trimmed) < 2 || len(trimmed) > 100 {
        return fmt.Errorf("name must be between 2-100 characters")
    }
    if strings.ContainsAny(trimmed, "\n\r\t\x00") {
        return fmt.Errorf("name contains invalid characters")
    }
    
    // Normalize the input in-place
    *name = trimmed
    return nil
}

func validateImageURL(urlStr string) error {
    if urlStr == "" {
        return fmt.Errorf("image URL cannot be empty")
    }
    u, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL format")
    }
    if u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("image URL must use http or https scheme")
    }
    return nil
}
```

**Empty string behavior:**
- `name: ""` → Rejected with error "name cannot be empty"
- `name: "   "` → Trimmed to empty, rejected
- `image: ""` → Rejected with error "image URL cannot be empty"
- Field omitted (`null`) → No update for that field (existing value preserved)

### 3.6 Session Handling

**For v1:** We update the database but don't force-refresh the session cookie immediately.

**Rationale:**
- Profile updates (name, image) are less critical than auth changes
- The next authenticated request will fetch fresh user data from DB via `auth.GetSessionView`
- Session expiry is 30 days with rolling refresh
- Avoids complexity of session re-issuing for non-critical updates

**Future consideration:** If profile updates become more frequent or critical, implement session refresh via `auth.Service.UpdateSessionData`.

---

## 4. Integration Points

### 4.1 Seed → Order Status Updates

- Seed creates orders in various statuses (paid, shipped, delivered)
- Admin can then transition these seeded orders through status workflow
- Audit trail captures all transitions including those on seeded data

### 4.2 Seed → User Profile Updates

- Seed creates 4 demo users with initial names
- Users can update their profiles via `/me` endpoint
- Admin user can update profiles via admin dashboard (future feature)

### 4.3 Combined Testing Flow

```
1. Run seed binary → DB populated with demo data
2. Log in as Customer A → See 3 orders in various states
3. Log in as admin → View all orders, update statuses
4. Verify order_history captures all status changes
5. Customer A updates profile name/image → See changes reflected
```

---

## 5. OpenAPI Updates

After implementation, run `make openapi` to regenerate:

- `PATCH /api/v1/me` — Updated to return 200 with profile (currently returns 501)
- `PATCH /api/v1/admin/orders/{id}/status` — Document audit trail behavior
- Add `order_history` to schemas section

---

## 6. Testing Considerations

### 6.1 Seed System Tests

- Unit: Test JSON loading and parsing
- Integration: Test seed binary against testcontainers Postgres
- Verify: Catalog conflict resolution works on re-run
- Verify: Demo user protection heuristic works

### 6.2 Order Status Update Tests

- Unit: Test validation logic (valid transitions matrix)
- Integration: Test transaction commits/rolls back correctly
- Test: Audit record created for each status change
- Test: Concurrent updates don't race (FOR UPDATE locking)

### 6.3 Profile Update Tests

- Unit: Test validation helpers (name length, image URL format)
- Integration: Test full request/response cycle
- Test: Email updates still return 501
- Test: Empty request (no changes) returns current profile

---

## 7. Migration Rollback

The `00007_order_history.sql` migration adds the `order_history` table and indexes. Rollback strategy:

```sql
-- Down migration
DROP INDEX IF EXISTS idx_order_history_changed_at;
DROP INDEX IF EXISTS idx_order_history_order_id;
DROP TABLE IF EXISTS order_history;
```

**Rollback considerations:**
- `ON DELETE CASCADE` ensures history rows are deleted when orders are deleted
- No foreign key constraints prevent rollback
- Data loss: All audit history would be lost on rollback (acceptable for local testing)

**Testing rollback:**
- Test both up and down migrations in CI
- Verify schema is consistent after rollback
- No orphaned objects or constraint violations

---

## 8. Error Taxonomy

Standard error codes used across all three features:

| Code | HTTP Status | Description | Used By |
|------|-------------|-------------|----------|
| `validation` | 400 | Request validation failed | Profile updates, order status |
| `unauthorized` | 401 | Authentication required | Profile updates |
| `forbidden` | 403 | Insufficient permissions | Order status updates (admin-only) |
| `not_found` | 404 | Resource not found | Order status (invalid order ID) |
| `conflict` | 409 | State conflict | Order status (invalid transition) |
| `internal` | 500 | Server error | All features (DB failures, etc.) |
| `not_implemented` | 501 | Feature not yet available | Email updates in profile |

**Error response shape (consistent):**
```json
{
  "error": {
    "code": "validation",
    "message": "name must be between 2-100 characters",
    "fields": {
      "name": "must be between 2-100 characters"
    }
  }
}
```

---

## 9. Authorization Model

All three features use the existing RBAC system with defense-in-depth:

**Profile updates (`PATCH /me`):**
- Layer 1: `auth.RequireSession` middleware ensures valid session
- Layer 2: Handler extracts `userID` from session context
- Layer 3: Users can only update their own profile (implicit via session)
- No admin override for profile updates (v1 scope)

**Order status updates (`PATCH /admin/orders/{id}/status`):**
- Layer 1: Router-level `auth.RequireRole("admin")` middleware
- Layer 2: Service-level `auth.MustBeAdmin(ctx)` check
- Layer 3: No ownership check required (admin can update any order)

**Seed system:**
- No authorization required (admin/dev tool)
- Runs with direct DB access, bypasses application auth layer
- Intentionally privileged for local development

---

## 10. Future Work

**Seed system:**
- Add `--reset` flag to drop and recreate all data
- Add `--dry-run` flag to preview changes without applying
- Support environment-specific seeds (dev vs staging)

**Order status:**
- Add webhook notifications on status changes
- Add customer email notifications on shipped/delivered
- Add "notes" field to capture reason for cancellation

**Profile updates:**
- Add email change with verification flow
- Add password change with current password confirmation
- Add account deletion with grace period
