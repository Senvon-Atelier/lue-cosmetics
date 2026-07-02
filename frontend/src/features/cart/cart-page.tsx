import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { useCart } from './cart-provider';
import { formatGhs, getImageUrl } from '../../lib/format/utils';

export function CartPage() {
  const { items, subtotalGhsMinor, freeShippingRemainderGhsMinor, updateItem, removeItem } = useCart();
  const navigate = useNavigate();

  // Rename to match skeleton bindings
  const subtotalMinor = subtotalGhsMinor;
  const remainderMinor = freeShippingRemainderGhsMinor;

  if (items.length === 0) {
    return (
      <main className="wrap cart-page fade-up">
        <div className="eyebrow">Your Bag</div>
        <h1 className="h-display" style={{ fontSize: 'clamp(40px, 6vw, 80px)', marginTop: 8 }}>
          Ready when <em>you are.</em>
        </h1>
        <div className="cart-empty" style={{ marginTop: 40 }}>
          <div
            className="ph"
            style={{ width: 120, height: 120, margin: '0 auto 24px', borderRadius: '50%' }}
          >
            <span className="ph-label">Empty</span>
          </div>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 24, margin: '0 0 8px' }}>
            Your bag is empty
          </h3>
          <p style={{ color: 'var(--ink-muted)', marginBottom: 24 }}>Let's change that.</p>
          <button
            className="btn btn-primary"
            onClick={() => void navigate({ to: '/shop' })}
          >
            Shop the edit <Icon name="arrow" size={14} />
          </button>
        </div>
      </main>
    );
  }

  return (
    <main className="wrap cart-page fade-up">
      <div className="eyebrow">Your Bag</div>
      <h1 className="h-display" style={{ fontSize: 'clamp(40px, 6vw, 80px)', marginTop: 8 }}>
        Ready when <em>you are.</em>
      </h1>
      <div className="cart-page-grid" style={{ marginTop: 40 }}>
        <div>
          {items.map((item) => (
            <div className="cart-item" key={item.id}>
              {item.product_image_path ? (
                <img
                  src={getImageUrl(item.product_image_path)}
                  alt={item.product_name || 'Product'}
                  style={{ width: 80, height: 100, flexShrink: 0, objectFit: 'cover', borderRadius: 'var(--radius)' }}
                  loading="lazy"
                />
              ) : (
                <div
                  className="ph ph--lavender"
                  style={{ width: 80, height: 100, flexShrink: 0 }}
                >
                  <span className="ph-label" style={{ fontSize: 8 }}>
                    {item.product_name?.substring(0, 2) || ''}
                  </span>
                </div>
              )}
              <div className="cart-item-body">
                <div className="cart-item-name">{item.product_name}</div>
                <div className="cart-item-row">
                  <div className="qty">
                    <button
                      onClick={() => {
                        const newQty = (item.qty || 1) - 1;
                        if (newQty < 1) {
                          void removeItem(item.id!);
                        } else {
                          void updateItem(item.id!, newQty);
                        }
                      }}
                      aria-label="Decrease quantity"
                    >
                      <Icon name="minus" size={12} />
                    </button>
                    <span>{item.qty || 1}</span>
                    <button
                      onClick={() => void updateItem(item.id!, (item.qty || 1) + 1)}
                      aria-label="Increase quantity"
                    >
                      <Icon name="plus" size={12} />
                    </button>
                  </div>
                  <div className="price">
                    {formatGhs((item.unit_price_ghs_minor || 0) * (item.qty || 1))}
                  </div>
                </div>
              </div>
              <button
                className="cart-item-remove"
                onClick={() => void removeItem(item.id!)}
                aria-label="Remove item"
              >
                <Icon name="close" size={14} />
              </button>
            </div>
          ))}
        </div>
        <aside className="cart-summary">
          <h3>Order summary</h3>
          <div className="drawer-row"><span>Subtotal</span><span className="price">{formatGhs(subtotalMinor)}</span></div>
          <div className="drawer-row muted"><span>Delivery</span><span>Calculated at checkout</span></div>
          {remainderMinor > 0 && (
            <div className="cart-free-note">Add {formatGhs(remainderMinor)} more for free delivery in Accra.</div>
          )}
          <button
            className="btn btn-primary"
            style={{ width: '100%', justifyContent: 'center', marginTop: 16 }}
            onClick={() => void navigate({ to: '/checkout' })}
          >
            Checkout · {formatGhs(subtotalMinor)}
          </button>
        </aside>
      </div>
    </main>
  );
}
