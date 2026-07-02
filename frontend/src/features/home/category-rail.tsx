import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getCategories } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogCategoryView } from '../../lib/api/generated/rueCosmeticsAPI';
import { Icon } from '../shared/ui/icons';

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
      <section className="section bg-paper">
        <div className="wrap">
          <div className="grid-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="aspect-square bg-lavender-50 rounded-full mb-3 flex items-center justify-center overflow-hidden">
                  <div className="w-20 h-20 bg-lavender-100 rounded-full" />
                </div>
                <div className="h-4 bg-lavender-100 rounded w-3/4 mx-auto" />
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  }

  if (categories.length === 0) return null;

  return (
    <section className="section bg-paper">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Categories</div>
            <h2 className="font-display text-[clamp(22px,2.5vw,32px)] font-normal tracking-[-0.01em]">
              Shop by Category
            </h2>
          </div>
          <button
            onClick={() => navigate({ to: '/shop' })}
            className="section-link"
          >
            <span>View all</span>
            <Icon name="arrow" size={16} />
          </button>
        </div>

        <div className="grid-4">
          {categories.map((category) => (
            <button
              key={category.id}
              onClick={() => navigate({ to: '/shop', search: { category: category.id || undefined } })}
              className="group text-center"
            >
              <div className="aspect-square rounded-full bg-lavender-50 flex items-center justify-center mb-4 overflow-hidden border border-line-soft group-hover:border-lavender-300 group-hover:scale-[1.03] transition-all duration-[600ms] ease-[cubic-bezier(0.2,0.8,0.2,1)] relative">
                <span className="font-display text-5xl text-lavender-300">
                  {(category.label || 'C').substring(0, 1)}
                </span>
              </div>
              <div className="font-label font-medium text-sm text-ink-soft group-hover:text-ink transition-colors duration-[var(--dur)]">
                {category.label || 'Category'}
              </div>
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}
