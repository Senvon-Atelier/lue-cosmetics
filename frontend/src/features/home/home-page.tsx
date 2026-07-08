import { CaseStudyBanner } from '../shared/case-study';
import { HomeHero } from './home-hero';
import { PromiseSection } from './promise-section';
import { CategoryRail } from './category-rail';
import { FeaturedProducts } from './featured-products';
import { RitualBanner } from './ritual-banner';
import { JournalSection } from './journal-section';
import { TestimonialsSection } from './testimonials-section';
import { NewsletterSection } from './newsletter-section';

export function HomePage() {
  return (
    <div>
      <CaseStudyBanner />
      <HomeHero />
      <PromiseSection />
      <CategoryRail />
      <FeaturedProducts />
      <RitualBanner />
      <JournalSection />
      <TestimonialsSection />
      <NewsletterSection />
    </div>
  );
}
