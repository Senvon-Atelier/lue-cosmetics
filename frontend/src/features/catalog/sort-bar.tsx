interface SortBarProps {
  sortBy: 'name' | 'price_asc' | 'price_desc' | 'rating' | 'newest';
  onSortChange: (sort: 'name' | 'price_asc' | 'price_desc' | 'rating' | 'newest') => void;
  productCount: number;
}

const sortOptions = [
  { value: 'name', label: 'Name A-Z' },
  { value: 'price_asc', label: 'Price: Low to High' },
  { value: 'price_desc', label: 'Price: High to Low' },
  { value: 'rating', label: 'Top Rated' },
  { value: 'newest', label: 'Newest' },
] as const;

export function SortBar({ sortBy, onSortChange, productCount }: SortBarProps) {
  return (
    <div className="border-b border-line-soft py-3">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
        <div className="flex items-center justify-between">
          <p className="text-sm text-ink-muted">
            {productCount} {productCount === 1 ? 'product' : 'products'}
          </p>

          <div className="flex items-center gap-2">
            <span className="text-sm text-ink-muted">Sort by:</span>
            <select
              value={sortBy}
              onChange={(e) => onSortChange(e.target.value as typeof sortBy)}
              className="px-3 py-1.5 border border-line rounded-lg bg-paper text-ink text-sm focus:outline-none focus:ring-2 focus:ring-lavender-400 focus:border-transparent"
            >
              {sortOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>
    </div>
  );
}
