# Account Suite UI Alignment (Tranche 2) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the account suite (layout/sidebar, dashboard, orders, order detail, addresses, wishlist, settings) look like the `Rue/` account mockup while keeping every existing hook, handler, guard, and route path working.

**Architecture:** Account routes move out of the storefront chrome into the mockup's standalone `.acct-layout` (sticky sidebar + per-page `.acct-main`). Each page port copies the mockup's JSX structure/class names from `Rue/src/acct-pages.jsx`, keeps the app's real data wiring, and strips dead Tailwind classNames. Page CSS goes in a new `src/styles/account.css` ported from `Rue/account.css` (dedup rule: skip anything already in globals/pages/auth CSS).

**Tech Stack:** React 18, TanStack Router v1 (code-based routes in `src/router.tsx`), Orval-generated client (`src/lib/api/generated/rueCosmeticsAPI.ts` — plain async functions, responses unwrapped), plain CSS over custom properties, vitest.

**Spec:** `docs/superpowers/specs/2026-07-02-account-ui-alignment-design.md` — binding. **Mockup source (canonical, read-only, NEVER modify):** `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/` (`src/acct-pages.jsx`, `account.css`, `styles.css`).

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."`.
- Frontend-only; nothing under `backend/` or `frontend/src/lib/api/generated/` changes.
- Functionality freeze: `getMeOrders`/`getMeOrdersId`/`getMeAddresses`/`postMeAddresses`/`patchMeAddressesId`/`deleteMeAddressesId`/`postMeAddressesIdDefault` calls, the `beforeLoad` `requireRole('authenticated')` guard, and all `/account/*` URLs keep working.
- Honest data only: no fake points, tiers, timelines, item counts, payment masks, or wishlist products.
- Class parity: ported markup keeps mockup class names verbatim; adapted/extrapolated classes go under a `/* adapted/extrapolated */` banner in `account.css`.
- Strip every dead Tailwind utility className from each file as it is ported (Tailwind is dormant — they generate nothing). Ported pages use plain `<button className="btn btn-primary">` etc., NOT the shared `Button` component.
- Prices via `formatGhs` (`src/lib/format/utils.ts:86`, minor units in). Images via `getImageUrl` (`utils.ts:75`). Icons via `Icon` from `features/shared/ui/icons.tsx` (`user`, `bag`, `heart`, `arrow`, `arrowLeft`, `plus`, `check` all exist).
- Layouts must NOT wrap `<Outlet/>` in `<main>` — each page renders its own `<main className="acct-main">`.
- Gate after every task: `pnpm typecheck` zero errors (run from `frontend/`). Full gate (typecheck/lint/vitest/build) in the final task.
- Commands run from `ruecosmetics/frontend/` unless stated.

---

### Task 1: Foundation — account.css, missing .brand fix, regions const, date formatter

**Files:**
- Create: `frontend/src/styles/account.css`
- Modify: `frontend/src/styles/globals.css` (add `@import './account.css';` after the auth.css import; add the missing `.brand*` block near the header styles ~line 989)
- Create: `frontend/src/content/regions.ts`
- Modify: `frontend/src/lib/format/utils.ts` (append `formatOrderDate`)
- Test: `frontend/src/lib/format/utils.test.ts` (extend)

**Interfaces:**
- Produces: `formatOrderDate(iso: string | undefined): string` — `"2026-04-04T11:24:00Z"` → `"Apr 04, 2026"`, undefined/invalid → `""`.
- Produces: `GHANA_REGIONS: readonly string[]` (16 regions) from `src/content/regions.ts`.
- Produces: all `acct-*`, `orders-*`, `status*`, `form-card`, `addr-*`, `sub-tabs`, `dash-link*`, `alert*` classes + adapted `.order-detail-grid`, `.kv-row`, `.od-item*`, `.acct-side-link-label`, `.acct-side-foot`, `.form-card-title`, status palette. Later tasks' markup depends on these exact names.

- [ ] **Step 1: Write the failing formatOrderDate test** (append to `src/lib/format/utils.test.ts`)

```ts
// add formatOrderDate to the file's existing `from './utils'` import — do not add a second import line

describe('formatOrderDate', () => {
  it('renders mockup-style dates', () => {
    expect(formatOrderDate('2026-04-04T11:24:00Z')).toBe('Apr 04, 2026');
  });
  it('returns empty string for missing/invalid input', () => {
    expect(formatOrderDate(undefined)).toBe('');
    expect(formatOrderDate('not-a-date')).toBe('');
  });
});
```

(`describe`/`expect`/`it` are already imported at the top of the file — reuse that import.)

- [ ] **Step 2: Run it to make sure it fails**

Run: `pnpm vitest run src/lib/format/utils.test.ts`
Expected: FAIL — `formatOrderDate` is not exported.

- [ ] **Step 3: Implement `formatOrderDate`** (append to `src/lib/format/utils.ts`)

```ts
/** Mockup-style order date: "Apr 04, 2026". Empty string when absent/invalid. */
export function formatOrderDate(iso: string | undefined): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  });
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm vitest run src/lib/format/utils.test.ts` — Expected: all pass.

- [ ] **Step 5: Create `src/content/regions.ts`**

```ts
/**
 * Ghana's 16 administrative regions, for address-form selects.
 * (Mockup hardcodes 6; we ship the real list — spec §2.5.)
 */
export const GHANA_REGIONS = [
  'Ahafo',
  'Ashanti',
  'Bono',
  'Bono East',
  'Central',
  'Eastern',
  'Greater Accra',
  'North East',
  'Northern',
  'Oti',
  'Savannah',
  'Upper East',
  'Upper West',
  'Volta',
  'Western',
  'Western North',
] as const;
```

- [ ] **Step 6: Fix the missing `.brand*` classes in `globals.css`**

Pre-existing bug (tranche-1 pitfall class): `features/shared/ui/brand.tsx` renders `.brand/.brand-mark/.brand-word/.brand-tag` but no frontend CSS defines them (`grep -rn '\.brand' src/styles/` → no hits). Port verbatim from `Rue/styles.css` lines 246–277, placed immediately BEFORE the `.header` rule (~line 989) with a comment:

```css
/* brand mark (ported from Rue/styles.css — was missing; used by header + account sidebar) */
.brand {
  justify-self: center;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.brand-mark {
  width: 34px; height: 34px;
  border-radius: 999px;
  border: 1px solid var(--lavender-400);
  display: grid;
  place-items: center;
  color: var(--lavender-600);
}
.brand-word {
  font-family: var(--font-serif);
  font-style: italic;
  font-weight: 400;
  font-size: 30px;
  color: var(--ink);
  line-height: 1;
  letter-spacing: -0.01em;
}
.brand-tag {
  font-family: var(--font-label);
  font-size: 9px;
  letter-spacing: 0.35em;
  text-transform: uppercase;
  color: var(--lavender-700);
  margin-top: 2px;
}
```

Also check `Rue/styles.css:379` (`.brand { justify-self: start; }` inside a media query) — find that media block and port the one-liner into the frontend's matching responsive section if the frontend has the corresponding `@media` block for the header; if the frontend header section has no such block, add it verbatim next to the `.brand` rules.

- [ ] **Step 7: Create `src/styles/account.css`**

Dedup verified against current files: `.field*`, `.form-row*`, `.field-error` live in `pages.css`; `.auth-*`, `.back-link` live in `auth.css`; `.badge` exists in `globals.css` (the sidebar's `.acct-side a .badge` is NOT ported — no badges per spec §2.2). Do not redefine any of those. Full file content:

```css
/* Ported from Rue/account.css — account suite (Tranche 2, spec §5).
   Omitted on purpose: .auth-* / .back-link (in auth.css), .field* / .form-row* (in pages.css),
   .timeline / .tl-* (no fulfilment backend), .loyalty-*, .star-picker (no backends),
   .legal-* (tranche 5), sidebar badges (no real counts). */

.acct-layout { min-height: 100vh; display: grid; grid-template-columns: 260px 1fr; background: var(--cream); }
@media (max-width: 900px) { .acct-layout { grid-template-columns: 1fr; } }
.acct-side { background: var(--surface); padding: 32px 24px; border-right: 1px solid var(--line); position: sticky; top: 0; height: 100vh; overflow-y: auto; }
@media (max-width: 900px) { .acct-side { position: static; height: auto; } }
.acct-side-brand { margin-bottom: 32px; }
.acct-side h5 { font-family: var(--font-label); font-size: 10px; font-weight: 700; letter-spacing: 0.2em; text-transform: uppercase; color: var(--ink-muted); margin: 20px 0 10px; }
.acct-side a { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; border-radius: 8px; font-family: var(--font-label); font-size: 14px; color: var(--ink-soft); cursor: pointer; transition: all var(--dur) var(--ease); }
.acct-side a:hover { background: var(--lavender-100); color: var(--ink); }
.acct-side a.active { background: var(--ink); color: white; }
.acct-me { background: white; border-radius: 12px; padding: 16px; margin-bottom: 24px; display: flex; align-items: center; gap: 12px; }
.acct-avatar { width: 44px; height: 44px; border-radius: 999px; background: var(--lavender-300); color: var(--ink); display: grid; place-items: center; font-family: var(--font-serif); font-weight: 500; font-size: 18px; }
.acct-me-name { font-family: var(--font-serif); font-size: 17px; font-style: italic; }
.acct-me-tier { font-family: var(--font-label); font-size: 11px; letter-spacing: 0.12em; text-transform: uppercase; color: var(--lavender-700); overflow-wrap: anywhere; }

.acct-main { padding: 48px clamp(24px, 4vw, 64px); max-width: 1200px; }
.acct-head { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 40px; gap: 20px; flex-wrap: wrap; }
.acct-head h1 { font-family: var(--font-display); font-size: clamp(36px, 5vw, 64px); margin: 8px 0 0; letter-spacing: 0.005em; font-weight: 400; }
.acct-head .eyebrow { color: var(--lavender-700); }

.acct-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: 40px; }
@media (max-width: 720px) { .acct-cards { grid-template-columns: 1fr; } }
.acct-card { background: white; border: 1px solid var(--line); border-radius: 12px; padding: 24px; }
.acct-card-k { font-family: var(--font-label); font-size: 10px; letter-spacing: 0.18em; text-transform: uppercase; color: var(--ink-muted); margin-bottom: 10px; }
.acct-card-v { font-family: var(--font-serif); font-weight: 400; font-size: 42px; letter-spacing: -0.01em; margin-bottom: 6px; color: var(--ink); }
.acct-card-sub { font-family: var(--font-label); font-size: 12px; color: var(--ink-muted); }
.acct-card-accent { background: var(--ink); color: white; border-color: var(--ink); }
.acct-card-accent .acct-card-k { color: var(--lavender-300); }
.acct-card-accent .acct-card-v { color: var(--lavender-300); font-style: italic; }
.acct-card-accent .acct-card-sub { color: rgba(255,255,255,0.7); }

.acct-section { margin-bottom: 48px; }
.acct-section-head { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 20px; }
.acct-section h2 { font-family: var(--font-display); font-weight: 400; font-size: clamp(22px, 2.5vw, 32px); margin: 0; }

/* orders table — grid adapted: mockup's Items column dropped (list API has no item count) */
.orders-table { background: white; border: 1px solid var(--line); border-radius: 12px; overflow: hidden; }
.orders-row { display: grid; grid-template-columns: 1.2fr 1fr 1fr 1fr 100px; gap: 20px; padding: 16px 24px; align-items: center; border-bottom: 1px solid var(--line-soft); font-family: var(--font-label); font-size: 13px; }
.orders-row:last-child { border-bottom: 0; }
.orders-row.head { background: var(--lavender-50); font-size: 10px; font-weight: 700; letter-spacing: 0.14em; text-transform: uppercase; color: var(--ink-muted); }
.orders-row .o-id { font-weight: 600; color: var(--ink); }
.orders-row .price { font-weight: 600; }
.status { display: inline-flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 999px; font-size: 11px; font-weight: 600; letter-spacing: 0.04em; }
.status-dot { width: 6px; height: 6px; border-radius: 999px; background: currentColor; }
/* adapted/extrapolated: real order statuses mapped onto the mockup's pill palette */
.status-paid { background: #E7F4EC; color: #2B7A4A; }
.status-pending { background: #FFF4D6; color: #8A6500; }
.status-failed { background: #FBE5E5; color: #9E3535; }
.status-cancelled { background: #EEE; color: var(--ink-muted); }
.orders-row .link-btn { font-size: 11px; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; color: var(--ink); border-bottom: 1px solid var(--ink); padding-bottom: 2px; cursor: pointer; }
@media (max-width: 900px) { .orders-row { grid-template-columns: 1fr 1fr; } .orders-row.head { display: none; } }

/* Forms */
.form-card { background: white; border: 1px solid var(--line); border-radius: 12px; padding: 32px; max-width: 720px; }

/* address cards */
.addr-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
@media (max-width: 720px) { .addr-grid { grid-template-columns: 1fr; } }
.addr-card { background: white; border: 1px solid var(--line); border-radius: 12px; padding: 24px; position: relative; }
.addr-card.default { border-color: var(--lavender-600); }
.addr-card .pill { position: absolute; top: 16px; right: 16px; background: var(--lavender-100); color: var(--lavender-700); padding: 4px 10px; border-radius: 999px; font-family: var(--font-label); font-size: 10px; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; }
.addr-card h4 { font-family: var(--font-serif); font-style: italic; font-size: 20px; margin: 0 0 4px; font-weight: 400; }
.addr-card p { font-family: var(--font-body); font-size: 14px; color: var(--ink-soft); line-height: 1.6; margin: 0; }
.addr-card .actions { display: flex; gap: 12px; margin-top: 16px; }
.addr-card .actions button { font-family: var(--font-label); font-size: 11px; font-weight: 600; letter-spacing: 0.1em; text-transform: uppercase; color: var(--ink); border-bottom: 1px solid var(--ink); padding-bottom: 2px; cursor: pointer; }
.addr-card .actions button.danger { color: #9E3535; border-color: #9E3535; }

/* Page tabs (sub-nav) */
.sub-tabs { display: flex; gap: 2px; background: var(--lavender-50); padding: 4px; border-radius: 999px; margin-bottom: 24px; }
.sub-tabs button { padding: 10px 18px; border-radius: 999px; font-family: var(--font-label); font-size: 12px; font-weight: 600; letter-spacing: 0.06em; color: var(--ink-soft); }
.sub-tabs button.active { background: white; color: var(--ink); box-shadow: 0 1px 3px rgba(0,0,0,0.06); }

.dash-link-card { background: white; border: 1px solid var(--line); border-radius: 12px; padding: 20px; display: flex; justify-content: space-between; align-items: center; cursor: pointer; transition: all var(--dur) var(--ease); }
.dash-link-card:hover { border-color: var(--lavender-400); transform: translateX(2px); }
.dash-link-card-title { font-family: var(--font-serif); font-size: 18px; font-style: italic; font-weight: 400; }
.dash-link-card-sub { font-family: var(--font-label); font-size: 12px; color: var(--ink-muted); margin-top: 4px; }
.dash-links { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
@media (max-width: 720px) { .dash-links { grid-template-columns: 1fr; } }

/* alerts */
.alert { padding: 14px 18px; border-radius: 10px; display: flex; gap: 12px; align-items: flex-start; font-family: var(--font-label); font-size: 13px; margin-bottom: 20px; }
.alert-info { background: var(--lavender-100); color: var(--lavender-800); }
.alert-success { background: #E7F4EC; color: #2B7A4A; }
.alert-warn { background: #FFF4D6; color: #8A6500; }

/* adapted/extrapolated — inline styles from the mockup lifted into real classes */
.acct-side-link-label { display: inline-flex; align-items: center; gap: 10px; }
.acct-side-foot { margin-top: 24px; padding-top: 20px; border-top: 1px solid var(--line); }
.acct-side-foot a { color: var(--ink-muted); }
.acct-loading { padding: 48px; font-family: var(--font-label); font-size: 13px; color: var(--ink-muted); }
.order-detail-grid { display: grid; grid-template-columns: 2fr 1fr; gap: 24px; }
@media (max-width: 900px) { .order-detail-grid { grid-template-columns: 1fr; } }
.form-card-title { font-family: var(--font-serif); font-style: italic; font-size: 22px; font-weight: 400; margin: 0 0 16px; }
.kv-row { display: flex; justify-content: space-between; font-family: var(--font-label); font-size: 13px; padding: 6px 0; color: var(--ink-soft); }
.kv-row strong { color: var(--ink); }
.kv-divider { border-top: 1px solid var(--line); margin: 12px 0; }
.kv-block { margin-top: 20px; font-family: var(--font-label); font-size: 12px; }
.kv-block-k { color: var(--ink-muted); margin-bottom: 6px; }
.od-item { display: flex; gap: 16px; padding: 16px 0; border-bottom: 1px solid var(--line-soft); }
.od-item:last-child { border-bottom: 0; }
.od-item-ph { width: 80px; height: 100px; flex-shrink: 0; border-radius: 4px; overflow: hidden; }
.od-item-ph img { width: 100%; height: 100%; object-fit: cover; }
.od-item-body { flex: 1; }
.od-item-name { font-family: var(--font-serif); font-style: italic; font-size: 18px; }
.od-item-meta { font-family: var(--font-label); font-size: 12px; color: var(--ink-muted); margin-top: 4px; }
.od-item .price { font-size: 16px; }
.acct-empty { text-align: center; padding: 48px 0; }
.acct-empty p { color: var(--ink-muted); font-family: var(--font-label); font-size: 13px; margin: 0 0 20px; }
.acct-placed { font-family: var(--font-label); font-size: 13px; color: var(--ink-muted); margin: -24px 0 32px; }
.acct-pagination { display: flex; justify-content: center; align-items: center; gap: 12px; margin-top: 24px; font-family: var(--font-label); font-size: 13px; color: var(--ink-muted); }
.acct-head-actions { display: flex; gap: 12px; align-items: center; }
```

- [ ] **Step 8: Import it from `globals.css`**

After the existing `@import './auth.css';` line add:

```css
@import './account.css';
```

- [ ] **Step 9: Verify**

Run: `pnpm typecheck && pnpm build`
Expected: both pass; `grep -c "acct-layout" dist/assets/*.css` ≥ 1; header brand now styled (spot-check `.brand-word` in dist CSS).

Dedup re-check (must all print 1, i.e. only account.css defines them):
`for c in acct-layout orders-table addr-card sub-tabs form-card alert-info status-dot; do echo -n "$c: "; grep -l "\.$c" src/styles/*.css | wc -l; done`

- [ ] **Step 10: Commit**

```bash
git add src/styles src/content/regions.ts src/lib/format
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): account.css port, missing .brand styles, regions const, order-date formatter"
```

---

### Task 2: Account layout + sidebar + router reparent + shared primitives

**Files:**
- Modify: `frontend/src/router.tsx` (reparent `accountRoute` from `storefrontLayoutRoute` to `rootRoute`; move it out of `storefrontLayoutRoute.addChildren` into the root children list)
- Rewrite: `frontend/src/features/account/account-layout.tsx`
- Create: `frontend/src/features/account/acct-primitives.tsx`
- Modify: `frontend/src/features/account/account-order-detail.tsx` (ONLY the `useParams` `from` string — full rewrite comes in Task 4)

**Interfaces:**
- Consumes: `useAuth()` → `{ user, isLoading, logout }` (`logout(): Promise<void>` — note: this is its FIRST UI call site in the app); `Brand` from `features/shared/ui/brand`; `Icon` from `features/shared/ui/icons`.
- Produces: `AcctHead({ eyebrow, title, children? })` and `StatusPill({ status })` from `./acct-primitives` — Tasks 3–7 import these exact names.
- Produces: route ids change from `/_storefront/account/...` to `/account/...` — order-detail's `useParams({ from: ... })` must become `'/account/orders/$id'`.

- [ ] **Step 1: Reparent the account routes in `router.tsx`**

Change `accountRoute`'s parent (currently `storefrontLayoutRoute` at ~line 118) and update the tree assembly:

```tsx
// ── Account: authenticated customers, standalone chrome (mockup acct-layout) ─
const accountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/account',
  beforeLoad: () => requireRole('authenticated'),
  component: AccountLayout,
});
```

Child route declarations (`accountDashboardRoute` … `accountSettingsRoute`) are unchanged. In the `routeTree` assembly, move `accountRoute.addChildren([...])` out of `storefrontLayoutRoute.addChildren([...])` to the top level, next to `authLayoutRoute.addChildren([...])`. (Spec §3 asked for a pathless `_account` layout; since every account page shares the `/account` path prefix, reparenting `accountRoute` itself is equivalent and smaller — the standalone-chrome intent is what's binding.)

- [ ] **Step 2: Update order-detail's `useParams` source**

In `account-order-detail.tsx` line 32: `useParams({ from: '/_storefront/account/orders/$id' })` → `useParams({ from: '/account/orders/$id' })`.

- [ ] **Step 3: Create `features/account/acct-primitives.tsx`**

```tsx
import type { ReactNode } from 'react';

// Ported from Rue/src/acct-pages.jsx (AcctHead, StatusPill)

export function AcctHead({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string;
  title: string;
  children?: ReactNode;
}) {
  return (
    <div className="acct-head">
      <div>
        <div className="eyebrow">{eyebrow}</div>
        <h1>{title}</h1>
      </div>
      {children}
    </div>
  );
}

// Real order statuses (pending/paid/failed/cancelled) on the mockup's pill palette.
const KNOWN_STATUSES = new Set(['pending', 'paid', 'failed', 'cancelled']);

export function StatusPill({ status }: { status: string | undefined }) {
  if (!status) return null;
  const cls = KNOWN_STATUSES.has(status) ? ` status-${status}` : '';
  return (
    <span className={`status${cls}`}>
      <span className="status-dot"></span>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}
```

- [ ] **Step 4: Rewrite `account-layout.tsx`**

```tsx
import { Link, Outlet, useNavigate } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Brand } from '../shared/ui/brand';
import { Icon } from '../shared/ui/icons';

// Ported from Rue/src/acct-pages.jsx (AcctSidebar) — real links only (spec §2.2):
// the mockup's tracking/subscriptions/returns/reviews/reorder/loyalty/referral/payments
// entries have no backend and are omitted.

export function AccountLayout() {
  const { user, isLoading, logout } = useAuth();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="acct-layout">
        <div className="acct-loading">Loading…</div>
      </div>
    );
  }

  const displayName = user?.name || user?.email?.split('@')[0] || 'Member';
  const initial = displayName.charAt(0).toUpperCase();

  const handleSignOut = async () => {
    await logout();
    navigate({ to: '/' });
  };

  return (
    <div className="acct-layout">
      <aside className="acct-side">
        <div className="acct-side-brand">
          <Link to="/" aria-label="Back to Rue home">
            <Brand />
          </Link>
        </div>
        <div className="acct-me">
          <div className="acct-avatar">{initial}</div>
          <div>
            <div className="acct-me-name">{displayName}</div>
            <div className="acct-me-tier">{user?.email}</div>
          </div>
        </div>
        <h5>Shop</h5>
        <Link
          to="/account"
          activeOptions={{ exact: true }}
          activeProps={{ className: 'active' }}
        >
          <span className="acct-side-link-label">
            <Icon name="user" size={14} /> Overview
          </span>
        </Link>
        <Link to="/account/orders" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">
            <Icon name="bag" size={14} /> Orders
          </span>
        </Link>
        <Link to="/account/wishlist" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">
            <Icon name="heart" size={14} /> Wishlist
          </span>
        </Link>
        <h5>Account</h5>
        <Link to="/account/addresses" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">Addresses</span>
        </Link>
        <Link to="/account/settings" activeProps={{ className: 'active' }}>
          <span className="acct-side-link-label">Settings</span>
        </Link>
        <div className="acct-side-foot">
          <a onClick={handleSignOut}>
            <span className="acct-side-link-label">
              <Icon name="arrowLeft" size={14} /> Sign out
            </span>
          </a>
        </div>
      </aside>
      <Outlet />
    </div>
  );
}
```

Notes for the implementer: the sidebar brand `<Link>` picks up `.acct-side a` padding/hover — if that visibly distorts the brand block, add `.acct-side-brand a { padding: 0; background: transparent; }` to the adapted section of `account.css` (the mockup does the same inline at `acct-pages.jsx:18`). The `<Outlet/>` is NOT wrapped in `<main>` (pages own `.acct-main`). While pages are still un-ported (Tasks 3–6 pending) they'll render without `.acct-main` padding — acceptable mid-plan state; do not add a wrapper.

- [ ] **Step 5: Typecheck + route smoke**

Run: `pnpm typecheck`
Expected: zero errors (this catches any stale `/_storefront/account/...` route references project-wide — fix any it reports the same way as Step 2).
Then: `pnpm vitest run` — the existing guard tests must still pass (guard moved with the route).

- [ ] **Step 6: Commit**

```bash
git add src/router.tsx src/features/account/account-layout.tsx src/features/account/acct-primitives.tsx src/features/account/account-order-detail.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): standalone account chrome — acct-layout sidebar, route reparent, shared AcctHead/StatusPill"
```

---

### Task 3: Dashboard

**Files:**
- Rewrite: `frontend/src/features/account/account-dashboard.tsx`

**Interfaces:**
- Consumes: `AcctHead`, `StatusPill` from `./acct-primitives`; `getMeOrders({ limit: 3 })` → `{ orders?: [{ id, status, total_ghs, created_at }], total?: number }`; `getMeAddresses()` → `{ addresses?: [{ label?, is_default? }] }`; `formatGhs`, `formatOrderDate` from `../../lib/format/utils`.

- [ ] **Step 1: Rewrite `account-dashboard.tsx`**

```tsx
import { useEffect, useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { getMeAddresses, getMeOrders } from '../../lib/api/generated/rueCosmeticsAPI';
import { useAuth } from '../../lib/auth/auth-provider';
import { formatGhs, formatOrderDate } from '../../lib/format/utils';
import { Icon } from '../shared/ui/icons';
import { AcctHead, StatusPill } from './acct-primitives';

type RecentOrder = {
  id?: string;
  status?: string;
  total_ghs?: number;
  created_at?: string;
};

type AddressSummary = { label?: string; is_default?: boolean };

export function AccountDashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [recent, setRecent] = useState<RecentOrder[]>([]);
  const [ordersTotal, setOrdersTotal] = useState<number | null>(null);
  const [addresses, setAddresses] = useState<AddressSummary[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    getMeOrders({ limit: 3 })
      .then((res) => {
        if (cancelled) return;
        setRecent(res.orders || []);
        setOrdersTotal(res.total ?? 0);
      })
      .catch(() => {
        if (!cancelled) setOrdersTotal(null);
      });
    getMeAddresses()
      .then((res) => {
        if (!cancelled) setAddresses(res.addresses || []);
      })
      .catch(() => {
        if (!cancelled) setAddresses(null);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const firstName = user?.name?.trim().split(/\s+/)[0];
  const defaultAddress = addresses?.find((a) => a.is_default);

  return (
    <main className="acct-main">
      <AcctHead
        eyebrow="Welcome back"
        title={firstName ? `Hello, ${firstName}.` : 'Hello.'}
      />

      <div className="acct-cards">
        <div className="acct-card">
          <div className="acct-card-k">Lifetime orders</div>
          <div className="acct-card-v">{ordersTotal ?? '—'}</div>
          <div className="acct-card-sub">All time</div>
        </div>
        <div className="acct-card">
          <div className="acct-card-k">Saved addresses</div>
          <div className="acct-card-v">{addresses ? addresses.length : '—'}</div>
          <div className="acct-card-sub">
            {defaultAddress?.label
              ? `${defaultAddress.label} set as default`
              : 'No default set'}
          </div>
        </div>
        <div className="acct-card acct-card-accent">
          <div className="acct-card-k">Member</div>
          <div className="acct-card-v">{firstName || 'Member'}</div>
          <div className="acct-card-sub">
            {user?.email} · {user?.email_verified ? 'Verified' : 'Unverified'}
          </div>
        </div>
      </div>

      <div className="acct-section">
        <div className="acct-section-head">
          <h2>Recent orders</h2>
          <Link className="auth-link" to="/account/orders">
            View all →
          </Link>
        </div>
        {recent.length === 0 ? (
          <div className="acct-empty">
            <p>No orders yet — when you place one, it will appear here.</p>
            <Link className="btn btn-primary" to="/shop">
              Start shopping
            </Link>
          </div>
        ) : (
          <div className="orders-table">
            <div className="orders-row head">
              <div>Order</div>
              <div>Date</div>
              <div>Total</div>
              <div>Status</div>
              <div></div>
            </div>
            {recent.map((o) => (
              <div key={o.id} className="orders-row">
                <div className="o-id">
                  #{(o.id || '').slice(0, 8).toUpperCase()}
                </div>
                <div>{formatOrderDate(o.created_at)}</div>
                <div className="price">{formatGhs(o.total_ghs || 0)}</div>
                <div>
                  <StatusPill status={o.status} />
                </div>
                <div>
                  <Link
                    className="link-btn"
                    to="/account/orders/$id"
                    params={{ id: o.id || '' }}
                  >
                    View
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="acct-section">
        <div className="acct-section-head">
          <h2>Quick links</h2>
        </div>
        <div className="dash-links">
          {[
            {
              to: '/account/orders',
              title: 'Orders',
              sub:
                ordersTotal !== null
                  ? `${ordersTotal} order${ordersTotal === 1 ? '' : 's'} placed`
                  : 'Your order history',
            },
            {
              to: '/account/addresses',
              title: 'Saved addresses',
              sub:
                addresses !== null
                  ? `${addresses.length} saved address${addresses.length === 1 ? '' : 'es'}`
                  : 'Manage delivery addresses',
            },
            {
              to: '/account/wishlist',
              title: 'Wishlist',
              sub: 'Saved items — coming soon',
            },
          ].map((l) => (
            <div
              key={l.to}
              className="dash-link-card"
              onClick={() => navigate({ to: l.to })}
            >
              <div>
                <div className="dash-link-card-title">{l.title}</div>
                <div className="dash-link-card-sub">{l.sub}</div>
              </div>
              <Icon name="arrow" size={16} />
            </div>
          ))}
        </div>
      </div>
    </main>
  );
}
```

(Fetch-failure behavior: card values show `—` and quick-link subs fall back to static copy — no fake zeros. The `Button` import is gone; the page uses mockup `btn` classes.)

- [ ] **Step 2: Typecheck**

Run: `pnpm typecheck` — Expected: zero errors.

- [ ] **Step 3: Visual + functional smoke (requires backend `make dev` + `pnpm dev`, logged-in session)**

`/account`: sidebar active state on Overview only; three cards show real total/count/member info; recent orders match `/account/orders`; quick links navigate. If Docker/backend is unavailable, note it in the ledger for the human walkthrough instead.

- [ ] **Step 4: Commit**

```bash
git add src/features/account/account-dashboard.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): account dashboard aligned to mockup — real KPI cards, recent orders, quick links"
```

---

### Task 4: Orders list + order detail

**Files:**
- Rewrite: `frontend/src/features/account/account-orders.tsx`
- Rewrite: `frontend/src/features/account/account-order-detail.tsx`

**Interfaces:**
- Consumes: `AcctHead`, `StatusPill`; `getMeOrders({ limit, offset, status })`; `getMeOrdersId(id)` → detail with `items[].{product_name_snapshot, product_brand_snapshot, product_image_snapshot, qty, unit_price_ghs}`, `shipping_address`, `paystack_reference`; `formatGhs`, `formatOrderDate`, `getImageUrl`.

- [ ] **Step 1: Rewrite `account-orders.tsx`** — keep ALL existing state/fetch logic (`orders/total/page/statusFilter/isLoading/error`, `loadOrders`, the `useEffect`, `limit = 10`); replace only the rendered JSX and drop the local `formatDate`/`formatCurrency`/`getStatusColor` helpers and the `Button` import.

```tsx
import { useEffect, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { getMeOrders } from '../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate } from '../../lib/format/utils';
import { AcctHead, StatusPill } from './acct-primitives';

type Order = {
  id?: string;
  user_id?: string;
  status?: string;
  subtotal_ghs?: number;
  shipping_ghs?: number;
  total_ghs?: number;
  paystack_reference?: string;
  created_at?: string;
  updated_at?: string;
};

const STATUS_TABS: Array<[value: string, label: string]> = [
  ['', 'All'],
  ['pending', 'Pending'],
  ['paid', 'Paid'],
  ['failed', 'Failed'],
  ['cancelled', 'Cancelled'],
];

export function AccountOrders() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [statusFilter, setStatusFilter] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const limit = 10;
  const offset = page * limit;

  const loadOrders = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getMeOrders({
        limit,
        offset,
        status: statusFilter || undefined,
      });
      setOrders(response.orders || []);
      setTotal(response.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load orders');
    } finally {
      setIsLoading(false);
    }
  };

  // Load orders on mount and when filter/page changes
  useEffect(() => {
    loadOrders();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, statusFilter]);

  const totalPages = Math.ceil(total / limit);

  return (
    <main className="acct-main">
      <AcctHead eyebrow="History" title="Your orders" />

      <div className="sub-tabs" style={{ width: 'fit-content' }}>
        {STATUS_TABS.map(([value, label]) => (
          <button
            key={label}
            className={statusFilter === value ? 'active' : ''}
            onClick={() => {
              setStatusFilter(value);
              setPage(0);
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {error && <div className="alert alert-warn">{error}</div>}

      {isLoading ? (
        <div className="acct-empty">
          <p>Loading orders…</p>
        </div>
      ) : orders.length === 0 ? (
        <div className="acct-empty">
          <p>
            {statusFilter
              ? 'No orders with this status.'
              : 'No orders yet — when you place one, it will appear here.'}
          </p>
          <Link className="btn btn-primary" to="/shop">
            Start shopping
          </Link>
        </div>
      ) : (
        <>
          <div className="orders-table">
            <div className="orders-row head">
              <div>Order</div>
              <div>Date</div>
              <div>Total</div>
              <div>Status</div>
              <div></div>
            </div>
            {orders.map((order) => (
              <div key={order.id} className="orders-row">
                <div className="o-id">
                  #{(order.id || '').slice(0, 8).toUpperCase()}
                </div>
                <div>{formatOrderDate(order.created_at)}</div>
                <div className="price">{formatGhs(order.total_ghs || 0)}</div>
                <div>
                  <StatusPill status={order.status} />
                </div>
                <div>
                  <Link
                    className="link-btn"
                    to="/account/orders/$id"
                    params={{ id: order.id || '' }}
                  >
                    View
                  </Link>
                </div>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="acct-pagination">
              <button
                className="btn btn-ghost"
                disabled={page === 0}
                onClick={() => setPage(page - 1)}
              >
                Previous
              </button>
              <span>
                Page {page + 1} of {totalPages}
              </span>
              <button
                className="btn btn-ghost"
                disabled={page >= totalPages - 1}
                onClick={() => setPage(page + 1)}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </main>
  );
}
```

(The `style={{ width: 'fit-content' }}` on `.sub-tabs` is the mockup's own inline style at `acct-pages.jsx:152` — kept verbatim. The eslint-disable on the effect deps matches the pre-existing pattern flagged in the audit-remediation backlog.)

- [ ] **Step 2: Rewrite `account-order-detail.tsx`** — keep the existing fetch effect/types; replace the JSX. Remove the fake `handleReorder` stub entirely.

```tsx
import { useEffect, useState } from 'react';
import { Link, useParams } from '@tanstack/react-router';
import { getMeOrdersId } from '../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate, getImageUrl } from '../../lib/format/utils';
import { Icon } from '../shared/ui/icons';
import { AcctHead, StatusPill } from './acct-primitives';

type OrderItem = {
  id?: string;
  order_id?: string;
  product_id?: string;
  qty?: number;
  unit_price_ghs?: number;
  product_name_snapshot?: string;
  product_brand_snapshot?: string;
  product_image_snapshot?: string;
};

type OrderDetailResponse = {
  id?: string;
  user_id?: string;
  status?: string;
  subtotal_ghs?: number;
  shipping_ghs?: number;
  total_ghs?: number;
  paystack_reference?: string;
  shipping_address?: string;
  created_at?: string;
  updated_at?: string;
  items?: OrderItem[];
};

export function AccountOrderDetail() {
  const { id } = useParams({ from: '/account/orders/$id' });
  const [order, setOrder] = useState<OrderDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadOrder = async () => {
      if (!id) return;
      setIsLoading(true);
      setError(null);
      try {
        const response = await getMeOrdersId(id);
        setOrder(response);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load order details');
      } finally {
        setIsLoading(false);
      }
    };
    loadOrder();
  }, [id]);

  if (isLoading) {
    return (
      <main className="acct-main">
        <div className="acct-empty">
          <p>Loading order details…</p>
        </div>
      </main>
    );
  }

  if (error || !order) {
    return (
      <main className="acct-main">
        <Link className="back-link" to="/account/orders">
          <Icon name="arrowLeft" size={12} /> All orders
        </Link>
        <div className="alert alert-warn">
          {error || 'Unable to load order details'}
        </div>
      </main>
    );
  }

  const shortId = (order.id || '').slice(0, 8).toUpperCase();

  return (
    <main className="acct-main">
      <Link className="back-link" to="/account/orders">
        <Icon name="arrowLeft" size={12} /> All orders
      </Link>
      <AcctHead eyebrow={`Order #${shortId}`} title="Order details">
        <div className="acct-head-actions">
          <StatusPill status={order.status} />
        </div>
      </AcctHead>
      <div className="acct-placed">
        Placed {formatOrderDate(order.created_at)}
      </div>

      <div className="order-detail-grid">
        <div>
          <div className="form-card" style={{ maxWidth: 'none' }}>
            <h3 className="form-card-title">
              Items ({order.items?.length || 0})
            </h3>
            {(order.items || []).map((it) => (
              <div key={it.id} className="od-item">
                <div className="od-item-ph ph ph--lavender">
                  {it.product_image_snapshot ? (
                    <img
                      src={getImageUrl(it.product_image_snapshot)}
                      alt={it.product_name_snapshot || ''}
                    />
                  ) : (
                    <span className="ph-label">
                      {it.product_name_snapshot}
                    </span>
                  )}
                </div>
                <div className="od-item-body">
                  <div className="od-item-name">
                    {it.product_name_snapshot}
                  </div>
                  {it.product_brand_snapshot && (
                    <div className="od-item-meta">
                      {it.product_brand_snapshot}
                    </div>
                  )}
                  <div className="od-item-meta">Qty {it.qty}</div>
                </div>
                <div className="price">
                  {formatGhs((it.unit_price_ghs || 0) * (it.qty || 0))}
                </div>
              </div>
            ))}
          </div>
        </div>
        <div>
          <div className="form-card" style={{ padding: 24 }}>
            <h3 className="form-card-title" style={{ fontSize: 18 }}>
              Summary
            </h3>
            <div className="kv-row">
              <span>Subtotal</span>
              <span>{formatGhs(order.subtotal_ghs || 0)}</span>
            </div>
            <div className="kv-row">
              <span>Shipping</span>
              <span>{formatGhs(order.shipping_ghs || 0)}</span>
            </div>
            <div className="kv-divider"></div>
            <div className="kv-row">
              <strong>Total</strong>
              <strong className="price">
                {formatGhs(order.total_ghs || 0)}
              </strong>
            </div>
            {order.shipping_address && (
              <div className="kv-block">
                <div className="kv-block-k">Delivering to</div>
                <div style={{ whiteSpace: 'pre-line' }}>
                  {order.shipping_address}
                </div>
              </div>
            )}
            {order.paystack_reference && (
              <div className="kv-block">
                <div className="kv-block-k">Payment</div>
                <div>Paystack · {order.paystack_reference}</div>
              </div>
            )}
          </div>
          <div style={{ marginTop: 16 }}>
            <Link className="btn btn-ghost" to="/shop">
              Continue shopping
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
```

(The three `style={{ … }}` props mirror the mockup's own inline styles at `acct-pages.jsx:209/227` (`maxWidth: 'none'`, `padding: 24`, summary title size) — kept inline like the mockup rather than minted as classes. No timeline, no return/track buttons, no reorder stub — all backend-less.)

- [ ] **Step 3: Typecheck**

Run: `pnpm typecheck` — Expected: zero errors.

- [ ] **Step 4: Functional smoke (with dev servers, logged in, an order placed)**

Orders list: tabs re-query with correct `status` param (watch the network panel); pagination works past 10 orders (or skip if fewer). Detail: real items with images, totals match, Paystack reference shown, back-link returns. Note in ledger if backend unavailable.

- [ ] **Step 5: Commit**

```bash
git add src/features/account/account-orders.tsx src/features/account/account-order-detail.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): orders list + order detail aligned to mockup — sub-tabs, orders-table, honest detail summary"
```

---

### Task 5: Addresses

**Files:**
- Rewrite: `frontend/src/features/account/account-addresses.tsx`

**Interfaces:**
- Consumes: `AcctHead`; `GHANA_REGIONS` from `../../content/regions`; existing generated calls (`getMeAddresses`, `postMeAddresses`, `patchMeAddressesId`, `deleteMeAddressesId`, `postMeAddressesIdDefault`); `Icon` (`plus`).

- [ ] **Step 1: Rewrite `account-addresses.tsx`** — keep ALL existing handlers (`loadAddresses`, `handleSetDefault`, `handleDelete` incl. its `confirm()`, the form's validation/submit flow, `AddressForm`'s controlled state). Two intentional behavior fixes, both to record in the ledger: (a) the Edit button now also opens the form (`setShowForm(true)` — previously it set `editingAddress` but nothing appeared: live bug); (b) Region becomes a `<select>` over `GHANA_REGIONS` (spec §2.5; same string submitted).

```tsx
import { useEffect, useState } from 'react';
import {
  deleteMeAddressesId,
  getMeAddresses,
  patchMeAddressesId,
  postMeAddresses,
  postMeAddressesIdDefault,
} from '../../lib/api/generated/rueCosmeticsAPI';
import { GHANA_REGIONS } from '../../content/regions';
import { Icon } from '../shared/ui/icons';
import { AcctHead } from './acct-primitives';

type Address = {
  id?: string;
  label?: string;
  line1?: string;
  line2?: string;
  city?: string;
  region?: string;
  phone?: string;
  is_default?: boolean;
  created_at?: string;
  updated_at?: string;
};

type AddressFormData = {
  label: string;
  line1: string;
  line2: string;
  city: string;
  region: string;
  phone: string;
};

export function AccountAddresses() {
  const [addresses, setAddresses] = useState<Address[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingAddress, setEditingAddress] = useState<Address | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadAddresses = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getMeAddresses();
      setAddresses(response.addresses || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load addresses');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadAddresses();
  }, []);

  const handleSetDefault = async (id: string) => {
    if (!id) return;
    try {
      await postMeAddressesIdDefault(id);
      await loadAddresses();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to set default address');
    }
  };

  const handleDelete = async (id: string) => {
    if (!id || !confirm('Are you sure you want to delete this address?')) return;
    try {
      await deleteMeAddressesId(id);
      await loadAddresses();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete address');
    }
  };

  if (isLoading) {
    return (
      <main className="acct-main">
        <div className="acct-empty">
          <p>Loading addresses…</p>
        </div>
      </main>
    );
  }

  return (
    <main className="acct-main">
      <AcctHead eyebrow="Delivery" title="Address book">
        <button className="btn btn-primary" onClick={() => setShowForm(true)}>
          <Icon name="plus" size={14} /> Add address
        </button>
      </AcctHead>

      {error && <div className="alert alert-warn">{error}</div>}

      {showForm && (
        <AddressForm
          address={editingAddress}
          onSubmit={async (data) => {
            try {
              if (editingAddress && editingAddress.id) {
                await patchMeAddressesId(editingAddress.id, data);
              } else {
                await postMeAddresses(data);
              }
              setShowForm(false);
              setEditingAddress(null);
              await loadAddresses();
            } catch (err) {
              setError(err instanceof Error ? err.message : 'Failed to save address');
            }
          }}
          onCancel={() => {
            setShowForm(false);
            setEditingAddress(null);
          }}
        />
      )}

      {addresses.length === 0 && !showForm ? (
        <div className="acct-empty">
          <p>No addresses yet — add one to make checkout easier.</p>
          <button className="btn btn-primary" onClick={() => setShowForm(true)}>
            Add your first address
          </button>
        </div>
      ) : (
        <div className="addr-grid">
          {addresses.map((a) => (
            <div key={a.id} className={`addr-card ${a.is_default ? 'default' : ''}`}>
              {a.is_default && <span className="pill">Default</span>}
              <h4>{a.label || 'Address'}</h4>
              <p>
                {a.line1}
                {a.line2 && (
                  <>
                    <br />
                    {a.line2}
                  </>
                )}
                <br />
                {[a.city, a.region].filter(Boolean).join(', ')}
                <br />
                {a.phone}
              </p>
              <div className="actions">
                <button
                  onClick={() => {
                    setEditingAddress(a);
                    setShowForm(true);
                  }}
                >
                  Edit
                </button>
                {!a.is_default && (
                  <button onClick={() => a.id && handleSetDefault(a.id)}>
                    Set default
                  </button>
                )}
                <button className="danger" onClick={() => a.id && handleDelete(a.id)}>
                  Remove
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </main>
  );
}

function AddressForm({
  address,
  onSubmit,
  onCancel,
}: {
  address: Address | null;
  onSubmit: (data: AddressFormData) => Promise<void>;
  onCancel: () => void;
}) {
  const [label, setLabel] = useState(address?.label || '');
  const [line1, setLine1] = useState(address?.line1 || '');
  const [line2, setLine2] = useState(address?.line2 || '');
  const [city, setCity] = useState(address?.city || '');
  const [region, setRegion] = useState(address?.region || '');
  const [phone, setPhone] = useState(address?.phone || '');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});
    setIsSubmitting(true);

    const newErrors: Record<string, string> = {};
    if (!label.trim()) newErrors.label = 'Label is required';
    if (!line1.trim()) newErrors.line1 = 'Street address is required';
    if (!city.trim()) newErrors.city = 'City is required';
    if (!region.trim()) newErrors.region = 'Region is required';
    if (!phone.trim()) newErrors.phone = 'Phone is required';

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      setIsSubmitting(false);
      return;
    }

    try {
      await onSubmit({ label, line1, line2, city, region, phone });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="form-card" style={{ marginBottom: 24, maxWidth: 'none' }}>
      <h3 className="form-card-title">
        {address ? 'Edit address' : 'New address'}
      </h3>
      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="field">
            <label>Label</label>
            <input
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="e.g. Home"
            />
            {errors.label && <span className="field-error">{errors.label}</span>}
          </div>
          <div className="field">
            <label>Phone</label>
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+233 24 000 0000"
            />
            {errors.phone && <span className="field-error">{errors.phone}</span>}
          </div>
        </div>
        <div className="form-row full">
          <div className="field">
            <label>Street address</label>
            <input
              value={line1}
              onChange={(e) => setLine1(e.target.value)}
              placeholder="14 Amilcar Cabral Ave"
            />
            {errors.line1 && <span className="field-error">{errors.line1}</span>}
          </div>
        </div>
        <div className="form-row full">
          <div className="field">
            <label>Apartment, suite, etc. (optional)</label>
            <input value={line2} onChange={(e) => setLine2(e.target.value)} />
          </div>
        </div>
        <div className="form-row">
          <div className="field">
            <label>City</label>
            <input
              value={city}
              onChange={(e) => setCity(e.target.value)}
              placeholder="Accra"
            />
            {errors.city && <span className="field-error">{errors.city}</span>}
          </div>
          <div className="field">
            <label>Region</label>
            <select value={region} onChange={(e) => setRegion(e.target.value)}>
              <option value="">Select a region…</option>
              {GHANA_REGIONS.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
            {errors.region && <span className="field-error">{errors.region}</span>}
          </div>
        </div>
        <div className="acct-head-actions">
          <button className="btn btn-primary" type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Saving…' : address ? 'Update address' : 'Save address'}
          </button>
          <button className="btn btn-ghost" type="button" onClick={onCancel} disabled={isSubmitting}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
```

(Note: the mockup's "Full name" field is omitted — the API has no such field. `AddressFormData` replaces the old `any`, resolving the audit-remediation minor. Edit-with-existing-region-not-in-list: a legacy free-text region value not matching the list renders as the placeholder; the value is preserved until the user picks — acceptable, note in ledger. The `style={{ marginBottom: 24, maxWidth: 'none' }}` mirrors the mockup's inline style at `acct-pages.jsx:302`. `AddressForm` remounts per edit target because `showForm` toggles render — if switching Edit between two cards without closing shows stale values, add `key={editingAddress?.id ?? 'new'}` to `<AddressForm>`; verify during smoke.)

- [ ] **Step 2: Typecheck**

Run: `pnpm typecheck` — Expected: zero errors.

- [ ] **Step 3: Functional smoke**

Add (region via select) → card appears; Edit prefills and now actually opens the form; Set default moves the pill + border; Remove asks confirm then deletes; validation errors render as `.field-error`.

- [ ] **Step 4: Commit**

```bash
git add src/features/account/account-addresses.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): address book aligned to mockup — addr-cards, mockup form, 16-region select, edit-button fix"
```

---

### Task 6: Wishlist + Settings (honest static states)

**Files:**
- Rewrite: `frontend/src/features/account/account-wishlist.tsx`
- Rewrite: `frontend/src/features/account/account-settings.tsx`

**Interfaces:**
- Consumes: `AcctHead`; `useAuth()` → `user.{name,email,email_verified}`; `Link` (router).

- [ ] **Step 1: Rewrite `account-wishlist.tsx`** — the page becomes a static honest empty state; delete the dead fake grid, stub handlers, commented-out generated imports, and the `WishlistItem` type (wishlist backend is queued — tranche-1 spec §8.2).

```tsx
import { Link } from '@tanstack/react-router';
import { AcctHead } from './acct-primitives';

// Wishlist has no backend yet (tranche-1 spec §8.2) — honest empty state only.

export function AccountWishlist() {
  return (
    <main className="acct-main">
      <AcctHead eyebrow="Saved for later" title="Wishlist" />
      <div className="alert alert-info">
        Saving items isn't available yet — wishlist is coming soon.
      </div>
      <div className="acct-empty">
        <p>Nothing saved here yet. In the meantime, browse the edit.</p>
        <Link className="btn btn-primary" to="/shop">
          Explore products
        </Link>
      </div>
    </main>
  );
}
```

- [ ] **Step 2: Rewrite `account-settings.tsx`** — real fields only; alerts explain what's unavailable; no disabled form theater (spec §4.7).

```tsx
import { Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { AcctHead } from './acct-primitives';

export function AccountSettings() {
  const { user } = useAuth();

  return (
    <main className="acct-main">
      <AcctHead eyebrow="Settings" title="Profile" />

      <div className="form-card" style={{ marginBottom: 24 }}>
        <div className="alert alert-info">
          Profile editing isn't available yet.
        </div>
        <div className="form-row">
          <div className="field">
            <label>Name</label>
            <input value={user?.name || ''} readOnly placeholder="Not set" />
          </div>
          <div className="field">
            <label>Email</label>
            <input type="email" value={user?.email || ''} readOnly />
            <span className="field-hint">
              {user?.email_verified
                ? 'Verified'
                : 'Unverified — check your inbox for the verification email'}
              {' · '}changes require verification
            </span>
          </div>
        </div>
      </div>

      <div className="form-card" style={{ marginBottom: 24 }}>
        <h3 className="form-card-title">Password</h3>
        <div className="alert alert-info">
          Password changes aren't available from settings yet — use the email
          reset flow instead.
        </div>
        <Link className="btn btn-ghost" to="/forgot-password">
          Reset via email
        </Link>
      </div>

      <div className="form-card">
        <h3 className="form-card-title">Delete account</h3>
        <div className="alert alert-warn">
          Account deletion isn't available yet. Contact us if you need your
          data removed.
        </div>
      </div>
    </main>
  );
}
```

- [ ] **Step 3: Typecheck + render smoke**

Run: `pnpm typecheck` — zero errors. Both pages render inside the sidebar layout with correct active nav states.

- [ ] **Step 4: Commit**

```bash
git add src/features/account/account-wishlist.tsx src/features/account/account-settings.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): wishlist + settings honest states in mockup styling"
```

---

### Task 7: Final gate, class-coverage audit, ledger

**Files:**
- Modify: `.superpowers/sdd/progress.md` (append Tranche 2 section — repo root, not frontend/)
- Modify (if audit finds gaps): `frontend/src/styles/account.css`

- [ ] **Step 1: Tailwind-residue + class-coverage audit**

From `frontend/`:

```bash
# 1. No Tailwind utilities left in account files (expect no output):
grep -nE 'className="[^"]*\b(flex|grid|px-|py-|mb-|mt-|text-|bg-|rounded|w-full|space-y)' src/features/account/*.tsx

# 2. Every static className in account files exists in some stylesheet (expect no "missing" lines):
for f in src/features/account/*.tsx; do
  grep -o 'className="[^"]*"' "$f" | sed 's/className="//;s/"$//' | tr ' ' '\n' | sort -u | while read -r c; do
    [ -z "$c" ] && continue
    grep -qs -- "\.$c" src/styles/*.css || echo "$f: missing .$c"
  done
done
```

Heuristic caveats: template-literal classNames (`` `addr-card ${...}` ``, `` `status${cls}` ``) are not caught by check 2 — verify by hand that `addr-card default`, `status status-paid/-pending/-failed/-cancelled`, `sub-tabs button.active`, and sidebar `active` are all defined (they are in Task 1's CSS; confirm). `btn`/`btn-primary`/`btn-ghost`/`eyebrow`/`ph`/`ph-label`/`back-link`/`auth-link` live in globals/pages/auth CSS — the grep covers all of `src/styles/`. Fix any real gap by adding the class to `account.css` (adapted banner) and note it in the ledger.

- [ ] **Step 2: Full gate**

Run: `pnpm typecheck && pnpm lint && pnpm vitest run && pnpm build`
Expected: all green, lint zero warnings. If lint flags the `onClick` on the sidebar sign-out `<a>` (a11y rule), convert it to `<button className="…">` styled by adding `.acct-side-foot button { …same link styles… }` — check whether the ESLint config even has jsx-a11y first (it may not; don't add rules).

- [ ] **Step 3: Route regression sweep**

With dev servers up (`make dev` in ruecosmetics/, `pnpm dev` in frontend/):

```bash
for p in /account /account/orders /account/addresses /account/wishlist /account/settings; do
  curl -s -o /dev/null -w "%{http_code} $p\n" "http://localhost:5173$p"
done
```
Expected: 200s (SPA shell). Logged-out browser visit to `/account` must redirect to login (guard moved with the route — verify in browser or note for human walkthrough).

- [ ] **Step 4: Update the progress ledger**

Append to `.superpowers/sdd/progress.md` (repo root) a `# Tranche 2 — Account UI Alignment` section: branch, start commit, per-task commit ranges + review outcomes, deferred minors, the backend follow-ups (order-list item counts, profile update endpoint, checkout region select), and the two behavior fixes (address Edit-button bug, region select) plus the `.brand` CSS restoration.

- [ ] **Step 5: Commit**

```bash
git add ../.superpowers/sdd/progress.md src/styles/account.css 2>/dev/null
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "chore(frontend): tranche-2 final gate — class audit, ledger update"
```

(If the audit found nothing and only the ledger changed, commit just the ledger.)
