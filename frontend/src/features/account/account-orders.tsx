import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getMeOrders } from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

type Order = {
  id?: string;
  user_id?: string;
  status?: string;
  subtotal_ghs?: number;
  shipping_ghs?: number;
  total_ghs?: number;
  paystack_reference?: string;
  created_at?: string;
  updated_at?: string;
};

export function AccountOrders() {
  const navigate = useNavigate();
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [statusFilter, setStatusFilter] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const limit = 10;
  const offset = page * limit;

  const loadOrders = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await getMeOrders({
        limit,
        offset,
        status: statusFilter || undefined,
      });
      setOrders(response.orders || []);
      setTotal(response.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load orders');
    } finally {
      setIsLoading(false);
    }
  };

  // Load orders on mount and when filter/page changes
  useEffect(() => {
    loadOrders();
  }, [page, statusFilter]);

  const totalPages = Math.ceil(total / limit);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toFixed(2)}`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'paid':
        return 'text-green-600';
      case 'pending':
        return 'text-yellow-600';
      case 'failed':
        return 'text-rose-600';
      case 'cancelled':
        return 'text-gray-600';
      default:
        return 'text-ink';
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="font-display text-xl mb-2">Order History</h2>
          <p className="text-ink-muted">View and track your orders.</p>
        </div>
      </div>

      {/* Status Filter */}
      <div className="mb-6">
        <select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value);
            setPage(0);
          }}
          className="px-4 py-2 border border-line rounded-lg bg-paper text-ink focus:outline-none focus:border-lavender-400"
        >
          <option value="">All Orders</option>
          <option value="pending">Pending</option>
          <option value="paid">Paid</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg mb-4">
          {error}
        </div>
      )}

      {/* Loading State */}
      {isLoading ? (
        <div className="text-center py-12">
          <div className="text-ink-muted">Loading orders...</div>
        </div>
      ) : orders.length === 0 ? (
        /* Empty State */
        <div className="text-center py-12">
          <div className="text-4xl mb-4">📦</div>
          <h3 className="font-display text-xl mb-2">No orders yet</h3>
          <p className="text-ink-muted mb-6">When you place an order, it will appear here.</p>
          <Button variant="primary" onClick={() => navigate({ to: '/shop' })}>
            Start Shopping
          </Button>
        </div>
      ) : (
        /* Orders List */
        <div className="space-y-4">
          {orders.map((order) => (
            <div
              key={order.id}
              className="bg-white rounded-lg p-6"
              style={{ border: '1px solid var(--line)' }}
            >
              <div className="flex items-start justify-between mb-4">
                <div>
                  <div className="text-sm text-ink-muted mb-1">
                    Order ID: {(order.id || '').slice(0, 8).toUpperCase()}
                  </div>
                  <div className="text-sm text-ink-muted">
                    {formatDate(order.created_at || '')}
                  </div>
                </div>
                <div className={`font-label font-medium capitalize ${getStatusColor(order.status || '')}`}>
                  {order.status}
                </div>
              </div>

              <div className="space-y-1 text-sm">
                <div className="flex justify-between">
                  <span className="text-ink-muted">Subtotal:</span>
                  <span className="font-medium">{formatCurrency(order.subtotal_ghs || 0)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-ink-muted">Shipping:</span>
                  <span className="font-medium">{formatCurrency(order.shipping_ghs || 0)}</span>
                </div>
                <div className="flex justify-between text-base">
                  <span className="font-medium">Total:</span>
                  <span className="font-display font-semibold">{formatCurrency(order.total_ghs || 0)}</span>
                </div>
              </div>

              <div className="mt-4 flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate({ to: `/account/orders/${order.id}` })}
                >
                  View Details
                </Button>
              </div>
            </div>
          ))}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-6">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 0}
                onClick={() => setPage(page - 1)}
              >
                Previous
              </Button>
              <span className="px-4 py-2 text-sm text-ink-muted">
                Page {page + 1} of {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages - 1}
                onClick={() => setPage(page + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
