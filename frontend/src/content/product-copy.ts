/**
 * Curated editorial copy for PDP lede + tabs, keyed by category slug.
 * Bridge until the backend gains per-product description/how_to/ingredients
 * (spec §8 backend follow-up #1) — API values take precedence when they exist.
 *
 * Category slugs are verbatim from backend/cmd/seed/data/categories.json.
 * Shipping threshold (GHS 500) is from backend/config/shipping_config.json
 * (free_over_ghs_minor: 50000 → GHS 500). Hardcoded here as the single
 * frontend source of truth; update if shipping_config.json changes.
 */

export interface ProductCopy {
  lede: string;
  description: string;
  howTo?: string;
  ingredients?: string;
}

const FALLBACK: ProductCopy = {
  lede: 'A considered formulation that delivers on its promise. Tested in-store, recommended by our beauty team, and loved by our regulars.',
  description:
    'A hydrating, lightweight treatment designed for daily use. Its texture absorbs quickly without residue, leaving skin visibly smoother and more even-toned after consistent use. Dermatologically tested. Suitable for sensitive skin types.',
};

const BY_CATEGORY: Record<string, ProductCopy> = {
  // slug: "skincare" (categories.json)
  skincare: {
    ...FALLBACK,
    howTo:
      'Apply morning and evening to cleansed skin. Gently press a few drops into your face and neck, avoiding the eye area. Follow with moisturiser and — in the morning — SPF 30 or higher.',
    ingredients:
      'Aqua, Glycerin, Niacinamide, Sodium Hyaluronate, Panthenol, Allantoin, Tocopherol, Propanediol, Citric Acid, Sodium Benzoate, Phenoxyethanol, Parfum.',
  },
  // slug: "hair-care" (categories.json — note hyphen, not "haircare")
  'hair-care': {
    lede: 'Salon-grade care for every texture. Chosen by our stylists, kept on our own shelves.',
    description:
      'A nourishing treatment that restores softness and shine without weighing hair down. Suitable for protective styles and colour-treated hair.',
    howTo:
      'Work a small amount through damp hair from mid-length to ends. Leave in, or rinse after 10 minutes for a deeper treatment.',
  },
  // slug: "body-care" (categories.json — not "body")
  'body-care': {
    lede: 'Everyday body care that feels like a ritual, not a routine.',
    description:
      'Rich, fast-absorbing care that leaves skin supple through harmattan and humidity alike. No residue, no heaviness.',
    howTo: 'Massage into damp skin after bathing. Reapply to hands and elbows as needed.',
  },
  // slug: "fragrance" (categories.json)
  fragrance: {
    lede: 'Scents with a memory. Composed to last through a Ghana afternoon.',
    description:
      'A layered composition that opens bright and settles into a warm, lasting base. Concentrated for longevity.',
    howTo: 'Spray onto pulse points — wrists, neck, behind the ears. Do not rub.',
  },
  // slug: "makeup" (categories.json) — no category-specific copy; uses FALLBACK
};

export function getProductCopy(categorySlug: string | undefined): ProductCopy {
  return (categorySlug && BY_CATEGORY[categorySlug]) || FALLBACK;
}

/**
 * Product detail page perks list.
 * Shipping threshold: free_over_ghs_minor 50000 → GHS 500
 * (backend/config/shipping_config.json).
 */
export const PERKS: string[] = [
  'Free delivery in Accra over GHS 500',
  '100% authentic, guaranteed',
  'Questions? WhatsApp us',
];
