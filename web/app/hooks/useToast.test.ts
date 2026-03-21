import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useToast } from './useToast';

describe('useToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('starts with an empty toast list', () => {
    const { result } = renderHook(() => useToast());
    expect(result.current.toasts).toEqual([]);
  });

  it('adds a success toast', () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast('Saved!', 'success');
    });
    expect(result.current.toasts).toHaveLength(1);
    expect(result.current.toasts[0]).toMatchObject({ message: 'Saved!', type: 'success' });
  });

  it('adds an error toast', () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast('Something went wrong', 'error');
    });
    expect(result.current.toasts[0]).toMatchObject({ message: 'Something went wrong', type: 'error' });
  });

  it('auto-dismisses a toast after 3 seconds', () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast('Temporary', 'success');
    });
    expect(result.current.toasts).toHaveLength(1);

    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.toasts).toHaveLength(0);
  });

  it('does not dismiss toast before 3 seconds', () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast('Still here', 'success');
    });
    act(() => {
      vi.advanceTimersByTime(2999);
    });
    expect(result.current.toasts).toHaveLength(1);
  });

  it('supports multiple simultaneous toasts', () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast('First', 'success');
      result.current.showToast('Second', 'error');
    });
    expect(result.current.toasts).toHaveLength(2);
  });

  it('dismisses toasts independently', () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast('First', 'success');
    });
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    act(() => {
      result.current.showToast('Second', 'success');
    });
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    // First toast (total 3000ms) dismissed, second (1000ms old) still present
    expect(result.current.toasts).toHaveLength(1);
    expect(result.current.toasts[0].message).toBe('Second');
  });
});
