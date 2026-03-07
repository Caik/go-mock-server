// API client utilities

const API_BASE = '/api/v1';

export async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

// SSE connection for traffic streaming
export function connectTrafficStream(
  onMessage: (entry: unknown) => void,
  onError?: (error: Event) => void,
  filters?: Record<string, string>
): EventSource {
  const params = new URLSearchParams(filters);
  const url = `${API_BASE}/traffic${params.toString() ? '?' + params.toString() : ''}`;
  
  const eventSource = new EventSource(url);
  
  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (e) {
      console.error('Failed to parse SSE message:', e);
    }
  };

  if (onError) {
    eventSource.onerror = onError;
  }

  return eventSource;
}

