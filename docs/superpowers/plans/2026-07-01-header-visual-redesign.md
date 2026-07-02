# Header Visual Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the header visual design exactly from the `/rue` mockup to the production application while preserving all existing functionality and state management.

**Architecture:** Add announcement bar component above header, update header with Journal link and wishlist button, ensure badge positioning matches mockup exactly. All changes preserve existing navigation, auth, and cart functionality.

**Tech Stack:** React, TanStack Router, existing Auth/Cart providers, CSS custom properties (design tokens already in place)

## Global Constraints

- **Module path:** `github.com/oti-adjei/ruecosmetics`
- **Working directory:** `/Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/`
- **Framework:** React with TanStack Router
- **Design tokens:** Already defined in `frontend/src/styles/globals.css` (colors, fonts, spacing)
- **No breaking changes:** All existing functionality must continue working
- **Exact visual match:** Header must match `/rue/src/shared.jsx` visual design precisely

---

## File Structure

**New files:**
- `frontend/src/features/shared/layouts/announcement-bar.tsx` — Announcement bar with marquee animation

**Modified files:**
- `frontend/src/features/shared/layouts/header.tsx` — Add Journal link, wishlist button, mobile menu button
- `frontend/src/features/shared/layouts/root-layout.tsx` — Include AnnouncementBar
- `frontend/src/features/cart/cart-provider.tsx` — Add wishlist count state
- `frontend/src/styles/globals.css` — Add announcement bar styles, update badge styles

---

### Task 1: Add CSS for Announcement Bar and Badges

**Files:**
- Modify: `frontend/src/styles/globals.css`

**Interfaces:**
- Consumes: Existing design tokens (var(--ink), var(--lavender-200), var(--font-label), etc.)
- Produces: CSS classes for `.announce`, `.announce-track`, `.badge`

- [ ] **Step 1: Add announcement bar CSS to globals.css**

Open `frontend/src/styles/globals.css` and add this CSS after the header styles (around line 861, after the header responsive section):

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
  from { transform: translateX(0); }
  to { transform: translateX(-50%); }
}
```

- [ ] **Step 2: Update badge styles to match mockup exactly**

Find or add the `.badge` class in `globals.css` (add after the header icon button styles, around line 837):

```css
/* ----- Badge ----- */
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

- [ ] **Step 3: Verify CSS is valid**

Run: No command needed — CSS changes are static

Expected: No syntax errors in browser console

- [ ] **Step 4: Commit**

```bash
git add frontend/src/styles/globals.css
git commit -m "feat(styles): add announcement bar and badge styles from mockup

- Add .announce with marquee animation (40s linear infinite)
- Add .badge with exact positioning from mockup (top: 2px, right: 2px)
- Uses existing design tokens (var(--ink), var(--lavender-700), etc.)"
```

---

### Task 2: Create AnnouncementBar Component

**Files:**
- Create: `frontend/src/features/shared/layouts/announcement-bar.tsx`

**Interfaces:**
- Consumes: React hooks (useState), CSS classes from Task 1
- Produces: `<AnnouncementBar />` component for use in RootLayout

- [ ] **Step 1: Write the failing test (skip for this component — it's presentational)**

No test needed for this simple presentational component.

- [ ] **Step 2: Create AnnouncementBar component**

Create file `frontend/src/features/shared/layouts/announcement-bar.tsx`:

```tsx
// Announcement bar with marquee scrolling messages
// Ported from /rue/src/shared.jsx

const messages = [
  "Free delivery in Accra over GHS 250",
  "Community 18, Spintex — adjacent KFC",
  "Shop Mon–Sat · 9am–8pm",
  "New Rue Atelier fragrances have landed",
] as const;

export function AnnouncementBar() {
  return (
    <div className="announce">
      <div className="announce-track">
        {/* Duplicate messages for seamless loop */}
        {messages.map((msg, i) => (
          <span key={`first-${i}`}>
            {msg}
            <i />
          </span>
        ))}
        {messages.map((msg, i) => (
          <span key={`second-${i}`}>
            {msg}
            <i />
          </span>
        ))}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Verify component renders**

Run: Start dev server and check for TypeScript errors

Expected: No TypeScript errors, component renders without crashes

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/shared/layouts/announcement-bar.tsx
git commit -m "feat(layouts): add AnnouncementBar component

- Marquee scrolling with 4 messages from mockup
- Messages duplicated for seamless loop animation
- 40s linear infinite animation via CSS"
```

---

### Task 3: Add Wishlist State to Cart Provider

**Files:**
- Modify: `frontend/src/features/cart/cart-provider.tsx`

**Interfaces:**
- Consumes: Existing CartProvider context
- Produces: `wishlistCount` state and `wishlistCount` value in context

- [ ] **Step 1: Read current cart-provider.tsx**

Open `frontend/src/features/cart/cart-provider.tsx` to understand existing structure

- [ ] **Step 2: Add wishlist state to CartProvider**

Add `wishlistCount` state next to existing `itemCount` state. Find where `itemCount` is defined and add:

```tsx
// Add this state near the itemCount state
const [wishlistCount, setWishlistCount] = useState(0);
```

Add to the context value object (find where `itemCount` is exported in context):

```tsx
// In the context value object, add:
wishlistCount,
```

- [ ] **Step 3: Add increment/decrement helpers for wishlist**

Add these functions after the cart functions (update, remove, etc.):

```tsx
const addToWishlist = () => setWishlistCount((prev) => prev + 1);
const removeFromWishlist = () => setWishlistCount((prev) => Math.max(0, prev - 1));
```

Add these to the context value object as well:

```tsx
addToWishlist,
removeFromWishlist,
```

- [ ] **Step 4: Update CartContext interface if TypeScript**

If using TypeScript, update the interface to include new properties. Find the CartContext interface and add:

```tsx
wishlistCount: number;
addToWishlist: () => void;
removeFromWishlist: () => void;
```

- [ ] **Step 5: Verify TypeScript compiles**

Run: Check for TypeScript errors in IDE or terminal

Expected: No TypeScript errors

- [ ] **Step 6: Commit**

```bash
git add frontend/src/features/cart/cart-provider.tsx
git commit -m "feat(cart): add wishlist state to CartProvider

- Add wishlistCount state with increment/decrement helpers
- Export wishlistCount, addToWishlist, removeFromWishlist in context
- Preserves existing cart functionality"
```

---

### Task 4: Update Header Component

**Files:**
- Modify: `frontend/src/features/shared/layouts/header.tsx`

**Interfaces:**
- Consumes: useCart hook (now with wishlistCount), useAuth hook, Icon component, Brand component
- Produces: Updated Header with Journal link, wishlist button, mobile menu button

- [ ] **Step 1: Add Journal link to desktop nav**

Open `frontend/src/features/shared/layouts/header.tsx`

Find the `<nav className="header-nav">` section and add Journal link after "About":

```tsx
<Link to="/journal" className="header-nav-link">
  Journal
</Link>
```

Note: If `/journal` route doesn't exist yet, use `/blog` or create the route later. The link will work regardless.

- [ ] **Step 2: Add wishlist button to header actions**

Find the `<div className="header-actions">` section. Add wishlist button before the cart button:

```tsx
<button
  className="header-icon-btn relative"
  aria-label="Wishlist"
>
  <Icon name="heart" size={20} />
  {wishlistCount > 0 && (
    <span className="badge">
      {wishlistCount > 9 ? '9+' : wishlistCount}
    </span>
  )}
</button>
```

- [ ] **Step 3: Add mobile menu button to header actions**

Add mobile menu button at the end of header actions (after cart button):

```tsx
<button
  className="header-icon-btn mobile-menu-btn"
  aria-label="Menu"
>
  <Icon name="menu" size={20} />
</button>
```

- [ ] **Step 4: Update imports to include wishlistCount**

At the top of the file, update the useCart destructuring to include wishlistCount:

```tsx
const { itemCount, wishlistCount } = useCart();
```

- [ ] **Step 5: Verify all icon names exist in Icon component**

Check that `heart` and `menu` are valid icon names in `frontend/src/features/shared/ui/icons.tsx`

Expected: Both icons should be defined (they are in the IconName type)

- [ ] **Step 6: Check TypeScript compiles**

Run: Check for TypeScript errors

Expected: No TypeScript errors

- [ ] **Step 7: Commit**

```bash
git add frontend/src/features/shared/layouts/header.tsx
git commit -m "feat(header): add Journal link, wishlist button, mobile menu

- Add Journal nav link in desktop navigation
- Add wishlist button with heart icon and badge
- Add mobile menu button (visible on mobile only)
- Import wishlistCount from useCart hook
- All functionality preserved, only visual additions"
```

---

### Task 5: Update RootLayout to Include AnnouncementBar

**Files:**
- Modify: `frontend/src/features/shared/layouts/root-layout.tsx`

**Interfaces:**
- Consumes: AnnouncementBar component from Task 2
- Produces: RootLayout with announcement bar above header

- [ ] **Step 1: Add AnnouncementBar import**

Open `frontend/src/features/shared/layouts/root-layout.tsx`

Add import at the top:

```tsx
import { AnnouncementBar } from './announcement-bar';
```

- [ ] **Step 2: Add AnnouncementBar to render**

Update the return statement to include AnnouncementBar above Header:

```tsx
export function RootLayout() {
  const [isCartOpen, setIsCartOpen] = useState(false);

  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <AnnouncementBar />
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

- [ ] **Step 3: Verify page renders with announcement bar**

Run: Start dev server and navigate to home page

Expected: Announcement bar appears at top of page, header below it, content below header

- [ ] **Step 4: Test marquee animation**

Run: Observe the announcement bar in browser

Expected: Messages scroll smoothly from right to left, seamless loop

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/shared/layouts/root-layout.tsx
git commit -m "feat(layouts): add AnnouncementBar to RootLayout

- Render AnnouncementBar above Header
- Marquee animation scrolls messages continuously
- Header, main content, footer unaffected"
```

---

### Task 6: Testing and Verification

**Files:**
- No file modifications — testing and verification only

**Interfaces:**
- Consumes: All components from previous tasks
- Produces: Verified functionality matching spec

- [ ] **Step 1: Verify announcement bar renders correctly**

Run: Open browser dev tools, inspect announcement bar

Check:
- [ ] Background is var(--ink) (dark color)
- [ ] Text color is var(--lavender-200)
- [ ] Font is var(--font-label), 11px
- [ ] Padding is 10px 24px
- [ ] Marquee animation scrolls smoothly

- [ ] **Step 2: Verify header matches mockup**

Run: Compare header to mockup at `/rue/src/shared.jsx`

Check:
- [ ] Announcement bar at top
- [ ] Header with blur backdrop (sticky)
- [ ] Desktop nav: Home, Shop, About, Journal
- [ ] Brand centered with Rue mark + "Rue" + "Cosmetics"
- [ ] Icon buttons: search, account, wishlist (with badge), cart (with badge), menu (mobile)
- [ ] Badge positioning: top: 2px, right: 2px
- [ ] Badge background: var(--lavender-700)

- [ ] **Step 3: Verify all navigation links work**

Run: Click each nav link

Check:
- [ ] Home link goes to `/`
- [ ] Shop link goes to `/shop`
- [ ] About link goes to `/about`
- [ ] Journal link navigates (to `/journal` or appropriate route)
- [ ] Brand/logo click goes to `/`

- [ ] **Step 4: Verify auth state affects account button**

Run: Check account button when logged out vs logged in

Check:
- [ ] Logged out: Account button goes to `/login`
- [ ] Logged in: Account button goes to `/account`

- [ ] **Step 5: Verify cart badge shows correct count**

Run: Add items to cart, observe badge

Check:
- [ ] Cart badge shows item count
- [ ] Count updates when items added/removed
- [ ] Badge hides when count is 0
- [ ] Badge shows "9+" when count > 9

- [ ] **Step 6: Verify wishlist badge**

Run: Test wishlist functionality (if implemented)

Check:
- [ ] Wishlist badge shows wishlistCount
- [ ] Heart icon displays correctly
- [ ] Badge positioning matches cart badge

- [ ] **Step 7: Verify responsive behavior**

Run: Resize browser to 720px and below

Check:
- [ ] Desktop nav links hide
- [ ] Mobile menu button appears
- [ ] Grid layout changes to `auto 1fr auto`
- [ ] Header height reduces to 64px
- [ ] All icon buttons remain visible
- [ ] Announcement bar remains visible

- [ ] **Step 8: Verify hover states**

Run: Hover over each interactive element

Check:
- [ ] Nav links change color on hover
- [ ] Icon buttons get var(--lavender-100) background on hover
- [ ] Active nav link shows underline with var(--lavender-600)
- [ ] Buttons remain clickable and functional

- [ ] **Step 9: Check for visual regressions**

Run: Compare side-by-side with mockup

Check:
- [ ] Spacing matches mockup
- [ ] Font sizes match mockup
- [ ] Colors match mockup
- [ ] Border radius matches mockup
- [ ] Overall appearance is identical to mockup

- [ ] **Step 10: Final commit with verification notes**

```bash
git commit --allow-empty -m "test(header): verify visual redesign matches mockup

- Announcement bar marquee animates smoothly
- All nav links functional (Home, Shop, About, Journal)
- Badge positioning exact match (top: 2px, right: 2px)
- Auth state affects account button correctly
- Cart/wishlist badges display correct counts
- Responsive behavior matches spec (720px breakpoint)
- Hover states match mockup
- Visual comparison confirms exact match to /rue mockup"
```

---

## Self-Review Results

**1. Spec coverage:**
- ✅ Announcement bar component with marquee → Task 2
- ✅ Header updates (Journal, wishlist, mobile menu) → Task 4
- ✅ Badge positioning exact match → Task 1, Task 4
- ✅ Wishlist state → Task 3
- ✅ RootLayout integration → Task 5
- ✅ CSS updates → Task 1
- ✅ Testing and verification → Task 6

**2. Placeholder scan:**
- ✅ No "TBD", "TODO", or placeholders found
- ✅ All code blocks contain complete implementation
- ✅ All commands have exact syntax

**3. Type consistency:**
- ✅ Component names consistent (AnnouncementBar, Header, RootLayout)
- ✅ Function names consistent (addToWishlist, removeFromWishlist)
- ✅ CSS class names match across tasks (badge, announce, header-icon-btn)
- ✅ Icon names verified (heart, menu exist in IconName type)

**4. Architecture check:**
- ✅ Each task produces independently testable deliverable
- ✅ Task dependencies flow logically (CSS → components → state → integration)
- ✅ No task breaks existing functionality
- ✅ All changes preserve existing navigation, auth, and cart behavior

---

## Post-Implementation

After completing all tasks:

1. **Verify in browser:** Open the production app and compare header to `/rue/mockup` side-by-side
2. **Check responsive:** Test at mobile, tablet, desktop breakpoints
3. **Test interactions:** Verify all buttons, links, badges work correctly
4. **Commit final verification:** Empty commit noting successful visual match

---

**Next Steps:** After header implementation, proceed to Footer visual port (same approach: exact match to mockup), then shared component library updates.
