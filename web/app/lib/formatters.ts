// Shared formatting utilities

/**
 * Format a timestamp to a time string (HH:MM:SS)
 */
export function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('en-US', { hour12: false });
}

/**
 * Get CSS class for HTTP status code
 */
export function getStatusClass(status: number): string {
  if (status >= 500) return 's5xx';
  if (status >= 400) return 's4xx';
  if (status >= 300) return 's3xx';
  if (status >= 200) return 's2xx';
  return 's1xx';
}

/**
 * Format headers object to string
 */
export function formatHeaders(headers: Record<string, string>): string {
  return Object.entries(headers)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n');
}

/**
 * Format JSON body with pretty printing
 */
export function formatBody(body?: string): string {
  if (!body) return '';
  try {
    const parsed = JSON.parse(body);
    return JSON.stringify(parsed, null, 2);
  } catch {
    return body;
  }
}

/**
 * Format error rate as percentage
 */
export function formatErrorRate(rate: number): string {
  return `${Math.round(rate * 100)}%`;
}

