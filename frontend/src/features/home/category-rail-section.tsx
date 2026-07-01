import { useNavigate } from '@tanstack/react-router';

const categories = [
  { name: 'Skincare', count: 124, color: 'lavender' },
  { name: 'Haircare', count: 89, color: 'cream' },
  { name: 'Bodycare', count: 67, color: 'rose' },
  { name: 'Fragrance', count: 45, color: 'lavender' },
  { name: 'Makeup', count: 78, color: 'cream' },
  { name: 'Wellness', count: 56, color: 'rose' },
  { name: 'Gifts', count: 34, color: 'lavender' },
];

export function CategoryRailSection() {
  const navigate = useNavigate();

  return (
    <section className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">Browse by category</div>
            <h2 className="h-display">Find your next favourite</h2>
          </div>
        </div>

        <div className="cat-rail">
          {categories.map((cat) => (
            <div
              key={cat.name}
              className="cat-tile"
              onClick={() => navigate({ to: '/shop', search: { category: cat.name.toLowerCase() } })}
              role="button"
              tabIndex={0}
            >
              <div className={`ph ph--${cat.color}`} style={{ aspectRatio: '1/1' }}>
                <span className="ph-label">{cat.name}</span>
              </div>
              <div className="cat-tile-foot">
                <span className="cat-tile-name">{cat.name}</span>
                <span className="cat-tile-count">
                  {cat.count}
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
