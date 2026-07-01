import { ProductCard } from './product-card';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

interface ProductGridProps {
  products: InternalCatalogProductView[];
  loading?: boolean;
}

export function ProductGrid({ products, loading }: ProductGridProps) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 mt-8">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="aspect-[4/5] bg-lavender-100 rounded animate-pulse" />
        ))}
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="text-center py-16 mt-8">
        <div className="inline-flex flex-col items-center justify-center w-20 h-20 bg-lavender-50 rounded-full mb-4">
          <span className="text-4xl">🔍</span>
        </div>
        <h3 className="font-display text-2xl mb-2">No products found</h3>
        <p className="text-ink-muted">Try adjusting your filters or search terms</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 mt-8">
      {products.map((product) => (
        <ProductCard key={product.id} product={product} />
      ))}
    </div>
  );
}
