import { useGetAdminDashboard } from '../../../lib/api/generated/rueCosmeticsAPI';
import { KPICard, StatusTag, Panel } from '../../shared/ui/admin';

export function AdminDashboard() {
  const { data: dashboard, isLoading, error } = useGetAdminDashboard();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-ink-muted">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg">
        Failed to load dashboard: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  const stats = dashboard?.stats;
  const recentOrders = dashboard?.recent_orders ?? [];

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toLocaleString()}`;
  };

  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4 flex-wrap">
        <div>
          <div className="text-lavender-700 text-sm mb-1">Overview</div>
          <h1 className="font-display text-4xl font-normal">
            Good morning, <em className="font-serif italic text-lavender-700">Admin.</em>
          </h1>
        </div>
        <div className="flex gap-2">
          <button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 transition-colors">
            Last 30 days
          </button>
          <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
            Export report
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        <KPICard
          title="Revenue"
          value={stats ? formatCurrency(stats.total_revenue_ghs_minor ?? 0) : '—'}
        />
        <KPICard
          title="Orders"
          value={stats?.total_orders ?? 0}
        />
        <KPICard
          title="Avg. order"
          value={stats ? formatCurrency((stats.total_revenue_ghs_minor ?? 0) / Math.max(stats.total_orders ?? 1, 1)) : '—'}
        />
        <KPICard
          title="Customers"
          value={stats?.total_customers ?? 0}
        />
      </div>

      {/* 2-column layout */}
      <div className="grid grid-cols-[2fr_1fr] gap-5 mb-5">
        {/* Revenue Chart Placeholder */}
        <Panel
          title="Revenue, last 12 months"
          actions={
            <div className="flex gap-1">
              <span className="px-3 py-1 rounded-full text-[11px] font-semibold bg-ink text-white cursor-pointer">
                12M
              </span>
              <span className="px-3 py-1 rounded-full text-[11px] font-semibold bg-white border border-line cursor-pointer hover:bg-lavender-50">
                90D
              </span>
              <span className="px-3 py-1 rounded-full text-[11px] font-semibold bg-white border border-line cursor-pointer hover:bg-lavender-50">
                30D
              </span>
            </div>
          }
        >
          <div className="h-[220px] flex items-end gap-1 px-4">
            {/* Placeholder chart bars - will be replaced with real data */}
            {[42, 58, 49, 72, 81, 68, 94, 88, 102, 118, 96, 134].map((v, i) => (
              <div
                key={i}
                className="flex-1 bg-gradient-to-t from-lavender-600 to-lavender-400 rounded-t hover:bg-ink transition-colors min-h-[10px]"
                style={{ height: `${v * 1.5}px` }}
                title={`Month ${i + 1}: GH₵${v}k`}
              />
            ))}
          </div>
          <div className="flex justify-between text-[10px] text-ink-muted uppercase tracking-wider mt-2">
            {['May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec', 'Jan', 'Feb', 'Mar', 'Apr'].map((m) => (
              <span key={m}>{m}</span>
            ))}
          </div>
        </Panel>

        {/* Category Segmentation */}
        <Panel title="Sales by category">
          <div className="h-[200px] bg-conic-gradient rounded-full mx-auto my-5 w-[200px] relative" style={{ backgroundImage: 'conic-gradient(var(--lavender-600) 0 42%, var(--lavender-400) 42% 68%, var(--lavender-300) 68% 86%, var(--ink) 86% 100%)' }}>
            <div className="absolute inset-[30px] bg-white rounded-full" />
          </div>
          <div className="grid gap-2 mt-4">
            {[
              ['Skincare', '42%', 'var(--lavender-600)'],
              ['Haircare', '26%', 'var(--lavender-400)'],
              ['Body', '18%', 'var(--lavender-300)'],
              ['Fragrance', '14%', 'var(--ink)'],
            ].map(([label, value, color]) => (
              <div key={label} className="flex items-center justify-between gap-3 text-sm">
                <div className="flex items-center gap-2">
                  <span className="w-2.5 h-2.5 rounded-sm" style={{ background: color }} />
                  {label}
                </div>
                <strong>{value}</strong>
              </div>
            ))}
          </div>
        </Panel>
      </div>

      {/* Recent Orders */}
      <div className="grid grid-cols-[2fr_1fr] gap-5">
        <Panel
          title="Recent orders"
          actions={
            <button className="text-lavender-700 font-semibold px-2 py-1 text-sm hover:underline">
              View all →
            </button>
          }
        >
          <table className="w-full border-collapse text-[13px]">
            <thead>
              <tr className="text-left">
                <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                  Order
                </th>
                <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                  Customer
                </th>
                <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                  Total
                </th>
                <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                  Status
                </th>
              </tr>
            </thead>
            <tbody>
              {recentOrders.slice(0, 5).map((order) => (
                <tr key={order.id ?? ''} className="hover:bg-[#FAFAFA] transition-colors">
                  <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                    {(order.id ?? '').slice(0, 8).toUpperCase()}
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft">
                    {order.customer_name || order.customer_email}
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                    {formatCurrency(order.total_ghs_minor ?? 0)}
                  </td>
                  <td className="px-3 py-3 border-b border-line-soft">
                    <StatusTag status={order.status ?? 'pending'} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Panel>

        {/* Activity Feed */}
        <Panel title="Activity">
          <div className="grid gap-0">
            {[
              ['New order', 'RUE-10482 from Ama Owusu — GH₵ 485', '2m ago'],
              ['Low stock', 'Niacinamide 10% + Zinc below threshold (8 left)', '14m ago'],
              ['Review posted', '★★★★★ on Rose Hydration Serum by Delali A.', '32m ago'],
              ['Refund issued', 'RUE-10471 — GH₵ 240 to MoMo', '1h ago'],
              ['Out of stock', 'Argan Gold Hair Oil — restock placed', '2h ago'],
            ].map(([type, description, time], i) => (
              <div
                key={i}
                className="grid grid-cols-[32px_1fr_auto] gap-3 py-3 border-b border-line-soft last:border-0 items-start"
              >
                <div className="w-2 h-2 rounded-full bg-lavender-600 mt-1.5 ml-3" />
                <div className="text-[13px]">
                  <strong className="text-ink">{type}.</strong> {description}
                </div>
                <div className="text-[11px] text-ink-muted whitespace-nowrap">{time}</div>
              </div>
            ))}
          </div>
        </Panel>
      </div>
    </div>
  );
}
