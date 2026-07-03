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
