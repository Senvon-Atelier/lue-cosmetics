# Account Suite UI Alignment — Design Spec (Tranche 2)

Date: 2026-07-02
Status: Approved direction (brainstorm answers 2026-07-02); spec committed for async user veto
Mockup sources: `Rue/src/acct-pages.jsx`, `Rue/account.css` (read-only reference outside the repo)
Foundation: all §2/§3 decisions from `2026-07-02-funnel-ui-alignment-design.md` are binding here (pure mockup CSS, Tailwind dormant, class-parity, dedup rule, functionality freeze, honest data, formatGhs, getImageUrl, shared icons).

## 1. Goal

Make the account suite — layout/sidebar, dashboard, orders list, order detail, addresses, wishlist, settings — look like the `Rue/` account mockup while keeping every existing hook, mutation, guard, and route path working. Wishlist stays an honest empty state (no backend).

## 2. Decisions from brainstorming (binding)

1. **Standalone chrome**: account moves out of `_storefront` into the mockup's full-viewport `.acct-layout` (260px sticky sidebar + content), no storefront header/footer. Same pattern as `_auth`/admin. URLs unchanged; the `beforeLoad` auth guard is untouched.
2. **Sidebar = real links only**: Overview, Orders, Wishlist (under "Shop"), Addresses, Settings (under "Account"), plus Sign out. The mockup's 8 backend-less links (tracking, subscriptions, returns, reviews, reorder, loyalty, referral, payments) and their badges are omitted entirely. No badges (no cheap real counts).
3. **Dashboard = real-data equivalents**: lifetime-orders card (real total), saved-addresses card (real count), accent card repurposed to account status (member name/email + verified state) instead of fake Rue Points; recent-orders table (3 rows, real); quick links Orders / Addresses / Wishlist with honest subtitles.
4. **Orders data adaptations**: (a) order detail drops the mockup's fake 4-step shipment timeline — status pill only, with extrapolated pill palette for the real statuses (`pending/paid/failed/cancelled`); `.timeline`/`.tl-*` CSS is NOT ported (YAGNI, like the swatch precedent). (b) Orders table drops the mockup's "Items" column — the list endpoint (`InternalMeOrderResponse`) has no item count; logged as backend follow-up (§8).
5. **Region select**: address form adopts the mockup's `<select>` but with Ghana's full 16-region list, exported as a shared const from `src/content/regions.ts`. Same string field submitted (freeze holds). Checkout's free-text region input switch is a noted follow-up, not in scope.

## 3. Routing and layout

- `router.tsx`: new pathless `accountLayoutRoute` (`id: '_account'`) parented on the root route, rendering `AccountLayout`. The existing `accountRoute` (path `/account`, with its `beforeLoad` guard) reparents from `storefrontLayoutRoute` to it; all child routes and paths stay identical. Order-detail param id `from` string (`/_storefront/account/orders/$id`) must be updated to the new route id — verify with typecheck (Route typing catches it).
- `account-layout.tsx` becomes the mockup shell: `<div className="acct-layout">` containing `AcctSidebar` + `<Outlet/>`. **The layout must not wrap `Outlet` in `<main>`** — each page renders its own `<main className="acct-main">` (tranche-1 pitfall).
- Loading state (auth `isLoading`) keeps working, restyled with muted mockup text (no Tailwind).

### Sidebar (`AcctSidebar`, inside account-layout.tsx)

Port mockup lines 6–53 with real data:
- `acct-side-brand`: existing brand mark linking to `/` (reuse whatever brand component/markup the storefront header uses, sized small like the mockup's `Brand size="sm"`).
- `acct-me`: avatar = first letter of `user.name` (fallback: first letter of email), `acct-me-name` = user name (fallback email local-part), `acct-me-tier` line = the user's email (no loyalty tiers — honest substitute; keep the class for the muted small-caps look).
- Nav links use TanStack `Link` with `activeProps={{ className: 'active' }}` (exact-match semantics so `/account` isn't active on children; use `activeOptions={{ exact: true }}` for Overview). Icons from `features/shared/ui/icons.tsx` (user, bag, heart) at size 14, matching mockup pairings; links without a matching mockup icon render without one (mockup itself has icon-less links).
- Sign out: the mockup's bordered-top bottom link, wired to the existing logout handler (whatever the storefront header currently calls — reuse it verbatim), then redirect as the existing logout flow does.

## 4. Page specs

All pages: strip every Tailwind utility className as the file is ported; keep mockup class names verbatim; prices via `formatGhs`; images via `getImageUrl` with `ph` placeholder fallback; dates via a small shared `formatOrderDate` helper (e.g. "Apr 04, 2026" — mockup style) placed in `src/lib/format/` next to formatGhs. Order numbers keep the current display rule: `#{id.slice(0,8).toUpperCase()}` styled as `.o-id`.

### 4.1 Shared account primitives (`features/account/`)

- `AcctHead` ({eyebrow, title, children}) — mockup lines 56–64, as a small exported component in the account feature (used by every page).
- `StatusPill` ({status}) — mockup lines 66–71 adapted to real statuses. Class map: `paid → status-paid`, `pending → status-pending`, `failed → status-failed`, `cancelled → status-cancelled`. CSS palette extrapolated from the mockup's semantics: paid = mockup's delivered green, pending = processing amber, failed = cancelled red, cancelled = refunded grey. Banner-comment these as extrapolated in `account.css`.

### 4.2 Dashboard (`account-dashboard.tsx`)

- `AcctHead` eyebrow "Welcome back", title `Hello, {firstName}.` (first word of name; fallback "Hello.").
- One `getMeOrders({ limit: 3 })` call serves both the lifetime-orders total and the 3 recent rows. One `getMeAddresses()` call for the address count. Both via the page's existing fetch style (plain async calls in state, as today).
- `acct-cards`: (1) "Lifetime orders" value = `total`, sub = "All time"; (2) "Saved addresses" value = count, sub = default label if any (`{label} set as default`) else "No default set"; (3) `acct-card-accent`: k = "Member", v = first name (serif italic per accent styles), sub = `email · Verified`/`· Unverified` from `user.email_verified`. No fake numbers anywhere.
- Recent orders: `acct-section` with head "Recent orders" + `auth-link` "View all →" (Link to `/account/orders`), `orders-table` with columns Order / Date / Total / Status / (View) — no Items column. Rows link "View" to the detail route. Empty state (total 0): keep the table region replaced by a muted line + "Start shopping" `btn btn-primary` to `/shop` (honest, mirrors current behavior).
- Quick links: `dash-links` grid with three real `dash-link-card`s — Orders ("{total} orders placed"), Addresses ("{n} saved address(es)"), Wishlist ("Saved items — coming soon"). Arrow icon size 16. Each navigates via router.

### 4.3 Orders list (`account-orders.tsx`)

- `AcctHead` eyebrow "History", title "Your orders".
- Filter: replace the `<select>` with mockup `sub-tabs` — All / Pending / Paid / Failed / Cancelled mapping to the existing `statusFilter` state values (`''`, `pending`, `paid`, `failed`, `cancelled`). Same reset-to-page-0 behavior.
- `orders-table`: head Order / Date / Total / Status / (blank action col); rows: `.o-id`, formatted date, `.price` via formatGhs, `StatusPill`, `.link-btn` View → detail. Grid template adapted to 5 columns (the mockup's 6-col template minus Items) — note as an adapted selector in `account.css` (`/* adapted: no items column */`).
- Loading/error/empty states keep existing logic; restyle as muted text / `alert alert-warn` for errors / the dashboard-style empty state.
- Pagination keeps existing state; restyle Previous/Next as `btn btn-ghost` with the muted page indicator between.

### 4.4 Order detail (`account-order-detail.tsx`)

- `back-link` "← All orders" (Link to `/account/orders`; arrowLeft icon 12).
- `AcctHead` eyebrow `Order #{shortId}`, title "Order details"; right side = `StatusPill` (replaces the mockup's Start-a-return/Track buttons — no backend for either).
- Two-column `order-detail-grid` (mockup line 207; lift its inline grid into a real `account.css` class since we can't carry JSX inline styles into clean ported markup — banner-comment as adapted): left = Items `form-card` (rows: image via `getImageUrl(product_image_snapshot)` in an 80×100 `ph`-fallback block, serif name from `product_name_snapshot`, label-font brand snapshot + `Qty {qty}`, `.price` line = unit×qty); right = Summary `form-card` (Row k/v pairs: Subtotal, Shipping, divider, Total bold `.price`; "Delivering to" block from the real `shipping_address` (pre-line); "Payment" block = `Paystack · {paystack_reference}` — real reference, no fake MoMo mask).
- The mockup's `Row` helper is ported as a tiny local component with a real class (`.kv-row`, adapted) instead of inline styles.
- Placed-date line under the head (label font, muted). Reorder button: current page has a stub `alert()` — REMOVE the fake button (honest data rule; reorder has no wiring). Keep "Continue shopping" as `btn btn-ghost` → `/shop`.
- Loading/error states keep logic; restyle (error = `alert alert-warn` + back-link).

### 4.5 Addresses (`account-addresses.tsx`)

- `AcctHead` eyebrow "Delivery", title "Address book"; right = `btn btn-primary` "+ Add address" (plus icon 14) toggling the existing form state.
- Form (add + edit share it, as today): `form-card` with `form-row`s — Label / (no Full-name field — API has none, omit mockup's), Street address (line1), Apartment etc (line2, optional), City / Region `<select>` (16 regions from `src/content/regions.ts`), Phone. Field errors render as `.field-error` under inputs (existing validation kept). Save/Cancel = `btn btn-primary` / `btn btn-ghost`, existing submit/cancel handlers kept, including edit prefill (note: current code sets `editingAddress` but never opens the form on Edit — if that's a live bug, fix minimally by also `setShowForm(true)` in the Edit handler; verify at implementation time and record in ledger).
- List: `addr-grid` of `addr-card`s — `default` modifier + `pill` "Default" for `is_default`; `h4` = label; `p` = line1 / line2 / city, region / phone; `.actions` = Edit, Set default (non-default only, existing handler), Remove (`danger`, keeps the `confirm()`).
- Empty state: AcctHead + muted copy + "Add your first address" primary button (existing behavior, mockup styling).

### 4.6 Wishlist (`account-wishlist.tsx`)

- `AcctHead` eyebrow "Saved for later", title "Wishlist".
- Honest empty state only: muted copy ("No saved items yet — wishlist is coming soon." + a line that saving isn't available yet) and `btn btn-primary` "Explore products" → `/shop`. Delete the dead fake-grid markup, the commented-out generated-client imports, the `WishlistItem` type, and the stub handlers — the page becomes a static empty state until the wishlist backend (§8 tranche-1 item 2) lands. No `ph` product grid, no fake counts.

### 4.7 Settings (`account-settings.tsx`)

- `AcctHead` eyebrow "Settings", title "Profile".
- Profile `form-card`: `alert alert-info` "Profile editing isn't available yet."; `form-row` with Name and Email `field`s, both real values, `readOnly`/`disabled` as today; email keeps its verification hint as `.field-hint`. No birthday / skin-type / marketing fields (all fake). No disabled save button — the alert already says why (drop the "Save Profile Unavailable" button).
- Password `form-card`: `alert alert-info` pointing at the login page's reset flow; drop the three disabled password inputs and disabled button (dead form theater) — alert + a `btn btn-ghost` "Reset via email" linking to `/forgot-password` instead.
- Danger zone: `form-card` with `alert alert-warn` "Account deletion isn't available yet." — no disabled button.
- This is the one page where markup intentionally diverges from the mockup (mockup's profile form is 90% fake fields); the honest-data rule wins.

## 5. CSS — `src/styles/account.css`

New file `@import`ed from `globals.css` after `auth.css`. Port from `Rue/account.css` ONLY:
- `.acct-layout`, `.acct-side*` (incl. badge rules — harmless), `.acct-me*`, `.acct-avatar`, `.acct-main`, `.acct-head`, `.acct-cards`, `.acct-card*`, `.acct-section*`
- `.orders-table`, `.orders-row*` (with the 5-column adapted grid), `.status`, `.status-dot`, adapted `.status-paid/-pending/-failed/-cancelled`, `.link-btn`
- `.form-card` (not in pages.css — verify by grep before adding)
- `.addr-grid`, `.addr-card*`
- `.sub-tabs*`, `.dash-link*`, `.dash-links`
- `.alert`, `.alert-info`, `.alert-success`, `.alert-warn`
- Adapted additions under a `/* adapted/extrapolated */` banner: `.order-detail-grid`, `.kv-row`, status palette.

Do NOT port (dedup / YAGNI): `.field*`, `.form-row*` (pages.css), `.auth-*`, `.back-link` (auth.css), `.timeline`/`.tl-*` (no fulfilment), `.loyalty-*`, `.star-picker` (no backends), `.legal-*` (tranche 5). Grep `globals.css`, `pages.css`, `auth.css` for every selector before appending (`.badge` exists in globals — reuse/verify it doesn't conflict with `.acct-side a .badge`, which is scoped and fine; confirm at implementation).

## 6. Non-goals

- No backend changes; no wishlist/loyalty/tracking/returns/reviews/subscriptions/payments features.
- No timeline UI or CSS.
- No checkout region-select change (follow-up only).
- Backlog trio (CartItemRow extraction, drawer a11y, Google glyph) does not touch account files — deferred to the tranche that touches cart/global chrome/auth.

## 7. Testing & verification

- Gate per task: `pnpm typecheck` clean. Full gate before merge: typecheck / lint / vitest / build.
- Route regression: all `/account/*` URLs render (guard still redirects anonymous → login); order-detail `useParams` `from` id updated and type-checked.
- Class-coverage check (tranche-1 pitfall): for each ported page, verify every className it emits is defined in account.css/globals/pages/auth (scripted grep in the plan).
- Tailwind-residue grep per ported file.
- Functional smoke: dashboard counts match orders/addresses endpoints; orders filter tabs hit the API with the right status param; address add/edit/set-default/delete round-trip; sign-out from sidebar works; wishlist/settings render their static states.

## 8. Backend follow-ups (append to the running backlog)

1. **Order list item counts** — add `item_count` (or item snippets) to `InternalMeOrderResponse` so the orders table can show the mockup's Items column.
2. **Profile update endpoint** (name/phone/marketing consent) — unlocks the settings form.
3. **Checkout region select** — switch checkout's free-text region input to the shared 16-region select (frontend, tranche-1 page touch-up).
(Existing queue unchanged: wishlist domain, product narrative fields, gallery images, cart brand/size enrichment, URL-driven shop filters, checkout shipping quote.)
