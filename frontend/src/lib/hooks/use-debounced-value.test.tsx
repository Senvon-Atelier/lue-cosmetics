/** @vitest-environment jsdom */
import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useDebouncedValue } from './use-debounced-value';

describe('useDebouncedValue', () => {
  it('holds the old value until the delay elapses', () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHook(({ v }) => useDebouncedValue(v, 300), {
      initialProps: { v: 'a' },
    });
    rerender({ v: 'ab' });
    expect(result.current).toBe('a');
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(result.current).toBe('ab');
    vi.useRealTimers();
  });
});
