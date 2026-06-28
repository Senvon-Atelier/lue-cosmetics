import { ProductCard } from './product-card';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

interface ProductGridProps {
  products: InternalCatalogProductView[];
  loading?: boolean;
}

export function ProductGrid({ products, loading }: ProductGridProps) {
  if (loading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4" style={{ padding: '0 2rem 2rem' }}>
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="aspect-[4/5] bg-lavender-100 rounded-lg animate-pulse" />
        ))}
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="text-center py-16" style={{ padding: '0 2rem' }}>
        <div className="inline-flex flex-col items-center justify-center w-20 h-20 bg-lavender-50 rounded-full mb-4">
          <span className="text-4xl">🔍</span>
        </div>
        <h3 className="font-display text-2xl mb-2">No products found</h3>
        <p className="text-ink-muted">Try adjusting your filters or search terms</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4" style={{ padding: '2rem' }}>
      {products.map((product) => (
        <ProductCard key={product.id} product={product} />
      ))}
    </div>
  );
}
