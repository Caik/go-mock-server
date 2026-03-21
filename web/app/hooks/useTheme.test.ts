import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';

// localStorage mock for jsdom environments that don't expose it directly
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; },
  };
})();

describe('useTheme', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', localStorageMock);
    localStorageMock.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    document.documentElement.removeAttribute('data-theme');
  });

  async function importFresh() {
    vi.resetModules();
    const { useTheme } = await import('./useTheme');
    return useTheme;
  }

  it('defaults to dark theme when localStorage is empty', async () => {
    const useTheme = await importFresh();
    const { result } = renderHook(() => useTheme());
    act(() => {});
    expect(result.current.theme).toBe('dark');
  });

  it('reads light theme from localStorage', async () => {
    localStorageMock.setItem('theme', 'light');
    const useTheme = await importFresh();
    const { result } = renderHook(() => useTheme());
    act(() => {});
    expect(result.current.theme).toBe('light');
  });

  it('setTheme writes to localStorage and updates data-theme attribute', async () => {
    const useTheme = await importFresh();
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.setTheme('light');
    });
    expect(result.current.theme).toBe('light');
    expect(localStorageMock.getItem('theme')).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('toggleTheme flips from dark to light', async () => {
    const useTheme = await importFresh();
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.toggleTheme();
    });
    expect(result.current.theme).toBe('light');
  });

  it('toggleTheme flips from light to dark', async () => {
    localStorageMock.setItem('theme', 'light');
    const useTheme = await importFresh();
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.toggleTheme();
    });
    expect(result.current.theme).toBe('dark');
  });

  it('hasSynced is true after mount', async () => {
    const useTheme = await importFresh();
    const { result } = renderHook(() => useTheme());
    act(() => {});
    expect(result.current.hasSynced).toBe(true);
  });
});
