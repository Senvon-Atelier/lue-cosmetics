# Footer Visual Redesign — Exact Port from Mockup

**Date:** 2026-07-01
**Status:** Approved for implementation
**Scope:** Layout components — Footer (following Header completion)

## Overview

Port the footer visual design exactly from the `/rue` mockup to the production application while preserving all existing navigation and functionality.

## Current State

**Existing Footer Component:**
- `Footer` in `frontend/src/features/shared/layouts/footer.tsx`
- Uses `.footer-inner` + `.footer-brand` structure
- 4 columns: Shop, Company, Help, Visit
- Social links with circular icons
- Footer bottom with copyright and legal links

**Existing CSS:**
- Footer styles in `globals.css`
- Uses existing design tokens
- Responsive breakpoints at 1024px and 768px

## Design Specification

### Structural Changes

**Mockup Structure:**
```tsx
<footer className="footer">
  <div className="wrap">
    <div className="footer-top">
      <div className="footer-lead">
        {/* brand, blurb, socials */}
      </div>
      <div className="footer-cols">
        {/* 4 columns */}
      </div>
    </div>
    <div className="footer-bottom">
      {/* copyright, legal */}
    </div>
  </div>
</footer>
```

**Current → Mockup Mapping:**
- `.footer-inner` → `.footer-top`
- `.footer-brand` → `.footer-lead`
- `.footer-cols` stays same (4 columns)

### Column 4 Updates: "Visit the shop"

**Mockup has icons with text:**
- Pin icon + "Community 18, Spintex" + "Adjacent KFC, Accra"
- Phone icon + "0594 701 345"
- Clock icon + "Mon–Sat · 9am – 8pm"

**Current has plain text:**
- "Spintex Road, Accra"
- "+233 20 123 4567"
- "Mon-Sat: 10am-7pm"

### CSS Updates Required

1. **Rename classes:**
   - `.footer-inner` → `.footer-top`
   - `.footer-brand` → `.footer-lead`

2. **Update footer-top layout:**
   - Change from `grid-template-columns: 1.2fr 3fr` to allow natural flow
   - Mockup doesn't use grid for footer-top, just natural block layout

3. **Add footer-contact styles:**
   ```css
   .footer-contact li {
     display: flex;
     align-items: flex-start;
     gap: 8px;
     line-height: 1.4;
   }
   .footer-contact li span {
     display: block;
   }
   ```

4. **Update column 4 content:**
   - Add icons (pin, phone, clock)
   - Use specific text from mockup
   - Format phone number as "0594 701 345"
   - Format hours as "Mon–Sat · 9am – 8pm"

5. **Social icon styling:**
   - Mockup doesn't have circular backgrounds for social icons
   - Current: 36px circles with lavender-100 background
   - Mockup: Plain icons, smaller (18px)

## Files to Modify

1. `frontend/src/features/shared/layouts/footer.tsx` — update structure and content
2. `frontend/src/styles/globals.css` — update CSS class names and styling

## Implementation Notes

- Preserve all TanStack Router navigation links
- Keep all routes functional
- No breaking changes to routing
- Only visual updates to match mockup exactly

## Next Steps

After footer implementation:
1. Shared component library updates (Button, Icon, Placeholder)
2. Page-level components (Home, Shop, Product, etc.)
