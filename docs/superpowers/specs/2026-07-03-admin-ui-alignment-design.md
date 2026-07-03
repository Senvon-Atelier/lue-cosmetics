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
   - **Amended after type inspection (2026-07-03):** the analytics endpoints already return real series — `getAdminAnalyticsRevenue` → `by_date` (time series), `by_category`, `order_stats`; `getAdminAnalyticsStats` → `customer_stats`, `product_stats`, `top_products`. So the ANALYTICS page renders a REAL `.admin-chart` bar chart from `by_date` (honest empty panel when the range has no data) and a real `.legend` category-revenue list from `by_category`. `.admin-chart` and `.legend*` CSS ARE ported; `.seg-chart` (hardcoded conic donut) is NOT. The DASHBOARD's chart + sales-by-category panels are deleted (its endpoint has no series; analytics owns them).
   - Dead controls removed everywhere: "Last 30 days"/"Export report"/"Export"/"Print labels"/"Print invoice"/"Update status"(detail)/"Import CSV"/"New product"/"New campaign"/"Media library"/"New post"/"Create segment"/"Export list"/"Invite" buttons; dead search input + row checkboxes on orders; dead category/status selects on products (their filter state is never applied); dead tier select on customers; dead "Edit" row button on products (no edit page).
   - Page-scoped fake KPIs removed: orders/products/customers KPI rows computed from the current page (mislabeled "New (24h)", "Churn risk", etc.) are dropped. Real replacements only where the response already carries them: dashboard uses `InternalAdminDashboardStats` (revenue/orders/avg/customers + per-status counts panel); products shows Total SKUs from `productsData.total`; analytics uses `customer_stats`/`order_stats` fields. Customers' invented Atelier/Bloom/Petal tier column and filter are dropped (no loyalty backend; LTV column already shows the real number).
   - Analytics placeholder fallbacks (fake sessions/conversion KPIs, fake top-products, fake traffic sources) removed; only endpoint-backed data renders.
   - Marketing and Content become honest stub pages (admin-head + one `admin-panel` with muted copy naming the missing backend); Settings becomes a read-only facts panel (store name, currency GHS, Paystack payments) + "no editable settings yet" note — no fake team/integrations/forms.
   - Backend follow-ups logged (§8): activity/events feed, KPI period deltas, marketing/content/settings domains.

## 3. Layout and shared components

### 3.1 `admin-layout.tsx`
Becomes the mockup shell: `<div className="admin-layout">` with `<aside className="admin-side">` and `<main className="admin-main"><Outlet/></main>`. Unlike account pages, the layout owns `<main className="admin-main">` because the mockup's admin layout does (admin pages are content fragments and none renders its own `<main>` — verify none does at implementation). Keeps: `useAuth` loading state (restyled muted), existing `navSections` data structure (relabeled to mockup section names — they already match), TanStack `Link` + `activeProps: 'active'` with `activeOptions={{ exact: true }}` on `/admin`. Sidebar bottom: bordered-top foot with `Link` to `/` "← Storefront" (adapted class `admin-side-foot`, mirroring `acct-side-foot`).

### 3.2 Shared UI (`features/shared/ui/admin/`)
- `Panel` → emits `admin-panel` / `admin-panel-head` (h3 + actions) / `admin-panel-body`. Same props interface (`title?`, `actions?`, `children`) so all nine consumer pages re-skin for free.
- `KPICard` → emits `admin-kpi` / `admin-kpi-k` / `admin-kpi-v`; `delta`/`deltaDirection` props DELETED (honest data; the CSS `.admin-kpi-delta` rules are still ported since a real deltas endpoint is a logged follow-up).
- `StatusTag` → emits `admin-tag tag-{variant}`. Mockup variants: `tag-live`, `tag-draft`, `tag-low`, `tag-oos`. Real order statuses (pending/paid/failed/cancelled) get extrapolated `tag-*` palette entries mapped from the account StatusPill colors (banner-commented). Inspect the current `status-tag.tsx` mapping at implementation and preserve its status→variant logic, only swapping emitted classes.

## 4. Page specs (all: strip Tailwind classNames, keep every hook/mutation/handler, `admin-head` pattern = eyebrow + display h1 + `admin-head-actions`)

- **Dashboard**: greeting head (real user name via useAuth, mockup's "Good morning" copy without fake buttons); 4 real KPIs (Revenue via formatGhs, Orders, Avg. order, Customers) with no deltas; an "Orders by status" panel from the real `pending/paid/shipped/delivered_orders` stats fields; recent-orders `admin-tbl` (id, customer, total, StatusTag) from `recent_orders`; fake chart, sales-by-category, and activity panels deleted.
- **Orders**: `admin-filter-bar` keeps only the LIVE control (status select; the dead search input and row checkboxes go); `admin-tbl` listing keeps the working inline per-row status `<select>` (the `patchAdminOrdersIdStatus` mutation — restyled, wiring untouched); pagination restyled `admin-btn admin-btn-sec`. Fake KPI row and fake header buttons removed. **Order detail**: `admin-2col` — items panel (`admin-tbl`) + status/totals/address panels; fake Print-invoice/Update-status buttons removed (status is updated from the list's inline select — adding a detail-page control would be a feature add, out of scope).
- **Products**: single real KPI (Total SKUs = `productsData.total`); `admin-filter-bar` keeps the live search input only; `admin-tbl` with `row-prod` cells (44px image via getImageUrl with `ph-sm` fallback, serif name, slug small), category/brand id shorts as today, price via formatGhs. Fake stock column ("—"), placeholder StatusTag, dead selects, dead Edit button, fake header buttons removed.
- **Customers**: `admin-filter-bar` keeps the live client-side search; `admin-tbl` columns Customer/Email/Orders/Lifetime (formatGhs)/Member since/View (route exists — a Coming-Soon stub page, honest). Tier column + tier select + fake KPI row + fake header buttons removed. (Client-side-only filtering is a known pre-existing backlog item — unchanged.)
- **Analytics**: real KPIs from `getAdminAnalyticsStats` (`customer_stats`: total/with-orders/active-30d) + `order_stats` revenue (from the revenue call); real `.admin-chart` bars from `by_date` scaled to the series max, with real date labels, honest empty panel when the series is empty; real top-products `admin-tbl` (name, units, revenue) with honest empty state; real category-revenue `.legend` list from `by_category`. All placeholder fallbacks and fake sessions/conversion KPIs deleted. The hardcoded 2024 date range stays (pre-existing; note as backlog: date-range picker).
- **Marketing / Content**: honest stub pages — admin-head + one `admin-panel` with muted copy stating the missing backend (campaigns/discounts/segments; posts/blocks/pages CMS) and pointing at the backlog. All fake tables deleted.
- **Settings**: admin-head + one `admin-panel` of read-only true facts as a small definition list (Store: Rue Cosmetics · Currency: GHS · Payments: Paystack) + muted "Store configuration lives in the backend deployment; editable settings need backend support." All fake forms/team/integrations deleted.

## 5. CSS — `src/styles/admin.css`

New file `@import`ed from `globals.css` after `account.css`. Port from `Rue/admin.css`: `.admin-body` (mockup sets it on body — fold its background/font/color onto `.admin-layout`, banner-noted adaptation), `.admin-layout`, `.admin-side*`, `.admin-main`, `.admin-head*`, `.admin-kpis`, `.admin-kpi*` (incl. delta rules for future), `.admin-panel*`, `.admin-chart` + `.admin-chart-bar` (real data exists per §2.3), `.admin-tbl*` (incl. `row-prod*`, `ph-sm`, `.num`), `.admin-filter-bar`, `.admin-chip`, `.admin-btn*`, `.admin-tag` + tag variants (+ extrapolated order-status tags), `.admin-form*`, `.admin-2col`, `.legend*`. NOT ported: `.seg-chart` (hardcoded conic donut), `.admin-activity`/`.activity-*` (fake-data only). Adapted additions (banner): `admin-side-foot`, order-status tag palette, anything the pages need that the mockup styled inline. Dedup rule: grep all existing style files first (`.badge` and `.pill` exist elsewhere — the mockup's `.admin-side-brand .badge` and `.admin-side .pill` are scoped so port carefully; pills are omitted anyway per §2.2, so skip `.admin-side .pill` rules). Button-chrome: `.admin-btn` sets `border: 0` and backgrounds already; `.admin-chip` sets background/border; verify every ported button/select selector defines background+border (tranche-2 lesson) and patch scoped resets where missing.

## 6. Non-goals

No backend changes; no new admin features (no real charts, exports, date-range filters, activity feeds); no route/guard changes; wishlist/marketing feature work out of scope; no Tailwind re-enable.

## 7. Testing & verification

Same gates as tranche 2: per-task `pnpm typecheck`; final full gate (typecheck/lint/vitest/build); Tailwind-residue + class-coverage audits over `features/admin/**` AND `features/shared/ui/admin/**`; template-literal class check (StatusTag variants); route sweep `/admin/*`; browser walkthrough deferred to human (needs admin login).

## 8. Backend follow-ups (append to backlog)

1. Admin activity/events endpoint (unlocks the activity feed).
2. KPI period-comparison deltas (unlocks `admin-kpi-delta` + KPICard delta props).
3. Analytics date-range picker (page currently hardcodes 2024; needs UI + maybe sensible server defaults).
4. Marketing domain (campaigns/discount codes/segments), Content CMS, editable Settings — each unlocks its stub page.
5. Server-side product/customer search (current search filters only the fetched page — pre-existing).
6. Admin customer detail page (route exists as a Coming-Soon stub).
(Existing queue unchanged.)
