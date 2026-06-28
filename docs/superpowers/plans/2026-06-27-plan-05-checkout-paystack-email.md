# Checkout + Paystack + Email Implementation Plan (Plan 5 of 15)

> **For executing agents:** This plan is structured as 5 independent bundles. Each bundle is implemented in its own session and committed as ONE commit at the end. Do NOT run the whole plan or final whole-branch review — those gate between bundles externally.

**Goal:** Land the full transactional checkout path — `POST /api/v1/checkout/init`, the Paystack-hosted-payment redirect flow, the public webhook + the auth-gated verify-poll endpoint that converge on a single idempotent `OrderService.MarkPaid`, real Resend-backed order-confirmation email (with allowlist gating), and the `orders` + `order_items` tables underneath.

**Architecture:**
- New `internal/orders/` package: repository + service + handlers + webhook handler. Three-layer per spec §3.1.
- New `internal/payments/paystack/` package: thin REST client (`InitializeTransaction`, `VerifyTransaction`) + HMAC-SHA512 webhook-signature verification primitive. Reusable from the order service AND the webhook handler.
- Email surface evolves: existing `email.LogSender` stub stays. Add `email.ResendSender` (real HTTP) + `email.AllowlistSender` wrapper. The wrapper decides "send via real provider vs log-only" per recipient — this collapses the existing per-service allowlist policy into one place.
- `OrderService.MarkPaid(ctx, reference, transactionID)` is the **single convergence point**. Both webhook and verify-poll call it. Transactional. Idempotent on `paystack_reference`. SELECT … FOR UPDATE so concurrent webhook + poll don't double-process.
- Address handling: **inline only in Plan 5.** The `POST /checkout/init` body carries the address inline; backend snapshots it directly into `orders.shipping_address jsonb`. The saved-addresses CRUD (`POST /me/addresses` etc.) is Plan 6's scope. Frontend (Plan 12) will start with inline-only and extend later.
- `orders.paystack_reference` follows the spec format `RUE-XXXXXXXX` (8 uppercase hex chars after the dash) for legibility in the Paystack dashboard. Generated server-side from a random source. UNIQUE constraint.
- Order status state machine in this plan covers the single transition `pending → paid`. All other transitions (`fulfilled`, `shipped`, `delivered`, `cancelled`, `refunded`) are Plan 7 (admin endpoints).

**Tech Stack:**
- New deps: none required (Paystack client uses `net/http`; HMAC uses `crypto/hmac`/`crypto/sha512`). Optional: `github.com/resendlabs/resend-go` for the Resend SDK — see Bundle 3 for the build-vs-buy decision.
- Reuses: zap, pgx/v5 + sqlc, chi, the existing `auth.RequireSession` middleware, `httpx.WriteError/WriteJSON`, `testsupport.StartPool`, `email.Sender` interface from Plan 3.

## Global Constraints

- **Module path:** `github.com/oti-adjei/ruecosmetics`.
- **Working directory:** `/Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/`. Backend paths relative to `backend/`.
- **Money:** every monetary column `BIGINT` pesewas. Paystack's `amount` field is in pesewas (Paystack uses the smallest currency unit). No conversion math beyond multiplying nothing — the value passes through.
- **Order status enum (DB CHECK):** `pending`, `paid`, `fulfilled`, `shipped`, `delivered`, `cancelled`, `refunded`. Plan 5 only writes `pending` (at init) and `paid` (at MarkPaid).
- **Paystack reference format:** `^RUE-[0-9A-F]{8}$`. Generate via `fmt.Sprintf("RUE-%08X", binary.BigEndian.Uint32(crypto/rand.Read(4 bytes)))` or equivalent. UNIQUE constraint on the column.
- **Cart pricing rule at checkout:** the order service **re-prices from the catalog** at `POST /checkout/init` time. It does NOT trust the cart's `unit_price_ghs_minor` snapshot. The cart's snapshot is for cart UI continuity only (per Plan 4 spec §4.5 forward-looking note). Order items are snapshotted from the **fresh** product lookup: `product_name_snapshot`, `product_brand_snapshot`, `product_image_snapshot`, `unit_price_ghs_minor` all reflect the moment of init.
- **Idempotent MarkPaid:** SELECT FOR UPDATE the order row by `paystack_reference`. If `status != 'pending'`, COMMIT and return — no error, no side effects. This is the rule for both webhook and verify-poll. Concurrent webhook + verify-poll arriving 50ms apart MUST produce exactly one email and one cart-clear.
- **Webhook signature:** HMAC-SHA512 of the raw request body, signed with `PAYSTACK_SECRET_KEY` (the same secret used for outbound auth). Header name `x-paystack-signature`. Constant-time compare via `crypto/subtle`. Invalid sig → 401 (Paystack will retry). Valid sig + unknown reference → 200 (idempotent). Valid sig + already-paid → 200 (idempotent).
- **Cart deletion on payment:** when an order goes `pending → paid`, the user's cart_items are DELETED inside the same transaction as the status update. The cart row itself stays (so future shopping resumes with the same `carts.id`). The cart's `updated_at` is touched.
- **Email allowlist behavior:** the new `email.AllowlistSender` wraps an inner sender and, on `Send(to, ...)`, checks if `to` is in the allowlist (or `*` for "all addresses"). Allowlisted → delegate to inner (Resend or LogSender). Non-allowlisted → log a slog line and return nil. The wrapper REPLACES the per-service allowlist check that auth.Service currently does. Refactor auth.Service in Bundle 3.
- **Order confirmation email:** sent after `MarkPaid` commits, inside the same handler (inline, not queued). If the email send fails, log it but do NOT roll back the order — the order is paid; resending the confirmation is the failure mode, not unpaying.
- **Error codes:** existing `httpx.CodeValidation`, `CodeNotFound`, `CodeUnauthorized`, `CodeInternal` cover this plan. New code `CodeUpstream` = `"upstream_error"` for Paystack outages (503).
- **No `fmt.Sprintf` building SQL.**
- **Commit identity:** `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit ...` on every commit.
- **Bundled commits:** 5 bundles → **5 commits**.
- **HEAD before Plan 5 begins:** `849d317` (Plan 4 Bundle 4 close).
- **New env vars to add to `Config`:**
  - `PAYSTACK_SECRET_KEY` (default empty; checkout/webhook return 503 `not_configured` if empty).
  - `PAYSTACK_BASE_URL` (default `https://api.paystack.co`; override for tests).
  - `PAYSTACK_CALLBACK_URL` (default `http://localhost:5173/checkout/return`; frontend lands here after Paystack).
  - `RESEND_API_KEY` (default empty; if empty, ResendSender refuses to construct and AllowlistSender falls through to LogSender for allowlisted addresses too — so the dev path "just works" without a key).
  - `RESEND_FROM_EMAIL` (default `noreply@rue.example.com`).

## File Structure

```
casestud/ruecosmetics/backend/
├── internal/
│   ├── orders/                                 # NEW
│   │   ├── repository.go                       # sqlc wrappers + ErrNotFound
│   │   ├── repository_test.go                  # integration
│   │   ├── service.go                          # InitCheckout, MarkPaid, VerifyCheckout
│   │   ├── service_test.go
│   │   ├── handler.go                          # POST /checkout/init, GET /checkout/verify/{reference}
│   │   ├── webhook.go                          # POST /webhooks/paystack (separate file for clarity)
│   │   ├── handler_test.go
│   │   └── reference.go                        # GenerateReference() → "RUE-XXXXXXXX"
│   ├── payments/                               # NEW
│   │   └── paystack/
│   │       ├── client.go                       # InitializeTransaction, VerifyTransaction
│   │       ├── client_test.go                  # httptest-driven
│   │       ├── signature.go                    # VerifyWebhookSignature (HMAC-SHA512)
│   │       └── signature_test.go               # known-vector + tampered-payload + missing-header
│   ├── email/                                  # MODIFY
│   │   ├── sender.go                           # existing: Sender interface + LogSender
│   │   ├── resend.go                           # NEW: ResendSender (real HTTP)
│   │   ├── allowlist.go                        # NEW: AllowlistSender wrapper
│   │   ├── templates/                          # NEW
│   │   │   ├── order_confirmation.html.tmpl
│   │   │   └── order_confirmation.txt.tmpl
│   │   └── *_test.go
│   ├── auth/
│   │   └── service.go                          # MODIFY (Bundle 3): drop the per-service allowlist check; trust AllowlistSender
│   ├── config/
│   │   └── config.go                           # MODIFY (Bundle 2 or 3): 5 new env vars
│   ├── app/
│   │   └── app.go                              # MODIFY (Bundle 4): wire Paystack client + AllowlistSender + ResendSender + Orders
│   └── cart/
│       └── repository.go                       # MODIFY (Bundle 4): expose ClearCartByUserID for MarkPaid
├── cmd/api/
│   ├── main.go                                 # MODIFY (Bundle 5): mount /checkout/* + /webhooks/paystack
│   └── main_test.go                            # MODIFY (Bundle 5): smoke test signup → cart → init → simulated webhook → DB-verify
├── migrations/
│   └── 00005_orders.sql                        # NEW (Bundle 1)
├── queries/
│   └── orders.sql                              # NEW (Bundle 1)
└── docs/                                       # REGENERATED (Bundle 5)
```

---

## Bundle 1 — Schema + sqlc + orders Repository

**Tasks:** migration, queries, typed repository, integration tests.

### Files
- Create: `backend/migrations/00005_orders.sql`
- Create: `backend/queries/orders.sql`
- Regenerate: `backend/internal/db/sqlc/`
- Create: `backend/internal/orders/repository.go`
- Create: `backend/internal/orders/repository_test.go`
- Create: `backend/internal/orders/reference.go` + `reference_test.go`

### Migration

File: `backend/migrations/00005_orders.sql`
```sql
-- +goose Up
CREATE TABLE orders (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                   uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status                    text NOT NULL CHECK (status IN ('pending','paid','fulfilled','shipped','delivered','cancelled','refunded')),
    subtotal_ghs_minor        bigint NOT NULL CHECK (subtotal_ghs_minor >= 0),
    shipping_ghs_minor        bigint NOT NULL CHECK (shipping_ghs_minor >= 0),
    total_ghs_minor           bigint NOT NULL CHECK (total_ghs_minor >= 0),
    paystack_reference        text NOT NULL UNIQUE,
    paystack_transaction_id   text,
    shipping_address          jsonb NOT NULL,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

CREATE TABLE order_items (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                    uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id                  uuid NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    qty                         int NOT NULL CHECK (qty >= 1),
    unit_price_ghs_minor        bigint NOT NULL CHECK (unit_price_ghs_minor >= 0),
    product_name_snapshot       text NOT NULL,
    product_brand_snapshot      text NOT NULL DEFAULT '',
    product_image_snapshot      text NOT NULL DEFAULT '',
    created_at                  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- +goose Down
DROP TABLE order_items;
DROP TABLE orders;
```

> Implementer note: `paystack_reference` is `NOT NULL` because an order without a reference is meaningless to us. We generate it server-side before the INSERT.

### Queries

File: `backend/queries/orders.sql`
```sql
-- name: CreateOrder :one
INSERT INTO orders (user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
                    total_ghs_minor, paystack_reference, shipping_address)
VALUES ($1, 'pending', $2, $3, $4, $5, $6)
RETURNING id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
          total_ghs_minor, paystack_reference, paystack_transaction_id,
          shipping_address, created_at, updated_at;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, qty, unit_price_ghs_minor,
                         product_name_snapshot, product_brand_snapshot,
                         product_image_snapshot)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, order_id, product_id, qty, unit_price_ghs_minor,
          product_name_snapshot, product_brand_snapshot,
          product_image_snapshot, created_at;

-- name: GetOrderByReference :one
SELECT id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
       total_ghs_minor, paystack_reference, paystack_transaction_id,
       shipping_address, created_at, updated_at
FROM orders WHERE paystack_reference = $1;

-- name: GetOrderByID :one
SELECT id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
       total_ghs_minor, paystack_reference, paystack_transaction_id,
       shipping_address, created_at, updated_at
FROM orders WHERE id = $1;

-- name: ListOrderItems :many
SELECT id, order_id, product_id, qty, unit_price_ghs_minor,
       product_name_snapshot, product_brand_snapshot,
       product_image_snapshot, created_at
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: GetOrderByReferenceForUpdate :one
SELECT id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor,
       total_ghs_minor, paystack_reference, paystack_transaction_id,
       shipping_address, created_at, updated_at
FROM orders WHERE paystack_reference = $1
FOR UPDATE;

-- name: MarkOrderPaid :exec
UPDATE orders
SET status = 'paid',
    paystack_transaction_id = $2,
    updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: CountOrdersByStatus :one
SELECT count(*) FROM orders WHERE status = $1;
```

### Reference generator

File: `backend/internal/orders/reference.go`
```go
package orders

import (
    "crypto/rand"
    "encoding/binary"
    "fmt"
)

// GenerateReference returns a Paystack reference of the form "RUE-XXXXXXXX"
// where X is a random uppercase hex digit. ~4 billion combinations; collision
// probability is negligible at case-study scale and the DB UNIQUE constraint
// is the backstop.
func GenerateReference() (string, error) {
    var b [4]byte
    if _, err := rand.Read(b[:]); err != nil {
        return "", err
    }
    return fmt.Sprintf("RUE-%08X", binary.BigEndian.Uint32(b[:])), nil
}
```

Tests:
- Format matches regex `^RUE-[0-9A-F]{8}$`.
- Two calls produce different strings (very high probability — assert inequality, not statistical uniqueness).

### Repository

File: `backend/internal/orders/repository.go` — typed wrappers exposing:

```go
type Repository struct { /* *sqlcq.Queries + db.Pool */ }
func NewRepository(pool db.Pool) *Repository
func (r *Repository) Pool() db.Pool
var ErrNotFound = errors.New("orders: not found")

func (r *Repository) CreateOrder(ctx, params sqlcq.CreateOrderParams) (sqlcq.Order, error)
func (r *Repository) CreateOrderItem(ctx, params sqlcq.CreateOrderItemParams) (sqlcq.OrderItem, error)
func (r *Repository) GetOrderByReference(ctx, ref string) (sqlcq.Order, error)   // ErrNotFound on miss
func (r *Repository) GetOrderByID(ctx, id uuid.UUID) (sqlcq.Order, error)        // ErrNotFound on miss
func (r *Repository) ListOrderItems(ctx, orderID uuid.UUID) ([]sqlcq.OrderItem, error)
func (r *Repository) CountOrdersByStatus(ctx, status string) (int64, error)
```

The tx-aware variants (`GetOrderByReferenceForUpdate`, `MarkOrderPaid`) are NOT exposed on the Repository — they're called inside `db.WithTx` closures using `sqlcq.New(tx)`, mirroring the established pattern from `cart.Service.MergeGuestCart` and `auth.Service.Signup`.

### Tests

`repository_test.go`: integration covering happy paths, ErrNotFound on miss, paystack_reference UNIQUE violation (insert same reference twice → error), and the order_items FK cascade (delete order → items gone). Use `testsupport.StartPool(t, "../../migrations")`.

### Commit message (single commit at end of bundle)

```
feat(orders): schema, sqlc queries, repository (foundation)

- migrations/00005_orders.sql: orders + order_items with CHECK status enum,
  UNIQUE paystack_reference, NOT NULL shipping_address jsonb, FK CASCADE on items
- queries/orders.sql: Create/Get*/ListItems/CountByStatus + tx-only
  GetOrderByReferenceForUpdate and MarkOrderPaid
- internal/orders/repository.go: typed wrappers + ErrNotFound
- internal/orders/reference.go: GenerateReference -> "RUE-XXXXXXXX"
- repository_test.go: happy paths, UNIQUE violation, FK cascade
```

---

## Bundle 2 — Paystack client + webhook signature primitives

**Tasks:** thin REST client, HMAC-SHA512 signature verifier, all tests.

### Files
- Create: `backend/internal/payments/paystack/client.go`
- Create: `backend/internal/payments/paystack/client_test.go`
- Create: `backend/internal/payments/paystack/signature.go`
- Create: `backend/internal/payments/paystack/signature_test.go`
- Modify: `backend/internal/config/config.go` — add 3 Paystack env vars.
- Modify: `backend/.env.example` — document them.

### Config additions

In `internal/config/config.go`, append to the `Config` struct:
```go
PaystackSecretKey   string `envconfig:"PAYSTACK_SECRET_KEY" default:""`
PaystackBaseURL     string `envconfig:"PAYSTACK_BASE_URL" default:"https://api.paystack.co"`
PaystackCallbackURL string `envconfig:"PAYSTACK_CALLBACK_URL" default:"http://localhost:5173/checkout/return"`
```

Extend `config_test.go` with assertions on the three new fields.

Append to `backend/.env.example`:
```
# Paystack (Plan 5)
# Get a test secret key from https://dashboard.paystack.com/#/settings/developer
# PAYSTACK_SECRET_KEY=sk_test_PLACEHOLDER
# PAYSTACK_BASE_URL=https://api.paystack.co
PAYSTACK_CALLBACK_URL=http://localhost:5173/checkout/return
```

### Paystack client

File: `backend/internal/payments/paystack/client.go`
```go
// Package paystack is a thin REST client for the two Paystack endpoints we
// need: transaction/initialize and transaction/verify. It also exposes
// VerifyWebhookSignature for the webhook handler.
//
// The client makes NO ASSUMPTION about the surrounding service — it's
// reusable from anywhere, returns typed responses, and never panics on
// missing fields (Paystack occasionally omits optional fields).
package paystack

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    BaseURL    string  // e.g., "https://api.paystack.co"
    SecretKey  string  // sk_test_xxxxx or sk_live_xxxxx
    HTTPClient *http.Client
}

func NewClient(baseURL, secretKey string) *Client {
    return &Client{
        BaseURL:   baseURL,
        SecretKey: secretKey,
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
    }
}

// IsConfigured returns true when both BaseURL and SecretKey are non-empty.
// Handlers should 503 with "not_configured" otherwise.
func (c *Client) IsConfigured() bool {
    return c.BaseURL != "" && c.SecretKey != ""
}

// InitializeTransactionInput models the subset of fields we send.
type InitializeTransactionInput struct {
    Email     string `json:"email"`
    Amount    int64  `json:"amount"`                // pesewas
    Reference string `json:"reference"`             // "RUE-XXXXXXXX"
    Callback  string `json:"callback_url,omitempty"`
    Currency  string `json:"currency,omitempty"`    // "GHS"
    Channels  []string `json:"channels,omitempty"`  // e.g. ["card","mobile_money"]
}

// InitializeTransactionOutput is the relevant subset of Paystack's response.
type InitializeTransactionOutput struct {
    AuthorizationURL string `json:"authorization_url"`
    AccessCode       string `json:"access_code"`
    Reference        string `json:"reference"`
}

var ErrPaystackUpstream = errors.New("paystack: upstream error")

// InitializeTransaction calls POST /transaction/initialize. On Paystack
// returning non-2xx, the error wraps ErrPaystackUpstream.
func (c *Client) InitializeTransaction(ctx context.Context, in InitializeTransactionInput) (InitializeTransactionOutput, error) {
    var out InitializeTransactionOutput
    body, _ := json.Marshal(in)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/transaction/initialize", bytes.NewReader(body))
    if err != nil {
        return out, err
    }
    req.Header.Set("Authorization", "Bearer "+c.SecretKey)
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return out, fmt.Errorf("%w: %v", ErrPaystackUpstream, err)
    }
    defer resp.Body.Close()
    raw, _ := io.ReadAll(resp.Body)
    if resp.StatusCode/100 != 2 {
        return out, fmt.Errorf("%w: status %d body=%s", ErrPaystackUpstream, resp.StatusCode, string(raw))
    }
    var envelope struct {
        Status  bool                          `json:"status"`
        Message string                        `json:"message"`
        Data    InitializeTransactionOutput   `json:"data"`
    }
    if err := json.Unmarshal(raw, &envelope); err != nil {
        return out, fmt.Errorf("%w: parse: %v", ErrPaystackUpstream, err)
    }
    if !envelope.Status {
        return out, fmt.Errorf("%w: %s", ErrPaystackUpstream, envelope.Message)
    }
    return envelope.Data, nil
}

// VerifyTransactionOutput is the subset we read from the verify endpoint.
type VerifyTransactionOutput struct {
    Reference       string `json:"reference"`
    Status          string `json:"status"`   // "success", "failed", "abandoned", ...
    Amount          int64  `json:"amount"`
    Currency        string `json:"currency"`
    GatewayResponse string `json:"gateway_response"`
    TransactionID   int64  `json:"id"`       // Paystack's numeric transaction id; convert to string at the boundary
}

// VerifyTransaction calls GET /transaction/verify/{reference}.
func (c *Client) VerifyTransaction(ctx context.Context, reference string) (VerifyTransactionOutput, error) {
    var out VerifyTransactionOutput
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/transaction/verify/"+reference, nil)
    if err != nil {
        return out, err
    }
    req.Header.Set("Authorization", "Bearer "+c.SecretKey)
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return out, fmt.Errorf("%w: %v", ErrPaystackUpstream, err)
    }
    defer resp.Body.Close()
    raw, _ := io.ReadAll(resp.Body)
    if resp.StatusCode/100 != 2 {
        return out, fmt.Errorf("%w: status %d body=%s", ErrPaystackUpstream, resp.StatusCode, string(raw))
    }
    var envelope struct {
        Status  bool                    `json:"status"`
        Message string                  `json:"message"`
        Data    VerifyTransactionOutput `json:"data"`
    }
    if err := json.Unmarshal(raw, &envelope); err != nil {
        return out, fmt.Errorf("%w: parse: %v", ErrPaystackUpstream, err)
    }
    if !envelope.Status {
        return out, fmt.Errorf("%w: %s", ErrPaystackUpstream, envelope.Message)
    }
    return envelope.Data, nil
}
```

### Signature verifier

File: `backend/internal/payments/paystack/signature.go`
```go
package paystack

import (
    "crypto/hmac"
    "crypto/sha512"
    "encoding/hex"
)

// VerifyWebhookSignature checks the x-paystack-signature header against
// HMAC-SHA512(secret, body). Returns true iff the signatures match in
// constant time.
func VerifyWebhookSignature(secret string, body []byte, headerValue string) bool {
    if secret == "" || headerValue == "" || len(body) == 0 {
        return false
    }
    mac := hmac.New(sha512.New, []byte(secret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(headerValue))
}
```

### Tests

`client_test.go`: spin up an `httptest.NewServer` whose handler asserts the bearer token + path + body shape, returns canned Paystack envelopes (success + failure variants). Exercise:
- Happy path on initialize → returns authorization_url.
- Happy path on verify → returns reference + status="success".
- Non-2xx response → ErrPaystackUpstream.
- `status: false` envelope → ErrPaystackUpstream wrapping the message.
- Malformed JSON → ErrPaystackUpstream.

`signature_test.go`: known vectors.
- A canonical body + secret produces a known HMAC-SHA512 hex string (compute once with `openssl dgst -sha512 -hmac SECRET` and embed in the test).
- Tampered body → returns false.
- Empty header → false.
- Empty body → false.
- Empty secret → false.

### Commit message

```
feat(payments): Paystack client (initialize/verify) and webhook HMAC

- internal/payments/paystack/client.go: Client.InitializeTransaction +
  VerifyTransaction with ErrPaystackUpstream wrap
- internal/payments/paystack/signature.go: VerifyWebhookSignature using
  HMAC-SHA512 + constant-time compare
- internal/config: PAYSTACK_SECRET_KEY, PAYSTACK_BASE_URL,
  PAYSTACK_CALLBACK_URL envs
- client_test.go: httptest-driven happy/failure paths
- signature_test.go: known HMAC vector + tampered + missing variants
```

---

## Bundle 3 — Email enhancements (AllowlistSender + ResendSender + templates)

**Tasks:** introduce AllowlistSender wrapper, ResendSender implementation, two simple HTML/text templates, refactor `auth.Service` to drop its per-service allowlist check.

### Files
- Create: `backend/internal/email/allowlist.go`
- Create: `backend/internal/email/allowlist_test.go`
- Create: `backend/internal/email/resend.go`
- Create: `backend/internal/email/resend_test.go`
- Create: `backend/internal/email/templates/order_confirmation.html.tmpl`
- Create: `backend/internal/email/templates/order_confirmation.txt.tmpl`
- Create: `backend/internal/email/templates.go` — embed.FS loader + Render(name, data) → (html, text, error)
- Create: `backend/internal/email/templates_test.go`
- Modify: `backend/internal/auth/service.go` — drop the `isAllowlisted` gate on Send calls (always Send; let AllowlistSender decide).
- Modify: `backend/internal/auth/service_test.go` — `TestSignupAllowlistedSendsVerifyToken` and similar tests now stub a Sender that records calls; allowlist behavior is asserted via the wrapper's tests.
- Modify: `backend/internal/config/config.go` — `RESEND_API_KEY`, `RESEND_FROM_EMAIL` envs.

### AllowlistSender

File: `backend/internal/email/allowlist.go`
```go
package email

import (
    "context"
    "log/slog"
    "strings"
)

// AllowlistSender wraps an inner Sender and decides at Send-time whether
// the recipient receives a real delivery. Non-allowlisted recipients are
// logged at Info and the call returns nil (so the caller's flow continues
// as if mail had been sent).
//
// The allowlist is a slice of lowercase email addresses; a single entry "*"
// allowlists every address.
type AllowlistSender struct {
    Inner     Sender
    Allowlist []string
    Log       *zap.Logger   // for the non-delivery path
}

func NewAllowlistSender(inner Sender, raw []string, log *zap.Logger) AllowlistSender {
    out := make([]string, 0, len(raw))
    for _, s := range raw {
        if s = strings.TrimSpace(strings.ToLower(s)); s != "" {
            out = append(out, s)
        }
    }
    return AllowlistSender{Inner: inner, Allowlist: out, Log: log}
}

func (s AllowlistSender) Send(ctx context.Context, to, template string, data map[string]any) error {
    addr := strings.ToLower(strings.TrimSpace(to))
    for _, a := range s.Allowlist {
        if a == "*" || a == addr {
            return s.Inner.Send(ctx, to, template, data)
        }
    }
    s.Log.Info("email suppressed (not allowlisted)",
        zap.String("to", addr),
        zap.String("template", template))
    return nil
}
```

> The import for zap: `"go.uber.org/zap"`. The brief shows `slog` in the import comment by mistake; ignore that.

### ResendSender

File: `backend/internal/email/resend.go`

Build vs buy: Resend's official Go SDK is `github.com/resendlabs/resend-go`. It's a small library (one file, no transitive deps beyond stdlib). **Use it.** Adds a real dep but saves a few hundred lines of HTTP-client + retry boilerplate.

```bash
cd backend
go get github.com/resendlabs/resend-go@latest
go mod tidy
```

```go
package email

import (
    "context"
    "errors"

    "github.com/resendlabs/resend-go"
    "go.uber.org/zap"
)

var ErrResendNotConfigured = errors.New("email: resend not configured")

type ResendSender struct {
    Client    *resend.Client
    FromEmail string
    Renderer  *Renderer   // see templates.go
    Log       *zap.Logger
}

// NewResendSender returns a sender bound to a real Resend account. If apiKey
// or fromEmail is empty, returns (nil, ErrResendNotConfigured) so app.New
// can fall back to a LogSender.
func NewResendSender(apiKey, fromEmail string, r *Renderer, log *zap.Logger) (*ResendSender, error) {
    if apiKey == "" || fromEmail == "" {
        return nil, ErrResendNotConfigured
    }
    return &ResendSender{
        Client: resend.NewClient(apiKey),
        FromEmail: fromEmail,
        Renderer: r,
        Log: log,
    }, nil
}

func (s *ResendSender) Send(ctx context.Context, to, template string, data map[string]any) error {
    html, text, err := s.Renderer.Render(template, data)
    if err != nil {
        return err
    }
    subject := subjectFor(template, data)
    _, err = s.Client.Emails.Send(&resend.SendEmailRequest{
        From:    s.FromEmail,
        To:      []string{to},
        Subject: subject,
        Html:    html,
        Text:    text,
    })
    if err != nil {
        return err
    }
    return nil
}

// subjectFor maps a template name + data to a subject line. Keep this tiny;
// proliferate as templates grow.
func subjectFor(template string, data map[string]any) string {
    switch template {
    case "verify_email":
        return "Verify your email for Rue Cosmetics"
    case "password_reset":
        return "Reset your password"
    case "welcome":
        return "Welcome to Rue Cosmetics"
    case "order_confirmation":
        if ref, ok := data["paystack_reference"].(string); ok {
            return "Your Rue Cosmetics order " + ref
        }
        return "Your Rue Cosmetics order"
    }
    return "Rue Cosmetics"
}
```

### Templates

File: `backend/internal/email/templates/order_confirmation.html.tmpl`
```html
<!doctype html>
<html><body style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
  <h1>Thanks for your order!</h1>
  <p>Hi {{.name}},</p>
  <p>Your order <strong>{{.paystack_reference}}</strong> is confirmed and on its way to being packed.</p>
  <h2>Order summary</h2>
  <table style="width: 100%; border-collapse: collapse;">
    <thead>
      <tr><th align="left">Item</th><th align="right">Qty</th><th align="right">Price</th></tr>
    </thead>
    <tbody>
      {{- range .items}}
      <tr>
        <td>{{.name}}{{if .brand}} <em>by {{.brand}}</em>{{end}}</td>
        <td align="right">{{.qty}}</td>
        <td align="right">GHS {{.line_total}}</td>
      </tr>
      {{- end}}
    </tbody>
    <tfoot>
      <tr><td colspan="2" align="right"><strong>Subtotal</strong></td><td align="right">GHS {{.subtotal}}</td></tr>
      <tr><td colspan="2" align="right"><strong>Shipping</strong></td><td align="right">GHS {{.shipping}}</td></tr>
      <tr><td colspan="2" align="right"><strong>Total</strong></td><td align="right"><strong>GHS {{.total}}</strong></td></tr>
    </tfoot>
  </table>
  <p>We'll email you again when it ships.</p>
  <p>— The Rue Cosmetics team</p>
</body></html>
```

File: `backend/internal/email/templates/order_confirmation.txt.tmpl`
```
Thanks for your order, {{.name}}!

Order: {{.paystack_reference}}

{{range .items -}}
  - {{.name}}{{if .brand}} by {{.brand}}{{end}} × {{.qty}} — GHS {{.line_total}}
{{end}}
Subtotal: GHS {{.subtotal}}
Shipping: GHS {{.shipping}}
Total:    GHS {{.total}}

We'll email again when it ships.
— The Rue Cosmetics team
```

> The pesewas-to-GHS formatting: every monetary value in `data` is already a STRING formatted as `"245.00"` by the caller (the order service formats `int64` pesewas before passing the map). Don't do arithmetic inside templates.

### Renderer

File: `backend/internal/email/templates.go`
```go
package email

import (
    "bytes"
    "embed"
    "fmt"
    "html/template"
    "io/fs"
    texttemplate "text/template"
)

//go:embed templates
var templatesFS embed.FS

type Renderer struct {
    html *template.Template
    text *texttemplate.Template
}

func NewRenderer() (*Renderer, error) {
    html, err := template.ParseFS(fs.FS(templatesFS), "templates/*.html.tmpl")
    if err != nil {
        return nil, fmt.Errorf("email: parse html templates: %w", err)
    }
    text, err := texttemplate.ParseFS(fs.FS(templatesFS), "templates/*.txt.tmpl")
    if err != nil {
        return nil, fmt.Errorf("email: parse text templates: %w", err)
    }
    return &Renderer{html: html, text: text}, nil
}

// Render returns (htmlBody, textBody, error). The template name is the
// logical name (e.g., "order_confirmation"); the renderer looks up
// {name}.html.tmpl and {name}.txt.tmpl.
func (r *Renderer) Render(name string, data map[string]any) (string, string, error) {
    var hb bytes.Buffer
    if err := r.html.ExecuteTemplate(&hb, name+".html.tmpl", data); err != nil {
        return "", "", fmt.Errorf("email: html render %s: %w", name, err)
    }
    var tb bytes.Buffer
    if err := r.text.ExecuteTemplate(&tb, name+".txt.tmpl", data); err != nil {
        return "", "", fmt.Errorf("email: text render %s: %w", name, err)
    }
    return hb.String(), tb.String(), nil
}
```

### Refactor auth.Service to drop its allowlist gate

In `internal/auth/service.go`:
- Drop the `Allowlist []string` field on Service.
- Drop the `isAllowlisted(email)` method.
- Drop the `normalizeAllowlist` function.
- In `Signup`, replace the `allow := s.isAllowlisted(in.Email)` block + the conditional logic with: **always** create the verify token and call `s.Email.Send(ctx, in.Email, "verify_email", ...)`. The `AllowlistSender` (set up in app.New) decides whether the email actually goes out.
- BUT — the `email_verified` auto-flag behavior was tied to the allowlist. That behavior must move somewhere. Two options:
  - (a) The Sender returns "actually sent" vs "suppressed" and the service marks `email_verified` based on the result.
  - (b) `auth.Service` keeps the allowlist FOR THE PURPOSE OF AUTO-VERIFY ONLY, separated from the Send decision.
  - Pick (b): `auth.Service` keeps an `Allowlist []string` field. Its only remaining purpose is the auto-verify-on-non-allowlist decision at signup. The Send call is unconditional; the wrapper handles delivery. Refactor the existing `isAllowlisted` to a thinner `autoVerifyDecision(emailAddr) bool` (true = mark verified at signup; false = require verify link).
- Update `auth.Service.Signup` accordingly. `ResendVerification` calls `Send` unconditionally; if the user is auto-verified, no verify token row was created at signup, so `ResendVerification` should refuse to act on an already-verified user (return nil silently).
- Update `auth.Service_test.go`: tests that previously asserted "non-allowlisted gets auto-verified" still pass because the auto-verify path is unchanged in semantics; only the Send call's outcome is different. Tests that asserted "allowlisted gets a real email" now assert it via the captured Sender (which is wrapped in AllowlistSender in the integration but the unit tests can use a non-wrapped capturing sender).

### Tests

`allowlist_test.go`:
- Inner sender records all Send calls. Wrapper with `[]string{"vip@y.test"}` allowlist: Send to "vip@y.test" → inner.Send called once with same args. Send to "rando@y.test" → inner.Send NOT called, log line emitted.
- "*" allowlist → every Send delegated.
- Empty allowlist → no Send delegated.
- Case-insensitive match: allowlist `["VIP@Y.TEST"]`, Send to `"vip@y.test"` → matches.

`resend_test.go`:
- `NewResendSender("", "from", r, log)` → returns ErrResendNotConfigured.
- `NewResendSender("key", "", r, log)` → returns ErrResendNotConfigured.
- `Send` happy path with a stub Renderer (or use the real one against a tiny test template) AND a stub `resend.Client` if `resend-go` exposes a swappable transport. If it doesn't, skip the live-HTTP test here and rely on the AllowlistSender path covering "real send" via integration.

`templates_test.go`:
- Render `order_confirmation` with realistic data → output contains expected strings (the name, the reference, the total).
- Render with a missing key (e.g., no `name`) → `<no value>` shows up; assert the renderer doesn't panic.

### Config additions

In `internal/config/config.go`:
```go
ResendAPIKey    string `envconfig:"RESEND_API_KEY" default:""`
ResendFromEmail string `envconfig:"RESEND_FROM_EMAIL" default:"noreply@rue.example.com"`
```

`.env.example`:
```
# Email (Plan 5)
# RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxx
RESEND_FROM_EMAIL=noreply@rue.example.com
```

### Commit message

```
feat(email): AllowlistSender, ResendSender, templates; simplify auth allowlist

- email.AllowlistSender: per-recipient gate on whether to delegate to the
  inner Sender (Resend) or log-suppress
- email.ResendSender: real HTTP via resend-go; refuses construction without
  api key + from email
- email/templates: embed.FS-loaded html + text order_confirmation templates
  rendered by Renderer
- auth.Service: drop isAllowlisted-gates-Send pattern; keep allowlist only
  for the auto-verify-at-signup decision; Send is unconditional, the
  wrapper decides
- internal/config: RESEND_API_KEY, RESEND_FROM_EMAIL envs
```

---

## Bundle 4 — OrderService (InitCheckout, MarkPaid, VerifyCheckout)

**Tasks:** the heart of the plan. Three service methods with deep DB integration, transactional MarkPaid that converges with the cart cleanup, idempotent everywhere.

### Files
- Create: `backend/internal/orders/service.go`
- Create: `backend/internal/orders/service_test.go`
- Modify: `backend/internal/cart/repository.go` — add `ClearItemsByUserCart(ctx, userID uuid.UUID) error` (or expose a tx-scoped variant). Discussed below.

### Cart clear helper

`MarkPaid` empties the user's cart inside the same tx as the order status update. Two options:

(a) Add `cart.Repository.ClearItemsByUserCart(ctx, userID)` that does `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM carts WHERE user_id = $1)`. Pool-level method.

(b) Inline the SQL inside the MarkPaid tx using `sqlcq.New(tx)` directly.

Pick (b) for atomicity. Add a new sqlc query:
```sql
-- name: DeleteCartItemsByUserID :exec
DELETE FROM cart_items
WHERE cart_id IN (SELECT id FROM carts WHERE user_id = $1);
```
Append to `backend/queries/cart.sql`. Regenerate.

MarkPaid's tx closure calls `q.DeleteCartItemsByUserID(ctx, userID)` after the status flip.

### OrderService

File: `backend/internal/orders/service.go`

Exports:
```go
type Service struct {
    Repo     *Repository
    Cart     *cart.Service                 // for re-pricing the cart at init
    Catalog  *catalog.Repository           // for product + brand lookups at snapshot time
    Shipping *shipping.Service
    Paystack *paystack.Client
    Email    email.Sender                   // AllowlistSender-wrapped Resend or LogSender
    Pool     db.Pool
    Log      *zap.Logger
    Now      func() time.Time
    CallbackURL string                      // from cfg.PaystackCallbackURL
}

func NewService(repo *Repository, cartSvc *cart.Service, catalog *catalog.Repository, ship *shipping.Service,
                paystack *paystack.Client, mail email.Sender, pool db.Pool, log *zap.Logger,
                callbackURL string) *Service

type ShippingAddress struct {
    Line1   string `json:"line1"`
    Line2   string `json:"line2,omitempty"`
    City    string `json:"city"`
    Region  string `json:"region"`
    Phone   string `json:"phone"`
    Label   string `json:"label,omitempty"`
}

type InitCheckoutInput struct {
    UserID          uuid.UUID
    UserEmail       string                  // captured from session; used as Paystack customer email
    UserName        string                  // for the email template
    ShippingAddress ShippingAddress
    ShippingMethod  string                  // "standard" for v1; ignored but recorded
}

type InitCheckoutOutput struct {
    OrderID          uuid.UUID
    Reference        string
    AuthorizationURL string
    TotalGhsMinor    int64
}

var (
    ErrEmptyCart        = errors.New("orders: cart is empty")
    ErrInvalidAddress   = errors.New("orders: invalid shipping address")
    ErrPaystackNotReady = errors.New("orders: paystack not configured")
)

func (s *Service) InitCheckout(ctx context.Context, in InitCheckoutInput) (InitCheckoutOutput, error)
func (s *Service) MarkPaid(ctx context.Context, reference string, paystackTransactionID string) error
func (s *Service) VerifyCheckout(ctx context.Context, reference string) (string, error)   // returns the order status
```

### InitCheckout implementation

1. Validate `ShippingAddress` (line1, city, region, phone non-empty). On miss → `ErrInvalidAddress`.
2. Validate Paystack is configured. If `!s.Paystack.IsConfigured()` → `ErrPaystackNotReady`.
3. Resolve the user's cart via the cart service. If empty → `ErrEmptyCart`.
4. RE-PRICE THE CART FROM CATALOG (do NOT trust cart's `unit_price_ghs_minor`):
   - For each cart item, look up `catalog.GetProductByID(item.ProductID)` to get current `price_ghs_minor`, `name`, `brand_id`, `image_path`.
   - Look up brand name via `catalog.GetBrandByID(brand_id)` (add this query if absent — see "Optional catalog repo additions" below).
   - Compute `subtotal = sum(qty * fresh_price)`. Compute shipping via `s.Shipping.Quote(subtotal).AppliedCostGhsMinor`. `total = subtotal + shipping`.
5. Generate a reference via `GenerateReference()`.
6. Marshal the `ShippingAddress` into JSON for the jsonb column.
7. `db.WithTx`:
   - `q.CreateOrder(...)` with all the totals + reference + address.
   - For each cart item: `q.CreateOrderItem(orderID, productID, qty, freshPrice, productName, brandName, imagePath)`.
8. After commit (NOT inside the tx — the external HTTP call shouldn't gate the DB write):
   - Call `s.Paystack.InitializeTransaction(ctx, ...)`. If it fails, the order row remains `pending` with no authorization_url; return the error wrapped. The frontend will report "checkout init failed."
9. Return `{OrderID, Reference, AuthorizationURL, TotalGhsMinor}`.

> Implementer note: the cart is NOT cleared at this step. Clear only happens on payment confirmation (MarkPaid). So a user can re-try checkout if Paystack init fails or if they abandon.

### Optional catalog repo addition

If `catalog.Repository.GetBrandByID(ctx, uuid.UUID)` doesn't exist, add it. Mirrors `GetProductByID`. Add the sqlc query if needed:
```sql
-- name: GetBrandByID :one
SELECT id, slug, name FROM brands WHERE id = $1;
```

### MarkPaid implementation

```go
func (s *Service) MarkPaid(ctx context.Context, reference, paystackTransactionID string) error {
    var userIDForEmail uuid.UUID
    var sendEmailAfterCommit *emailPayload    // populated only on actual transition

    err := db.WithTx(ctx, s.Pool, func(tx pgx.Tx) error {
        q := sqlcq.New(tx)
        order, err := q.GetOrderByReferenceForUpdate(ctx, reference)
        if errors.Is(err, pgx.ErrNoRows) {
            return ErrNotFound
        }
        if err != nil {
            return err
        }
        if order.Status != "pending" {
            // Idempotent no-op. Either already paid, or a terminal state.
            // No transition log; just succeed.
            return nil
        }
        if err := q.MarkOrderPaid(ctx, sqlcq.MarkOrderPaidParams{
            ID: order.ID,
            PaystackTransactionID: pgxNullable(paystackTransactionID),
        }); err != nil {
            return err
        }
        if err := q.DeleteCartItemsByUserID(ctx, order.UserID); err != nil {
            return err
        }
        // Build the email payload from the tx-consistent snapshot.
        items, err := q.ListOrderItems(ctx, order.ID)
        if err != nil {
            return err
        }
        // Look up user email/name for the template.
        user, err := q.GetUserByID(ctx, order.UserID)
        if err != nil {
            return err
        }
        userIDForEmail = user.ID
        sendEmailAfterCommit = buildOrderEmailPayload(user, order, items)
        return nil
    })
    if err != nil {
        return err
    }
    if sendEmailAfterCommit != nil {
        if sendErr := s.Email.Send(ctx, sendEmailAfterCommit.To,
            "order_confirmation", sendEmailAfterCommit.Data); sendErr != nil {
            s.Log.Error("order confirmation send failed",
                zap.String("user_id", userIDForEmail.String()),
                zap.String("reference", reference),
                zap.Error(sendErr))
            // Do NOT return the error. The order is paid; the email is best-effort.
        }
    }
    return nil
}
```

Where:
- `pgxNullable(s string)` returns `pgtype.Text{String: s, Valid: s != ""}` (or whatever the sqlc-emitted nullable text type is — adapt to actual generated code).
- `buildOrderEmailPayload(user, order, items)` produces a struct with the recipient and a `map[string]any` for the template, formatting pesewas to GHS strings (`"245.00"`).

`ErrNotFound` here is the orders-package sentinel; webhook handler treats it as 200 (Paystack might webhook a reference we've never seen because of a clock skew or test misconfig — don't 5xx).

### VerifyCheckout implementation

```go
// VerifyCheckout looks up the order locally. If still pending, calls
// Paystack's verify endpoint. On Paystack reporting "success", calls
// MarkPaid (which is idempotent — if the webhook already arrived, MarkPaid
// is a no-op). Returns the order's final status (or "pending" if Paystack
// hasn't confirmed yet).
func (s *Service) VerifyCheckout(ctx context.Context, reference string) (string, error) {
    order, err := s.Repo.GetOrderByReference(ctx, reference)
    if errors.Is(err, ErrNotFound) {
        return "", ErrNotFound
    }
    if err != nil {
        return "", err
    }
    if order.Status != "pending" {
        return order.Status, nil
    }
    if !s.Paystack.IsConfigured() {
        return "", ErrPaystackNotReady
    }
    res, err := s.Paystack.VerifyTransaction(ctx, reference)
    if err != nil {
        return "", err
    }
    if res.Status != "success" {
        // Still pending from our side; Paystack hasn't confirmed.
        return "pending", nil
    }
    txID := strconv.FormatInt(res.TransactionID, 10)
    if err := s.MarkPaid(ctx, reference, txID); err != nil {
        return "", err
    }
    return "paid", nil
}
```

### Tests

`service_test.go`:
- `TestInitCheckout_HappyPath` — seed cart with 2 products, call InitCheckout (with a stubbed Paystack server via httptest), assert order row exists with status=pending, all order_items rows present with correct snapshots (name, brand, image, price), total computed from FRESH prices.
- `TestInitCheckout_EmptyCart_ReturnsErrEmptyCart`.
- `TestInitCheckout_RePricesFromCatalog` — add item to cart at price P1, bump product price to P2, call InitCheckout, assert order_items.unit_price = P2 (NOT P1).
- `TestInitCheckout_PaystackInitFails_OrderRemainsPending` — Paystack stub returns 500; assert order row still exists with status=pending, no authorization_url returned.
- `TestMarkPaid_IdempotentSecondCall` — call MarkPaid twice; assert status=paid (unchanged), assert email sent ONLY ONCE (capture via stub Sender).
- `TestMarkPaid_DeletesUserCartItems` — populate user cart, call MarkPaid, assert cart_items rows for that user are gone, cart row remains.
- `TestMarkPaid_EmailFailureDoesNotRollback` — stub Sender returns error; assert order status=paid (committed) despite the send failure.
- `TestVerifyCheckout_OrderAlreadyPaid_ShortCircuits` — pre-populate paid order; assert Paystack stub was NOT called.
- `TestVerifyCheckout_PaystackReportsSuccess_CallsMarkPaid` — pre-populate pending order; stub Paystack verify returns success; assert order transitions to paid.
- `TestVerifyCheckout_PaystackReportsFailure_StaysPending` — stub returns status="failed"; assert order still pending.

### Commit message

```
feat(orders): InitCheckout, MarkPaid (idempotent), VerifyCheckout

- service.InitCheckout: re-prices cart from catalog, snapshots fresh
  name/brand/image/price into order_items, persists order pending,
  initializes Paystack transaction, returns authorization_url
- service.MarkPaid: tx-scoped pending→paid transition, deletes user's
  cart_items in the same tx, idempotent on any non-pending status,
  sends order_confirmation email AFTER commit (best-effort, never rolls
  back the order on send failure)
- service.VerifyCheckout: convergence helper for the poll path; calls
  Paystack verify, then MarkPaid on success
- queries/cart.sql: DeleteCartItemsByUserID for the tx-scoped cart clear
- catalog: GetBrandByID for the order-item brand snapshot
```

---

## Bundle 5 — HTTP handlers + main.go wire + OpenAPI regen + smoke test

**Tasks:** the three new endpoints; mount in main.go; regen OpenAPI; extend the smoke test to exercise the full payment flow with a stubbed Paystack server + a signed webhook.

### Files
- Create: `backend/internal/orders/handler.go` — POST /checkout/init, GET /checkout/verify/{reference}
- Create: `backend/internal/orders/webhook.go` — POST /webhooks/paystack
- Create: `backend/internal/orders/handler_test.go`
- Modify: `backend/internal/app/app.go` — wire Paystack client + ResendSender + AllowlistSender + Orders service
- Modify: `backend/cmd/api/main.go` — mount handlers
- Modify: `backend/cmd/api/main_test.go` — end-to-end smoke
- Regenerate: `backend/docs/swagger.{json,yaml,go}`

### Handler shapes

```go
type Handlers struct {
    Svc           *Service
    PaystackSecret string  // for webhook signature verification
    Log           *zap.Logger
}

func NewHandlers(svc *Service, paystackSecret string, log *zap.Logger) *Handlers
func (h *Handlers) MountAuthGated(r chi.Router)   // POST /checkout/init, GET /checkout/verify/{reference}
func (h *Handlers) MountPublic(r chi.Router)      // POST /webhooks/paystack
```

`POST /checkout/init` body:
```json
{
  "shipping_address": {
    "line1": "...",
    "line2": "...",
    "city": "...",
    "region": "...",
    "phone": "...",
    "label": "Home"
  },
  "shipping_method": "standard"
}
```
Auth required. 200 with `{order_id, reference, authorization_url, total_ghs_minor}`. 400 on invalid address/empty cart. 503 on Paystack not configured or upstream error.

`GET /checkout/verify/{reference}`:
- Auth required. 200 with `{status}` where status ∈ {`pending`,`paid`,...}. 404 if reference doesn't belong to ANY order (or doesn't belong to the authenticated user — IDOR concern; reject 404 to avoid enumeration). Confirms user ownership by checking the order's user_id matches the session's userID.

`POST /webhooks/paystack`:
- Public. Reads raw body BEFORE JSON-decoding (signature verification needs the unmodified bytes). Reads `x-paystack-signature` header. Calls `paystack.VerifyWebhookSignature(secret, rawBody, headerValue)`. Invalid → 401 (Paystack will retry). Valid → parse the body, extract `data.reference` and `data.id` (transaction id). Call `svc.MarkPaid(ctx, reference, transactionID)`. 200 on success OR ErrNotFound. 500 on other errors (so Paystack retries).

### Webhook body shape

```go
type paystackWebhookEvent struct {
    Event string `json:"event"`  // e.g., "charge.success"
    Data  struct {
        Reference string `json:"reference"`
        ID        int64  `json:"id"`        // numeric transaction id
        Status    string `json:"status"`    // "success", "failed"
        Amount    int64  `json:"amount"`
    } `json:"data"`
}
```

Process ONLY `event == "charge.success" && data.status == "success"` events. All other events: log + 200 (acknowledge so Paystack doesn't retry, but no state change).

### main.go mount

Inside the existing `/api/v1` Route:
```go
ordersHandlers := orders.NewHandlers(a.Orders, cfg.PaystackSecretKey, a.Logger)
ordersHandlers.MountPublic(api)  // /webhooks/paystack — public, no session middleware

// Inside the existing RequireSession Group:
ordersHandlers.MountAuthGated(r)
```

The webhook must NOT live inside the RequireSession group — Paystack won't send a session cookie.

### app.New wiring

```go
// Build the email Sender chain.
renderer, err := email.NewRenderer()
if err != nil { ... }

var inner email.Sender = email.LogSender{Log: logger}
if resendSender, rerr := email.NewResendSender(cfg.ResendAPIKey, cfg.ResendFromEmail, renderer, logger); rerr == nil {
    inner = resendSender
}
mailSender := email.NewAllowlistSender(inner, cfg.EmailAllowlist, logger)

// Paystack client.
paystackClient := paystack.NewClient(cfg.PaystackBaseURL, cfg.PaystackSecretKey)

// Orders service.
ordersSvc := orders.NewService(
    orders.NewRepository(pool),
    cartSvc, catalogRepo, ship, paystackClient, mailSender, pool, logger,
    cfg.PaystackCallbackURL,
)
```

Update `Application` struct fields: `Email email.Sender` (already exists), `Orders *orders.Service` (new), `Paystack *paystack.Client` (new — exposed for tests/health).

### Smoke test extension

Extend `TestServerBootsAndHealthzReturnsOK` with:

1. Stand up a stub Paystack server via `httptest.NewServer(...)` BEFORE launching the binary. The stub handles:
   - `POST /transaction/initialize` → returns `{status:true, data:{authorization_url:"https://stub/abc", access_code:"AC", reference:<received>}}`.
   - `GET /transaction/verify/:reference` → returns `{status:true, data:{reference:<param>, status:"success", id:1234567, amount:<...>}}`.
2. Pass the stub URL to the binary via `PAYSTACK_BASE_URL=...` env var. Also set `PAYSTACK_SECRET_KEY=sk_test_smoke`.
3. After the existing cart flow (guest → signup → merge → /cart shows merged item), continue:
   - Compute the cart's current total via GET /cart.
   - POST /api/v1/checkout/init with a stub shipping_address. Capture `reference` and `authorization_url`.
   - Assert authorization_url is the stub server's URL.
   - Simulate a webhook: build the canonical Paystack webhook body referencing the captured reference, compute the HMAC-SHA512 signature with the configured secret, POST it to `/api/v1/webhooks/paystack` with the `x-paystack-signature` header. Assert 200.
   - Query the side pool directly to assert `orders.status = 'paid'` for that reference and `cart_items` for the user is empty.

### OpenAPI regen + drift-check

```bash
cd backend
PATH="$(go env GOPATH)/bin:$PATH" swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency
```

Verify the three new routes appear exactly once:
```bash
for path in /checkout/init '/checkout/verify/{reference}' /webhooks/paystack; do
    echo -n "$path: "; grep -c "\"$path\"" docs/swagger.json
done
```

`make drift-check` and `make test` both exit 0.

### Commit message

```
feat(orders): checkout/init + checkout/verify + paystack webhook, mount, smoke

- internal/orders/handler.go: POST /checkout/init (auth) and GET
  /checkout/verify/{reference} (auth, user-scoped) returning the order status
- internal/orders/webhook.go: POST /webhooks/paystack (public) reads raw
  body, HMAC-verifies, decodes, calls MarkPaid; 401 on invalid sig
- app + main: build email Sender chain (LogSender → ResendSender →
  AllowlistSender), construct Paystack client + Orders service, mount
  public webhook outside the RequireSession group
- cmd/api/main_test.go: extends smoke with stubbed Paystack server,
  /checkout/init, signed webhook, DB-asserts status=paid + empty cart
- docs: regenerated openapi with the 3 new routes
```

---

## Verification — end of Plan 5

- `make test` exits 0.
- `make drift-check` exits 0.
- `make up && make dev` boots the server. With a real `PAYSTACK_SECRET_KEY` from a Paystack test account, the following flow works in a browser:
  - Sign up, add an item to cart, POST /checkout/init with a valid address.
  - Get redirected to Paystack's test card UI, complete a payment with the test card.
  - Get redirected back to the frontend's `/checkout/return?reference=...`.
  - Backend's webhook handler receives Paystack's webhook, flips the order to paid, empties the cart, fires the confirmation email (logged via slog if not in EMAIL_ALLOWLIST, otherwise delivered via Resend).
- 5 new commits on top of `849d317`, all prefixed `feat(`.
- No `fmt.Sprintf` constructing SQL.
- No raw `PAYSTACK_SECRET_KEY` logged anywhere outside test files.
- The full RBAC matrix from Plan 3 still passes.

## Self-Review Notes

- **Spec coverage:** §4.1 (orders + order_items), §5.2 (checkout endpoints + webhook), §6.2 (full Paystack flow + idempotent convergence on MarkPaid), §6.3 (Resend + allowlist + order confirmation template), §14 (HMAC-SHA512 webhook verification, constant-time compare, no enumeration on verify).
- **Addresses deferred:** Plan 5 inlines the address into `orders.shipping_address jsonb`. Plan 6 adds the `addresses` table + saved-addresses CRUD + extends `/checkout/init` to also accept `{address_id}`.
- **Status state machine deferred:** Plan 5 implements only `pending → paid`. The other transitions (`fulfilled`, `shipped`, `delivered`, `cancelled`, `refunded`) are Plan 7's `PATCH /admin/orders/{id}/status` with validation of allowed transitions.
- **Email queue deferred:** confirmation email is sent inline inside the MarkPaid handler. If Resend is slow, the handler is slow. Spec §6.3 documents this; queue is post-v1.
- **No frontend in this plan:** Plan 12 wires the checkout flow into the React frontend (TanStack Query polling the verify endpoint after the Paystack redirect).
- **Tests use a stub Paystack server** via `httptest.NewServer` — neither the smoke test nor the unit tests hit `https://api.paystack.co`. The only place a real Paystack key is used is local dev / production.
- **Carryover deferrals from Plan 4 (per the whole-branch review):** the `deleteItem` discards-View and `postItem` double-buildView issues are NOT addressed here. They may naturally surface if `checkout/init` calls `cartSvc.GetCart` and feels the same wasted-read sting; if so, fold into Bundle 4. Otherwise defer further.
