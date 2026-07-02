import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { ProductPlaceholder } from '../shared/ui/placeholder';
import { formatPrice, formatRating, getImageUrl } from '../../lib/format/utils';
import { useCart } from '../cart/cart-provider';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

interface ProductCardProps {
  product: InternalCatalogProductView;
  variant?: 'default' | 'minimal' | 'bordered';
}

export function ProductCard({ product, variant = 'default' }: ProductCardProps) {
  const { addItem } = useCart();
  const navigate = useNavigate();
  const {
    id,
    name = 'Product',
    price_ghs_minor: price = 0,
    was_price_ghs_minor: wasPrice,
    rating,
    review_count: reviewCount,
    image_path: imagePath,
    tags,
  } = product;

  const hasDiscount = wasPrice && wasPrice > price;
  const displayPrice = formatPrice(price);
  const displayWasPrice = wasPrice ? formatPrice(wasPrice) : null;
  const displayRating = formatRating(rating, reviewCount);
  const imageUrl = getImageUrl(imagePath);
  const tone = (imagePath?.includes('tone=') ? imagePath.split('tone=')[1]?.split('&')[0] : 'lavender') as 'lavender' | 'cream' | 'ink' | 'rose';

  const handleAddToCart = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!id) return;
    try {
      await addItem(id, 1, name);
    } catch (error) {
      console.error('Failed to add to cart:', error);
    }
  };

  const cardStyles = {
    default: 'bg-paper border border-line-soft hover:border-lavender-300 transition-colors',
    minimal: 'bg-transparent border-0',
    bordered: 'bg-paper border-2 border-line hover:border-lavender-400',
  };

  return (
    <div
      onClick={() => id && navigate({ to: '/shop/$slug', params: { slug: id } })}
      className={`block ${cardStyles[variant]} rounded overflow-hidden group hover:-translate-y-px transition-transform duration-[600ms] ease-[cubic-bezier(0.2,0.8,0.2,1)]`}
    >
      <div className="relative aspect-[4/5] overflow-hidden bg-lavender-50">
        {imageUrl ? (
          <img
            src={imageUrl}
            alt={name || 'Product'}
            className="w-full h-full object-cover group-hover:scale-[1.03] transition-transform duration-[600ms] ease-[cubic-bezier(0.2,0.8,0.2,1)]"
            loading="lazy"
          />
        ) : (
          <ProductPlaceholder tone={tone} label={name?.substring(0, 2)} />
        )}

        {tags && tags.length > 0 && (
          <div className="absolute top-2 left-2 flex gap-1">
            {tags.map((tag) => (
              <span
                key={tag}
                className="px-2 py-1 bg-paper/90 backdrop-blur-sm text-xs font-label font-medium rounded-full"
                style={{ fontSize: '10px' }}
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {hasDiscount && (
          <span className="absolute top-2 right-2 px-2 py-1 bg-ink text-paper text-xs font-label font-medium rounded-full">
            Sale
          </span>
        )}

        {/* Wishlist heart button - appears on hover */}
        <button
          onClick={(e) => {
            e.stopPropagation();
            // Wishlist functionality to be implemented
          }}
          className="absolute top-2 right-2 w-8 h-8 rounded-full bg-paper/90 backdrop-blur-sm flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-[280ms] hover:bg-paper"
          aria-label="Add to wishlist"
        >
          <Icon name="heart" size={16} />
        </button>
      </div>

      <div className="p-4">
        <div className="flex items-start justify-between gap-2 mb-2">
          <div className="flex-1 min-w-0">
            <h3 className="font-label font-medium text-ink text-sm mb-1 truncate">{name ?? 'Product'}</h3>
            {displayRating && (
              <div className="flex items-center gap-1 text-xs">
                <Icon name="star" size={12} />
                <span className="text-ink-muted">{displayRating}</span>
              </div>
            )}
          </div>
        </div>

        <div className="flex items-end justify-between">
          <div className="flex flex-col">
            <span className="font-label font-semibold text-ink">{displayPrice}</span>
            {hasDiscount && (
              <span className="text-xs text-ink-muted line-through">{displayWasPrice}</span>
            )}
          </div>

          <button
            onClick={handleAddToCart}
            className="p-2 rounded-full bg-lavender-100 hover:bg-lavender-200 text-ink transition-colors"
            aria-label="Add to cart"
          >
            <Icon name="plus" size={16} />
          </button>
        </div>
      </div>
    </div>
  );
}
