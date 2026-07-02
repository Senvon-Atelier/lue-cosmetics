import { useParams } from '@tanstack/react-router';
import { useGetAdminOrdersId } from '../../../lib/api/generated/rueCosmeticsAPI';
import { StatusTag, Panel } from '../../shared/ui/admin';

export function AdminOrderDetail() {
  const { id } = useParams({ from: '/admin/orders/$id' });
  const { data: orderDetail, isLoading, error } = useGetAdminOrdersId(id, {
    query: { enabled: !!id },
  });

  const formatCurrency = (amount: number) => {
    return `GH₵${(amount / 100).toLocaleString()}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-ink-muted">Loading order details...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg">
        Failed to load order: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    );
  }

  const order = orderDetail?.order;
  const items = orderDetail?.items ?? [];

  if (!order) {
    return <div>Order not found</div>;
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-start justify-between mb-7 gap-4">
        <div>
          <div className="text-lavender-700 text-sm mb-1">Order Details</div>
          <h1 className="font-display text-4xl font-normal">#{(order.id ?? '').slice(0, 8).toUpperCase()}</h1>
        </div>
        <div className="flex gap-2">
          <button className="px-3 py-2 border border-line rounded-lg text-sm font-semibold hover:bg-lavender-50 transition-colors">
            Print invoice
          </button>
          <button className="px-3 py-2 bg-ink text-white rounded-lg text-sm font-semibold hover:bg-lavender-700 transition-colors">
            Update status
          </button>
        </div>
      </div>

      <div className="grid grid-cols-[2fr_1fr] gap-5">
        {/* Order Items */}
        <div>
          <Panel title="Items">
            <table className="w-full">
              <thead>
                <tr className="text-left border-b border-line">
                  <th className="py-2 text-sm font-semibold">Product</th>
                  <th className="py-2 text-sm font-semibold text-right">Qty</th>
                  <th className="py-2 text-sm font-semibold text-right">Price</th>
                  <th className="py-2 text-sm font-semibold text-right">Total</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.id} className="border-b border-line-soft">
                    <td className="py-3">
                      <div className="font-display text-sm">{item.product_name_snapshot}</div>
                      <div className="text-xs text-ink-muted">{item.product_brand_snapshot}</div>
                    </td>
                    <td className="py-3 text-right font-variant-numeric tabular-nums">{item.qty ?? 0}</td>
                    <td className="py-3 text-right font-variant-numeric tabular-nums text-sm">
                      {formatCurrency(item.unit_price_ghs_minor ?? 0)}
                    </td>
                    <td className="py-3 text-right font-variant-numeric tabular-nums font-semibold">
                      {formatCurrency((item.unit_price_ghs_minor ?? 0) * (item.qty ?? 0))}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Panel>
        </div>

        {/* Order Info */}
        <div className="space-y-5">
          {/* Status */}
          <Panel>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-ink-muted">Status</span>
                <StatusTag status={order.status ?? 'pending'} />
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-ink-muted">Created</span>
                <span className="text-sm">{formatDate(order.created_at ?? '')}</span>
              </div>
              {order.updated_at !== order.created_at && (
                <div className="flex justify-between items-center">
                  <span className="text-sm text-ink-muted">Updated</span>
                  <span className="text-sm">{formatDate(order.updated_at ?? '')}</span>
                </div>
              )}
            </div>
          </Panel>

          {/* Totals */}
          <Panel title="Order totals">
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-ink-muted">Subtotal</span>
                <span className="font-variant-numeric tabular-nums">{formatCurrency(order.subtotal_ghs_minor ?? 0)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-ink-muted">Shipping</span>
                <span className="font-variant-numeric tabular-nums">{formatCurrency(order.shipping_ghs_minor ?? 0)}</span>
              </div>
              <div className="flex justify-between text-base font-semibold border-t border-line pt-2 mt-2">
                <span>Total</span>
                <span className="font-display font-variant-numeric tabular-nums">{formatCurrency(order.total_ghs_minor ?? 0)}</span>
              </div>
            </div>
          </Panel>

          {/* Shipping Address */}
          {order.shipping_address && (
            <Panel title="Shipping address">
              <div className="text-sm space-y-1">
                <div className="font-semibold">{order.shipping_address.line1}</div>
                {order.shipping_address.line2 && <div>{order.shipping_address.line2}</div>}
                <div>
                  {order.shipping_address.city}, {order.shipping_address.region}
                </div>
                <div className="text-ink-muted">{order.shipping_address.phone}</div>
              </div>
            </Panel>
          )}
        </div>
      </div>
    </div>
  );
}
