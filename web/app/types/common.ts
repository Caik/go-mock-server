// Shared types used across multiple pages

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
export type StatusCategory = '2xx' | '3xx' | '4xx' | '5xx';

export const ALL_HTTP_METHODS: HttpMethod[] = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'];
export const ALL_STATUS_CATEGORIES: StatusCategory[] = ['2xx', '3xx', '4xx', '5xx'];

