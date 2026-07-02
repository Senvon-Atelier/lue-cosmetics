# Homepage Mockup Alignment Design

**Project:** Rue Cosmetics Frontend UI Refresh
**Date:** 2026-07-01
**Scope:** Align production frontend with mockup design while preserving all existing functionality

---

## Overview

The production frontend has complete functionality but lacks the visual polish and some sections present in the mockup. This design outlines all visual and structural changes needed to match the mockup exactly.

---

## Section-by-Section Changes

### 1. Hero Section (Highest Priority)

**Current State:**
- Basic three-column layout with placeholder imagery
- Compressed typography
- Flat white background
- Basic CTA buttons

**Target State (Mockup):**
- Sophisticated three-column editorial layout
- Large serif typography with animations
- Subtle lavender gradient background with animated blobs
- Pill-shaped CTA buttons
- Better spacing and visual hierarchy

**Implementation Requirements:**
- **Layout:** 3-column grid (1.15fr | 1.2fr | 0.7fr)
- **Typography:** 
  - Title: `clamp(64px, 11vw, 196px)` Italiana display font
  - Animated line-by-line reveal
  - "rituals" and "glow" words use Cormorant Garamond italic in lavender-600
- **Background Elements:**
  - Two blurred gradient blobs (lavender-300 at ~520px, rose at ~420px)
  - `filter: blur(100px)` and `opacity: 0.55`
- **Editorial Cards:**
  - Center: Large product placeholder (aspect ratio 3/4, min-height 560px)
  - Chip overlay with product info and CTA button
  - Right column: stacked "Still life" card and "07 Categories" dark card
- **CTAs:**
  - "Shop the edit" pill button (btn-primary class)
  - "Our story →" link with underline
- **Brand Marquee:**
  - Infinite scroll animation
  - Serif italic font for brand names
  - Lavender dot separators

**File Changes:**
- `src/features/home/home-hero.tsx` - Restructure to match mockup layout
- `src/styles/globals.css` - Add/update hero-e2 styles with animations

---

### 2. Announcement Bar

**Current State:** Functional but lacks refinement

**Target State:**
- Refined spacing (padding: 10px 24px)
- Smaller uppercase typography (11px, letter-spacing: 0.18em)
- Additional items: "Community", "Shop Mon-Sat", "New arrivals"
- Dark background (var(--ink)) with lavender text

**File Changes:**
- `src/features/shared/layouts.tsx` - Update announcement bar content
- `src/styles/globals.css` - Refine typography and spacing

---

### 3. Header/Navigation

**Current State:** Basic header with navigation

**Target State:**
- Better logo treatment with circular mark
- Thin, minimal icon style for search/account/wishlist/cart
- Proper spacing and hover states
- Active state indicators

**File Changes:**
- `src/features/shared/layouts.tsx` - Update header structure
- `src/styles/globals.css` - Refine icon button styles

---

### 4. Trust/Features Section (Promise Section)

**Current State:** Card-based layout

**Target State:**
- Luxury info strip aesthetic
- Thinner, more refined icons (44px circles with lavender-400 border)
- Better vertical spacing
- Divider lines top and bottom
- Cream background

**File Changes:**
- `src/features/home/promise-section.tsx` - Update layout
- `src/styles/globals.css` - Refine promise styles

---

### 5. Category Rail (NEW SECTION)

**Current State:** Missing completely

**Target State:**
- Section heading: "Find your next favourite"
- 6-column grid of category cards
- Each card:
  - Pastel background (lavender/cream/rose)
  - Category label (Italiana display font, 18px)
  - Item count (Manrope label font, 11px)
  - Hover scale effect
- Categories: Skincare, Haircare, Bodycare, Fragrance, Makeup, Wellness, Gifts

**File Changes:**
- NEW: `src/features/home/category-rail-section.tsx`
- `src/router.tsx` - Add to home route
- `src/styles/globals.css` - Add cat-rail styles

---

### 6. Featured Products Section

**Current State:** Basic product grid

**Target State:**
- Section heading: "What Accra is reaching for"
- Product cards with:
  - Pastel image backgrounds with striped placeholder pattern
  - Product badges (top left)
  - Wishlist icon (top right, circular white button)
  - Price layout with strikethrough for sale prices
  - Star ratings
  - "Add to cart" button that appears on hover
  - Proper hover states (image scale 1.03)
- Better grid spacing (32px column gap, 24px row gap)

**File Changes:**
- `src/features/home/featured-products.tsx` - Update card rendering
- `src/styles/globals.css` - Add pcard styles with variants

---

### 7. Journal Section

**Current State:** Basic cards

**Target State:**
- Category tags above title
- Article metadata (category · time · date)
- Proper card proportions
- Striped placeholder pattern for images
- "Read story →" link

**File Changes:**
- `src/features/home/journal-section.tsx` - Add metadata and tags
- `src/styles/globals.css` - Refine journal card styles

---

### 8. Testimonials Section

**Current State:** Functional but lacks polish

**Target State:**
- Lavender-50 background (var(--surface))
- Smaller label: "FROM OUR PEOPLE" (eyebrow style)
- Large quotation typography (clamp(26px, 3.5vw, 48px))
- Cormorant Garamond italic
- Carousel dots with active state (lavender-700, width: 24px)

**File Changes:**
- `src/features/home/testimonials-section.tsx` - Update styling
- `src/styles/globals.css` - Refine testimonials styles

---

### 9. Newsletter Section

**Current State:** Basic form

**Target State:**
- Full dark section (var(--ink) background)
- Two-column layout:
  - Left: Large heading with lavender italic em text
  - Right: Email input with integrated button
- Subtle lavender bottom border on focus

**File Changes:**
- `src/features/home/newsletter-section.tsx` - Restructure layout
- `src/styles/globals.css` - Refine nl styles

---

### 10. Footer

**Current State:** Basic footer

**Target State:**
- Two-column top section (brand info + navigation)
- Store location block with icons
- Social icons with hover states
- Better spacing and divider lines
- Responsive grid (4 cols → 2 cols → 1 col)

**File Changes:**
- `src/features/shared/layouts.tsx` - Update footer structure
- `src/styles/globals.css` - Refine footer styles

---

## Component Architecture

### New Components Needed
1. `CategoryRailSection` - Category cards grid
2. `PromoBannerSection` - Dark promotional banner (if needed)

### Existing Components to Update
1. `HomeHero` - Complete restructure
2. `PromiseSection` - Refine layout
3. `FeaturedProducts` - Update card styling
4. `JournalSection` - Add metadata
5. `TestimonialsSection` - Refine styling
6. `NewsletterSection` - Restructure

---

## Design Tokens Reference

All changes use existing design tokens in `:root`:
- Colors: lavender palette, cream, paper, ink variants
- Typography: Italiana (display), Cormorant Garamond (serif), Epilogue (body), Manrope (label)
- Spacing: `var(--gut)`, `var(--max)`
- Motion: `var(--ease)`, `var(--dur)`
- Radius: `var(--radius)`, `var(--radius-lg)`

---

## Responsive Breakpoints

- **Desktop:** > 1100px (full layout)
- **Tablet:** 720px - 1100px (2-column grids)
- **Mobile:** < 720px (1-column stacks)

---

## Animation Requirements

1. **Hero text:** Line-by-line reveal (heroRise keyframes)
2. **Hero blobs:** Pulse animation
3. **Marquee:** Infinite scroll (50s linear)
4. **Cards:** Scale on hover (500-600ms ease)
5. **Buttons:** Transform on hover (translateY -1px)

---

## Implementation Priority

1. **Phase 1 (Hero):** Restructure hero section with proper layout and animations
2. **Phase 2 (Categories):** Add missing category rail section
3. **Phase 3 (Products):** Update product card styling
4. **Phase 4 (Polish):** Refine all other sections (journal, testimonials, newsletter, footer)
5. **Phase 5 (Details):** Announcement bar, header refinements

---

## Success Criteria

- [ ] Hero matches mockup layout exactly (3-column editorial)
- [ ] All typography matches mockup (sizes, weights, fonts)
- [ ] All spacing matches mockup (gaps, padding, margins)
- [ ] All colors match mockup (lavender palette, gradients)
- [ ] All animations match mockup (reveals, hovers, marquees)
- [ ] All sections present (including new category rail)
- [ ] All interactions preserved (buttons, links, navigation)
- [ ] Responsive behavior matches mockup (breakpoints, stacks)
