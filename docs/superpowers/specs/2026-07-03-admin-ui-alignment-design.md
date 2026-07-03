# Admin Suite UI Alignment — Design Spec (Tranche 3)

Date: 2026-07-03
Status: Approved direction (brainstorm answers 2026-07-03); spec committed for async user veto
Mockup sources: `Rue/src/admin.jsx`, `Rue/admin.css` (read-only reference outside the repo)
Foundation: tranche-1 spec §2/§3 decisions binding (pure mockup CSS, Tailwind dormant, class parity, dedup rule, functionality freeze, honest data, formatGhs, getImageUrl, shared icons). Tranche-2 precedents apply (scoped button-chrome resets, keyboard-actionable controls where cheap, `<button>` over `<a onClick>`).

## 1. Goal

Re-skin the admin suite (layout, dashboard, orders, order detail, products, customers, marketing, content, analytics, settings) with the mockup's real `admin-*` CSS. The existing pages already mirror the mockup's structure through shared `Panel`/`KPICard`/`StatusTag` components — but everything is styled in dead Tailwind utilities (Tailwind is dormant), so the admin renders nearly unstyled today. This tranche swaps the class vocabulary, not the architecture: every generated hook, mutation, guard, and route stays.

## 2. Decisions from brainstorming (binding)

1. **Mockup chrome + store link**: full `.admin-layout`/`.admin-side` dark sidebar (Rue brand-word + "Admin" badge, Overview/Commerce/Growth/System sections). The current light header and its "Back to Store" button are removed; a muted "← Storefront" link at the sidebar bottom (account-sidebar sign-out pattern, rendered as a `Link` to `/`) replaces it.
2. **No nav count pills**: the mockup's Products/Orders pills are omitted — no layout-level fetches, matching the account sidebar's no-badges precedent.
3. **Strip fake data to honest states**:
   - KPI deltas (`+12.4%` etc.) removed — `KPICard` loses its `delta`/`deltaDirection` props and call sites; cards show real values only.
   - Fake activity feed panel deleted (no events endpoint). `.admin-activity`/`.activity-*` CSS NOT ported.
   - Placeholder chart bars deleted; `.admin-chart`/`.seg-chart`/`.legend` CSS NOT ported (no time-series/segment endpoints). Where a chart panel existed, either drop the panel or show real tabular data the endpoint already returns — decided per page in §4.
   - Dead "Last 30 days" / "Export report" buttons removed.
   - Analytics placeholder series removed; only endpoint-backed data renders.
   - Backend follow-ups logged (§8): revenue time-series, activity/events feed, KPI period deltas.

## 3. Layout and shared components

### 3.1 `admin-layout.tsx`
Becomes the mockup shell: `<div className="admin-layout">` with `<aside className="admin-side">` and `<main className="admin-main"><Outlet/></main>`. Unlike account pages, the layout owns `<main className="admin-main">` because the mockup's admin layout does (admin pages are content fragments and none renders its own `<main>` — verify none does at implementation). Keeps: `useAuth` loading state (restyled muted), existing `navSections` data structure (relabeled to mockup section names — they already match), TanStack `Link` + `activeProps: 'active'` with `activeOptions={{ exact: true }}` on `/admin`. Sidebar bottom: bordered-top foot with `Link` to `/` "← Storefront" (adapted class `admin-side-foot`, mirroring `acct-side-foot`).

### 3.2 Shared UI (`features/shared/ui/admin/`)
- `Panel` → emits `admin-panel` / `admin-panel-head` (h3 + actions) / `admin-panel-body`. Same props interface (`title?`, `actions?`, `children`) so all nine consumer pages re-skin for free.
- `KPICard` → emits `admin-kpi` / `admin-kpi-k` / `admin-kpi-v`; `delta`/`deltaDirection` props DELETED (honest data; the CSS `.admin-kpi-delta` rules are still ported since a real deltas endpoint is a logged follow-up).
- `StatusTag` → emits `admin-tag tag-{variant}`. Mockup variants: `tag-live`, `tag-draft`, `tag-low`, `tag-oos`. Real order statuses (pending/paid/failed/cancelled) get extrapolated `tag-*` palette entries mapped from the account StatusPill colors (banner-commented). Inspect the current `status-tag.tsx` mapping at implementation and preserve its status→variant logic, only swapping emitted classes.

## 4. Page specs (all: strip Tailwind classNames, keep every hook/mutation/handler, `admin-head` pattern = eyebrow + display h1 + `admin-head-actions`)

- **Dashboard**: greeting head (real user name via useAuth, mockup's "Good morning" copy without fake buttons); 4 real KPIs (Revenue via formatGhs, Orders, Avg. order, Customers) with no deltas; recent-orders `admin-tbl` (id, date, total, StatusTag) from `useGetAdminDashboard`'s `recent_orders`; fake chart + activity panels deleted. If the dashboard response exposes other real aggregates, render them as KPI cards or table rows — nothing else.
- **Orders**: `admin-filter-bar` (search input + status select or `admin-chip` row — match current filter state), `admin-tbl` listing, existing pagination restyled with `admin-btn admin-btn-sec`. **Order detail**: `admin-2col` — items panel + summary/status panel; the existing `patchAdminOrdersIdStatus` status-update control becomes `admin-form` select + `admin-btn admin-btn-pri`; back link. All wiring untouched.
- **Products**: `admin-tbl` with `row-prod` cells (44px image via getImageUrl with `ph-sm` fallback, serif name, sku/slug small), price via formatGhs, StatusTag for stock/live state as the current page computes it; existing filters → `admin-filter-bar`/`admin-chip`.
- **Customers**: `admin-tbl` port of existing columns; the pre-existing client-side search/tier filters keep working, restyled (`admin-filter-bar`). (Client-side-only filtering is a known pre-existing backlog item — unchanged.)
- **Marketing / Content / Settings**: restyle whatever currently renders — panels, forms (`admin-form` classes), tables — preserving existing honest empty/disabled states; anything currently fake in these pages is stripped under §2.3. No new features.
- **Analytics**: only endpoint-backed data renders as `admin-tbl`/KPI panels; placeholder series and the conic `seg-chart` never ported. If nothing real remains for a section, render an honest `admin-panel` with a muted "No analytics data yet — see backend follow-ups" body.

## 5. CSS — `src/styles/admin.css`

New file `@import`ed from `globals.css` after `account.css`. Port from `Rue/admin.css`: `.admin-body` (apply on the layout root? mockup sets it on body — instead fold its background/font onto `.admin-layout`, banner-noted adaptation), `.admin-layout`, `.admin-side*`, `.admin-main`, `.admin-head*`, `.admin-kpis`, `.admin-kpi*` (incl. delta rules for future), `.admin-panel*`, `.admin-tbl*` (incl. `row-prod*`, `ph-sm`, `.num`), `.admin-filter-bar`, `.admin-chip`, `.admin-btn*`, `.admin-tag` + tag variants (+ extrapolated order-status tags), `.admin-form*`, `.admin-2col`. NOT ported: `.admin-chart*`, `.seg-chart`, `.legend*`, `.admin-activity`/`.activity-*` (fake-data only). Adapted additions (banner): `admin-side-foot`, order-status tag palette, anything the pages need that the mockup styled inline. Dedup rule: grep all existing style files first (`.badge` and `.pill` exist elsewhere — the mockup's `.admin-side-brand .badge` and `.admin-side .pill` are scoped so port carefully; pills are omitted anyway per §2.2, so skip `.admin-side .pill` rules). Button-chrome: `.admin-btn` sets `border: 0` and backgrounds already; `.admin-chip` sets background/border; verify every ported button selector defines background+border (tranche-2 lesson) and patch scoped resets where missing.

## 6. Non-goals

No backend changes; no new admin features (no real charts, exports, date-range filters, activity feeds); no route/guard changes; wishlist/marketing feature work out of scope; no Tailwind re-enable.

## 7. Testing & verification

Same gates as tranche 2: per-task `pnpm typecheck`; final full gate (typecheck/lint/vitest/build); Tailwind-residue + class-coverage audits over `features/admin/**` AND `features/shared/ui/admin/**`; template-literal class check (StatusTag variants); route sweep `/admin/*`; browser walkthrough deferred to human (needs admin login).

## 8. Backend follow-ups (append to backlog)

1. Revenue time-series endpoint (unlocks `admin-chart` — port CSS then).
2. Admin activity/events endpoint (unlocks activity feed).
3. KPI period-comparison deltas (unlocks `admin-kpi-delta` + KPICard delta props).
4. Customer segments aggregate (unlocks `seg-chart`).
(Existing queue unchanged.)
