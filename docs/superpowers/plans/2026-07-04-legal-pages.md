# Legal / Policy Pages (Tranche 5) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship five legal/policy pages (Privacy, Terms, Cookies, Shipping & delivery, Returns & refunds) at `/legal/$slug`, driven by a content registry and a shared layout, and wire the footer's dead links to them.

**Architecture:** One dynamic route `/legal/$slug` renders a presentational `LegalPageView` that reads an ordered registry (`src/content/legal.tsx`). The view draws the mockup's hero + sticky-sidebar + body shell; the sidebar is auto-built from the registry with the active entry marked. Content is realistic prose grounded in the app's real mechanics (Paystack, GHS, `shipping_config.json` rates, the 16 regions, `STORE_INFO`).

**Tech Stack:** React 18, TanStack Router v1 (code-based; param route + `useParams`), plain CSS ported from the mockup, vitest + @testing-library/react (jsdom pragma, same setup as `add-toast.test.tsx`).

**Spec:** `docs/superpowers/specs/2026-07-04-legal-pages-design.md` — binding. **Mockup CSS (read-only, never modify):** `frontend/reference/legacy-css/legal.css`.

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."`. Commit subjects in the branch's `feat(frontend): ...` style. NO AI attribution, NO Co-Authored-By, NO session URLs.
- Frontend-only; nothing under `backend/` or `frontend/src/lib/api/generated/` changes.
- Honest data: content is realistic but store-consistent. No invented guarantees, no generic boilerplate. Shipping figures MUST match `backend/config/shipping_config.json`: flat **GHS 25** (`flat_rate_ghs_minor: 2500`), free over **GHS 500** (`free_over_ghs_minor: 50000`). Payments are **Paystack**; the store never stores card details. Contact details come from `STORE_INFO` — do not hardcode a second copy of the address/phone/hours strings.
- Class parity with the mockup; adapted rules under a `/* adapted */` banner; every new selector defined once (dedup — grep before appending). `.safety-grid`, `.safety-card`, `.patch-steps`, `.patch-step` are NOT ported.
- Actionable elements are `<button>`/`<Link>` — no bare `<a onClick>` and no `href="#"`.
- Per-task gate: `pnpm typecheck` zero errors. Full gate (typecheck/lint/vitest/build) in the final task. Commands run from `ruecosmetics/frontend/`.

---

### Task 1: Content registry — `src/content/legal.tsx`

**Files:**
- Create: `frontend/src/content/legal.tsx`
- Test: `frontend/src/content/legal.test.tsx`

**Interfaces:**
- Consumes: `STORE_INFO` from `src/content/store-info.ts` (`{ addressLine1, addressLine2, phone, hours }`).
- Produces:
  - `type LegalPage = { slug: string; navLabel: string; title: ReactNode; lastUpdated: string; lead: string; body: ReactNode }`
  - `const LEGAL_PAGES: readonly LegalPage[]` — five entries, in sidebar order: `privacy`, `terms`, `cookies`, `shipping`, `returns`.
  - `function getLegalPage(slug: string): LegalPage | undefined`

- [ ] **Step 1: Write the failing test** — `frontend/src/content/legal.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { LEGAL_PAGES, getLegalPage } from './legal';

describe('legal registry', () => {
  it('has the five expected pages in order', () => {
    expect(LEGAL_PAGES.map((p) => p.slug)).toEqual([
      'privacy',
      'terms',
      'cookies',
      'shipping',
      'returns',
    ]);
  });

  it('every page has a navLabel, title, lastUpdated, lead, and body', () => {
    for (const p of LEGAL_PAGES) {
      expect(p.navLabel).toBeTruthy();
      expect(p.title).toBeTruthy();
      expect(p.lastUpdated).toBeTruthy();
      expect(p.lead.length).toBeGreaterThan(0);
      expect(p.body).toBeTruthy();
    }
  });

  it('slugs are unique', () => {
    const slugs = LEGAL_PAGES.map((p) => p.slug);
    expect(new Set(slugs).size).toBe(slugs.length);
  });

  it('getLegalPage returns the entry for a known slug and undefined otherwise', () => {
    expect(getLegalPage('privacy')?.navLabel).toBe('Privacy');
    expect(getLegalPage('nope')).toBeUndefined();
  });
});
```

- [ ] **Step 2: Run it to make sure it fails**

Run: `pnpm vitest run src/content/legal.test.tsx`
Expected: FAIL — cannot resolve `./legal`.

- [ ] **Step 3: Implement the registry** — `frontend/src/content/legal.tsx`. Transcribe verbatim; the copy is the deliverable. Body markup uses only ported classes (`callout`, `callout warn`, `callout info`, `contact-card`, plus base `h2`/`h3`/`p`/`ul`/`li`/`strong` under `.legal-body`).

```tsx
import type { ReactNode } from 'react';
import { STORE_INFO } from './store-info';

export type LegalPage = {
  slug: string;
  navLabel: string;
  title: ReactNode;
  lastUpdated: string;
  lead: string;
  body: ReactNode;
};

function ContactCard() {
  return (
    <div className="contact-card">
      <div className="k">Get in touch</div>
      <div className="v">
        {STORE_INFO.addressLine1}
        <br />
        {STORE_INFO.addressLine2}
        <br />
        {STORE_INFO.phone} · {STORE_INFO.hours}
      </div>
    </div>
  );
}

export const LEGAL_PAGES: readonly LegalPage[] = [
  {
    slug: 'privacy',
    navLabel: 'Privacy',
    title: (
      <>
        Privacy <em>policy</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'This policy explains what Rue Cosmetics collects when you shop with us, how we use it, and the choices you have. It is written to match how this store actually works — nothing more.',
    body: (
      <>
        <h2>What we collect</h2>
        <p>
          When you create an account we store your <strong>name</strong> and{' '}
          <strong>email</strong>. When you place an order we store your{' '}
          <strong>delivery address</strong> and <strong>order history</strong>{' '}
          so you can track and reorder. We collect only what we need to fulfil
          your orders.
        </p>
        <h2>Payments</h2>
        <p>
          Payments are processed by <strong>Paystack</strong>. Your full card
          details are entered on Paystack&rsquo;s secure checkout — Rue never
          sees or stores them. We keep a payment reference so we can match a
          payment to your order.
        </p>
        <h2>Cookies</h2>
        <p>
          We use only essential session and cart cookies. See our{' '}
          <strong>Cookie policy</strong> for details. We do not run
          advertising or third-party tracking cookies.
        </p>
        <h2>Your choices</h2>
        <p>
          You can view and update your details any time from your account. To
          delete your account and associated data, contact us using the details
          below.
        </p>
        <ContactCard />
      </>
    ),
  },
  {
    slug: 'terms',
    navLabel: 'Terms',
    title: (
      <>
        Terms <em>of service</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'These terms govern your use of the Rue Cosmetics store and the orders you place with us. By placing an order you agree to them.',
    body: (
      <>
        <h2>Orders &amp; acceptance</h2>
        <p>
          Placing an order is an offer to buy. An order is accepted once payment
          is confirmed; until then we may decline or cancel it (for example, if
          an item is unavailable).
        </p>
        <h2>Pricing</h2>
        <p>
          All prices are shown in <strong>Ghana cedis (GHS)</strong> and include
          any applicable taxes. We take care to price accurately; if a clear
          pricing error occurs we will contact you before proceeding.
        </p>
        <h2>Payment</h2>
        <p>
          Payment is taken through <strong>Paystack</strong>. Your order is
          confirmed once Paystack reports a successful charge.
        </p>
        <h2>Availability</h2>
        <p>
          Stock is limited and can sell out. If we cannot fulfil an item after
          payment, we will refund it to your original payment method.
        </p>
        <h2>Liability</h2>
        <p>
          To the extent permitted by law, Rue Cosmetics is not liable for
          indirect or consequential loss arising from use of the store. Nothing
          here limits rights you have under Ghanaian consumer law.
        </p>
        <h2>Governing law</h2>
        <p>These terms are governed by the laws of <strong>Ghana</strong>.</p>
      </>
    ),
  },
  {
    slug: 'cookies',
    navLabel: 'Cookies',
    title: (
      <>
        Cookie <em>policy</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'Cookies are small files a site stores in your browser. We keep our use of them to the minimum needed to run the store.',
    body: (
      <>
        <div className="callout info">
          Rue uses <strong>essential cookies only</strong> — no advertising or
          cross-site tracking.
        </div>
        <h2>What we use</h2>
        <p>
          A <strong>session cookie</strong> keeps you signed in, and a{' '}
          <strong>cart cookie</strong> remembers the items in your bag between
          visits. These are required for the store to function.
        </p>
        <h2>What we don&rsquo;t use</h2>
        <p>
          We do not set third-party advertising, profiling, or cross-site
          tracking cookies.
        </p>
        <h2>Managing cookies</h2>
        <p>
          You can clear or block cookies in your browser settings. Blocking the
          essential cookies above will stop the cart and sign-in from working.
        </p>
      </>
    ),
  },
  {
    slug: 'shipping',
    navLabel: 'Shipping & delivery',
    title: (
      <>
        Shipping <em>&amp; delivery</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'How and where we deliver, what it costs, and how long it takes.',
    body: (
      <>
        <div className="callout">
          Delivery times are estimates from dispatch and can vary with courier
          and location.
        </div>
        <h2>Rates</h2>
        <p>
          A flat delivery fee of <strong>GHS 25</strong> applies to every order.
          Orders of <strong>GHS 500 or more ship free</strong>.
        </p>
        <h2>Coverage</h2>
        <p>
          We deliver to all <strong>16 regions of Ghana</strong>. You choose
          your region and enter your address at checkout.
        </p>
        <h2>Timeframes</h2>
        <p>
          Greater Accra typically arrives within{' '}
          <strong>1&ndash;3 working days</strong>; other regions typically take{' '}
          <strong>3&ndash;7 working days</strong> after dispatch.
        </p>
        <h2>Tracking</h2>
        <p>
          You can follow your order&rsquo;s status any time from{' '}
          <strong>your account</strong>.
        </p>
      </>
    ),
  },
  {
    slug: 'returns',
    navLabel: 'Returns & refunds',
    title: (
      <>
        Returns <em>&amp; refunds</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'We want you to love your order. Because cosmetics are personal-care products, some hygiene limits apply to returns.',
    body: (
      <>
        <div className="callout warn">
          For safety and hygiene, <strong>opened or used</strong> cosmetics
          cannot be returned unless they arrived damaged or faulty.
        </div>
        <h2>Return window</h2>
        <p>
          You can return <strong>unopened, unused</strong> items in their
          original packaging within <strong>7 days</strong> of delivery.
        </p>
        <h2>How to start a return</h2>
        <p>
          Contact us with your order number and we&rsquo;ll guide you through
          the next steps.
        </p>
        <h2>Refunds</h2>
        <p>
          Approved refunds are issued through <strong>Paystack</strong> to your
          original payment method, usually within a few working days of us
          receiving the item.
        </p>
        <h2>Damaged or wrong items</h2>
        <p>
          If an item arrives damaged or we sent the wrong product, we cover it —
          contact us and we&rsquo;ll make it right at no cost to you.
        </p>
        <ContactCard />
      </>
    ),
  },
];

export function getLegalPage(slug: string): LegalPage | undefined {
  return LEGAL_PAGES.find((p) => p.slug === slug);
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `pnpm vitest run src/content/legal.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 5: Typecheck**

Run: `pnpm typecheck`
Expected: zero errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/content/legal.tsx frontend/src/content/legal.test.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): legal content registry — 5 store-consistent policy pages"
```

---

### Task 2: LegalPageView + CSS port + route

**Files:**
- Create: `frontend/src/features/content/legal-page.tsx`
- Create: `frontend/src/features/content/legal-page.test.tsx`
- Modify: `frontend/src/styles/globals.css` (append `.legal-*` port)
- Modify: `frontend/src/styles/account.css` (remove stale placeholder comment on line ~4)
- Modify: `frontend/src/router.tsx` (add `/legal/$slug` route + tree entry)

**Interfaces:**
- Consumes: `LEGAL_PAGES`, `getLegalPage`, `type LegalPage` from `src/content/legal` (Task 1). `Link`, `useParams` from `@tanstack/react-router`. `pageShell` and `storefrontLayoutRoute` in `router.tsx`.
- Produces: `function LegalPageView({ slug }: { slug: string }): JSX.Element` (default-less named export). A route matching `/legal/$slug`.

- [ ] **Step 1: Append the CSS** to `frontend/src/styles/globals.css`. First grep to confirm none of these selectors already exist (`grep -nE "\.legal-|\.callout|\.contact-card" src/styles/globals.css src/styles/pages.css` → expect no matches). Then append this block verbatim (a straight port of the used rules from `frontend/reference/legacy-css/legal.css`):

```css
/* ========== Legal / policy pages (ported from mockup legal.css) ========== */

.legal-hero {
  padding: 80px var(--gut) 40px;
  background: var(--surface);
  border-bottom: 1px solid var(--line);
}
.legal-hero-inner { max-width: var(--max); margin: 0 auto; }
.legal-hero h1 { font-family: var(--font-display); font-weight: 400; font-size: clamp(40px, 6vw, 88px); margin: 8px 0 0; letter-spacing: 0.005em; }
.legal-hero h1 em { font-family: var(--font-serif); font-style: italic; color: var(--lavender-700); font-weight: 300; }
.legal-hero .meta { font-family: var(--font-label); font-size: 12px; color: var(--ink-muted); letter-spacing: 0.1em; text-transform: uppercase; margin-top: 20px; }

.legal-wrap { display: grid; grid-template-columns: 260px 1fr; gap: 64px; max-width: var(--max); margin: 0 auto; padding: 64px var(--gut); }
@media (max-width: 900px) { .legal-wrap { grid-template-columns: 1fr; gap: 32px; padding: 40px var(--gut); } }
.legal-side { position: sticky; top: 80px; align-self: start; }
.legal-side h4 { font-family: var(--font-label); font-size: 10px; font-weight: 700; letter-spacing: 0.2em; text-transform: uppercase; color: var(--ink-muted); margin: 0 0 14px; }
.legal-side a { display: block; padding: 10px 0; border-top: 1px solid var(--line); font-family: var(--font-label); font-size: 13px; color: var(--ink-soft); cursor: pointer; }
.legal-side a:last-child { border-bottom: 1px solid var(--line); }
.legal-side a:hover { color: var(--ink); }
.legal-side a.active { color: var(--ink); font-weight: 700; }

.legal-body { max-width: 760px; }
.legal-body .lead { font-size: 17px; color: var(--ink-soft); line-height: 1.65; margin-bottom: 40px; padding-bottom: 32px; border-bottom: 1px solid var(--line); }
.legal-body h2 { font-family: var(--font-display); font-weight: 400; font-size: clamp(24px, 3vw, 32px); margin: 48px 0 16px; letter-spacing: 0.005em; }
.legal-body h3 { font-family: var(--font-serif); font-weight: 500; font-size: 18px; margin: 24px 0 10px; letter-spacing: 0; }
.legal-body p, .legal-body li { color: var(--ink-soft); line-height: 1.75; font-size: 15px; }
.legal-body p { margin: 0 0 14px; }
.legal-body ul, .legal-body ol { padding-left: 20px; }
.legal-body li { margin-bottom: 6px; }
.legal-body strong { color: var(--ink); }

.callout {
  background: var(--surface); border-left: 3px solid var(--lavender-600);
  padding: 20px 24px; border-radius: 6px; margin: 20px 0;
  font-size: 14px; color: var(--ink-soft); line-height: 1.65;
}
.callout.warn { background: #FFF4D6; border-color: #8A6500; color: #5A4500; }
.callout.info { background: var(--lavender-100); border-color: var(--lavender-700); }

.contact-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; margin-top: 24px; }
@media (max-width: 720px) { .contact-grid { grid-template-columns: 1fr; } }
.contact-card { background: var(--cream); border-radius: 10px; padding: 20px; }
.contact-card .k { font-family: var(--font-label); font-size: 10px; letter-spacing: 0.18em; text-transform: uppercase; color: var(--lavender-700); margin-bottom: 6px; }
.contact-card .v { font-family: var(--font-serif); font-size: 16px; color: var(--ink); }
```

- [ ] **Step 2: Remove the stale placeholder** in `frontend/src/styles/account.css`. The comment near line 4 lists `.legal-* (tranche 5)` as not-yet-ported. Delete just the `.legal-* (tranche 5)` mention from that comment (leave the rest of the comment intact). Verify with `grep -n "legal-" src/styles/account.css` → no matches.

- [ ] **Step 3: Write the failing component test** — `frontend/src/features/content/legal-page.test.tsx`. `Link` requires router context, so mock it to a plain anchor; the view takes `slug` as a prop so no router is needed:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <a className={className}>{children}</a>
  ),
}));

import { LegalPageView } from './legal-page';

describe('LegalPageView', () => {
  it('renders the page title and lead for a known slug', () => {
    render(<LegalPageView slug="privacy" />);
    expect(screen.getByRole('heading', { level: 1 }).textContent).toContain('Privacy');
    expect(screen.getByText(/what Rue Cosmetics collects/i)).toBeTruthy();
  });

  it('marks the active sidebar entry with aria-current', () => {
    render(<LegalPageView slug="terms" />);
    const current = document.querySelector('[aria-current="page"]');
    expect(current?.textContent).toBe('Terms');
  });

  it('renders an in-shell fallback for an unknown slug', () => {
    render(<LegalPageView slug="does-not-exist" />);
    expect(screen.getByText(/could not be found/i)).toBeTruthy();
    // sidebar still renders every page
    expect(screen.getAllByText('Privacy').length).toBeGreaterThan(0);
  });
});
```

- [ ] **Step 4: Run it to make sure it fails**

Run: `pnpm vitest run src/features/content/legal-page.test.tsx`
Expected: FAIL — cannot resolve `./legal-page`.

- [ ] **Step 5: Implement the component** — `frontend/src/features/content/legal-page.tsx`:

```tsx
import { Link } from '@tanstack/react-router';
import { LEGAL_PAGES, getLegalPage } from '../../content/legal';

function Sidebar({ activeSlug }: { activeSlug: string }) {
  return (
    <nav className="legal-side" aria-label="Legal pages">
      <h4>Legal</h4>
      {LEGAL_PAGES.map((p) => {
        const isActive = p.slug === activeSlug;
        return (
          <Link
            key={p.slug}
            to="/legal/$slug"
            params={{ slug: p.slug }}
            className={isActive ? 'active' : undefined}
            aria-current={isActive ? 'page' : undefined}
          >
            {p.navLabel}
          </Link>
        );
      })}
    </nav>
  );
}

export function LegalPageView({ slug }: { slug: string }) {
  const page = getLegalPage(slug);

  if (!page) {
    return (
      <>
        <div className="legal-hero">
          <div className="legal-hero-inner">
            <h1>Not found</h1>
          </div>
        </div>
        <div className="legal-wrap">
          <Sidebar activeSlug="" />
          <div className="legal-body">
            <p>This policy could not be found.</p>
            <p>
              <Link to="/legal/$slug" params={{ slug: 'privacy' }}>
                Back to Privacy policy
              </Link>
            </p>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <div className="legal-hero">
        <div className="legal-hero-inner">
          <h1>{page.title}</h1>
          <div className="meta">Last updated {page.lastUpdated}</div>
        </div>
      </div>
      <div className="legal-wrap">
        <Sidebar activeSlug={page.slug} />
        <div className="legal-body">
          <p className="lead">{page.lead}</p>
          {page.body}
        </div>
      </div>
    </>
  );
}
```

- [ ] **Step 6: Run the component test to verify it passes**

Run: `pnpm vitest run src/features/content/legal-page.test.tsx`
Expected: PASS (3 tests).

- [ ] **Step 7: Add the route** in `frontend/src/router.tsx`. Mirror `productDetailRoute` (which reads a `$slug` param). Add near `aboutRoute`:

```tsx
function LegalRouteComponent() {
  const { slug } = useParams({ from: '/_storefront/legal/$slug' });
  return pageShell(<LegalPageView slug={slug || ''} />);
}

const legalRoute = createRoute({
  getParentRoute: () => storefrontLayoutRoute,
  path: '/legal/$slug',
  component: LegalRouteComponent,
});
```

Add the import at the top with the other `features/content` imports:

```tsx
import { LegalPageView } from './features/content/legal-page';
```

Then add `legalRoute` to the `storefrontLayoutRoute.addChildren([...])` array (the same array that lists `aboutRoute`).

- [ ] **Step 8: Typecheck + full test run**

Run: `pnpm typecheck` → zero errors.
Run: `pnpm vitest run` → all pass (existing + 2 new files).

- [ ] **Step 9: Commit**

```bash
git add frontend/src/features/content/legal-page.tsx frontend/src/features/content/legal-page.test.tsx frontend/src/styles/globals.css frontend/src/styles/account.css frontend/src/router.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): legal page view + route + CSS port — hero/sidebar/body shell"
```

---

### Task 3: Wire the footer links — `footer.tsx`

**Files:**
- Modify: `frontend/src/features/shared/layouts/footer.tsx`

**Interfaces:**
- Consumes: `Link` from `@tanstack/react-router` (already imported in footer, per tranche 4; if not, add the import), the `/legal/$slug` route from Task 2.
- Produces: no exports; live footer links.

- [ ] **Step 1: Replace the dead legal anchors.** In `footer.tsx` the "Visit the shop"/support column has `<a href="#">Shipping & delivery</a>` and the `.footer-legal` block has `<a href="#">Privacy</a>`, `<a href="#">Terms</a>`, `<a href="#">Cookies</a>`. Replace each dead `<a href="#">…</a>` with a `Link`:

```tsx
<Link to="/legal/$slug" params={{ slug: 'shipping' }}>Shipping &amp; delivery</Link>
```
```tsx
<Link to="/legal/$slug" params={{ slug: 'privacy' }}>Privacy</Link>
<Link to="/legal/$slug" params={{ slug: 'terms' }}>Terms</Link>
<Link to="/legal/$slug" params={{ slug: 'cookies' }}>Cookies</Link>
```

Keep each element's existing wrapping (`<li>` in the support column, inline in `.footer-legal`) and classes unchanged — only swap the tag and add the `to`/`params`.

- [ ] **Step 2: Add the Returns & refunds link.** In the same support column as "Shipping & delivery" (the `<ul>` of `<li>` links), add a new item:

```tsx
<li><Link to="/legal/$slug" params={{ slug: 'returns' }}>Returns &amp; refunds</Link></li>
```

- [ ] **Step 3: Confirm no dead links remain.**

Run: `grep -n 'href="#"' src/features/shared/layouts/footer.tsx`
Expected: no matches.

- [ ] **Step 4: Typecheck + lint + tests**

Run: `pnpm typecheck` → zero errors (this confirms every `Link to="/legal/$slug"` has valid `params`).
Run: `pnpm lint` → zero warnings.
Run: `pnpm vitest run` → all pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/shared/layouts/footer.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): wire footer legal links to /legal pages + add returns link"
```

---

### Task 4: Final gate, audits

**Files:** none expected (verification only; fix minimally if the gate surfaces anything in touched files).

- [ ] **Step 1: Class-coverage audit** (from `frontend/`). Every className used by the new files must have a matching CSS selector:

```bash
# Collect classes used in the new/modified files and confirm each is defined in a CSS file.
for c in legal-hero legal-hero-inner legal-wrap legal-side legal-body callout contact-card; do
  echo -n "$c: "; grep -rl "\.$c" src/styles/*.css | wc -l
done
# Template-literal / conditional classes to eyeball: `.callout warn`, `.callout info`,
# `.legal-side a.active` — all defined in globals.css.
```
Expected: each count ≥ 1.

- [ ] **Step 2: No Tailwind residue** in touched files:

```bash
grep -rnE "className=\"[^\"]*\b(flex|grid|gap-|px-|py-|mt-|mb-|text-(xs|sm|lg|xl)|bg-[a-z]+-[0-9])" \
  src/content/legal.tsx src/features/content/legal-page.tsx src/features/shared/layouts/footer.tsx || echo "clean"
```
Expected: `clean`.

- [ ] **Step 3: Full gate**

Run: `pnpm typecheck && pnpm lint && pnpm vitest run && pnpm build`
Expected: typecheck 0 errors; lint 0 warnings; vitest all pass (existing + `legal.test.tsx` + `legal-page.test.tsx`); build succeeds. Confirm the legal CSS shipped: `grep -c "legal-wrap" dist/assets/*.css` → ≥ 1.

- [ ] **Step 4: Functional smoke** — only if dev servers are already running; otherwise note "deferred to human walkthrough": footer links navigate to each `/legal/<slug>`; sidebar highlights the current page (`aria-current`); switching pages via the sidebar works and back/forward moves between them; `/legal/bogus` shows the in-shell fallback with a working "Back to Privacy policy" link; shipping page shows GHS 25 / free over GHS 500.

- [ ] **Step 5: Commit** (only if Steps 1–3 changed any files; otherwise no commit for this task).

---

## Self-review notes

- **Spec coverage:** 5 pages (Task 1 registry) ✓; `/legal/$slug` route + shared view + sidebar active state + in-shell fallback (Task 2) ✓; CSS port scope incl. skip-list + stale-comment removal (Task 2) ✓; footer wiring incl. new Returns link (Task 3) ✓; content grounded in shipping_config/regions/Paystack/STORE_INFO (Task 1 copy + Global Constraints) ✓; testing incl. title/active-nav/fallback (Task 2 test) ✓; final gate + audits (Task 4) ✓.
- **Type consistency:** `LegalPage`/`LEGAL_PAGES`/`getLegalPage` defined in Task 1 and consumed with the same names/signatures in Task 2; `LegalPageView({ slug })` produced in Task 2 and used by the route (Task 2) and footer depends only on the route path (Task 3).
- **Out of scope confirmed absent:** no markdown pipeline, no `<title>`/meta management, no safety/patch CSS, no admin/CMS.
