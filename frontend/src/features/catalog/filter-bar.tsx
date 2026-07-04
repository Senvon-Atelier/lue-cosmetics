import { Icon } from '../shared/ui/icons';

interface FilterBarProps {
  categories: Array<{ id: string; label: string; slug: string }>;
  brands: Array<{ id: string; name: string; slug: string }>;
  selectedCategory: string | null;
  selectedBrand: string | null;
  searchQuery: string;
  onCategoryChange: (slug: string | null) => void;
  onBrandChange: (slug: string | null) => void;
  onSearchChange: (query: string) => void;
}

export function FilterBar({
  categories,
  brands,
  selectedCategory,
  selectedBrand,
  searchQuery,
  onCategoryChange,
  onBrandChange,
  onSearchChange,
}: FilterBarProps) {
  return (
    <div className="space-y-6">
      {/* Search */}
      <div className="relative">
        <Icon name="search" size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-muted" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search products..."
          className="w-full pl-10 pr-4 py-3 border border-line rounded bg-paper text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-600 focus:ring-2 focus:ring-lavender-600 transition-all duration-[var(--dur)]"
        />
        {searchQuery && (
          <button
            onClick={() => onSearchChange('')}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-ink-muted hover:text-ink transition-colors"
            aria-label="Clear search"
          >
            <Icon name="close" size={14} />
          </button>
        )}
      </div>

      {/* Filter Groups */}
      {categories.length > 0 && (
        <div className="border border-line p-6 rounded">
          <h3 className="font-label font-semibold text-xs uppercase tracking-wider mb-4 text-ink-soft">Categories</h3>
          <div className="space-y-2">
            <button
              onClick={() => onCategoryChange(null)}
              className={`w-full text-left px-3 py-2 text-sm rounded-full transition-colors duration-[var(--dur)] ${
                selectedCategory === null
                  ? 'bg-lavender-600 text-paper'
                  : 'bg-lavender-50 text-ink hover:bg-lavender-100'
              }`}
            >
              All Categories
            </button>
            {categories.map((category) => (
              <button
                key={category.id}
                onClick={() => onCategoryChange(category.slug)}
                className={`w-full text-left px-3 py-2 text-sm rounded-full transition-colors duration-[var(--dur)] ${
                  selectedCategory === category.slug
                    ? 'bg-lavender-600 text-paper'
                    : 'bg-lavender-50 text-ink hover:bg-lavender-100'
                }`}
              >
                {category.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {brands.length > 0 && (
        <div className="border border-line p-6 rounded">
          <h3 className="font-label font-semibold text-xs uppercase tracking-wider mb-4 text-ink-soft">Brands</h3>
          <div className="space-y-2">
            <button
              onClick={() => onBrandChange(null)}
              className={`w-full text-left px-3 py-2 text-sm rounded-full transition-colors duration-[var(--dur)] ${
                selectedBrand === null
                  ? 'bg-lavender-600 text-paper'
                  : 'bg-lavender-50 text-ink hover:bg-lavender-100'
              }`}
            >
              All Brands
            </button>
            {brands.map((brand) => (
              <button
                key={brand.id}
                onClick={() => onBrandChange(brand.id)}
                className={`w-full text-left px-3 py-2 text-sm rounded-full transition-colors duration-[var(--dur)] ${
                  selectedBrand === brand.id
                    ? 'bg-lavender-600 text-paper'
                    : 'bg-lavender-50 text-ink hover:bg-lavender-100'
                }`}
              >
                {brand.name}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Clear All */}
      {(selectedCategory || selectedBrand) && (
        <button
          onClick={() => {
            onCategoryChange(null);
            onBrandChange(null);
          }}
          className="font-label text-sm text-ink-soft hover:text-ink transition-colors"
        >
          Clear all filters
        </button>
      )}
    </div>
  );
}
