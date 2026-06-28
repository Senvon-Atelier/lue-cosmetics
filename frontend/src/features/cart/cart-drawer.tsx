import { useEffect } from 'react';
import { Link } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { Button } from '../shared/ui/button';
import { useCart } from './cart-provider';
import { formatPrice, getImageUrl } from '../../lib/format/utils';
import { ProductPlaceholder } from '../shared/ui/placeholder';

interface CartDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function CartDrawer({ open, onClose }: CartDrawerProps) {
  const { items, subtotalGhsMinor, shippingCostGhsMinor, totalGhsMinor, isLoading } = useCart();

  // Close drawer on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && open) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleEscape);
    return () => window.removeEventListener('keydown', handleEscape);
  }, [open, onClose]);

  // Prevent body scroll when drawer is open
  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [open]);

  const handleRemoveItem = async (_itemId: string | undefined) => {
    // This would call the cart provider's removeItem
    // For now, we'll just close the drawer
    onClose();
  };

  const handleUpdateQuantity = async (_itemId: string | undefined, newQty: number) => {
    if (newQty < 1) return;
    // This would call the cart provider's updateItem
    // For now, we'll just update locally
  };

  const itemCount = items.reduce((sum, item) => sum + (item.qty || 0), 0);
  const subtotal = formatPrice(subtotalGhsMinor);
  const shipping = formatPrice(shippingCostGhsMinor);
  const total = formatPrice(totalGhsMinor);

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 bg-ink/50 transition-opacity duration-300 z-40 ${open ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
        onClick={onClose}
      />

      {/* Drawer */}
      <aside
        className={`fixed top-0 right-0 h-full w-full max-w-md bg-paper shadow-2xl z-50 transition-transform duration-300 ease-[var(--ease)] ${
          open ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-line-soft">
            <div>
              <div className="font-label text-sm text-ink-muted">Your Bag</div>
              <div className="font-display text-2xl">{itemCount} {itemCount === 1 ? 'item' : 'items'}</div>
            </div>
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-lavender-50 transition-colors"
              aria-label="Close cart"
            >
              <Icon name="close" size={20} />
            </button>
          </div>

          {/* Body */}
          <div className="flex-1 overflow-y-auto p-6">
            {isLoading ? (
              <div className="space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="flex gap-4 animate-pulse">
                    <div className="w-20 h-24 bg-lavender-100 rounded" />
                    <div className="flex-1 space-y-2">
                      <div className="h-4 bg-lavender-100 rounded w-3/4" />
                      <div className="h-3 bg-lavender-100 rounded w-1/2" />
                    </div>
                  </div>
                ))}
              </div>
            ) : items.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full text-center">
                <div className="w-20 h-20 rounded-full bg-lavender-100 flex items-center justify-center mb-4">
                  <Icon name="bag" size={32} className="text-lavender-300" />
                </div>
                <h3 className="font-display text-2xl mb-2">Your bag is empty</h3>
                <p className="text-ink-muted mb-6">Let's change that.</p>
                <Link to="/shop">
                  <Button onClick={onClose} variant="primary" icon="arrow" iconPosition="right">
                    Shop the edit
                  </Button>
                </Link>
              </div>
            ) : (
              <div className="space-y-4">
                {items.map((item) => (
                  <div key={item.id} className="flex gap-4 p-4 bg-lavender-50 rounded-lg">
                    {/* Product Image */}
                    <div className="relative w-20 h-24 flex-shrink-0 overflow-hidden rounded bg-lavender-100">
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
                        <h4 className="font-label font-medium text-ink mb-1">{item.product_name}</h4>
                      </div>

                      {/* Quantity and Price */}
                      <div className="flex items-center justify-between mt-2">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleUpdateQuantity(item.id, (item.qty || 1) - 1)}
                            className="w-8 h-8 flex items-center justify-center border border-line rounded hover:bg-lavender-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                            disabled={(item.qty || 1) <= 1}
                          >
                            <Icon name="minus" size={12} />
                          </button>
                          <span className="w-8 text-center font-label">{item.qty || 1}</span>
                          <button
                            onClick={() => handleUpdateQuantity(item.id, (item.qty || 1) + 1)}
                            className="w-8 h-8 flex items-center justify-center border border-line rounded hover:bg-lavender-50 transition-colors"
                          >
                            <Icon name="plus" size={12} />
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
                      className="self-start p-2 text-ink-muted hover:text-ink transition-colors"
                      aria-label="Remove item"
                    >
                      <Icon name="close" size={14} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Footer */}
          {items.length > 0 && !isLoading && (
            <div className="border-t border-line-soft p-6 space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-ink-muted">Subtotal</span>
                <span className="font-label font-semibold">{subtotal}</span>
              </div>
              <div className="flex justify-between text-sm text-ink-muted">
                <span>Delivery</span>
                <span>{shipping}</span>
              </div>
              <Link to="/checkout" onClick={onClose}>
                <Button className="w-full" icon="arrow" iconPosition="right">
                  Checkout · {total}
                </Button>
              </Link>
              <button
                onClick={onClose}
                className="w-full text-sm text-ink-muted hover:text-ink transition-colors py-2"
              >
                Continue shopping
              </button>
            </div>
          )}
        </div>
      </aside>
    </>
  );
}
