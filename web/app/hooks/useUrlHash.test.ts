import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useUrlHash } from './useUrlHash';

describe('useUrlHash', () => {
  beforeEach(() => {
    // Reset hash before each test
    window.history.replaceState(null, '', window.location.pathname);
    vi.spyOn(window.history, 'replaceState');
  });

  it('returns initialValue when no hash is present', () => {
    const { result } = renderHook(() => useUrlHash('default'));
    expect(result.current[0]).toBe('default');
  });

  it('returns empty string as default initialValue', () => {
    const { result } = renderHook(() => useUrlHash());
    expect(result.current[0]).toBe('');
  });

  it('reads existing hash from window.location on mount', () => {
    window.history.replaceState(null, '', '#existing-hash');
    const { result } = renderHook(() => useUrlHash(''));
    expect(result.current[0]).toBe('existing-hash');
  });

  it('setter updates the state value', () => {
    const { result } = renderHook(() => useUrlHash(''));
    act(() => {
      result.current[1]('new-value');
    });
    expect(result.current[0]).toBe('new-value');
  });

  it('setter writes to window.location via replaceState', () => {
    const { result } = renderHook(() => useUrlHash(''));
    act(() => {
      result.current[1]('my-hash');
    });
    expect(window.history.replaceState).toHaveBeenCalledWith(null, '', '#my-hash');
  });

  it('setter with empty string removes the hash', () => {
    const { result } = renderHook(() => useUrlHash('initial'));
    act(() => {
      result.current[1]('');
    });
    expect(window.history.replaceState).toHaveBeenCalledWith(null, '', window.location.pathname);
  });
});
