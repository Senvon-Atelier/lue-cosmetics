import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui';

export function HomeHero() {
  const navigate = useNavigate();

  return (
    <section className="hero-e2">
      <div className="hero-e2-bg" aria-hidden="true">
        <div className="hero-e2-blob hero-e2-blob-1" />
        <div className="hero-e2-blob hero-e2-blob-2" />
      </div>

      <div className="hero-e2-inner">
        <div className="hero-e2-top">
          <div className="hero-e2-eyebrow">
            <span className="dot" /> Spring 2026 — The Lavender Edit
          </div>
          <div className="hero-e2-rating">
            <span className="stars-row">
              {[0, 1, 2, 3, 4].map((i) => (
                <Icon key={i} name="starFilled" size={12} />
              ))}
            </span>
            <span>Rated 4.9 · 1,200+ Accra reviews</span>
          </div>
        </div>

        <div className="hero-e2-grid">
          <div className="hero-e2-col-l">
            <h1 className="hero-e2-title">
              <span className="line line-1">Soft</span>
              <span className="line line-2">rituals,</span>
              <span className="line line-3">
                <em>quiet</em>
              </span>
              <span className="line line-4">glow.</span>
            </h1>
          </div>

          <div className="hero-e2-col-c">
            <div className="hero-e2-frame">
              <img src="/products/hero-editorial.jpg" alt="Lue Cosmetics editorial" style={{ minHeight: '560px', display: 'block', width: '100%', objectFit: 'cover' }} />
              <div className="hero-e2-chip">
                <div className="hero-e2-chip-dot" />
                <div>
                  <div className="hero-e2-chip-k">Today's ritual</div>
                  <div className="hero-e2-chip-v">Rose Hydration Serum · GHS 245</div>
                </div>
                <button
                  className="hero-e2-chip-go"
                  onClick={() => navigate({ to: '/shop' })}
                  aria-label="View product"
                >
                  <Icon name="arrow" size={14} />
                </button>
              </div>
            </div>
          </div>

          <div className="hero-e2-col-r">
            <div className="hero-e2-stack-t">
              <div className="ph ph--cream" style={{ height: '100%' }}>
                <span className="ph-label">still life</span>
              </div>
            </div>
            <div className="hero-e2-stack-b">
              <div className="hero-e2-number">07</div>
              <div className="hero-e2-numlabel">
                categories
                <br />
                edited weekly
              </div>
            </div>
          </div>
        </div>

        <div className="hero-e2-bottom">
          <div className="hero-e2-lede">
            Home of authentic beauty and wellness. A shelf of trusted names — and a few of our own —
            stocked in Accra, shipped across Ghana.
          </div>
          <div className="hero-e2-ctas">
            <button className="btn btn-primary" onClick={() => navigate({ to: '/shop' })}>
              Shop the edit <Icon name="arrow" size={14} />
            </button>
            <button
              className="hero-e2-link"
              onClick={() => navigate({ to: '/about' })}
            >
              <span>Our story</span>
              <Icon name="arrow" size={14} />
            </button>
          </div>
        </div>

        <div className="hero-e2-marquee">
          <div className="hero-e2-track">
            {[...Array(2)].map((_, k) => (
              <div key={k}>
                {['Nuxe', 'CeraVe', 'The Ordinary', 'La Roche-Posay', 'Shea Moisture', 'Cantu', 'Lue Atelier', "Palmer's", 'Garnier', 'Eucerin'].map(
                  (b, i) => (
                    <span key={`${k}-${i}`} className="hero-e2-brand">
                      {b}
                      <i />
                    </span>
                  )
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
