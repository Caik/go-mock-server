import type { TrafficEntry } from '~/types/traffic';

const API_BASE_URL = import.meta.env.DEV ? 'http://localhost:9090' : '';

/**
 * Subscribe to real-time traffic entries via SSE.
 * The backend sends catch-up entries immediately on connect, then live entries.
 * Returns an unsubscribe function that closes the EventSource.
 */
export function subscribeToTraffic(callback: (entry: TrafficEntry) => void): () => void {
  const es = new EventSource(`${API_BASE_URL}/api/v1/traffic`);

  es.onmessage = (event) => {
    try {
      callback(JSON.parse(event.data) as TrafficEntry);
    } catch {
      // ignore parse errors
    }
  };

  return () => es.close();
}
