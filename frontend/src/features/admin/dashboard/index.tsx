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
        <Panel title="Recent orders" flush>
          {recentOrders.length === 0 ? (
            <div className="admin-panel-body"><p className="admin-empty">No orders yet.</p></div>
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

        <Panel title="Orders by status" flush>
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
            <div className="admin-panel-body"><p className="admin-empty">No stats available.</p></div>
          )}
        </Panel>
      </div>
    </>
  );
}
