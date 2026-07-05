import { useEffect, useState } from 'react';
import { Link } from '@tanstack/react-router';
import {
  getMeOrders,
  type InternalMeOrderResponse,
} from '../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate } from '../../lib/format/utils';
import { AcctHead, StatusPill } from './acct-primitives';

const STATUS_TABS: Array<[value: string, label: string]> = [
  ['', 'All'],
  ['pending', 'Pending'],
  ['paid', 'Paid'],
  ['failed', 'Failed'],
  ['cancelled', 'Cancelled'],
];

export function AccountOrders() {
  const [orders, setOrders] = useState<InternalMeOrderResponse[]>([]);
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

  return (
    <main className="acct-main">
      <AcctHead eyebrow="History" title="Your orders" />

      <div className="sub-tabs" style={{ width: 'fit-content' }}>
        {STATUS_TABS.map(([value, label]) => (
          <button
            key={label}
            className={statusFilter === value ? 'active' : ''}
            onClick={() => {
              setStatusFilter(value);
              setPage(0);
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {error && <div className="alert alert-warn">{error}</div>}

      {isLoading ? (
        <div className="acct-empty">
          <p>Loading orders…</p>
        </div>
      ) : orders.length === 0 ? (
        <div className="acct-empty">
          <p>
            {statusFilter
              ? 'No orders with this status.'
              : 'No orders yet — when you place one, it will appear here.'}
          </p>
          <Link className="btn btn-primary" to="/shop" search={{}}>
            Start shopping
          </Link>
        </div>
      ) : (
        <>
          <div className="orders-table">
            <div className="orders-row head">
              <div>Order</div>
              <div>Date</div>
              <div>Total</div>
              <div>Status</div>
              <div></div>
            </div>
            {orders.map((order) => (
              <div key={order.id} className="orders-row">
                <div className="o-id">
                  #{(order.id || '').slice(0, 8).toUpperCase()}
                </div>
                <div>{formatOrderDate(order.created_at)}</div>
                <div className="price">{formatGhs(order.total_ghs_minor || 0)}</div>
                <div>
                  <StatusPill status={order.status} />
                </div>
                <div>
                  <Link
                    className="link-btn"
                    to="/account/orders/$id"
                    params={{ id: order.id || '' }}
                  >
                    View
                  </Link>
                </div>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="acct-pagination">
              <button
                className="btn btn-ghost"
                disabled={page === 0}
                onClick={() => setPage(page - 1)}
              >
                Previous
              </button>
              <span>
                Page {page + 1} of {totalPages}
              </span>
              <button
                className="btn btn-ghost"
                disabled={page >= totalPages - 1}
                onClick={() => setPage(page + 1)}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </main>
  );
}
