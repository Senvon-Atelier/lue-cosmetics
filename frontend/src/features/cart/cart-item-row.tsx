import { Icon } from '../shared/ui/icons';
import { formatGhs, getImageUrl } from '../../lib/format/utils';
import type { InternalCartCartItemResponse } from '../../lib/api/generated/rueCosmeticsAPI';

interface CartItemRowProps {
  item: InternalCartCartItemResponse;
  onUpdateQty: (id: string, qty: number) => void;
  onRemove: (id: string) => void;
}

/** One bag line — shared by the cart drawer and the cart page (identical markup). */
export function CartItemRow({ item, onUpdateQty, onRemove }: CartItemRowProps) {
  return (
    <div className="cart-item">
      {item.product_image_path ? (
        <img
          src={getImageUrl(item.product_image_path)}
          alt={item.product_name || 'Product'}
          style={{ width: 80, height: 100, flexShrink: 0, objectFit: 'cover', borderRadius: 'var(--radius)' }}
          loading="lazy"
        />
      ) : (
        <div className="ph ph--lavender" style={{ width: 80, height: 100, flexShrink: 0 }}>
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
                  onRemove(item.id!);
                } else {
                  onUpdateQty(item.id!, newQty);
                }
              }}
              aria-label="Decrease quantity"
            >
              <Icon name="minus" size={12} />
            </button>
            <span>{item.qty || 1}</span>
            <button
              onClick={() => onUpdateQty(item.id!, (item.qty || 1) + 1)}
              aria-label="Increase quantity"
            >
              <Icon name="plus" size={12} />
            </button>
          </div>
          <div className="price">{formatGhs((item.unit_price_ghs_minor || 0) * (item.qty || 1))}</div>
        </div>
      </div>
      <button className="cart-item-remove" onClick={() => onRemove(item.id!)} aria-label="Remove item">
        <Icon name="close" size={14} />
      </button>
    </div>
  );
}
