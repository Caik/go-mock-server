// Hook for managing URL hash state
import { useState, useEffect, useCallback } from 'react';

/**
 * Custom hook to sync state with URL hash
 * @param initialValue - Default value when no hash is present
 * @returns [value, setValue] - Current hash value and setter function
 */
export function useUrlHash(initialValue: string = ''): [string, (value: string) => void] {
  const [hash, setHashState] = useState<string>(() => {
    if (typeof window === 'undefined') return initialValue;
    return window.location.hash.slice(1) || initialValue;
  });

  // Sync hash from URL on mount
  useEffect(() => {
    const currentHash = window.location.hash.slice(1);
    if (currentHash) {
      setHashState(currentHash);
    }
  }, []);

  // Update URL and state
  const setHash = useCallback((value: string) => {
    setHashState(value);
    if (value) {
      window.history.replaceState(null, '', `#${value}`);
    } else {
      window.history.replaceState(null, '', window.location.pathname);
    }
  }, []);

  return [hash, setHash];
}

