import { HomeHero } from './home-hero';
import { PromiseSection } from './promise-section';
import { CategoryRail } from './category-rail';
import { FeaturedProducts } from './featured-products';
import { JournalSection } from './journal-section';
import { TestimonialsSection } from './testimonials-section';
import { NewsletterSection } from './newsletter-section';

export function HomePage() {
  return (
    <div>
      <HomeHero />
      <PromiseSection />
      <CategoryRail />
      <FeaturedProducts />
      <JournalSection />
      <TestimonialsSection />
      <NewsletterSection />
    </div>
  );
}
