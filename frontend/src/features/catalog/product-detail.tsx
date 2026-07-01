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
    <div className="section">
      <div className="wrap">
        {/* Breadcrumb Navigation */}
        <nav className="font-label text-xs text-ink-muted mb-6 flex items-center gap-2">
          <a href="/" className="hover:text-ink transition-colors">Home</a>
          <Icon name="chevronRight" size={12} />
          <a href="/shop" className="hover:text-ink transition-colors">Shop</a>
          <Icon name="chevronRight" size={12} />
          <span className="text-ink">{name ?? 'Product'}</span>
        </nav>

        <div className="grid grid-cols-1 lg:grid-cols-[1.1fr_1fr] gap-12">
          {/* Product Images */}
          <div className="space-y-4">
            <div className="aspect-[4/5] rounded overflow-hidden bg-lavender-50 group">
              {imageUrl ? (
                <img
                  src={imageUrl}
                  alt={name || 'Product'}
                  className="w-full h-full object-cover group-hover:scale-[1.02] transition-transform duration-[280ms] ease-[cubic-bezier(0.2,0.8,0.2,1)]"
                  loading="eager"
                />
              ) : (
                <ProductPlaceholder tone={tone} label={name?.substring(0, 2)} />
              )}
            </div>
          </div>

          {/* Product Info */}
          <div className="space-y-6 lg:sticky lg:top-24 lg:self-start">
            <div>
              {tags && tags.length > 0 && (
                <div className="flex gap-2 mb-3">
                  {tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-2 py-1 bg-paper/90 backdrop-blur-sm text-ink text-xs font-label font-medium rounded-full"
                      style={{ fontSize: '10px' }}
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}

              <h1 className="font-display text-[clamp(32px,4vw,56px)] font-normal leading-[1.1] tracking-[-0.01em] mb-3">
                {name ?? 'Product'}
              </h1>

              {displayRating && (
                <div className="flex items-center gap-2 mb-4">
                  <div className="flex items-center gap-1">
                    <Icon name="starFilled" size={16} className="text-lavender-600" />
                    <span className="font-label font-semibold text-sm">{displayRating}</span>
                  </div>
                </div>
              )}

              <div className="flex items-baseline gap-3 mb-4">
                <span className="font-label font-semibold text-xl">{displayPrice}</span>
                {hasDiscount && (
                  <span className="text-sm text-ink-muted line-through">{displayWasPrice}</span>
                )}
              </div>

              {size && (
                <p className="text-sm text-ink-muted">Size: {size}</p>
              )}
            </div>

            {/* Perks List */}
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm">
                <Icon name="check" size={16} className="text-lavender-600" />
                <span>In stock, ready to ship</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <Icon name="truck" size={16} className="text-lavender-600" />
                <span>Free delivery on orders over GHS 250</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <Icon name="shield" size={16} className="text-lavender-600" />
                <span>100% authentic, sourced directly</span>
              </div>
            </div>

            {/* Quantity Selector */}
            <div>
              <label className="font-label font-semibold text-xs uppercase tracking-wider mb-2 block">Quantity</label>
              <div className="flex items-center gap-3">
                <button
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  className="w-10 h-10 flex items-center justify-center border border-line rounded-full hover:bg-lavender-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={quantity <= 1}
                >
                  <Icon name="minus" size={16} />
                </button>
                <span className="w-12 text-center font-label font-semibold">{quantity}</span>
                <button
                  onClick={() => setQuantity(quantity + 1)}
                  className="w-10 h-10 flex items-center justify-center border border-line rounded-full hover:bg-lavender-50 transition-colors"
                >
                  <Icon name="plus" size={16} />
                </button>
              </div>
            </div>

            {/* Add to Cart Button */}
            <Button onClick={handleAddToCart} className="w-full" size="lg" icon="arrowRight" iconPosition="right">
              Add to Cart
            </Button>

            {/* Product Features - Tabbed Content */}
            <div className="pt-6 border-t border-line-soft">
              <div className="flex gap-6 border-b border-line mb-4">
                <button className="pb-2 font-label font-semibold text-xs uppercase tracking-wider text-ink border-b-2 border-lavender-600">
                  Description
                </button>
                <button className="pb-2 font-label font-semibold text-xs uppercase tracking-wider text-ink-muted hover:text-ink transition-colors">
                  How to Use
                </button>
                <button className="pb-2 font-label font-semibold text-xs uppercase tracking-wider text-ink-muted hover:text-ink transition-colors">
                  Ingredients
                </button>
              </div>
              <div className="text-sm text-ink-soft leading-relaxed">
                <p>A premium skincare product carefully formulated to deliver visible results. Perfect for daily use as part of your skincare routine.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
