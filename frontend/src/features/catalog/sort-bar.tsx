import { Icon } from '../shared/ui/icons';

interface SortBarProps {
  sortBy: string;
  onSortChange: (sort: string) => void;
  productCount: number;
  view: 'grid' | 'list';
  onViewChange: (view: 'grid' | 'list') => void;
  onOpenFilters: () => void;
}

const sortOptions = [
  { value: 'featured', label: 'Featured' },
  { value: 'price_asc', label: 'Price · low to high' },
  { value: 'price_desc', label: 'Price · high to low' },
  { value: 'rating', label: 'Highest rated' },
] as const;

export function SortBar({ sortBy, onSortChange, productCount, view, onViewChange, onOpenFilters }: SortBarProps) {
  return (
    <div className="shop-bar">
      <button className="btn btn-ghost shop-filter-btn" onClick={onOpenFilters}>
        <Icon name="sliders" size={14} /> Filters
      </button>
      <div className="shop-bar-right">
        <span className="shop-count">{productCount} items</span>
        <div className="shop-view">
          <button className={view === 'grid' ? 'active' : ''} onClick={() => onViewChange('grid')}>
            <Icon name="grid" size={14} />
          </button>
          <button className={view === 'list' ? 'active' : ''} onClick={() => onViewChange('list')}>
            <Icon name="list" size={14} />
          </button>
        </div>
        <div className="shop-sort">
          <select value={sortBy} onChange={(e) => onSortChange(e.target.value as typeof sortBy)}>
            {sortOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
          <Icon name="chevronDown" size={14} />
        </div>
      </div>
    </div>
  );
}
