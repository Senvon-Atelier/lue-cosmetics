import type { Session } from './session';

export type GuardRequirement = 'authenticated' | 'admin';

/**
 * Pure routing decision: where to redirect (or null to allow).
 * Kept free of router/query imports so it is trivially unit-testable.
 */
export function redirectPathFor(
  session: Session | null,
  requirement: GuardRequirement,
): string | null {
  if (!session) return '/login';
  if (requirement === 'admin' && session.role !== 'admin') return '/';
  return null;
}
