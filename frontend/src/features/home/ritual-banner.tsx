import { Icon } from '../shared/ui/icons';

export function RitualBanner() {
  return (
    <section className="section ritual">
      <div className="ritual-wrap">
        <div className="ritual-copy">
          <div className="eyebrow" style={{ color: 'var(--lavender-300)' }}>The Lue Atelier</div>
          <h2 className="h-display" style={{ fontSize: 'clamp(40px, 6vw, 88px)' }}>
            <em>Nuit de Prélude</em>
            <br />
            has arrived.
          </h2>
          <p style={{ fontSize: 'clamp(16px, 1.4vw, 20px)', lineHeight: 1.55, maxWidth: 520, margin: '24px 0 32px' }}>
            Our first in-house fragrance. Two years in the making,
            composed in Grasse, bottled in Accra. Velvet iris, warm amber,
            and a finish of damp pavement after rain.
          </p>
          <div style={{ display: 'flex', gap: 20, alignItems: 'center', flexWrap: 'wrap' }}>
            <button className="btn btn-secondary">
              Discover the scent <Icon name="arrow" size={14} />
            </button>
          </div>
        </div>
        <div className="ritual-visual">
          {/* TODO: Load from API if/when ritual content becomes dynamic */}
          <img src="/ritual/nuit-de-prelude.jpg" alt="Nuit de Prélude fragrance" />
        </div>
      </div>
    </section>
  );
}
