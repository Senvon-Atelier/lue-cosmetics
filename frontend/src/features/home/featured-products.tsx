import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getProducts } from '../../lib/api/generated/rueCosmeticsAPI';
import { ProductCard } from '../catalog/product-card';
import { Icon } from '../shared/ui';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

export function FeaturedProducts() {
  const navigate = useNavigate();
  const [products, setProducts] = useState<InternalCatalogProductView[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadProducts = async () => {
      try {
        const response = await getProducts({ limit: 8 });
        const productsArray = Array.isArray(response?.items) ? response.items : [];
        setProducts(productsArray);
      } catch (error) {
        console.error('Failed to load featured products:', error);
        setProducts([]);
      } finally {
        setIsLoading(false);
      }
    };
    loadProducts();
  }, []);

  return (
    <section className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Most loved this month</div>
            <h2 className="h-display" style={{ fontSize: 'clamp(32px, 4vw, 56px)' }}>What Accra is <em>reaching for.</em></h2>
          </div>
          <button onClick={() => navigate({ to: '/shop' })} className="section-link">
            Shop all <Icon name="arrow" size={14} />
          </button>
        </div>

        {isLoading ? (
          <div className="grid-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="pcard">
                <div className="pcard-media">
                  <div className="ph" style={{ aspectRatio: '4/5' }}>
                    <span className="ph-label">loading</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : products.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-ink-muted">No featured products available.</p>
          </div>
        ) : (
          <div className="grid-4">
            {products.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
