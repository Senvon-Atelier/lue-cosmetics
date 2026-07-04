// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <a className={className}>{children}</a>
  ),
}));

import { LegalPageView } from './legal-page';

describe('LegalPageView', () => {
  it('renders the page title and lead for a known slug', () => {
    render(<LegalPageView slug="privacy" />);
    expect(screen.getByRole('heading', { level: 1 }).textContent).toContain('Privacy');
    expect(screen.getByText(/what Rue Cosmetics collects/i)).toBeTruthy();
  });

  it('marks the active sidebar entry with aria-current', () => {
    render(<LegalPageView slug="terms" />);
    const current = document.querySelector('[aria-current="page"]');
    expect(current?.textContent).toBe('Terms');
  });

  it('renders an in-shell fallback for an unknown slug', () => {
    render(<LegalPageView slug="does-not-exist" />);
    expect(screen.getByText(/could not be found/i)).toBeTruthy();
    // sidebar still renders every page
    expect(screen.getAllByText('Privacy').length).toBeGreaterThan(0);
  });
});
