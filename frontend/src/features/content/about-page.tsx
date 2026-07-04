import { Link } from '@tanstack/react-router';
import { Brand, Button, Icon } from '../shared/ui';

export function AboutPage() {
  return (
    <div className="min-h-screen bg-paper text-ink font-body">
      <section className="py-20 bg-lavender-50">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <div className="max-w-3xl mx-auto text-center">
            <h1 className="font-display text-5xl mb-6">Our Story</h1>
            <p className="font-serif text-xl text-ink-soft leading-relaxed">
              Founded in Accra in 2024, Rue Cosmetics began with a simple mission: to bring authentic,
              trusted beauty products to Ghanaians — without the markup, without the wait.
            </p>
          </div>
        </div>
      </section>

      <section className="py-20">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="font-display text-3xl mb-4">Why Rue?</h2>
              <div className="space-y-4 text-ink-soft">
                <p>
                  We noticed something frustrating: Ghanaians were paying premium prices for international
                  beauty brands, or waiting weeks for deliveries that might never arrive.
                </p>
                <p>
                  So we built Rue differently. We source directly from brands and authorized distributors,
                  stock locally in Accra, and ship nationwide within days — not weeks.
                </p>
                <p>
                  Our shelves carry trusted names like CeraVe, The Ordinary, and La Roche-Posay alongside
                  curated skincare picks from emerging clean brands. Every product is verified authentic,
                  every order is tracked.
                </p>
              </div>
            </div>
            <div className="bg-lavender-100 rounded-lg p-8 aspect-square flex items-center justify-center">
              <Brand />
            </div>
          </div>
        </div>
      </section>

      <section className="py-20 bg-lavender-50">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <h2 className="font-display text-3xl mb-8 text-center">Our Values</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-paper rounded-lg p-8">
              <div className="w-12 h-12 rounded-full bg-lavender-100 flex items-center justify-center mb-4">
                <Icon name="shield" size={24} className="text-lavender-600" />
              </div>
              <h3 className="font-label font-semibold text-lg mb-2">Authenticity Guaranteed</h3>
              <p className="text-ink-soft text-sm">
                Every product is sourced directly from brands or authorized distributors. No counterfeits,
                no compromises.
              </p>
            </div>

            <div className="bg-paper rounded-lg p-8">
              <div className="w-12 h-12 rounded-full bg-lavender-100 flex items-center justify-center mb-4">
                <Icon name="truck" size={24} className="text-lavender-600" />
              </div>
              <h3 className="font-label font-semibold text-lg mb-2">Fast Delivery</h3>
              <p className="text-ink-soft text-sm">
                Same-day delivery in Accra, 2-3 days nationwide. No more waiting weeks for international
                shipments.
              </p>
            </div>

            <div className="bg-paper rounded-lg p-8">
              <div className="w-12 h-12 rounded-full bg-lavender-100 flex items-center justify-center mb-4">
                <Icon name="heart" size={24} className="text-lavender-600" />
              </div>
              <h3 className="font-label font-semibold text-lg mb-2">Curated with Care</h3>
              <p className="text-ink-soft text-sm">
                Our team tests every product before it hits our shelves. We only stock what we'd use
                ourselves.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="py-20">
        <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '2rem' }}>
          <div className="max-w-2xl mx-auto text-center">
            <h2 className="font-display text-3xl mb-4">Shop the Collection</h2>
            <p className="text-ink-soft mb-8">
              Explore hundreds of authentic beauty products, ready to ship nationwide.
            </p>
            <Link to="/shop" search={{}}>
              <Button variant="primary" icon="arrow" iconPosition="right">
                Browse Products
              </Button>
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
