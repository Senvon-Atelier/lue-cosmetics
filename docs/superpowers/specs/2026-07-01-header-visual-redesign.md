# Header Visual Redesign — Exact Port from Mockup

**Date:** 2026-07-01
**Status:** Approved for implementation
**Scope:** Layout components (Header → Footer → Shared components)

## Overview

Port the header visual design exactly from the `/rue` mockup to the production application while preserving all existing functionality and state management.

## Current State

**Existing Header Components:**
- `Header` in `frontend/src/features/shared/layouts/header.tsx`
- `Brand` component with Rue mark
- Integration with Auth provider (user authentication state)
- Integration with Cart provider (cart count)
- TanStack Router navigation
- Sticky header with blur backdrop
- Desktop nav: Home, Shop, About
- Icon buttons: search, account, cart with badge

**Existing CSS:**
- Header styles already in `globals.css`
- Design tokens (colors, fonts, spacing) already defined
- Responsive breakpoints in place

## Design Specification

### 1. Announcement Bar Component

**New File:** `frontend/src/features/shared/layouts/announcement-bar.tsx`

**Visual Requirements:**
- Dark background: `var(--ink)`
- Text color: `var(--lavender-200)` (or `var(--cream)` with reduced opacity per mockup)
- Font: `var(--font-label)`, 11px, font-weight 500, letter-spacing 0.18em, uppercase
- Padding: 10px 24px
- Full-width, text-align center

**Animation:**
- Marquee scrolling animation: 40s linear infinite
- Track contains messages duplicated 2x for seamless loop
- Each message separated by dot separator (4px circle, `var(--lavender-400)`)

**Messages (from mockup):**
- "Free delivery in Accra over GHS 250"
- "Community 18, Spintex — adjacent KFC"
- "Shop Mon–Sat · 9am–8pm"
- "New Rue Atelier fragrances have landed"

**Component Structure:**
```tsx
export function AnnouncementBar() {
  const messages = [
    "Free delivery in Accra over GHS 250",
    "Community 18, Spintex — adjacent KFC",
    "Shop Mon–Sat · 9am–8pm",
    "New Rue Atelier fragrances have landed"
  ];
  // Render track with duplicated messages
  // Apply marquee animation
}
```

### 2. Header Component Updates

**File:** `frontend/src/features/shared/layouts/header.tsx`

**Add to Desktop Nav:**
- Journal link (after "About")
- Route: `/journal` or `/blog` (to be determined based on app structure)

**Add to Header Actions:**
- Wishlist button (before cart button)
- Mobile menu button (visible on mobile only)
- Both use existing `header-icon-btn` class

**Wishlist Button:**
- Icon: `heart` size 20
- Badge: Shows `wishlistCount` from state
- Same badge positioning as cart (top: 2px, right: 2px)

**Mobile Menu Button:**
- Icon: `menu` size 20
- Only visible on mobile (media query at 720px)
- Opens mobile navigation drawer (functionality to be preserved/implemented)

**Badge Positioning (exact match to mockup):**
```css
.badge {
  position: absolute;
  top: 2px;
  right: 2px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  background: var(--lavender-700);
  color: white;
  border-radius: 999px;
  font-family: var(--font-label);
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
```

**Hover States:**
- All icon buttons: `background: var(--lavender-100)` on hover
- Already in CSS, ensure no conflicts

### 3. Cart Provider Updates

**File:** `frontend/src/features/cart/cart-provider.tsx` (or separate wishlist context)

**Add:**
- `wishlistCount` state
- `addToWishlist` / `removeFromWishlist` functions
- Or create separate `WishlistProvider` if preferred

### 4. Root Layout Updates

**File:** `frontend/src/features/shared/layouts/root-layout.tsx`

**Change:**
```tsx
export function RootLayout() {
  const [isCartOpen, setIsCartOpen] = useState(false);

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <AnnouncementBar />  {/* NEW */}
      <Header onCartOpen={() => setIsCartOpen(true)} />
      <main>
        <Outlet />
      </main>
      <Footer />
      <CartDrawer open={isCartOpen} onClose={() => setIsCartOpen(false)} />
    </div>
  );
}
```

### 5. CSS Updates

**File:** `frontend/src/styles/globals.css`

**Add Announcement Bar Styles:**
```css
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
  from { transform: translateX(0); }
  to { transform: translateX(-50%); }
}
```

**Ensure Badge Styles Match:**
- Update or add `.badge` class to match mockup exactly
- Remove any Tailwind badge classes that conflict

## Functional Requirements

### Preserved Functionality

All existing functionality MUST continue working:

1. **Navigation:**
   - Home link goes to `/`
   - Shop link goes to `/shop`
   - About link goes to `/about`
   - Journal link goes to appropriate route

2. **Authentication:**
   - Account button links to `/account` if authenticated
   - Account button links to `/login` if not authenticated
   - Auth state from `useAuth()` hook

3. **Cart:**
   - Cart button opens drawer
   - Badge shows correct item count
   - Cart state from `useCart()` hook

4. **Search:**
   - Search button opens search overlay
   - Functionality to be preserved

### New Functionality

1. **Wishlist:**
   - Wishlist button toggles wishlist state
   - Badge shows wishlist count
   - Click goes to `/account/wishlist`

2. **Mobile Menu:**
   - Mobile menu button opens navigation drawer
   - Implementation matches existing drawer pattern

3. **Announcement Bar:**
   - Auto-scrolling marquee
   - No user interaction required

## Responsive Behavior

### Desktop (720px and above)
- Show: Announcement bar, full header with all nav links, all icon buttons
- Hide: Mobile menu button

### Mobile (below 720px)
- Show: Announcement bar, condensed header, mobile menu button
- Hide: Desktop nav links
- Layout: Grid becomes `auto 1fr auto` (menu button | brand | icon buttons)

## Implementation Checklist

1. [ ] Create `AnnouncementBar` component
2. [ ] Add marquee animation CSS
3. [ ] Update `Header` component with Journal link
4. [ ] Add wishlist button to header
5. [ ] Add mobile menu button to header
6. [ ] Update badge positioning to match mockup
7. [ ] Add wishlist state to cart provider (or create separate context)
8. [ ] Update `RootLayout` to include `AnnouncementBar`
9. [ ] Test responsive behavior at 720px breakpoint
10. [ ] Verify all nav links work correctly
11. [ ] Verify auth state affects account button correctly
12. [ ] Verify cart badge shows correct count
13. [ ] Verify wishlist badge shows correct count
14. [ ] Test hover states on all interactive elements

## Testing Notes

- All existing tests for Header component should continue passing
- No breaking changes to props or interfaces
- Visual regression testing against mockup

## Next Steps

After header implementation:
1. Footer visual port (same approach: exact match to mockup)
2. Shared component library updates (Button, Icon, Placeholder)
3. Page-level components (Home, Shop, Product, etc.)

## Out of Scope

- Mobile navigation drawer implementation (use existing pattern)
- Search overlay implementation (preserve existing)
- Wishlist backend integration (state only for now)
- Announcements CMS/static content management (hardcoded for now)

## References

- Mockup header: `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/src/shared.jsx` (lines 47-90)
- Mockup styles: `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/styles.css` (lines 206-284)
- Current header: `frontend/src/features/shared/layouts/header.tsx`
- Current styles: `frontend/src/styles/globals.css`
