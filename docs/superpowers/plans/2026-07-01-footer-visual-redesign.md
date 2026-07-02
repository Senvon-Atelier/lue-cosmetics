# Footer Visual Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the footer visual design exactly from the `/rue` mockup to the production application while preserving all existing navigation and functionality.

**Architecture:** Update footer structure to match mockup (footer-top + footer-lead), add icons to column 4, update social icon styling, preserve all TanStack Router links.

**Tech Stack:** React, TanStack Router, existing design tokens, CSS custom properties

## Global Constraints

- **Module path:** `github.com/oti-adjei/ruecosmetics`
- **Working directory:** `/Volumes/Georgie/reformat-audit/Downloads/casestud/ruecosmetics/`
- **Framework:** React with TanStack Router
- **Design tokens:** Already defined in `frontend/src/styles/globals.css`
- **No breaking changes:** All existing navigation must continue working
- **Exact visual match:** Footer must match `/rue/src/shared.jsx` visual design precisely

---

## File Structure

**Modified files:**
- `frontend/src/features/shared/layouts/footer.tsx` — Update structure and column 4 content
- `frontend/src/styles/globals.css` — Update CSS class names and styling

---

### Task 1: Update Footer Structure in Component

**Files:**
- Modify: `frontend/src/features/shared/layouts/footer.tsx`

**Interfaces:**
- Consumes: React hooks, TanStack Router Link, Icon component, Brand component
- Produces: Updated Footer with mockup structure

- [ ] **Step 1: Update footer structure to match mockup**

Open `frontend/src/features/shared/layouts/footer.tsx`

Replace `.footer-inner` with `.footer-top` and `.footer-brand` with `.footer-lead`:

```tsx
export function Footer() {
  return (
    <footer className="footer">
      <div className="wrap">
        <div className="footer-top">
          <div className="footer-lead">
            <div className="footer-brand-logo">
              <Brand />
            </div>
            <p className="footer-blurb">
              Home of authentic beauty and wellness. A shelf of trusted names — and a few of our own — stocked in
              Accra, shipped across Ghana.
            </p>
            <div className="footer-socials">
              <a href="#" className="footer-social-link" aria-label="Instagram">
                <Icon name="instagram" size={18} />
              </a>
              <a href="#" className="footer-social-link" aria-label="TikTok">
                <Icon name="tiktok" size={18} />
              </a>
              <a href="#" className="footer-social-link" aria-label="WhatsApp">
                <Icon name="whatsapp" size={18} />
              </a>
            </div>
          </div>

          <div className="footer-cols">
            {/* Shop column */}
            <div className="footer-col">
              <h5>Shop</h5>
              <ul>
                <li><Link to="/shop">Skincare</Link></li>
                <li><Link to="/shop">Haircare</Link></li>
                <li><Link to="/shop">Fragrance</Link></li>
                <li><Link to="/shop">Bodycare</Link></li>
                <li><Link to="/shop">Sets & Gifts</Link></li>
                <li><Link to="/shop">All products</Link></li>
              </ul>
            </div>

            {/* Company column */}
            <div className="footer-col">
              <h5>Company</h5>
              <ul>
                <li><Link to="/about">About Rue</Link></li>
                <li><Link to="/journal">The Journal</Link></li>
                <li><a href="#">Store locator</a></li>
                <li><a href="#">Careers</a></li>
                <li><a href="#">Press</a></li>
              </ul>
            </div>

            {/* Help column */}
            <div className="footer-col">
              <h5>Help</h5>
              <ul>
                <li><a href="#">Contact us</a></li>
                <li><a href="#">Shipping & delivery</a></li>
                <li><a href="#">Returns</a></li>
                <li><a href="#">FAQs</a></li>
                <li><a href="#">Authenticity</a></li>
              </ul>
            </div>

            {/* Visit column with icons */}
            <div className="footer-col">
              <h5>Visit the shop</h5>
              <ul className="footer-contact">
                <li>
                  <Icon name="pin" size={14} />
                  Community 18, Spintex<br />
                  <span>Adjacent KFC, Accra</span>
                </li>
                <li>
                  <Icon name="phone" size={14} />
                  0594 701 345
                </li>
                <li>
                  <Icon name="clock" size={14} />
                  Mon–Sat · 9am – 8pm
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div className="footer-bottom">
          <div>© 2026 Rue Cosmetics Ghana · All rights reserved</div>
          <div className="footer-legal">
            <a href="#">Privacy</a>
            <a href="#">Terms</a>
            <a href="#">Cookies</a>
          </div>
        </div>
      </div>
    </footer>
  );
}
```

- [ ] **Step 2: Verify component compiles**

Run: Check for TypeScript errors

Expected: No TypeScript errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/features/shared/layouts/footer.tsx
git commit -m "feat(footer): update structure to match mockup

- Rename footer-inner to footer-top
- Rename footer-brand to footer-lead
- Add icons to column 4 (pin, phone, clock)
- Update contact details: address, phone, hours
- Preserve all TanStack Router navigation links"
```

---

### Task 2: Update Footer CSS

**Files:**
- Modify: `frontend/src/styles/globals.css`

**Interfaces:**
- Consumes: Existing design tokens
- Produces: Updated CSS classes for footer

- [ ] **Step 1: Update footer class names**

Find and replace in `globals.css`:

Replace `.footer-inner` with `.footer-top`:
```css
.footer-top {
  display: block;
  margin-bottom: 40px;
}
```

Replace `.footer-brand` references with `.footer-lead`:
```css
.footer-lead {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
```

- [ ] **Step 2: Add footer-contact styles**

Add after `.footer-cols` styles:

```css
/* ----- Footer Contact ----- */
.footer-contact {
  list-style: none;
  padding: 0;
  margin: 0;
}

.footer-contact li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  line-height: 1.4;
  font-size: 14px;
  color: var(--ink-soft);
  margin-bottom: 12px;
}

.footer-contact li:last-child {
  margin-bottom: 0;
}

.footer-contact li span {
  display: block;
}
```

- [ ] **Step 3: Update footer-bottom styling**

Ensure `.footer-bottom` has correct flex layout:

```css
.footer-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 24px;
  border-top: 1px solid var(--line);
  font-size: 12px;
  color: var(--ink-muted);
}
```

- [ ] **Step 4: Update footer-legal styling**

```css
.footer-legal {
  display: flex;
  gap: 24px;
}

.footer-legal a {
  color: var(--ink-muted);
  text-decoration: none;
  transition: color var(--dur) var(--ease);
}

.footer-legal a:hover {
  color: var(--ink);
}
```

- [ ] **Step 5: Verify CSS is valid**

Run: No command needed — CSS changes are static

Expected: No syntax errors

- [ ] **Step 6: Commit**

```bash
git add frontend/src/styles/globals.css
git commit -m "feat(styles): update footer CSS to match mockup

- Rename footer-inner to footer-top, footer-brand to footer-lead
- Add footer-contact styles for icons + text layout
- Update footer-bottom and footer-legal styling
- Preserve existing responsive breakpoints"
```

---

### Task 3: Testing and Verification

**Files:**
- No file modifications — testing and verification only

**Interfaces:**
- Consumes: Updated Footer component and CSS
- Produces: Verified footer matching mockup

- [ ] **Step 1: Verify footer renders correctly**

Run: Open browser and inspect footer

Check:
- [ ] Background is var(--cream)
- [ ] Structure: footer-top with footer-lead + footer-cols
- [ ] 4 columns in footer-cols
- [ ] footer-bottom with copyright + legal links

- [ ] **Step 2: Verify column 4 has icons**

Check:
- [ ] Pin icon displayed
- [ ] Phone icon displayed
- [ ] Clock icon displayed
- [ ] Icons aligned with text
- [ ] Text matches mockup exactly

- [ ] **Step 3: Verify all navigation links work**

Run: Click each link

Check:
- [ ] Shop links go to `/shop`
- [ ] About Rue goes to `/about`
- [ ] Journal link goes to `/journal`
- [ ] Links navigate correctly

- [ ] **Step 4: Verify responsive behavior**

Run: Resize browser to 1024px and 768px

Check:
- [ ] At 1024px: footer-cols becomes 3 columns (from current CSS)
- [ ] At 768px: footer-cols becomes 2 columns, footer-bottom stacks
- [ ] All content remains readable

- [ ] **Step 5: Check visual match to mockup**

Run: Compare side-by-side with `/rue/src/shared.jsx`

Check:
- [ ] Spacing matches mockup
- [ ] Font sizes match mockup
- [ ] Colors match mockup
- [ ] Layout matches mockup

- [ ] **Step 6: Final commit**

```bash
git commit --allow-empty -m "test(footer): verify visual redesign matches mockup

- Footer structure matches mockup (footer-top + footer-lead)
- Column 4 icons render correctly (pin, phone, clock)
- All navigation links functional
- Responsive behavior preserved
- Visual comparison confirms exact match to /rue mockup"
```

---

## Self-Review Results

**1. Spec coverage:**
- ✅ Footer structure update → Task 1
- ✅ CSS updates → Task 2
- ✅ Testing and verification → Task 3

**2. Placeholder scan:**
- ✅ No "TBD", "TODO", or placeholders found
- ✅ All code blocks contain complete implementation

**3. Type consistency:**
- ✅ Class names consistent (footer-top, footer-lead, footer-contact)
- ✅ Icon names verified (pin, phone, clock exist)

**4. Architecture check:**
- ✅ Each task produces independently testable deliverable
- ✅ No breaking changes to navigation
- ✅ All routes preserved

---

## Post-Implementation

After completing all tasks:

1. **Verify in browser:** Open the production app and compare footer to `/rue/mockup`
2. **Check responsive:** Test at tablet, mobile breakpoints
3. **Test interactions:** Verify all links work correctly
4. **Commit final verification:** Empty commit noting successful visual match

---

**Next Steps:** After footer implementation, proceed to shared component library updates (Button, Icon, Placeholder).
