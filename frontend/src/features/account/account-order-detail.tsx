import { useEffect, useState } from 'react';
import { Link, useParams } from '@tanstack/react-router';
import {
  getMeOrdersId,
  type InternalMeOrderDetailResponse,
} from '../../lib/api/generated/rueCosmeticsAPI';
import { formatGhs, formatOrderDate, getImageUrl } from '../../lib/format/utils';
import { Icon } from '../shared/ui/icons';
import { AcctHead, StatusPill } from './acct-primitives';

export function AccountOrderDetail() {
  const { id } = useParams({ from: '/_storefront/account/orders/$id' });
  const [order, setOrder] = useState<InternalMeOrderDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadOrder = async () => {
      if (!id) return;
      setIsLoading(true);
      setError(null);
      try {
        const response = await getMeOrdersId(id);
        setOrder(response);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load order details');
      } finally {
        setIsLoading(false);
      }
    };
    loadOrder();
  }, [id]);

  if (isLoading) {
    return (
      <main className="acct-main">
        <div className="acct-empty">
          <p>Loading order details…</p>
        </div>
      </main>
    );
  }

  if (error || !order) {
    return (
      <main className="acct-main">
        <Link className="back-link" to="/account/orders">
          <Icon name="arrowLeft" size={12} /> All orders
        </Link>
        <div className="alert alert-warn">
          {error || 'Unable to load order details'}
        </div>
      </main>
    );
  }

  const shortId = (order.id || '').slice(0, 8).toUpperCase();

  return (
    <main className="acct-main">
      <Link className="back-link" to="/account/orders">
        <Icon name="arrowLeft" size={12} /> All orders
      </Link>
      <AcctHead eyebrow={`Order #${shortId}`} title="Order details">
        <div className="acct-head-actions">
          <StatusPill status={order.status} />
        </div>
      </AcctHead>
      <div className="acct-placed">
        Placed {formatOrderDate(order.created_at)}
      </div>

      <div className="order-detail-grid">
        <div>
          <div className="form-card" style={{ maxWidth: 'none' }}>
            <h3 className="form-card-title">
              Items ({order.items?.length || 0})
            </h3>
            {(order.items || []).map((it) => (
              <div key={it.id} className="od-item">
                <div className="od-item-ph ph ph--lavender">
                  {it.product_image_snapshot ? (
                    <img
                      src={getImageUrl(it.product_image_snapshot)}
                      alt={it.product_name_snapshot || ''}
                    />
                  ) : (
                    <span className="ph-label">
                      {it.product_name_snapshot}
                    </span>
                  )}
                </div>
                <div className="od-item-body">
                  <div className="od-item-name">
                    {it.product_name_snapshot}
                  </div>
                  {it.product_brand_snapshot && (
                    <div className="od-item-meta">
                      {it.product_brand_snapshot}
                    </div>
                  )}
                  <div className="od-item-meta">Qty {it.qty}</div>
                </div>
                <div className="price">
                  {formatGhs((it.unit_price_ghs_minor || 0) * (it.qty || 0))}
                </div>
              </div>
            ))}
          </div>
        </div>
        <div>
          <div className="form-card" style={{ padding: 24 }}>
            <h3 className="form-card-title" style={{ fontSize: 18 }}>
              Summary
            </h3>
            <div className="kv-row">
              <span>Subtotal</span>
              <span>{formatGhs(order.subtotal_ghs_minor || 0)}</span>
            </div>
            <div className="kv-row">
              <span>Shipping</span>
              <span>{formatGhs(order.shipping_ghs_minor || 0)}</span>
            </div>
            <div className="kv-divider"></div>
            <div className="kv-row">
              <strong>Total</strong>
              <strong className="price">
                {formatGhs(order.total_ghs_minor || 0)}
              </strong>
            </div>
            {order.shipping_address && (
              <div className="kv-block">
                <div className="kv-block-k">Delivering to</div>
                <div style={{ whiteSpace: 'pre-line' }}>
                  {order.shipping_address}
                </div>
              </div>
            )}
            {order.paystack_reference && (
              <div className="kv-block">
                <div className="kv-block-k">Payment</div>
                <div>Paystack · {order.paystack_reference}</div>
              </div>
            )}
          </div>
          <div style={{ marginTop: 16 }}>
            <Link className="btn btn-ghost" to="/shop" search={{}}>
              Continue shopping
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
