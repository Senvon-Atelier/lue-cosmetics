import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { useCart } from './cart-provider';
import { formatGhs } from '../../lib/format/utils';
import { useEscToClose, useLockBodyScroll } from '../shared/use-overlay';
import { CartItemRow } from './cart-item-row';

interface CartDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function CartDrawer({ open, onClose }: CartDrawerProps) {
  const { items, subtotalGhsMinor, updateItem, removeItem } = useCart();
  const navigate = useNavigate();

  useEscToClose(open, onClose);
  useLockBodyScroll(open);

  const itemCount = items.reduce((sum, item) => sum + (item.qty || 0), 0);

  return (
    <>
      <div className={`drawer-scrim${open ? ' open' : ''}`} onClick={onClose} />
      <aside className={`drawer${open ? ' open' : ''}`} aria-hidden={!open} inert={open ? undefined : ''}>
        <div className="drawer-head">
          <div>
            <div className="eyebrow">Your Bag</div>
            <div className="drawer-title">
              {itemCount} {itemCount === 1 ? 'item' : 'items'}
            </div>
          </div>
          <button className="icon-btn" onClick={onClose} aria-label="Close cart">
            <Icon name="close" size={20} />
          </button>
        </div>

        <div className="drawer-body">
          {items.length === 0 ? (
            <div className="cart-empty">
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
                onClick={() => {
                  onClose();
                  void navigate({ to: '/shop' });
                }}
              >
                Shop the edit <Icon name="arrow" size={14} />
              </button>
            </div>
          ) : (
            items.map((item) => (
              <CartItemRow
                key={item.id}
                item={item}
                onUpdateQty={(id, qty) => void updateItem(id, qty)}
                onRemove={(id) => void removeItem(id)}
              />
            ))
          )}
        </div>

        {items.length > 0 && (
          <div className="drawer-foot">
            <div className="drawer-row">
              <span>Subtotal</span>
              <span className="price">{formatGhs(subtotalGhsMinor)}</span>
            </div>
            <div className="drawer-row muted">
              <span>Delivery</span>
              <span>Calculated at checkout</span>
            </div>
            <button
              className="btn btn-primary"
              style={{ width: '100%', justifyContent: 'center', marginTop: 16 }}
              onClick={() => {
                onClose();
                void navigate({ to: '/checkout' });
              }}
            >
              Checkout · {formatGhs(subtotalGhsMinor)}
            </button>
            <button className="drawer-link" onClick={onClose}>
              Continue shopping
            </button>
          </div>
        )}
      </aside>
    </>
  );
}
