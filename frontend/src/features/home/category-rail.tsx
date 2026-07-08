import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { getCategories } from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCatalogCategoryView } from '../../lib/api/generated/rueCosmeticsAPI';
import { Icon } from '../shared/ui/icons';

const TONES = ['lavender', 'cream', 'rose', 'ink', 'lavender', 'cream'] as const;

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
          {categories.slice(0, 6).map((c, i) => (
            <button
              key={c.id}
              className="cat-tile"
              onClick={() => navigate({ to: '/shop', search: { category: c.slug || undefined } })}
            >
              <div className={`ph ph--${TONES[i % TONES.length]}`} style={{ aspectRatio: '3/4' }}>
                <span className="ph-label">{(c.label || 'Category').substring(0, 1)}</span>
              </div>
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
