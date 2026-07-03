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
        <Panel title="Top products" flush>
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
            <div className="admin-panel-body">
              <p className="admin-empty">No sales data yet.</p>
            </div>
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
