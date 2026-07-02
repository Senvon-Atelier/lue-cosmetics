# Purchase-Funnel UI Alignment (Tranche 1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the product-detail page, cart (drawer + toast + page), checkout, and the five auth pages look exactly like the `Rue/` mockup while keeping every existing behavior (hooks, mutations, guards, routes).

**Architecture:** Pure mockup CSS (Tailwind stays installed but commented out). Each page port copies the mockup's JSX structure and class names verbatim from `casestud/Rue/src/*.jsx`, swaps its fake `window.RueData` reads for the app's generated hooks, and strips the file's dead Tailwind classNames. Page CSS is ported into `src/styles/pages.css` / `src/styles/auth.css`, imported from `globals.css`.

**Tech Stack:** React 18, TanStack Router v1 (code-based), TanStack Query v5 via Orval-generated hooks (`src/lib/api/generated/rueCosmeticsAPI.ts`, responses unwrapped), plain CSS over custom properties, vitest (+ jsdom via per-file `@vitest-environment` pragma).

**Spec:** `docs/superpowers/specs/2026-07-02-funnel-ui-alignment-design.md` — binding. **Mockup source (canonical for markup/CSS):** `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/` (read-only reference; never modify it).

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."`.
- Frontend-only: nothing under `backend/` changes; nothing under `frontend/src/lib/api/generated/` is hand-edited.
- Functionality freeze: every existing flow keeps working — PDP fetch by slug, add-to-cart, drawer qty/remove, cart totals + free-shipping remainder, checkout quote → `postCheckoutInit` → Paystack redirect, return-page verify polling, login/signup/logout/forgot/reset/verify handlers, Google OAuth start, route guards. Visual changes only.
- Class parity: ported markup keeps the mockup's class names verbatim. Extrapolated (non-mockup) classes go under a `/* … (extrapolated) */` comment banner in the CSS.
- Strip dead Tailwind classNames from every file this plan touches (they generate nothing — Tailwind isn't wired). Rule of thumb: after porting, `className` values in touched files contain only mockup/extrapolated class names.
- Currency format on funnel pages: `GHS 480` style via the shared `formatGhs` helper added in Task 1. Amounts are GHS minor units (pesewas).
- Icons come from the existing `features/shared/ui/icons.tsx` `Icon` component (`name` prop). The mockup's icon names largely match; if a name is missing from the map, use the closest existing glyph (e.g. `chevronRight` exists; `whatsapp` may not — check, and substitute `sparkle` if absent, noting it).
- Gate after every task: `pnpm typecheck` must pass with zero errors project-wide (run from `frontend/`). Full gate (`typecheck`, `build`, `lint`, `vitest run`) runs in the final task.
- Commands run from `ruecosmetics/frontend/` unless stated.

---

### Task 1: Styling foundation — dormant Tailwind, CSS files, GHS formatter

**Files:**
- Modify: `frontend/src/styles/globals.css` (top of file)
- Modify: `frontend/tailwind.config.ts` (header comment)
- Create: `frontend/src/styles/pages.css`, `frontend/src/styles/auth.css`
- Modify: `frontend/src/lib/format/utils.ts`
- Test: `frontend/src/lib/format/utils.test.ts` (new)

**Interfaces:**
- Produces: `formatGhs(minor: number): string` — `formatGhs(48000)` → `"GHS 480"`, `formatGhs(48050)` → `"GHS 480.50"`. All later tasks use it for funnel prices.
- Produces: empty `pages.css` / `auth.css` imported by `globals.css`; later tasks append classes to them.

- [ ] **Step 1: Write the failing formatter test**

`frontend/src/lib/format/utils.test.ts`:

```ts
import { describe, expect, it } from 'vitest';
import { formatGhs } from './utils';

describe('formatGhs', () => {
  it('renders whole cedis without decimals (mockup style)', () => {
    expect(formatGhs(48000)).toBe('GHS 480');
    expect(formatGhs(0)).toBe('GHS 0');
  });
  it('keeps pesewas when present instead of lying by rounding', () => {
    expect(formatGhs(48050)).toBe('GHS 480.50');
  });
  it('groups thousands', () => {
    expect(formatGhs(123456700)).toBe('GHS 1,234,567');
  });
});
```

- [ ] **Step 2: Run it to make sure it fails**

Run: `pnpm vitest run src/lib/format/utils.test.ts`
Expected: FAIL — `formatGhs` is not exported.

- [ ] **Step 3: Implement `formatGhs`** (append to `src/lib/format/utils.ts`, keep existing exports untouched)

```ts
/** Mockup-style GHS price: "GHS 480", "GHS 480.50". Input is minor units. */
export function formatGhs(minor: number): string {
  const cedis = minor / 100;
  const hasPesewas = minor % 100 !== 0;
  return `GHS ${cedis.toLocaleString('en-US', {
    minimumFractionDigits: hasPesewas ? 2 : 0,
    maximumFractionDigits: hasPesewas ? 2 : 0,
  })}`;
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm vitest run src/lib/format/utils.test.ts` — Expected: 3 passed.

- [ ] **Step 5: Comment out Tailwind, add dormant notes, create CSS files**

In `globals.css`, replace line 1:

```css
/*
 * Tailwind is deliberately DORMANT. The UI is pure mockup CSS for now;
 * a future CSS→Tailwind migration re-enables this import (see
 * docs/superpowers/specs/2026-07-02-funnel-ui-alignment-design.md §3).
 */
/* @import "tailwindcss"; */
@import './pages.css';
@import './auth.css';
```

Prepend to `tailwind.config.ts`:

```ts
// DORMANT: Tailwind is not wired into the build (no PostCSS/Vite plugin) and
// this config is currently unused. Kept for the future CSS→Tailwind migration.
// See docs/superpowers/specs/2026-07-02-funnel-ui-alignment-design.md §3.
```

Create `src/styles/pages.css` and `src/styles/auth.css`, each starting with a one-line banner comment (`/* Ported from Rue mockup — see spec §3. */`).

- [ ] **Step 6: Verify the app still builds and renders**

Run: `pnpm typecheck && pnpm build`
Expected: both pass (Tailwind import was inert, so the bundle CSS should be byte-similar; spot-check `dist/assets/*.css` still contains `.eyebrow`).

- [ ] **Step 7: Commit**

```bash
git add src/styles src/lib/format tailwind.config.ts
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): dormant tailwind, page CSS files, mockup-style GHS formatter"
```

---

### Task 2: Cart provider owns drawer + toast; drawer restyled to mockup

**Files:**
- Modify: `frontend/src/features/cart/cart-provider.tsx`
- Create: `frontend/src/features/cart/add-toast.tsx`
- Modify: `frontend/src/features/cart/cart-drawer.tsx` (full JSX restyle)
- Modify: `frontend/src/features/shared/layouts/root-layout.tsx`, `frontend/src/features/shared/layouts/header.tsx` (drawer state moves to provider)
- Modify: `frontend/src/styles/pages.css` (drawer + toast + shared button/qty classes)
- Test: `frontend/src/features/cart/add-toast.test.tsx` (new)

**Interfaces:**
- Consumes: existing `useCart()` (`addItem(productId, qty)`, items, totals, mutations), `formatGhs` (Task 1).
- Produces on the cart context (later tasks rely on these exact names): `isDrawerOpen: boolean`, `openDrawer(): void`, `closeDrawer(): void`, `lastAdded: { name: string } | null`, `dismissToast(): void`. `addItem` now also sets `lastAdded` (auto-cleared after 2400 ms) on success.

- [ ] **Step 1: Port the CSS**

Append to `src/styles/pages.css` the following class blocks copied **verbatim** from the mockup CSS. Source of truth: grep each selector in `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/styles.css` and `Rue/pages.css` and copy the whole rule(s): `.drawer-scrim`, `.drawer`, `.drawer.open`, `.drawer-head`, `.drawer-title`, `.drawer-body`, `.drawer-foot`, `.drawer-row`, `.drawer-row.muted`, `.drawer-link`, `.cart-item`, `.cart-item-body`, `.cart-item-brand`, `.cart-item-name`, `.cart-item-meta`, `.cart-item-row`, `.cart-item-remove`, `.cart-empty`, `.qty`, `.qty-lg`, `.icon-btn`, `.icon-btn-lg`, `.btn`, `.btn-primary`, `.toast`, `.ph`, `.ph-label`, `.ph--*` tone variants, `.price`.
**Dedup rule:** before appending each block, grep `globals.css` for the same selector — several (`.btn`, `.ph`, `.price`, `.icon-btn`) may already exist from the home/shop port. If a selector already exists with identical rules, skip it; if it exists with different rules, keep the `globals.css` version and note the diff in your report (do not define the same selector twice).

- [ ] **Step 2: Write the failing toast test**

`frontend/src/features/cart/add-toast.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AddToast } from './add-toast';

describe('AddToast', () => {
  it('renders nothing when there is no recent add', () => {
    const { container } = render(
      <AddToast lastAdded={null} onView={() => {}} onDismiss={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('shows the product name and opens the bag on click', async () => {
    const onView = vi.fn();
    render(<AddToast lastAdded={{ name: 'Rose Serum' }} onView={onView} onDismiss={() => {}} />);
    expect(screen.getByText(/Rose Serum/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: /view bag/i }));
    expect(onView).toHaveBeenCalledOnce();
  });
});
```

If `toBeInTheDocument` is untyped, add `import '@testing-library/jest-dom/vitest';` at the top of the test (the dep exists).

- [ ] **Step 3: Run it to make sure it fails**

Run: `pnpm vitest run src/features/cart/add-toast.test.tsx`
Expected: FAIL — `./add-toast` does not exist.

- [ ] **Step 4: Implement `AddToast`** (`frontend/src/features/cart/add-toast.tsx`; markup from mockup `Rue/src/app.jsx:113-119`)

```tsx
import { Icon } from '../shared/ui/icons';

interface AddToastProps {
  lastAdded: { name: string } | null;
  onView: () => void;
  onDismiss: () => void;
}

export function AddToast({ lastAdded, onView, onDismiss }: AddToastProps) {
  if (!lastAdded) return null;
  return (
    <div className="toast">
      <Icon name="check" size={14} />
      <span>
        <strong>Added.</strong> {lastAdded.name}
      </span>
      <button
        onClick={() => {
          onDismiss();
          onView();
        }}
      >
        View bag
      </button>
    </div>
  );
}
```

(If `check` is missing from the icon map, use `starFilled` and note it.)

- [ ] **Step 5: Run the test to verify it passes**

Run: `pnpm vitest run src/features/cart/add-toast.test.tsx` — Expected: 2 passed.

- [ ] **Step 6: Move drawer + toast state into the provider**

In `cart-provider.tsx`: add to the context value `isDrawerOpen`, `openDrawer`, `closeDrawer`, `lastAdded`, `dismissToast` backed by `useState` (`isDrawerOpen: boolean`, `lastAdded: { name: string } | null`). In `addItem`, after the existing successful add path, resolve the product name (the function already has product data in scope or can take an optional `name` — if not in scope, change the signature to `addItem(productId: string, qty: number, name?: string)` and update all call sites), then:

```ts
setLastAdded(name ? { name } : { name: 'Item' });
window.setTimeout(() => setLastAdded(null), 2400);
```

In `root-layout.tsx`: delete the local `isCartOpen` state; consume `useCart()` for `isDrawerOpen`/`closeDrawer`/`openDrawer`; render `<CartDrawer open={isDrawerOpen} onClose={closeDrawer} />` and, alongside it, `<AddToast lastAdded={lastAdded} onView={openDrawer} onDismiss={dismissToast} />`. In `header.tsx`: the bag button's open handler switches to `openDrawer` from `useCart()` (remove any prop-drilled open callback).

- [ ] **Step 7: Restyle the drawer**

Rewrite `cart-drawer.tsx`'s JSX to the mockup structure (`Rue/src/shared.jsx:181-238`): scrim div + `aside.drawer.open`, `drawer-head` (eyebrow "Your Bag", `drawer-title` "{n} item(s)", `icon-btn` close), body with either the `cart-empty` block (circular `ph`, display-font "Your bag is empty", "Let's change that.", `btn btn-primary` "Shop the edit →" navigating to `/shop` and closing) or `cart-item` rows — image from `image_path` (`<img>` inside the 80×100 box) falling back to `ph ph--{tone}`; brand, name, size, `qty` stepper wired to the existing update mutation, `price` via `formatGhs(line total)`, `cart-item-remove` ×. Foot: subtotal row, muted delivery row, `btn btn-primary` "Checkout · {formatGhs(subtotal)}" → `navigate({ to: '/checkout' })` + close, `drawer-link` "Continue shopping" → close. Keep ALL existing mutation wiring; strip every Tailwind className from the file.

- [ ] **Step 8: Verify**

Run: `pnpm typecheck && pnpm vitest run src/features/cart/`
Expected: clean. Then `pnpm dev`, add an item from the shop page: toast appears bottom (mockup position), "View bag" opens the restyled drawer, qty/remove still work. Kill the server.

- [ ] **Step 9: Commit**

```bash
git add src/features/cart src/features/shared/layouts src/styles/pages.css
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(cart): mockup drawer visuals, add-to-cart toast, provider-owned drawer state"
```

---

### Task 3: Product detail page

**Files:**
- Create: `frontend/src/content/product-copy.ts`
- Modify: `frontend/src/features/catalog/product-detail.tsx` (full rewrite)
- Modify: `frontend/src/styles/pages.css` (PDP classes)
- Reference: `Rue/src/pages.jsx:144-260` (markup), `Rue/pages.css` (classes)

**Interfaces:**
- Consumes: `useGetProductsSlug(slug)` (verify exact hook name in the generated file; the plain function is `getProductsSlug`), `useGetBrands()`, `useGetCategories()`, `useGetProducts(params)` for related, `useCart().addItem(productId, qty, name)` (Task 2 signature), `formatGhs`, existing `ProductCard`.
- Produces: `getProductCopy(categorySlug: string | undefined): ProductCopy` from `product-copy.ts`.

- [ ] **Step 1: Port PDP CSS**

Append to `pages.css` (verbatim from `Rue/pages.css`, same dedup rule as Task 2): `.breadcrumb`, `.product-grid`, `.product-gallery`, `.product-thumbs`, `.product-thumb`, `.product-main`, `.product-info`, `.product-name`, `.product-rating`, `.stars`, `.product-price`, `.price-was`, `.product-size`, `.product-lede`, `.product-actions`, `.product-perks`, `.product-tabs`, `.tab-heads`, `.tab-head`, `.tab-body`, `.ing-list`. Do NOT port `.product-swatches`/`.swatch-row`/`.label` (spec §4.1: swatches omitted).

- [ ] **Step 2: Create the curated copy module**

`frontend/src/content/product-copy.ts`:

```ts
/**
 * Curated editorial copy for PDP lede + tabs, keyed by category slug.
 * Bridge until the backend gains per-product description/how_to/ingredients
 * (spec §8 backend follow-up #1) — API values take precedence when they exist.
 */
export interface ProductCopy {
  lede: string;
  description: string;
  howTo?: string;
  ingredients?: string;
}

const FALLBACK: ProductCopy = {
  lede: 'A considered formulation that delivers on its promise. Tested in-store, recommended by our beauty team, and loved by our regulars.',
  description:
    'A hydrating, lightweight treatment designed for daily use. Its texture absorbs quickly without residue, leaving skin visibly smoother and more even-toned after consistent use. Dermatologically tested. Suitable for sensitive skin types.',
};

const BY_CATEGORY: Record<string, ProductCopy> = {
  skincare: {
    ...FALLBACK,
    howTo:
      'Apply morning and evening to cleansed skin. Gently press a few drops into your face and neck, avoiding the eye area. Follow with moisturiser and — in the morning — SPF 30 or higher.',
    ingredients:
      'Aqua, Glycerin, Niacinamide, Sodium Hyaluronate, Panthenol, Allantoin, Tocopherol, Propanediol, Citric Acid, Sodium Benzoate, Phenoxyethanol, Parfum.',
  },
  haircare: {
    lede: 'Salon-grade care for every texture. Chosen by our stylists, kept on our own shelves.',
    description:
      'A nourishing treatment that restores softness and shine without weighing hair down. Suitable for protective styles and colour-treated hair.',
    howTo:
      'Work a small amount through damp hair from mid-length to ends. Leave in, or rinse after 10 minutes for a deeper treatment.',
  },
  body: {
    lede: 'Everyday body care that feels like a ritual, not a routine.',
    description:
      'Rich, fast-absorbing care that leaves skin supple through harmattan and humidity alike. No residue, no heaviness.',
    howTo: 'Massage into damp skin after bathing. Reapply to hands and elbows as needed.',
  },
  fragrance: {
    lede: 'Scents with a memory. Composed to last through a Ghana afternoon.',
    description:
      'A layered composition that opens bright and settles into a warm, lasting base. Concentrated for longevity.',
    howTo: 'Spray onto pulse points — wrists, neck, behind the ears. Do not rub.',
  },
};

export function getProductCopy(categorySlug: string | undefined): ProductCopy {
  return (categorySlug && BY_CATEGORY[categorySlug]) || FALLBACK;
}
```

(Adjust the four keys to the REAL category slugs in the seed data — check `backend/cmd/seed/data/categories.json`; use its slugs verbatim.)

- [ ] **Step 3: Rewrite `product-detail.tsx`**

Full structure from `Rue/src/pages.jsx:144-260` with these bindings (keep the component's exported name and `slug` prop so `router.tsx` is untouched):

- Data: `const { data: product, isLoading, error } = useGetProductsSlug(slug)`; `useGetCategories()` + `useGetBrands()` for label resolution; related products via `useGetProducts({ category: categorySlug, limit: 5 })` filtered to exclude the current slug, capped at 4, section skipped when empty. Keep existing loading/error render paths (restyle copy with `eyebrow`/muted text, no Tailwind classes).
- Breadcrumb: `wrap breadcrumb` with `Link to="/"`, `Link to="/shop"`, `Link to="/shop" search={{ category: categorySlug }}` (match however the shop page currently encodes its category filter — read `shop-page.tsx` first; if the shop filter is client-state-only and not URL-driven, link to plain `/shop` and note it), then the product name as a span.
- Gallery: no thumb rail (single image). `div.product-main` — if `product.image_path` render `<img src={image_path} alt={name} />`, else `ph ph--{tone ?? 'lavender'}` with `ph-label` = brand name.
- Info column exactly in mockup order; price row: `formatGhs(price_ghs_minor ?? 0)`, `price-was` only when `was_price_ghs_minor` is a number > 0, `· {size}` only when `size` is non-empty; rating row only when `rating` is a number (5 `starFilled` icons per mockup, text `{rating} · {review_count} reviews`).
- Lede: `getProductCopy(categorySlug).lede` + trailing link "Part of our {category label} edit." to the shop category.
- Actions: local `qty` state (min 1); CTA text `Add to bag · {formatGhs((price_ghs_minor ?? 0) * qty)}`; onClick `addItem(product.id ?? '', qty, product.name)`; wishlist heart: `<button className="icon-btn icon-btn-lg" disabled title="Saved items coming soon"><Icon name="heart" /></button>`.
- Perks (`product-perks`): add to `product-copy.ts` an export `export const PERKS: string[]` with three lines: `Free delivery in Accra over GHS {threshold}` (read the real threshold from `backend/config/shipping_config.json` at execution time and hardcode the figure — it's config that changes rarely, and this keeps it in ONE frontend place), `100% authentic, guaranteed`, and `Questions? WhatsApp us`. Render each as an `<li>` with icons `truck`, `shield`, and `sparkle` (if a `whatsapp` glyph is absent from the icon map).
- Tabs: heads Description / How to use / Ingredients, rendering only tabs whose copy exists (`description` always; `howTo`/`ingredients` conditional). Body from `getProductCopy`.
- Strip all Tailwind classNames from the file; remove the `min-h-screen bg-paper …` shell if present (the route already wraps in `pageShell`).

- [ ] **Step 4: Verify**

Run: `pnpm typecheck && pnpm build` — clean. `pnpm dev`: open a product from the shop grid — breadcrumb links work, price/rating/size show real data, was-price appears only on discounted seed products, add-to-bag updates drawer count and fires the toast, disabled heart shows the tooltip, tabs switch. Compare side-by-side against the mockup (`open '/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/Rue Cosmetics.html'`, navigate to a product; needs network for CDN React — if offline, compare against `pages.jsx` structure). Kill the server.

- [ ] **Step 5: Commit**

```bash
git add src/features/catalog/product-detail.tsx src/content/product-copy.ts src/styles/pages.css
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(catalog): mockup-aligned product detail page with curated copy bridge"
```

---

### Task 4: Cart page (extrapolated)

**Files:**
- Modify: `frontend/src/features/cart/cart-page.tsx` (full restyle)
- Modify: `frontend/src/styles/pages.css` (extrapolated cart-page classes)

**Interfaces:**
- Consumes: `useCart()` items/totals/mutations (unchanged), `formatGhs`, `cart-item` classes from Task 2.

- [ ] **Step 1: Add extrapolated CSS**

Append to `pages.css`:

```css
/* ── cart page (extrapolated — no mockup source; follows drawer vocabulary) ── */
.cart-page { padding: 48px 0 96px; }
.cart-page-grid { display: grid; grid-template-columns: 1fr 380px; gap: 48px; align-items: start; }
@media (max-width: 900px) { .cart-page-grid { grid-template-columns: 1fr; } }
.cart-page .cart-item { padding: 20px 0; border-bottom: 1px solid var(--line-soft); }
.cart-summary { background: var(--surface); border-radius: var(--radius-lg); padding: 28px; position: sticky; top: 96px; }
.cart-summary h3 { font-family: var(--font-display); font-size: 22px; font-weight: 400; margin: 0 0 16px; }
.cart-summary .drawer-row { padding: 8px 0; }
.cart-free-note { font-family: var(--font-label); font-size: 12px; color: var(--lavender-700); background: var(--lavender-100); border-radius: var(--radius); padding: 10px 12px; margin-top: 12px; }
```

- [ ] **Step 2: Restyle `cart-page.tsx`**

Read the current file first; keep every hook/mutation/computed value (subtotal, free-shipping remainder). New JSX skeleton:

```tsx
<main className="wrap cart-page fade-up">
  <div className="eyebrow">Your Bag</div>
  <h1 className="h-display" style={{ fontSize: 'clamp(40px, 6vw, 80px)', marginTop: 8 }}>
    Ready when <em>you are.</em>
  </h1>
  {/* empty state: same block as the drawer's cart-empty, centered, with Shop the edit CTA */}
  <div className="cart-page-grid" style={{ marginTop: 40 }}>
    <div>{/* cart-item rows exactly as in the drawer (Task 2 markup), wired to the same mutations */}</div>
    <aside className="cart-summary">
      <h3>Order summary</h3>
      <div className="drawer-row"><span>Subtotal</span><span className="price">{formatGhs(subtotalMinor)}</span></div>
      <div className="drawer-row muted"><span>Delivery</span><span>Calculated at checkout</span></div>
      {/* keep the existing free-shipping remainder, restyled: */}
      {remainderMinor > 0 && (
        <div className="cart-free-note">Add {formatGhs(remainderMinor)} more for free delivery in Accra.</div>
      )}
      <button className="btn btn-primary" style={{ width: '100%', justifyContent: 'center', marginTop: 16 }}
              onClick={() => navigate({ to: '/checkout' })}>
        Checkout · {formatGhs(subtotalMinor)}
      </button>
    </aside>
  </div>
</main>
```

Bind `subtotalMinor`/`remainderMinor` to whatever the current file's equivalents are named (keep their computation). The headline `Ready when you are.` mirrors the mockup's wishlist-page headline pattern (`app.jsx:92-94`). Strip all Tailwind classNames.

- [ ] **Step 3: Verify**

Run: `pnpm typecheck` — clean. `pnpm dev`: `/cart` with items shows rows + sticky summary; empty cart shows the empty block; totals match the drawer; checkout CTA navigates. Kill the server.

- [ ] **Step 4: Commit**

```bash
git add src/features/cart/cart-page.tsx src/styles/pages.css
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(cart): extrapolated mockup-language cart page"
```

---

### Task 5: Checkout + return page (extrapolated)

**Files:**
- Modify: `frontend/src/features/checkout/checkout-page.tsx`, `frontend/src/features/checkout/checkout-return.tsx` (JSX restyle only)
- Modify: `frontend/src/styles/pages.css` (extrapolated checkout classes + `field` form classes)

**Interfaces:**
- Consumes: ALL existing checkout state/handlers (address form state, shipping quote query, `postCheckoutInit` mutation, Paystack redirect, return-page verify polling), `formatGhs`, `field` classes (shared with auth — define here, Task 6 reuses).

- [ ] **Step 1: Add CSS**

Append to `pages.css`: first the `field` form classes copied **verbatim** from `Rue/account.css` (grep `.field` — label + input/select/textarea styling), then:

```css
/* ── checkout (extrapolated) ── */
.checkout-page { max-width: 720px; margin: 0 auto; padding: 48px var(--gut) 96px; }
.checkout-section { border-top: 1px solid var(--line); padding: 32px 0; }
.checkout-section h2 { font-family: var(--font-display); font-size: 26px; font-weight: 400; margin: 0 0 20px; }
.checkout-methods { display: grid; gap: 12px; }
.method-card { display: flex; justify-content: space-between; align-items: center; border: 1px solid var(--line); border-radius: var(--radius-lg); padding: 16px 20px; cursor: pointer; transition: border-color var(--dur) var(--ease), background var(--dur) var(--ease); font-family: var(--font-label); }
.method-card.selected { border-color: var(--lavender-600); background: var(--surface); }
.method-card .price { font-weight: 600; }
.checkout-summary-rows .drawer-row { padding: 6px 0; }
.checkout-status-card { max-width: 560px; margin: 96px auto; background: var(--surface); border-radius: var(--radius-lg); padding: 48px; text-align: center; }
.checkout-status-card h1 { font-family: var(--font-display); font-weight: 400; font-size: clamp(32px, 5vw, 56px); margin: 0 0 12px; }
.checkout-status-card p { color: var(--ink-muted); margin: 0 0 24px; }
```

- [ ] **Step 2: Restyle `checkout-page.tsx`**

Read the current file first and inventory its state/handlers — they all stay. Replace the JSX with: `main.checkout-page.fade-up` → eyebrow "Checkout" + `h-display` headline ("Almost <em>there.</em>") → three `checkout-section`s:
1. **Delivery address** — existing form fields re-rendered as `div.field` (label + input) in the current field order; keep names, values, onChange handlers, and any validation/error rendering (errors render as muted `--lavender-700`-tinted text under the field).
2. **Delivery method** — map the existing quote/method options to `label.method-card` (+`.selected` when chosen) each containing the method name and `formatGhs(priceMinor)`; the underlying radio input stays (visually hidden is fine: `style={{ position: 'absolute', opacity: 0 }}`) so keyboard/form semantics keep working.
3. **Order summary** — existing items/totals as `checkout-summary-rows` drawer-rows (item name × qty | line total; subtotal; delivery; total row with `price` class).
Then the CTA: `btn btn-primary` full-width, `Pay with Paystack · {formatGhs(totalMinor)}`, wired to the existing submit handler with its existing disabled/pending states (pending text: "Preparing secure payment…"). Strip all Tailwind classNames.

- [ ] **Step 3: Restyle `checkout-return.tsx`**

Keep the existing verify-polling logic and its three states, re-rendered as one `checkout-status-card`:
- success: `h1` "Thank you." + `p` "Order {reference} is confirmed. A receipt is on its way to your inbox." + `btn btn-primary` "View your orders" → `/account/orders`;
- pending: `h1` "One moment." + `p` "Confirming your payment with Paystack…" (keep the poll);
- failed: `h1` "Payment didn't complete." + `p` existing error copy + `btn btn-primary` "Back to checkout" → `/checkout`.
Strip Tailwind classNames.

- [ ] **Step 4: Verify**

Run: `pnpm typecheck` — clean. `pnpm dev` with the backend up if available (`make up && make dev` from repo root): fill the address, pick a method (card highlights, quoted price real), summary totals correct; if Paystack test key configured, CTA redirects. Without a backend, verify render + interaction states only and note it. Kill servers.

- [ ] **Step 5: Commit**

```bash
git add src/features/checkout src/styles/pages.css
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(checkout): extrapolated mockup-language checkout and return pages"
```

---

### Task 6: Auth shell + five auth pages

**Files:**
- Create: `frontend/src/features/auth/auth-shell.tsx`
- Modify: `frontend/src/features/auth/login-page.tsx`, `signup-page.tsx`, `forgot-password-page.tsx`, `reset-password-page.tsx`, `verify-email-page.tsx` (restyle, keep handlers)
- Modify: `frontend/src/router.tsx` (auth routes move to a pathless `_auth` layout)
- Modify: `frontend/src/styles/auth.css` (ported auth classes)
- Reference: `Rue/src/acct-auth.jsx` (all 141 lines), `Rue/account.css`

**Interfaces:**
- Consumes: existing page handlers (`useAuth().login/signup`, forgot/reset/verify API calls), `field` classes (Task 5), `Icon`.
- Produces: `AuthShell({ title, sub, children, footer }: { title: ReactNode; sub: string; children: ReactNode; footer?: ReactNode })`.

- [ ] **Step 1: Port auth CSS**

Append to `src/styles/auth.css`, verbatim from `Rue/account.css` (same dedup rule): `.auth-wrap`, `.auth-visual`, `.auth-visual-copy`, `.auth-form-wrap`, `.auth-form`, `.auth-social`, `.auth-divider`, `.auth-meta`, `.auth-link`, `.back-link`. (`.field` already landed in `pages.css` in Task 5 — do not duplicate.)

- [ ] **Step 2: Create `auth-shell.tsx`** (markup from `acct-auth.jsx:5-25`, back-link becomes a router Link)

```tsx
import type { ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';

interface AuthShellProps {
  title: ReactNode;
  sub: string;
  children: ReactNode;
  footer?: ReactNode;
}

export function AuthShell({ title, sub, children, footer }: AuthShellProps) {
  return (
    <div className="auth-wrap">
      <div className="auth-visual">
        <div className="ph ph--lavender"><span className="ph-label">editorial · lifestyle</span></div>
        <div className="auth-visual-copy">
          <div className="eyebrow" style={{ color: 'var(--lavender-300)' }}>Rue · Members</div>
          <h2>
            Small rituals,<br />
            <em style={{ fontFamily: 'var(--font-serif)', fontStyle: 'italic', color: 'var(--lavender-300)' }}>
              long kept.
            </em>
          </h2>
          <p>Track your orders, build your rituals, and earn points toward the shelf you've always wanted.</p>
        </div>
      </div>
      <div className="auth-form-wrap">
        <div className="auth-form">
          <Link className="back-link" to="/"><Icon name="arrowLeft" size={12} /> Back to Rue</Link>
          <div className="eyebrow" style={{ color: 'var(--lavender-700)' }}>{sub}</div>
          <h1>{title}</h1>
          {children}
          {footer}
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Move auth routes into a pathless `_auth` layout**

In `router.tsx`, add after the `_checkout` layout definition:

```tsx
const authLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_auth',
  component: () => <Outlet />,
});
```

Change the five auth routes' `getParentRoute` from `storefrontLayoutRoute` to `authLayoutRoute` (paths unchanged), remove them from `storefrontLayoutRoute.addChildren([...])`, and add to the tree: `authLayoutRoute.addChildren([loginRoute, signupRoute, forgotPasswordRoute, resetPasswordRoute, verifyEmailRoute])` as a sibling of the storefront/admin/checkout subtrees. The `redirectPathFor` guard target `/login` still resolves (path unchanged).

- [ ] **Step 4: Restyle the five pages**

Read each page first; keep its state, submit handler, error/success rendering, and navigation. Login (from `acct-auth.jsx:27-58`):

```tsx
<AuthShell
  title="Welcome back."
  sub="Sign in"
  footer={<div className="auth-meta">New here? <Link to="/signup">Create an account</Link></div>}
>
  <div className="auth-social">
    <button type="button" onClick={/* existing Google start: window.location.href to the /auth/google/start URL the current page uses */}>
      <Icon name="user" size={14} /> Continue with Google
    </button>
  </div>
  <div className="auth-divider">or with email</div>
  <form onSubmit={/* existing handler */}>
    <div className="field" style={{ marginBottom: 16 }}>
      <label>Email</label>
      <input type="email" value={email} onChange={…} placeholder="you@somewhere.com" />
    </div>
    <div className="field" style={{ marginBottom: 8 }}>
      <label>Password</label>
      <input type={show ? 'text' : 'password'} value={password} onChange={…} />
    </div>
    {/* existing error message, className="auth-meta" with color var(--lavender-700) */}
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
      <label style={{ display: 'inline-flex', gap: 8, fontFamily: 'var(--font-label)', fontSize: 12, color: 'var(--ink-soft)' }}>
        <input type="checkbox" onChange={(e) => setShow(e.target.checked)} /> Show password
      </label>
      <Link className="auth-link" to="/forgot-password">Forgot password?</Link>
    </div>
    <button type="submit" className="btn btn-primary" style={{ width: '100%', justifyContent: 'center' }} disabled={/* existing pending state */}>
      Sign in <Icon name="arrow" size={14} />
    </button>
  </form>
</AuthShell>
```

(No Apple button — spec §4.4. If the current login page has no Google button, add it only if a Google start URL already exists in the codebase — grep `google/start`; otherwise omit the social row entirely and note it.)

The other four use the same shell with their existing fields/handlers:
- **Signup** — title "Begin your ritual.", sub "Create account", fields Name (if the current page collects it) / Email / Password, footer "Already have an account? → /login".
- **Forgot** — title "No worries.", sub "Reset password", single Email field, CTA "Send reset link"; keep the existing submitted-state message as `auth-meta`.
- **Reset** — title "Choose a new one.", sub "Reset password", Password (+ confirm if present), keep token handling.
- **Verify** — title "Check your inbox." (or the current page's states: verifying / success / failure), sub "Verify email"; keep the existing token-verification logic and its state renders, each as simple `auth-meta`/`p` copy inside the shell.

Strip all Tailwind classNames from the five pages.

- [ ] **Step 5: Verify**

Run: `pnpm typecheck && pnpm vitest run` — clean (guard tests unaffected: `/login` path unchanged). `pnpm dev`: `/login` renders the two-panel shell with no storefront header; login round-trip works against a running backend (or note backend-down); links between auth pages navigate; "Back to Rue" returns home. Kill servers.

- [ ] **Step 6: Commit**

```bash
git add src/features/auth src/router.tsx src/styles/auth.css
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(auth): mockup AuthShell layout for all five auth pages"
```

---

### Task 7: Final gate, dead-class sweep, visual pass

**Files:** none new (fixes only).

- [ ] **Step 1: Full gate**

Run from `frontend/`:

```bash
pnpm typecheck && pnpm lint && pnpm vitest run && pnpm build
```

Expected: all pass. Fix anything that fails (lint may flag unused imports left by the restyles).

- [ ] **Step 2: Dead Tailwind-class sweep on tranche files**

```bash
grep -nE 'className="[^"]*\b(flex|grid|px-|py-|mb-|mt-|text-(sm|xs|lg|xl)|bg-|rounded|items-center|justify-)' \
  src/features/cart src/features/checkout src/features/auth src/features/catalog/product-detail.tsx \
  src/features/shared/layouts/root-layout.tsx -r
```

Expected: no hits (heuristic — inspect any hit and remove it if it's a Tailwind utility; inline `style` and mockup classes are fine).

- [ ] **Step 3: Visual pass**

`pnpm dev` (+ backend via `make up && make dev` from repo root when Docker is available). Walk the funnel: shop → product → add (toast) → drawer → `/cart` → `/checkout` → (test-key payment or stop at redirect) → `/checkout/return`; then `/login`, `/signup`, `/forgot-password`. Screenshot each against the mockup where one exists. Record discrepancies in the report; fix layout-level ones, note pixel-level ones for the human's final eyeball.

- [ ] **Step 4: Commit any fixes**

```bash
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -am "fix(frontend): funnel UI polish from final visual pass"
```

(Skip if the tree is clean.)
