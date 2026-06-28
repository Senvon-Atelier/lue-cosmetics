import { useState } from 'react';
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
  const [showFilters, setShowFilters] = useState(false);

  return (
    <div className="border-b border-line-soft py-4">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
        <div className="flex items-center justify-between gap-4 mb-4">
          <div className="flex items-center gap-2 flex-1">
            <div className="relative flex-1 max-w-md">
              <Icon name="search" size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-muted" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => onSearchChange(e.target.value)}
                placeholder="Search products, brands..."
                className="w-full pl-10 pr-4 py-2 border border-line rounded-lg bg-paper text-ink placeholder:text-ink-muted focus:outline-none focus:ring-2 focus:ring-lavender-400 focus:border-transparent"
              />
              {searchQuery && (
                <button
                  onClick={() => onSearchChange('')}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-ink-muted hover:text-ink"
                  aria-label="Clear search"
                >
                  <Icon name="close" size={14} />
                </button>
              )}
            </div>
          </div>

          <button
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2 px-4 py-2 border border-line rounded-lg bg-paper hover:bg-lavender-50 transition-colors"
          >
            <Icon name="filter" size={16} />
            <span className="font-label text-sm">Filters</span>
            {(selectedCategory || selectedBrand) && (
              <span className="w-2 h-2 bg-lavender-600 rounded-full" />
            )}
          </button>
        </div>

        {showFilters && (
          <div className="space-y-4 mt-4">
            <div>
              <h3 className="font-label font-semibold text-sm mb-2">Categories</h3>
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => onCategoryChange(null)}
                  className={`px-3 py-1.5 text-sm rounded-lg transition-colors ${
                    selectedCategory === null
                      ? 'bg-lavender-600 text-paper'
                      : 'bg-lavender-100 text-ink hover:bg-lavender-200'
                  }`}
                >
                  All
                </button>
                {categories.map((category) => (
                  <button
                    key={category.id}
                    onClick={() => onCategoryChange(category.slug)}
                    className={`px-3 py-1.5 text-sm rounded-lg transition-colors ${
                      selectedCategory === category.slug
                        ? 'bg-lavender-600 text-paper'
                        : 'bg-lavender-100 text-ink hover:bg-lavender-200'
                    }`}
                  >
                    {category.label}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <h3 className="font-label font-semibold text-sm mb-2">Brands</h3>
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => onBrandChange(null)}
                  className={`px-3 py-1.5 text-sm rounded-lg transition-colors ${
                    selectedBrand === null
                      ? 'bg-lavender-600 text-paper'
                      : 'bg-lavender-100 text-ink hover:bg-lavender-200'
                  }`}
                >
                  All
                </button>
                {brands.map((brand) => (
                  <button
                    key={brand.id}
                    onClick={() => onBrandChange(brand.slug)}
                    className={`px-3 py-1.5 text-sm rounded-lg transition-colors ${
                      selectedBrand === brand.slug
                        ? 'bg-lavender-600 text-paper'
                        : 'bg-lavender-100 text-ink hover:bg-lavender-200'
                    }`}
                  >
                    {brand.name}
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
