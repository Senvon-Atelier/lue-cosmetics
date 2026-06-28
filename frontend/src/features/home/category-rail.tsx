import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getCategories } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogCategoryView } from '../../lib/api/generated/rueCosmeticsAPI';

export function CategoryRail() {
  const navigate = useNavigate();
  const [categories, setCategories] = useState<InternalCatalogCategoryView[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadCategories = async () => {
      try {
        const response = await getCategories();
        const data = response.data;
        // Ensure data is an array
        const categoriesArray = Array.isArray(data) ? data : [];
        setCategories(categoriesArray);
      } catch (error) {
        console.error('Failed to load categories:', error);
        setCategories([]);
      } finally {
        setIsLoading(false);
      }
    };
    loadCategories();
  }, []);

  if (isLoading) {
    return (
      <section className="py-16 bg-paper">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
          <div className="flex gap-4 overflow-hidden">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="flex-shrink-0 w-32 animate-pulse">
                <div className="w-32 h-32 bg-lavender-100 rounded-full mb-3" />
                <div className="h-4 bg-lavender-100 rounded w-3/4" />
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  }

  if (categories.length === 0) return null;

  return (
    <section className="py-16 bg-paper">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
        <div className="mb-6">
          <h2 className="font-display text-2xl mb-2">Shop by Category</h2>
          <p className="text-ink-muted">Browse our curated collections</p>
        </div>

        <div className="flex gap-6 overflow-x-auto pb-4 scrollbar-hide">
          {categories.map((category) => (
            <button
              key={category.id}
              onClick={() => navigate({ to: '/shop', search: { category: category.id || undefined } })}
              className="flex-shrink-0 group text-center"
            >
              <div className="w-32 h-32 rounded-full bg-lavender-50 flex items-center justify-center mb-3 overflow-hidden border-2 border-transparent group-hover:border-lavender-300 transition-colors">
                <span className="font-display text-3xl text-lavender-300">
                  {(category.label || 'C').substring(0, 1)}
                </span>
              </div>
              <div className="font-label font-medium text-sm text-ink-soft group-hover:text-ink transition-colors">
                {category.label || 'Category'}
              </div>
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}
