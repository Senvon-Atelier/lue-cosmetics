# Plan 6 — Saved-addresses CRUD

**Goal:** Backfill the saved-addresses feature that Plan 5 deferred. Land a pure CRUD resource at `/me/addresses` with default-address semantics (at most one per user, auto-promote on delete-of-default), a per-user soft cap, and an extended smoke test. Frontend integration is out of scope (Plan 12 / future frontend handoff).

**Architecture:**
- New `internal/addresses/` package matching the established `internal/cart/` and `internal/orders/` shape: schema → sqlc queries → typed repository → service → handlers.
- Pure CRUD with three special operations: `SetDefault` (transactional flip), default-on-first-create (auto-set), and default-promote-on-delete (auto-elevate the next-oldest sibling).
- `/checkout/init` is **NOT modified.** It continues to accept inline `shipping_address` only. The frontend resolves a saved address to inline shape on its side before posting. This keeps Plan 5's checkout contract stable.
- At-most-one-default-per-user is enforced by a **partial unique index** on `(user_id) WHERE is_default = true` — the DB is the backstop; the service still maintains it explicitly in transactions for predictable error handling.
- Soft cap: 20 addresses per user. Beyond the cap → `400 validation` with code `address_limit_reached`. Cap is enforced in the service via a `COUNT` before `INSERT`, not by DB constraint (so it's easy to bump later via config if needed).

**Tech Stack:**
- New deps: none. Stdlib + existing pgx/v5 + sqlc + chi + zap.
- Reuses: `auth.RequireSession` middleware, `httpx.WriteError/WriteJSON`, `db.WithTx`, `testsupport.StartPool`, the existing user-scoped routing pattern from `/me/*`.

## Global Constraints

- **Module path:** `github.com/oti-adjei/ruecosmetics`.
- **Working directory:** `/Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/`. Backend paths relative to `backend/`.
- **Endpoint group:** All addresses endpoints live under `/api/v1/me/addresses*` and require an authenticated session (mount inside the existing `RequireSession` group).
- **IDOR rule:** Every endpoint that resolves an address by `{id}` MUST verify `address.user_id == session.user_id`. On mismatch, return `404 not_found` (not `403 forbidden`) — 403 leaks existence. The repository returns the row regardless; ownership check is the **service's** responsibility.
- **Status codes:**
  - `POST /me/addresses` → `201 Created` with the address row.
  - `GET /me/addresses` → `200` with `{addresses: [...]}` (array; empty array if none).
  - `PATCH /me/addresses/{id}` → `200` with the updated row; `404` if not theirs; `400 validation` on bad input.
  - `DELETE /me/addresses/{id}` → `204 No Content`; `404` if not theirs.
  - `POST /me/addresses/{id}/default` → `200` with the now-default row; `404` if not theirs; idempotent if already default.
  - Address-cap-reached on create → `400 validation` with code `address_limit_reached` (use existing `httpx.CodeValidation`; the code suffix lives in the error envelope's optional `details.code` field — match the established pattern from cart/checkout errors).
- **Default-address semantics:**
  - First address for a user is auto-set to `is_default=true` on create (service decides; INSERT carries the flag).
  - Subsequent creates default to `is_default=false`.
  - `SetDefault` is a tx: `UPDATE … SET is_default=false WHERE user_id=$1 AND is_default=true` followed by `UPDATE … SET is_default=true WHERE id=$2 AND user_id=$1`. Idempotent: setting an already-default address re-runs the same statements with no observable change.
  - Deleting the default address auto-promotes the **oldest remaining address** (by `created_at ASC`) to default. If no other addresses exist, the deletion just removes the row and the user is back to zero addresses.
  - The partial unique index `idx_addresses_one_default_per_user ON addresses(user_id) WHERE is_default = true` is the DB-level invariant. If the service ever violates it (concurrent SetDefaults from the same user), the unique violation surfaces as `500 internal` — that's a code bug, not a user-visible state.
- **Validation rules (service-level, not DB):**
  - `line1`: required; trimmed; 1–200 chars after trim.
  - `line2`: optional; 0–200 chars after trim.
  - `city`: required; trimmed; 1–100 chars.
  - `region`: required; trimmed; 1–100 chars.
  - `phone`: required; trimmed; 1–30 chars. No format validation (Ghana phone numbers vary).
  - `label`: optional; 1–50 chars after trim; defaults to `"Home"` on create if absent/empty.
- **PATCH semantics:** PATCH only updates fields present in the request body. Missing fields are unchanged. To allow this cleanly, the request DTO uses `*string` for every field and the service merges into the existing row before validation. Per-field validation rules above still apply to whatever fields are present.
- **Soft cap:** `MaxAddressesPerUser = 20` constant in the `addresses` package. Enforced in `Create` via `CountAddressesByUserID`. Out-of-scope to make this config-driven in Plan 6.
- **Money:** N/A for this plan.
- **No `fmt.Sprintf` building SQL.**
- **Commit identity:** local git config already noreply; plain `git commit -m '…'` is correct.
- **Bundled commits:** 3 bundles → **3 commits**.
- **HEAD before Plan 6 begins:** `c8fe5a0` (Plan 5 / F2 close).
- **No new env vars.**

## File Structure

```
backend/
├── migrations/
│   └── 00006_addresses.sql                          # NEW
├── queries/
│   └── addresses.sql                                # NEW
├── internal/
│   ├── addresses/                                   # NEW package
│   │   ├── repository.go
│   │   ├── repository_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── app/
│   │   └── app.go                                   # MODIFY: wire AddressService
│   ├── db/sqlc/
│   │   └── addresses.sql.go                         # REGENERATED
│   └── ...
├── cmd/api/
│   ├── main.go                                      # MODIFY: mount handlers in /me group
│   └── main_test.go                                 # MODIFY: extend smoke
└── docs/                                            # REGENERATED swagger.json/yaml/go
```

---

## Bundle 1 — Schema + sqlc + Repository

**Tasks:** migration, queries, typed repository, integration tests.

### Files
- Create: `backend/migrations/00006_addresses.sql`
- Create: `backend/queries/addresses.sql`
- Regenerate: `backend/internal/db/sqlc/`
- Create: `backend/internal/addresses/repository.go`
- Create: `backend/internal/addresses/repository_test.go`

### Migration

File: `backend/migrations/00006_addresses.sql`
```sql
-- +goose Up
CREATE TABLE addresses (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label       text NOT NULL DEFAULT 'Home',
    line1       text NOT NULL CHECK (length(line1) > 0),
    line2       text NOT NULL DEFAULT '',
    city        text NOT NULL CHECK (length(city) > 0),
    region      text NOT NULL CHECK (length(region) > 0),
    phone       text NOT NULL CHECK (length(phone) > 0),
    is_default  boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_addresses_user_id ON addresses(user_id);
CREATE UNIQUE INDEX idx_addresses_one_default_per_user
    ON addresses(user_id) WHERE is_default = true;

-- +goose Down
DROP TABLE addresses;
```

> Implementer note: the CHECK constraints on `line1/city/region/phone` are belt-and-suspenders — the service validates these too, but a buggy direct DB write must not produce nonsense rows. The partial unique index is the cornerstone of the default-address invariant.

### Queries

File: `backend/queries/addresses.sql`
```sql
-- name: CreateAddress :one
INSERT INTO addresses (user_id, label, line1, line2, city, region, phone, is_default)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at;

-- name: GetAddressByID :one
SELECT id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at
FROM addresses
WHERE id = $1;

-- name: ListAddressesByUserID :many
SELECT id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at
FROM addresses
WHERE user_id = $1
ORDER BY is_default DESC, created_at DESC;

-- name: UpdateAddress :one
UPDATE addresses
SET label = $2, line1 = $3, line2 = $4, city = $5, region = $6, phone = $7, updated_at = now()
WHERE id = $1
RETURNING id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at;

-- name: DeleteAddress :exec
DELETE FROM addresses WHERE id = $1;

-- name: CountAddressesByUserID :one
SELECT count(*) FROM addresses WHERE user_id = $1;

-- name: ClearDefaultForUser :exec
UPDATE addresses SET is_default = false, updated_at = now()
WHERE user_id = $1 AND is_default = true;

-- name: SetDefaultAddress :one
UPDATE addresses SET is_default = true, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, label, line1, line2, city, region, phone, is_default, created_at, updated_at;

-- name: GetOldestOtherAddress :one
SELECT id
FROM addresses
WHERE user_id = $1 AND id != $2
ORDER BY created_at ASC
LIMIT 1;
```

> Note: `ClearDefaultForUser`, `SetDefaultAddress`, and `GetOldestOtherAddress` are tx-only — used inside `db.WithTx` closures from the service. They are NOT exposed on the public Repository surface (mirroring the Plan 5 `MarkOrderPaid` pattern).

### Repository

File: `backend/internal/addresses/repository.go` — typed wrappers exposing:

```go
type Repository struct { /* *sqlcq.Queries + db.Pool */ }
func NewRepository(pool db.Pool) *Repository
func (r *Repository) Pool() db.Pool

var ErrNotFound = errors.New("addresses: not found")

func (r *Repository) CreateAddress(ctx, params sqlcq.CreateAddressParams) (sqlcq.Address, error)
func (r *Repository) GetAddressByID(ctx, id uuid.UUID) (sqlcq.Address, error)  // ErrNotFound on miss
func (r *Repository) ListAddressesByUserID(ctx, userID uuid.UUID) ([]sqlcq.Address, error)
func (r *Repository) UpdateAddress(ctx, params sqlcq.UpdateAddressParams) (sqlcq.Address, error)  // ErrNotFound on miss
func (r *Repository) DeleteAddress(ctx, id uuid.UUID) error
func (r *Repository) CountAddressesByUserID(ctx, userID uuid.UUID) (int64, error)
```

The tx-only queries (`ClearDefaultForUser`, `SetDefaultAddress`, `GetOldestOtherAddress`) are NOT exposed. They're called inside `db.WithTx` closures via `sqlcq.New(tx)` from the service.

### Tests

`repository_test.go`: integration tests using `testsupport.StartPool(t, "../../migrations")`. Coverage:
- `TestCreateAddress_HappyPath` — insert, assert returned row matches input.
- `TestGetAddressByID_NotFound_ReturnsErrNotFound`.
- `TestListAddressesByUserID_OrdersDefaultFirstThenNewest` — seed 3 addresses, one default; assert default first, then created_at DESC.
- `TestListAddressesByUserID_EmptyForUserWithNone`.
- `TestUpdateAddress_HappyPath`.
- `TestUpdateAddress_NotFound_ReturnsErrNotFound`.
- `TestDeleteAddress_HappyPath_AndCascadesNothing`.
- `TestCountAddressesByUserID`.
- `TestPartialUniqueIndex_RejectsTwoDefaultsForSameUser` — insert two addresses both with `is_default=true` for the same user; assert the second INSERT errors with unique violation.
- `TestPartialUniqueIndex_AllowsTwoDefaultsForDifferentUsers` — sanity check that the partial index doesn't over-constrain.
- `TestCascade_OnUserDelete_AddressesGone` — delete the user row; assert addresses are gone (ON DELETE CASCADE on FK).

### Commit message

```
feat(addresses): schema, sqlc queries, repository (foundation)

- migrations/00006_addresses.sql: addresses table with FK CASCADE on user_id,
  CHECK constraints on required fields, partial unique index enforcing at
  most one default per user
- queries/addresses.sql: Create/Get/List/Update/Delete/Count plus tx-only
  ClearDefaultForUser, SetDefaultAddress, GetOldestOtherAddress
- internal/addresses/repository.go: typed wrappers + ErrNotFound; tx-only
  queries kept off the public surface
- repository_test.go: happy paths, ErrNotFound, partial-unique invariants,
  ordering, FK cascade
```

---

## Bundle 2 — Service (CRUD + SetDefault + auto-promote)

**Tasks:** business logic, transactional default-flip, auto-default-on-first-create, auto-promote-on-delete-default, soft cap.

### Files
- Create: `backend/internal/addresses/service.go`
- Create: `backend/internal/addresses/service_test.go`

### Service

File: `backend/internal/addresses/service.go`

```go
package addresses

import (
    "context"
    "errors"
    "strings"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

const MaxAddressesPerUser = 20

var (
    ErrInvalidAddress      = errors.New("addresses: invalid address")
    ErrAddressLimitReached = errors.New("addresses: limit reached")
    ErrNotOwned            = errors.New("addresses: not owned by caller")  // surfaced as 404 at the handler
)

type Service struct {
    Repo *Repository
    Pool db.Pool
    Log  *zap.Logger
}

func NewService(repo *Repository, pool db.Pool, log *zap.Logger) *Service

// AddressInput is the validated, normalized shape used by Create.
// Trim and normalize happens in Validate(); Service.Create takes the
// already-normalized form.
type AddressInput struct {
    Label  string  // empty → defaults to "Home"
    Line1  string
    Line2  string
    City   string
    Region string
    Phone  string
}

// AddressPatch carries optional fields for Update. Nil = leave unchanged.
type AddressPatch struct {
    Label  *string
    Line1  *string
    Line2  *string
    City   *string
    Region *string
    Phone  *string
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, in AddressInput) (sqlcq.Address, error)
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]sqlcq.Address, error)
func (s *Service) Update(ctx context.Context, userID, addrID uuid.UUID, patch AddressPatch) (sqlcq.Address, error)
func (s *Service) Delete(ctx context.Context, userID, addrID uuid.UUID) error
func (s *Service) SetDefault(ctx context.Context, userID, addrID uuid.UUID) (sqlcq.Address, error)
```

### Method behaviors

**`Create`:**
1. Apply defaults (`label` → `"Home"` if empty after trim) and validate. On miss → `ErrInvalidAddress`.
2. `CountAddressesByUserID`. If `>= MaxAddressesPerUser` → `ErrAddressLimitReached`.
3. If count == 0, `is_default = true`. Else `is_default = false`. (Just compute the bool from the count we already fetched; no second query.)
4. `Repo.CreateAddress(ctx, params)` → return the row.

**`List`:** trivial wrapper around `Repo.ListAddressesByUserID`.

**`Update`:**
1. `Repo.GetAddressByID(ctx, addrID)` → `ErrNotFound` if missing, `ErrNotOwned` if `addr.UserID != userID`. (Handlers map both to 404.)
2. Merge `patch` into the loaded row's fields.
3. Validate the merged result. On miss → `ErrInvalidAddress`.
4. `Repo.UpdateAddress(ctx, params)` → return updated row.

**`Delete`:**
1. `Repo.GetAddressByID(ctx, addrID)` → `ErrNotFound` / `ErrNotOwned` as above.
2. If `!addr.IsDefault`: single-statement `Repo.DeleteAddress`. Done.
3. If `addr.IsDefault`: tx scope:
   - `q := sqlcq.New(tx)`
   - `successorID, err := q.GetOldestOtherAddress(ctx, GetOldestOtherAddressParams{UserID: userID, ID: addrID})`. If `pgx.ErrNoRows` → no successor; just delete and commit. If other error → return.
   - If a successor exists: `q.DeleteAddress(ctx, addrID)` then `q.SetDefaultAddress(ctx, SetDefaultAddressParams{ID: successorID, UserID: userID})`. The DELETE removes the default row first, then the partial unique index allows the new default to be set. Order matters — if you SET first then DELETE, the partial unique index briefly sees two defaults inside the same tx (some PG versions/configs reject that; do DELETE first to be safe).
   - Commit.

**`SetDefault`:**
1. `Repo.GetAddressByID(ctx, addrID)` → `ErrNotFound` / `ErrNotOwned`.
2. If already default (`addr.IsDefault`): return the row as-is. Idempotent no-op.
3. Tx scope:
   - `q := sqlcq.New(tx)`
   - `q.ClearDefaultForUser(ctx, userID)` — flips current default (if any) to false.
   - `updated, err := q.SetDefaultAddress(ctx, SetDefaultAddressParams{ID: addrID, UserID: userID})` — if `pgx.ErrNoRows`, the address was deleted between step 1 and now (rare race); return `ErrNotFound`.
   - Commit.
4. Return `updated`.

### Validation helper

Internal function (not exported):
```go
func validateInput(in *AddressInput) error {
    in.Label = strings.TrimSpace(in.Label)
    if in.Label == "" { in.Label = "Home" }
    if len(in.Label) > 50 { return ErrInvalidAddress }
    in.Line1 = strings.TrimSpace(in.Line1)
    if l := len(in.Line1); l < 1 || l > 200 { return ErrInvalidAddress }
    in.Line2 = strings.TrimSpace(in.Line2)
    if len(in.Line2) > 200 { return ErrInvalidAddress }
    in.City = strings.TrimSpace(in.City)
    if l := len(in.City); l < 1 || l > 100 { return ErrInvalidAddress }
    in.Region = strings.TrimSpace(in.Region)
    if l := len(in.Region); l < 1 || l > 100 { return ErrInvalidAddress }
    in.Phone = strings.TrimSpace(in.Phone)
    if l := len(in.Phone); l < 1 || l > 30 { return ErrInvalidAddress }
    return nil
}
```

`Update` calls a sibling helper `applyPatchAndValidate(existing sqlcq.Address, patch AddressPatch) (AddressInput, error)` that merges the patch into the existing row and runs the same validation.

### Tests

`service_test.go` — integration (real DB via `testsupport.StartPool`):

- `TestCreate_FirstAddressIsDefault` — fresh user, create one address; assert `is_default=true`.
- `TestCreate_SecondAddressIsNotDefault` — create two; assert first remains default, second is not.
- `TestCreate_TrimsAndDefaultsLabel` — pass `label=""`, assert returned label is `"Home"`.
- `TestCreate_ValidationFailures` — table-driven: empty line1, empty city, overlong fields, empty phone. Each → `ErrInvalidAddress`.
- `TestCreate_CapReached_ReturnsErrAddressLimitReached` — seed 20 addresses, attempt 21st, assert error.
- `TestUpdate_HappyPath_MergesPatch` — create with all fields, PATCH only `label` and `phone`; assert other fields unchanged.
- `TestUpdate_NotOwned_ReturnsErrNotOwned` — user A's address, user B tries to PATCH; expect `ErrNotOwned`.
- `TestUpdate_ValidationOnMergedResult` — PATCH sets `line1` to empty string; expect `ErrInvalidAddress`.
- `TestDelete_NonDefault_JustRemoves`.
- `TestDelete_Default_PromotesNextOldest` — three addresses (A default, B, C in order); delete A; assert B is now default (B is older than C).
- `TestDelete_Default_NoOthers_LeavesUserWithZero` — delete the only address; assert user has no addresses, no error.
- `TestDelete_NotOwned_ReturnsErrNotOwned`.
- `TestSetDefault_HappyPath_FlipsAtomically` — A default, set B; assert A is no longer default, B is.
- `TestSetDefault_Idempotent_AlreadyDefault` — set the already-default; no error, no extra updated_at churn beyond what the no-op returns.
- `TestSetDefault_NotOwned_ReturnsErrNotOwned`.
- `TestPartialUniqueIndex_HoldsAcrossConcurrentSetDefaults` — fire two `SetDefault` goroutines targeting different addresses; assert exactly one wins (the loser sees a unique-violation; service surfaces it as a generic error — log + return). This is the "the DB is the backstop" assertion.

### Commit message

```
feat(addresses): service with CRUD, SetDefault, auto-promote-on-delete

- service.Create: validation, soft cap (20 per user), first-address-is-
  default auto-flag
- service.Update: PATCH semantics — merge then validate; ErrNotOwned for
  cross-user attempts
- service.Delete: single-statement for non-default; tx-scoped delete +
  promote-oldest-sibling-to-default for the default-deletion case
- service.SetDefault: idempotent; tx-scoped clear-then-set; surfaces
  unique-violation as a 500 (DB is the backstop)
- service_test.go: covers default semantics, cap, validation, PATCH merge,
  ownership rejection, atomic default flip, concurrent-flip backstop
```

---

## Bundle 3 — HTTP handlers + mount + OpenAPI regen + smoke test

**Tasks:** the 5 endpoints, app wiring, mount in main.go, OpenAPI regen, smoke test extension.

### Files
- Create: `backend/internal/addresses/handler.go`
- Create: `backend/internal/addresses/handler_test.go`
- Modify: `backend/internal/app/app.go` — construct AddressService, attach to Application.
- Modify: `backend/cmd/api/main.go` — mount handlers under `/me/addresses*` inside the existing RequireSession group.
- Modify: `backend/cmd/api/main_test.go` — extend smoke with address CRUD.
- Regenerate: `backend/docs/swagger.{json,yaml,go}`.

### Handler shapes

```go
type Handlers struct {
    Svc *Service
    Log *zap.Logger
}

func NewHandlers(svc *Service, log *zap.Logger) *Handlers
func (h *Handlers) Mount(r chi.Router)   // mounts all 5 endpoints; caller mounts inside /me + RequireSession
```

`Mount` shape:
```go
r.Route("/addresses", func(r chi.Router) {
    r.Post("/", h.create)
    r.Get("/", h.list)
    r.Patch("/{id}", h.update)
    r.Delete("/{id}", h.delete)
    r.Post("/{id}/default", h.setDefault)
})
```

### Request/response shapes

`POST /me/addresses` body:
```json
{
  "label": "Home",
  "line1": "...",
  "line2": "...",
  "city": "...",
  "region": "...",
  "phone": "..."
}
```
Response 201:
```json
{
  "id": "...",
  "label": "Home",
  "line1": "...",
  "line2": "...",
  "city": "...",
  "region": "...",
  "phone": "...",
  "is_default": true,
  "created_at": "...",
  "updated_at": "..."
}
```

`GET /me/addresses` response 200:
```json
{
  "addresses": [ <Address>, ... ]
}
```

`PATCH /me/addresses/{id}` body: same fields as POST, all optional. Response 200 = full address row.

`DELETE /me/addresses/{id}` → 204 No Content, empty body.

`POST /me/addresses/{id}/default` → 200 with the address row.

### Error mapping (handler-level)

| Service error | HTTP | `httpx.Code…` |
|---|---|---|
| `ErrInvalidAddress` | 400 | `CodeValidation` |
| `ErrAddressLimitReached` | 400 | `CodeValidation` (with `details: {"code": "address_limit_reached"}` so the frontend can disambiguate) |
| `ErrNotFound` / `ErrNotOwned` | 404 | `CodeNotFound` |
| Anything else | 500 | `CodeInternal` |

### Handler tests

`handler_test.go`:
- `TestCreate_201_HappyPath`
- `TestCreate_400_MissingLine1`
- `TestCreate_400_LimitReached` (seed 20, request 21st)
- `TestList_200_OrdersDefaultFirst`
- `TestUpdate_200_HappyPath`
- `TestUpdate_404_NotOwned` (user B tries to PATCH user A's address)
- `TestUpdate_404_DoesNotExist`
- `TestDelete_204_HappyPath_NonDefault`
- `TestDelete_204_HappyPath_DefaultPromotesNextOldest` (assert DB state after: other address is now default)
- `TestDelete_404_NotOwned`
- `TestSetDefault_200_FlipsAtomically`
- `TestSetDefault_200_IdempotentOnAlreadyDefault`
- `TestSetDefault_404_NotOwned`
- `TestAllEndpoints_401_WithoutSession` — single test loop hitting each endpoint without a session cookie; assert 401 (RequireSession middleware fires).

### main.go mount

Inside the existing `/api/v1` route, inside the existing RequireSession group, inside the existing `/me` route group, add:
```go
addressesHandlers := addresses.NewHandlers(a.Addresses, a.Logger)
addressesHandlers.Mount(meRouter)   // mounts /addresses* relative to /me
```
The exact wiring depends on the established `/me` mount pattern — read `internal/me/handler.go` (or wherever `/me/*` lives today) and mirror it. Do not invent a new mounting style.

### app.New wiring

In `internal/app/app.go`, alongside the existing service constructions:
```go
addressesRepo := addresses.NewRepository(pool)
addressesSvc := addresses.NewService(addressesRepo, pool, logger)
```
Add `Addresses *addresses.Service` to the `Application` struct.

### Smoke test extension

Extend `cmd/api/main_test.go`'s existing flow with:

1. After auth + cart + checkout flow that's already in the smoke test, add an addresses-CRUD section.
2. `POST /api/v1/me/addresses` with a full address payload. Assert 201, capture the returned `id`, assert `is_default=true` (first address).
3. `POST /api/v1/me/addresses` with a second address payload. Assert 201, capture `id2`, assert `is_default=false`.
4. `GET /api/v1/me/addresses`. Assert response is `{addresses: [...]}` with 2 entries; the default (first) is first in the list.
5. `POST /api/v1/me/addresses/{id2}/default`. Assert 200, `is_default=true`.
6. `GET /api/v1/me/addresses` again. Assert the second address is now first (default ordering).
7. `PATCH /api/v1/me/addresses/{id}` with `{label: "Work"}`. Assert 200 and only the label changed.
8. `DELETE /api/v1/me/addresses/{id2}`. Assert 204. (Deleting the current default.)
9. `GET /api/v1/me/addresses`. Assert only the first address remains, and it's now `is_default=true` (auto-promoted because it was the only remaining address).
10. `DELETE /api/v1/me/addresses/{id}`. Assert 204.
11. `GET /api/v1/me/addresses`. Assert `{addresses: []}` (empty array, not null).

### OpenAPI regen + drift-check

```bash
cd backend
PATH="$(go env GOPATH)/bin:$PATH" swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency
```

Verify the 5 new routes appear exactly once each:
```bash
for path in /me/addresses '/me/addresses/{id}' '/me/addresses/{id}/default'; do
    echo -n "$path: "; grep -c "\"$path\"" docs/swagger.json
done
```
Expected: `/me/addresses: 1` (POST + GET share the path), `/me/addresses/{id}: 1` (PATCH + DELETE share), `/me/addresses/{id}/default: 1`.

`make drift-check` and `make test` both exit 0.

### Commit message

```
feat(addresses): /me/addresses CRUD endpoints, mount, smoke

- internal/addresses/handler.go: POST/GET/PATCH/DELETE /me/addresses* and
  POST /me/addresses/{id}/default, all under RequireSession
- error mapping: validation → 400, not-owned/not-found → 404 (IDOR-safe),
  cap-reached → 400 with details.code=address_limit_reached
- app + main: AddressService wired into Application; handlers mounted inside
  the existing /me group
- cmd/api/main_test.go: extends smoke with create-default → second non-
  default → list-ordering → SetDefault flip → PATCH label → delete-default
  auto-promote → delete-last → empty-list
- docs: regenerated openapi with the 5 new operations across 3 paths
```

---

## Verification — end of Plan 6

- `make test` exits 0.
- `make drift-check` exits 0.
- The 3 address paths each appear exactly once in `docs/swagger.json` (5 operations across 3 path entries).
- `/checkout/init` contract from Plan 5 is **unchanged** — verify by `git diff HEAD~3 -- internal/orders/handler.go internal/orders/service.go backend/docs/swagger.json` showing no orders-side modifications.

## Self-Review Notes

- **Why partial unique index, not a CHECK or trigger?** A trigger would work but obscures the invariant; a CHECK can't express "at most one true per user." The partial unique index is the standard Postgres idiom for "at most one row matching a predicate per group" and shows up clearly in `\d addresses`.
- **Why is the soft cap in the service, not the DB?** Caps that need to change later (config-driven, per-user-tier, A/B-tested) want to live in code. Hard-CHECK in the DB locks it in.
- **Why not extend `/checkout/init` to accept `address_id`?** Decided in the brainstorm: keeps the checkout contract stable (Plan 5 already shipped + the frontend is reconciled against it), avoids a second validation path on the hottest endpoint in the system, and the frontend's saved→inline lookup is a small one-shot fetch. Revisit only if measurement shows the extra round-trip matters.
- **Why `ErrNotOwned` distinct from `ErrNotFound` in the service?** Both surface as 404 at the handler, but separating them at the service layer lets tests assert "did we reject because the address doesn't exist or because it isn't yours?" which is a meaningfully different code path. The handler collapses them on the way out.
- **Why DELETE-then-SET-default order in the default-deletion tx?** The partial unique index permits at most one `is_default=true` per user. Inside a tx, if you SET the successor's default to true while the current default still exists with default=true, some Postgres setups will reject the SET (deferred constraints aren't enabled here). DELETE first, then SET — the unique check sees zero defaults at the moment of SET, allowing it.
- **No frontend work in Plan 6.** A future frontend handoff (Plan 12 or earlier) will add: address-book UI in account settings, "save this address" checkbox on checkout, "pick from saved" selector on checkout. Plan 6 ships the backend; frontend is decoupled.
