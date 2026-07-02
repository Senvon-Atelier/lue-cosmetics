// Formatting utilities for Rue Cosmetics

/**
 * Format price from minor units (pesewas) to GHS string
 * @param priceMinor - Price in minor units (1 GHS = 100 pesewas)
 * @returns Formatted price string (e.g., "GHS 45.00")
 */
export function formatPrice(priceMinor: number | undefined | null): string {
  if (priceMinor == null) return 'GHS 0.00';
  const ghs = priceMinor / 100;
  return `GHS ${ghs.toFixed(2)}`;
}

/**
 * Format date to readable string
 * @param dateString - ISO date string
 * @returns Formatted date string (e.g., "Jun 28, 2026")
 */
export function formatDate(dateString: string | undefined | null): string {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-GH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Format date and time
 * @param dateString - ISO date string
 * @returns Formatted datetime string (e.g., "Jun 28, 2026, 4:05 PM")
 */
export function formatDateTime(dateString: string | undefined | null): string {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-GH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
}

/**
 * Truncate text with ellipsis
 * @param text - Text to truncate
 * @param maxLength - Maximum length before truncation
 * @returns Truncated text with ellipsis if needed
 */
export function truncate(text: string | undefined | null, maxLength: number): string {
  if (!text) return '';
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + '...';
}

/**
 * Format rating to display string
 * @param rating - Rating number (e.g., 4.8)
 * @param reviewCount - Number of reviews
 * @returns Formatted rating string (e.g., "4.8 (142 reviews)")
 */
export function formatRating(rating: number | undefined | null, reviewCount: number | undefined | null): string {
  if (rating == null) return '';
  const reviews = reviewCount ? ` (${reviewCount} ${reviewCount === 1 ? 'review' : 'reviews'})` : '';
  return `${rating.toFixed(1)}${reviews}`;
}

/**
 * Convert product image path to full URL
 * @param imagePath - Relative image path from API
 * @returns Full image URL
 */
export function getImageUrl(imagePath: string | undefined | null): string {
  if (!imagePath) return '';
  // In dev, use relative path to public folder
  // In prod, use absolute URL to CDN or backend
  if (import.meta.env.DEV) {
    return `/${imagePath}`;
  }
  return imagePath; // In prod, assume imagePath is already a full URL or relative to CDN
}

/** Mockup-style GHS price: "GHS 480", "GHS 480.50". Input is minor units. */
export function formatGhs(minor: number): string {
  const cedis = minor / 100;
  const hasPesewas = minor % 100 !== 0;
  return `GHS ${cedis.toLocaleString('en-US', {
    minimumFractionDigits: hasPesewas ? 2 : 0,
    maximumFractionDigits: hasPesewas ? 2 : 0,
  })}`;
}
