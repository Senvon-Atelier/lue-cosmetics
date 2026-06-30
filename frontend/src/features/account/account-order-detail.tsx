import { useState, useEffect } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { getApiV1MeOrdersId } from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

type OrderItem = {
  id?: string;
  order_id?: string;
  product_id?: string;
  qty?: number;
  unit_price_ghs?: number;
  product_name_snapshot?: string;
  product_brand_snapshot?: string;
  product_image_snapshot?: string;
};

type OrderDetailResponse = {
  id?: string;
  user_id?: string;
  status?: string;
  subtotal_ghs?: number;
  shipping_ghs?: number;
  total_ghs?: number;
  paystack_reference?: string;
  shipping_address?: string;
  created_at?: string;
  updated_at?: string;
  items?: OrderItem[];
};

export function AccountOrderDetail() {
  const { id } = useParams({ from: '/_marketing/account/orders/$id' });
  const navigate = useNavigate();
  const [order, setOrder] = useState<OrderDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadOrder = async () => {
      if (!id) return;

      setIsLoading(true);
      setError(null);
      try {
        const response = await getApiV1MeOrdersId(id);
        setOrder(response.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load order details');
      } finally {
        setIsLoading(false);
      }
    };

    loadOrder();
  }, [id]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
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

  const handleReorder = async () => {
    if (!order || !order.items || !order.items.length) return;

    // Add all items to cart
    // This would require cart context integration
    alert('Reorder functionality would add all items to cart');
  };

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="text-ink-muted">Loading order details...</div>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="text-center py-12">
        <div className="text-4xl mb-4">⚠️</div>
        <h3 className="font-display text-xl mb-2">Order Not Found</h3>
        <p className="text-ink-muted mb-6">{error || 'Unable to load order details'}</p>
        <Button variant="outline" onClick={() => navigate({ to: '/account/orders' })}>
          Back to Orders
        </Button>
      </div>
    );
  }

  return (
    <div>
      {/* Back Button */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate({ to: '/account/orders' })}
        className="mb-4"
      >
        ← Back to Orders
      </Button>

      {/* Order Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between mb-4">
          <div>
            <h2 className="font-display text-xl mb-2">Order Details</h2>
            <div className="text-sm text-ink-muted">
              Order ID: {order.id?.slice(0, 8).toUpperCase() || 'N/A'}
            </div>
          </div>
          <div className={`font-label font-medium text-lg capitalize ${getStatusColor(order.status || '')}`}>
            {order.status}
          </div>
        </div>
        <div className="text-sm text-ink-muted">
          Placed on {formatDate(order.created_at || '')}
        </div>
      </div>

      {/* Shipping Address */}
      <div className="bg-white rounded-lg p-6 mb-6" style={{ border: '1px solid var(--line)' }}>
        <h3 className="font-label font-semibold mb-4">Shipping Address</h3>
        <div className="text-ink-soft">
          {order.shipping_address ? (
            <div className="whitespace-pre-line">{order.shipping_address}</div>
          ) : (
            <div className="text-ink-muted">No shipping address available</div>
          )}
        </div>
      </div>

      {/* Order Items */}
      <div className="bg-white rounded-lg p-6 mb-6" style={{ border: '1px solid var(--line)' }}>
        <h3 className="font-label font-semibold mb-4">Items ({order.items?.length || 0})</h3>
        <div className="space-y-4">
          {(order.items || []).map((item) => (
            <div
              key={item.id}
              className="flex items-start gap-4 pb-4"
              style={{ borderBottom: '1px solid var(--line-soft)' }}
            >
              {/* Product Image */}
              <div className="w-16 h-16 bg-lavender-50 rounded flex-shrink-0">
                {item.product_image_snapshot ? (
                  <img
                    src={item.product_image_snapshot}
                    alt={item.product_name_snapshot || ''}
                    className="w-full h-full object-cover rounded"
                    onError={(e) => {
                      (e.target as HTMLImageElement).style.display = 'none';
                    }}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-ink-muted text-xs">
                    No image
                  </div>
                )}
              </div>

              {/* Product Info */}
              <div className="flex-1 min-w-0">
                <h4 className="font-label font-medium mb-1">{item.product_name_snapshot}</h4>
                <p className="text-sm text-ink-muted mb-1">{item.product_brand_snapshot}</p>
                <div className="flex items-center gap-4 text-sm">
                  <div className="text-ink-muted">Qty: {item.qty}</div>
                  <div className="font-medium">{formatCurrency(item.unit_price_ghs || 0)}</div>
                </div>
              </div>

              {/* Line Total */}
              <div className="font-label font-medium">
                {formatCurrency((item.unit_price_ghs || 0) * (item.qty || 0))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Order Summary */}
      <div className="bg-white rounded-lg p-6 mb-6" style={{ border: '1px solid var(--line)' }}>
        <h3 className="font-label font-semibold mb-4">Order Summary</h3>
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-ink-muted">Subtotal</span>
            <span>{formatCurrency(order.subtotal_ghs || 0)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-ink-muted">Shipping</span>
            <span>{formatCurrency(order.shipping_ghs || 0)}</span>
          </div>
          <div className="flex justify-between text-base font-semibold pt-2" style={{ borderTop: '1px solid var(--line-soft)' }}>
            <span>Total</span>
            <span className="font-display">{formatCurrency(order.total_ghs || 0)}</span>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-4">
        <Button
          variant="primary"
          onClick={handleReorder}
        >
          Reorder
        </Button>
        <Button
          variant="outline"
          onClick={() => navigate({ to: '/shop' })}
        >
          Continue Shopping
        </Button>
      </div>
    </div>
  );
}
