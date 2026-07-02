import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { useCart } from './cart-provider';
import { formatGhs, getImageUrl } from '../../lib/format/utils';

interface CartDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function CartDrawer({ open, onClose }: CartDrawerProps) {
  const { items, subtotalGhsMinor, updateItem, removeItem } = useCart();
  const navigate = useNavigate();

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

  const itemCount = items.reduce((sum, item) => sum + (item.qty || 0), 0);

  return (
    <>
      <div className={`drawer-scrim${open ? ' open' : ''}`} onClick={onClose} />
      <aside className={`drawer${open ? ' open' : ''}`} aria-hidden={!open}>
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
