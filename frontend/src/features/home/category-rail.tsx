import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getCategories } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogCategoryView } from '../../lib/api/generated/rueCosmeticsAPI';
import { Icon } from '../shared/ui/icons';

// TODO: Load category images from API. The categories table needs an image_path
// column. Requires: DB migration → update Go model + SQL query → sqlc generate
// → update handler → regenerate API client → then remove this hardcoded map.
const CATEGORY_IMAGE_MAP: Record<string, string> = {
  skincare:  '/categories/skincare.jpg',
  makeup:    '/categories/makeup.jpg',
  'hair-care':'/categories/hair-care.jpg',
  'body-care':'/categories/body-care.jpg',
  fragrance: '/categories/fragrance.jpg',
};

export function CategoryRail() {
  const navigate = useNavigate();
  const [categories, setCategories] = useState<InternalCatalogCategoryView[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadCategories = async () => {
      try {
        const response = await getCategories();
        const categoriesArray = Array.isArray(response) ? response : [];
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
      <section className="section">
        <div className="wrap">
          <div className="section-head">
            <div>
              <div className="eyebrow">Shop by category</div>
              <h2 className="h-display" style={{ fontSize: 'clamp(32px, 4vw, 56px)' }}>Find your <em>next favourite.</em></h2>
            </div>
            <button onClick={() => navigate({ to: '/shop' })} className="section-link">
              View all <Icon name="arrow" size={14} />
            </button>
          </div>
          <div className="cat-rail">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="cat-tile">
                <div className="ph" style={{ aspectRatio: '3/4' }}>
                  <span className="ph-label">loading</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>
    );
  }

  if (categories.length === 0) return null;

  return (
    <section className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Shop by category</div>
            <h2 className="h-display" style={{ fontSize: 'clamp(32px, 4vw, 56px)' }}>Find your <em>next favourite.</em></h2>
          </div>
          <button onClick={() => navigate({ to: '/shop' })} className="section-link">
            View all <Icon name="arrow" size={14} />
          </button>
        </div>

        <div className="cat-rail">
          {categories.slice(0, 6).map((c) => (
            <button
              key={c.id}
              className="cat-tile"
              onClick={() => navigate({ to: '/shop', search: { category: c.slug || undefined } })}
            >
              <img src={CATEGORY_IMAGE_MAP[c.slug!]} alt={c.label || 'Category'} className="cat-tile-img" />
              <div className="cat-tile-foot">
                <div className="cat-tile-name">{c.label || 'Category'}</div>
                {/* <div className="cat-tile-count">
                    {c.product_count} items <Icon name="arrow" size={12} />
                  </div> */}
              </div>
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}
