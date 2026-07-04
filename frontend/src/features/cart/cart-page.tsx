import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { useCart } from './cart-provider';
import { formatGhs } from '../../lib/format/utils';
import { CartItemRow } from './cart-item-row';

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
            <CartItemRow
              key={item.id}
              item={item}
              onUpdateQty={(id, qty) => void updateItem(id, qty)}
              onRemove={(id) => void removeItem(id)}
            />
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
          <button className="drawer-link" onClick={() => navigate({ to: '/shop' })}>Continue shopping</button>
        </aside>
      </div>
    </main>
  );
}
