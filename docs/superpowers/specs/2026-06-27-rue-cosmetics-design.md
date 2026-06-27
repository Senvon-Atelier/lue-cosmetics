# Rue Cosmetics — Case Study Design

**Date:** 2026-06-27
**Status:** Draft for review
**Goal:** Turn the existing Rue Cosmetics frontend mockup (in `casestud/Rue/`) into a fully functional e-commerce case study, deployed to the user's Hetzner server behind Caddy.

## 1. Purpose & Success Criteria

Rue Cosmetics is a Ghana-based cosmetics & wellness e-commerce concept. It was originally designed for potential clients who did not engage. It is being repurposed as a **public case study**: a working website visitors can sign up to, browse, add to cart, check out (in payment-gateway test mode), and explore as either a fresh customer or as one of several pre-seeded demo accounts.

**Success means:**

- A visitor can land on the homepage, browse the catalog, add items, register an account (or log in to a seeded account), check out via Paystack test mode, and see the resulting order in their account dashboard.
- An admin user can log in to the admin dashboard, see real orders/users/revenue from the database, and update an order's fulfillment status.
- The system is built and presented in a way that is impressive to a reviewer evaluating the user's portfolio: real database, real auth, real payments (in test mode), real email, type-safe end-to-end, RBAC enforced in depth.
- Deployed at `rue.example.com` (frontend) and `api.rue.example.com` (backend) on the user's existing Hetzner host behind their existing Caddy setup.

**Explicit non-goals (v1):**

- Real money / production-mode payments.
- Admin CRUD for the product catalog (catalog is seeded and code-managed).
- Customer-written product reviews (rating + review count are seeded fields only).
- Coupon / discount code system.
- Return/RMA workflow (the `refunded` and `cancelled` order statuses exist; the workflow does not).
- A blog/CMS UI for editing posts (blog content lives as JSON files in the repo).
- Multi-currency (everything is GHS).
- Metrics / tracing observability stack.

## 2. Stack Decisions

| Layer | Choice | Reason |
|---|---|---|
| Frontend framework | **Vite + React 18 + TanStack Router + TanStack Query** | Cleanest split with a separate Go backend (TanStack Start's server-functions value would be wasted); TanStack ecosystem still showcases modern React DX. |
| Frontend styling | **Tailwind CSS v4** | User preference. Existing mockup's palette + typography mapped into Tailwind theme tokens; components rebuilt with utility classes. |
| Frontend validation | **Zod** | Best ecosystem fit with BetterAuth + TanStack Form; bundle cost is acceptable for a case-study site. |
| Frontend package manager | **pnpm** | User preference. |
| Auth | **BetterAuth (Go) with admin/roles plugin** | Handles email/password + Google OAuth + session cookies + role separation in one library. |
| Auth methods | **Email/password + Google OAuth** | Demonstrates breadth without the email-sending dependency of magic links. |
| Backend language | **Go** | Stronger portfolio showcase than a single TS stack. |
| Web framework | **chi** | Idiomatic Go, minimal, stdlib-flavored. |
| DB access | **sqlc** | Typed Go from plain SQL — no ORM, no injection paths. |
| Migrations | **goose** | Simple CLI, supports both SQL and Go migrations. |
| Backend validation | **`go-playground/validator`** | Struct-tag-driven, standard in Go ecosystem. |
| Database | **Postgres 16** | Locally installed on the Hetzner host (not managed). |
| Payments | **Paystack** | Native GHS + Mobile Money support (MTN MoMo, Vodafone Cash). Test mode for the case study. |
| Email | **Resend** in allowlist mode (real outbound to allowlisted addresses only); **Mailpit** in local dev | Real-shaped flow without deliverability risk. |
| Repo structure | **Monorepo** (`frontend/` and `backend/` side by side) | Easiest mental model for reviewers. |
| Deploy | **Caddy reverse proxy → static frontend + Go binary (systemd)** | Uses the user's existing infrastructure. |
| CI | **GitHub Actions** | Standard. |

## 3. Architecture

```
                 ┌──────────────────────────────────────────┐
                 │                Caddy (Hetzner)           │
                 │  rue.example.com → frontend (static)     │
                 │  api.rue.example.com → backend :8080     │
                 └──────────────────────────────────────────┘
                          │                       │
                ┌─────────▼─────────┐    ┌────────▼─────────┐
                │ React SPA          │    │ Go API (chi)     │
                │ Vite build         │    │ BetterAuth-Go    │
                │ TanStack Router/Q  │    │ sqlc + goose     │
                │ Zod schemas        │    │ Resend, Paystack │
                └────────────────────┘    └────────┬─────────┘
                                                   │
                                          ┌────────▼─────────┐
                                          │ Postgres 16      │
                                          │ (Hetzner, local) │
                                          └──────────────────┘
```

**Folder layout:**

```
casestud/ruecosmetics/
├── README.md
├── docker-compose.yml          # postgres + mailpit (dev only)
├── docs/superpowers/specs/...  # this design + future plans
├── .github/workflows/          # ci.yml
├── Caddyfile.example           # production reverse-proxy config
├── deploy/                     # systemd unit, Makefile deploy target
├── backend/                    # Go module
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── auth/               # BetterAuth wiring + middleware
│   │   ├── catalog/            # products, categories, brands handlers
│   │   ├── cart/               # cart + merge logic
│   │   ├── orders/             # checkout init, order lifecycle
│   │   ├── admin/              # admin handlers + RBAC middleware
│   │   ├── payments/           # Paystack client + webhook verifier
│   │   ├── email/              # Resend client + Mailpit dev provider
│   │   ├── content/            # blog.json, testimonials.json, legal/*.md, shipping_config.json
│   │   └── db/                 # sqlc-generated code
│   ├── migrations/             # goose .sql files
│   ├── queries/                # sqlc input .sql files
│   ├── sqlc.yaml
│   ├── seed/                   # seed program: ports data.js → DB
│   └── go.mod
└── frontend/                   # pnpm workspace root
    ├── package.json
    ├── pnpm-workspace.yaml
    ├── vite.config.ts
    ├── tsconfig.json
    ├── public/products/        # stock photos (B); tone CSS fallback if needed
    ├── tailwind.config.ts      # palette tokens, font families, animations
    └── src/
        ├── routes/             # TanStack Router file-based routes
        ├── components/         # rebuilt from existing Rue/src/*.jsx as TS + Tailwind
        ├── lib/
        │   ├── api/            # client, Zod schemas, openapi-types
        │   ├── auth/           # BetterAuth React client
        │   └── format/         # GHS formatting, dates
        └── styles/
            └── globals.css     # @tailwind directives + a small layer for keyframes / toast
```

**Local dev:**
- `docker compose up` → Postgres 16 on `:5432`, Mailpit on `:8025`.
- Backend: `air` (hot reload) on `:8080`, reads `.env`.
- Frontend: `pnpm dev` on `:5173`, Vite proxies `/api/*` → `:8080`.

**Production:**
- Single Hetzner host, single Postgres instance.
- Backend deployed as a single Go binary + systemd unit (`rue-api.service`), reads `/etc/rue/api.env`.
- Frontend deployed as static `dist/` served by Caddy directly.
- Two Caddy sites: `rue.example.com` (static) and `api.rue.example.com` (reverse proxy to `127.0.0.1:8080`).
- Both subdomains share an eTLD+1 so the BetterAuth session cookie is valid across them.

## 4. Data Model

UUID primary keys via `gen_random_uuid()` (pgcrypto). All monetary values stored as `BIGINT` in **pesewas** (1 GHS = 100 pesewas) to avoid float drift.

**Identity & auth** (managed by BetterAuth schemas + admin plugin)

- `users` — `id, email, name, image, email_verified, created_at, updated_at`
- `accounts` — OAuth provider links (Google) + credential rows for email/password
- `sessions` — `id, user_id, token, expires_at, ip, user_agent`
- `verification` — magic tokens for email verify + password reset
- `user_roles` — `user_id, role` where role ∈ {`customer`, `admin`}

**Catalog** (seeded from the existing `Rue/src/data.js`; read-only at runtime in v1)

- `categories` — `id, slug, label, sort_order`
- `brands` — `id, name, slug`
- `products` — `id, slug, name, brand_id, category_id, price_ghs_minor, was_price_ghs_minor (nullable), tone, size, rating, review_count, tags text[], image_path, created_at`

**Commerce**

- `carts` — `id, user_id (nullable), guest_token (nullable), created_at, updated_at`. Constraint: exactly one of `user_id`/`guest_token` must be set.
- `cart_items` — `id, cart_id, product_id, qty, unit_price_ghs_minor` (price captured at add-time)
- `orders` — `id, user_id, status, subtotal_ghs_minor, shipping_ghs_minor, total_ghs_minor, paystack_reference (unique), shipping_address jsonb, created_at, updated_at`. Status ∈ {`pending`, `paid`, `fulfilled`, `shipped`, `delivered`, `cancelled`, `refunded`}.
- `order_items` — `id, order_id, product_id, qty, unit_price_ghs_minor, product_name_snapshot, product_brand_snapshot` (snapshot fields so historical orders survive catalog changes)
- `addresses` — `id, user_id, label, line1, line2 (nullable), city, region, phone, is_default boolean`
- `wishlist_items` — `(user_id, product_id)` composite PK, `created_at`

**Static content (not in DB; JSON / Markdown files in `backend/internal/content/`)**

- `blog.json` — blog posts (seeded from `data.js`)
- `testimonials.json` — testimonials (seeded from `data.js`)
- `legal/*.md` — privacy, terms, returns policy
- `shipping_config.json` — `{flat_rate_ghs_minor: 2500, free_over_ghs_minor: 50000}` (GHS 25 flat, free over GHS 500)

**Cart-merge rule:** on login, the frontend reads its guest cart token from localStorage and POSTs it to `/cart/merge`. The backend looks up the guest cart, merges items into the user's existing cart by `product_id` (summing quantities; keeping the user's existing `unit_price_ghs_minor`), then deletes the guest cart row.

## 5. API Surface

REST under `/api/v1`. JSON in, JSON out. BetterAuth owns `/api/v1/auth/*`. Everything else is hand-written chi handlers.

**Catalog (public)**

- `GET /products` — query: `category`, `brand`, `tag`, `q`, `sort` (allowlisted enum), `page`, `limit`
- `GET /products/:slug`
- `GET /categories`
- `GET /brands`

**Content (public, served from JSON/Markdown files)**

- `GET /blog` · `GET /blog/:id`
- `GET /testimonials`
- `GET /legal/:slug`

**Cart (public — uses session or guest token)**

- `GET /cart` — returns `{items, subtotal_ghs_minor, shipping_cost_ghs_minor, free_shipping_remainder_ghs_minor, total_ghs_minor}`; creates a guest cart on first call if no auth + no token
- `POST /cart/items` — `{product_id, qty}`
- `PATCH /cart/items/:id` — `{qty}`
- `DELETE /cart/items/:id`
- `POST /cart/merge` — body: `{guest_token}`. Auth required.

**Shipping (public)**

- `GET /shipping/quote?subtotal=NNN` — returns `{flat_rate_ghs_minor, free_over_ghs_minor, applied_cost_ghs_minor, free_shipping_remainder_ghs_minor}`

**Wishlist (auth required)**

- `GET /wishlist`
- `POST /wishlist/:product_id`
- `DELETE /wishlist/:product_id`

**Account (auth required)**

- `GET /me` · `PATCH /me`
- `GET /me/addresses` · `POST /me/addresses` · `PATCH /me/addresses/:id` · `DELETE /me/addresses/:id`
- `GET /me/orders` · `GET /me/orders/:id`

**Checkout & payments**

- `POST /checkout/init` — auth required. Body: `{address_id, shipping_method}`. Re-prices cart from DB, snapshots line items, inserts `orders` + `order_items` with status `pending`, calls Paystack `transaction/initialize`. Returns `{authorization_url, reference}`.
- `POST /webhooks/paystack` — public; signature-verified via HMAC-SHA512 against `PAYSTACK_SECRET`. Idempotent: flips `pending → paid`, empties cart, fires confirmation email.
- `GET /checkout/verify/:reference` — auth required. Fallback for late webhooks. Frontend polls every 2s for up to 30s after redirect-back. Backend may call Paystack's verify endpoint if local status is still `pending`.

**Admin (admin role required)**

- `GET /admin/orders` — paginated, filterable by status
- `GET /admin/orders/:id`
- `PATCH /admin/orders/:id/status` — `paid → fulfilled → shipped → delivered`, or `→ cancelled`, or `→ refunded`. State machine validates transitions.
- `GET /admin/users` · `GET /admin/users/:id`
- `GET /admin/stats` — revenue, order counts by status, top products. Computed live, no caching.

**Cross-cutting:**

- Non-public routes go through `auth.RequireSession` middleware that validates the BetterAuth session cookie and injects `userID` + `role` into `r.Context()`.
- Admin routes go through an additional `auth.RequireRole("admin")` middleware **and** each handler calls `auth.MustBeAdmin(ctx)` at the top (redundant by design).
- Request bodies validated with `go-playground/validator` struct tags; failures return `400` with field-keyed error map.
- Errors follow a single shape: `{"error": {"code": "...", "message": "...", "fields": {...}?}}`.
- Rate limit on `/auth/*`: 10 req/min/IP via `httprate`.
- CORS: backend allows only the frontend origin.

## 6. Auth, Payments, Email Flows

### 6.1 Auth

- BetterAuth Go SDK + admin/roles plugin handles signup, login, OAuth (Google), email verification, password reset.
- Session: HTTP-only cookie, SameSite=Lax, 30-day rolling. Cookie scoped to eTLD+1 so it works on both subdomains.
- Frontend uses BetterAuth's React client for session state. Protected routes use TanStack Router's `beforeLoad` to redirect unauthenticated users; admin routes additionally check `role === 'admin'`.

### 6.2 Checkout (Paystack)

1. Frontend POSTs `/checkout/init` with `{address_id, shipping_method}`.
2. Backend re-prices cart from DB (never trusts client prices), snapshots line items, inserts `orders` + `order_items` with status `pending`, calls Paystack `transaction/initialize` with `amount = total_ghs_minor` and `reference = order_id`.
3. Backend returns `{authorization_url, reference}`. Frontend redirects.
4. After payment, Paystack redirects to `rue.example.com/checkout/return?reference=...`. Frontend polls `GET /checkout/verify/:reference` every 2s for up to 30s.
5. In parallel, Paystack POSTs `/webhooks/paystack`. Backend verifies HMAC-SHA512 signature, looks up the order by reference, flips `pending → paid` (idempotent), empties the cart, fires order-confirmation email.
6. Mobile money flows through the same Paystack hosted page (provider selection happens on Paystack's side).

**Failure modes:**

- Webhook never arrives → frontend's verify-poll asks the backend, which calls Paystack's verify endpoint if local status is still `pending`.
- User abandons Paystack page → order stays `pending`. No automatic expiry in v1 (documented as future work).
- Cart changes between init and webhook → irrelevant; the order is snapshotted at init time.

### 6.3 Email (Resend, allowlist mode)

- Env var `EMAIL_ALLOWLIST` = comma-separated addresses that receive real mail.
- For non-allowlisted recipients: backend logs the payload to stdout and returns success without calling Resend. Verification tokens for non-allowlisted addresses are auto-consumed server-side immediately so self-signup completes without a dead-end.
- Templates: signup welcome + verify, password reset, order confirmation, order shipped. Simple HTML, no templating engine beyond Go's `html/template`.
- In dev, `EMAIL_PROVIDER=mailpit` swaps Resend for SMTP-to-Mailpit (visible at `localhost:8025`).

## 7. Seed Data

A `backend/seed/main.go` program populates the DB from a clean state.

- **Catalog:** ports `Rue/src/data.js` (25 products, 9 categories, 13 brands).
- **Demo users:** one of each shape, with known email + password (printed on seed run):
  - Customer with 3 paid + 1 delivered orders, populated wishlist, 2 addresses.
  - Customer with delivered history (no active orders).
  - Customer with only wishlist items.
  - Customer in fresh/empty state.
  - Admin user.
- Open self-signup remains enabled — anyone can register a fresh account alongside the demo ones.
- Seed is idempotent: re-running drops and recreates demo data without touching real signups.

## 8. Frontend Routes (TanStack Router)

- `/` — home
- `/shop` — catalog list, filters, sort
- `/shop/:slug` — product detail
- `/about`, `/blog`, `/blog/:id`
- `/legal/:slug` — privacy, terms, returns
- `/cart` — full-page cart (drawer still exists for quick add)
- `/checkout` — auth-gated; address + shipping selection
- `/checkout/return` — Paystack redirect target, polls verify
- `/account` — auth-gated; overview
- `/account/orders`, `/account/orders/:id`, `/account/addresses`, `/account/wishlist`, `/account/settings`
- `/admin` — admin-only; overview/stats
- `/admin/orders`, `/admin/orders/:id`, `/admin/users`, `/admin/users/:id`
- `/login`, `/signup`, `/forgot-password`, `/reset-password`, `/verify-email`

> **Marketing pages from the mockup** (`marketing-pages.jsx` / `marketing.css`) are **deferred to future work**. They are press/newsletter/campaign-style landing pages that don't block the core e-commerce loop. The CSS file is ported into `frontend/src/styles/` so the visual language is available; routes are not wired in v1.

> **Domain names** (`rue.example.com`, `api.rue.example.com`) are placeholders throughout this spec. The actual subdomain pair will be chosen at deploy time and substituted into `Caddyfile`, the BetterAuth cookie config, and the CORS origin allowlist.

## 9. Visual Assets

- **Ship:** stock photos (~25) sourced from Unsplash/Pexels, committed to `frontend/public/products/`, referenced by filename in the seed.
- **Fallback if sourcing is a slog:** keep the existing CSS color-tone tiles (`lavender`, `cream`, `ink`, `rose`).
- **Styling approach:** Tailwind CSS v4. The existing mockup's design system is **mapped into Tailwind theme tokens** (palette swatches `--lavender-*`, `--ink`, `--cream`, font families, spacing rhythm, typography scale) in `tailwind.config.ts`. The legacy CSS files (`styles.css`, `pages.css`, `account.css`, `admin.css`, `legal.css`, `marketing.css`) are kept in the repo as `frontend/reference/legacy-css/` for **visual reference only** — they are not imported into the build.
- **Palette switching** (the existing `PALETTES` object in `app.jsx` with lavender/rose/sand/mint): preserved as CSS custom property overrides on `:root`, with Tailwind tokens referencing those custom properties. The tweaks panel UI is dropped (it was a design-tool artifact, not part of the product).
- Existing JSX components (`home.jsx`, `pages.jsx`, `shared.jsx`, `acct-pages.jsx`, `admin.jsx`, `legal-pages.jsx`) are **rebuilt** as TypeScript React under `frontend/src/components/` and `frontend/src/routes/` using Tailwind utilities, matching the visual reference. Rewired to use the API client instead of the in-memory `RueData` object. `marketing-pages.jsx` is deferred (see Section 8).

## 10. Quality Guarantees

### 10.1 Type safety (continuous)

- **Backend:** `go vet` + `staticcheck` + `golangci-lint` in pre-commit and CI. sqlc generates typed Go from SQL — query results are never `interface{}`.
- **Frontend:** `tsc --noEmit` in pre-commit and CI. `tsconfig` has `"strict": true` and `"noUncheckedIndexedAccess": true`. ESLint with `@typescript-eslint/recommended-type-checked`.
- **Wire boundary:** every API response is parsed through a Zod schema in `frontend/src/lib/api/schemas.ts`. CI generates TS types from Go via `swaggo/swag` → `openapi-typescript` and checks Zod schemas against generated types — drift fails the build.
- Pre-commit (lefthook) runs `gofmt`, `golangci-lint`, `tsc --noEmit`, `eslint`, `prettier` (with `prettier-plugin-tailwindcss` for class sorting).

### 10.2 SQL injection — structurally impossible

- All queries flow through sqlc with `$1, $2` placeholders.
- Dynamic sort/filter columns on `/products` and `/admin/orders` use an **allowlist enum** mapped to fixed `ORDER BY` clauses. Unknown sort values return 400.
- CI grep rule: `grep -rE "fmt\.Sprintf.*(SELECT|INSERT|UPDATE|DELETE)" backend/` must return nothing. `sqlclosecheck` lints `database/sql` resource handling.

### 10.3 RBAC — defense in depth

- **Layer 1 — router group:** admin routes mounted under a chi group with `RequireRole("admin")` middleware.
- **Layer 2 — handler:** every admin handler calls `auth.MustBeAdmin(ctx)` at the top. Redundant by design — survives router refactors.
- **Layer 3 — query scope:** `/me/*`, `wishlist`, `cart` queries always include `WHERE user_id = $1`. The `user_id` is taken from session context, **never** from request body or URL.
- **Tests:** an integration test matrix asserts customer-session → admin route → 403 for every admin endpoint, and customer-A → customer-B's resources → 404/403 (IDOR coverage).
- **Frontend:** UI hides admin nav for non-admins but treats backend as source of truth — any 403 redirects to `/`.

## 11. Testing

- **Backend unit:** table-driven tests for pure functions (cart pricing, shipping calc, cart merge, order status transitions).
- **Backend integration:** `testcontainers-go` spins a real Postgres per test package. Covers happy paths, the RBAC matrix, Paystack webhook signature verification (known-good + tampered payload), and idempotent webhook delivery.
- **Frontend unit:** Vitest for utilities + Zod schema round-trips.
- **Frontend component:** Testing Library for stateful components (cart drawer, checkout form).
- **E2E:** one Playwright happy-path: signup → browse → add to cart → checkout (Paystack test card) → see order in account. Runs in CI against a docker-compose stack.

## 12. CI/CD

- **GitHub Actions:**
  - On push: lint, type-check, unit tests (frontend + backend).
  - On PR to main: integration tests + E2E.
  - On merge to main: build artifacts (backend binary + frontend `dist/`).
- **Deploy:** manual `make deploy` target — `scp` binary + `dist/` to Hetzner, `systemctl restart rue-api`, `caddy reload`. Sub-second restart is acceptable (not zero-downtime).

## 13. Observability (v1, minimal)

- `slog` structured logs to stdout, captured by `journalctl`.
- `GET /healthz` endpoint for uptime checks.
- No metrics/tracing stack — documented as out of scope.

## 14. Security Checklist (v1)

- Paystack webhook HMAC-SHA512 signature verification.
- BetterAuth handles password hashing (argon2).
- All money math server-side; client prices ignored.
- BetterAuth cookie + SameSite=Lax provides CSRF protection.
- CORS restricted to the frontend origin.
- Secrets in `.env` (gitignored); `.env.example` ships in the repo.
- Rate limit on `/auth/*` via `httprate`.

## 15. Future Work (out of v1 scope)

- Admin product CRUD + image upload pipeline (Hetzner Object Storage).
- Customer-written product reviews.
- Coupon / discount code system.
- Return / RMA workflow.
- Blog/CMS UI for editing posts.
- Nightly job to expire `pending` orders older than 24 hours.
- Real product photos via upload pipeline (replacing the committed stock photo workflow).
- Metrics + tracing (Prometheus / OpenTelemetry).
- Wire the marketing pages from the original mockup (`marketing-pages.jsx`) into actual routes once their content/purpose is finalized.
