import { useState, useEffect } from 'react';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { FilterBar } from './filter-bar';
import { SortBar } from './sort-bar';
import { ProductGrid } from './product-grid';
import { getProducts, getCategories, getBrands } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

type SortBy = 'name' | 'price_asc' | 'price_desc' | 'rating' | 'newest';

export function ShopPage() {
  const { category: categoryParam } = useSearch({ from: '/_storefront/shop' });
  const navigate = useNavigate();
  const [products, setProducts] = useState<InternalCatalogProductView[]>([]);
  const [categories, setCategories] = useState<Array<{ id: string; label: string; slug: string }>>([]);
  const [brands, setBrands] = useState<Array<{ id: string; name: string; slug: string }>>([]);
  const [loading, setLoading] = useState(true);
  // selectedCategory holds the category SLUG (the API filters by slug — see handler.go:118)
  const [selectedCategory, setSelectedCategory] = useState<string | null>(categoryParam ?? null);
  const [selectedBrand, setSelectedBrand] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<SortBy>('name');

  // Load categories and brands
  useEffect(() => {
    const loadFilters = async () => {
      try {
        const [categoriesRes, brandsRes] = await Promise.all([
          getCategories(),
          getBrands(),
        ]);
        setCategories(categoriesRes.map(cat => ({
          id: cat.id || '',
          label: cat.label || '',
          slug: cat.slug || '',
        })));
        setBrands(brandsRes.map(brand => ({
          id: brand.id || '',
          name: brand.name || '',
          slug: brand.slug || '',
        })));
      } catch (error) {
        console.error('Failed to load filters:', error);
      }
    };
    loadFilters();
  }, []);

  // Load products with filters
  useEffect(() => {
    const loadProducts = async () => {
      setLoading(true);
      try {
        const params: Record<string, string> = {};
        if (selectedCategory) params.category = selectedCategory;
        if (selectedBrand) params.brand = selectedBrand;
        if (searchQuery) params.q = searchQuery;
        params.sort = sortBy;

        const response = await getProducts(params);
        setProducts(response?.items || []);
      } catch (error) {
        console.error('Failed to load products:', error);
        setProducts([]);
      } finally {
        setLoading(false);
      }
    };
    loadProducts();
  }, [selectedCategory, selectedBrand, searchQuery, sortBy]);

  // URL is the source of truth: back/forward + deep links update the filter
  useEffect(() => {
    setSelectedCategory(categoryParam ?? null);
  }, [categoryParam]);

  const handleCategoryChange = (slug: string | null) => {
    setSelectedCategory(slug);
    void navigate({
      to: '/shop',
      search: slug ? { category: slug } : {},
      replace: true,
    });
  };

  const handleBrandChange = (slug: string | null) => {
    setSelectedBrand(slug);
  };

  const handleSearchChange = (query: string) => {
    setSearchQuery(query);
  };

  const handleSortChange = (sort: SortBy) => {
    setSortBy(sort);
  };

  return (
    <div className="section">
      <div className="wrap">
        {/* Page Header */}
        <div className="mb-12">
          <div className="eyebrow">Shop</div>
          <h1 className="font-display text-[clamp(32px,4vw,56px)] font-normal tracking-[-0.01em]">
            All Products
          </h1>
          <p className="text-ink-muted">Browse our curated collection of skincare, haircare, and wellness products.</p>
        </div>

        {/* Category Chips */}
        {categories.length > 0 && (
          <div className="flex gap-3 mb-12 flex-wrap">
            <button
              onClick={() => handleCategoryChange(null)}
              className={`chip transition-colors duration-[var(--dur)] ${
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
                onClick={() => handleCategoryChange(category.slug)}
                className={`chip transition-colors duration-[var(--dur)] ${
                  selectedCategory === category.slug
                    ? 'bg-lavender-600 text-paper'
                    : 'bg-lavender-100 text-ink hover:bg-lavender-200'
                }`}
              >
                {category.label}
              </button>
            ))}
          </div>
        )}

        {/* Two-Column Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-12">
          {/* Filters Sidebar */}
          <aside className="lg:sticky lg:top-24 lg:self-start h-fit">
            <FilterBar
              categories={categories}
              brands={brands}
              selectedCategory={selectedCategory}
              selectedBrand={selectedBrand}
              searchQuery={searchQuery}
              onCategoryChange={handleCategoryChange}
              onBrandChange={handleBrandChange}
              onSearchChange={handleSearchChange}
            />
          </aside>

          {/* Main Content */}
          <div>
            <SortBar
              sortBy={sortBy}
              onSortChange={handleSortChange}
              productCount={products.length}
            />

            <ProductGrid products={products} loading={loading} />
          </div>
        </div>
      </div>
    </div>
  );
}
