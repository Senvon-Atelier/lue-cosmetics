import { useState, useEffect } from 'react';
import { useCart } from '../cart/cart-provider';
import { Icon } from '../shared/ui/icons';
import { Button } from '../shared/ui/button';
import { ProductPlaceholder } from '../shared/ui/placeholder';
import { formatPrice, formatRating, getImageUrl } from '../../lib/format/utils';
import { getProductsSlug } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

interface ProductDetailProps {
  slug: string;
}

export function ProductDetail({ slug }: ProductDetailProps) {
  const [product, setProduct] = useState<InternalCatalogProductView | null>(null);
  const [loading, setLoading] = useState(true);
  const [quantity, setQuantity] = useState(1);
  const { addItem } = useCart();

  useEffect(() => {
    const loadProduct = async () => {
      if (!slug) return;
      setLoading(true);
      try {
        const response = await getProductsSlug(slug);
        setProduct(response.data);
      } catch (error) {
        console.error('Failed to load product:', error);
      } finally {
        setLoading(false);
      }
    };
    loadProduct();
  }, [slug]);

  const handleAddToCart = async () => {
    if (!product || !product.id) return;
    try {
      await addItem(product.id, quantity);
    } catch (error) {
      console.error('Failed to add to cart:', error);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-ink-muted">Loading...</div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center">
        <div className="text-center">
          <h1 className="font-display text-4xl mb-4">Product Not Found</h1>
          <p className="text-ink-muted">The product you're looking for doesn't exist.</p>
        </div>
      </div>
    );
  }

  const {
    name = 'Product',
    price_ghs_minor: price = 0,
    was_price_ghs_minor: wasPrice,
    rating,
    review_count: reviewCount,
    image_path: imagePath,
    tags,
    size,
  } = product;

  const hasDiscount = wasPrice && wasPrice > price;
  const displayPrice = formatPrice(price);
  const displayWasPrice = wasPrice ? formatPrice(wasPrice) : null;
  const displayRating = formatRating(rating, reviewCount);
  const imageUrl = getImageUrl(imagePath);
  const tone = (imagePath?.includes('tone=') ? imagePath.split('tone=')[1]?.split('&')[0] : 'lavender') as 'lavender' | 'cream' | 'ink' | 'rose';

  return (
    <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
      <div className="grid md:grid-cols-2 gap-8 lg:gap-12">
        {/* Product Images */}
        <div className="space-y-4">
          <div className="aspect-square rounded-lg overflow-hidden bg-lavender-50">
            {imageUrl ? (
              <img
                src={imageUrl}
                alt={name || 'Product'}
                className="w-full h-full object-cover"
                loading="eager"
              />
            ) : (
              <ProductPlaceholder tone={tone} label={name?.substring(0, 2)} />
            )}
          </div>
        </div>

        {/* Product Info */}
        <div className="space-y-6">
          <div>
            {tags && tags.length > 0 && (
              <div className="flex gap-2 mb-3">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="px-2 py-1 bg-lavender-100 text-ink text-xs font-label font-medium rounded"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            )}

            <h1 className="font-display text-4xl mb-2">{name ?? 'Product'}</h1>

            {displayRating && (
              <div className="flex items-center gap-2 mb-4">
                <div className="flex items-center gap-1">
                  <Icon name="star" size={16} />
                  <span className="font-label font-semibold">{displayRating}</span>
                </div>
              </div>
            )}

            <div className="flex items-baseline gap-3">
              <span className="font-display text-3xl font-semibold">{displayPrice}</span>
              {hasDiscount && (
                <span className="text-lg text-ink-muted line-through">{displayWasPrice}</span>
              )}
            </div>

            {size && (
              <p className="text-sm text-ink-muted">Size: {size}</p>
            )}
          </div>

          {/* Quantity Selector */}
          <div>
            <label className="font-label font-semibold mb-2 block">Quantity</label>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setQuantity(Math.max(1, quantity - 1))}
                className="w-10 h-10 flex items-center justify-center border border-line rounded-lg hover:bg-lavender-50 transition-colors"
                disabled={quantity <= 1}
              >
                <Icon name="minus" size={16} />
              </button>
              <span className="w-12 text-center font-label font-semibold">{quantity}</span>
              <button
                onClick={() => setQuantity(quantity + 1)}
                className="w-10 h-10 flex items-center justify-center border border-line rounded-lg hover:bg-lavender-50 transition-colors"
              >
                <Icon name="plus" size={16} />
              </button>
            </div>
          </div>

          {/* Add to Cart Button */}
          <Button onClick={handleAddToCart} className="w-full" icon="arrow" iconPosition="right">
            Add to Cart
          </Button>

          {/* Product Features */}
          <div className="space-y-3 pt-6 border-t border-line-soft">
            <div className="flex items-start gap-3">
              <Icon name="truck" size={20} className="text-lavender-600 flex-shrink-0" />
              <div>
                <p className="font-label font-semibold">Free delivery</p>
                <p className="text-sm text-ink-muted">On orders over GHS 250 in Accra</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Icon name="shield" size={20} className="text-lavender-600 flex-shrink-0" />
              <div>
                <p className="font-label font-semibold">Authentic products</p>
                <p className="text-sm text-ink-muted">Sourced directly from brands</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Icon name="leaf" size={20} className="text-lavender-600 flex-shrink-0" />
              <div>
                <p className="font-label font-semibold">Clean ingredients</p>
                <p className="text-sm text-ink-muted">Carefully curated formulations</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
