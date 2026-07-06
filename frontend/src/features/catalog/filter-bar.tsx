import { Icon } from '../shared/ui/icons';

const CONCERNS = ['Hydration', 'Brightening', 'Anti-aging', 'Sensitive skin', 'Acne-prone', 'Damaged hair'];

interface FilterBarProps {
  brands: Array<{ id: string; name: string; slug: string }>;
  selectedBrand: string | null;
  searchQuery: string;
  showFilters: boolean;
  onBrandChange: (slug: string | null) => void;
  onSearchChange: (query: string) => void;
  onCloseFilters: () => void;
  onClear: () => void;
}

export function FilterBar({
  brands,
  selectedBrand,
  searchQuery,
  showFilters,
  onBrandChange,
  onSearchChange,
  onCloseFilters,
  onClear,
}: FilterBarProps) {
  const hasActiveFilters = !!selectedBrand;

  return (
    <aside className={`shop-filters ${showFilters ? 'open' : ''}`}>
      <div className="shop-filters-head">
        <div>
          <div className="eyebrow">Filters</div>
          <h3 className="h-display" style={{ fontSize: 28, margin: 0 }}>Refine</h3>
        </div>
        <button className="icon-btn mobile-close-filters" onClick={onCloseFilters}>
          <Icon name="close" />
        </button>
      </div>

      {/* Search */}
      <div className="filter-group" style={{ paddingTop: 0 }}>
        <div className="label">Search</div>
        <div style={{ position: 'relative' }}>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search products..."
            style={{
              width: '100%',
              padding: '10px 12px',
              border: '1px solid var(--line)',
              borderRadius: 8,
              fontFamily: 'var(--font-body)',
              fontSize: 14,
              background: 'transparent',
              color: 'var(--ink)',
              outline: 'none',
              boxSizing: 'border-box',
            }}
          />
          {searchQuery && (
            <button
              onClick={() => onSearchChange('')}
              style={{
                position: 'absolute', right: 8, top: '50%',
                transform: 'translateY(-50%)',
                background: 'none', border: 'none',
                cursor: 'pointer', color: 'var(--ink-muted)',
                padding: 4,
              }}
              aria-label="Clear search"
            >
              <Icon name="close" size={14} />
            </button>
          )}
        </div>
      </div>

      {/* Price */}
      <div className="filter-group">
        <div className="label">Price · up to GHS 500</div>
        <input type="range" min="50" max="700" step="10" defaultValue={700} className="price-slider" disabled />
        <div className="price-range-labels"><span>GHS 50</span><span>GHS 700</span></div>
      </div>

      {/* Brands */}
      {brands.length > 0 && (
        <div className="filter-group">
          <div className="label">Brand</div>
          <div className="brand-list">
            {brands.map((b) => (
              <label key={b.id} className="brand-check">
                <input
                  type="checkbox"
                  checked={selectedBrand === b.id}
                  onChange={() => onBrandChange(selectedBrand === b.id ? null : b.id)}
                />
                <span className="check-box"><Icon name="check" size={12} /></span>
                <span>{b.name}</span>
              </label>
            ))}
          </div>
        </div>
      )}

      {/* Concern */}
      <div className="filter-group">
        <div className="label">Concern</div>
        <div className="tag-list">
          {CONCERNS.map((t) => (
            <button key={t} className="chip">{t}</button>
          ))}
        </div>
      </div>

      {hasActiveFilters && (
        <button className="btn btn-ghost" style={{ width: '100%', justifyContent: 'center' }} onClick={onClear}>
          Clear all filters
        </button>
      )}
    </aside>
  );
}
