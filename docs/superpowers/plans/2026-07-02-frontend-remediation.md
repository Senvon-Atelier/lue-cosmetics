# Frontend Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the frontend to a buildable, type-clean state with a correct API contract chain (swag → OpenAPI → Orval react-query hooks), route-level auth guards grouped by audience (public / account / admin / checkout), and CI that prevents regression.

**Architecture:** The backend swagger annotations are normalized first (they are the root cause of broken admin URLs), then Orval is switched to generate TanStack Query hooks through a configured axios instance (mutator). Call sites are then fixed compiler-error-by-compiler-error. Routing is restructured into audience-grouped layout routes with `beforeLoad` guards backed by a shared session query.

**Tech Stack:** React 18, TanStack Router v1 (code-based routes), TanStack Query v5, Orval 7, axios, Zod, Vite 5, TypeScript 5.5, vitest, Go 1.25 + swaggo (annotation fixes only).

## Context (read this first)

This plan remediates audit findings. Key facts an implementer needs:

1. **`frontend/src/lib/api/client.ts` is dead code** — nothing imports it. The generated client (`src/lib/api/generated/rueCosmeticsAPI.ts`) imports the global `axios` directly, so no baseURL, no `withCredentials`, no error interceptor are ever applied.
2. **Swagger annotations are inconsistent.** `@BasePath /api/v1` is set globally, but `backend/internal/admin/handler.go` (10 routes) and two routes in `backend/internal/me/handler.go` embed `/api/v1` in their `@Router` paths. Generated URLs for those endpoints are double-prefixed.
3. **Admin list endpoints read query params** (`page`, `page_size`, `status`, `date_from`, `date_to`, `granularity`) **but declare no `@Param` annotations**, so the generated functions take no params argument, and pages pass filter objects as the axios config where they are silently dropped.
4. **`tsconfig.json` is invalid** (`"ignoreDeprecations": "6.0"` is rejected by TS 5.5), so `pnpm typecheck` and `pnpm build` fail before checking anything. With it removed there are **84 real type errors** underneath.
5. **`docs/superpowers/` is gitignored**, so newer specs/plans (including this one) are untracked.
6. Generated code names derive from URL paths. After the annotation fix, names change: `getApiV1AdminOrders` → `getAdminOrders`, etc. Every admin/me call site's imports change; tsc finds them all.

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."` (never commit with a personal address).
- Backend commands run from `ruecosmetics/backend/`; frontend commands from `ruecosmetics/frontend/`; `make` targets from `ruecosmetics/`.
- Never hand-edit anything under `frontend/src/lib/api/generated/` or `backend/docs/` — always regenerate (`make openapi`, `pnpm orval`).
- Requires: Go 1.25.8+, `swag` v1.16.4 (`go install github.com/swaggo/swag/cmd/swag@v1.16.4`), pnpm, Node 20+.
- `pnpm typecheck` must never get *worse* after a task than the count stated in that task's verification step.
- Currency amounts are GHS minor units (pesewas); render via `GH₵${(amount / 100).toLocaleString()}` as existing code does.

---

### Task 1: Repo hygiene — track docs, fix workspace config, remove strays

**Files:**
- Modify: `.gitignore`
- Modify: `frontend/pnpm-workspace.yaml`
- Create: `frontend/.env.example`
- Delete: `internal/` (empty stray directory at repo root, untracked)

**Interfaces:**
- Produces: `docs/superpowers/` tracked in git (later tasks commit plan checkboxes and specs).

- [ ] **Step 1: Fix `.gitignore`**

Remove these two lines from the end of `.gitignore` (keep everything else):

```
docs/superpowers/plans/2026-06-28-rue-cosmetics-frontend.md
docs/superpowers/
```

- [ ] **Step 2: Remove the stray empty directory at repo root**

```bash
rmdir internal/payments/paystack internal/payments internal
```

(The real Paystack client lives at `backend/internal/payments/paystack/` — do not touch that.)

- [ ] **Step 3: Fix `frontend/pnpm-workspace.yaml`**

Replace the entire file content (the current `allowBuilds` block is a literal unfilled placeholder, not valid config):

```yaml
packages:
  - '.'
onlyBuiltDependencies:
  - esbuild
```

- [ ] **Step 4: Create `frontend/.env.example`**

```bash
# Base URL for the API. Leave unset in dev — Vite proxies /api/* to :8080.
# In production set to the API origin, e.g. https://api.rue.example.com/api/v1
# VITE_API_URL=
```

- [ ] **Step 5: Verify pnpm still resolves the workspace**

Run from `frontend/`: `pnpm install --frozen-lockfile`
Expected: completes without error.

- [ ] **Step 6: Commit**

```bash
git add .gitignore docs/superpowers frontend/pnpm-workspace.yaml frontend/.env.example
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "chore: track docs/superpowers, fix pnpm workspace config, add frontend env example"
```

---

### Task 2: Restore a truthful typecheck

**Files:**
- Modify: `frontend/tsconfig.json:19`

- [ ] **Step 1: Remove the invalid compiler option**

Delete line 19 from `frontend/tsconfig.json`:

```json
    "ignoreDeprecations": "6.0",
```

- [ ] **Step 2: Verify typecheck now reports real errors**

Run from `frontend/`: `pnpm typecheck 2>&1 | wc -l`
Expected: TS5103 (invalid value) is gone; ~84 lines of real errors (TS6133/TS2339/TS2345/TS18048/TS2322 etc.). This is the expected baseline — later tasks burn it down to zero.

- [ ] **Step 3: Commit**

```bash
git add frontend/tsconfig.json
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "fix(frontend): remove invalid ignoreDeprecations from tsconfig"
```

---

### Task 3: Normalize swagger annotations (backend files, contract fix)

**Files:**
- Modify: `backend/internal/admin/handler.go` (lines 73, 160, 272, 388, 443, 519, 610, 704, 768, 893 — the `@Router` lines — plus `@Param` additions)
- Modify: `backend/internal/me/handler.go:120` and `backend/internal/me/handler.go:201`
- Regenerate: `backend/docs/{docs.go,swagger.json,swagger.yaml}`

**Interfaces:**
- Produces: OpenAPI paths relative to `@BasePath /api/v1` (e.g. `/admin/orders`, not `/api/v1/admin/orders`), and `GetAdminOrdersParams`-style query param schemas that Task 4's regeneration turns into typed params.

- [ ] **Step 1: Strip the `/api/v1` prefix from every `@Router` annotation that has one**

Run from `backend/` to find them (expect 12 matches):

```bash
grep -rn '@Router */api/v1' internal/
```

Edit each so e.g. `// @Router /api/v1/admin/orders [get]` becomes `// @Router /admin/orders [get]`. Full list: `admin/handler.go` — `/admin/dashboard`, `/admin/orders`, `/admin/orders/{id}`, `/admin/orders/{id}/status`, `/admin/customers`, `/admin/customers/{id}`, `/admin/products`, `/admin/products/{id}`, `/admin/analytics/revenue`, `/admin/analytics/stats`; `me/handler.go` — `/me/orders`, `/me/orders/{id}`.

- [ ] **Step 2: Add missing `@Param` annotations for query parameters**

The handlers read these via `httpx.QueryInt` / `r.URL.Query().Get` (verify against each handler body — `listOrders` is at `admin/handler.go:161`, `listCustomers` ~`:449`, `listProducts` ~`:616`, `getRevenueAnalytics` ~`:770`). Insert directly above the corresponding `@Router` line:

`listOrders` (`/admin/orders`):
```go
// @Param page      query int    false "Page number (1-based)"
// @Param page_size query int    false "Items per page"
// @Param status    query string false "Filter by order status"
// @Param date_from query string false "RFC3339 lower bound"
// @Param date_to   query string false "RFC3339 upper bound"
```

`listCustomers` (`/admin/customers`) and `listProducts` (`/admin/products`):
```go
// @Param page      query int false "Page number (1-based)"
// @Param page_size query int false "Items per page"
```

`getRevenueAnalytics` (`/admin/analytics/revenue`):
```go
// @Param granularity query string false "day|week|month"
// @Param date_from   query string false "RFC3339 lower bound"
// @Param date_to     query string false "RFC3339 upper bound"
```

- [ ] **Step 3: Regenerate OpenAPI docs**

Run from repo root: `make openapi`
Expected: `backend/docs/swagger.json` changes; verify no double-prefixed path remains:

```bash
grep -c '"/api/v1/' backend/docs/swagger.json
```
Expected: `0`.

- [ ] **Step 4: Verify backend still compiles and its tests still pass for the touched packages**

```bash
cd backend && go build ./... && go test ./internal/admin/... ./internal/me/... -timeout=600s
```
Expected: build OK; note `internal/admin` currently has no test files (that's a backend-plan item, not a failure).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/admin/handler.go backend/internal/me/handler.go backend/docs
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "fix(api): make @Router paths relative to BasePath and document admin query params"
```

---

### Task 4: Switch Orval to react-query hooks through a real axios instance

**Files:**
- Modify: `frontend/orval.config.ts` (full rewrite)
- Rewrite: `frontend/src/lib/api/client.ts`
- Regenerate: `frontend/src/lib/api/generated/`

**Interfaces:**
- Produces: `apiClient` (configured axios instance), `ApiError` (carries `status`, `code`, `fields`), `customInstance<T>(config): Promise<T>` (Orval mutator that unwraps `.data`), and generated hooks named `useGetAdminOrders`, `useGetAdminDashboard`, `useGetAdminOrdersId`, `usePatchAdminOrdersIdStatus`, `useGetAdminAnalyticsStats`, `useGetAdminAnalyticsRevenue`, `useGetAdminProducts`, `useGetAdminCustomers`, plus plain functions (`getAuthSession`, `postAuthLogin`, …) that now return **unwrapped response bodies** (no `AxiosResponse` wrapper).
- Consumes: Task 3's normalized swagger.json.

- [ ] **Step 1: Rewrite `frontend/src/lib/api/client.ts`**

```ts
import Axios, { AxiosError, AxiosRequestConfig } from 'axios';

export const apiClient = Axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api/v1',
  withCredentials: true,
});

interface ErrorEnvelope {
  error?: { code?: string; message?: string; fields?: Record<string, string> };
}

/** Typed error preserving the backend's httpx error envelope. */
export class ApiError extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;

  constructor(status: number, code: string, message: string, fields?: Record<string, string>) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ErrorEnvelope>) => {
    if (error.response) {
      const body = error.response.data?.error;
      throw new ApiError(
        error.response.status,
        body?.code ?? 'unknown',
        body?.message ?? `Request failed with status ${error.response.status}`,
        body?.fields,
      );
    }
    if (error.request) {
      throw new ApiError(0, 'network', 'No response from server. Please check your connection.');
    }
    throw new ApiError(0, 'unknown', error.message || 'An unexpected error occurred');
  },
);

/** Orval mutator: every generated call goes through apiClient and unwraps .data. */
export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> =>
  apiClient(config).then((res) => res.data as T);
```

- [ ] **Step 2: Rewrite `frontend/orval.config.ts`**

(The old config had `definitions`/`tags`/`operations`/`override.axios` keys that are not valid Orval options and were silently ignored.)

```ts
import { defineConfig } from 'orval';

export default defineConfig({
  rue: {
    input: {
      target: '../backend/docs/swagger.json',
    },
    output: {
      target: './src/lib/api/generated/',
      client: 'react-query',
      httpClient: 'axios',
      override: {
        mutator: {
          path: './src/lib/api/client.ts',
          name: 'customInstance',
        },
      },
    },
    hooks: {
      afterAllFilesWrite: 'prettier --write "src/lib/api/generated/**/*.{ts,tsx}"',
    },
  },
});
```

- [ ] **Step 3: Regenerate the client**

Run from `frontend/`: `rm -rf src/lib/api/generated && pnpm orval`
Expected: generation succeeds; verify hooks + params types exist:

```bash
grep -c "useGetAdminOrders\|GetAdminOrdersParams\|customInstance" src/lib/api/generated/*.ts
```
Expected: nonzero for each (file may be named after the API title; adjust glob accordingly).

- [ ] **Step 4: Record the new typecheck baseline**

Run: `pnpm typecheck 2>&1 | grep -c "error TS"`
Expected: errors *increase* (import names like `getApiV1AdminOrders` no longer exist; `.data` unwraps are now wrong). This is the compiler enumerating every call site Tasks 6–9 fix. Note the count in the commit message.

- [ ] **Step 5: Commit**

```bash
git add frontend/orval.config.ts frontend/src/lib/api/client.ts frontend/src/lib/api/generated
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): generate react-query hooks via configured axios mutator"
```

---

### Task 5: Session query, guard logic (TDD), and auth provider rewrite

**Files:**
- Modify: `frontend/src/features/shared/providers/query-provider.tsx` (export the client)
- Create: `frontend/src/lib/auth/session.ts`
- Create: `frontend/src/lib/auth/guards.ts`
- Test: `frontend/src/lib/auth/guards.test.ts`
- Rewrite: `frontend/src/lib/auth/auth-provider.tsx`

**Interfaces:**
- Consumes: `getAuthSession` (Task 4, now returns `InternalAuthSessionResponse` directly), `postAuthSignup`, `postAuthLogin`, `postAuthLogout`.
- Produces: `queryClient` (module export), `sessionQueryOptions`, `Session` type, `redirectPathFor(session, requirement): string | null`, and `useAuth()` with the same shape as today (`user`, `isLoading`, `isAuthenticated`, `isAdmin`, `signup`, `login`, `logout`, `refreshSession`). Task 6's router `beforeLoad` uses `queryClient` + `sessionQueryOptions` + `redirectPathFor`.

- [ ] **Step 1: Write the failing guard tests**

`frontend/src/lib/auth/guards.test.ts`:

```ts
import { describe, expect, it } from 'vitest';
import { redirectPathFor } from './guards';

describe('redirectPathFor', () => {
  it('sends anonymous users to /login regardless of requirement', () => {
    expect(redirectPathFor(null, 'authenticated')).toBe('/login');
    expect(redirectPathFor(null, 'admin')).toBe('/login');
  });

  it('lets authenticated customers into authenticated areas', () => {
    expect(redirectPathFor({ role: 'customer' }, 'authenticated')).toBeNull();
  });

  it('bounces non-admins from admin areas to the homepage', () => {
    expect(redirectPathFor({ role: 'customer' }, 'admin')).toBe('/');
    expect(redirectPathFor({}, 'admin')).toBe('/');
  });

  it('lets admins into both areas', () => {
    expect(redirectPathFor({ role: 'admin' }, 'authenticated')).toBeNull();
    expect(redirectPathFor({ role: 'admin' }, 'admin')).toBeNull();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run from `frontend/`: `pnpm vitest run src/lib/auth/guards.test.ts`
Expected: FAIL — `./guards` does not exist. (vitest needs no config for pure TS tests.)

- [ ] **Step 3: Implement `frontend/src/lib/auth/guards.ts`**

```ts
import type { Session } from './session';

export type GuardRequirement = 'authenticated' | 'admin';

/**
 * Pure routing decision: where to redirect (or null to allow).
 * Kept free of router/query imports so it is trivially unit-testable.
 */
export function redirectPathFor(
  session: Session | null,
  requirement: GuardRequirement,
): string | null {
  if (!session) return '/login';
  if (requirement === 'admin' && session.role !== 'admin') return '/';
  return null;
}
```

And `frontend/src/lib/auth/session.ts`:

```ts
import { queryOptions } from '@tanstack/react-query';
import { getAuthSession } from '../api/generated/rueCosmeticsAPI';
import type { InternalAuthSessionResponse } from '../api/generated/rueCosmeticsAPI';
import { ApiError } from '../api/client';

export type Session = InternalAuthSessionResponse;

export const sessionQueryOptions = queryOptions({
  queryKey: ['auth', 'session'] as const,
  queryFn: async (): Promise<Session | null> => {
    try {
      const session = await getAuthSession();
      return session ?? null;
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) return null;
      throw err;
    }
  },
  staleTime: 60_000,
});
```

(If Task 4's regeneration produced a different generated filename, fix the import path; do not rename the generated file.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm vitest run src/lib/auth/guards.test.ts`
Expected: 4 passed.

- [ ] **Step 5: Export the query client**

In `frontend/src/features/shared/providers/query-provider.tsx` change line 5 from `const queryClient = new QueryClient({` to `export const queryClient = new QueryClient({`. Nothing else changes.

- [ ] **Step 6: Rewrite `frontend/src/lib/auth/auth-provider.tsx`**

```tsx
import { createContext, useContext, ReactNode } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  postAuthSignup,
  postAuthLogin,
  postAuthLogout,
} from '../api/generated/rueCosmeticsAPI';
import { sessionQueryOptions, Session } from './session';

interface AuthContextType {
  user: Session | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  signup: (email: string, password: string, name?: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const { data: user = null, isLoading } = useQuery(sessionQueryOptions);

  const refreshSession = async () => {
    await queryClient.invalidateQueries({ queryKey: sessionQueryOptions.queryKey });
  };

  const signup = async (email: string, password: string, name?: string) => {
    await postAuthSignup({ email, password, name });
    await refreshSession();
  };

  const login = async (email: string, password: string) => {
    await postAuthLogin({ email, password });
    await refreshSession();
  };

  const logout = async () => {
    await postAuthLogout();
    queryClient.setQueryData(sessionQueryOptions.queryKey, null);
    await queryClient.invalidateQueries();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        isAdmin: user?.role === 'admin',
        signup,
        login,
        logout,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
```

Check `postAuthSignup`/`postAuthLogin` body types in the generated file — if signup takes `InternalAuthSignupBody`, the object literal `{ email, password, name }` already matches.

- [ ] **Step 7: Verify**

Run: `pnpm vitest run src/lib/auth/ && pnpm typecheck 2>&1 | grep -c 'lib/auth'`
Expected: tests pass; zero type errors under `src/lib/auth/`.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/lib/auth frontend/src/features/shared/providers/query-provider.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): shared session query, testable route-guard logic, query-backed auth provider"
```

---

### Task 6: Restructure the router into audience-grouped trees with beforeLoad guards

**Files:**
- Create: `frontend/src/features/home/home-page.tsx`
- Rewrite: `frontend/src/router.tsx`
- Modify: `frontend/src/features/admin/admin-layout.tsx` (remove useEffect guard, lines 1, 36–74)
- Modify: `frontend/src/features/account/account-layout.tsx` (remove its equivalent guard + the unused `Button` import on line 4)
- Modify: `frontend/src/features/home/journal-section.tsx` (add `id="journal"` anchor)

**Interfaces:**
- Consumes: `queryClient`, `sessionQueryOptions`, `redirectPathFor` (Task 5).
- Produces: route tree with ids `/_storefront/...` (public + auth pages), `/account/...` (guarded), `/admin/...` (guarded, own chrome — **no longer wrapped in the storefront header/footer**), `/_checkout/...`. Route id for admin order detail becomes `/admin/orders/$id` (Task 8 depends on this in `useParams({ from: '/admin/orders/$id' })`).

- [ ] **Step 1: Create `frontend/src/features/home/home-page.tsx`** (moves the composition out of the router)

```tsx
import { HomeHero } from './home-hero';
import { PromiseSection } from './promise-section';
import { CategoryRail } from './category-rail';
import { FeaturedProducts } from './featured-products';
import { JournalSection } from './journal-section';
import { TestimonialsSection } from './testimonials-section';
import { NewsletterSection } from './newsletter-section';

export function HomePage() {
  return (
    <div>
      <HomeHero />
      <PromiseSection />
      <CategoryRail />
      <FeaturedProducts />
      <JournalSection />
      <TestimonialsSection />
      <NewsletterSection />
    </div>
  );
}
```

- [ ] **Step 2: Rewrite `frontend/src/router.tsx`**

Delete the commented-out `AdminRoute` block (old lines 316–385) entirely. New structure — page wrappers that previously lived inline (the `min-h-screen bg-paper …` divs) move with their pages; keep them as thin wrappers here only where the page component doesn't already provide chrome:

```tsx
import { createRouter, createRoute, createRootRoute, Outlet, redirect, useParams } from '@tanstack/react-router';
import { QueryProvider, queryClient } from './features/shared/providers/query-provider';
import { AuthProvider } from './lib/auth/auth-provider';
import { CartProvider } from './features/cart/cart-provider';
import { sessionQueryOptions } from './lib/auth/session';
import { redirectPathFor, GuardRequirement } from './lib/auth/guards';
import { RootLayout, CheckoutLayout } from './features/shared/layouts';
import { HomePage } from './features/home/home-page';
import { ShopPage } from './features/catalog/shop-page';
import { ProductDetail } from './features/catalog/product-detail';
import { CartPage } from './features/cart/cart-page';
import { CheckoutPage } from './features/checkout/checkout-page';
import { CheckoutReturnPage } from './features/checkout/checkout-return';
import { AboutPage } from './features/content/about-page';
import { AccountLayout } from './features/account/account-layout';
import { AccountDashboard } from './features/account/account-dashboard';
import { AccountOrders } from './features/account/account-orders';
import { AccountOrderDetail } from './features/account/account-order-detail';
import { AccountAddresses } from './features/account/account-addresses';
import { AccountWishlist } from './features/account/account-wishlist';
import { AccountSettings } from './features/account/account-settings';
import { LoginPage } from './features/auth/login-page';
import { SignupPage } from './features/auth/signup-page';
import { ForgotPasswordPage } from './features/auth/forgot-password-page';
import { ResetPasswordPage } from './features/auth/reset-password-page';
import { VerifyEmailPage } from './features/auth/verify-email-page';
import { AdminLayout } from './features/admin/admin-layout';
import { AdminDashboard } from './features/admin/dashboard';
import { AdminOrders } from './features/admin/orders';
import { AdminOrderDetail } from './features/admin/orders/order-detail';
import { AdminProducts } from './features/admin/products';
import { AdminCustomers } from './features/admin/customers';
import { AdminAnalytics } from './features/admin/analytics';
import { AdminMarketing } from './features/admin/marketing';
import { AdminContent } from './features/admin/content';
import { AdminSettings } from './features/admin/settings';

// Session-backed route guard. Throws a redirect before anything renders.
async function requireRole(requirement: GuardRequirement) {
  const session = await queryClient.ensureQueryData(sessionQueryOptions);
  const target = redirectPathFor(session, requirement);
  if (target) throw redirect({ to: target });
}

const rootRoute = createRootRoute({
  component: () => (
    <QueryProvider>
      <AuthProvider>
        <CartProvider>
          <Outlet />
        </CartProvider>
      </AuthProvider>
    </QueryProvider>
  ),
});

// ── Storefront: public pages + auth pages (Header + Footer + CartDrawer) ────
const storefrontLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_storefront',
  component: RootLayout,
});

const pageShell = (children: React.ReactNode) => (
  <div className="min-h-screen bg-paper text-ink font-body">{children}</div>
);

const homeRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/',
  component: HomePage,
});

const shopRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/shop',
  component: () => pageShell(<ShopPage />),
});

function ProductDetailComponent() {
  const { slug } = useParams({ from: '/_storefront/shop/$slug' });
  return pageShell(<ProductDetail slug={slug || ''} />);
}

const productDetailRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/shop/$slug',
  component: ProductDetailComponent,
});

const cartRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/cart',
  component: () => pageShell(<CartPage />),
});

const aboutRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/about',
  component: () => pageShell(<AboutPage />),
});

const loginRoute = createRoute({ getParentRoute: () => storefrontLayoutRoute, path: '/login', component: LoginPage });
const signupRoute = createRoute({ getParentRoute: () => storefrontLayoutRoute, path: '/signup', component: SignupPage });
const forgotPasswordRoute = createRoute({ getParentRoute: () => storefrontLayoutRoute, path: '/forgot-password', component: ForgotPasswordPage });
const resetPasswordRoute = createRoute({ getParentRoute: () => storefrontLayoutRoute, path: '/reset-password', component: ResetPasswordPage });
const verifyEmailRoute = createRoute({ getParentRoute: () => storefrontLayoutRoute, path: '/verify-email', component: VerifyEmailPage });

// ── Account: authenticated customers, storefront chrome ─────────────────────
const accountRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/account',
  beforeLoad: () => requireRole('authenticated'),
  component: AccountLayout,
});

const accountDashboardRoute = createRoute({ getParentRoute: () => accountRoute, path: '/', component: AccountDashboard });
const accountOrdersRoute = createRoute({ getParentRoute: () => accountRoute, path: '/orders', component: AccountOrders });
const accountOrderDetailRoute = createRoute({ getParentRoute: () => accountOrdersRoute, path: '$id', component: AccountOrderDetail });
const accountAddressesRoute = createRoute({ getParentRoute: () => accountRoute, path: '/addresses', component: AccountAddresses });
const accountWishlistRoute = createRoute({ getParentRoute: () => accountRoute, path: '/wishlist', component: AccountWishlist });
const accountSettingsRoute = createRoute({ getParentRoute: () => accountRoute, path: '/settings', component: AccountSettings });

// ── Admin: admin role required, own chrome (no storefront header/footer) ────
const adminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  beforeLoad: () => requireRole('admin'),
  component: AdminLayout,
});

const adminDashboardRoute = createRoute({ getParentRoute: () => adminRoute, path: '/', component: AdminDashboard });
const adminOrdersRoute = createRoute({ getParentRoute: () => adminRoute, path: '/orders', component: AdminOrders });
const adminOrderDetailRoute = createRoute({ getParentRoute: () => adminRoute, path: '/orders/$id', component: AdminOrderDetail });
const adminProductsRoute = createRoute({ getParentRoute: () => adminRoute, path: '/products', component: AdminProducts });
const adminProductDetailRoute = createRoute({ getParentRoute: () => adminRoute, path: '/products/$id', component: () => <div>Admin Product Detail - Coming Soon...</div> });
const adminCustomersRoute = createRoute({ getParentRoute: () => adminRoute, path: '/customers', component: AdminCustomers });
const adminCustomerDetailRoute = createRoute({ getParentRoute: () => adminRoute, path: '/customers/$id', component: () => <div>Admin Customer Detail - Coming Soon...</div> });
const adminAnalyticsRoute = createRoute({ getParentRoute: () => adminRoute, path: '/analytics', component: AdminAnalytics });
const adminMarketingRoute = createRoute({ getParentRoute: () => adminRoute, path: '/marketing', component: AdminMarketing });
const adminContentRoute = createRoute({ getParentRoute: () => adminRoute, path: '/content', component: AdminContent });
const adminSettingsRoute = createRoute({ getParentRoute: () => adminRoute, path: '/settings', component: AdminSettings });

// ── Checkout: minimal chrome ─────────────────────────────────────────────────
const checkoutLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_checkout',
  component: CheckoutLayout,
});

const checkoutRoute = createRoute({ getParentRoute: () => checkoutLayoutRoute, path: '/checkout', component: CheckoutPage });
const checkoutReturnRoute = createRoute({
  getParentRoute: () => checkoutLayoutRoute,
  path: '/checkout/return',
  component: () => pageShell(<CheckoutReturnPage />),
});

const routeTree = rootRoute.addChildren([
  storefrontLayoutRoute.addChildren([
    homeRoute,
    shopRoute,
    productDetailRoute,
    cartRoute,
    aboutRoute,
    loginRoute,
    signupRoute,
    forgotPasswordRoute,
    resetPasswordRoute,
    verifyEmailRoute,
    accountRoute.addChildren([
      accountDashboardRoute,
      accountOrdersRoute.addChildren([accountOrderDetailRoute]),
      accountAddressesRoute,
      accountWishlistRoute,
      accountSettingsRoute,
    ]),
  ]),
  adminRoute.addChildren([
    adminDashboardRoute,
    adminOrdersRoute,
    adminOrderDetailRoute,
    adminProductsRoute,
    adminProductDetailRoute,
    adminCustomersRoute,
    adminCustomerDetailRoute,
    adminAnalyticsRoute,
    adminMarketingRoute,
    adminContentRoute,
    adminSettingsRoute,
  ]),
  checkoutLayoutRoute.addChildren([checkoutRoute, checkoutReturnRoute]),
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
```

Note the `/shop` heading block that used to live in the router: move it into `shop-page.tsx` as the first element ShopPage renders (an `<h1 className="font-display text-4xl mb-2">Shop</h1>` plus the description paragraph inside the existing max-width wrapper), so the route stays a one-liner.

- [ ] **Step 3: Strip the imperative guards from the layouts**

In `admin-layout.tsx`: delete the `useEffect` import and guard block (old lines 40–49) and the `if (!isAuthenticated)` / `if (!isAdmin)` early returns (old lines 59–74). Keep the `isLoading` spinner and `useAuth()` for the welcome text. The route's `beforeLoad` now owns access control.

In `account-layout.tsx`: same treatment — remove its redirect `useEffect` and unauthenticated early-return, and delete the unused `Button` import (line 4).

- [ ] **Step 4: Add the journal anchor**

In `frontend/src/features/home/journal-section.tsx`, add `id="journal"` to the section's root element (e.g. `<section id="journal" …>`). Task 7 points the header/footer "Journal" links at it.

- [ ] **Step 5: Verify**

```bash
pnpm typecheck 2>&1 | grep -c "router.tsx\|admin-layout\|account-layout"
```
Expected: `0` errors in these files. Then `pnpm dev` and manually confirm: `/` renders; `/admin` while logged out redirects to `/login`; `/account` while logged out redirects to `/login`.

- [ ] **Step 6: Commit**

```bash
git add frontend/src
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "refactor(frontend): audience-grouped route trees with beforeLoad session guards"
```

---

### Task 7: Mechanical type-error cleanup outside admin

**Files:**
- Modify: `frontend/src/features/cart/cart-drawer.tsx:109,189`, `cart/cart-page.tsx:59,150`, `catalog/product-detail.tsx:191`, `checkout/checkout-page.tsx:308`, `home/featured-products.tsx:47,74`, `home/category-rail.tsx:66` — icon name
- Modify: `frontend/src/features/shared/layouts/header.tsx:28`, `footer.tsx:50` — journal links
- Modify: `frontend/src/features/home/category-rail.tsx:81-83`, `checkout/checkout-page.tsx:361`, `catalog/filter-bar.tsx:25`, `catalog/shop-page.tsx:5`

**Interfaces:** none new; pure cleanup driven by the tsc error inventory.

- [ ] **Step 1: Replace the nonexistent `arrowRight` icon name with `arrow`**

The icon map in `features/shared/ui/icons.tsx` has `arrow` (right-pointing), `arrowLeft`, `arrowUp` — no `arrowRight`. Replace the string at all 8 call sites:

```bash
cd frontend && grep -rl '"arrowRight"' src/features | xargs sed -i '' 's/"arrowRight"/"arrow"/g'
```

- [ ] **Step 2: Fix the `/journal` links (route doesn't exist)**

In `header.tsx:28` and `footer.tsx:50`, change the Journal nav entry from `to: '/journal'` to link to the homepage journal anchor added in Task 6:

```tsx
<Link to="/" hash="journal">Journal</Link>
```

(Adapt to how each file structures its nav items — if links are data-driven objects, change the entry to `{ label: 'Journal', to: '/', hash: 'journal' }` and pass `hash` through to `Link`.)

- [ ] **Step 3: Remove the `product_count` display in `category-rail.tsx`**

`InternalCatalogCategoryView` has no `product_count` field (the backend doesn't return counts). At lines 81–83, delete the count rendering and keep only the category label. If a count is wanted later, that's a backend feature request, not a frontend cast.

- [ ] **Step 4: Fix the remaining small errors**

- `checkout-page.tsx:361` — `Icon` is used but not imported: add `Icon` to the existing import from `'../shared/ui'` (or `'../shared/ui/icons'`, matching the file's other imports).
- `filter-bar.tsx:25` — `showFilters`/`setShowFilters` unused: delete the `useState` line (and the `useState` import if now unused).
- `shop-page.tsx:5` — unused `Icon` import: delete it.

- [ ] **Step 5: Verify and commit**

```bash
pnpm typecheck 2>&1 | grep -c "features/cart\|features/catalog\|features/home\|features/checkout\|shared/layouts"
```
Expected: `0` (checkout may still show generated-client call-site errors — those belong to Task 9; only the errors listed in this task must be gone).

```bash
git add frontend/src
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "fix(frontend): icon names, journal anchor links, stale field usages, unused imports"
```

---

### Task 8: Move admin pages onto generated hooks

**Files:**
- Modify: `frontend/src/features/admin/dashboard/index.tsx`, `orders/index.tsx`, `orders/order-detail.tsx`, `analytics/index.tsx`, `products/index.tsx`, `customers/index.tsx`, `settings/index.tsx:97,112`, `content/index.tsx:101`

**Interfaces:**
- Consumes: generated hooks from Task 4 (`useGetAdminDashboard`, `useGetAdminOrders`, `useGetAdminOrdersId`, `usePatchAdminOrdersIdStatus`, `useGetAdminAnalyticsStats`, `useGetAdminAnalyticsRevenue`, `useGetAdminProducts`, `useGetAdminCustomers`) and their `Params` types. Hooks return **unwrapped** bodies: `data` is `InternalAdminOrdersResponse`, not `AxiosResponse`.
- Route id `/admin/orders/$id` from Task 6.

General rules for every page in this task:
1. Replace `useQuery({ queryKey: [...], queryFn: () => getApiV1... })` with the generated hook; delete the hand-rolled queryKey.
2. Replace `data?.data.X` with `data?.X` (mutator unwraps).
3. All generated response fields are optional — guard with `?? 0` / `?? ''` / `?? []` at use sites rather than non-null assertions.
4. Placeholder/demo content (fake activity feeds, fallback chart bars) stays as-is; this task only fixes the data plumbing and type errors.

- [ ] **Step 1: `dashboard/index.tsx`**

Replace lines 1–9 with:

```tsx
import { useGetAdminDashboard } from '../../../lib/api/generated/rueCosmeticsAPI';
import { KPICard, StatusTag, Panel } from '../../shared/ui/admin';

export function AdminDashboard() {
  const { data: dashboard, isLoading, error } = useGetAdminDashboard();
```

Then: `const stats = dashboard?.stats;` and `const recentOrders = dashboard?.recent_orders ?? [];` (drop `.data`). Delete the unused `formatDate` (lines 34–40). Field guards: line 66 `formatCurrency(stats.total_revenue_ghs_minor ?? 0)`; line 72 `stats?.total_orders ?? 0`; line 78 `formatCurrency((stats.total_revenue_ghs_minor ?? 0) / Math.max(stats.total_orders ?? 1, 1))`; line 84 `stats?.total_customers ?? 0`; line 180/182 `key={order.id}` → `key={order.id ?? ''}` and `{(order.id ?? '').slice(0, 8).toUpperCase()}`; line 188 `formatCurrency(order.total_ghs_minor ?? 0)`; line 191 `<StatusTag status={order.status ?? 'pending'} />`.

- [ ] **Step 2: `orders/index.tsx`**

Replace the query and mutation blocks (lines 4, 13–25):

```tsx
import {
  useGetAdminOrders,
  getGetAdminOrdersQueryKey,
  usePatchAdminOrdersIdStatus,
  getGetAdminDashboardQueryKey,
} from '../../../lib/api/generated/rueCosmeticsAPI';

// inside the component:
const { data: ordersData, isLoading, error } = useGetAdminOrders({
  page: page + 1,
  page_size: 20,
  ...(statusFilter ? { status: statusFilter } : {}),
});

const updateStatusMutation = usePatchAdminOrdersIdStatus({
  mutation: {
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: getGetAdminOrdersQueryKey() });
      queryClient.invalidateQueries({ queryKey: getGetAdminDashboardQueryKey() });
    },
  },
});
// call sites change shape: updateStatusMutation.mutate({ id, data: { status } })
```

(Check the generated file for the exact hook/key names and mutation variable shape — Orval generates `mutate({ id, data })` for path-param + body operations.) Unwrap: `ordersData?.total_pages ?? 1`, `ordersData?.orders ?? []`. Remove the unused `StatusTag` import if truly unused after edits (line 5). Field guards at lines 149–159 per the same `?? ''` / `?? 0` pattern as the dashboard.

- [ ] **Step 3: `orders/order-detail.tsx`**

Lines 1–13 become:

```tsx
import { useParams } from '@tanstack/react-router';
import { useGetAdminOrdersId } from '../../../lib/api/generated/rueCosmeticsAPI';
import { StatusTag, Panel } from '../../shared/ui/admin';

export function AdminOrderDetail() {
  const { id } = useParams({ from: '/admin/orders/$id' });
  const { data: orderDetail, isLoading, error } = useGetAdminOrdersId(id, {
    query: { enabled: !!id },
  });
```

Unwrap: `const order = orderDetail?.order;` / `const items = orderDetail?.items ?? [];`. Guards: line 58 `#{(order.id ?? '').slice(0, 8).toUpperCase()}`; line 90–95 `{item.qty ?? 0}`, `formatCurrency(item.unit_price_ghs_minor ?? 0)`, `formatCurrency((item.unit_price_ghs_minor ?? 0) * (item.qty ?? 0))`; line 111 `status={order.status ?? 'pending'}`; lines 115/120 `formatDate(order.created_at ?? '')` / `formatDate(order.updated_at ?? '')`; lines 131–139 wrap each `_ghs_minor` in `?? 0`.

- [ ] **Step 4: `analytics/index.tsx`**

Lines 1–22 become:

```tsx
import {
  useGetAdminAnalyticsStats,
  useGetAdminAnalyticsRevenue,
} from '../../../lib/api/generated/rueCosmeticsAPI';
import { KPICard, Panel } from '../../shared/ui/admin';

export function AdminAnalytics() {
  const granularity = 'month'; // switcher UI not built yet; keep a constant, not dead state
  const { data: stats, isLoading: statsLoading } = useGetAdminAnalyticsStats();
  const { data: revenueData, isLoading: revenueLoading } = useGetAdminAnalyticsRevenue({
    granularity,
    date_from: '2024-01-01T00:00:00Z',
    date_to: '2024-12-31T23:59:59Z',
  });
```

Unwrap: `const topProducts = stats?.top_products ?? [];`, `const revenueByDate = revenueData?.by_date ?? [];`, `const revenueByCategory = revenueData?.by_category ?? [];`. Fix the wrong field names against the real generated types: `InternalAdminTopProduct` has `id`, `name`, `revenue_ghs_minor` (there is **no** `product_id`/`units_sold`) → line 123 `key={product.id}`, line 125 drop the Units cell content to `—` or remove the Units column; line 127 `formatCurrency(product.revenue_ghs_minor ?? 0)`. `InternalAdminRevenueByCategory` has `category_name`/`category_slug` (no `category`) → lines 158/161 use `item.category_name`. Guards at lines 85–86 (`item.revenue_ghs_minor ?? 0`) and 163. Line 144 (placeholder table) — its tuple type confuses tsc: type the placeholder array explicitly as `[string, number, number][]`.

- [ ] **Step 5: `products/index.tsx` and `customers/index.tsx`**

Products: swap to `useGetAdminProducts({ page: page + 1, page_size: 20 })`; unwrap `productsData?.total_pages ?? 1` / `productsData?.products ?? []`; delete unused `formatDate` (line 25) and the unused `product` parameter in `getStockStatus` (line 33 — change to `getStockStatus()` or remove the helper if the column just renders 'live'); guards: line 43 `(p.name ?? '').toLowerCase()` and `(p.slug ?? '').toLowerCase()`; lines 176–179 `product.category_id ?? '—'`, `product.brand_id ?? '—'`, `formatCurrency(product.price_ghs_minor ?? 0)`.

Customers: swap to `useGetAdminCustomers({ page: page + 1, page_size: 20 })` (the current code nests `{ params: {...} }` in axios config — replace it) and unwrap `customersData?.customers ?? []` etc. This file has no tsc errors today but the hook swap keeps the section consistent.

- [ ] **Step 6: `settings/index.tsx` and `content/index.tsx`**

`settings/index.tsx:97` — `name` possibly undefined: guard with `name ?? ''` (or `?? '—'` to match surrounding copy). `settings/index.tsx:112` and `content/index.tsx:101` — unused `i` in `.map((x, i) => …)`: remove the second parameter.

- [ ] **Step 7: Verify and commit**

```bash
pnpm typecheck 2>&1 | grep -c "features/admin"
```
Expected: `0`.

```bash
git add frontend/src/features/admin
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "refactor(admin): use generated react-query hooks with typed params and unwrapped responses"
```

---

### Task 9: Account/checkout call-site fixes and the zero-error gate

**Files:**
- Modify: `frontend/src/features/account/account-orders.tsx`, `account-order-detail.tsx`, `account-addresses.tsx`, `checkout/checkout-page.tsx`, `checkout/checkout-return.tsx` (call sites of renamed/unwrapped generated functions)
- Modify: whatever `pnpm typecheck` still flags — this task ends at zero

**Interfaces:**
- Consumes: renamed generated functions (`getMeOrders`, `getMeOrdersId`, address + checkout functions — read the generated file for exact names) returning unwrapped bodies.

- [ ] **Step 1: Enumerate every remaining error**

```bash
pnpm typecheck 2>&1 | tee /tmp/remaining-errors.txt
```

- [ ] **Step 2: Apply the two mechanical transformations everywhere they appear**

1. Import renames: `getApiV1MeOrders` → `getMeOrders` style (compiler names the missing import; the generated file names the replacement). Where a page fetches in `useQuery`, prefer the generated hook (`useGetMeOrders(...)`) exactly as in Task 8.
2. Unwraps: `x?.data.field` → `x?.field`; responses are already bodies.

- [ ] **Step 3: Zero-error gate — all three must pass**

```bash
pnpm typecheck   # expected: exit 0, no output
pnpm build       # expected: tsc passes, vite build emits dist/
pnpm lint        # expected: exit 0 (zero warnings allowed by the script)
pnpm vitest run  # expected: guard tests pass
```

If lint flags issues beyond what typecheck caught (unused eslint-disable, hook deps), fix them here.

- [ ] **Step 4: Manual smoke test (requires backend running: `make dev` from repo root, DB migrated + seeded)**

`pnpm dev`, then verify in the browser: home page loads products (network tab shows `/api/v1/products` → 200 JSON, **not** index.html); login works and `/account` shows the session's name; an admin user reaches `/admin` and the orders list paginates (network tab shows `page=2` actually sent).

- [ ] **Step 5: Commit**

```bash
git add frontend/src
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "fix(frontend): migrate remaining call sites to regenerated client; typecheck, build, lint all green"
```

---

### Task 10: CI workflow

**Files:**
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: `make drift-check` (regenerates OpenAPI + sqlc and fails on diff), `make test`, frontend scripts.

- [ ] **Step 1: Create `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
      - name: Install codegen tools
        run: |
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0
          go install github.com/swaggo/swag/cmd/swag@v1.16.4
      - name: Drift check (OpenAPI + sqlc)
        run: make drift-check
      - name: Tests (testcontainers use the runner's Docker daemon)
        run: make test

  frontend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm
          cache-dependency-path: frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm typecheck
      - run: pnpm lint
      - run: pnpm vitest run
      - run: pnpm build
```

- [ ] **Step 2: Verify locally what CI will run**

```bash
make drift-check        # from repo root; expected: passes with no diff
cd frontend && pnpm typecheck && pnpm lint && pnpm vitest run && pnpm build
```
Expected: all pass (they did at the end of Task 9).

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "ci: add backend drift-check/tests and frontend typecheck/lint/test/build"
```

---

### Task 11: Documentation truth-up

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Fix the README**

- Prerequisites: change `Go 1.22+` to `Go 1.25.8+` (matches `backend/go.mod`); add `pnpm 9+` and `Node 20+` for the frontend.
- Add a frontend quickstart section after the backend one:

```markdown
## Quickstart (frontend)

```bash
cd frontend
pnpm install
pnpm dev                      # Vite on :5173, proxies /api/* to :8080
```

Regenerate the API client after backend contract changes:

```bash
make openapi && cd frontend && pnpm orval
```
```

- In the Tests section add: `cd frontend && pnpm vitest run` and note that `make seed-run` requires `DATABASE_URL` exported in the shell (it does not read `backend/.env`).

- [ ] **Step 2: Commit**

```bash
git add README.md
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "docs: correct Go version, add frontend quickstart and client-regen workflow"
```
