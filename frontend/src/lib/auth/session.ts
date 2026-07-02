import { queryOptions } from '@tanstack/react-query';
import { getAuthSession } from '../api/generated/rueCosmeticsAPI';
import type { InternalAuthSessionResponse } from '../api/generated/rueCosmeticsAPI';
import { ApiError } from '../api/client';

export type Session = InternalAuthSessionResponse;

export const sessionQueryOptions = queryOptions({
  queryKey: ['auth', 'session'] as const,
  queryFn: async (): Promise<Session | null> => {
    try {
      const session = await getAuthSession();
      return session ?? null;
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) return null;
      throw err;
    }
  },
  staleTime: 60_000,
});
