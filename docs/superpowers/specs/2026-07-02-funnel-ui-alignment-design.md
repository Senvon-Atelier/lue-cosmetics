# Purchase-Funnel UI Alignment — Design Spec (Tranche 1)

Date: 2026-07-02
Status: Approved direction; spec pending user review

## 1. Goal

Make the purchase funnel — product detail, cart (drawer + page), checkout, and the five auth pages — look exactly like the `Rue/` mockup while keeping every piece of existing functionality (generated-hook data fetching, cart mutations, checkout init/verify, auth flows). This is the first tranche of the full UI alignment; account, admin, global chrome (search overlay, mobile menu), and the legal/marketing suites come later.

## 2. Context and constraints

- The mockup at `casestud/Rue/` is a React JSX prototype (`src/*.jsx`) styled by plain CSS files (`styles.css`, `pages.css`, `account.css`, …) over CSS custom properties. The frontend already carries the tokens verbatim in `src/styles/globals.css` and ~143 shared classes (`wrap`, `eyebrow`, `chip`, `section`, header/footer/home classes).
- **Tailwind is installed but not wired into the build** — no PostCSS config, no Vite plugin, so no utilities are generated. Every Tailwind className in the app is dead code today.
- Home, header, footer, announcement bar, and most of the shop page were already aligned to the mockup (2026-07-01 work). Do not redo them.
- Currency is GHS minor units; render via the existing `formatCurrency`-style helpers (`GH₵`/`GHS` per mockup copy — the mockup uses `GHS 480`; match the mockup's format in ported markup).

## 3. Styling foundation (decided)

**Pure mockup CSS now; Tailwind migration deferred to a future project.**

1. In `globals.css`, comment out `@import "tailwindcss";` with a note: re-enable when the future CSS→Tailwind migration project starts. Do NOT uninstall the tailwind packages or delete `tailwind.config.ts`; add a header comment to `tailwind.config.ts` marking it dormant and pointing at this spec.
2. CSS file organization mirrors the mockup: `globals.css` keeps tokens + shared classes and gains two plain `@import './pages.css'` / `@import './auth.css'` lines. New files:
   - `src/styles/pages.css` — PDP + drawer + toast + extrapolated cart/checkout classes, ported from `Rue/pages.css` (and `Rue/styles.css` where drawer/toast classes live). Port only the classes this tranche uses; later tranches extend it.
   - `src/styles/auth.css` — the `auth-*`, `field`, `back-link` classes from `Rue/account.css`.
3. Dead Tailwind classNames are stripped from each file **as it is ported** in this tranche. Files outside the tranche keep theirs until their own tranche.
4. Class parity rule: ported markup keeps the mockup's class names verbatim so future diffs against the mockup stay trivial.

## 4. Page specs

### 4.1 Product detail (`features/catalog/product-detail.tsx`)

Port `Rue/src/pages.jsx` `ProductPage` (lines 144–260) structure 1:1:

- Breadcrumb (`wrap breadcrumb`): Home → Shop → {category label} → {product name}; links use TanStack `Link` (`/`, `/shop`, `/shop?category=…` matching current shop filter behavior).
- Two-column `wrap product-grid`: gallery (thumb rail + main image) and info column.
- Gallery: the backend has ONE image per product (`image_path`). Render the main image from `image_path` with a `ph ph--{tone}` placeholder fallback (the `tone` field exists on the API). Render the thumb rail ONLY if more than one image exists — i.e. omitted for now; keep the CSS so it lights up when images arrive. No fake duplicate thumbs.
- Info column, in mockup order: brand eyebrow (resolve `brand_id` via existing brands data), `h-display product-name`, rating row (real `rating` + `review_count`; star icons from the existing `Icon` set), price row (`price`, conditional `price-was` from `was_price_ghs_minor`, `product-size` from `size`), lede paragraph, actions, perks, tabs.
- Lede: static category-keyed copy (see 4.5); links to the category's shop filter.
- **Swatch row omitted** — no variant data in the backend. Do not port `product-swatches` markup (keep the CSS out too; YAGNI).
- Actions (`product-actions`): qty stepper (`qty qty-lg`, local state), primary CTA `Add to bag · GHS {price×qty}` wired to the existing add-to-cart mutation (which must also open the toast, 4.2), and the wishlist heart button rendered **disabled** with `title="Saved items coming soon"` — no fake interaction (backend wishlist doesn't exist).
- Perks list (`product-perks`): free-delivery line must show the REAL threshold from the shipping quote config (the shipping quote endpoint exposes free-over; if not cheaply available on this page, use the static copy from `src/content/` with the correct configured value, not the mockup's hardcoded one). Authenticity + WhatsApp lines are static copy.
- Tabs (`product-tabs`): Description / How to use / Ingredients. Content comes from a static curated map (4.5). If a product's category has no curated copy, render only the Description tab with the generic fallback.
- Related products (`section` below the grid): reuse the existing aligned `ProductCard` with products from the same category (existing list endpoint, `category` filter, exclude self, cap 4). Skip the section when empty.

### 4.2 Cart drawer, toast, and cart page

**Drawer** (`features/cart/cart-drawer.tsx`): port `Rue/src/shared.jsx` `CartDrawer` (lines 173–241) visuals onto the existing drawer logic — scrim + slide-in `drawer` aside, `drawer-head` (eyebrow "Your Bag" + item count + close), item rows (`cart-item`: image/placeholder, brand, name, size, qty stepper, line price, remove ×), empty state (circular placeholder, display-type headline "Your bag is empty", "Shop the edit" CTA), footer (`drawer-foot`: subtotal row, muted "Delivery — Calculated at checkout" row, full-width `Checkout · GHS {subtotal}` CTA navigating to `/checkout`, "Continue shopping" link). All quantities/removals keep using the existing cart mutations; line images use real `image_path` with `ph` fallback.

**Add-to-cart toast**: new shared component (`features/cart/add-toast.tsx` or inside the cart provider) matching the mockup `.toast`: check icon, "**Added.** {product name}", "View bag" button that opens the drawer. Triggered by successful add-to-cart from any page (PDP, shop cards). Auto-dismiss ~2.4s (mockup timing). Rendered once at root-layout level.

**Cart page** (`features/cart/cart-page.tsx`) — extrapolated (no mockup): `wrap` page with eyebrow "Your Bag" + `h-display` headline (mirrors the mockup's wishlist-page pattern at `Rue/src/app.jsx:89–107`), then a two-column layout: left = `cart-item` rows reused from the drawer at page scale; right = summary card (surface-tinted panel, radius-lg) with subtotal, delivery note, checkout CTA, and the free-shipping-remainder line the current page already computes. Empty state mirrors the drawer's. Same class vocabulary; new classes go in `pages.css` under a `/* cart page (extrapolated) */` banner so future readers know they're not mockup-sourced.

### 4.3 Checkout (`features/checkout/checkout-page.tsx`, `checkout-return.tsx`) — extrapolated

Keeps `CheckoutLayout` (brand + minimal chrome) and ALL existing logic (address form state, shipping quote, `postCheckoutInit`, redirect to Paystack, return-page verify polling).

- Layout: centered column (max ~720px) with eyebrow "Checkout" + `h-display` headline, then sections separated by `line` borders: 1) Delivery address — mockup `field` inputs (label above input, Manrope labels, generous spacing); 2) Delivery method — bordered radio cards (border `--line`, selected state `--lavender-600` border + `--surface` fill) showing method name + real quoted price; 3) Order summary — item rows (name × qty, line total) + subtotal/delivery/total rows in the drawer's row style; 4) full-width `btn btn-primary` "Pay with Paystack · GHS {total}".
- Return page: single centered status card (radius-lg, surface fill): success = display-type "Thank you." + order reference + "View your orders" CTA (`/account/orders`); pending = muted spinner copy (polling continues as today); failed = display-type headline + retry CTA back to `/checkout`.
- New classes live in `pages.css` under `/* checkout (extrapolated) */`.

### 4.4 Auth (`features/auth/*`, router)

Port `Rue/src/acct-auth.jsx` (`AuthShell` + Login/Signup/Forgot/Reset/Verify):

- `AuthShell` becomes `features/auth/auth-shell.tsx`: two-panel full-viewport layout — left visual panel (lavender `ph` block, "Rue · Members" eyebrow, "Small rituals, *long kept.*" display copy) and right form panel (`auth-form`: "← Back to Rue" link to `/`, sub eyebrow, display h1, children, footer link).
- Router: auth routes (`/login`, `/signup`, `/forgot-password`, `/reset-password`, `/verify-email`) move out of the `_storefront` layout into a new pathless `_auth` layout route rendering the shell chrome-free (same pattern as `_checkout`). URLs unchanged. The `beforeLoad` guards on `/account` and `/admin` are untouched.
- Login: social row = **Google only** (wired to the existing `/auth/google/start` redirect; mockup's Apple button omitted — no backend), `auth-divider` "or with email", email/password `field`s, show-password checkbox, "Forgot password?" link, full-width primary CTA. Keep existing submit handler, error display (map `ApiError` message into the form's error area), and post-login redirect behavior.
- Signup/Forgot/Reset/Verify: same shell, mockup field layouts, existing handlers and validation kept. Any current UX the mockup lacks (e.g. success messages after forgot-password submit) is kept and styled with `auth-meta`/muted text.

### 4.5 Static curated copy

New `src/content/product-copy.ts`: category-slug-keyed `{ lede, description, howTo, ingredients? }` curated from the mockup's copy (e.g. `pages.jsx:190–236`) — editorial site copy, clearly not per-product DB data. Also the perks copy with the real free-delivery threshold. When the backend later gains per-product `description/how_to/ingredients` columns, the API values take precedence; note this as a backend feature request (out of scope here).

## 5. Non-goals

- No Tailwind wiring (explicitly deferred; import stays commented out).
- No backend changes of any kind.
- No search overlay, mobile menu, wishlist page, blog, account/admin/legal/marketing pages (later tranches).
- No new cart/checkout/auth behavior — visual only; every handler, hook, guard, and route path stays.

## 6. Testing & verification

- Existing gates stay green: `pnpm typecheck`, `pnpm build`, `pnpm lint`, `pnpm vitest run` (guard tests).
- Per-page visual verification during implementation: dev server screenshots compared against the mockup opened from `Rue/*.html` (needs network for CDN React; if offline, compare against the JSX/CSS source directly).
- Functional smoke per page after port: PDP add-to-cart (toast appears, drawer count updates), drawer qty/remove, cart page totals, checkout quote → Paystack redirect init (test key), auth login/signup round-trip, verify/reset flows reachable.
- Lint note: ported markup must not reintroduce dead Tailwind classNames in tranche files; a grep check per file (`grep -c "className=\"[^\"]*\b(flex|grid|px-|py-|text-)" …`) belongs in the plan's verification steps (heuristic — reviewer judgment applies).

## 7. Risks / open points

- **Curated-copy decision (4.5)**: product tabs/lede use static category-level copy until the backend has real fields. Flagged for user veto.
- The mockup renders prices as `GHS 480` while some existing components use `GH₵`; this tranche standardizes funnel pages on the mockup's `GHS {amount}` format via the shared formatter (single place to flip later).
- Auth pages moving out of the storefront layout removes header/footer from them (mockup-intended); confirm nobody relies on the cart drawer being reachable from auth pages.

## 8. Backend follow-ups (deferred, decided 2026-07-02)

Frontend-only for this tranche. Queued for a later backend round, in priority order:

1. **Product narrative fields** — migration adding nullable `description`, `how_to`, `ingredients` text columns; update sqlc query + product view + swag annotations; regenerate; extend `products.json` seed. PDP then prefers API values over the curated static copy automatically (§4.5).
2. **Wishlist domain** — the originally-specced handler→repository single-table CRUD under `/me/wishlist`; unlocks the PDP heart and the account wishlist page.
3. **Product gallery images** — `product_images` table (`product_id`, `path`, `sort_order`) + detail-query join; PDP thumb rail lights up automatically (§4.1).
4. **Variants/shades** — full variant modeling (cart/order ripple); only if the business needs it. The swatch UI stays out until then.
