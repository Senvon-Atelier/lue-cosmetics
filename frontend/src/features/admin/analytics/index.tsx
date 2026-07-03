import {
  useGetAdminAnalyticsStats,
  useGetAdminAnalyticsRevenue,
} from '../../../lib/api/generated/rueCosmeticsAPI';
import { KPICard, Panel } from '../../shared/ui/admin';

export function AdminAnalytics() {
  const granularity = 'month'; // switcher UI not built yet; keep a constant, not dead state

  const { data: stats, isLoading: statsLoading } = useGetAdminAnalyticsStats();
  const { data: revenueData, isLoading: revenueLoading } = useGetAdminAnalyticsRevenue({
    granularity,
    date_from: '2024-01-01T00:00:00Z',
    date_to: '2024-12-31T23:59:59Z',
  });

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toLocaleString()}`;
  };

  const topProducts = stats?.top_products ?? [];
  const revenueByDate = revenueData?.by_date ?? [];
  const revenueByCategory = revenueData?.by_category ?? [];

  if (statsLoading || revenueLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-ink-muted">Loading analytics...</div>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-7">
        <div className="text-lavender-700 text-sm mb-1">Insights</div>
        <h1 className="font-display text-4xl font-normal">Analytics</h1>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        <KPICard
          title="Sessions (30d)"
          value="184,200"
        />
        <KPICard
          title="Conversion"
          value="3.8%"
        />
        <KPICard
          title="Add-to-cart"
          value="12.4%"
        />
        <KPICard
          title="Checkout rate"
          value="62.1%"
        />
      </div>

      {/* Revenue Chart */}
      <Panel title={`Revenue, last 12 ${granularity}s`}>
        <div className="h-[220px] flex items-end gap-1 px-4">
          {revenueByDate.length > 0 ? (
            revenueByDate.map((item, i) => (
              <div
                key={i}
                className="flex-1 bg-gradient-to-t from-lavender-600 to-lavender-400 rounded-t hover:bg-ink transition-colors min-h-[10px]"
                style={{ height: `${Math.min(100, ((item.revenue_ghs_minor ?? 0) / 1000000) * 100)}%` }}
                title={`${item.date}: GH₵${((item.revenue_ghs_minor ?? 0) / 100).toLocaleString()}`}
              />
            ))
          ) : (
            // Placeholder chart if no data
            [42, 58, 49, 72, 81, 68, 94, 88, 102, 118, 96, 134].map((v, i) => (
              <div
                key={i}
                className="flex-1 bg-gradient-to-t from-lavender-600 to-lavender-400 rounded-t hover:bg-ink transition-colors min-h-[10px]"
                style={{ height: `${v * 1.5}px` }}
                title={`Month ${i + 1}: GH₵${v}k`}
              />
            ))
          )}
        </div>
        <div className="flex justify-between text-[10px] text-ink-muted uppercase tracking-wider mt-2">
          {['May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec', 'Jan', 'Feb', 'Mar', 'Apr'].map((m) => (
            <span key={m}>{m}</span>
          ))}
        </div>
      </Panel>

      {/* 2-column layout */}
      <div className="grid grid-cols-[2fr_1fr] gap-5 mt-5">
        {/* Top Products */}
        <Panel title="Top products">
          <table className="w-full">
            <thead>
              <tr className="text-left border-b border-line">
                <th className="py-2 text-sm font-semibold">Product</th>
                <th className="py-2 text-sm font-semibold text-right">Units</th>
                <th className="py-2 text-sm font-semibold text-right">Revenue</th>
              </tr>
            </thead>
            <tbody>
              {topProducts.length > 0 ? (
                topProducts.map((product) => (
                  <tr key={product.id ?? ''} className="border-b border-line-soft">
                    <td className="py-2 font-display">{product.name ?? 'Product'}</td>
                    <td className="py-2 text-right font-variant-numeric tabular-nums">{product.total_sold ?? 0}</td>
                    <td className="py-2 text-right font-variant-numeric tabular-nums font-semibold">
                      {formatCurrency(product.revenue_ghs_minor ?? 0)}
                    </td>
                  </tr>
                ))
              ) : (
                // Placeholder data if none available
                ([
                  ['Rose Hydration Serum', 412, 101940],
                  ['Argan Gold Hair Oil', 389, 36955],
                  ['Cocoa Body Lotion', 318, 20670],
                  ['Niacinamide 10%', 284, 22152],
                  ['Nuit de Prelude', 86, 58480],
                ] as [string, number, number][]).map(([name, units, revenue]) => (
                  <tr key={name} className="border-b border-line-soft">
                    <td className="py-2 font-display">{name}</td>
                    <td className="py-2 text-right font-variant-numeric tabular-nums">{units}</td>
                    <td className="py-2 text-right font-variant-numeric tabular-nums font-semibold">
                      {formatCurrency(revenue)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </Panel>

        {/* Traffic Sources */}
        <Panel title="Traffic sources">
          <div className="grid gap-2">
            {revenueByCategory.length > 0 ? (
              revenueByCategory.map((item) => (
                <div key={item.category_name ?? ''} className="flex items-center justify-between gap-3 text-sm">
                  <div className="flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-sm bg-ink" />
                    {item.category_name}
                  </div>
                  <strong>{formatCurrency(item.revenue_ghs_minor ?? 0)}</strong>
                </div>
              ))
            ) : (
              // Placeholder data
              [
                ['Organic search', '38%', 'var(--ink)'],
                ['Instagram', '24%', 'var(--lavender-600)'],
                ['Direct', '18%', 'var(--lavender-400)'],
                ['Affiliate', '12%', 'var(--lavender-300)'],
                ['Email', '8%', 'var(--lavender-200)'],
              ].map(([source, percent, color]) => (
                <div key={source} className="flex items-center justify-between gap-3 text-sm">
                  <div className="flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-sm" style={{ background: color }} />
                    {source}
                  </div>
                  <strong>{percent}</strong>
                </div>
              ))
            )}
          </div>
        </Panel>
      </div>
    </div>
  );
}
