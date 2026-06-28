# Cart with Guest Tokens + Merge Implementation Plan (Plan 4 of 15)

> **For executing agents:** This plan is structured as 4 independent bundles. Each bundle is implemented in its own session and committed as ONE commit at the end. Do NOT run the whole plan or final whole-branch review — that gates between bundles externally.

**Goal:** Add the cart surface — anonymous guest carts backed by a `rue_guest_cart` cookie + a `guest_token` column, authenticated carts tied to `user_id`, line items, server-side pricing (catalog + shipping), and `POST /api/v1/cart/merge` that atomically folds a guest cart into the user's cart on first login. After this plan, the cart is the second writable domain in the system (auth was first).

**Architecture:**
- `internal/cart/` package: `repository` + `service` + `handler`. The cart has real business logic (merge, pricing, item upserts), so it gets a service layer per the spec's three-layer convention.
- Guest token: UUIDv4 in the body response AND a non-HttpOnly `rue_guest_cart` cookie (frontend reads both). Stored in `carts.guest_token` (text, UNIQUE WHERE user_id IS NULL).
- Cart-merge is transactional: SELECT guest cart + items FOR UPDATE → upsert each item into user cart by `product_id` (sum qty; keep user's existing `unit_price_ghs_minor` when both rows exist) → DELETE guest cart → clear cookie.
- Pricing: every item's `unit_price_ghs_minor` is snapshotted at add-time from the product. The `GET /cart` response computes subtotal + shipping (via the existing `shipping.Service` from Plan 2) + total.
- Layer-3 RBAC (row-scoping): every cart query uses `WHERE (user_id = $userID OR guest_token = $token)` and never accepts a `cart_id` from the client.

**Tech Stack:** No new external deps. Reuses chi, pgx/v5, sqlc, goose, slog, swaggo, the existing `shipping.Service`, `auth.Handlers` middleware.

## Global Constraints

- **Module path:** `github.com/oti-adjei/ruecosmetics`.
- **Working directory:** `/Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/`. Backend paths relative to `backend/`.
- **Money:** every monetary column `BIGINT` pesewas.
- **Guest cart cookie:** name `rue_guest_cart`, NOT HttpOnly (frontend reads it for the localStorage mirror), SameSite=Lax, Path=`/`, Secure iff `Env != "development"`. 30-day Max-Age.
- **Guest token format:** UUIDv4 string (use `google/uuid.NewString()` so it's URL-safe and 36 chars). Stored in `carts.guest_token`. No signing — possession of the cookie value authorizes cart operations on that cart.
- **Cart row CHECK constraint:** exactly one of `user_id` / `guest_token` is non-null.
- **Cart resolution rule:** authenticated requests (cookie present and valid via `RequireSession`-like resolution) use the user's cart; otherwise the guest cart is used (creating one + setting the cookie on the first `GET /cart` with no `rue_guest_cart`).
- **Cart-merge on login:** the FRONTEND calls `POST /cart/merge` immediately after a successful login when it holds a guest token in localStorage. The backend never auto-merges. (Frontend logic comes in Plan 11; backend exposes the endpoint now and the smoke test exercises it.)
- **`GET /cart` totals:** include `subtotal_ghs_minor`, `shipping_cost_ghs_minor`, `free_shipping_remainder_ghs_minor`, `total_ghs_minor`. Empty cart returns `{items: [], subtotal_ghs_minor: 0, shipping_cost_ghs_minor: 0, free_shipping_remainder_ghs_minor: free_over, total_ghs_minor: 0}`.
- **Validation:** `qty >= 1` on add/update. Unknown `product_id` → 400 `validation_failed`. Item not in the caller's cart → 404 `not_found` (NOT 403 — anti-IDOR enumeration).
- **No `fmt.Sprintf` building SQL anywhere** (the spec's structural-injection-impossible rule).
- **Commit identity:** every commit uses `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit ...`.
- **Bundled commits (per user preference):** the 4 bundles land as **4 commits**.
- **HEAD before Plan 4 begins:** `1bfd801` (Plan 3 final cleanup).

## File Structure

```
casestud/ruecosmetics/backend/
├── internal/
│   ├── cart/                              # NEW (Bundles 1-3)
│   │   ├── repository.go                  # sqlc wrappers
│   │   ├── repository_test.go             # integration
│   │   ├── service.go                     # GetOrCreate, AddItem, UpdateQty, RemoveItem, Merge
│   │   ├── service_test.go
│   │   ├── handler.go                     # GET /cart, POST /cart/items, PATCH/DELETE /cart/items/{id}, POST /cart/merge
│   │   ├── handler_test.go
│   │   └── cookie.go                      # rue_guest_cart helpers (set/clear)
│   ├── catalog/
│   │   └── repository.go                  # MODIFY (Bundle 1): add GetProductPriceBySlug or GetProductByID for pricing lookup
│   ├── app/
│   │   └── app.go                         # MODIFY (Bundle 3): add Cart *cart.Service to Application; wire in New()
│   └── auth/
│       └── handler.go                     # MODIFY (Bundle 3): mount /cart/merge inside MountAuthGated group
├── cmd/api/
│   ├── main.go                            # MODIFY (Bundle 3): mount cart handlers (some public, /merge auth-gated)
│   └── main_test.go                       # MODIFY (Bundle 4): smoke test guest-cart → signup → merge → /cart shows merged items
├── migrations/
│   └── 00004_cart.sql                     # NEW (Bundle 1)
├── queries/
│   └── cart.sql                           # NEW (Bundle 1)
└── docs/                                  # REGENERATED (Bundle 4)
```

---

## Bundle 1 — Schema + sqlc + repository

**Tasks:** migration, queries, catalog-side helper for product price lookup, cart repository.

### Files
- Create: `backend/migrations/00004_cart.sql`
- Create: `backend/queries/cart.sql`
- Regenerate: `backend/internal/db/sqlc/`
- Modify: `backend/internal/catalog/repository.go` — add `GetProductByID(ctx, id uuid.UUID) (sqlcq.Product, error)` for the cart service to look up price + slug at add-time.
- Create: `backend/internal/cart/repository.go`
- Create: `backend/internal/cart/repository_test.go`

### Migration

File: `backend/migrations/00004_cart.sql`
```sql
-- +goose Up
CREATE TABLE carts (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid REFERENCES users(id) ON DELETE CASCADE,
    guest_token  text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    -- exactly one of user_id / guest_token must be set
    CHECK ((user_id IS NULL) <> (guest_token IS NULL))
);
CREATE UNIQUE INDEX idx_carts_user_id ON carts(user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_carts_guest_token ON carts(guest_token) WHERE guest_token IS NOT NULL;

CREATE TABLE cart_items (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id              uuid NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id           uuid NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    qty                  int  NOT NULL CHECK (qty >= 1),
    unit_price_ghs_minor bigint NOT NULL CHECK (unit_price_ghs_minor >= 0),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    UNIQUE (cart_id, product_id)
);
CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);

-- +goose Down
DROP TABLE cart_items;
DROP TABLE carts;
```

### Queries

File: `backend/queries/cart.sql`
```sql
-- name: GetCartByUserID :one
SELECT id, user_id, guest_token, created_at, updated_at FROM carts WHERE user_id = $1;

-- name: GetCartByGuestToken :one
SELECT id, user_id, guest_token, created_at, updated_at FROM carts WHERE guest_token = $1;

-- name: CreateCartForUser :one
INSERT INTO carts (user_id) VALUES ($1) RETURNING id, user_id, guest_token, created_at, updated_at;

-- name: CreateCartForGuest :one
INSERT INTO carts (guest_token) VALUES ($1) RETURNING id, user_id, guest_token, created_at, updated_at;

-- name: TouchCart :exec
UPDATE carts SET updated_at = now() WHERE id = $1;

-- name: DeleteCart :exec
DELETE FROM carts WHERE id = $1;

-- name: ListCartItems :many
SELECT id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at
FROM cart_items
WHERE cart_id = $1
ORDER BY created_at ASC;

-- name: GetCartItemByID :one
SELECT id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at
FROM cart_items
WHERE id = $1 AND cart_id = $2;

-- name: GetCartItemByProduct :one
SELECT id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at
FROM cart_items
WHERE cart_id = $1 AND product_id = $2;

-- name: UpsertCartItemAddQty :one
INSERT INTO cart_items (cart_id, product_id, qty, unit_price_ghs_minor)
VALUES ($1, $2, $3, $4)
ON CONFLICT (cart_id, product_id) DO UPDATE
SET qty = cart_items.qty + EXCLUDED.qty, updated_at = now()
RETURNING id, cart_id, product_id, qty, unit_price_ghs_minor, created_at, updated_at;

-- name: SetCartItemQty :exec
UPDATE cart_items SET qty = $3, updated_at = now()
WHERE id = $1 AND cart_id = $2;

-- name: DeleteCartItem :exec
DELETE FROM cart_items WHERE id = $1 AND cart_id = $2;
```

> Implementer note: regenerate sqlc and adjust the catalog `GetProductByID` query (it may not exist yet — add `-- name: GetProductByID :one SELECT * FROM products WHERE id = $1;` to `queries/catalog.sql` and regenerate).

### Catalog repository addition

In `internal/catalog/repository.go`, add:
```go
func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (sqlcq.Product, error) {
    p, err := r.q.GetProductByID(ctx, id)
    if errors.Is(err, pgx.ErrNoRows) {
        return sqlcq.Product{}, ErrNotFound  // reuse existing sentinel; export if not already
    }
    return p, err
}
```

If `ErrNotFound` is not currently exported from `catalog`, export it: `var ErrNotFound = errors.New("catalog: not found")`. (The auth package has its own `ErrNotFound` — that's fine; they're scoped per package.)

### Cart Repository

File: `backend/internal/cart/repository.go` — wraps sqlc Queries. Exports:
```go
type Repository struct { /* *sqlcq.Queries + db.Pool */ }
func NewRepository(pool db.Pool) *Repository
func (r *Repository) Pool() db.Pool
var ErrNotFound = errors.New("cart: not found")

func (r *Repository) GetCartByUserID(ctx, userID uuid.UUID) (sqlcq.Cart, error)        // ErrNotFound on miss
func (r *Repository) GetCartByGuestToken(ctx, token string) (sqlcq.Cart, error)        // ErrNotFound on miss
func (r *Repository) CreateCartForUser(ctx, userID uuid.UUID) (sqlcq.Cart, error)
func (r *Repository) CreateCartForGuest(ctx, token string) (sqlcq.Cart, error)
func (r *Repository) DeleteCart(ctx, id uuid.UUID) error
func (r *Repository) TouchCart(ctx, id uuid.UUID) error
func (r *Repository) ListCartItems(ctx, cartID uuid.UUID) ([]sqlcq.CartItem, error)
func (r *Repository) GetCartItemByID(ctx, itemID, cartID uuid.UUID) (sqlcq.CartItem, error) // ErrNotFound — note BOTH ids in WHERE for row scoping
func (r *Repository) GetCartItemByProduct(ctx, cartID, productID uuid.UUID) (sqlcq.CartItem, error)
func (r *Repository) UpsertCartItemAddQty(ctx, cartID, productID uuid.UUID, qty int32, unitPrice int64) (sqlcq.CartItem, error)
func (r *Repository) SetCartItemQty(ctx, itemID, cartID uuid.UUID, qty int32) error
func (r *Repository) DeleteCartItem(ctx, itemID, cartID uuid.UUID) error
```

**Always include `cart_id` in WHERE clauses for item operations** — this is the Layer-3 RBAC row-scoping per spec §10.3.

### Tests

`repository_test.go`: integration tests using `testsupport.StartPostgres` + `testsupport.Migrate(t, url, "../../migrations")`. Cover happy paths for each method, the CHECK constraint (exactly-one of user_id/guest_token), the unique-by-cart-product constraint via upsert.

### Commit

Bundle 1 commit message (single commit at end):
```
feat(cart): schema, sqlc queries, repository (foundation)

- migrations/00004_cart.sql: carts + cart_items with CHECK and unique
  partial indexes for user_id and guest_token
- queries/cart.sql: GetByUserID/GuestToken, CreateForUser/Guest, item
  upsert-by-product-with-qty-add, set-qty, delete
- internal/catalog/repository.go: GetProductByID + exported ErrNotFound
- internal/cart/repository.go: typed wrappers, integration tests
```

---

## Bundle 2 — cart.Service + handlers (excluding /merge)

**Tasks:** service with GetOrCreateCart / AddItem / UpdateQty / RemoveItem; handlers for GET /cart, POST /cart/items, PATCH /cart/items/{id}, DELETE /cart/items/{id}; cookie helper.

### Files
- Create: `backend/internal/cart/service.go`
- Create: `backend/internal/cart/service_test.go`
- Create: `backend/internal/cart/handler.go`
- Create: `backend/internal/cart/handler_test.go`
- Create: `backend/internal/cart/cookie.go`

### Service interfaces

```go
type Service struct {
    Repo     *Repository
    Catalog  *catalog.Repository    // for price/slug lookup
    Shipping *shipping.Service
    Log      *slog.Logger
    Now      func() time.Time       // injectable clock
}
func NewService(repo *Repository, catalog *catalog.Repository, ship *shipping.Service, log *slog.Logger) *Service

// CartIdentity is what the handler resolves before calling the service.
// Exactly one field is non-zero.
type CartIdentity struct {
    UserID     uuid.UUID
    GuestToken string
}

type View struct {
    CartID                       uuid.UUID
    GuestToken                   string         // empty if owned by user
    Items                        []ItemView
    SubtotalGhsMinor             int64
    ShippingCostGhsMinor         int64
    FreeShippingRemainderGhsMinor int64
    TotalGhsMinor                int64
}
type ItemView struct {
    ID                  uuid.UUID
    ProductID           uuid.UUID
    ProductSlug         string
    ProductName         string
    ProductImagePath    string
    Qty                 int32
    UnitPriceGhsMinor   int64
    LineTotalGhsMinor   int64
}

// GetOrCreate resolves identity → cart, creating a guest cart (and minting a
// guest token via uuid.NewString()) if identity is empty.
// Returns the View AND the guest token if newly minted (caller sets cookie).
func (s *Service) GetOrCreate(ctx, id CartIdentity) (View, mintedToken string, err error)

// AddItem upserts an item; qty must be >= 1.
func (s *Service) AddItem(ctx, id CartIdentity, productID uuid.UUID, qty int32) (View, error)

// UpdateQty sets a specific item's qty; itemID is row-scoped to the resolved cart.
func (s *Service) UpdateQty(ctx, id CartIdentity, itemID uuid.UUID, qty int32) (View, error)

// RemoveItem deletes; row-scoped.
func (s *Service) RemoveItem(ctx, id CartIdentity, itemID uuid.UUID) (View, error)

var (
    ErrInvalidQty     = errors.New("cart: qty must be >= 1")
    ErrUnknownProduct = errors.New("cart: unknown product")
    ErrItemNotFound   = errors.New("cart: item not found in cart")
)
```

### Pricing rule

`AddItem` calls `s.Catalog.GetProductByID(productID)` to get `price_ghs_minor`. Snapshot that as `unit_price_ghs_minor` on first insert. On upsert (item already exists), the SQL keeps the existing `unit_price_ghs_minor` (the upsert query only adds qty). This means the price-at-add-time semantics is preserved even if the catalog price changes between adds.

`buildView(items, cart)` iterates items, fetches each product for `slug/name/image_path` (cheap; ≤ ~20 items), computes line totals, calls `s.Shipping.Quote(subtotal)`, returns the View.

### Cookie helper

File: `backend/internal/cart/cookie.go`
```go
package cart

import (
    "net/http"
    "time"
)

const GuestCookieName = "rue_guest_cart"
const GuestCookieMaxAge = 30 * 24 * 3600

func SetGuestCookie(w http.ResponseWriter, token, domain string, secure bool) {
    http.SetCookie(w, &http.Cookie{
        Name:     GuestCookieName,
        Value:    token,
        Path:     "/",
        Domain:   domain,
        MaxAge:   GuestCookieMaxAge,
        HttpOnly: false,            // frontend reads this to mirror to localStorage
        Secure:   secure,
        SameSite: http.SameSiteLaxMode,
    })
}

func ClearGuestCookie(w http.ResponseWriter, domain string, secure bool) {
    http.SetCookie(w, &http.Cookie{
        Name: GuestCookieName, Value: "", Path: "/", Domain: domain,
        MaxAge: -1, HttpOnly: false, Secure: secure, SameSite: http.SameSiteLaxMode,
    })
}
```

### Handlers

File: `backend/internal/cart/handler.go`

```go
type Handlers struct {
    Svc          *Service
    AuthSvc      *auth.Service   // for resolving session
    CookieName   string          // session cookie name (e.g., "rue_session")
    CookieDomain string
    Secure       bool
}
func NewHandlers(svc *Service, authSvc *auth.Service, sessionCookieName, cookieDomain string, secure bool) *Handlers

func (h *Handlers) Mount(r chi.Router)              // public: GET /cart, POST /cart/items, PATCH /cart/items/{id}, DELETE /cart/items/{id}
func (h *Handlers) MountAuthGated(r chi.Router)     // auth-required: POST /cart/merge (implemented in Bundle 3)
```

The handlers MUST resolve identity in a uniform way:
```go
func (h *Handlers) resolveIdentity(r *http.Request) cart.CartIdentity {
    // Try session cookie first.
    if c, err := r.Cookie(h.CookieName); err == nil {
        if view, err := h.AuthSvc.GetSession(r.Context(), c.Value); err == nil {
            return cart.CartIdentity{UserID: view.UserID}
        }
    }
    // Fall back to guest cookie.
    if c, err := r.Cookie(cart.GuestCookieName); err == nil && c.Value != "" {
        return cart.CartIdentity{GuestToken: c.Value}
    }
    return cart.CartIdentity{} // empty — service will mint
}
```

`GET /cart`:
- Resolve identity. Call `Svc.GetOrCreate`. If a token was minted, set the cookie. Return the View.

`POST /cart/items`:
- Body `{product_id: uuid, qty: int}`. Validate qty >= 1, parse UUID. Resolve identity (mint if needed; set cookie). Call `Svc.AddItem`. Return View. 400 on unknown product / bad qty.

`PATCH /cart/items/{id}`:
- Body `{qty: int}`. Resolve identity. Call `Svc.UpdateQty`. 404 if item not in cart (covers IDOR — never 403 to avoid enumeration).

`DELETE /cart/items/{id}`:
- Resolve identity. Call `Svc.RemoveItem`. 204 on success; 404 if item not in cart.

### Tests

`service_test.go`:
- GetOrCreate with empty identity mints a guest cart + token (assert `mintedToken != ""`).
- GetOrCreate with guest token re-uses cart.
- GetOrCreate with user_id mints a user cart on first call, reuses on second.
- AddItem stores `unit_price_ghs_minor` from the catalog at add-time.
- AddItem twice for same product → upsert sums qty, keeps original price.
- UpdateQty changes qty.
- RemoveItem deletes the row.
- UpdateQty / RemoveItem on a cart_item_id from a DIFFERENT cart → `ErrItemNotFound` (IDOR).
- View totals include shipping computed from the existing shipping.Service.

`handler_test.go`:
- GET /cart with no cookies → 200 with empty cart + `Set-Cookie: rue_guest_cart=...`.
- POST /cart/items adds an item.
- GET /cart with the guest cookie returns the cart (no new cookie set).
- Customer A cannot PATCH customer B's cart_item_id → 404.

### Commit

```
feat(cart): service, GET/POST/PATCH/DELETE handlers with guest cookies

- internal/cart/service.go: GetOrCreate, AddItem, UpdateQty, RemoveItem
  with at-add-time price snapshot and shipping-aware View totals
- internal/cart/cookie.go: rue_guest_cart helpers (set/clear)
- internal/cart/handler.go: resolveIdentity (session → guest → mint),
  4 public endpoints with row-scoped 404s
- service_test.go + handler_test.go: integration coverage including IDOR
```

---

## Bundle 3 — Cart merge + mount + IDOR matrix

**Tasks:** `MergeGuestCart` service method + handler; mount cart in main.go; comprehensive IDOR test for cross-cart access.

### Files
- Modify: `backend/internal/cart/service.go` — add `MergeGuestCart` method
- Modify: `backend/internal/cart/service_test.go` — add merge tests
- Modify: `backend/internal/cart/handler.go` — add `/cart/merge` handler + extend `MountAuthGated`
- Modify: `backend/internal/cart/handler_test.go` — add merge handler tests
- Modify: `backend/internal/app/app.go` — wire `Cart *cart.Service` into `Application`
- Modify: `backend/cmd/api/main.go` — mount cart handlers (public + auth-gated)
- Create: `backend/internal/cart/idor_test.go` — comprehensive cross-tenant matrix

### Merge service method

```go
type MergeRequest struct {
    GuestToken string
}

// MergeGuestCart upserts each item from the guest cart into the user cart by
// product_id (summing qty; KEEPING the user's existing unit_price_ghs_minor
// when both rows exist). Deletes the guest cart on success. Returns the
// resulting user cart View. If the guest token doesn't resolve to a cart
// (already merged, expired, never existed), the method is idempotent: just
// returns the user's current cart, no error.
func (s *Service) MergeGuestCart(ctx, userID uuid.UUID, guestToken string) (View, error)
```

Implementation outline (transactional):
1. Open tx.
2. Load (or create) the user's cart.
3. Look up the guest cart by token. Missing → commit + return user cart view (idempotent no-op).
4. Read guest items. For each, run an upsert against the user cart:
   - If `(user_cart_id, product_id)` exists: UPDATE user_cart_items SET qty = user.qty + guest.qty (keep user's unit price).
   - Else: INSERT with the guest item's qty + unit_price.
5. DELETE guest cart (cascades to items via FK).
6. Commit.
7. Return view.

### Merge handler

```go
// merge godoc
//
// @Summary  Merge a guest cart into the user's cart (auth required)
// @Tags     cart
// @Accept   json
// @Produce  json
// @Param    body body mergeRequestBody true "Guest token from localStorage"
// @Success  200 {object} cartResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  401 {object} httpx.ErrorEnvelope
// @Router   /cart/merge [post]
func (h *Handlers) merge(w http.ResponseWriter, r *http.Request) {
    userID, ok := auth.GetUserID(r.Context())
    if !ok { httpx.WriteError(w, 401, httpx.CodeUnauthorized, "authentication required", nil); return }
    var body struct { GuestToken string `json:"guest_token"` }
    if err := httpx.ReadJSON(r, &body); err != nil {
        httpx.WriteError(w, 400, httpx.CodeBadRequest, "invalid body", nil); return
    }
    view, err := h.Svc.MergeGuestCart(r.Context(), userID, body.GuestToken)
    if err != nil {
        h.logger().ErrorContext(r.Context(), "merge", "err", err)
        httpx.WriteError(w, 500, httpx.CodeInternal, "merge failed", nil); return
    }
    // Clear the guest cookie regardless of whether anything was merged.
    cart.ClearGuestCookie(w, h.CookieDomain, h.Secure)
    httpx.WriteJSON(w, 200, viewToResponse(view))
}
```

`MountAuthGated(r)` registers `r.Post("/cart/merge", h.merge)`.

### Application wiring

In `internal/app/app.go`, add field `Cart *cart.Service`. In `New`:
```go
cartRepo := cart.NewRepository(pool)
catalogRepo := catalog.NewRepository(pool)   // reuse if already constructed
cartSvc := cart.NewService(cartRepo, catalogRepo, ship, logger)
```

Then `return &Application{... Cart: cartSvc, ...}`.

### main.go mount

Inside the existing `r.Route("/api/v1", func(api chi.Router) { ... })`:
```go
cartHandlers := cart.NewHandlers(a.Cart, a.Auth, cfg.SessionCookieName, cfg.SessionCookieDomain, secure)
cartHandlers.Mount(api)                 // public: GET/POST/PATCH/DELETE
// inside the existing RequireSession group:
api.Group(func(r chi.Router) {
    r.Use(authHandlers.RequireSession)
    me.NewHandlers().Mount(r)
    authHandlers.MountAuthGated(r)
    cartHandlers.MountAuthGated(r)      // POST /cart/merge
})
```

### IDOR test matrix

`backend/internal/cart/idor_test.go` exercises:
- User A's cart_item_id is NOT addressable from User B's session (PATCH/DELETE → 404).
- Guest token A's cart_item_id is NOT addressable with guest token B (PATCH/DELETE → 404).
- User A's session + Guest token B's `rue_guest_cart` cookie → session wins (the user cart is touched, the guest cookie is ignored unless `/cart/merge` is called).
- Anonymous PATCH on an item_id from an existing user cart → 404 (no fishing).

### Commit

```
feat(cart): MergeGuestCart, /cart/merge endpoint, mount, IDOR matrix

- service: MergeGuestCart — transactional upsert-by-product summing qty,
  user's existing unit_price wins on conflict; idempotent for unknown tokens
- handler: POST /cart/merge (auth required); clears rue_guest_cart cookie
- app + main: wire Cart service into Application; mount public + auth-gated
- idor_test.go: cross-cart access returns 404 across user/user, guest/guest,
  anonymous, and session-precedence scenarios
```

---

## Bundle 4 — OpenAPI regen + end-to-end smoke

**Tasks:** regenerate OpenAPI; extend `cmd/api/main_test.go` smoke test to exercise the full guest-cart → signup → merge flow.

### Files
- Regenerate: `backend/docs/{swagger.json,swagger.yaml,docs.go}`
- Modify: `backend/cmd/api/main_test.go` — extend smoke test

### Smoke test additions

After the existing smoke test sequence (signup → cookie → /me), add:

1. `GET /api/v1/cart` (no cookies) → 200 + empty cart + capture `rue_guest_cart` cookie.
2. `POST /api/v1/cart/items` with that guest cookie + a real product_id from the catalog (the smoke test will need to insert a product — either via direct SQL or by calling `testsupport.Migrate` against the catalog seed). For brevity, insert a single test product directly via the pool.
3. `GET /api/v1/cart` (guest cookie) → 200 + one item, subtotal correct.
4. `POST /api/v1/auth/signup` → cookie captured.
5. `POST /api/v1/cart/merge` with the new session cookie + body `{"guest_token": "<token-from-step-1>"}` → 200 + view shows the item in the user cart.
6. `GET /api/v1/cart` with session cookie → 200 + same item.

### OpenAPI regen + drift-check

```bash
cd backend
PATH="$(go env GOPATH)/bin:$PATH" swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency
```

Verify all 5 cart routes appear exactly once:
```bash
for path in /cart /cart/items '/cart/items/{id}' /cart/merge; do
    echo -n "$path: "; grep -c "\"$path\"" docs/swagger.json
done
```

(`/cart/items` appears under `paths` with both `post` and `delete`+`patch` for the `{id}` variant. Check the actual entries land.)

```bash
cd .. && make drift-check && make test
```

### Commit

```
feat(cart): regenerate openapi, smoke test guest→signup→merge

- docs: 5 cart routes in swagger.json
- cmd/api/main_test.go: full guest cart → add item → signup → merge →
  /cart shows merged item
```

---

## Verification — end of Plan 4

- [ ] `make test` exits 0.
- [ ] `make drift-check` exits 0.
- [ ] `make up && make dev` boots the server. The following works end-to-end:
  - `curl -c jar -X GET :8080/api/v1/cart` → 200 + sets `rue_guest_cart`.
  - `curl -b jar -X POST :8080/api/v1/cart/items -H 'Content-Type: application/json' -d '{"product_id":"<id>","qty":2}'` → 200 + item appears.
  - signup → merge → `/cart` shows the merged item.
- [ ] 4 new commits on top of `1bfd801`, all prefixed `feat(cart):` (the bundle messages above).
- [ ] No `fmt.Sprintf` constructing SQL.
- [ ] No raw guest token logged outside test files.

## Self-Review Notes

- **Spec coverage:** Section 4.1 (carts + cart_items), Section 4.4 (cart-merge rule), Section 4.5 (guest cart lifecycle), Section 5.2 cart endpoints, Section 10.3 Layer-3 row scoping (cart_id always in WHERE for item ops).
- **Pricing snapshot:** `unit_price_ghs_minor` is captured at add-time and survives upserts. Plan 5 (checkout) will reuse this when snapshotting line items into orders.
- **Guest cookie is NOT HttpOnly** — by design, the frontend reads it to mirror to localStorage so the post-login merge call can include the token. Possession of the cookie is the only authorization for that guest cart; no signing.
- **Merge idempotency:** unknown/missing guest token returns the user's current cart with no error — this lets the frontend call `/cart/merge` unconditionally after login without having to track whether it was already merged.
- **Why session precedence in `resolveIdentity`:** an authenticated user who happens to have an orphan guest cookie sitting in their browser should NOT have their user cart shadowed by a stale guest cart. Merge is opt-in via the explicit POST.
- **Forward-looking for Plan 5 (checkout):** checkout will read the cart, re-price from DB (never trust the snapshot for the order total — re-resolve to catch admin price changes between cart and checkout), then write order_items with the freshly-resolved prices. The cart's `unit_price_ghs_minor` snapshot is for *display continuity* in the cart UI, not the final order total.
