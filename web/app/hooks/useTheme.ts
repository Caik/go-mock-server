// Theme hook for dark/light mode management
import { useState, useEffect, useCallback } from 'react';

type Theme = 'light' | 'dark';

// Module-level flag - once we've synced on first mount, subsequent mounts can read directly
let hasInitialized = false;

function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';
  const saved = localStorage.getItem('theme');
  return saved === 'light' ? 'light' : 'dark';
}

export function useTheme() {
  // After first initialization, read directly from localStorage to avoid flicker on navigation
  const [theme, setThemeState] = useState<Theme>(() =>
    hasInitialized ? getStoredTheme() : 'dark'
  );
  const [hasSynced, setHasSynced] = useState(hasInitialized);

  // On first mount only, sync with localStorage
  useEffect(() => {
    if (!hasInitialized) {
      const saved = getStoredTheme();
      setThemeState(saved);
      hasInitialized = true;
    }
    setHasSynced(true);
  }, []);

  // Apply theme changes to DOM and localStorage
  const setTheme = useCallback((newTheme: Theme) => {
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    setThemeState(newTheme);
  }, []);

  const toggleTheme = useCallback(() => {
    setThemeState((currentTheme) => {
      const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
      document.documentElement.setAttribute('data-theme', newTheme);
      localStorage.setItem('theme', newTheme);
      return newTheme;
    });
  }, []);

  return { theme, setTheme, toggleTheme, hasSynced };
}

