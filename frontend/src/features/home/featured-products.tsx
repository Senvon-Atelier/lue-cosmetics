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
    <section className="py-16 bg-lavender-50">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
        <div className="flex items-center justify-between mb-8">
          <div>
            <h2 className="font-display text-3xl mb-2">Featured Products</h2>
            <p className="text-ink-muted">Our most loved items, curated for you</p>
          </div>
          <button
            onClick={() => navigate({ to: '/shop' })}
            className="flex items-center gap-2 font-label font-medium text-sm text-lavender-600 hover:text-lavender-700 transition-colors"
          >
            View all
            <Icon name="arrow" size={14} />
          </button>
        </div>

        {isLoading ? (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="aspect-[4/5] bg-lavender-100 rounded-lg mb-4" />
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
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-8">
              {products.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>

            <div className="text-center">
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
