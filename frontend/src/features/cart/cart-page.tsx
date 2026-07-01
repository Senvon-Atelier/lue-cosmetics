import { Link } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { Button } from '../shared/ui/button';
import { useCart } from './cart-provider';
import { formatPrice, getImageUrl } from '../../lib/format/utils';
import { ProductPlaceholder } from '../shared/ui/placeholder';

export function CartPage() {
  const { items, subtotalGhsMinor, shippingCostGhsMinor, totalGhsMinor, isLoading, refreshCart } = useCart();

  const handleUpdateQuantity = async (_itemId: string | undefined, newQty: number) => {
    if (newQty < 1) return;
    // This would call the cart provider's updateItem
    // For now, we'll just update locally
  };

  const handleRemoveItem = async (_itemId: string | undefined) => {
    // This would call the cart provider's removeItem
    // For now, we'll just refresh the cart
    await refreshCart();
  };

  const itemCount = items.reduce((sum, item) => sum + (item.qty || 0), 0);
  const subtotal = formatPrice(subtotalGhsMinor);
  const shipping = formatPrice(shippingCostGhsMinor);
  const total = formatPrice(totalGhsMinor);

  return (
    <div className="section">
      <div className="wrap">
        <div className="mb-12">
          <div className="eyebrow">Cart</div>
          <h1 className="font-display text-[clamp(32px,4vw,56px)] font-normal tracking-[-0.01em]">
            Shopping Cart
          </h1>
          <p className="text-ink-muted mt-2">{itemCount} {itemCount === 1 ? 'item' : 'items'}</p>
        </div>

        {isLoading ? (
          <div className="space-y-5">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex gap-4 animate-pulse">
                <div className="w-32 h-32 bg-lavender-100 rounded" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-lavender-100 rounded w-3/4" />
                  <div className="h-3 bg-lavender-100 rounded w-1/2" />
                </div>
              </div>
            ))}
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-20 h-20 rounded-full bg-lavender-100 flex items-center justify-center mb-4">
              <Icon name="bag" size={32} className="text-lavender-300" />
            </div>
            <h2 className="font-display text-2xl mb-2">Your cart is empty</h2>
            <p className="text-ink-muted mb-6">Let's change that.</p>
            <Link to="/shop">
              <Button variant="primary" icon="arrowRight" iconPosition="right">
                Shop the edit
              </Button>
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-[2fr_1fr] gap-12">
            {/* Cart Items */}
            <div className="space-y-5">
              {items.map((item) => (
                <div key={item.id} className="flex gap-6 py-5 border-b border-line-soft last:border-0">
                  {/* Product Image */}
                  <div className="relative w-32 h-32 flex-shrink-0 overflow-hidden rounded bg-lavender-100">
                    {item.product_image_path ? (
                      <img
                        src={getImageUrl(item.product_image_path || '')}
                        alt={item.product_name || 'Product'}
                        className="w-full h-full object-cover"
                        loading="lazy"
                      />
                    ) : (
                      <ProductPlaceholder tone="lavender" label={item.product_name?.substring(0, 2) || ''} />
                    )}
                  </div>

                  {/* Product Info */}
                  <div className="flex-1 flex flex-col">
                    <div className="flex-1">
                      <h3 className="font-label font-semibold text-ink mb-1">{item.product_name}</h3>
                    </div>

                    {/* Quantity and Price */}
                    <div className="flex items-center justify-between mt-4">
                      <div className="flex items-center gap-3">
                        <button
                          onClick={() => handleUpdateQuantity(item.id, (item.qty || 1) - 1)}
                          className="w-10 h-10 flex items-center justify-center border border-line rounded-full hover:bg-lavender-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={(item.qty || 1) <= 1}
                        >
                          <Icon name="minus" size={16} />
                        </button>
                        <span className="w-12 text-center font-label font-semibold">{item.qty || 1}</span>
                        <button
                          onClick={() => handleUpdateQuantity(item.id, (item.qty || 1) + 1)}
                          className="w-10 h-10 flex items-center justify-center border border-line rounded-full hover:bg-lavender-50 transition-colors"
                        >
                          <Icon name="plus" size={16} />
                        </button>
                      </div>

                      <div className="text-right">
                        <div className="font-label font-semibold">{formatPrice((item.unit_price_ghs_minor || 0) * (item.qty || 1))}</div>
                      </div>
                    </div>
                  </div>

                  {/* Remove Button */}
                  <button
                    onClick={() => handleRemoveItem(item.id)}
                    className="self-start p-2 text-ink-muted hover:text-red-600 transition-colors"
                    aria-label="Remove item"
                  >
                    <Icon name="close" size={16} />
                  </button>
                </div>
              ))}
            </div>

            {/* Order Summary */}
            <div>
              <div className="border border-line-soft rounded p-6 sticky top-24">
                <h2 className="font-display text-2xl mb-4">Order Summary</h2>

                <div className="space-y-3 mb-6">
                  <div className="flex justify-between text-sm">
                    <span className="text-ink-muted">Subtotal</span>
                    <span className="font-label font-semibold">{subtotal}</span>
                  </div>
                  <div className="flex justify-between text-sm text-ink-muted">
                    <span>Delivery</span>
                    <span>{shipping}</span>
                  </div>
                  <div className="border-t border-line-soft pt-3 mt-3">
                    <div className="flex justify-between">
                      <span className="font-label font-semibold">Total</span>
                      <span className="font-display text-xl">{total}</span>
                    </div>
                  </div>
                </div>

                <Link to="/checkout">
                  <Button className="w-full" icon="arrowRight" iconPosition="right">
                    Proceed to Checkout
                  </Button>
                </Link>

                <Link to="/shop">
                  <button
                    className="w-full text-sm text-ink-muted hover:text-ink transition-colors py-2 mt-4"
                  >
                    Continue shopping
                  </button>
                </Link>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
