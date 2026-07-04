# Legal / Policy Pages (Tranche 5) — Design Spec

**Date:** 2026-07-04
**Status:** Approved (brainstorming) — pending implementation plan
**Branch:** `ui-tranche-5-legal` (off `main` @ `bc8ccf1`)

## Goal

Ship the storefront's legal/policy pages, the last deferred UI tranche. The footer currently renders four dead links (`href="#"`): **Shipping & delivery**, **Privacy**, **Terms**, **Cookies**. Tranche 5 builds these pages — plus a **Returns & refunds** policy — using the mockup's `.legal-*` layout, and wires every footer link. This closes the honest-data gap tranche 4 left (dead links promising pages that don't exist).

This was explicitly deferred from the account tranche (`docs/superpowers/specs/2026-07-02-account-ui-alignment-design.md:99` and `docs/superpowers/plans/2026-07-02-account-ui-alignment.md:163` both mark `.legal-*` as "tranche 5").

## Scope

**In scope — 5 policy pages:**

1. Privacy
2. Terms
3. Cookies
4. Shipping & delivery
5. Returns & refunds

**Out of scope (YAGNI):**

- No CMS / admin editing of policy content.
- No markdown pipeline or new content-loading dependency.
- No per-page SEO / `<title>` / meta-tag management (the SPA does not manage document head today; not introducing it here).
- No ingredient-safety or patch-test pages (`.safety-grid`, `.patch-step` CSS is **not** ported).
- No real legal review — this is portfolio case-study copy, not lawyer-drafted policy.

## Architecture

Both new surfaces follow existing storefront patterns: code-based TanStack Router route, `pageShell()` wrapper, static content in `src/content/`, page component in `src/features/content/`.

### Route

- **One dynamic route:** `/legal/$slug`, created with `createRoute` parented on the storefront layout route (same parent as `/about`), component wrapped in `pageShell(<LegalPageView />)`.
- The route reads `slug` from params and looks it up in the registry.

### Content registry — `src/content/legal.tsx`

Exports an **ordered** array (order defines sidebar order):

```ts
export type LegalPage = {
  slug: string;          // URL segment, e.g. 'privacy'
  navLabel: string;      // sidebar label (plain string), e.g. 'Privacy'
  title: ReactNode;      // hero <h1> content; may embed an <em> accent per .legal-hero h1 em
  lastUpdated: string;   // display string, e.g. 'July 2026'
  lead: string;          // .legal-body .lead paragraph
  body: ReactNode;       // the policy body JSX (headings, paragraphs, lists, callouts, contact-card)
};

export const LEGAL_PAGES: readonly LegalPage[] = [ /* 5 entries */ ];

export function getLegalPage(slug: string): LegalPage | undefined;
```

Adding or editing a page is a single registry entry. `title` is a `ReactNode` so it can carry the mockup's `<em>` accent (`<h1>Privacy <em>policy</em></h1>`); `navLabel` stays a plain string for the sidebar. The hero renders `{title}` directly inside `<h1>`.

### Page component — `src/features/content/legal-page.tsx`

Renders the mockup shell (classes from `legal.css`):

- `.legal-hero` → `.legal-hero-inner` → `<h1>{title}</h1>` + `.meta` ("Last updated {lastUpdated}").
- `.legal-wrap` (grid) containing:
  - `.legal-side` — sticky sidebar `<nav>`; an `<h4>` label ("Legal") then one `<Link>` per `LEGAL_PAGES` entry (`to="/legal/$slug"`, `params={{ slug }}`). The entry matching the current slug gets `className="active"` and `aria-current="page"`.
  - `.legal-body` — `.lead` paragraph then `{page.body}`.
- All actionable elements are `<Link>` / `<button>` (no bare `<a onClick>`), consistent with the project constraint.

### Invalid slug

If `getLegalPage(slug)` returns `undefined`, the component renders an **in-shell fallback**: the `.legal-hero` + `.legal-wrap` chrome stays intact, the body shows a short "This policy could not be found." message with a `<Link>` back to `/legal/privacy`. The sidebar still renders (no active entry). This keeps global chrome and the sidebar usable rather than dropping to a bare 404.

## Content (realistic, store-consistent)

Content is professional policy prose grounded in the app's **real mechanics** — no invented guarantees, no generic boilerplate disconnected from how the store actually works. Anchors:

- **Payments:** Paystack. The store never stores card details.
- **Currency:** GHS (Ghana cedis).
- **Shipping rates (`backend/config/shipping_config.json`):** flat **GHS 25** (`flat_rate_ghs_minor: 2500`), **free over GHS 500** (`free_over_ghs_minor: 50000`). Content must reflect these exact values; if the config changes, the copy is the source of divergence to update.
- **Coverage:** the 16 Ghana regions (`src/content/regions.ts`).
- **Contact:** `STORE_INFO` (address, phone, hours) — reference it in copy where a contact block appears (do not hardcode a second copy of these strings).

Per-page substance:

- **Privacy** — data collected (account name/email, saved addresses, order history); Paystack handles payment data; session/cart cookies; how to contact re: data. Uses a `.contact-card` (or `.contact-grid`) for the contact block, sourced from `STORE_INFO`.
- **Terms** — ordering process, GHS pricing, Paystack payment, order acceptance/availability, limitation of liability, governed by the laws of Ghana.
- **Cookies** — honest: **essential session/cart cookies only**, no third-party advertising/tracking cookies. Uses `.callout.info`.
- **Shipping & delivery** — flat GHS 25, free over GHS 500, delivery windows, coverage across the 16 regions. Uses `.callout` for the delivery-timeframe note.
- **Returns & refunds** — hygiene caveat for opened cosmetics, return window, return process, Paystack refund path. Uses `.callout.warn` for the hygiene note.

## CSS

Append to `frontend/src/styles/globals.css` a verbatim port of the used blocks from `frontend/reference/legacy-css/legal.css`:

- Port: `.legal-hero`, `.legal-hero-inner`, `.legal-hero h1`/`em`/`.meta`, `.legal-wrap` (+ responsive), `.legal-side` (+ `h4`, `a`, `a:last-child`, `a:hover`, `a.active`), `.legal-body` (+ `.lead`, `h2`, `h3`, `p`, `ul`/`ol`, `li`, `strong`), `.callout` (+ `.warn`, `.info`), `.contact-grid`, `.contact-card` (+ `.k`, `.v`).
- **Do not port:** `.safety-grid`, `.safety-card`, `.patch-steps`, `.patch-step` (no pages use them).
- Dedup rule: grep `globals.css`/`pages.css` for each selector before appending; do not redefine classes that already exist. Adapted rules (if any) under a `/* adapted */` banner. Remove the stale `.legal-* (tranche 5)` placeholder comment in `account.css:4` now that the classes exist.

## Footer wiring — `frontend/src/features/shared/layouts/footer.tsx`

Replace the dead `href="#"` anchors with `<Link>`s:

- `Privacy` → `/legal/privacy`
- `Terms` → `/legal/terms`
- `Cookies` → `/legal/cookies`
- `Shipping & delivery` → `/legal/shipping`
- **Add** `Returns & refunds` → `/legal/returns` (in the same support/help column as "Shipping & delivery").

Markup structure (columns, `.footer-legal`) is otherwise unchanged.

## Testing

One vitest file `frontend/src/features/content/legal-page.test.tsx` (same setup as existing component tests):

- Known slug renders its hero title and `.lead`.
- Sidebar marks the active entry with `aria-current="page"`.
- Unknown slug renders the in-shell fallback (fallback message + back link), not a crash.

## Verification gate

Per the standing per-tranche gate: `pnpm typecheck` (0), `pnpm lint` (0 warnings), `pnpm vitest run` (all pass, incl. the new test), `pnpm build`. Class-coverage audit over touched files (no missing selectors) and no Tailwind residue. Functional smoke (nav from footer → each policy, sidebar active state, back/forward, unknown slug) deferred to human walkthrough unless dev servers are already running.

## Files

- **Create:** `src/content/legal.tsx`, `src/features/content/legal-page.tsx`, `src/features/content/legal-page.test.tsx`
- **Modify:** `src/router.tsx` (add `/legal/$slug` route), `src/features/shared/layouts/footer.tsx` (wire links), `src/styles/globals.css` (append `.legal-*` port), `src/styles/account.css` (drop stale placeholder comment)
- **Backend:** none.
