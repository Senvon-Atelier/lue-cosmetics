import { Icon } from '../shared/ui/icons';

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
    <div className="flex items-center justify-between mb-8 pb-6 border-b border-line-soft">
      <p className="font-label text-sm text-ink-soft">
        {productCount} {productCount === 1 ? 'product' : 'products'}
      </p>

      <div className="flex items-center gap-3">
        <span className="font-label text-sm text-ink-soft">Sort by:</span>
        <div className="relative">
          <select
            value={sortBy}
            onChange={(e) => onSortChange(e.target.value as typeof sortBy)}
            className="appearance-none pr-8 pl-3 py-2 border border-line rounded bg-paper text-ink text-sm focus:outline-none focus:border-lavender-600 focus:ring-2 focus:ring-lavender-600 transition-all duration-[var(--dur)] cursor-pointer"
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          <Icon name="chevronDown" size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-ink-muted pointer-events-none" />
        </div>
      </div>
    </div>
  );
}
