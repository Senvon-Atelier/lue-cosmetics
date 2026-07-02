import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getProducts } from '../../lib/api/generated/rueCosmeticsAPI';
import { ProductCard } from '../catalog/product-card';
import { Button, Icon } from '../shared/ui';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

export function FeaturedProducts() {
  const navigate = useNavigate();
  const [products, setProducts] = useState<InternalCatalogProductView[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadProducts = async () => {
      try {
        const response = await getProducts({
          limit: 8,
        });
        const items = response.data?.items;
        const productsArray = Array.isArray(items) ? items : [];
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
    <section className="section bg-lavender-50">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Featured</div>
            <h2 className="font-display text-[clamp(22px,2.5vw,32px)] font-normal tracking-[-0.01em]">
              Featured Products
            </h2>
          </div>
          <button
            onClick={() => navigate({ to: '/shop' })}
            className="section-link"
          >
            <span>View all</span>
            <Icon name="arrow" size={16} />
          </button>
        </div>

        {isLoading ? (
          <div className="grid-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="aspect-[4/5] bg-lavender-100 rounded mb-4" />
                <div className="h-4 bg-lavender-100 rounded w-3/4 mb-2" />
                <div className="h-3 bg-lavender-100 rounded w-1/2" />
              </div>
            ))}
          </div>
        ) : products.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-ink-muted">No featured products available.</p>
          </div>
        ) : (
          <>
            <div className="grid-4">
              {products.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>

            <div className="text-center mt-12">
              <Button onClick={() => navigate({ to: '/shop' })} variant="outline" icon="arrow" iconPosition="right">
                Shop All Products
              </Button>
            </div>
          </>
        )}
      </div>
    </section>
  );
}
