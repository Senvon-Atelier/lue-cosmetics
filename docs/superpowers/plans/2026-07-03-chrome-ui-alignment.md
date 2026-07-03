# Global Chrome UI Alignment (Tranche 4) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the missing global chrome — SearchOverlay wired to the products `q` param and a MobileMenu — plus URL-driven shop category filters (which also fixes a live bug: the shop page sends category **ids** to an API that filters by **slug**, so category chips currently return zero products), and three folded backlog items (drawer a11y, shared CartItemRow, Google glyph).

**Architecture:** Both new surfaces are Header-owned (local `useState`), rendered by Header, styled by mockup CSS ported into `globals.css` (their classes come from the mockup's global `styles.css`/`pages.css` chrome blocks). Search uses the generated `useGetProducts` hook with a debounced `q`. Category filtering standardizes on slugs end-to-end: URL `?category=<slug>` → shop state → API param. Closed overlays get the native `inert` attribute (React 18 types augmented).

**Tech Stack:** React 18, TanStack Router v1 (code-based; `validateSearch` on the shop route), Orval-generated hooks/functions (`useGetProducts`, `getCategories`), plain CSS, vitest (+ @testing-library/react, jsdom pragma — same setup as `add-toast.test.tsx`).

**Spec:** `docs/superpowers/specs/2026-07-03-chrome-ui-alignment-design.md` — binding. **Mockup source (read-only, NEVER modify):** `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/` (`src/shared.jsx` 244–382, `styles.css` drawer/search-overlay blocks, `pages.css` SEARCH + MOBILE NAV blocks).

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."`.
- Frontend-only; nothing under `backend/` or `frontend/src/lib/api/generated/` changes.
- Functionality freeze except the spec's sanctioned changes: shop category identity switches id→slug (bug fix, spec §3.5); `wishlistCount` removed from the cart provider; heart button disabled. Everything else (cart mutations, PDP fetches, auth flows, existing shop q/brand/sort behavior) keeps working.
- Honest data: no fake counts (`.mnav-count` NOT ported), no fake popularity ("From the shop" label, plain product list), search errors render honest copy.
- Prices `formatGhs`, images `getImageUrl` with `ph` fallback, icons from `features/shared/ui/icons.tsx` (`search`, `close`, `menu`, `chevronRight`, `pin`, `phone`, `clock`, `minus`, `plus`, `heart`, `arrow` all exist).
- Class parity with the mockup; adapted rules under `/* adapted */` banners; every new button selector defines its own background/border; dedup rule (chip/eyebrow/wrap/icon-btn/drawer-scrim/drawer classes exist — do not redefine).
- Actionable elements are `<button>`/`<Link>` (no bare `<a onClick>`); closed overlays are `inert` and Esc-closable.
- Gate per task: `pnpm typecheck` zero errors. Full gate (typecheck/lint/vitest/build) in the final task.
- Commands run from `ruecosmetics/frontend/` unless stated.

---

### Task 1: Foundation — chrome CSS, content consts, inert typing, hooks

**Files:**
- Modify: `frontend/src/styles/globals.css` (append chrome blocks)
- Create: `frontend/src/content/search-terms.ts`, `frontend/src/content/store-info.ts`
- Modify: `frontend/src/vite-env.d.ts` (inert augmentation)
- Create: `frontend/src/features/shared/use-overlay.ts`
- Create: `frontend/src/lib/hooks/use-debounced-value.ts`
- Test: `frontend/src/lib/hooks/use-debounced-value.test.tsx` (new)

**Interfaces:**
- Produces: `SEARCH_TERMS: readonly string[]`; `STORE_INFO = { addressLine1, addressLine2, phone, hours }`; `useEscToClose(open: boolean, onClose: () => void): void`; `useLockBodyScroll(open: boolean): void`; `useDebouncedValue<T>(value: T, delayMs?: number): T`; global `inert?: ''` prop on all HTML elements; CSS classes `search-*`, `drawer-left`, `mobile-nav`, `mnav-section`, `mnav-contact`, `.mobile-menu-btn` responsive rules.

- [ ] **Step 1: Write the failing debounce test** — `src/lib/hooks/use-debounced-value.test.tsx`:

```tsx
/** @vitest-environment jsdom */
import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useDebouncedValue } from './use-debounced-value';

describe('useDebouncedValue', () => {
  it('holds the old value until the delay elapses', () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHook(({ v }) => useDebouncedValue(v, 300), {
      initialProps: { v: 'a' },
    });
    rerender({ v: 'ab' });
    expect(result.current).toBe('a');
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(result.current).toBe('ab');
    vi.useRealTimers();
  });
});
```

(If `renderHook` is missing from the installed @testing-library/react version, check `package.json` — v13.1+ has it; the repo's add-toast test already uses @testing-library/react.)

- [ ] **Step 2: Run it to make sure it fails**

Run: `pnpm vitest run src/lib/hooks/use-debounced-value.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement the hook** — `src/lib/hooks/use-debounced-value.ts`:

```ts
import { useEffect, useState } from 'react';

/** Returns `value`, but only after it has been stable for `delayMs`. */
export function useDebouncedValue<T>(value: T, delayMs = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(t);
  }, [value, delayMs]);
  return debounced;
}
```

- [ ] **Step 4: Run test to verify it passes** — `pnpm vitest run src/lib/hooks/use-debounced-value.test.tsx` → PASS.

- [ ] **Step 5: Create the content consts**

`src/content/search-terms.ts`:

```ts
/** Curated editorial search suggestions for the search overlay (mockup's "Trending searches"). */
export const SEARCH_TERMS = [
  'Hyaluronic acid',
  'Shea butter',
  'SPF 50',
  'Lip tint',
  'Fragrance',
  'Hair oil',
] as const;
```

`src/content/store-info.ts`:

```ts
/** Physical store details — single source for footer + mobile menu. */
export const STORE_INFO = {
  addressLine1: 'Community 18, Spintex',
  addressLine2: 'Adjacent KFC, Accra',
  phone: '0594 701 345',
  hours: 'Mon–Sat · 9am – 8pm',
} as const;
```

- [ ] **Step 6: Augment React types for `inert`** — append to `src/vite-env.d.ts`:

```ts
// React 18's types lack the native `inert` attribute. '' = inert, undefined = interactive.
declare module 'react' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface HTMLAttributes<T> {
    inert?: '';
  }
}
```

(If the eslint-disable is unnecessary per lint output, drop it — lint runs with `--report-unused-disable-directives`.)

- [ ] **Step 7: Create the overlay hooks** — `src/features/shared/use-overlay.ts`:

```ts
import { useEffect } from 'react';

/** Close on Escape while open. */
export function useEscToClose(open: boolean, onClose: () => void) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);
}

/** Prevent body scroll while an overlay is open; restores the previous value. */
export function useLockBodyScroll(open: boolean) {
  useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = prev;
    };
  }, [open]);
}
```

- [ ] **Step 8: Append the chrome CSS to `globals.css`** (verbatim port + adapted):

```css
/* ----- Search overlay (ported from Rue/styles.css + Rue/pages.css SEARCH block) ----- */
.search-overlay {
  position: fixed;
  inset: 0;
  background: rgba(255, 255, 255, 0.97);
  backdrop-filter: blur(16px);
  z-index: 120;
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--dur) var(--ease);
  display: flex;
  flex-direction: column;
}
.search-overlay.open { opacity: 1; pointer-events: auto; }
.search-head {
  display: flex;
  gap: 16px;
  align-items: center;
  padding-top: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--line);
}
.search-input-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 14px;
  color: var(--ink-muted);
  padding: 12px 4px;
}
.search-input {
  flex: 1;
  border: 0;
  background: transparent;
  outline: none;
  font-family: var(--font-display);
  font-size: clamp(24px, 4vw, 40px);
  color: var(--ink);
  letter-spacing: -0.01em;
  font-style: italic;
  font-weight: 400;
}
.search-input::placeholder { color: var(--ink-muted); opacity: 0.6; }
.search-clear { color: var(--ink-muted); width: 32px; height: 32px; border-radius: 999px; display: grid; place-items: center; }
.search-clear:hover { background: var(--lavender-100); color: var(--ink); }
.search-body { padding-top: 48px; flex: 1; overflow-y: auto; padding-bottom: 80px; }
.search-section { margin-bottom: 48px; }
.search-section .eyebrow { margin-bottom: 20px; display: block; }
.search-chips { display: flex; gap: 10px; flex-wrap: wrap; }
.search-picks { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.search-pick { display: flex; gap: 16px; cursor: pointer; padding: 12px; border-radius: var(--radius); transition: background var(--dur) var(--ease); }
.search-pick:hover { background: var(--lavender-50); }
.search-empty { padding: 48px 0; font-family: var(--font-display); font-size: 22px; color: var(--ink-soft); }
.search-empty a { border-bottom: 1px solid; cursor: pointer; }
@media (max-width: 720px) { .search-picks { grid-template-columns: 1fr; } }

/* ----- Mobile nav drawer (ported; .mnav-count omitted — no category counts exist) ----- */
.drawer-left { right: auto; left: 0; transform: translateX(-100%); width: min(380px, 88vw); }
.drawer-left.open { transform: translateX(0); }
.mobile-nav a {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  font-family: var(--font-display);
  font-size: 22px;
  color: var(--ink);
  border-bottom: 1px solid var(--line-soft);
  cursor: pointer;
}
.mobile-nav a.active { color: var(--lavender-700); }
.mobile-nav .mnav-section { margin-bottom: 32px; }
.mobile-nav .eyebrow { display: block; margin-bottom: 12px; }
.mnav-contact a { border: 0; padding: 4px 0; font-size: 14px; font-family: var(--font-body); }
.mnav-contact p { font-size: 14px; color: var(--ink-soft); display: flex; gap: 10px; align-items: flex-start; margin: 10px 0; }
.mnav-contact svg { color: var(--lavender-700); margin-top: 4px; flex-shrink: 0; }

/* adapted: chrome buttons need their own chrome (no global button reset) */
button.search-pick { background: none; border: 0; text-align: left; font: inherit; color: inherit; }
.search-clear { background: none; border: 0; }
/* adapted: header menu button — hidden on desktop, shown when the nav collapses */
.mobile-menu-btn { display: none; }
```

Then the responsive rule: FIRST check whether `globals.css` already collapses `.header-nav` on small screens (`grep -n "header-nav" src/styles/globals.css` and look for an existing `@media` rule). If a collapse rule exists at some breakpoint, add `.mobile-menu-btn { display: <same display value as .header-icon-btn — check its rule, likely inline-flex or grid>; }` inside that same media block. If none exists, add (adapted):

```css
@media (max-width: 720px) {
  .header-nav { display: none; }
  .mobile-menu-btn { display: inline-flex; }
}
```

(match `.header-icon-btn`'s display value exactly — read its rule first).

- [ ] **Step 9: Verify** — `pnpm typecheck && pnpm vitest run && pnpm build`. All green; `grep -c "search-overlay" dist/assets/*.css` ≥ 1. Dedup check: `for c in search-overlay drawer-left mobile-nav mnav-contact; do echo -n "$c: "; grep -l "\.$c" src/styles/*.css | wc -l; done` → all 1.

- [ ] **Step 10: Commit**

```bash
git add src/styles/globals.css src/content src/vite-env.d.ts src/features/shared/use-overlay.ts src/lib/hooks
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): chrome foundation — search/mnav CSS, store info + search terms, inert typing, overlay + debounce hooks"
```

---

### Task 2: SearchOverlay + header search wiring

**Files:**
- Create: `frontend/src/features/shared/search-overlay.tsx`
- Modify: `frontend/src/features/shared/layouts/header.tsx` (search button wiring only — menu/heart come in Task 4)

**Interfaces:**
- Consumes: Task 1's `SEARCH_TERMS`, `useDebouncedValue`, `useEscToClose`, `useLockBodyScroll`, `inert` typing, CSS. Generated `useGetProducts(params?, options?)` (same hook the PDP uses for related products) — response `{ items?: InternalCatalogProductView[] }`; product fields `id, slug, name, price_ghs_minor, image_path, tone`.
- Produces: `SearchOverlay({ open, onClose }: { open: boolean; onClose: () => void })`.

- [ ] **Step 1: Create `search-overlay.tsx`**

```tsx
import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useGetProducts } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, getImageUrl } from '../../lib/format/utils';
import { SEARCH_TERMS } from '../../content/search-terms';
import { useDebouncedValue } from '../../lib/hooks/use-debounced-value';
import { Icon } from './ui/icons';
import { useEscToClose, useLockBodyScroll } from './use-overlay';

// Ported from Rue/src/shared.jsx SearchOverlay (lines 244–335); real data via GET /products?q=.
// Idle sections: curated trending chips + honest "From the shop" rail (no fake popularity).

interface SearchOverlayProps {
  open: boolean;
  onClose: () => void;
}

export function SearchOverlay({ open, onClose }: SearchOverlayProps) {
  const [q, setQ] = useState('');
  const debouncedQ = useDebouncedValue(q, 300);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  useEscToClose(open, onClose);
  useLockBodyScroll(open);

  useEffect(() => {
    if (!open) return;
    const t = setTimeout(() => inputRef.current?.focus(), 100);
    return () => clearTimeout(t);
  }, [open]);

  const term = debouncedQ.trim();

  const { data: resultsData, isLoading: searching, error: searchError } = useGetProducts(
    { q: term, limit: 6 },
    { query: { enabled: open && term.length > 0 } },
  );
  const { data: idleData } = useGetProducts(
    { limit: 3 },
    { query: { enabled: open && term.length === 0 } },
  );
  const results = resultsData?.items ?? [];
  const idlePicks = idleData?.items ?? [];

  const openProduct = (slug?: string) => {
    if (!slug) return;
    onClose();
    setQ('');
    void navigate({ to: `/shop/${slug}` });
  };

  return (
    <div className={`search-overlay${open ? ' open' : ''}`} inert={open ? undefined : ''}>
      <div className="search-head wrap">
        <div className="search-input-wrap">
          <Icon name="search" size={20} />
          <input
            ref={inputRef}
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search products, brands, rituals..."
            className="search-input"
          />
          {q && (
            <button onClick={() => setQ('')} className="search-clear" aria-label="Clear search">
              <Icon name="close" size={14} />
            </button>
          )}
        </div>
        <button className="icon-btn" onClick={onClose} aria-label="Close search">
          <Icon name="close" />
        </button>
      </div>
      <div className="search-body wrap">
        {term.length === 0 ? (
          <>
            <div className="search-section">
              <div className="eyebrow">Trending searches</div>
              <div className="search-chips">
                {SEARCH_TERMS.map((t) => (
                  <button key={t} className="chip" onClick={() => setQ(t)}>
                    {t}
                  </button>
                ))}
              </div>
            </div>
            {idlePicks.length > 0 && (
              <div className="search-section">
                <div className="eyebrow">From the shop</div>
                <div className="search-picks">
                  {idlePicks.map((p) => (
                    <SearchPick key={p.id} product={p} onOpen={openProduct} />
                  ))}
                </div>
              </div>
            )}
          </>
        ) : searchError ? (
          <div className="search-empty">
            <p>Search is unavailable right now. Please try again in a moment.</p>
          </div>
        ) : searching ? (
          <div className="search-empty">
            <p>Searching…</p>
          </div>
        ) : results.length === 0 ? (
          <div className="search-empty">
            <p>
              No results for <em>"{term}"</em>. Try a different term or browse{' '}
              <Link to="/shop" onClick={onClose}>
                the full shop
              </Link>
              .
            </p>
          </div>
        ) : (
          <div className="search-section">
            <div className="eyebrow">
              {results.length} result{results.length === 1 ? '' : 's'}
            </div>
            <div className="search-picks">
              {results.map((p) => (
                <SearchPick key={p.id} product={p} onOpen={openProduct} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function SearchPick({
  product,
  onOpen,
}: {
  product: InternalCatalogProductView;
  onOpen: (slug?: string) => void;
}) {
  return (
    <button type="button" className="search-pick" onClick={() => onOpen(product.slug)}>
      {product.image_path ? (
        <img
          src={getImageUrl(product.image_path)}
          alt=""
          style={{ width: 64, height: 80, objectFit: 'cover', borderRadius: 'var(--radius)', flexShrink: 0 }}
          loading="lazy"
        />
      ) : (
        <div className={`ph ph--${product.tone ?? 'lavender'}`} style={{ width: 64, height: 80, flexShrink: 0 }}>
          <span className="ph-label" style={{ fontSize: 8 }}>
            {product.name?.slice(0, 2)}
          </span>
        </div>
      )}
      <div>
        <div className="cart-item-name">{product.name}</div>
        <div className="price" style={{ marginTop: 4 }}>
          {formatGhs(product.price_ghs_minor ?? 0)}
        </div>
      </div>
    </button>
  );
}
```

Adaptation notes baked in: the mockup's `cart-item-brand` line is omitted (product view carries `brand_id`, not a name — cart rows already omit it; backend enrichment is a logged follow-up). If the generated hook's exact name/signature differs (check `grep -n "export function useGetProducts" src/lib/api/generated/rueCosmeticsAPI.ts` and how `product-detail.tsx` calls it), match the PDP's call style exactly.

- [ ] **Step 2: Wire the header search button** — in `header.tsx` add state and render:

```tsx
import { useState } from 'react';
// ...existing imports...
import { SearchOverlay } from '../search-overlay';

export function Header() {
  const { isAuthenticated } = useAuth();
  const { itemCount, wishlistCount, openDrawer } = useCart();
  const [searchOpen, setSearchOpen] = useState(false);
  // ...
```

Search button becomes:

```tsx
<button className="header-icon-btn" aria-label="Search" onClick={() => setSearchOpen(true)}>
  <Icon name="search" size={20} />
</button>
```

And before the closing `</header>` tag add:

```tsx
<SearchOverlay open={searchOpen} onClose={() => setSearchOpen(false)} />
```

(Leave `wishlistCount`, heart, and menu button untouched — Task 4 handles them.)

- [ ] **Step 3: Typecheck + tests** — `pnpm typecheck` zero errors; `pnpm vitest run` all pass.

- [ ] **Step 4: Commit**

```bash
git add src/features/shared/search-overlay.tsx src/features/shared/layouts/header.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): search overlay — debounced live product search, curated idle state, header wiring"
```

---

### Task 3: URL-driven shop category filter (+ slug bug fix, PDP deep-links)

**Files:**
- Modify: `frontend/src/router.tsx` (shop route gains `validateSearch`)
- Modify: `frontend/src/features/catalog/shop-page.tsx`
- Modify: `frontend/src/features/catalog/filter-bar.tsx` (category identity id→slug — read the file and change every category compare/emit consistently)
- Modify: `frontend/src/features/catalog/product-detail.tsx` (breadcrumb + lede links)

**Interfaces:**
- Consumes: nothing new.
- Produces: shop route search schema `{ category?: string }` (slug). Task 4's mobile menu links depend on `Link to="/shop" search={{ category: slug }}` typechecking.

**Background (why this fixes a live bug):** the backend's products `category` query param is a SLUG (`backend/internal/catalog/handler.go:118` — `@Param category query string false "Category slug"`; `handler.go:160` — `CategorySlug: q.Get("category")`). The shop page currently passes `category.id` (`shop-page.tsx:117`), so selecting a category chip filters to zero products. The PDP's related-products call already passes slugs and works. This task standardizes on slugs.

- [ ] **Step 1: Add `validateSearch` to the shop route** in `router.tsx` (the route with `path: '/shop'` parented on `storefrontLayoutRoute`):

```tsx
const shopRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/shop',
  validateSearch: (search: Record<string, unknown>): { category?: string } => ({
    category:
      typeof search.category === 'string' && search.category !== ''
        ? search.category
        : undefined,
  }),
  component: ShopPage,
});
```

(Keep whatever component wrapper the route currently uses; only add `validateSearch`. If ShopPage is invoked with `initialCategory/initialBrand` props anywhere, note it — Step 2 removes those props.)

- [ ] **Step 2: Sync shop page state with the URL** — in `shop-page.tsx`:

1. Drop the `ShopPageProps` interface and the `initialCategory`/`initialBrand` props (they were only a dead prop-drill; verify with `grep -rn "initialCategory" src/` that no caller passes them — remove any that do).
2. Read the param and sync (add imports `useNavigate, useSearch` from `@tanstack/react-router`):

```tsx
export function ShopPage() {
  const { category: categoryParam } = useSearch({ from: '/_storefront/shop' });
  const navigate = useNavigate();
  // selectedCategory now holds the category SLUG (the API filters by slug — see handler.go:118)
  const [selectedCategory, setSelectedCategory] = useState<string | null>(categoryParam ?? null);
  // ...other state unchanged...

  // URL is the source of truth: back/forward + deep links update the filter
  useEffect(() => {
    setSelectedCategory(categoryParam ?? null);
  }, [categoryParam]);

  const handleCategoryChange = (slug: string | null) => {
    setSelectedCategory(slug);
    void navigate({
      to: '/shop',
      search: slug ? { category: slug } : {},
      replace: true,
    });
  };
```

3. Category chips: `onClick={() => handleCategoryChange(category.slug)}` and active compare `selectedCategory === category.slug` (was `category.id` — the bug).
4. The products fetch (`params.category = selectedCategory`) is unchanged — it now receives a slug, which is what the API expects.

- [ ] **Step 3: Update `filter-bar.tsx`** — read the file; wherever it emits or compares the selected category (option values, radio/link onClick args), switch `category.id` → `category.slug`. Do not touch brand filtering (stays id/client-side as-is). No markup changes.

- [ ] **Step 4: PDP deep-links** — in `product-detail.tsx`: replace the two plain `/shop` category links and the stale comment (lines ~84–93 and ~138):

```tsx
// Category links deep-link via /shop?category=<slug> (URL-driven filter, tranche 4).
```

Breadcrumb:

```tsx
<Link to="/shop" search={categorySlug ? { category: categorySlug } : {}}>
  {categoryLabel ?? 'Category'}
</Link>
```

Lede category link (~line 138): same `search` prop on its existing `Link`.
(The plain "Shop" breadcrumb link keeps `search={{}}`— add it if typecheck demands a search prop after `validateSearch` lands; TanStack usually allows omitting optional search.)

- [ ] **Step 5: Typecheck + tests** — `pnpm typecheck` zero errors (this sweeps every `Link to="/shop"` in the app — fix any that now need a `search` prop by adding `search={{}}`); `pnpm vitest run` all pass.

- [ ] **Step 6: Commit**

```bash
git add src/router.tsx src/features/catalog
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): URL-driven shop category filter — slug end-to-end (fixes id-vs-slug dead filter), PDP deep-links"
```

---

### Task 4: MobileMenu + header/footer/provider changes

**Files:**
- Create: `frontend/src/features/shared/mobile-menu.tsx`
- Modify: `frontend/src/features/shared/layouts/header.tsx` (menu wiring, heart disabled, wishlistCount removed)
- Modify: `frontend/src/features/shared/layouts/footer.tsx` (contact strings → `STORE_INFO`)
- Modify: `frontend/src/features/cart/cart-provider.tsx` (remove dead `wishlistCount` state + interface field)

**Interfaces:**
- Consumes: `STORE_INFO`, `useEscToClose`, `useLockBodyScroll`, `inert`, `getCategories()` (plain generated fn returning `InternalCatalogCategoryView[]` — `{ id?, label?, slug? }`), shop route search schema from Task 3, `Brand`, `Icon`.
- Produces: `MobileMenu({ open, onClose })`.

- [ ] **Step 1: Create `mobile-menu.tsx`**

```tsx
import { useEffect, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { getCategories } from '../../lib/api/generated/rueCosmeticsAPI';
import { STORE_INFO } from '../../content/store-info';
import { Brand } from './ui/brand';
import { Icon } from './ui/icons';
import { useEscToClose, useLockBodyScroll } from './use-overlay';

// Ported from Rue/src/shared.jsx MobileMenu (lines 336–382).
// No category counts (.mnav-count) — the API has none. Contact copy shared with the footer.

interface MobileMenuProps {
  open: boolean;
  onClose: () => void;
}

type CategoryLink = { id?: string; label?: string; slug?: string };

export function MobileMenu({ open, onClose }: MobileMenuProps) {
  const [categories, setCategories] = useState<CategoryLink[]>([]);

  useEscToClose(open, onClose);
  useLockBodyScroll(open);

  // Fetch categories the first time the menu opens (not on every page load)
  useEffect(() => {
    if (!open || categories.length > 0) return;
    getCategories()
      .then((cats) => setCategories(cats ?? []))
      .catch(() => {
        /* categories section simply stays empty — nav links still work */
      });
  }, [open, categories.length]);

  return (
    <>
      <div className={`drawer-scrim${open ? ' open' : ''}`} onClick={onClose} />
      <aside className={`drawer drawer-left${open ? ' open' : ''}`} inert={open ? undefined : ''} aria-label="Menu">
        <div className="drawer-head">
          <Link to="/" onClick={onClose} aria-label="Rue home">
            <Brand />
          </Link>
          <button className="icon-btn" onClick={onClose} aria-label="Close menu">
            <Icon name="close" />
          </button>
        </div>
        <div className="drawer-body mobile-nav">
          <div className="mnav-section">
            <div className="eyebrow">Pages</div>
            <Link to="/" onClick={onClose} activeOptions={{ exact: true }} activeProps={{ className: 'active' }}>
              Home <Icon name="chevronRight" size={14} />
            </Link>
            <Link to="/shop" onClick={onClose} activeProps={{ className: 'active' }}>
              Shop <Icon name="chevronRight" size={14} />
            </Link>
            <Link to="/about" onClick={onClose} activeProps={{ className: 'active' }}>
              About <Icon name="chevronRight" size={14} />
            </Link>
            <Link to="/" hash="journal" onClick={onClose}>
              Journal <Icon name="chevronRight" size={14} />
            </Link>
          </div>
          {categories.length > 0 && (
            <div className="mnav-section">
              <div className="eyebrow">Shop by category</div>
              {categories.map((c) => (
                <Link key={c.id} to="/shop" search={{ category: c.slug }} onClick={onClose}>
                  {c.label} <Icon name="chevronRight" size={14} />
                </Link>
              ))}
            </div>
          )}
          <div className="mnav-section mnav-contact">
            <div className="eyebrow">Visit us</div>
            <p>
              <Icon name="pin" size={14} /> {STORE_INFO.addressLine1} · {STORE_INFO.addressLine2}
            </p>
            <p>
              <Icon name="phone" size={14} /> {STORE_INFO.phone}
            </p>
            <p>
              <Icon name="clock" size={14} /> {STORE_INFO.hours}
            </p>
          </div>
        </div>
      </aside>
    </>
  );
}
```

- [ ] **Step 2: Header** — add `menuOpen` state, wire the menu button (`onClick={() => setMenuOpen(true)}`), render `<MobileMenu open={menuOpen} onClose={() => setMenuOpen(false)} />` next to the SearchOverlay. Heart button becomes:

```tsx
<button className="header-icon-btn" aria-label="Wishlist" disabled title="Saved items coming soon">
  <Icon name="heart" size={20} />
</button>
```

(no badge, no `style` prop). Remove `wishlistCount` from the `useCart()` destructure.

- [ ] **Step 3: Footer** — import `STORE_INFO` and replace the three hardcoded contact strings in the "Visit the shop" column (`Community 18, Spintex` / `Adjacent KFC, Accra` / `0594 701 345` / `Mon–Sat · 9am – 8pm`) with `{STORE_INFO.addressLine1}`, `{STORE_INFO.addressLine2}`, `{STORE_INFO.phone}`, `{STORE_INFO.hours}` — markup (br/span structure) unchanged.

- [ ] **Step 4: Cart provider** — remove the `wishlistCount: number;` interface field, the `const [wishlistCount, setWishlistCount] = useState(0);` line, and `wishlistCount,` from the provider value. `pnpm typecheck` sweeps any remaining consumer (after Step 2 there should be none — fix any it finds by removing the usage).

- [ ] **Step 5: Typecheck + tests** — `pnpm typecheck` zero errors; `pnpm vitest run` all pass. Also verify the disabled heart doesn't break lint (`--max-warnings 0`).

- [ ] **Step 6: Commit**

```bash
git add src/features/shared/mobile-menu.tsx src/features/shared/layouts src/features/cart/cart-provider.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): mobile menu — pages/category/contact drawer, header wiring, honest heart, dead wishlist state removed"
```

---

### Task 5: CartItemRow extraction + drawer inert + Google glyph

**Files:**
- Create: `frontend/src/features/cart/cart-item-row.tsx`
- Modify: `frontend/src/features/cart/cart-drawer.tsx` (use CartItemRow; inert; switch to shared hooks)
- Modify: `frontend/src/features/cart/cart-page.tsx` (use CartItemRow)
- Modify: `frontend/src/features/auth/login-page.tsx` (Google glyph)

**Interfaces:**
- Consumes: `useEscToClose`/`useLockBodyScroll` (Task 1), `inert` typing; cart item type `InternalCartCartItemResponse` from the generated client (fields used: `id, product_name, product_image_path, qty, unit_price_ghs_minor`).
- Produces: `CartItemRow({ item, onUpdateQty, onRemove })`.

- [ ] **Step 1: Create `cart-item-row.tsx`** — pure extraction of the JSX currently duplicated at `cart-drawer.tsx:82–137` and `cart-page.tsx:52–107` (byte-identical rendering):

```tsx
import { Icon } from '../shared/ui/icons';
import { formatGhs, getImageUrl } from '../../lib/format/utils';
import type { InternalCartCartItemResponse } from '../../lib/api/generated/rueCosmeticsAPI';

interface CartItemRowProps {
  item: InternalCartCartItemResponse;
  onUpdateQty: (id: string, qty: number) => void;
  onRemove: (id: string) => void;
}

/** One bag line — shared by the cart drawer and the cart page (identical markup). */
export function CartItemRow({ item, onUpdateQty, onRemove }: CartItemRowProps) {
  return (
    <div className="cart-item">
      {item.product_image_path ? (
        <img
          src={getImageUrl(item.product_image_path)}
          alt={item.product_name || 'Product'}
          style={{ width: 80, height: 100, flexShrink: 0, objectFit: 'cover', borderRadius: 'var(--radius)' }}
          loading="lazy"
        />
      ) : (
        <div className="ph ph--lavender" style={{ width: 80, height: 100, flexShrink: 0 }}>
          <span className="ph-label" style={{ fontSize: 8 }}>
            {item.product_name?.substring(0, 2) || ''}
          </span>
        </div>
      )}
      <div className="cart-item-body">
        <div className="cart-item-name">{item.product_name}</div>
        <div className="cart-item-row">
          <div className="qty">
            <button
              onClick={() => {
                const newQty = (item.qty || 1) - 1;
                if (newQty < 1) {
                  onRemove(item.id!);
                } else {
                  onUpdateQty(item.id!, newQty);
                }
              }}
              aria-label="Decrease quantity"
            >
              <Icon name="minus" size={12} />
            </button>
            <span>{item.qty || 1}</span>
            <button
              onClick={() => onUpdateQty(item.id!, (item.qty || 1) + 1)}
              aria-label="Increase quantity"
            >
              <Icon name="plus" size={12} />
            </button>
          </div>
          <div className="price">{formatGhs((item.unit_price_ghs_minor || 0) * (item.qty || 1))}</div>
        </div>
      </div>
      <button className="cart-item-remove" onClick={() => onRemove(item.id!)} aria-label="Remove item">
        <Icon name="close" size={14} />
      </button>
    </div>
  );
}
```

- [ ] **Step 2: Use it in both call sites.** In `cart-drawer.tsx` and `cart-page.tsx`, replace the duplicated `items.map((item) => ( <div className="cart-item" ...>…</div> ))` bodies with:

```tsx
items.map((item) => (
  <CartItemRow
    key={item.id}
    item={item}
    onUpdateQty={(id, qty) => void updateItem(id, qty)}
    onRemove={(id) => void removeItem(id)}
  />
))
```

Remove now-unused imports from each file (`getImageUrl`, and `Icon`/`formatGhs` ONLY if genuinely unused afterwards — both files use them elsewhere in foot/summary sections; typecheck+lint will tell).

- [ ] **Step 3: Drawer a11y** — in `cart-drawer.tsx`: add `inert={open ? undefined : ''}` to the `<aside>` (keep `aria-hidden={!open}`), and replace its two inline effects (escape-key + body-scroll) with the shared hooks:

```tsx
useEscToClose(open, onClose);
useLockBodyScroll(open);
```

(import from `../shared/use-overlay`; delete the two `useEffect` blocks at lines 17–37 and the now-unused `useEffect` import if nothing else uses it).

- [ ] **Step 4: Google glyph** — in `login-page.tsx`, replace `<Icon name="user" size={14} />` inside the Google button with the multi-color G (decorative):

```tsx
<svg width="14" height="14" viewBox="0 0 18 18" aria-hidden="true">
  <path fill="#4285F4" d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62z" />
  <path fill="#34A853" d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.84.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18z" />
  <path fill="#FBBC05" d="M3.97 10.72a5.4 5.4 0 0 1 0-3.44V4.95H.96a9 9 0 0 0 0 8.1l3.01-2.33z" />
  <path fill="#EA4335" d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.59A9 9 0 0 0 .96 4.95l3.01 2.33C4.68 5.16 6.66 3.58 9 3.58z" />
</svg>
```

(`Icon` import stays if the page uses it elsewhere — check; remove if orphaned.)

- [ ] **Step 5: Typecheck + tests** — `pnpm typecheck` zero errors; `pnpm vitest run` all pass (add-toast tests exercise the cart provider tree — they must stay green).

- [ ] **Step 6: Commit**

```bash
git add src/features/cart src/features/auth/login-page.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "refactor(frontend): shared CartItemRow, drawer inert + shared overlay hooks, Google glyph on login"
```

---

### Task 6: Final gate, audits, ledger

**Files:**
- Modify (only if audits find gaps): `frontend/src/styles/globals.css`
- The controller maintains `.superpowers/sdd/progress.md` — do not edit it.

- [ ] **Step 1: Audits** (from `frontend/`)

```bash
# 1. Class coverage over files this tranche created/touched (expect no "missing" lines):
for f in src/features/shared/search-overlay.tsx src/features/shared/mobile-menu.tsx src/features/cart/cart-item-row.tsx src/features/shared/layouts/header.tsx; do
  grep -o 'className="[^"]*"' "$f" | sed 's/className="//;s/"$//' | tr ' ' '\n' | sort -u | while read -r c; do
    [ -z "$c" ] && continue
    grep -qs -- "\.$c" src/styles/*.css || echo "$f: missing .$c"
  done
done
# Template-literal classes to verify by hand: `search-overlay open`, `drawer drawer-left open`,
# `drawer-scrim open`, `ph ph--${tone}` — all defined (search-overlay.open, drawer-left.open in globals).

# 2. No new Tailwind residue in touched files:
grep -nE 'className="[^"]*\b(px-|py-|mb-|mt-|text-|bg-|rounded|w-full|space-y|gap-)' \
  src/features/shared/search-overlay.tsx src/features/shared/mobile-menu.tsx src/features/cart/cart-item-row.tsx
```

- [ ] **Step 2: Full gate** — `pnpm typecheck && pnpm lint && pnpm vitest run && pnpm build`. All green, lint zero warnings. Fix minimally anything arising from touched files; report every fix.

- [ ] **Step 3: Functional smoke** — only if dev servers are already running; otherwise note "deferred to human walkthrough": search overlay opens/focuses/searches/navigates; Esc closes all three surfaces; closed surfaces untabbable (Tab from the page never lands inside); mobile menu categories navigate to `/shop?category=<slug>` and the chip state matches; back/forward moves the filter; category chips now actually filter products (the id→slug fix); cart drawer/page rows render identically to before; Google button shows the glyph.

- [ ] **Step 4: Commit** (only if the audits/gate changed files)

```bash
git add src/
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "chore(frontend): tranche-4 final gate — audit fixes"
```
