import { describe, it, expect } from 'vitest';
import { formatTime, getStatusClass, formatHeaders, formatBody, formatErrorRate } from './formatters';

describe('formatTime', () => {
  it('formats a UTC timestamp to HH:MM:SS', () => {
    // Use a fixed timestamp and assert the shape, not the exact value (locale-dependent)
    const result = formatTime('2024-01-15T13:04:05.000Z');
    expect(result).toMatch(/^\d{2}:\d{2}:\d{2}$/);
  });

  it('returns a valid time string for any ISO timestamp', () => {
    const result = formatTime('2024-06-30T00:00:00.000Z');
    expect(result).toMatch(/^\d{2}:\d{2}:\d{2}$/);
  });
});

describe('getStatusClass', () => {
  it('returns s1xx for 1xx statuses', () => {
    expect(getStatusClass(100)).toBe('s1xx');
    expect(getStatusClass(199)).toBe('s1xx');
  });

  it('returns s2xx for 2xx statuses', () => {
    expect(getStatusClass(200)).toBe('s2xx');
    expect(getStatusClass(201)).toBe('s2xx');
    expect(getStatusClass(299)).toBe('s2xx');
  });

  it('returns s3xx for 3xx statuses', () => {
    expect(getStatusClass(300)).toBe('s3xx');
    expect(getStatusClass(304)).toBe('s3xx');
    expect(getStatusClass(399)).toBe('s3xx');
  });

  it('returns s4xx for 4xx statuses', () => {
    expect(getStatusClass(400)).toBe('s4xx');
    expect(getStatusClass(404)).toBe('s4xx');
    expect(getStatusClass(499)).toBe('s4xx');
  });

  it('returns s5xx for 5xx statuses', () => {
    expect(getStatusClass(500)).toBe('s5xx');
    expect(getStatusClass(503)).toBe('s5xx');
    expect(getStatusClass(599)).toBe('s5xx');
  });
});

describe('formatHeaders', () => {
  it('returns empty string for empty headers', () => {
    expect(formatHeaders({})).toBe('');
  });

  it('formats a single header', () => {
    expect(formatHeaders({ 'Content-Type': 'application/json' })).toBe('Content-Type: application/json');
  });

  it('formats multiple headers joined by newlines', () => {
    const result = formatHeaders({
      'Content-Type': 'application/json',
      'Authorization': 'Bearer token123',
    });
    expect(result).toBe('Content-Type: application/json\nAuthorization: Bearer token123');
  });
});

describe('formatBody', () => {
  it('returns empty string for undefined body', () => {
    expect(formatBody(undefined)).toBe('');
  });

  it('returns empty string for empty string body', () => {
    expect(formatBody('')).toBe('');
  });

  it('pretty-prints valid JSON', () => {
    const result = formatBody('{"a":1,"b":2}');
    expect(result).toBe(JSON.stringify({ a: 1, b: 2 }, null, 2));
  });

  it('returns raw string for invalid JSON', () => {
    expect(formatBody('not json at all')).toBe('not json at all');
  });

  it('returns raw string for partial JSON', () => {
    expect(formatBody('{"a":')).toBe('{"a":');
  });
});

describe('formatErrorRate', () => {
  it('formats 0 as 0%', () => {
    expect(formatErrorRate(0)).toBe('0%');
  });

  it('formats 0.5 as 50%', () => {
    expect(formatErrorRate(0.5)).toBe('50%');
  });

  it('formats 1 as 100%', () => {
    expect(formatErrorRate(1)).toBe('100%');
  });

  it('rounds fractional percentages', () => {
    expect(formatErrorRate(0.333)).toBe('33%');
    expect(formatErrorRate(0.666)).toBe('67%');
  });
});
