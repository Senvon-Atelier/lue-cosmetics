# Admin Suite UI Alignment (Tranche 3) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Re-skin the admin suite (layout, dashboard, orders + detail, products, customers, analytics, marketing, content, settings) with the mockup's real `admin-*` CSS while keeping every generated hook, mutation, guard, and route, and stripping all fake data to honest states.

**Architecture:** The admin pages already mirror the mockup's structure through shared `Panel`/`KPICard`/`StatusTag` components — but everything is styled in dead Tailwind utilities (Tailwind is dormant), so the admin renders nearly unstyled. This tranche re-skins the three shared components to emit real `admin-*` classes (all nine pages inherit it), ports `Rue/admin.css` → `src/styles/admin.css`, converts the layout to the mockup's dark sidebar, and rewrites each page's JSX with mockup classes + honest data only.

**Tech Stack:** React 18, TanStack Router v1 (code-based, `src/router.tsx` — admin routes need NO changes; adminRoute already has standalone chrome), Orval-generated TanStack Query hooks (`useGetAdminDashboard`, `useGetAdminOrders`, `usePatchAdminOrdersIdStatus`, `useGetAdminOrdersId`, `useGetAdminProducts`, `useGetAdminCustomers`, `useGetAdminAnalyticsStats`, `useGetAdminAnalyticsRevenue`), plain CSS over custom properties, vitest.

**Spec:** `docs/superpowers/specs/2026-07-03-admin-ui-alignment-design.md` — binding. **Mockup source (read-only, NEVER modify):** `/Volumes/Georgie/reformat-audit/Downloads/casestud/Rue/` (`src/admin.jsx`, `admin.css`).

## Global Constraints

- All git commits MUST use: `git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "..."`.
- Frontend-only; nothing under `backend/` or `frontend/src/lib/api/generated/` changes; no router changes.
- Functionality freeze: every generated hook/mutation call listed in Tech Stack keeps its exact parameters and cache-invalidation behavior; the orders list keeps its working inline per-row status `<select>` mutation; pagination state logic unchanged; client-side search filters on products/customers unchanged.
- Honest data only (spec §2.3 is the binding inventory): no fake KPIs/deltas/feeds/campaigns/team/integrations/placeholder chart fallbacks; every dead control listed there is removed; unknown values render `—` or honest empty copy.
- Prices via `formatGhs`, dates via `formatOrderDate` (both `src/lib/format/utils.ts`), images via `getImageUrl` with `ph`-class fallback.
- Strip every Tailwind utility className from each file as it is touched; ported markup uses mockup class names verbatim; adapted/extrapolated classes go under a `/* adapted/extrapolated */` banner in `admin.css`.
- Dedup rule: no selector defined in two files under `src/styles/` — `alert`/`alert-warn`/`alert-info`, `kv-row`/`kv-divider` live in `account.css` (REUSE them, do not redefine); `.badge`, `.ph*` live in `globals.css`.
- Every ported button/select selector must define its own `background` and `border` (tranche-2 lesson: there is no global button reset).
- Gate after every task: `pnpm typecheck` zero errors (from `frontend/`). Full gate (typecheck/lint/vitest/build) in the final task.
- Commands run from `ruecosmetics/frontend/` unless stated.

---

### Task 1: `admin.css`

**Files:**
- Create: `frontend/src/styles/admin.css`
- Modify: `frontend/src/styles/globals.css` (add `@import './admin.css';` after the account.css import)

**Interfaces:**
- Produces: every `admin-*`, `tag-*`, `row-prod*`, `legend*` class later tasks' markup uses — exact names below.

- [ ] **Step 1: Create `src/styles/admin.css`** with exactly this content:

```css
/* Ported from Rue/admin.css — admin suite (Tranche 3, spec §5).
   Omitted on purpose: .seg-chart (hardcoded conic donut), .admin-activity/.activity-* (fake feed),
   .admin-chip / .admin-form* (no consumers after dead-control strip — port when one arrives),
   .admin-side .pill (nav count pills omitted per spec §2.2).
   Reused from other files (dedup): .alert/.alert-warn/.alert-info, .kv-row/.kv-divider (account.css),
   .badge, .ph/.ph--* (globals.css). */

/* adapted: mockup sets .admin-body on <body>; folded onto the layout root */
.admin-layout { display: grid; grid-template-columns: 240px 1fr; min-height: 100vh; background: #F5F2ED; font-family: var(--font-label); color: var(--ink); }
@media (max-width: 900px) { .admin-layout { grid-template-columns: 1fr; } }

.admin-side { background: var(--ink); color: rgba(255,255,255,0.7); padding: 28px 20px; position: sticky; top: 0; height: 100vh; overflow-y: auto; }
@media (max-width: 900px) { .admin-side { position: static; height: auto; } }
.admin-side-brand { color: white; margin-bottom: 32px; display: flex; align-items: center; gap: 10px; }
.admin-side-brand .brand-word { font-family: var(--font-serif); font-size: 22px; font-style: italic; color: white; }
.admin-side-brand .badge { background: var(--lavender-600); color: white; padding: 2px 8px; border-radius: 4px; font-size: 9px; font-weight: 700; letter-spacing: 0.14em; text-transform: uppercase; }
.admin-side h5 { font-size: 10px; letter-spacing: 0.18em; text-transform: uppercase; color: rgba(255,255,255,0.4); margin: 24px 0 8px; font-weight: 600; }
.admin-side a { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 9px 12px; border-radius: 6px; font-size: 13px; color: rgba(255,255,255,0.7); cursor: pointer; transition: all 0.2s; }
.admin-side a:hover { background: rgba(255,255,255,0.06); color: white; }
.admin-side a.active { background: var(--lavender-600); color: white; }

.admin-main { padding: 32px 40px; max-width: 1400px; }
@media (max-width: 720px) { .admin-main { padding: 20px; } }
.admin-head { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 28px; gap: 16px; flex-wrap: wrap; }
.admin-head h1 { font-family: var(--font-display); font-size: clamp(28px, 4vw, 44px); font-weight: 400; margin: 4px 0 0; letter-spacing: 0.005em; }
.admin-head .eyebrow { color: var(--lavender-700); }
.admin-head-actions { display: flex; gap: 10px; }

.admin-kpis { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; margin-bottom: 24px; }
@media (max-width: 900px) { .admin-kpis { grid-template-columns: repeat(2, 1fr); } }
.admin-kpi { background: white; border: 1px solid var(--line); border-radius: 10px; padding: 20px; }
.admin-kpi-k { font-size: 10px; letter-spacing: 0.16em; text-transform: uppercase; color: var(--ink-muted); }
.admin-kpi-v { font-family: var(--font-serif); font-size: 32px; font-weight: 400; letter-spacing: -0.01em; margin: 8px 0 4px; }
.admin-kpi-delta { font-size: 11px; font-weight: 600; }
.admin-kpi-delta.up { color: #2B7A4A; }
.admin-kpi-delta.down { color: #9E3535; }

.admin-panel { background: white; border: 1px solid var(--line); border-radius: 12px; margin-bottom: 20px; overflow: hidden; }
.admin-panel-head { display: flex; justify-content: space-between; align-items: center; padding: 18px 24px; border-bottom: 1px solid var(--line); gap: 12px; flex-wrap: wrap; }
.admin-panel-head h3 { font-family: var(--font-serif); font-size: 17px; font-weight: 500; margin: 0; letter-spacing: 0; }
.admin-panel-body { padding: 20px 24px; }

.admin-chart { height: 220px; padding: 16px 0; display: flex; align-items: flex-end; gap: 6px; }
.admin-chart-bar { flex: 1; background: linear-gradient(to top, var(--lavender-600), var(--lavender-400)); border-radius: 4px 4px 0 0; min-height: 10px; position: relative; }
.admin-chart-bar:hover { background: var(--ink); }

.admin-tbl { width: 100%; border-collapse: collapse; font-size: 13px; }
.admin-tbl th { text-align: left; padding: 10px 14px; background: #FAFAFA; border-bottom: 1px solid var(--line); font-size: 10px; letter-spacing: 0.12em; text-transform: uppercase; color: var(--ink-muted); font-weight: 700; }
.admin-tbl td { padding: 12px 14px; border-bottom: 1px solid var(--line-soft); }
.admin-tbl tr:last-child td { border-bottom: 0; }
.admin-tbl tr:hover td { background: #FAFAFA; }
.admin-tbl .num { font-variant-numeric: tabular-nums; font-weight: 600; }
.admin-tbl img, .admin-tbl .ph-sm { width: 44px; height: 44px; border-radius: 4px; }
.admin-tbl .row-prod { display: flex; align-items: center; gap: 12px; }
.admin-tbl .row-prod-name { font-family: var(--font-serif); font-size: 14px; }
.admin-tbl .row-prod-sku { font-size: 11px; color: var(--ink-muted); }

.admin-filter-bar { display: flex; gap: 8px; padding: 14px 24px; border-bottom: 1px solid var(--line); background: #FAFAFA; flex-wrap: wrap; align-items: center; }
.admin-filter-bar input, .admin-filter-bar select { padding: 8px 12px; border: 1px solid var(--line); border-radius: 6px; font-size: 12px; font-family: inherit; background: white; }
.admin-filter-bar input[type=search] { flex: 1; min-width: 200px; }

.admin-btn { padding: 8px 14px; border-radius: 6px; font-size: 12px; font-weight: 600; letter-spacing: 0.06em; cursor: pointer; border: 0; display: inline-flex; align-items: center; gap: 6px; }
.admin-btn-pri { background: var(--ink); color: white; }
.admin-btn-pri:hover { background: var(--lavender-700); }
.admin-btn-sec { background: white; color: var(--ink); border: 1px solid var(--line); }
.admin-btn-sec:hover { border-color: var(--ink); }
.admin-btn-dng { background: #FBE5E5; color: #9E3535; }
.admin-btn-link { background: transparent; color: var(--lavender-700); font-weight: 600; padding: 4px 8px; }

.admin-tag { display: inline-flex; padding: 3px 8px; border-radius: 999px; font-size: 10px; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; }
.tag-live { background: #E7F4EC; color: #2B7A4A; }
.tag-draft { background: #EEE; color: #666; }
.tag-low { background: #FFF4D6; color: #8A6500; }
.tag-oos { background: #FBE5E5; color: #9E3535; }

.admin-2col { display: grid; grid-template-columns: 2fr 1fr; gap: 20px; }
@media (max-width: 900px) { .admin-2col { grid-template-columns: 1fr; } }

.legend { display: grid; gap: 8px; margin-top: 16px; }
.legend-item { display: flex; justify-content: space-between; align-items: center; gap: 12px; font-size: 12px; }
.legend-item .dot { width: 10px; height: 10px; border-radius: 2px; }

/* adapted/extrapolated */
.admin-side-foot { margin-top: 24px; padding-top: 20px; border-top: 1px solid rgba(255,255,255,0.12); }
.admin-loading { padding: 40px; font-size: 13px; color: var(--ink-muted); }
.admin-empty { font-size: 13px; color: var(--ink-muted); margin: 0; }
.admin-pagination { display: flex; justify-content: center; align-items: center; gap: 10px; padding: 14px; font-size: 12px; color: var(--ink-muted); }
.admin-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.admin-tbl select { padding: 6px 8px; border: 1px solid var(--line); border-radius: 6px; font-size: 12px; font-family: inherit; background: white; }
.admin-chart-labels { display: flex; justify-content: space-between; font-size: 10px; color: var(--ink-muted); text-transform: uppercase; letter-spacing: 0.12em; margin-top: 8px; }
.legend-item-label { display: flex; align-items: center; gap: 8px; }
.legend-item .dot { background: var(--lavender-600); }
/* real order statuses on the mockup tag palette */
.tag-paid, .tag-delivered, .tag-shipped { background: #E7F4EC; color: #2B7A4A; }
.tag-pending, .tag-processing { background: #FFF4D6; color: #8A6500; }
.tag-failed, .tag-cancelled { background: #FBE5E5; color: #9E3535; }
.tag-fulfilled { background: var(--lavender-100); color: var(--lavender-700); }
.tag-default { background: #EEE; color: var(--ink-muted); }
```

(Note the duplicated `.legend-item .dot` selector is intentional within the same file: base geometry in the ported block, default background in the adapted block — CSS cascade merges them. If you prefer, merge into one rule in the ported block with a `/* adapted: default background */` inline note; either is acceptable.)

- [ ] **Step 2: Import it** — in `globals.css`, after `@import './account.css';` add:

```css
@import './admin.css';
```

- [ ] **Step 3: Verify**

Run: `pnpm typecheck && pnpm build`
Expected: pass; `grep -c "admin-layout" dist/assets/*.css` ≥ 1.
Dedup check (each must print exactly 1):
`for c in admin-layout admin-tbl admin-panel admin-kpi tag-live legend-item; do echo -n "$c: "; grep -l "\.$c" src/styles/*.css | wc -l; done`

- [ ] **Step 4: Commit**

```bash
git add src/styles
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): port admin.css from mockup with honest-data omissions"
```

---

### Task 2: Shared admin components + layout

**Files:**
- Rewrite: `frontend/src/features/shared/ui/admin/panel.tsx`
- Rewrite: `frontend/src/features/shared/ui/admin/kpi-card.tsx`
- Rewrite: `frontend/src/features/shared/ui/admin/status-tag.tsx`
- Rewrite: `frontend/src/features/admin/admin-layout.tsx`

**Interfaces:**
- Consumes: Task 1's classes; `useAuth()` → `{ isLoading }`; TanStack `Link`.
- Produces (all nine pages and Tasks 3–7 depend on these exact signatures):
  - `Panel({ title?: string; actions?: React.ReactNode; children })` — unchanged interface, new classes.
  - `KPICard({ title: string; value: string | number })` — **`delta`/`deltaDirection` props REMOVED** (Tasks 3–6 must not pass them; typecheck enforces).
  - `StatusTag({ status: string })` — unchanged interface, emits `admin-tag tag-*`.

- [ ] **Step 1: Rewrite `panel.tsx`**

```tsx
interface PanelProps {
  title?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

// Ported from Rue/admin.css .admin-panel structure
export function Panel({ title, actions, children }: PanelProps) {
  return (
    <div className="admin-panel">
      {(title || actions) && (
        <div className="admin-panel-head">
          {title && <h3>{title}</h3>}
          {actions && <div className="admin-head-actions">{actions}</div>}
        </div>
      )}
      <div className="admin-panel-body">{children}</div>
    </div>
  );
}
```

- [ ] **Step 2: Rewrite `kpi-card.tsx`** (delta props deleted — honest data, spec §3.2; `.admin-kpi-delta` CSS stays for the deltas backend follow-up)

```tsx
interface KPICardProps {
  title: string;
  value: string | number;
}

export function KPICard({ title, value }: KPICardProps) {
  return (
    <div className="admin-kpi">
      <div className="admin-kpi-k">{title}</div>
      <div className="admin-kpi-v">{value}</div>
    </div>
  );
}
```

- [ ] **Step 3: Rewrite `status-tag.tsx`** (same status→label logic, mockup classes; unknown statuses fall back to `tag-default`)

```tsx
interface StatusTagProps {
  status: string;
}

const KNOWN_TAGS = new Set([
  'paid', 'delivered', 'shipped', 'pending', 'processing', 'failed',
  'cancelled', 'fulfilled', 'live', 'draft', 'low', 'oos',
]);

export function StatusTag({ status }: StatusTagProps) {
  const lower = status.toLowerCase();
  const key = lower === 'low stock' ? 'low' : lower === 'out of stock' ? 'oos' : lower;
  const label =
    key === 'low' ? 'Low stock'
    : key === 'oos' ? 'Out of stock'
    : status.charAt(0).toUpperCase() + status.slice(1);
  const cls = KNOWN_TAGS.has(key) ? `tag-${key}` : 'tag-default';
  return <span className={`admin-tag ${cls}`}>{label}</span>;
}
```

- [ ] **Step 4: Rewrite `admin-layout.tsx`** (dark sidebar; navSections kept; `useNavigate` and the header removed; storefront link in the foot)

```tsx
import { Outlet, Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';

const navSections = [
  {
    title: 'Overview',
    items: [
      { label: 'Dashboard', to: '/admin' },
      { label: 'Analytics', to: '/admin/analytics' },
    ],
  },
  {
    title: 'Commerce',
    items: [
      { label: 'Orders', to: '/admin/orders' },
      { label: 'Products', to: '/admin/products' },
      { label: 'Customers', to: '/admin/customers' },
    ],
  },
  {
    title: 'Growth',
    items: [
      { label: 'Marketing', to: '/admin/marketing' },
      { label: 'Content', to: '/admin/content' },
    ],
  },
  {
    title: 'System',
    items: [{ label: 'Settings', to: '/admin/settings' }],
  },
];

export function AdminLayout() {
  const { isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="admin-layout">
        <div className="admin-loading">Loading…</div>
      </div>
    );
  }

  return (
    <div className="admin-layout">
      <aside className="admin-side">
        <div className="admin-side-brand">
          <span className="brand-word">Rue</span>
          <span className="badge">Admin</span>
        </div>
        {navSections.map((section) => (
          <div key={section.title}>
            <h5>{section.title}</h5>
            {section.items.map((item) =>
              item.to === '/admin' ? (
                <Link
                  key={item.to}
                  to={item.to}
                  activeOptions={{ exact: true }}
                  activeProps={{ className: 'active' }}
                >
                  {item.label}
                </Link>
              ) : (
                <Link key={item.to} to={item.to} activeProps={{ className: 'active' }}>
                  {item.label}
                </Link>
              ),
            )}
          </div>
        ))}
        <div className="admin-side-foot">
          <Link to="/">← Storefront</Link>
        </div>
      </aside>
      <main className="admin-main">
        <Outlet />
      </main>
    </div>
  );
}
```

Note: `.admin-side-brand .badge` scoped rules override globals' `.badge` — verify visually that the badge renders dark-sidebar style. The layout owns `<main className="admin-main">` (mockup-faithful; admin pages are fragments — verify none renders its own `<main>`: `grep -rn "<main" src/features/admin/` should only hit admin-layout.tsx after this tranche).

- [ ] **Step 5: Typecheck**

Run: `pnpm typecheck`
Expected: errors ONLY in files passing `delta`/`deltaDirection` to KPICard (`dashboard/index.tsx`, `analytics/index.tsx`). Those pages are rewritten in Tasks 3 and 6 — to keep this task's gate green, remove the `delta`/`deltaDirection` props from those call sites now (minimal edit, only deleting the two props per call; page rewrites come later). Re-run until zero errors.

- [ ] **Step 6: Commit**

```bash
git add src/features/shared/ui/admin src/features/admin/admin-layout.tsx src/features/admin/dashboard/index.tsx src/features/admin/analytics/index.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): admin shell + shared Panel/KPICard/StatusTag on mockup classes, dark sidebar, honest KPI props"
```

---

### Task 3: Dashboard

**Files:**
- Rewrite: `frontend/src/features/admin/dashboard/index.tsx`

**Interfaces:**
- Consumes: `useGetAdminDashboard()` → `{ stats?: { total_revenue_ghs_minor, total_orders, total_customers, total_products, pending_orders, paid_orders, shipped_orders, delivered_orders }, recent_orders?: [{ id, customer_name, customer_email, total_ghs_minor, status }] }`; `KPICard({ title, value })`; `StatusTag`; `Panel`; `formatGhs`; `useAuth()` → `{ user }`.

- [ ] **Step 1: Rewrite `dashboard/index.tsx`** with exactly:

```tsx
import { useGetAdminDashboard } from '../../../lib/api/generated/rueCosmeticsAPI';
import { useAuth } from '../../../lib/auth/auth-provider';
import { formatGhs } from '../../../lib/format/utils';
import { KPICard, StatusTag, Panel } from '../../shared/ui/admin';

export function AdminDashboard() {
  const { user } = useAuth();
  const { data: dashboard, isLoading, error } = useGetAdminDashboard();

  if (isLoading) {
    return <div className="admin-loading">Loading dashboard…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load dashboard: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  const stats = dashboard?.stats;
  const recentOrders = dashboard?.recent_orders ?? [];
  const firstName = user?.name?.trim().split(/\s+/)[0];

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Overview</div>
          <h1>Good morning{firstName ? `, ${firstName}` : ''}.</h1>
        </div>
      </div>

      <div className="admin-kpis">
        <KPICard
          title="Revenue"
          value={stats ? formatGhs(stats.total_revenue_ghs_minor ?? 0) : '—'}
        />
        <KPICard title="Orders" value={stats?.total_orders ?? '—'} />
        <KPICard
          title="Avg. order"
          value={
            stats && (stats.total_orders ?? 0) > 0
              ? formatGhs(
                  Math.round(
                    (stats.total_revenue_ghs_minor ?? 0) / (stats.total_orders ?? 1),
                  ),
                )
              : '—'
          }
        />
        <KPICard title="Customers" value={stats?.total_customers ?? '—'} />
      </div>

      <div className="admin-2col">
        <Panel title="Recent orders">
          {recentOrders.length === 0 ? (
            <p className="admin-empty">No orders yet.</p>
          ) : (
            <table className="admin-tbl">
              <thead>
                <tr>
                  <th>Order</th>
                  <th>Customer</th>
                  <th>Total</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {recentOrders.slice(0, 5).map((order) => (
                  <tr key={order.id ?? ''}>
                    <td className="num">{(order.id ?? '').slice(0, 8).toUpperCase()}</td>
                    <td>{order.customer_name || order.customer_email}</td>
                    <td className="num">{formatGhs(order.total_ghs_minor ?? 0)}</td>
                    <td>
                      <StatusTag status={order.status ?? 'pending'} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </Panel>

        <Panel title="Orders by status">
          {stats ? (
            <table className="admin-tbl">
              <tbody>
                <tr>
                  <td>Pending</td>
                  <td className="num">{stats.pending_orders ?? 0}</td>
                </tr>
                <tr>
                  <td>Paid</td>
                  <td className="num">{stats.paid_orders ?? 0}</td>
                </tr>
                <tr>
                  <td>Shipped</td>
                  <td className="num">{stats.shipped_orders ?? 0}</td>
                </tr>
                <tr>
                  <td>Delivered</td>
                  <td className="num">{stats.delivered_orders ?? 0}</td>
                </tr>
                <tr>
                  <td>Products in catalog</td>
                  <td className="num">{stats.total_products ?? 0}</td>
                </tr>
              </tbody>
            </table>
          ) : (
            <p className="admin-empty">No stats available.</p>
          )}
        </Panel>
      </div>
    </>
  );
}
```

(Deleted: fake deltas, "Last 30 days"/"Export report" buttons, placeholder revenue chart, fake sales-by-category donut, fake activity feed, local `formatCurrency`.)

- [ ] **Step 2: Typecheck** — `pnpm typecheck`, zero errors.

- [ ] **Step 3: Commit**

```bash
git add src/features/admin/dashboard/index.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): admin dashboard on mockup classes — real KPIs and status breakdown, fake panels removed"
```

---

### Task 4: Orders list + order detail

**Files:**
- Rewrite: `frontend/src/features/admin/orders/index.tsx`
- Rewrite: `frontend/src/features/admin/orders/order-detail.tsx`

**Interfaces:**
- Consumes: `useGetAdminOrders({ page, page_size, status? })` → `{ orders?, total_pages? }`; `usePatchAdminOrdersIdStatus` with the existing onSuccess invalidations; `useGetAdminOrdersId(id, { query: { enabled: !!id } })` → `{ order?, items? }` (items carry `product_image_snapshot`/`product_name_snapshot`/`product_brand_snapshot`/`qty`/`unit_price_ghs_minor`); `formatGhs`, `formatOrderDate`, `getImageUrl`; `Panel`, `StatusTag`.

- [ ] **Step 1: Rewrite `orders/index.tsx`** — ALL hook/mutation/state logic preserved verbatim (page, statusFilter, page reset, invalidations, the six status options in the row select); UI rewritten:

```tsx
import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import {
  useGetAdminOrders,
  getGetAdminOrdersQueryKey,
  usePatchAdminOrdersIdStatus,
  getGetAdminDashboardQueryKey,
} from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate } from '../../../lib/format/utils';
import { Panel } from '../../shared/ui/admin';

export function AdminOrders() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [statusFilter, setStatusFilter] = useState('');

  const { data: ordersData, isLoading, error } = useGetAdminOrders({
    page: page + 1,
    page_size: 20,
    ...(statusFilter ? { status: statusFilter } : {}),
  });

  const updateStatusMutation = usePatchAdminOrdersIdStatus({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetAdminOrdersQueryKey() });
        queryClient.invalidateQueries({ queryKey: getGetAdminDashboardQueryKey() });
      },
    },
  });

  const totalPages = ordersData?.total_pages ?? 1;

  if (isLoading) {
    return <div className="admin-loading">Loading orders…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load orders: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  const orders = ordersData?.orders ?? [];

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Fulfilment</div>
          <h1>Orders</h1>
        </div>
      </div>

      <Panel>
        <div className="admin-filter-bar">
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              setPage(0);
            }}
          >
            <option value="">All statuses</option>
            <option value="pending">Pending</option>
            <option value="paid">Paid</option>
            <option value="fulfilled">Fulfilled</option>
            <option value="shipped">Shipped</option>
            <option value="delivered">Delivered</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>

        {orders.length === 0 ? (
          <div className="admin-panel-body">
            <p className="admin-empty">No orders{statusFilter ? ' with this status' : ''} yet.</p>
          </div>
        ) : (
          <table className="admin-tbl">
            <thead>
              <tr>
                <th>Order</th>
                <th>Customer</th>
                <th>Date</th>
                <th>Total</th>
                <th>Status</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id ?? ''}>
                  <td className="num">{(order.id ?? '').slice(0, 8).toUpperCase()}</td>
                  <td>{order.customer_name || order.customer_email}</td>
                  <td>{formatOrderDate(order.created_at)}</td>
                  <td className="num">{formatGhs(order.total_ghs_minor ?? 0)}</td>
                  <td>
                    <select
                      value={order.status ?? ''}
                      onChange={(e) =>
                        updateStatusMutation.mutate({
                          id: order.id ?? '',
                          data: { status: e.target.value },
                        })
                      }
                      disabled={updateStatusMutation.isPending}
                    >
                      <option value="pending">Pending</option>
                      <option value="paid">Paid</option>
                      <option value="fulfilled">Fulfilled</option>
                      <option value="shipped">Shipped</option>
                      <option value="delivered">Delivered</option>
                      <option value="cancelled">Cancelled</option>
                    </select>
                  </td>
                  <td>
                    <button
                      className="admin-btn admin-btn-link"
                      onClick={() => navigate({ to: `/admin/orders/${order.id}` })}
                    >
                      Open
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div className="admin-pagination">
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
            >
              Previous
            </button>
            <span>
              Page {page + 1} of {totalPages}
            </span>
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
            >
              Next
            </button>
          </div>
        )}
      </Panel>
    </>
  );
}
```

(Deleted: fake page-scoped KPI row, dead search input, dead row checkboxes, fake Export/Print-labels buttons, local `formatCurrency`/`formatDate`. Added honest empty state — the old page rendered an empty table.)

- [ ] **Step 2: Rewrite `orders/order-detail.tsx`** — same hook, mockup classes; fake Print-invoice/Update-status buttons removed; item images added from the real `product_image_snapshot` field:

```tsx
import { useParams, Link } from '@tanstack/react-router';
import { useGetAdminOrdersId } from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate, getImageUrl } from '../../../lib/format/utils';
import { StatusTag, Panel } from '../../shared/ui/admin';

export function AdminOrderDetail() {
  const { id } = useParams({ from: '/admin/orders/$id' });
  const { data: orderDetail, isLoading, error } = useGetAdminOrdersId(id, {
    query: { enabled: !!id },
  });

  if (isLoading) {
    return <div className="admin-loading">Loading order details…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load order: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  const order = orderDetail?.order;
  const items = orderDetail?.items ?? [];

  if (!order) {
    return <p className="admin-empty">Order not found.</p>;
  }

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Order details</div>
          <h1>#{(order.id ?? '').slice(0, 8).toUpperCase()}</h1>
        </div>
        <div className="admin-head-actions">
          <StatusTag status={order.status ?? 'pending'} />
          <Link className="admin-btn admin-btn-sec" to="/admin/orders">
            All orders
          </Link>
        </div>
      </div>

      <div className="admin-2col">
        <Panel title="Items">
          <table className="admin-tbl">
            <thead>
              <tr>
                <th>Product</th>
                <th>Qty</th>
                <th>Price</th>
                <th>Total</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id}>
                  <td>
                    <div className="row-prod">
                      {item.product_image_snapshot ? (
                        <img
                          src={getImageUrl(item.product_image_snapshot)}
                          alt={item.product_name_snapshot ?? ''}
                          loading="lazy"
                        />
                      ) : (
                        <div className="ph ph--lavender ph-sm" />
                      )}
                      <div>
                        <div className="row-prod-name">{item.product_name_snapshot}</div>
                        <div className="row-prod-sku">{item.product_brand_snapshot}</div>
                      </div>
                    </div>
                  </td>
                  <td className="num">{item.qty ?? 0}</td>
                  <td className="num">{formatGhs(item.unit_price_ghs_minor ?? 0)}</td>
                  <td className="num">
                    {formatGhs((item.unit_price_ghs_minor ?? 0) * (item.qty ?? 0))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Panel>

        <div>
          <Panel title="Summary">
            <div className="kv-row">
              <span>Status</span>
              <StatusTag status={order.status ?? 'pending'} />
            </div>
            <div className="kv-row">
              <span>Created</span>
              <span>{formatOrderDate(order.created_at)}</span>
            </div>
            {order.updated_at !== order.created_at && (
              <div className="kv-row">
                <span>Updated</span>
                <span>{formatOrderDate(order.updated_at)}</span>
              </div>
            )}
            <div className="kv-divider"></div>
            <div className="kv-row">
              <span>Subtotal</span>
              <span>{formatGhs(order.subtotal_ghs_minor ?? 0)}</span>
            </div>
            <div className="kv-row">
              <span>Shipping</span>
              <span>{formatGhs(order.shipping_ghs_minor ?? 0)}</span>
            </div>
            <div className="kv-row">
              <strong>Total</strong>
              <strong>{formatGhs(order.total_ghs_minor ?? 0)}</strong>
            </div>
            {order.paystack_reference && (
              <>
                <div className="kv-divider"></div>
                <div className="kv-row">
                  <span>Payment</span>
                  <span>Paystack · {order.paystack_reference}</span>
                </div>
              </>
            )}
          </Panel>

          {order.shipping_address && (
            <Panel title="Shipping address">
              <div className="kv-row">
                <span>{order.shipping_address.label || 'Address'}</span>
                <span></span>
              </div>
              <p className="admin-empty" style={{ color: 'var(--ink)' }}>
                {order.shipping_address.line1}
                {order.shipping_address.line2 && (
                  <>
                    <br />
                    {order.shipping_address.line2}
                  </>
                )}
                <br />
                {order.shipping_address.city}, {order.shipping_address.region}
                <br />
                {order.shipping_address.phone}
              </p>
            </Panel>
          )}
        </div>
      </div>
    </>
  );
}
```

- [ ] **Step 3: Typecheck** — `pnpm typecheck`, zero errors.

- [ ] **Step 4: Commit**

```bash
git add src/features/admin/orders
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): admin orders list + detail on mockup classes — live controls only, honest states"
```

---

### Task 5: Products + Customers

**Files:**
- Rewrite: `frontend/src/features/admin/products/index.tsx`
- Rewrite: `frontend/src/features/admin/customers/index.tsx`

**Interfaces:**
- Consumes: `useGetAdminProducts({ page, page_size })` → `{ products?, total?, total_pages? }` (product: `id, name, slug, image_path, category_id, brand_id, price_ghs_minor`); `useGetAdminCustomers({ page, page_size })` → `{ customers?, total_pages? }` (customer: `id, name, email, order_count, lifetime_value_ghs_minor, created_at`); `KPICard`, `Panel`; `formatGhs`, `formatOrderDate`, `getImageUrl`.

- [ ] **Step 1: Rewrite `products/index.tsx`** — keep hook, page state, and the client-side search filter; strip dead category/status selects (their state never filtered anything), fake KPI row, placeholder `getStockStatus`, stock/status columns, dead Edit button, fake header buttons:

```tsx
import { useState } from 'react';
import { useGetAdminProducts } from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, getImageUrl } from '../../../lib/format/utils';
import { KPICard, Panel } from '../../shared/ui/admin';

export function AdminProducts() {
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: productsData, isLoading, error } = useGetAdminProducts({
    page: page + 1,
    page_size: 20,
  });

  const totalPages = productsData?.total_pages ?? 1;
  const products = productsData?.products ?? [];

  // Client-side search over the fetched page (pre-existing; server-side search is a backlog item)
  const filteredProducts = products.filter((p) => {
    if (
      searchQuery &&
      !(p.name ?? '').toLowerCase().includes(searchQuery.toLowerCase()) &&
      !(p.slug ?? '').toLowerCase().includes(searchQuery.toLowerCase())
    ) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return <div className="admin-loading">Loading products…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load products: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Catalog</div>
          <h1>Products</h1>
        </div>
      </div>

      <div className="admin-kpis">
        <KPICard title="Total SKUs" value={productsData?.total ?? '—'} />
      </div>

      <Panel>
        <div className="admin-filter-bar">
          <input
            type="search"
            placeholder="Search name or slug…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {filteredProducts.length === 0 ? (
          <div className="admin-panel-body">
            <p className="admin-empty">No products match.</p>
          </div>
        ) : (
          <table className="admin-tbl">
            <thead>
              <tr>
                <th>Product</th>
                <th>Category</th>
                <th>Brand</th>
                <th>Price</th>
              </tr>
            </thead>
            <tbody>
              {filteredProducts.map((product) => (
                <tr key={product.id}>
                  <td>
                    <div className="row-prod">
                      {product.image_path ? (
                        <img
                          src={getImageUrl(product.image_path)}
                          alt={product.name}
                          loading="lazy"
                        />
                      ) : (
                        <div className="ph ph--lavender ph-sm" />
                      )}
                      <div>
                        <div className="row-prod-name">{product.name}</div>
                        <div className="row-prod-sku">{product.slug}</div>
                      </div>
                    </div>
                  </td>
                  <td>{(product.category_id ?? '—').slice(0, 8)}</td>
                  <td>{(product.brand_id ?? '—').slice(0, 8)}</td>
                  <td className="num">{formatGhs(product.price_ghs_minor ?? 0)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div className="admin-pagination">
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
            >
              Previous
            </button>
            <span>
              Page {page + 1} of {totalPages}
            </span>
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
            >
              Next
            </button>
          </div>
        )}
      </Panel>
    </>
  );
}
```

(Note: the old page passed raw `product.image_path` to `<img>` — this rewrite routes it through `getImageUrl` per the global constraint.)

- [ ] **Step 2: Rewrite `customers/index.tsx`** — keep hook, page state, live client-side search; strip fake KPI row, invented tier column + tier select, fake header buttons:

```tsx
import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useGetAdminCustomers } from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate } from '../../../lib/format/utils';
import { Panel } from '../../shared/ui/admin';

export function AdminCustomers() {
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: customersData, isLoading, error } = useGetAdminCustomers({
    page: page + 1,
    page_size: 20,
  });

  const totalPages = customersData?.total_pages ?? 1;
  const customers = customersData?.customers ?? [];

  // Client-side search over the fetched page (pre-existing; server-side search is a backlog item)
  const filteredCustomers = customers.filter((c) => {
    if (
      searchQuery &&
      !c.name?.toLowerCase().includes(searchQuery.toLowerCase()) &&
      !(c.email ?? '').toLowerCase().includes(searchQuery.toLowerCase())
    ) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return <div className="admin-loading">Loading customers…</div>;
  }

  if (error) {
    return (
      <div className="alert alert-warn">
        Failed to load customers: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">CRM</div>
          <h1>Customers</h1>
        </div>
      </div>

      <Panel>
        <div className="admin-filter-bar">
          <input
            type="search"
            placeholder="Search name or email…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {filteredCustomers.length === 0 ? (
          <div className="admin-panel-body">
            <p className="admin-empty">No customers match.</p>
          </div>
        ) : (
          <table className="admin-tbl">
            <thead>
              <tr>
                <th>Customer</th>
                <th>Email</th>
                <th>Orders</th>
                <th>Lifetime</th>
                <th>Member since</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {filteredCustomers.map((customer) => (
                <tr key={customer.id}>
                  <td className="num">{customer.name || 'No name'}</td>
                  <td>{customer.email}</td>
                  <td className="num">{customer.order_count}</td>
                  <td className="num">{formatGhs(customer.lifetime_value_ghs_minor ?? 0)}</td>
                  <td>{formatOrderDate(customer.created_at)}</td>
                  <td>
                    <button
                      className="admin-btn admin-btn-link"
                      onClick={() => navigate({ to: `/admin/customers/${customer.id}` })}
                    >
                      View
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div className="admin-pagination">
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
            >
              Previous
            </button>
            <span>
              Page {page + 1} of {totalPages}
            </span>
            <button
              className="admin-btn admin-btn-sec"
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
            >
              Next
            </button>
          </div>
        )}
      </Panel>
    </>
  );
}
```

(The `/admin/customers/$id` route exists as an honest Coming-Soon stub — View stays.)

- [ ] **Step 3: Typecheck** — `pnpm typecheck`, zero errors.

- [ ] **Step 4: Commit**

```bash
git add src/features/admin/products src/features/admin/customers
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): admin products + customers on mockup classes — live search only, invented tiers and fake KPIs removed"
```

---

### Task 6: Analytics

**Files:**
- Rewrite: `frontend/src/features/admin/analytics/index.tsx`

**Interfaces:**
- Consumes: `useGetAdminAnalyticsStats()` → `{ customer_stats?: { total_customers, customers_with_orders, active_customers_30d }, top_products?: [{ id, name, total_sold, revenue_ghs_minor }] }`; `useGetAdminAnalyticsRevenue({ granularity, date_from, date_to })` → `{ by_date?: [{ date, revenue_ghs_minor }], by_category?: [{ category_id, category_name, revenue_ghs_minor }], order_stats?: { total_completed_revenue_ghs_minor } }`; `KPICard`, `Panel`; `formatGhs`.

- [ ] **Step 1: Rewrite `analytics/index.tsx`** — same two hooks with the same (hardcoded-2024, pre-existing) parameters; ALL placeholder fallbacks and fake KPIs deleted; real chart/legend/tables only:

```tsx
import {
  useGetAdminAnalyticsStats,
  useGetAdminAnalyticsRevenue,
} from '../../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs } from '../../../lib/format/utils';
import { KPICard, Panel } from '../../shared/ui/admin';

export function AdminAnalytics() {
  const granularity = 'month'; // switcher UI not built yet; keep a constant, not dead state

  const { data: stats, isLoading: statsLoading } = useGetAdminAnalyticsStats();
  const { data: revenueData, isLoading: revenueLoading } = useGetAdminAnalyticsRevenue({
    granularity,
    date_from: '2024-01-01T00:00:00Z',
    date_to: '2024-12-31T23:59:59Z',
  });

  if (statsLoading || revenueLoading) {
    return <div className="admin-loading">Loading analytics…</div>;
  }

  const customerStats = stats?.customer_stats;
  const topProducts = stats?.top_products ?? [];
  const byDate = revenueData?.by_date ?? [];
  const byCategory = revenueData?.by_category ?? [];
  const maxRevenue = Math.max(...byDate.map((d) => d.revenue_ghs_minor ?? 0), 1);

  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Insights</div>
          <h1>Analytics</h1>
        </div>
      </div>

      <div className="admin-kpis">
        <KPICard title="Customers" value={customerStats?.total_customers ?? '—'} />
        <KPICard title="With orders" value={customerStats?.customers_with_orders ?? '—'} />
        <KPICard title="Active (30d)" value={customerStats?.active_customers_30d ?? '—'} />
        <KPICard
          title="Completed revenue"
          value={
            revenueData?.order_stats
              ? formatGhs(revenueData.order_stats.total_completed_revenue_ghs_minor ?? 0)
              : '—'
          }
        />
      </div>

      <Panel title={`Revenue by ${granularity} (2024)`}>
        {byDate.length > 0 ? (
          <>
            <div className="admin-chart">
              {byDate.map((item, i) => (
                <div
                  key={item.date ?? i}
                  className="admin-chart-bar"
                  style={{
                    height: `${Math.max(5, ((item.revenue_ghs_minor ?? 0) / maxRevenue) * 100)}%`,
                  }}
                  title={`${item.date}: ${formatGhs(item.revenue_ghs_minor ?? 0)}`}
                />
              ))}
            </div>
            <div className="admin-chart-labels">
              {byDate.map((item, i) => (
                <span key={item.date ?? i}>{(item.date ?? '').slice(5, 7)}</span>
              ))}
            </div>
          </>
        ) : (
          <p className="admin-empty">No revenue data for this range yet.</p>
        )}
      </Panel>

      <div className="admin-2col">
        <Panel title="Top products">
          {topProducts.length > 0 ? (
            <table className="admin-tbl">
              <thead>
                <tr>
                  <th>Product</th>
                  <th>Units</th>
                  <th>Revenue</th>
                </tr>
              </thead>
              <tbody>
                {topProducts.map((product) => (
                  <tr key={product.id ?? product.name}>
                    <td className="row-prod-name">{product.name ?? 'Product'}</td>
                    <td className="num">{product.total_sold ?? 0}</td>
                    <td className="num">{formatGhs(product.revenue_ghs_minor ?? 0)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="admin-empty">No sales data yet.</p>
          )}
        </Panel>

        <Panel title="Revenue by category">
          {byCategory.length > 0 ? (
            <div className="legend">
              {byCategory.map((item) => (
                <div key={item.category_id ?? item.category_name} className="legend-item">
                  <div className="legend-item-label">
                    <span className="dot" />
                    {item.category_name}
                  </div>
                  <strong>{formatGhs(item.revenue_ghs_minor ?? 0)}</strong>
                </div>
              ))}
            </div>
          ) : (
            <p className="admin-empty">No category data yet.</p>
          )}
        </Panel>
      </div>
    </>
  );
}
```

(Deleted: fake Sessions/Conversion/Add-to-cart/Checkout-rate KPIs, placeholder chart bars, fake month labels, fake top-products rows, fake traffic sources. The chart's bars, labels, and tooltips all derive from `by_date`.)

- [ ] **Step 2: Typecheck** — `pnpm typecheck`, zero errors.

- [ ] **Step 3: Commit**

```bash
git add src/features/admin/analytics/index.tsx
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): admin analytics on mockup classes — real chart/legend/KPIs, all placeholder fallbacks removed"
```

---

### Task 7: Marketing + Content + Settings (honest stubs)

**Files:**
- Rewrite: `frontend/src/features/admin/marketing/index.tsx`
- Rewrite: `frontend/src/features/admin/content/index.tsx`
- Rewrite: `frontend/src/features/admin/settings/index.tsx`

**Interfaces:**
- Consumes: `Panel`; `kv-row` class (account.css).

- [ ] **Step 1: Rewrite `marketing/index.tsx`**

```tsx
import { Panel } from '../../shared/ui/admin';

// Honest stub: campaigns, discount codes, and segments have no backend yet (spec §2.3, §8.4).

export function AdminMarketing() {
  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Growth</div>
          <h1>Marketing</h1>
        </div>
      </div>
      <Panel title="Not wired up yet">
        <p className="admin-empty">
          Campaigns, discount codes, and customer segments need backend support before this
          page can show real data. See the backend follow-ups in the tranche-3 spec.
        </p>
      </Panel>
    </>
  );
}
```

- [ ] **Step 2: Rewrite `content/index.tsx`**

```tsx
import { Panel } from '../../shared/ui/admin';

// Honest stub: there is no CMS backend. Site copy ships as static files under src/content/.

export function AdminContent() {
  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">CMS</div>
          <h1>Content</h1>
        </div>
      </div>
      <Panel title="Not wired up yet">
        <p className="admin-empty">
          Journal posts, homepage blocks, and page management need a CMS backend. Today the
          site's editorial copy is maintained as static files in the frontend repo
          (src/content/). See the backend follow-ups in the tranche-3 spec.
        </p>
      </Panel>
    </>
  );
}
```

- [ ] **Step 3: Rewrite `settings/index.tsx`**

```tsx
import { Panel } from '../../shared/ui/admin';

// Read-only true facts; editable settings need backend support (spec §2.3, §8.4).

export function AdminSettings() {
  return (
    <>
      <div className="admin-head">
        <div>
          <div className="eyebrow">Configuration</div>
          <h1>Settings</h1>
        </div>
      </div>
      <Panel title="Store">
        <div className="kv-row">
          <span>Store</span>
          <span>Rue Cosmetics</span>
        </div>
        <div className="kv-row">
          <span>Currency</span>
          <span>GHS — Ghanaian cedi</span>
        </div>
        <div className="kv-row">
          <span>Payments</span>
          <span>Paystack</span>
        </div>
      </Panel>
      <Panel title="Editable settings">
        <p className="admin-empty">
          Store configuration lives in the backend deployment (env + config files); editable
          settings need backend support. See the backend follow-ups in the tranche-3 spec.
        </p>
      </Panel>
    </>
  );
}
```

- [ ] **Step 4: Typecheck** — `pnpm typecheck`, zero errors.

- [ ] **Step 5: Commit**

```bash
git add src/features/admin/marketing src/features/admin/content src/features/admin/settings
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "feat(frontend): marketing/content/settings honest stubs — fake campaigns, CMS tables, team and integrations removed"
```

---

### Task 8: Final gate, audits, ledger

**Files:**
- Modify (only if audits find gaps): `frontend/src/styles/admin.css`
- The controller maintains `.superpowers/sdd/progress.md` — do not edit it.

- [ ] **Step 1: Audits** (from `frontend/`)

```bash
# 1. Tailwind residue in admin + shared admin files (expect no REAL hits; class names
#    containing hyphenated 'grid'/'flex' like admin-... are false positives to ignore):
grep -nE 'className="[^"]*\b(flex|grid|px-|py-|mb-|mt-|text-|bg-|rounded|w-full|space-y|gap-|h-\[|min-h)' \
  src/features/admin/**/*.tsx src/features/shared/ui/admin/*.tsx

# 2. Class coverage — every static className resolves to a stylesheet definition:
for f in src/features/admin/*.tsx src/features/admin/*/*.tsx src/features/shared/ui/admin/*.tsx; do
  grep -o 'className="[^"]*"' "$f" | sed 's/className="//;s/"$//' | tr ' ' '\n' | sort -u | while read -r c; do
    [ -z "$c" ] && continue
    grep -qs -- "\.$c" src/styles/*.css || echo "$f: missing .$c"
  done
done

# 3. Template-literal classes (StatusTag) — verify each tag class is defined:
for t in tag-live tag-draft tag-low tag-oos tag-paid tag-pending tag-failed tag-cancelled tag-fulfilled tag-default tag-delivered tag-shipped tag-processing; do
  grep -qs -- "\.$t" src/styles/admin.css || echo "missing .$t"
done

# 4. No page-owned <main> in admin (layout owns it):
grep -rn "<main" src/features/admin/ | grep -v admin-layout.tsx
```

Fix any real gap by adding the class to `admin.css` under the adapted banner; note it in your report.

- [ ] **Step 2: Full gate**

Run: `pnpm typecheck && pnpm lint && pnpm vitest run && pnpm build`
Expected: all green, lint zero warnings. Fix minimally anything arising from files this tranche touched; report every fix.

- [ ] **Step 3: Route smoke** — only if a dev server is already running; otherwise note "deferred to human walkthrough". Logged-out `/admin` must still redirect (guard untouched).

- [ ] **Step 4: Commit** (only if the audits/gate changed files)

```bash
git add src/styles/admin.css src/features
git -c user.email='52512684+oti-adjei@users.noreply.github.com' commit -m "chore(frontend): tranche-3 final gate — class audit fixes"
```
