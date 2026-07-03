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
