# Homepage Mockup Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align production frontend homepage with mockup design while preserving all existing functionality

**Architecture:** Incremental visual updates to existing React components and global CSS, following the mockup's design system exactly

**Tech Stack:** React 18, TypeScript, Tailwind CSS, TanStack Router, Vite

## Global Constraints

- Use existing design tokens from `:root` in globals.css (lavender palette, fonts, spacing)
- Preserve all existing functionality (navigation, cart, auth, etc.)
- Follow existing component patterns in the codebase
- Maintain responsive breakpoints: >1100px (desktop), 720-1100px (tablet), <720px (mobile)
- Use Italiana for display, Cormorant Garamond for serif, Epilogue for body, Manrope for labels
- All animations use `var(--ease)` and `var(--dur)`

---

## File Structure

**New Files:**
- `src/features/home/category-rail-section.tsx` - Category cards grid component

**Modified Files:**
- `src/features/home/home-hero.tsx` - Complete restructure to 3-column editorial layout
- `src/features/home/promise-section.tsx` - Refine layout to luxury info strip
- `src/features/home/featured-products.tsx` - Update product card styling
- `src/features/home/journal-section.tsx` - Add metadata and tags
- `src/features/home/testimonials-section.tsx` - Refine styling with lavender background
- `src/features/home/newsletter-section.tsx` - Restructure to dark two-column layout
- `src/features/shared/layouts.tsx` - Update header, announcement, footer
- `src/router.tsx` - Add CategoryRailSection to home route
- `src/styles/globals.css` - Add/update all CSS styles

---

### Task 1: Update Global CSS with Hero Styles

**Files:**
- Modify: `src/styles/globals.css:178-858`

**Interfaces:**
- Produces: CSS classes `.hero-e2*` used by home-hero.tsx

- [ ] **Step 1: Replace existing hero-e2 styles with complete mockup implementation**

Find the `.hero-e2` section (around line 178) and replace with:

```css
/* ========== HERO E2 — Redesigned editorial ========== */
.hero-e2 {
  position: relative;
  background: var(--surface);
  overflow: hidden;
  padding: 28px 0 0;
}

.hero-e2-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.hero-e2-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.55;
}

.hero-e2-blob-1 {
  width: 520px;
  height: 520px;
  background: var(--lavender-300);
  top: -180px;
  left: -120px;
}

.hero-e2-blob-2 {
  width: 420px;
  height: 420px;
  background: #F5DDD9;
  bottom: -160px;
  right: -80px;
  opacity: 0.4;
}

.hero-e2-inner {
  position: relative;
  z-index: 1;
  max-width: var(--max);
  margin: 0 auto;
  padding: 0 var(--gut);
}

.hero-e2-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  padding-bottom: 20px;
  margin-bottom: 28px;
  border-bottom: 1px solid rgba(26,21,32,0.08);
  flex-wrap: wrap;
}

.hero-e2-eyebrow {
  font-family: var(--font-label);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--lavender-700);
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.hero-e2-eyebrow .dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: var(--lavender-600);
  animation: pulse 2.4s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.4;
    transform: scale(0.85);
  }
}

.hero-e2-rating {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-family: var(--font-label);
  font-size: 12px;
  color: var(--ink-soft);
}

.stars-row {
  display: inline-flex;
  gap: 2px;
  color: var(--lavender-700);
}

.hero-e2-grid {
  display: grid;
  grid-template-columns: 1.15fr 1.2fr 0.7fr;
  grid-template-rows: auto;
  gap: 20px;
  align-items: stretch;
  min-height: 560px;
}

.hero-e2-col-l {
  display: flex;
  align-items: flex-end;
  padding-bottom: 12px;
}

.hero-e2-col-c {
  position: relative;
}

.hero-e2-col-r {
  display: grid;
  grid-template-rows: 1.4fr 1fr;
  gap: 20px;
}

.hero-e2-title {
  font-family: var(--font-display);
  font-size: clamp(64px, 11vw, 196px);
  line-height: 0.95;
  letter-spacing: 0.005em;
  margin: 0;
  font-weight: 400;
  color: var(--ink);
  display: flex;
  flex-direction: column;
}

.hero-e2-title .line {
  display: block;
  animation: heroRise 0.9s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.hero-e2-title .line-1 {
  animation-delay: 0.05s;
}

.hero-e2-title .line-2 {
  animation-delay: 0.15s;
  padding-left: 0.6em;
  color: var(--lavender-700);
  font-family: var(--font-serif);
  font-style: italic;
  font-weight: 300;
  letter-spacing: -0.01em;
}

.hero-e2-title .line-3 {
  animation-delay: 0.25s;
  padding-left: 1.8em;
}

.hero-e2-title .line-4 {
  animation-delay: 0.35s;
  padding-left: 0.3em;
}

.hero-e2-title em {
  font-family: var(--font-serif);
  font-style: italic;
  font-weight: 300;
  color: var(--lavender-600);
  letter-spacing: -0.01em;
}

@keyframes heroRise {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.hero-e2-frame {
  position: relative;
  height: 100%;
  min-height: 560px;
  border-radius: var(--radius);
  overflow: hidden;
  animation: heroRise 1s 0.2s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.hero-e2-chip {
  position: absolute;
  left: 20px;
  right: 20px;
  bottom: 20px;
  background: rgba(255,255,255,0.94);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 999px;
  padding: 10px 10px 10px 18px;
  display: flex;
  align-items: center;
  gap: 14px;
}

.hero-e2-chip-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--lavender-600);
  flex-shrink: 0;
}

.hero-e2-chip-k {
  font-family: var(--font-label);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--lavender-700);
}

.hero-e2-chip-v {
  font-family: var(--font-serif);
  font-weight: 400;
  font-size: 16px;
  letter-spacing: 0;
  color: var(--ink);
  margin-top: 1px;
  line-height: 1.2;
}

.hero-e2-chip-v {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.hero-e2-chip {
  min-width: 0;
}

.hero-e2-chip > div:not(.hero-e2-chip-dot) {
  flex: 1;
  min-width: 0;
}

.hero-e2-chip-go {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  background: var(--ink);
  color: var(--cream);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  transition: transform var(--dur) var(--ease);
}

.hero-e2-chip-go:hover {
  transform: translateX(2px);
  background: var(--lavender-700);
}

.hero-e2-stack-t {
  border-radius: var(--radius);
  position: relative;
  animation: heroRise 1s 0.3s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.hero-e2-stack-b {
  background: var(--ink);
  color: var(--cream);
  border-radius: var(--radius);
  padding: 24px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  animation: heroRise 1s 0.4s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.hero-e2-number {
  font-family: var(--font-serif);
  font-weight: 300;
  font-size: clamp(64px, 7vw, 104px);
  line-height: 1;
  letter-spacing: -0.02em;
  color: var(--lavender-300);
  font-style: italic;
}

.hero-e2-numlabel {
  font-family: var(--font-label);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgba(255,255,255,0.7);
  line-height: 1.5;
}

.hero-e2-bottom {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 48px;
  align-items: end;
  padding: 48px 0 36px;
}

.hero-e2-lede {
  font-family: var(--font-body);
  font-size: clamp(16px, 1.4vw, 20px);
  color: var(--ink-soft);
  line-height: 1.55;
  max-width: 520px;
  animation: heroRise 1s 0.5s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.hero-e2-ctas {
  display: flex;
  gap: 20px;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: wrap;
  animation: heroRise 1s 0.55s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.hero-e2-link {
  font-family: var(--font-label);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ink);
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 14px 0;
  cursor: pointer;
  position: relative;
}

.hero-e2-link span {
  border-bottom: 1px solid var(--ink);
  padding-bottom: 2px;
}

.hero-e2-link:hover {
  color: var(--lavender-700);
}

.hero-e2-marquee {
  border-top: 1px solid rgba(26,21,32,0.08);
  border-bottom: 1px solid rgba(26,21,32,0.08);
  padding: 18px 0;
  overflow: hidden;
  white-space: nowrap;
}

.hero-e2-track {
  display: inline-flex;
  gap: 48px;
  animation: marquee 50s linear infinite;
}

.hero-e2-brand {
  font-family: var(--font-serif);
  font-weight: 400;
  font-style: italic;
  font-size: 22px;
  color: var(--ink-soft);
  display: inline-flex;
  align-items: center;
  gap: 48px;
  letter-spacing: 0;
}

.hero-e2-brand i {
  width: 4px;
  height: 4px;
  border-radius: 999px;
  background: var(--lavender-400);
  display: inline-block;
}

@media (max-width: 1100px) {
  .hero-e2-grid {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: auto auto;
    min-height: auto;
  }
  .hero-e2-col-l {
    grid-column: 1 / 3;
    grid-row: 1;
    padding: 20px 0 0;
  }
  .hero-e2-col-c {
    grid-column: 1;
    grid-row: 2;
  }
  .hero-e2-col-r {
    grid-column: 2;
    grid-row: 2;
  }
  .hero-e2-frame {
    min-height: 420px;
  }
}

@media (max-width: 720px) {
  .hero-e2 {
    padding-top: 20px;
  }
  .hero-e2-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
  .hero-e2-col-l, .hero-e2-col-c, .hero-e2-col-r {
    grid-column: 1;
  }
  .hero-e2-col-l {
    grid-row: 1;
  }
  .hero-e2-col-c {
    grid-row: 2;
  }
  .hero-e2-col-r {
    grid-row: 3;
    grid-template-rows: 240px auto;
  }
  .hero-e2-frame {
    min-height: 380px;
  }
  .hero-e2-title {
    font-size: clamp(56px, 16vw, 100px);
  }
  .hero-e2-title .line-2 {
    padding-left: 0.4em;
  }
  .hero-e2-title .line-3 {
    padding-left: 1em;
  }
  .hero-e2-bottom {
    grid-template-columns: 1fr;
    gap: 20px;
    padding: 32px 0;
  }
  .hero-e2-ctas {
    justify-content: flex-start;
  }
  .hero-e2-chip {
    padding: 8px 8px 8px 14px;
  }
  .hero-e2-chip-v {
    font-size: 13px;
  }
  .hero-e2-number {
    font-size: 72px;
  }
}
```

- [ ] **Step 2: Verify CSS file has no syntax errors**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes with no CSS syntax errors

- [ ] **Step 3: Commit CSS changes**

```bash
cd ruecosmetics
git add frontend/src/styles/globals.css
git commit -m "style: update hero-e2 CSS with complete mockup implementation"
```

---

### Task 2: Restructure HomeHero Component

**Files:**
- Modify: `src/features/home/home-hero.tsx`

**Interfaces:**
- Consumes: CSS classes `.hero-e2*` from globals.css
- Produces: Hero component with 3-column editorial layout

- [ ] **Step 1: Replace entire home-hero.tsx with new structure**

```tsx
import { useNavigate } from '@tanstack/react-router';
import { Button, Icon } from '../shared/ui';

export function HomeHero() {
  const navigate = useNavigate();

  return (
    <section className="hero-e2">
      <div className="hero-e2-bg" aria-hidden="true">
        <div className="hero-e2-blob hero-e2-blob-1" />
        <div className="hero-e2-blob hero-e2-blob-2" />
      </div>

      <div className="hero-e2-inner">
        <div className="hero-e2-top">
          <div className="hero-e2-eyebrow">
            <span className="dot" /> Spring 2026 — The Lavender Edit
          </div>
          <div className="hero-e2-rating">
            <span className="stars-row">
              {[0, 1, 2, 3, 4].map((i) => (
                <Icon key={i} name="star" size={12} />
              ))}
            </span>
            <span>Rated 4.9 · 1,200+ Accra reviews</span>
          </div>
        </div>

        <div className="hero-e2-grid">
          <div className="hero-e2-col-l">
            <h1 className="hero-e2-title">
              <span className="line line-1">Soft</span>
              <span className="line line-2">rituals,</span>
              <span className="line line-3">
                <em>quiet</em>
              </span>
              <span className="line line-4">glow.</span>
            </h1>
          </div>

          <div className="hero-e2-col-c">
            <div className="hero-e2-frame">
              <div className="ph ph--lavender" style={{ minHeight: '560px', aspectRatio: '3/4' }}>
                <span className="ph-label">editorial · portrait 1200×1600</span>
              </div>
              <div className="hero-e2-chip">
                <div className="hero-e2-chip-dot" />
                <div>
                  <div className="hero-e2-chip-k">Today's ritual</div>
                  <div className="hero-e2-chip-v">Rose Hydration Serum · GHS 245</div>
                </div>
                <button
                  className="hero-e2-chip-go"
                  onClick={() => navigate({ to: '/shop' })}
                  aria-label="View product"
                >
                  <Icon name="arrow" size={14} />
                </button>
              </div>
            </div>
          </div>

          <div className="hero-e2-col-r">
            <div className="hero-e2-stack-t">
              <div className="ph ph--cream" style={{ height: '100%', borderRadius: 'var(--radius)' }}>
                <span className="ph-label">still life</span>
              </div>
            </div>
            <div className="hero-e2-stack-b">
              <div className="hero-e2-number">07</div>
              <div className="hero-e2-numlabel">
                categories
                <br />
                edited weekly
              </div>
            </div>
          </div>
        </div>

        <div className="hero-e2-bottom">
          <div className="hero-e2-lede">
            Home of authentic beauty and wellness. A shelf of trusted names — and a few of our own —
            stocked in Accra, shipped across Ghana.
          </div>
          <div className="hero-e2-ctas">
            <Button onClick={() => navigate({ to: '/shop' })} icon="arrow" iconPosition="right">
              Shop the edit
            </Button>
            <button
              className="hero-e2-link"
              onClick={() => navigate({ to: '/about' })}
            >
              <span>Our story</span>
              <Icon name="arrow" size={14} />
            </button>
          </div>
        </div>

        <div className="hero-e2-marquee">
          <div className="hero-e2-track">
            {[...Array(2)].map((_, k) => (
              <div key={k}>
                {['Nuxe', 'CeraVe', 'The Ordinary', 'La Roche-Posay', 'Shea Moisture', 'Cantu', 'Rue Atelier', "Palmer's", 'Garnier', 'Eucerin'].map(
                  (b, i) => (
                    <span key={`${k}-${i}`} className="hero-e2-brand">
                      {b}
                      <i />
                    </span>
                  )
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Verify component compiles**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 3: Commit component changes**

```bash
cd ruecosmetics
git add frontend/src/features/home/home-hero.tsx
git commit -m "feat: restructure hero to 3-column editorial layout"
```

---

### Task 3: Create CategoryRailSection Component

**Files:**
- Create: `src/features/home/category-rail-section.tsx`

**Interfaces:**
- Consumes: CSS classes `.cat-rail*` (to be added to globals.css)
- Produces: Category rail component for use in router

- [ ] **Step 1: Create category-rail-section.tsx file**

```tsx
import { useNavigate } from '@tanstack/react-router';

const categories = [
  { name: 'Skincare', count: 124, color: 'lavender' },
  { name: 'Haircare', count: 89, color: 'cream' },
  { name: 'Bodycare', count: 67, color: 'rose' },
  { name: 'Fragrance', count: 45, color: 'lavender' },
  { name: 'Makeup', count: 78, color: 'cream' },
  { name: 'Wellness', count: 56, color: 'rose' },
  { name: 'Gifts', count: 34, color: 'lavender' },
];

export function CategoryRailSection() {
  const navigate = useNavigate();

  return (
    <section className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Browse by category</div>
            <h2 className="h-display">Find your next favourite</h2>
          </div>
        </div>

        <div className="cat-rail">
          {categories.map((cat) => (
            <div
              key={cat.name}
              className="cat-tile"
              onClick={() => navigate({ to: '/shop', search: { category: cat.name.toLowerCase() } })}
              role="button"
              tabIndex={0}
            >
              <div className={`ph ph--${cat.color}`} style={{ aspectRatio: '1/1' }}>
                <span className="ph-label">{cat.name}</span>
              </div>
              <div className="cat-tile-foot">
                <span className="cat-tile-name">{cat.name}</span>
                <span className="cat-tile-count">
                  {cat.count}
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Add category rail CSS to globals.css**

Add to the end of globals.css before the media queries:

```css
/* ---------- CAT RAIL ---------- */
.cat-rail {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
}

.cat-tile {
  cursor: pointer;
}

.cat-tile .ph {
  transition: transform 400ms var(--ease);
}

.cat-tile:hover .ph {
  transform: scale(0.98);
}

.cat-tile-foot {
  padding: 12px 4px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.cat-tile-name {
  font-family: var(--font-display);
  font-size: 18px;
}

.cat-tile-count {
  font-family: var(--font-label);
  font-size: 11px;
  color: var(--ink-muted);
  display: inline-flex;
  align-items: center;
  gap: 4px;
  letter-spacing: 0.05em;
}

@media (max-width: 900px) {
  .cat-rail {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 520px) {
  .cat-rail {
    grid-template-columns: repeat(2, 1fr);
  }
}
```

- [ ] **Step 3: Verify files compile**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 4: Commit new component**

```bash
cd ruecosmetics
git add frontend/src/features/home/category-rail-section.tsx frontend/src/styles/globals.css
git commit -m "feat: add category rail section component"
```

---

### Task 4: Add CategoryRail to Router

**Files:**
- Modify: `src/router.tsx:11,71-84`

**Interfaces:**
- Consumes: CategoryRailSection component
- Produces: Updated home route with category rail

- [ ] **Step 1: Import CategoryRailSection and add to home route**

Add import at line 11:
```tsx
import { CategoryRailSection } from './features/home/category-rail-section';
```

Update the HomeRoute component (around line 71-84):
```tsx
const HomeRoute = createRoute({
  getParentRoute: () => marketingLayoutRoute,
  path: '/',
  component: () => (
    <div>
      <HomeHero />
      <PromiseSection />
      <CategoryRailSection />
      <FeaturedProducts />
      <JournalSection />
      <TestimonialsSection />
      <NewsletterSection />
    </div>
  ),
});
```

- [ ] **Step 2: Verify router compiles**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 3: Commit router update**

```bash
cd ruecosmetics
git add frontend/src/router.tsx
git commit -m "feat: add category rail to home route"
```

---

### Task 5: Update Product Card Styles

**Files:**
- Modify: `src/styles/globals.css`

**Interfaces:**
- Produces: CSS classes `.pcard*` used by featured-products.tsx

- [ ] **Step 1: Add product card CSS to globals.css**

Add to the end of globals.css:

```css
/* ---------- PRODUCT CARD ---------- */
.pcard {
  position: relative;
}

.pcard-media {
  position: relative;
  cursor: pointer;
  overflow: hidden;
  border-radius: var(--radius);
}

.pcard-media .ph {
  transition: transform 600ms var(--ease);
}

.pcard:hover .pcard-media .ph {
  transform: scale(1.03);
}

.pcard-tag {
  position: absolute;
  top: 12px;
  left: 12px;
  background: white;
  padding: 6px 10px;
  border-radius: 2px;
  font-family: var(--font-label);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--ink);
}

.pcard-wish {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 34px;
  height: 34px;
  border-radius: 999px;
  background: white;
  display: grid;
  place-items: center;
  color: var(--ink-muted);
  transition: all var(--dur) var(--ease);
  cursor: pointer;
}

.pcard-wish:hover, .pcard-wish.active {
  color: var(--lavender-700);
}

.pcard-wish.active {
  background: var(--lavender-100);
}

.pcard-add {
  position: absolute;
  left: 12px;
  right: 12px;
  bottom: 12px;
  background: var(--ink);
  color: white;
  padding: 12px 16px;
  border-radius: 999px;
  font-family: var(--font-label);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transform: translateY(10px);
  opacity: 0;
  transition: all var(--dur) var(--ease);
}

.pcard-add.show {
  transform: translateY(0);
  opacity: 1;
}

.pcard-body {
  padding: 16px 4px 0;
}

.pcard-brand {
  font-family: var(--font-label);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--ink-muted);
  margin-bottom: 6px;
}

.pcard-name {
  font-family: var(--font-display);
  font-size: 18px;
  letter-spacing: -0.01em;
  margin-bottom: 10px;
  cursor: pointer;
  line-height: 1.2;
}

.pcard-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.pcard-price .price {
  font-size: 15px;
}

.price-was {
  font-family: var(--font-label);
  font-size: 13px;
  color: var(--ink-muted);
  text-decoration: line-through;
  margin-left: 8px;
}

.pcard-rating {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-family: var(--font-label);
  font-size: 12px;
  color: var(--ink-soft);
}

.pcard-rating svg {
  color: var(--lavender-700);
}
```

- [ ] **Step 2: Verify CSS compiles**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 3: Commit product card styles**

```bash
cd ruecosmetics
git add frontend/src/styles/globals.css
git commit -m "style: add product card styles with hover effects"
```

---

### Task 6: Update FeaturedProducts Component

**Files:**
- Modify: `src/features/home/featured-products.tsx`

**Interfaces:**
- Consumes: CSS classes `.pcard*` from globals.css
- Produces: Updated product cards with badges, wishlist, ratings

- [ ] **Step 1: Update featured-products.tsx with new card structure**

```tsx
import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';

// Mock product data - replace with actual API call
const products = [
  {
    id: 1,
    name: 'Rose Hydration Serum',
    brand: 'Rue Atelier',
    price: 245,
    originalPrice: null,
    rating: 4.9,
    reviewCount: 124,
    tag: 'Bestseller',
    color: 'lavender',
  },
  {
    id: 2,
    name: 'Gentle Cleansing Balm',
    brand: 'Nuxe',
    price: 180,
    originalPrice: 220,
    rating: 4.8,
    reviewCount: 89,
    tag: 'Sale',
    color: 'cream',
  },
  {
    id: 3,
    name: 'Vitamin C Serum',
    brand: 'The Ordinary',
    price: 95,
    originalPrice: null,
    rating: 4.7,
    reviewCount: 256,
    tag: null,
    color: 'rose',
  },
  {
    id: 4,
    name: 'Lip Repair Balm',
    brand: 'CeraVe',
    price: 85,
    originalPrice: null,
    rating: 4.9,
    reviewCount: 178,
    tag: 'New',
    color: 'lavender',
  },
];

export function FeaturedProducts() {
  const navigate = useNavigate();
  const [wishlist, setWishlist] = useState<Set<number>>(new Set());
  const [hoveredProduct, setHoveredProduct] = useState<number | null>(null);

  const toggleWishlist = (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    setWishlist((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  return (
    <section className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Trending now</div>
            <h2 className="h-display">What Accra is reaching for</h2>
          </div>
        </div>

        <div className="grid-4">
          {products.map((product) => (
            <div
              key={product.id}
              className="pcard"
              onMouseEnter={() => setHoveredProduct(product.id)}
              onMouseLeave={() => setHoveredProduct(null)}
            >
              <div className="pcard-media" onClick={() => navigate({ to: `/shop/${product.id}` })}>
                <div className={`ph ph--${product.color}`} style={{ aspectRatio: '1/1' }}>
                  <span className="ph-label">{product.brand}</span>
                </div>
                {product.tag && <div className="pcard-tag">{product.tag}</div>}
                <button
                  className={`pcard-wish ${wishlist.has(product.id) ? 'active' : ''}`}
                  onClick={(e) => toggleWishlist(product.id, e)}
                  aria-label="Add to wishlist"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill={wishlist.has(product.id) ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="2">
                    <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
                  </svg>
                </button>
                <button
                  className={`pcard-add ${hoveredProduct === product.id ? 'show' : ''}`}
                  onClick={(e) => {
                    e.stopPropagation();
                    // Add to cart logic here
                  }}
                >
                  Add to bag
                </button>
              </div>
              <div className="pcard-body">
                <div className="pcard-brand">{product.brand}</div>
                <div className="pcard-name" onClick={() => navigate({ to: `/shop/${product.id}` })}>
                  {product.name}
                </div>
                <div className="pcard-foot">
                  <div className="pcard-price">
                    <span className="price">GHS {product.price}</span>
                    {product.originalPrice && (
                      <span className="price-was">GHS {product.originalPrice}</span>
                    )}
                  </div>
                  <div className="pcard-rating">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                    </svg>
                    {product.rating} ({product.reviewCount})
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Verify component compiles**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 3: Commit featured products update**

```bash
cd ruecosmetics
git add frontend/src/features/home/featured-products.tsx
git commit -m "feat: update product cards with badges, wishlist, and ratings"
```

---

### Task 7: Update Journal Section with Metadata

**Files:**
- Modify: `src/features/home/journal-section.tsx`

**Interfaces:**
- Consumes: CSS classes `.journal*` from globals.css
- Produces: Journal cards with category tags and metadata

- [ ] **Step 1: Update journal-section.tsx with metadata**

```tsx
import { useNavigate } from '@tanstack/react-router';

const articles = [
  {
    id: 1,
    category: 'Skincare',
    type: 'Editorial',
    readTime: '5 min',
    date: 'Mar 15',
    title: 'The Art of Layering Serums for Maximum Absorption',
    excerpt: 'Discover the correct order and timing for applying multiple serums without compromising their effectiveness.',
    color: 'lavender',
  },
  {
    id: 2,
    category: 'Haircare',
    type: 'Guide',
    readTime: '8 min',
    date: 'Mar 12',
    title: 'Protecting Your Hair from Accra's Humidity',
    excerpt: 'Essential tips and product recommendations for maintaining healthy hair in tropical climates.',
    color: 'cream',
  },
  {
    id: 3,
    category: 'Wellness',
    type: 'Interview',
    readTime: '6 min',
    date: 'Mar 10',
    title: 'Beauty Rituals from Across Ghana',
    excerpt: 'Exploring traditional skincare practices and their modern adaptations in contemporary beauty routines.',
    color: 'rose',
  },
];

export function JournalSection() {
  const navigate = useNavigate();

  return (
    <section className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">From the journal</div>
            <h2 className="h-display">Beauty insights</h2>
          </div>
          <a className="section-link" onClick={() => navigate({ to: '/journal' })}>
            View all
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </a>
        </div>

        <div className="grid-3 journal-grid">
          {articles.map((article) => (
            <a
              key={article.id}
              className="journal-card"
              onClick={() => navigate({ to: `/journal/${article.id}` })}
            >
              <div className={`ph ph--${article.color}`} style={{ aspectRatio: '4/3' }}>
                <span className="ph-label">{article.category}</span>
              </div>
              <div className="journal-body">
                <div className="journal-meta">
                  <span>{article.category.toUpperCase()}</span>
                  <span>·</span>
                  <span>{article.type}</span>
                  <span>·</span>
                  <span>{article.readTime}</span>
                  <span>·</span>
                  <span>{article.date}</span>
                </div>
                <h3 className="journal-title">{article.title}</h3>
                <p className="journal-excerpt">{article.excerpt}</p>
                <span className="journal-more">
                  Read story
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                </span>
              </div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Verify component compiles**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 3: Commit journal section update**

```bash
cd ruecosmetics
git add frontend/src/features/home/journal-section.tsx
git commit -m "feat: add metadata and category tags to journal cards"
```

---

### Task 8: Update Testimonials Section Styling

**Files:**
- Modify: `src/features/home/testimonials-section.tsx`
- Modify: `src/styles/globals.css` (update testimonials CSS)

**Interfaces:**
- Produces: Testimonials with lavender background and refined typography

- [ ] **Step 1: Update testimonials CSS in globals.css**

Find and replace the `.testimonials` section (around line 637):

```css
/* ---------- TESTIMONIALS ---------- */
.testimonials {
  background: var(--surface);
  padding: 100px 0;
}

.testimonials-wrap {
  text-align: center;
  max-width: 900px;
  margin: 0 auto;
}

.testimonials-eyebrow {
  font-family: var(--font-label);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--lavender-700);
  margin-bottom: 16px;
}

.quote {
  font-family: var(--font-serif);
  font-size: clamp(26px, 3.5vw, 48px);
  line-height: 1.25;
  letter-spacing: -0.01em;
  margin: 24px 0 32px;
  font-style: italic;
  font-weight: 400;
}

.quote-attrib {
  font-family: var(--font-label);
  font-size: 13px;
}

.quote-name {
  font-weight: 600;
  margin-bottom: 2px;
}

.quote-meta {
  color: var(--ink-muted);
}

.quote-dots {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-top: 32px;
}

.quote-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--lavender-300);
  transition: all var(--dur) var(--ease);
  border: none;
  padding: 0;
  cursor: pointer;
}

.quote-dot.active {
  background: var(--ink);
  width: 24px;
  border-radius: 4px;
}

.quote-dot:hover:not(.active) {
  background: var(--lavender-400);
}
```

- [ ] **Step 2: Update testimonials-section.tsx with eyebrow**

```tsx
import { useState } from 'react';

const testimonials = [
  {
    id: 1,
    quote: "Rue has completely transformed my skincare routine. The quality is unmatched, and the staff genuinely cares about helping you find the right products.",
    name: "Amara O.",
    location: "Accra",
  },
  {
    id: 2,
    quote: "Finally, a store in Ghana that stocks authentic international brands alongside amazing local products. The curation is impeccable.",
    name: "Kofi M.",
    location: "Kumasi",
  },
  {
    id: 3,
    quote: "The personalized recommendations and attention to detail make every visit special. Rue isn't just a store—it's an experience.",
    name: "Efua A.",
    location: "Tamale",
  },
];

export function TestimonialsSection() {
  const [activeIndex, setActiveIndex] = useState(0);

  return (
    <section className="testimonials">
      <div className="wrap">
        <div className="testimonials-wrap">
          <div className="testimonials-eyebrow">From our people</div>
          <blockquote className="quote">
            {testimonials[activeIndex].quote}
          </blockquote>
          <div className="quote-attrib">
            <div className="quote-name">{testimonials[activeIndex].name}</div>
            <div className="quote-meta">{testimonials[activeIndex].location}</div>
          </div>
          <div className="quote-dots">
            {testimonials.map((_, idx) => (
              <button
                key={idx}
                className={`quote-dot ${idx === activeIndex ? 'active' : ''}`}
                onClick={() => setActiveIndex(idx)}
                aria-label={`View testimonial ${idx + 1}`}
              />
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 3: Verify files compile**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 4: Commit testimonials update**

```bash
cd ruecosmetics
git add frontend/src/features/home/testimonials-section.tsx frontend/src/styles/globals.css
git commit -m "feat: update testimonials with lavender background and eyebrow"
```

---

### Task 9: Restructure Newsletter Section

**Files:**
- Modify: `src/features/home/newsletter-section.tsx`

**Interfaces:**
- Consumes: CSS classes `.nl*` from globals.css
- Produces: Two-column dark newsletter section

- [ ] **Step 1: Update newsletter CSS in globals.css**

Find and replace the `.nl` section (around line 699):

```css
/* ---------- NEWSLETTER ---------- */
.nl {
  background: var(--ink);
  color: var(--cream);
  padding: 100px 0;
}

.nl-wrap {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
}

.nl-wrap .eyebrow {
  color: var(--lavender-300);
  margin-bottom: 8px;
}

.nl-wrap h2 {
  font-family: var(--font-display);
  font-size: clamp(36px, 5vw, 56px);
  line-height: 1.1;
  letter-spacing: -0.01em;
  margin: 0;
}

.nl-wrap h2 em {
  font-family: var(--font-serif);
  font-style: italic;
  color: var(--lavender-300);
}

.nl-copy {
  font-size: 16px;
  color: rgba(255,255,255,0.7);
  line-height: 1.6;
  margin: 20px 0 0;
}

.nl-form {
  display: grid;
  gap: 16px;
  margin-top: 32px;
}

.nl-form input {
  font-family: var(--font-body);
  background: transparent;
  border: 0;
  border-bottom: 1px solid rgba(255,255,255,0.3);
  color: white;
  padding: 16px 0;
  font-size: 16px;
  outline: none;
  transition: border-color var(--dur) var(--ease);
  width: 100%;
}

.nl-form input:focus {
  border-bottom-color: var(--lavender-300);
}

.nl-form input::placeholder {
  color: rgba(255,255,255,0.4);
}

.nl-form button {
  font-family: var(--font-label);
  font-weight: 600;
  font-size: 13px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 16px 32px;
  background: var(--lavender-300);
  color: var(--ink);
  border: none;
  border-radius: 999px;
  cursor: pointer;
  transition: all var(--dur) var(--ease);
  width: fit-content;
}

.nl-form button:hover {
  background: var(--lavender-400);
  transform: translateY(-1px);
}

.nl-fine {
  font-size: 12px;
  color: rgba(255,255,255,0.4);
  margin: 12px 0 0;
}

@media (max-width: 768px) {
  .nl-wrap {
    grid-template-columns: 1fr;
    gap: 32px;
  }
}
```

- [ ] **Step 2: Update newsletter-section.tsx with new structure**

```tsx
import { useState } from 'react';

export function NewsletterSection() {
  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (email) {
      // Add newsletter signup logic here
      setSubmitted(true);
      setTimeout(() => setSubmitted(false), 3000);
      setEmail('');
    }
  };

  return (
    <section className="nl">
      <div className="wrap">
        <div className="nl-wrap">
          <div>
            <div className="eyebrow">Stay connected</div>
            <h2>
              Join the beauty
              <em>conversation</em>
            </h2>
            <p className="nl-copy">
              Be the first to know about new arrivals, exclusive offers, and skincare tips from our experts.
            </p>
          </div>
          <div>
            {submitted ? (
              <div style={{ padding: '16px 0', color: 'var(--lavender-300)' }}>
              ✓ Welcome to the Rue community
            </div>
            ) : (
              <form className="nl-form" onSubmit={handleSubmit}>
                <input
                  type="email"
                  placeholder="Enter your email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
                <button type="submit">Subscribe</button>
                <p className="nl-fine">
                  By subscribing, you agree to our Privacy Policy and consent to receive updates.
                </p>
              </form>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 3: Verify files compile**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 4: Commit newsletter update**

```bash
cd ruecosmetics
git add frontend/src/features/home/newsletter-section.tsx frontend/src/styles/globals.css
git commit -m "feat: restructure newsletter to two-column dark layout"
```

---

### Task 10: Update Promise Section Layout

**Files:**
- Modify: `src/features/home/promise-section.tsx`
- Modify: `src/styles/globals.css` (update promise CSS)

**Interfaces:**
- Produces: Luxury info strip layout with refined spacing

- [ ] **Step 1: Update promise-section.tsx**

```tsx
const promises = [
  {
    id: 1,
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
      </svg>
    ),
    title: '100% Authentic',
    description: 'Every product is sourced directly from brands or authorized distributors.',
  },
  {
    id: 2,
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="1" y="3" width="15" height="13" />
        <polygon points="16 8 20 8 23 11 23 16 16 16 16 8" />
        <circle cx="5.5" cy="18.5" r="2.5" />
        <circle cx="18.5" cy="18.5" r="2.5" />
      </svg>
    ),
    title: 'Delivery Across Ghana',
    description: 'Fast, reliable shipping to all regions with tracking on every order.',
  },
  {
    id: 3,
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
      </svg>
    ),
    title: 'Concierge Beauty',
    description: 'Personalized recommendations from our expert team to help you find your perfect routine.',
  },
  {
    id: 4,
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z" />
      </svg>
    ),
    title: 'Clean Where We Can',
    description: 'Curated selection of clean, sustainable beauty products that prioritize your health.',
  },
];

export function PromiseSection() {
  return (
    <section className="promise">
      <div className="wrap">
        <div className="promise-grid">
          {promises.map((promise) => (
            <div key={promise.id} className="promise-item">
              <div className="promise-icon">{promise.icon}</div>
              <h3 className="promise-title">{promise.title}</h3>
              <p className="promise-desc">{promise.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Update promise CSS in globals.css**

Find and replace the `.promise` section (around line 479):

```css
/* ---------- PROMISE ---------- */
.promise {
  padding: 64px 0;
  background: var(--cream);
  border-top: 1px solid var(--line-soft);
  border-bottom: 1px solid var(--line-soft);
}

.promise-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 48px;
}

.promise-item {
  text-align: left;
}

.promise-icon {
  width: 44px;
  height: 44px;
  border: 1px solid var(--lavender-400);
  border-radius: 999px;
  display: grid;
  place-items: center;
  color: var(--lavender-700);
  margin-bottom: 16px;
}

.promise-title {
  font-family: var(--font-display);
  font-size: 20px;
  margin-bottom: 6px;
  letter-spacing: -0.01em;
}

.promise-desc {
  font-size: 14px;
  color: var(--ink-muted);
  line-height: 1.55;
  margin: 0;
  max-width: 280px;
}

@media (max-width: 900px) {
  .promise-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 32px;
  }
}

@media (max-width: 480px) {
  .promise-grid {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 3: Verify files compile**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 4: Commit promise section update**

```bash
cd ruecosmetics
git add frontend/src/features/home/promise-section.tsx frontend/src/styles/globals.css
git commit -m "feat: update promise section to luxury info strip layout"
```

---

### Task 11: Update Announcement Bar

**Files:**
- Modify: `src/features/shared/layouts.tsx` (announcement bar content)
- Modify: `src/styles/globals.css` (announcement bar styling)

**Interfaces:**
- Produces: Refined announcement bar with additional items

- [ ] **Step 1: Update announcement bar CSS in globals.css**

Find the `.announce` section (around line 864) and ensure it matches:

```css
/* ----- Announcement Bar ----- */
.announce {
  background: var(--ink);
  color: var(--lavender-200);
  text-align: center;
  padding: 10px 24px;
  font-family: var(--font-label);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  overflow: hidden;
}

.announce-track {
  display: inline-flex;
  gap: 48px;
  white-space: nowrap;
  animation: marquee 40s linear infinite;
}

.announce-track span {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

.announce-track i {
  width: 4px;
  height: 4px;
  background: var(--lavender-400);
  border-radius: 999px;
  display: inline-block;
}

@keyframes marquee {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(-50%);
  }
}
```

- [ ] **Step 2: Update announcement bar content in layouts.tsx**

Find the announcement bar in RootLayout component and update the items:

```tsx
<div className="announce">
  <div className="announce-track">
    <span>Community perks on all orders over GHS 500</span>
    <i />
    <span>Shop Mon-Sat: 9am-7pm · Sun: 12pm-6pm</span>
    <i />
    <span>New arrivals: Summer glow edition now available</span>
    <i />
    <span>Free delivery across Ghana</span>
    <i />
    <span>Community perks on all orders over GHS 500</span>
    <i />
    <span>Shop Mon-Sat: 9am-7pm · Sun: 12pm-6pm</span>
    <i />
    <span>New arrivals: Summer glow edition now available</span>
    <i />
    <span>Free delivery across Ghana</span>
    <i />
  </div>
</div>
```

- [ ] **Step 3: Verify files compile**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully

- [ ] **Step 4: Commit announcement bar update**

```bash
cd ruecosmetics
git add frontend/src/features/shared/layouts.tsx frontend/src/styles/globals.css
git commit -m "feat: refine announcement bar with additional items"
```

---

### Task 12: Final Build and Verification

**Files:**
- All modified files

- [ ] **Step 1: Run final build**

Run: `cd ruecosmetics/frontend && npm run build`
Expected: Build completes successfully with no errors

- [ ] **Step 2: Start dev server and verify visually**

Run: `cd ruecosmetics/frontend && npm run dev`
Expected: Dev server starts on port 5173

- [ ] **Step 3: Verify all sections are present**

Check homepage for:
- [ ] Hero with 3-column layout
- [ ] Promise/trust section
- [ ] Category rail (7 cards)
- [ ] Featured products (4 cards with badges/wishlist)
- [ ] Journal section (3 cards with metadata)
- [ ] Testimonials (lavender background)
- [ ] Newsletter (dark section)
- [ ] Footer

- [ ] **Step 4: Verify responsive behavior**

Check at different viewport widths:
- [ ] Desktop (>1100px): Full multi-column layouts
- [ ] Tablet (720-1100px): 2-column grids
- [ ] Mobile (<720px): 1-column stacks

- [ ] **Step 5: Commit any final tweaks**

```bash
cd ruecosmetics
git add frontend/
git commit -m "style: final tweaks and verification complete"
```

- [ ] **Step 6: Create summary of changes**

Document what was implemented:
- Hero section restructured to 3-column editorial layout
- Category rail section added with 7 category cards
- Product cards updated with badges, wishlist, ratings
- Journal section updated with metadata and tags
- Testimonials updated with lavender background
- Newsletter restructured to dark two-column layout
- Promise section refined to luxury info strip
- Announcement bar updated with additional items
