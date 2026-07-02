# Backend Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the backend audit gaps: 500 responses that silently swallow errors, no HTTP access logs, route wiring duplicated outside the DI container, a misleadingly named admin mount, a silent email-sender fallback, and runtime config living under `seed/`.

**Architecture:** Add two small `httpx` building blocks (an error-logging 500 helper and an access-log middleware), then move router construction from `cmd/api/main.go` into a new `internal/api` package (a new package avoids the import cycle: `internal/health` imports `internal/app`, so `app` itself cannot own the router). Everything else is targeted refactors with tests.

**Tech Stack:** Go 1.25, chi v5, zap, pgx/v5 + sqlc, testcontainers (existing test infra), goose.

## Context (read this first)

Audit facts an implementer needs:

1. Handlers write 500s without logging the underlying error — e.g. `internal/catalog/handler.go:62-65` returns "failed to list categories" and drops `err`. The plumbing to fix this already exists: `internal/logging.From(ctx, fallback)` returns the request-scoped zap logger with `request_id` attached.
2. `httpx.RequestLogger` (`internal/httpx/middleware.go:43`) only *injects* a logger into context; nothing logs method/path/status/duration. There are zero access logs.
3. `cmd/api/main.go:63-101` builds handlers and re-creates repositories that `app.New` already built (`catalog.NewRepository` twice; `orders.NewRepository` and `me.NewProfileService` built inline in the auth-gated group).
4. `admin.Handlers.MountPublic` (`internal/admin/handler.go:32`) mounts routes that require session + admin role — the name lies. The per-handler `auth.MustBeAdmin` re-checks are an intentional belt-and-suspenders convention (see `internal/auth/rbac_test.go` comments) — keep them, codify the convention in a doc comment.
5. `internal/app/app.go:55-58` silently falls back to the log-only email sender when `NewResendSender` errors. A typo'd `RESEND_API_KEY` in production means no emails and no signal.
6. `backend/seed/config/shipping_config.json` is runtime config, not seed data; the default path lives in `internal/config/config.go:15`. Tests are unaffected by moving it — `testsupport.WriteShippingConfig` writes its own temp file.
7. `internal/admin` has **no test files** at all.
8. House test style: package-external tests (`package xxx_test`), table-driven, `httptest`, testcontainers via `internal/testsupport` for DB-backed tests, zap's `observer` for log assertions.

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."`.
- Run Go commands from `ruecosmetics/backend/`; `make` targets from `ruecosmetics/`.
- Go tests: `go test ./internal/<pkg>/ -run TestName -timeout=600s`; DB-backed tests need Docker running.
- `make drift-check` must pass after any change that touches swag annotations or `queries/` (this plan touches neither, but verify before the final commit).
- Public error messages must stay generic; detail goes to logs only. Error envelope shape (`httpx.ErrorEnvelope`) must not change — the frontend parses it.

---

### Task 1: `httpx.WriteInternal` — 500s that log the real error

**Files:**
- Modify: `backend/internal/httpx/error.go`
- Test: `backend/internal/httpx/error_test.go` (append)
- Modify: `backend/internal/catalog/handler.go` (all `StatusInternalServerError` sites), `backend/internal/shipping/handler.go`, `backend/internal/admin/handler.go` (same pattern)

**Interfaces:**
- Consumes: `logging.From(ctx, fallback)` (`internal/logging/logging.go:24`).
- Produces: `func WriteInternal(w http.ResponseWriter, r *http.Request, fallback *zap.Logger, publicMsg string, err error)` — logs `err` at Error level via the request-scoped logger, then writes the standard 500 envelope. Later tasks and future handlers use this instead of bare `WriteError(..., 500, ...)`.

- [ ] **Step 1: Write the failing test** (append to `internal/httpx/error_test.go`; match its existing style)

```go
func TestWriteInternalLogsAndWritesEnvelope(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	base := zap.New(core)

	req := httptest.NewRequest(http.MethodGet, "/things", nil)
	rec := httptest.NewRecorder()

	httpx.WriteInternal(rec, req, base, "failed to list things", errors.New("pg: connection refused"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var env httpx.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error.Code != httpx.CodeInternal || env.Error.Message != "failed to list things" {
		t.Fatalf("envelope = %+v", env.Error)
	}
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	if entries[0].Message != "failed to list things" {
		t.Fatalf("log message = %q", entries[0].Message)
	}
	found := false
	for _, f := range entries[0].Context {
		if f.Key == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected zap error field on log entry")
	}
}
```

Imports needed: `errors`, `encoding/json`, `net/http`, `net/http/httptest`, `testing`, `go.uber.org/zap`, `go.uber.org/zap/zaptest/observer`, and the `httpx` package.

- [ ] **Step 2: Run it to make sure it fails**

Run: `go test ./internal/httpx/ -run TestWriteInternalLogsAndWritesEnvelope -timeout=600s`
Expected: FAIL — `undefined: httpx.WriteInternal`.

- [ ] **Step 3: Implement** (append to `internal/httpx/error.go`)

```go
// WriteInternal logs err (with the request-scoped logger, so request_id is
// attached) and writes the standard 500 envelope with a generic public
// message. Use this for every unexpected-error branch instead of a bare
// WriteError, so no 500 is ever silent.
func WriteInternal(w http.ResponseWriter, r *http.Request, fallback *zap.Logger, publicMsg string, err error) {
	logging.From(r.Context(), fallback).Error(publicMsg,
		zap.Error(err),
		zap.String("path", r.URL.Path),
	)
	WriteError(w, http.StatusInternalServerError, CodeInternal, publicMsg, nil)
}
```

Add imports `go.uber.org/zap` and `github.com/oti-adjei/ruecosmetics/internal/logging` to `error.go`. (No import cycle: `logging` imports only `context` and `zap`.)

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/httpx/ -timeout=600s`
Expected: PASS (all httpx tests, not just the new one).

- [ ] **Step 5: Adopt it at every silent-500 site**

Enumerate them:

```bash
grep -rn 'StatusInternalServerError' internal/ --include='*.go' | grep -v _test.go
```

For each hit in `catalog`, `shipping`, and `admin` handlers, apply this transformation (example from `catalog/handler.go:61-66`):

```go
// before
cats, err := h.repo.ListCategories(r.Context())
if err != nil {
	httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list categories", nil)
	return
}

// after
cats, err := h.repo.ListCategories(r.Context())
if err != nil {
	httpx.WriteInternal(w, r, h.log, "failed to list categories", err)
	return
}
```

`catalog.Handlers` and `shipping.Handlers` don't currently hold a logger — add a `log *zap.Logger` field and a logger parameter to their `NewHandlers` constructors, and update the call sites in `cmd/api/main.go` (`catalog.NewHandlers(catalog.NewRepository(a.Pool), a.Logger)` etc.). `admin.Handlers` gets the same field. Handlers that already log their errors before writing 500 (e.g. `orders`, via `logging.From`) may be left as-is or converted — convert only if the conversion is a pure simplification.

- [ ] **Step 6: Verify the whole tree still passes**

Run: `go build ./... && go test ./... -timeout=600s`
Expected: PASS. (Full suite needs Docker.)

- [ ] **Step 7: Commit**

```bash
git add backend/internal backend/cmd
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(httpx): WriteInternal logs the underlying error on every 500"
```

---

### Task 2: Move router construction into `internal/api`

**Files:**
- Create: `backend/internal/api/routes.go`
- Test: `backend/internal/api/routes_test.go`
- Modify: `backend/cmd/api/main.go` (shrinks to config + app + serve)
- Modify: `backend/internal/app/app.go` (expose the catalog repository instead of letting main rebuild it)

**Interfaces:**
- Consumes: `*app.Application` and its fields (`Auth`, `Cart`, `Orders`, `Addresses`, `Admin`, `Shipping`, `Pool`, `Logger`, `Config`).
- Produces: `func New(a *app.Application) http.Handler` in package `api` — the complete production router including middleware and `/healthz`. `main.go` and integration tests both use it, so tests exercise the real route table.

- [ ] **Step 1: Expose the catalog repository on Application**

In `internal/app/app.go`, add a field `Catalog *catalog.Repository` to the `Application` struct and set `Catalog: catalogRepo` in the returned struct (the local `catalogRepo` already exists at line 61). This removes main.go's duplicate `catalog.NewRepository(a.Pool)`.

- [ ] **Step 2: Write the failing test**

`backend/internal/api/routes_test.go`:

```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/api"
	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/config"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// TestRoutesMountsExpectedPaths boots the real router against a test DB and
// asserts representative routes exist (non-404) with correct gating.
func TestRoutesMountsExpectedPaths(t *testing.T) {
	pool := testsupport.NewMigratedPool(t) // use the existing helper name from internal/testsupport/postgres.go — check and match it
	cfg := &config.Config{
		Env:                "test",
		DatabaseURL:        "unused-pool-injected",
		LogLevel:           "error",
		ShippingConfigPath: testsupport.WriteShippingConfig(t, 2000, 30000),
		SessionCookieName:  "rue_session",
	}
	a := app.NewWithPool(t.Context(), cfg, pool) // see Step 4
	t.Cleanup(a.Close)

	h := api.New(a)

	cases := []struct {
		method, path string
		wantStatus   int
	}{
		{http.MethodGet, "/healthz", http.StatusOK},
		{http.MethodGet, "/api/v1/products", http.StatusOK},
		{http.MethodGet, "/api/v1/categories", http.StatusOK},
		{http.MethodGet, "/api/v1/cart", http.StatusOK},
		{http.MethodGet, "/api/v1/me", http.StatusUnauthorized},          // auth-gated
		{http.MethodGet, "/api/v1/admin/orders", http.StatusUnauthorized}, // admin-gated
		{http.MethodPost, "/api/v1/webhooks/paystack", http.StatusBadRequest}, // public webhook, bad signature ≠ 404
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.wantStatus {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, rec.Code, tc.wantStatus)
		}
	}
}
```

Before running: open `internal/testsupport/postgres.go` and use its actual helper names (migrated-pool setup) — adjust the two helper calls to match; the webhook expected status may be 401/400 depending on signature handling in `orders.paystackWebhook` — set the expectation to whatever the handler returns for a missing signature (read `internal/orders/handler.go` / `webhook.go`).

- [ ] **Step 3: Run it to make sure it fails**

Run: `go test ./internal/api/ -timeout=600s`
Expected: FAIL — package `api` does not exist.

- [ ] **Step 4: Implement `internal/api/routes.go`**

Move the entire router block from `cmd/api/main.go:53-103` verbatim, with these changes: use `a.Catalog` instead of `catalog.NewRepository(a.Pool)`; read `cfg := a.Config` at the top; keep the middleware order exactly (`Recovery`, `RequestID`, `RequestLogger`, `CORS`).

```go
// Package api assembles the full production HTTP router from an Application.
// It lives outside package app because internal/health imports app.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oti-adjei/ruecosmetics/internal/addresses"
	"github.com/oti-adjei/ruecosmetics/internal/admin"
	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/cart"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/me"
	"github.com/oti-adjei/ruecosmetics/internal/orders"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

// New builds the production router. main.go and integration tests share it.
func New(a *app.Application) http.Handler {
	cfg := a.Config
	r := chi.NewRouter()
	r.Use(httpx.Recovery(a.Logger))
	r.Use(httpx.RequestID)
	r.Use(httpx.RequestLogger(a.Logger))
	r.Use(httpx.CORS(cfg.CORSOrigins))

	r.Get("/healthz", health.Handler(a))

	catalogHandlers := catalog.NewHandlers(a.Catalog, a.Logger)
	shippingHandlers := shipping.NewHandlers(a.Shipping, a.Logger)
	r.Route("/api/v1", func(apiRouter chi.Router) {
		catalogHandlers.Mount(apiRouter)
		shippingHandlers.Mount(apiRouter)

		secure := cfg.Env != "development"
		authHandlers := auth.NewHandlers(a.Auth, cfg.SessionCookieName, cfg.SessionCookieDomain, secure)
		authHandlers.GoogleClientID = cfg.GoogleClientID
		authHandlers.GoogleClientSecret = cfg.GoogleClientSecret
		authHandlers.GoogleRedirectURL = cfg.GoogleRedirectURL
		authHandlers.FrontendBaseURL = cfg.FrontendBaseURL
		authHandlers.Mount(apiRouter)

		cartHandlers := cart.NewHandlers(a.Cart, a.Auth, cfg.SessionCookieName, cfg.SessionCookieDomain, secure)
		cartHandlers.Mount(apiRouter)

		ordersHandlers := orders.NewHandlers(a.Orders, cfg.PaystackSecretKey, a.Logger)
		ordersHandlers.MountPublic(apiRouter)

		adminHandlers := admin.NewHandlers(a.Admin, authHandlers)
		adminHandlers.MountPublic(apiRouter) // renamed to Mount in Task 4

		apiRouter.Group(func(gated chi.Router) {
			gated.Use(authHandlers.RequireSession)
			ordersRepo := orders.NewRepository(a.Pool)
			profileSvc := me.NewProfileService(auth.NewRepository(a.Pool), a.Logger)
			meHandlers := me.NewHandlers(ordersRepo, profileSvc, a.Logger)
			authHandlers.MountAuthGated(gated)
			cartHandlers.MountAuthGated(gated)
			ordersHandlers.MountAuthGated(gated)

			addressesHandlers := addresses.NewHandlers(a.Addresses, a.Logger)
			gated.Route("/me", func(meRouter chi.Router) {
				meHandlers.MountRoutes(meRouter)
				addressesHandlers.Mount(meRouter)
			})
		})
	})

	return r
}
```

(Constructor signatures with the added logger params come from Task 1 — if Task 1 hasn't run, keep the current signatures; the two tasks are order-independent apart from that.)

If the test needs an `app.NewWithPool` variant (constructing `Application` without dialing `cfg.DatabaseURL`): check how existing DB-backed handler tests build services — if `app.New` is never used in tests today, add to `internal/app/app.go` a small `NewWithPool(ctx, cfg, pool)` that runs the same wiring with an injected pool, and have `New` call it after `db.NewPool`. Keep the change minimal and mirror `New`'s body.

Then shrink `cmd/api/main.go`: delete the router block and the now-unused imports; replace with `handler := api.New(a)` and `srv := &http.Server{Addr: ..., Handler: handler, ReadHeaderTimeout: 5 * time.Second}`.

- [ ] **Step 5: Run the tests**

Run: `go test ./internal/api/ ./cmd/... -timeout=600s`
Expected: PASS, including the existing `cmd/api/main_test.go` boot test.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/api backend/internal/app backend/cmd/api
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "refactor(api): assemble router in internal/api; main.go is config+serve only"
```

---

### Task 3: Access-log middleware

**Files:**
- Modify: `backend/internal/httpx/middleware.go`
- Test: `backend/internal/httpx/middleware_test.go` (append)
- Modify: `backend/internal/api/routes.go` (wire it)

**Interfaces:**
- Consumes: `GetRequestID(ctx)`, chi's `middleware.NewWrapResponseWriter`.
- Produces: `func AccessLog(base *zap.Logger) func(http.Handler) http.Handler` — one Info line per request with `method`, `path`, `status`, `duration_ms`, `bytes`, `request_id`.

- [ ] **Step 1: Write the failing test** (append to `internal/httpx/middleware_test.go`)

```go
func TestAccessLogEmitsOneEntryPerRequest(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	base := zap.New(core)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short and stout"))
	})
	h := httpx.RequestID(httpx.AccessLog(base)(inner))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	got := entries[0].ContextMap()
	if got["method"] != "GET" || got["path"] != "/api/v1/products" {
		t.Fatalf("context = %+v", got)
	}
	if got["status"] != int64(http.StatusTeapot) {
		t.Fatalf("status = %v, want 418", got["status"])
	}
	if _, ok := got["request_id"]; !ok {
		t.Fatal("missing request_id")
	}
}
```

- [ ] **Step 2: Run it to make sure it fails**

Run: `go test ./internal/httpx/ -run TestAccessLogEmitsOneEntryPerRequest -timeout=600s`
Expected: FAIL — `undefined: httpx.AccessLog`.

- [ ] **Step 3: Implement** (append to `internal/httpx/middleware.go`)

```go
// AccessLog emits one structured log line per request after it completes.
// Install AFTER RequestID so the request_id field is populated. Query strings
// are deliberately not logged (they can carry tokens).
func AccessLog(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			base.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.String("request_id", GetRequestID(r.Context())),
			)
		})
	}
}
```

Add imports: `time`, `github.com/go-chi/chi/v5/middleware`.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/httpx/ -timeout=600s`
Expected: PASS.

- [ ] **Step 5: Wire it into the router**

In `internal/api/routes.go`, add `r.Use(httpx.AccessLog(a.Logger))` immediately after `r.Use(httpx.RequestID)` (and before `RequestLogger`). Run `go test ./internal/api/ -timeout=600s` — the route test still passes.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/httpx backend/internal/api
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(httpx): structured access log with status, duration, and request_id"
```

---

### Task 4: Admin mount rename + first admin RBAC tests

**Files:**
- Modify: `backend/internal/admin/handler.go:30-63` (rename), `backend/internal/api/routes.go` (call site)
- Test: `backend/internal/admin/handler_rbac_test.go` (new — the package's first test file)

**Interfaces:**
- Consumes: the RBAC test scaffolding pattern from `internal/auth/rbac_test.go` (router builder + testsupport session/cookie helpers — read that file and reuse its helpers verbatim where exported).
- Produces: `admin.Handlers.Mount(r chi.Router)` (replacing `MountPublic`); no route paths change.

- [ ] **Step 1: Write the failing test**

`backend/internal/admin/handler_rbac_test.go` — follow `internal/auth/rbac_test.go` for session setup (it creates users with roles and issues session cookies via testsupport; copy its arrangement):

```go
package admin_test

// The admin package's routes all require session + admin role. This matrix
// pins that gating for representative endpoints so a future mount refactor
// can't silently expose them.

func TestAdminRoutesRequireAdminRole(t *testing.T) {
	// Arrange: build a router exactly like internal/auth/rbac_test.go does,
	// then mount admin handlers on it:
	//   adminHandlers := admin.NewHandlers(nil, authHandlers) // Svc unused for 401/403 paths
	//   adminHandlers.Mount(r)   // <- new name; compile fails until Step 3
	//
	// Matrix (per endpoint: /admin/dashboard, /admin/orders, /admin/analytics/stats):
	//   anonymous          -> 401
	//   customer session   -> 403
	//
	// Use the same helpers rbac_test.go uses for signup/login and cookie
	// extraction (testsupport.NewMigratedPool / cookie helpers).
}
```

Write the real body by transplanting the arrange/act/assert code from `rbac_test.go`'s customer/anonymous cases, swapping `/admin/ping` for the three admin paths above. A 200-path admin case would need a seeded admin user and a live `admin.Service` — add it only if `rbac_test.go` already has an admin-session helper you can reuse; otherwise the 401/403 matrix is this task's scope.

- [ ] **Step 2: Run it to make sure it fails**

Run: `go test ./internal/admin/ -timeout=600s`
Expected: FAIL — `adminHandlers.Mount` undefined (only `MountPublic` exists).

- [ ] **Step 3: Rename and document the convention**

In `internal/admin/handler.go` rename `MountPublic` → `Mount` and replace its doc comment:

```go
// Mount mounts all admin routes under the given router. Every route is
// gated by RequireSession + RequireRole("admin") here, and each handler
// additionally calls auth.MustBeAdmin as belt-and-suspenders (the same
// convention rbac_test.go documents) so a future mis-mount cannot expose
// an admin endpoint.
func (h *Handlers) Mount(r chi.Router) {
```

Update the call site in `internal/api/routes.go` (`adminHandlers.Mount(apiRouter)`) and its stale `// renamed in Task 4` comment. Leave the per-handler `MustBeAdmin` calls in place.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/admin/ ./internal/api/ -timeout=600s`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/admin backend/internal/api
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "test(admin): RBAC matrix for admin routes; rename MountPublic to Mount"
```

---

### Task 5: Explicit email-sender selection

**Files:**
- Create: `backend/internal/email/select.go`
- Test: `backend/internal/email/select_test.go`
- Modify: `backend/internal/app/app.go:48-59` (use the new selector)

**Interfaces:**
- Consumes: `email.LogSender`, `email.NewResendSender(apiKey, from, renderer, logger)`, `email.NewAllowlistSender(inner, allowlist, logger)`, `email.NewRenderer()` — all existing.
- Produces: `func Select(env, resendAPIKey, resendFrom string, allowlist []string, renderer *Renderer, log *zap.Logger) (Sender, error)` — check `NewRenderer`'s actual return type in `templates.go` and use it for the `renderer` parameter. Behavior: no API key + `env == "production"` → error; no API key otherwise → log-only sender (logged at Warn); API key present → Resend sender (logged at Info); result always wrapped in AllowlistSender.

- [ ] **Step 1: Write the failing test**

`backend/internal/email/select_test.go` (internal package test — `package email` — so sender types can be asserted directly):

```go
package email

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func newTestRenderer(t *testing.T) *Renderer { // match NewRenderer's real return type
	t.Helper()
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	return r
}

func TestSelectWithoutKeyInDevelopmentUsesLogSender(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	s, err := Select("development", "", "noreply@rue.example.com", nil, newTestRenderer(t), zap.New(core))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if s == nil {
		t.Fatal("nil sender")
	}
	if logs.Len() != 1 || !strings.Contains(logs.All()[0].Message, "log-only") {
		t.Fatalf("expected one log-only warning, got %v", logs.All())
	}
}

func TestSelectWithoutKeyInProductionFails(t *testing.T) {
	_, err := Select("production", "", "noreply@rue.example.com", nil, newTestRenderer(t), zap.NewNop())
	if err == nil {
		t.Fatal("expected error when production has no email provider")
	}
}

func TestSelectWithKeyUsesResend(t *testing.T) {
	s, err := Select("production", "re_test_key", "noreply@rue.example.com", nil, newTestRenderer(t), zap.NewNop())
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	allow, ok := s.(*AllowlistSender) // match the real concrete type in allowlist.go
	if !ok {
		t.Fatalf("outermost sender = %T, want *AllowlistSender", s)
	}
	if _, ok := allow.Inner.(*ResendSender); !ok { // match the real field/type names
		t.Fatalf("inner sender = %T, want *ResendSender", allow.Inner)
	}
}
```

Before running: open `allowlist.go`, `resend.go`, `sender.go`, `templates.go` and align the concrete type names, exported/unexported fields, and `NewRenderer` return type. If `AllowlistSender`'s inner sender is unexported, either export nothing and assert via behavior (send to a disallowed address and inspect logs) or add a small `Unwrap() Sender` method — prefer the behavioral assertion.

- [ ] **Step 2: Run to make sure they fail**

Run: `go test ./internal/email/ -run TestSelect -timeout=600s`
Expected: FAIL — `undefined: Select`.

- [ ] **Step 3: Implement `internal/email/select.go`**

```go
package email

import (
	"errors"

	"go.uber.org/zap"
)

// Select picks the outgoing email implementation from config, loudly.
// Production without a configured provider is a startup error, not a
// silent fallback to log-only delivery.
func Select(env, resendAPIKey, resendFrom string, allowlist []string, renderer *Renderer, log *zap.Logger) (Sender, error) {
	var inner Sender
	switch {
	case resendAPIKey != "":
		rs, err := NewResendSender(resendAPIKey, resendFrom, renderer, log)
		if err != nil {
			return nil, err
		}
		log.Info("email: using Resend sender", zap.String("from", resendFrom))
		inner = rs
	case env == "production":
		return nil, errors.New("email: RESEND_API_KEY is required in production")
	default:
		log.Warn("email: no provider configured, using log-only sender")
		inner = LogSender{Log: log}
	}
	return NewAllowlistSender(inner, allowlist, log), nil
}
```

(Adjust `*Renderer` and constructor signatures to the real ones found in Step 1.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/email/ -timeout=600s`
Expected: PASS.

- [ ] **Step 5: Use it in `app.New`**

Replace `internal/app/app.go` lines 48–59 with:

```go
renderer, rerr := email.NewRenderer()
if rerr != nil {
	pool.Close()
	return nil, rerr
}
mailSender, merr := email.Select(cfg.Env, cfg.ResendAPIKey, cfg.ResendFromEmail, cfg.EmailAllowlist, renderer, logger)
if merr != nil {
	pool.Close()
	return nil, merr
}
```

Run: `go test ./internal/app/ ./internal/api/ -timeout=600s`
Expected: PASS (test configs use `Env: "test"`/`development`, so no key is fine).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/email backend/internal/app
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(email): explicit sender selection; production requires a real provider"
```

---

### Task 6: Relocate runtime shipping config out of `seed/`

**Files:**
- Move: `backend/seed/config/shipping_config.json` → `backend/config/shipping_config.json`
- Modify: `backend/internal/config/config.go:15` (default path), `backend/.env.example:14-15`, `README.md` (if it references the old path)

- [ ] **Step 1: Move the file**

```bash
cd backend && mkdir -p config && git mv seed/config/shipping_config.json config/shipping_config.json && rmdir seed/config seed 2>/dev/null || true
```

(If `seed/` has other contents, leave the directory; only `config/` moves.)

- [ ] **Step 2: Update the default path**

`internal/config/config.go:15`:

```go
ShippingConfigPath   string   `envconfig:"SHIPPING_CONFIG_PATH" default:"config/shipping_config.json"`
```

And `.env.example` line 15:

```
# SHIPPING_CONFIG_PATH=config/shipping_config.json
```

Check for other references: `grep -rn 'seed/config' . --include='*.go' --include='*.md' --include='Makefile' --include='*.example'` — update every hit (the design spec under `docs/` may mention it; update the README, leave historical specs alone).

- [ ] **Step 3: Verify**

Run: `go test ./internal/config/ ./internal/shipping/ ./internal/app/ -timeout=600s` — PASS (tests write their own temp config via `testsupport.WriteShippingConfig`, so only the default string changed). Then boot locally: `make dev` from repo root and `curl http://localhost:8080/healthz` → `{"status":"ok","db":"ok"}`.

- [ ] **Step 4: Commit**

```bash
git add -A backend/config backend/seed backend/internal/config backend/.env.example README.md
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "refactor(config): shipping config is runtime config, move out of seed/"
```

---

### Task 7: Final verification sweep

**Files:** none new.

- [ ] **Step 1: Full-tree gate**

Run from repo root:

```bash
make drift-check   # OpenAPI + sqlc still in sync (nothing in this plan should change them)
make test          # full backend suite with Docker
```
Expected: both pass.

- [ ] **Step 2: Grep for leftovers**

```bash
grep -rn 'MountPublic' backend/internal/admin backend/internal/api   # expect: no hits
grep -rn 'StatusInternalServerError' backend/internal --include='*.go' | grep -v _test.go | grep -v WriteInternal  # expect: only sites that log before writing (orders/health)
grep -rn 'seed/config' backend --include='*.go'                       # expect: no hits
```

- [ ] **Step 3: Commit any stragglers**

If the greps surfaced misses, fix and commit with the same identity flag:

```bash
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -am "chore: backend polish follow-ups from final sweep"
```
