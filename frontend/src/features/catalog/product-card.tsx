import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';
import { formatPrice, formatRating, getImageUrl } from '../../lib/format/utils';
import { useCart } from '../cart/cart-provider';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

interface ProductCardProps {
  product: InternalCatalogProductView;
  variant?: 'default' | 'minimal' | 'bordered' | 'list';
}

export function ProductCard({ product, variant = 'default' }: ProductCardProps) {
  const [hover, setHover] = useState(false);
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

  const handleOpen = () => {
    if (id) navigate({ to: '/shop/$slug', params: { slug: id } });
  };

  return (
    <article
      className={`pcard pcard-${variant}`}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
    >
      <div className="pcard-media" onClick={handleOpen}>
        {imageUrl ? (
          <img
            src={imageUrl}
            alt={name || 'Product'}
            className="w-full h-full object-cover"
            style={{ aspectRatio: '4/5', display: 'block', width: '100%' }}
            loading="lazy"
          />
        ) : (
          <div className="ph" style={{ aspectRatio: '4/5' }}>
            <span className="ph-label">{name?.substring(0, 2) || 'P'}</span>
          </div>
        )}

        {tags && tags.length > 0 && (
          <span className="pcard-tag">{tags[0]}</span>
        )}

        <button
          className="pcard-wish"
          onClick={(e) => { e.stopPropagation(); }}
          aria-label="Wishlist"
        >
          <Icon name="heart" size={16} />
        </button>

        <button
          className={`pcard-add ${hover ? 'show' : ''}`}
          onClick={handleAddToCart}
        >
          Add to bag <Icon name="plus" size={14} />
        </button>
      </div>

      <div className="pcard-body">
        <div className="pcard-name" onClick={handleOpen}>{name}</div>
        <div className="pcard-foot">
          <div className="pcard-price">
            <span className="price">{displayPrice}</span>
            {hasDiscount && <span className="price-was">{displayWasPrice}</span>}
          </div>
          {displayRating && (
            <div className="pcard-rating">
              <Icon name="starFilled" size={11} />
              <span>{displayRating}</span>
            </div>
          )}
        </div>
      </div>
    </article>
  );
}
