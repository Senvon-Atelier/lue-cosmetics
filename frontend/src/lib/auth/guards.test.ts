import { describe, expect, it } from 'vitest';
import { redirectPathFor } from './guards';

describe('redirectPathFor', () => {
  it('sends anonymous users to /login regardless of requirement', () => {
    expect(redirectPathFor(null, 'authenticated')).toBe('/login');
    expect(redirectPathFor(null, 'admin')).toBe('/login');
  });

  it('lets authenticated customers into authenticated areas', () => {
    expect(redirectPathFor({ role: 'customer' }, 'authenticated')).toBeNull();
  });

  it('bounces non-admins from admin areas to the homepage', () => {
    expect(redirectPathFor({ role: 'customer' }, 'admin')).toBe('/');
    expect(redirectPathFor({}, 'admin')).toBe('/');
  });

  it('lets admins into both areas', () => {
    expect(redirectPathFor({ role: 'admin' }, 'authenticated')).toBeNull();
    expect(redirectPathFor({ role: 'admin' }, 'admin')).toBeNull();
  });
});
