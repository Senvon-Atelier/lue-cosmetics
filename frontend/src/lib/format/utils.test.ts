import { describe, expect, it } from 'vitest';
import { formatGhs, formatOrderDate } from './utils';

describe('formatGhs', () => {
  it('renders whole cedis without decimals (mockup style)', () => {
    expect(formatGhs(48000)).toBe('GHS 480');
    expect(formatGhs(0)).toBe('GHS 0');
  });
  it('keeps pesewas when present instead of lying by rounding', () => {
    expect(formatGhs(48050)).toBe('GHS 480.50');
  });
  it('groups thousands', () => {
    expect(formatGhs(123456700)).toBe('GHS 1,234,567');
  });
});

describe('formatOrderDate', () => {
  it('renders mockup-style dates', () => {
    expect(formatOrderDate('2026-04-04T11:24:00Z')).toBe('Apr 04, 2026');
  });
  it('returns empty string for missing/invalid input', () => {
    expect(formatOrderDate(undefined)).toBe('');
    expect(formatOrderDate('not-a-date')).toBe('');
  });
});
