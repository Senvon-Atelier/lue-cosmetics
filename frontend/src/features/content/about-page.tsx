import { useNavigate } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';

const values = [
  { n: '01', t: 'Edited, not stocked', d: 'If we wouldn\'t use it, we don\'t sell it. That rule narrows the shelf — and sharpens it.' },
  { n: '02', t: 'Authentic, no exceptions', d: 'Every product is sourced directly from authorised distributors. You can ask. We keep the paper trail.' },
  { n: '03', t: 'Service, not selling', d: 'Our team is trained to advise, not to upsell. If a GHS 65 cream is the right answer, that\'s the answer.' },
  { n: '04', t: 'Home-grown pride', d: 'We champion African ingredients — shea, black soap, baobab — and the brands building with them.' },
  { n: '05', t: 'Made for Ghana\'s weather', d: 'Harmattan, humidity, harsh sun — we test for the conditions you actually live in.' },
  { n: '06', t: 'Quiet, not loud', d: 'No fear-selling. No miracle claims. Just consistent, considered beauty.' },
];

export function AboutPage() {
  const navigate = useNavigate();

  return (
    <div>
      <section className="about-hero">
        <div className="wrap about-hero-wrap">
          <div className="eyebrow">Our story</div>
          <h1 className="h-display" style={{ fontSize: 'clamp(48px, 8vw, 128px)' }}>
            Beauty is a <em>practice,</em><br />not a purchase.
          </h1>
          <p className="hero-lede" style={{ maxWidth: 680 }}>
            Founded in Accra in 2024, Lue Cosmetics began with a simple mission: to bring authentic,
            trusted beauty products to Ghanaians — without the markup, without the wait.
          </p>
        </div>
      </section>

      <section className="section">
        <div className="wrap about-split">
          <div className="ph ph--lavender about-ph"><span className="ph-label">founder · portrait</span></div>
          <div>
            <div className="eyebrow">Founded in Accra</div>
            <h2 className="h-display" style={{ fontSize: 'clamp(32px, 4.5vw, 64px)' }}>
              Why <em>Lue?</em>
            </h2>
            <p>
              We noticed something frustrating: Ghanaians were paying premium prices for international
              beauty brands, or waiting weeks for deliveries that might never arrive.
            </p>
            <p>
              So we built Lue differently. We source directly from brands and authorized distributors,
              stock locally in Accra, and ship nationwide within days — not weeks.
            </p>
            <p>
              Our shelves carry trusted names like CeraVe, The Ordinary, and La Roche-Posay alongside
              curated skincare picks from emerging clean brands. Every product is verified authentic,
              every order is tracked.
            </p>
            <button className="btn btn-ghost" onClick={() => navigate({ to: '/shop' })}>
              Browse the edit <Icon name="arrow" size={14} />
            </button>
          </div>
        </div>
      </section>

      <section className="section values">
        <div className="wrap">
          <div className="eyebrow" style={{ textAlign: 'center', marginBottom: 12 }}>Our principles</div>
          <h2 className="h-display" style={{ textAlign: 'center', fontSize: 'clamp(36px, 5vw, 72px)', marginBottom: 60 }}>
            What we stand <em>behind.</em>
          </h2>
          <div className="grid-3">
            {values.map(v => (
              <div key={v.n} className="value">
                <div className="value-n">{v.n}</div>
                <h3 className="value-t">{v.t}</h3>
                <p className="value-d">{v.d}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section stats">
        <div className="wrap">
          <div className="stats-grid">
            <div><div className="stat-n h-display">2021</div><div className="stat-l">The shop opened</div></div>
            <div><div className="stat-n h-display">240+</div><div className="stat-l">Curated products</div></div>
            <div><div className="stat-n h-display">12,000</div><div className="stat-l">Customers served</div></div>
            <div><div className="stat-n h-display">10</div><div className="stat-l">Regions delivered to</div></div>
          </div>
        </div>
      </section>

      <section className="section visit">
        <div className="wrap visit-wrap">
          <div>
            <div className="eyebrow">Come say hi</div>
            <h2 className="h-display" style={{ fontSize: 'clamp(36px, 5vw, 72px)' }}>
              Visit the <em>shop.</em>
            </h2>
            <div className="visit-info">
              <div><Icon name="pin" /> <div><strong>Community 18, Spintex</strong><br />Adjacent KFC, Accra, Ghana</div></div>
              <div><Icon name="clock" /> <div><strong>Monday – Saturday</strong><br />9:00am – 8:00pm</div></div>
              <div><Icon name="phone" /> <div><strong>0594 701 345</strong><br />WhatsApp welcome</div></div>
            </div>
          </div>
          <div className="ph ph--cream visit-map"><span className="ph-label">map · Spintex, Accra</span></div>
        </div>
      </section>
    </div>
  );
}
