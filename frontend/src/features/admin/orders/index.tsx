import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { getApiV1AdminOrders, patchApiV1AdminOrdersIdStatus } from '../../../lib/api/generated/rueCosmeticsAPI';
import { StatusTag, Panel } from '../../shared/ui/admin';

export function AdminOrders() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [statusFilter, setStatusFilter] = useState('');

  const { data: ordersData, isLoading, error } = useQuery({
    queryKey: ['admin', 'orders', page, statusFilter],
    queryFn: () => getApiV1AdminOrders({ page: page + 1, page_size: 20, status: statusFilter || undefined }),
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      patchApiV1AdminOrdersIdStatus({ id, data: { status } }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'orders'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'dashboard'] });
    },
  });

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toLocaleString()}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    });
  };

  const totalPages = ordersData?.data.total_pages || 1;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-ink-muted">Loading orders...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg">
        Failed to load orders: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  const orders = ordersData?.data.orders || [];

  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4 flex-wrap">
        <div>
          <div className="text-lavender-700 text-sm mb-1">Fulfilment</div>
          <h1 className="font-display text-4xl font-normal">Orders</h1>
        </div>
        <div className="flex gap-2">
          <button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 transition-colors">
            Export
          </button>
          <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
            Print labels ({orders.length})
          </button>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        {[
          ['New (24h)', orders.filter((o) => o.status === 'pending').length || 0],
          ['To pack', orders.filter((o) => o.status === 'paid').length || 0],
          ['To ship', orders.filter((o) => o.status === 'fulfilled').length || 0],
          ['Shipped today', orders.filter((o) => o.status === 'shipped').length || 0],
        ].map(([label, value]) => (
          <div key={label} className="bg-white border border-line rounded-xl p-5">
            <div className="text-[10px] uppercase tracking-wider text-ink-muted">{label}</div>
            <div className="font-display text-[32px] font-normal tracking-tight mt-2">{value}</div>
          </div>
        ))}
      </div>

      {/* Orders Table */}
      <Panel>
        {/* Filter Bar */}
        <div className="flex gap-2 px-6 py-3 border-b border-line bg-[#FAFAFA] flex-wrap items-center">
          <input
            type="search"
            placeholder="Order id, customer, email…"
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white flex-1 min-w-[200px] focus:outline-none focus:border-lavender-400"
          />
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              setPage(0);
            }}
            className="px-3 py-2 border border-line rounded-lg text-sm bg-white focus:outline-none focus:border-lavender-400"
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

        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="text-left">
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                <input type="checkbox" className="rounded" />
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Order
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Customer
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Date
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Total
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold">
                Status
              </th>
              <th className="px-3 py-2 bg-[#FAFAFA] border-b border-line text-[10px] uppercase tracking-wider text-ink-muted font-bold" />
            </tr>
          </thead>
          <tbody>
            {orders.map((order) => (
              <tr key={order.id} className="hover:bg-[#FAFAFA] transition-colors">
                <td className="px-3 py-3 border-b border-line-soft">
                  <input type="checkbox" className="rounded" />
                </td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                  {order.id.slice(0, 8).toUpperCase()}
                </td>
                <td className="px-3 py-3 border-b border-line-soft">{order.customer_name || order.customer_email}</td>
                <td className="px-3 py-3 border-b border-line-soft">{formatDate(order.created_at)}</td>
                <td className="px-3 py-3 border-b border-line-soft font-variant-numeric tabular-nums font-semibold">
                  {formatCurrency(order.total_ghs_minor)}
                </td>
                <td className="px-3 py-3 border-b border-line-soft">
                  <select
                    value={order.status}
                    onChange={(e) => updateStatusMutation.mutate({ id: order.id, status: e.target.value })}
                    disabled={updateStatusMutation.isPending}
                    className="text-xs rounded-lg px-2 py-1 border-0 focus:ring-2 focus:ring-lavender-400 cursor-pointer"
                    style={{ backgroundColor: 'transparent' }}
                  >
                    <option value="pending">Pending</option>
                    <option value="paid">Paid</option>
                    <option value="fulfilled">Fulfilled</option>
                    <option value="shipped">Shipped</option>
                    <option value="delivered">Delivered</option>
                    <option value="cancelled">Cancelled</option>
                  </select>
                </td>
                <td className="px-3 py-3 border-b border-line-soft">
                  <button
                    onClick={() => navigate({ to: `/admin/orders/${order.id}` })}
                    className="text-lavender-700 font-semibold px-2 py-1 text-sm hover:underline"
                  >
                    Open
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-4">
            <button
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
              className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="px-4 py-2 text-sm text-ink-muted">Page {page + 1} of {totalPages}</span>
            <button
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
              className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        )}
      </Panel>
    </div>
  );
}
