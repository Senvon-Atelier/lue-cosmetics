import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useCart } from '../cart/cart-provider';
import { Icon } from '../shared/ui/icons';
import { ProductPlaceholder } from '../shared/ui/placeholder';
import { formatGhs, getImageUrl } from '../../lib/format/utils';
import {
  useGetProductsSlug,
  useGetBrands,
  useGetCategories,
  useGetProducts,
} from '../../lib/api/generated/rueCosmeticsAPI';
import { ProductCard } from './product-card';
import { getProductCopy, PERKS } from '../../content/product-copy';

interface ProductDetailProps {
  slug: string;
}

type TabKey = 'desc' | 'how' | 'ings';

export function ProductDetail({ slug }: ProductDetailProps) {
  const [qty, setQty] = useState(1);
  const [tab, setTab] = useState<TabKey>('desc');

  const { addItem } = useCart();

  const { data: product, isLoading, error } = useGetProductsSlug(slug);
  const { data: categories } = useGetCategories();
  const { data: brands } = useGetBrands();

  const category = categories?.find(c => c.id === product?.category_id);
  const categorySlug = category?.slug;
  const categoryLabel = category?.label;
  const brand = brands?.find(b => b.id === product?.brand_id);
  const brandName = brand?.name ?? '';

  const { data: relatedData } = useGetProducts(
    { category: categorySlug, limit: 5 },
    { query: { enabled: !!categorySlug } },
  );
  const related = (relatedData?.items ?? [])
    .filter(p => p.slug !== slug)
    .slice(0, 4);

  if (isLoading) {
    return (
      <main className="fade-up">
        <div className="wrap" style={{ padding: '80px 0', textAlign: 'center' }}>
          <p className="eyebrow" style={{ color: 'var(--ink-muted)' }}>Loading product…</p>
        </div>
      </main>
    );
  }

  if (error || !product) {
    return (
      <main className="fade-up">
        <div className="wrap" style={{ padding: '80px 0', textAlign: 'center' }}>
          <p className="eyebrow" style={{ color: 'var(--ink-muted)' }}>Product not found.</p>
        </div>
      </main>
    );
  }

  const {
    name = 'Product',
    price_ghs_minor: priceMinor = 0,
    was_price_ghs_minor: wasPriceMinor,
    rating,
    review_count: reviewCount,
    image_path: imagePath,
    size,
    tone,
  } = product;

  const copy = getProductCopy(categorySlug);
  const safeTone = (
    tone === 'lavender' || tone === 'cream' || tone === 'ink' || tone === 'rose'
      ? tone
      : 'lavender'
  ) as 'lavender' | 'cream' | 'ink' | 'rose';

  // Shop category filter is client-state-only (not URL-driven) — all breadcrumb
  // and lede category links go to /shop without search params. See task-3-report.
  return (
    <main className="fade-up">
      <nav className="wrap breadcrumb">
        <Link to="/">Home</Link>
        <Icon name="chevronRight" size={12} />
        <Link to="/shop">Shop</Link>
        <Icon name="chevronRight" size={12} />
        <Link to="/shop">{categoryLabel ?? 'Category'}</Link>
        <Icon name="chevronRight" size={12} />
        <span>{name}</span>
      </nav>

      <section className="wrap product-grid">
        {/* Gallery — single image, no thumb rail per spec §4.1 */}
        <div className="product-main">
          {imagePath ? (
            <img
              src={getImageUrl(imagePath)}
              alt={name}
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
            />
          ) : (
            <ProductPlaceholder tone={safeTone} label={brandName} />
          )}
        </div>

        <div className="product-info">
          <div className="eyebrow">{brandName}</div>
          <h1 className="h-display product-name">{name}</h1>

          {typeof rating === 'number' && (
            <div className="product-rating">
              <div className="stars">
                {[0, 1, 2, 3, 4].map(i => (
                  <Icon key={i} name="starFilled" size={14} />
                ))}
              </div>
              <span>{rating} · {reviewCount ?? 0} reviews</span>
            </div>
          )}

          <div className="product-price">
            <span className="price">{formatGhs(priceMinor)}</span>
            {typeof wasPriceMinor === 'number' && wasPriceMinor > 0 && (
              <span className="price-was">{formatGhs(wasPriceMinor)}</span>
            )}
            {size && <span className="product-size">· {size}</span>}
          </div>

          <p className="product-lede">
            {copy.lede}{' '}Part of our{' '}
            <Link to="/shop">
              {categoryLabel?.toLowerCase() ?? 'beauty'}
            </Link>{' '}
            edit.
          </p>

          <div className="product-actions">
            <div className="qty qty-lg">
              <button onClick={() => setQty(Math.max(1, qty - 1))}>
                <Icon name="minus" size={14} />
              </button>
              <span>{qty}</span>
              <button onClick={() => setQty(qty + 1)}>
                <Icon name="plus" size={14} />
              </button>
            </div>
            <button
              className="btn btn-primary"
              style={{ flex: 1, justifyContent: 'center' }}
              onClick={async () => {
                if (!product.id) return;
                try {
                  await addItem(product.id, qty, product.name);
                } catch (error) {
                  console.error('Failed to add to cart:', error);
                }
              }}
            >
              Add to bag · {formatGhs(priceMinor * qty)}
            </button>
            <button
              className="icon-btn icon-btn-lg"
              disabled
              title="Saved items coming soon"
            >
              <Icon name="heart" />
            </button>
          </div>

          <ul className="product-perks">
            {PERKS.map((perk, i) => (
              <li key={i}>
                <Icon
                  name={i === 0 ? 'truck' : i === 1 ? 'shield' : 'whatsapp'}
                  size={14}
                />
                {perk}
              </li>
            ))}
          </ul>

          <div className="product-tabs">
            <div className="tab-heads">
              <button
                className={`tab-head${tab === 'desc' ? ' active' : ''}`}
                onClick={() => setTab('desc')}
              >
                Description
              </button>
              {copy.howTo && (
                <button
                  className={`tab-head${tab === 'how' ? ' active' : ''}`}
                  onClick={() => setTab('how')}
                >
                  How to use
                </button>
              )}
              {copy.ingredients && (
                <button
                  className={`tab-head${tab === 'ings' ? ' active' : ''}`}
                  onClick={() => setTab('ings')}
                >
                  Ingredients
                </button>
              )}
            </div>
            <div className="tab-body">
              {tab === 'desc' && <p>{copy.description}</p>}
              {tab === 'how' && copy.howTo && <p>{copy.howTo}</p>}
              {tab === 'ings' && copy.ingredients && (
                <p className="ing-list">{copy.ingredients}</p>
              )}
            </div>
          </div>
        </div>
      </section>

      {related.length > 0 && (
        <section className="section">
          <div className="wrap">
            <div className="section-head">
              <h2 className="h-display" style={{ fontSize: 'clamp(28px, 3.5vw, 48px)' }}>
                You might also <em>love</em>.
              </h2>
            </div>
            <div className="grid-4">
              {related.map(r => (
                <ProductCard key={r.id} product={r} variant="default" />
              ))}
            </div>
          </div>
        </section>
      )}
    </main>
  );
}
