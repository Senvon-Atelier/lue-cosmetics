import { useState, useEffect } from 'react';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { FilterBar } from './filter-bar';
import { SortBar } from './sort-bar';
import { ProductGrid } from './product-grid';
import { getProducts, getCategories, getBrands } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogProductView, GetProductsParams } from '../../lib/api/generated/rueCosmeticsAPI';
import { Pagination } from './pagination';

type SortBy = 'featured' | 'price_asc' | 'price_desc' | 'rating';

const SORT_API_MAP: Record<string, string> = {
  featured: 'name',
  price_asc: 'price_asc',
  price_desc: 'price_desc',
  rating: 'rating',
};
type ViewMode = 'grid' | 'list';

export function ShopPage() {
  const { category: categoryParam } = useSearch({ from: '/_storefront/shop' });
  const navigate = useNavigate();
  const [products, setProducts] = useState<InternalCatalogProductView[]>([]);
  const [categories, setCategories] = useState<Array<{ id: string; label: string; slug: string }>>([]);
  const [brands, setBrands] = useState<Array<{ id: string; name: string; slug: string }>>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(categoryParam ?? null);
  const [selectedBrand, setSelectedBrand] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<SortBy>('featured');
  const [view, setView] = useState<ViewMode>('grid');
  const [showFilters, setShowFilters] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

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

  useEffect(() => {
    const loadProducts = async () => {
      setLoading(true);
      try {
        const params: GetProductsParams = {};
        if (selectedCategory) params.category = selectedCategory;
        if (selectedBrand) params.brand = selectedBrand;
        if (searchQuery) params.q = searchQuery;
        const apiSort = SORT_API_MAP[sortBy] || sortBy;
        if (sortBy !== 'featured') params.sort = apiSort;
        params.page = page;
        params.limit = 16;

        const response = await getProducts(params);
        setProducts(response?.items || []);
        setTotal(response?.total ?? 0);
      } catch (error) {
        console.error('Failed to load products:', error);
        setProducts([]);
      } finally {
        setLoading(false);
      }
    };
    loadProducts();
  }, [selectedCategory, selectedBrand, searchQuery, sortBy, page]);

  useEffect(() => {
    if (categories.length === 0) return;
    const isKnown = categories.some((c) => c.slug === categoryParam);
    setSelectedCategory(isKnown ? (categoryParam ?? null) : null);
  }, [categoryParam, categories]);

  const handleCategoryChange = (slug: string | null) => {
    setSelectedCategory(slug);
    setPage(1);
    void navigate({
      to: '/shop',
      search: slug ? { category: slug } : {},
      replace: true,
    });
  };

  const handleBrandChange = (slug: string | null) => {
    setSelectedBrand(slug);
    setPage(1);
  };

  const handleSearchChange = (query: string) => {
    setSearchQuery(query);
    setPage(1);
  };

  const handleSortChange = (sort: string) => {
    setSortBy(sort as SortBy);
    setPage(1);
  };

  const handleViewChange = (v: ViewMode) => {
    setView(v);
  };

  const handleClear = () => {
    setSelectedCategory(null);
    setSelectedBrand(null);
    setSearchQuery('');
    setPage(1);
    void navigate({ to: '/shop', search: {}, replace: true });
  };

  const categoryLabel = selectedCategory
    ? categories.find((c) => c.slug === selectedCategory)?.label || 'All products'
    : 'All products';

  return (
    <div>
      <section className="shop-head">
        <div className="wrap">
          <div className="eyebrow">The shop</div>
          <h1 className="h-display shop-title">{categoryLabel}</h1>
          <p className="shop-sub">
            {selectedCategory
              ? `Our edit of ${categoryLabel.toLowerCase()} — trusted names and new discoveries.`
              : `${total || '…'} curated products. Filter to find yours.`}
          </p>
          <div className="shop-cats">
            <button
              className={`chip ${selectedCategory === null ? 'active' : ''}`}
              onClick={() => handleCategoryChange(null)}
            >
              All
            </button>
            {categories.map((c) => (
              <button
                key={c.id}
                className={`chip ${selectedCategory === c.slug ? 'active' : ''}`}
                onClick={() => handleCategoryChange(c.slug)}
              >
                {c.label}
              </button>
            ))}
          </div>
        </div>
      </section>

      <section className="shop-body">
        <div className="wrap shop-body-inner">
        <FilterBar
          brands={brands}
          selectedBrand={selectedBrand}
          searchQuery={searchQuery}
          showFilters={showFilters}
          onBrandChange={handleBrandChange}
          onSearchChange={handleSearchChange}
          onCloseFilters={() => setShowFilters(false)}
          onClear={handleClear}
        />

        <div>
          <SortBar
            sortBy={sortBy}
            onSortChange={handleSortChange}
            productCount={total}
            view={view}
            onViewChange={handleViewChange}
            onOpenFilters={() => setShowFilters(true)}
          />

          <ProductGrid
            products={products}
            loading={loading}
            view={view}
            onReset={handleClear}
          />

          <Pagination
            page={page}
            totalPages={Math.ceil(total / 16)}
            onPageChange={setPage}
          />
        </div>
      </div>
    </section>
    </div>
  );
}
