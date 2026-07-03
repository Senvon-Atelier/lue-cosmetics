# Global Chrome UI Alignment — Design Spec (Tranche 4)

Date: 2026-07-03
Status: Approved direction (brainstorm answers 2026-07-03); spec committed for async user veto
Mockup sources: `Rue/src/shared.jsx` (`SearchOverlay` lines 244–335, `MobileMenu` lines 336–382), `Rue/styles.css` (search-overlay / mnav / drawer-left blocks) — read-only reference
Foundation: tranche-1 §2/§3 decisions binding (pure mockup CSS, class parity, dedup, functionality freeze, honest data, formatGhs/getImageUrl/icons). Tranche precedents: scoped button-chrome resets; `<button>`/`<Link>` over `<a onClick>` for actionables.

## 1. Goal

Give the site its missing global chrome: a working SearchOverlay (wired to the products list `q` param) and a MobileMenu (the site currently has NO mobile nav), plus the scope adds George approved: URL-driven shop category filters (clears the tranche-1 backlog item) and three folded backlog items (drawer focus/visibility a11y, shared CartItemRow, Google glyph on login).

## 2. Decisions from brainstorming (binding)

1. **Search idle state**: trending chips as curated static copy (`src/content/search-terms.ts`) + a product rail relabeled **"From the shop"** fed by the plain products list (`getProducts({ limit: 3 })`) — no fake popularity claim.
2. **Search results**: debounced (~300ms) `getProducts({ q, limit: 6 })`; result rows navigate to `/shop/$slug`; honest empty state with a link to `/shop`.
3. **URL-driven category filter THIS tranche**: shop route gains `validateSearch` for `?category=<slug>`; the shop page resolves slug→id against the loaded categories and syncs its existing `selectedCategory` state both ways (URL → state on load/param change; filter clicks navigate with the new slug or clear it). Mobile-menu category rows and PDP breadcrumb/category links deep-link with `?category=<slug>`. Brand filter stays client-only (backlog).
4. **Header wishlist heart**: stays, disabled, `title="Saved items coming soon"` (PDP-heart precedent); the dead `wishlistCount` state in the cart provider is removed (honest: it is a constant 0 today).
5. **Backlog folds**: (a) drawer focus/visibility a11y — closed drawers/overlays must not be tabbable (visibility/`inert` handling) and Esc closes them, applied to the cart drawer AND the two new surfaces; (b) `CartItemRow` extraction shared by cart drawer + cart page (visual output unchanged); (c) Google "G" glyph SVG added to the login social button.

## 3. Components

### 3.1 CSS (`src/styles/globals.css` — these classes come from the mockup's global `styles.css`, which globals.css mirrors)
Port under banners: `.search-overlay` (+ `.open`), `.search-head`, `.search-input-wrap`, `.search-input`, `.search-clear`, `.search-body`, `.search-section`, `.search-chips`, `.search-picks`, `.search-pick`, `.search-empty`; `.drawer-left` (left-side variant of the existing ported `.drawer`); `.mobile-nav`, `.mnav-section`, `.mnav-contact` (`.mnav-count` NOT ported — no category counts exist). Verify `.mobile-menu-btn` styling exists (header renders it today but no frontend CSS defines it — same missing-class bug as `.brand` was; port the mockup's rule that hides it ≥900px and shows it below, or adapt if the mockup names it differently). Dedup rule applies (chip/eyebrow/wrap/icon-btn/drawer classes already exist — do not redefine). Every new button selector defines its own background/border.

### 3.2 SearchOverlay (`features/shared/search-overlay.tsx`)
Controlled by Header state (`open`, `onClose`). Mockup structure verbatim: search head (icon + input + clear + close), body with idle sections / results / empty. Real data: debounced `q` → `getProducts({ q, limit: 6 })` via the generated hook (only enabled when `q` trimmed non-empty); result rows use real `image_path` via `getImageUrl` in the 64×80 `ph` block (fallback tone class), brand resolved like the PDP does, `formatGhs` price, `Link`/navigate to `/shop/$slug` then close. Idle: chips from `SEARCH_TERMS` const (clicking fills the input); "From the shop" = `getProducts({ limit: 3 })` (fetched only while open+idle, or reuse a small hook — plan decides; no fetch when closed). Focus input on open; Esc closes; closed overlay is not tabbable (visibility + inert per §2.5a). Loading state: muted "Searching…" line; error: muted "Search is unavailable right now." (honest, no fake results).

### 3.3 MobileMenu (`features/shared/mobile-menu.tsx`)
Controlled by Header state. Mockup structure: `drawer-scrim` + `aside.drawer.drawer-left` with drawer-head (Brand → `/`, close btn) and `drawer-body.mobile-nav`: Pages section (Home `/`, Shop `/shop`, About `/about`, Journal `/` hash=journal — same targets as the desktop header, active state from router where cheap), Shop-by-category section (categories from the existing categories data → `/shop?category=<slug>`, no counts), Visit-us contact block. Contact copy: extract the footer's existing address/phone/hours strings into `src/content/store-info.ts` and consume from BOTH footer and menu (no duplicated copy). All rows are `Link`s/`button`s (keyboard-actionable); menu closes on navigation; Esc closes; closed menu not tabbable.

### 3.4 Header (`features/shared/layouts/header.tsx`)
Gains `useState` for `searchOpen`/`menuOpen`; search button and menu button wired; overlay + menu rendered by Header (position-fixed surfaces). Heart button: `disabled` + `title="Saved items coming soon"`; `wishlistCount` usage removed here and the dead state removed from `cart-provider.tsx` (interface shrinks — typecheck sweeps consumers).

### 3.5 Shop URL filter (`router.tsx` shop route + `features/catalog/shop-page.tsx`)
`validateSearch: (s) => ({ category: typeof s.category === 'string' ? s.category : undefined })` on the shop route. Shop page: on categories load / param change, resolve slug→id and set `selectedCategory` (unknown slug = no filter); `handleCategoryChange` additionally navigates with `search: { category: slugOrUndefined }` (replace: true) so state and URL stay in lockstep. Existing fetch logic (params.category = id) unchanged. PDP breadcrumb + category links switch from plain `/shop` to `/shop?category=<slug>` (slug available on the PDP's category data — verify exact field at plan time).

### 3.6 Folds
- **Drawer a11y**: cart drawer scrim/aside get `inert` (or visibility-based) treatment when closed + Esc handling — shared tiny hook (`useEscToClose`, `useInertWhenClosed`) if it keeps the three surfaces consistent; plan decides the exact mechanism (note: React 18 supports the `inert` attribute lowercase; verify TS typing — may need `// @ts-expect-error` or `inert=""` cast).
- **CartItemRow** (`features/cart/cart-item-row.tsx`): one component renders the `cart-item` row (image/ph, brand line when present, name, size when present, qty stepper wired to the passed mutation callbacks, line price, remove) consumed by drawer and cart page; markup/classes identical to today's duplicated JSX — pure extraction, zero visual change.
- **Google glyph**: multi-color "G" SVG inline in `features/auth/login` social button (decorative, `aria-hidden`).

## 4. Non-goals

No backend changes; no wishlist wiring; no brand-filter URL param; no dedicated search results page (`?q` on /shop not added); no search ranking/popularity (backlog).

## 5. Testing & verification

Standard gates (typecheck per task; full typecheck/lint/vitest/build at the end). New unit test for the debounce hook if one is written (vitest fake timers). Tailwind-residue + class-coverage audits over touched files. Functional smoke (dev servers or deferred to human): search returns live results and navigates; menu categories deep-link and the shop chip state matches the URL; back/forward updates the filter; drawer/overlay/menu all Esc-close and are untabbable when closed; login Google button shows the glyph.

## 6. Backend follow-ups (append)

1. Search endpoint improvements (ranking, brand/category match like the mockup's multi-field filter — today `q` semantics are whatever the products list implements).
2. Category product counts (unlocks `.mnav-count`).
(Existing queue unchanged; wishlist still pending for the header heart.)
