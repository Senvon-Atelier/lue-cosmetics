import { useState, useEffect } from 'react';
import { FilterBar } from './filter-bar';
import { SortBar } from './sort-bar';
import { ProductGrid } from './product-grid';
import { getProducts, getCategories, getBrands } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogProductView } from '../../lib/api/generated/rueCosmeticsAPI';

type SortBy = 'name' | 'price_asc' | 'price_desc' | 'rating' | 'newest';

interface ShopPageProps {
  initialCategory?: string;
  initialBrand?: string;
}

export function ShopPage({ initialCategory, initialBrand }: ShopPageProps) {
  const [products, setProducts] = useState<InternalCatalogProductView[]>([]);
  const [categories, setCategories] = useState<Array<{ id: string; label: string; slug: string }>>([]);
  const [brands, setBrands] = useState<Array<{ id: string; name: string; slug: string }>>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(initialCategory || null);
  const [selectedBrand, setSelectedBrand] = useState<string | null>(initialBrand || null);
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
        setCategories(categoriesRes.data.map(cat => ({
          id: cat.id || '',
          label: cat.label || '',
          slug: cat.slug || '',
        })));
        setBrands(brandsRes.data.map(brand => ({
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
        // The response might be directly the products array or wrapped
        const productsData = Array.isArray(response.data) ? response.data : (response.data as any)?.items || [];
        setProducts(productsData);
      } catch (error) {
        console.error('Failed to load products:', error);
        setProducts([]);
      } finally {
        setLoading(false);
      }
    };
    loadProducts();
  }, [selectedCategory, selectedBrand, searchQuery, sortBy]);

  const handleCategoryChange = (slug: string | null) => {
    setSelectedCategory(slug);
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
    <div>
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

      <SortBar
        sortBy={sortBy}
        onSortChange={handleSortChange}
        productCount={products.length}
      />

      <ProductGrid products={products} loading={loading} />
    </div>
  );
}
