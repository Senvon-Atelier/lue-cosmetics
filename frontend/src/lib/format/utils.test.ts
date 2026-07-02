import { describe, expect, it } from 'vitest';
import { formatGhs } from './utils';

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
