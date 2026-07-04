// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { LEGAL_PAGES, getLegalPage } from './legal';

describe('legal registry', () => {
  it('has the five expected pages in order', () => {
    expect(LEGAL_PAGES.map((p) => p.slug)).toEqual([
      'privacy',
      'terms',
      'cookies',
      'shipping',
      'returns',
    ]);
  });

  it('every page has a navLabel, title, lastUpdated, lead, and body', () => {
    for (const p of LEGAL_PAGES) {
      expect(p.navLabel).toBeTruthy();
      expect(p.title).toBeTruthy();
      expect(p.lastUpdated).toBeTruthy();
      expect(p.lead.length).toBeGreaterThan(0);
      expect(p.body).toBeTruthy();
    }
  });

  it('slugs are unique', () => {
    const slugs = LEGAL_PAGES.map((p) => p.slug);
    expect(new Set(slugs).size).toBe(slugs.length);
  });

  it('getLegalPage returns the entry for a known slug and undefined otherwise', () => {
    expect(getLegalPage('privacy')?.navLabel).toBe('Privacy');
    expect(getLegalPage('nope')).toBeUndefined();
  });
});
