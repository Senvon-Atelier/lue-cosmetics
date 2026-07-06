import { ProductCard } from './product-card';
import { Icon } from '../shared/ui/icons';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

interface ProductGridProps {
  products: InternalCatalogProductView[];
  loading?: boolean;
  view?: 'grid' | 'list';
  onReset?: () => void;
}

export function ProductGrid({ products, loading, view = 'grid', onReset }: ProductGridProps) {
  if (loading) {
    return (
      <div className={view === 'grid' ? 'grid-4 shop-grid' : 'shop-list'} style={{ marginTop: 0 }}>
        {Array.from({ length: 16 }).map((_, i) => (
          <div key={i} className="pcard">
            <div className="pcard-media">
              <div className="ph" style={{ aspectRatio: '4/5' }}>
                <span className="ph-label">loading</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="shop-empty">
        <div className="shop-empty-icon"><Icon name="search" size={40} /></div>
        <p>We couldn't find any matches. Try adjusting your filters or browsing our full collection.</p>
        {onReset && (
          <button className="btn btn-primary" onClick={onReset}>
            Clear All Filters
          </button>
        )}
      </div>
    );
  }

  const gridClass = view === 'grid' ? 'grid-4 shop-grid' : 'shop-list';

  return (
    <div className={gridClass} style={{ marginTop: 0 }}>
      {products.map((product) => (
        <ProductCard key={product.id} product={product} variant={view === 'list' ? 'list' : 'default'} />
      ))}
    </div>
  );
}
