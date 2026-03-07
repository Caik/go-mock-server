import type { MockDefinition } from '~/types/mock';

const API_BASE_URL = import.meta.env.DEV ? 'http://localhost:9090' : '';

// API response types
interface ApiMock {
  id: string;
  host: string;
  uri: string;
  method: string;
}

interface ApiResponse<T> {
  status: string;
  data: T;
}

/**
 * Convert API mock to MockDefinition
 */
function toMockDefinition(item: ApiMock): MockDefinition {
  return {
    id: item.id,
    endpoint: item.uri,
    method: item.method,
    host: item.host,
  };
}

/**
 * Get all mock definitions from the API
 */
export async function getMocks(): Promise<MockDefinition[]> {
  const response = await fetch(`${API_BASE_URL}/api/v1/mocks`);

  if (!response.ok) {
    throw new Error(`Failed to fetch mocks: ${response.statusText}`);
  }

  const data: ApiResponse<ApiMock[]> = await response.json();

  return data.data.map(toMockDefinition);
}

interface MockContentResponse {
  body: string;
}

/**
 * Get mock content (response body) for a specific mock by ID
 */
export async function getMockContent(id: string): Promise<string> {
  const response = await fetch(`${API_BASE_URL}/api/v1/mocks/${encodeURIComponent(id)}/content`);

  if (!response.ok) {
    throw new Error(`Failed to fetch mock content: ${response.statusText}`);
  }

  const data: ApiResponse<MockContentResponse> = await response.json();

  return data.data.body;
}

/**
 * Get a single mock definition by ID
 */
export function getMockById(mocks: MockDefinition[], id: string): MockDefinition | undefined {
  return mocks.find((mock) => mock.id === id);
}

export interface MockData {
  host: string;
  uri: string;
  method: string;
  body: string;
}

/**
 * Create a new mock
 */
export async function createMock(mock: MockData): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/mocks`, {
    method: 'POST',
    headers: {
      'Content-Type': 'text/plain',
      'x-mock-host': mock.host,
      'x-mock-uri': mock.uri,
      'x-mock-method': mock.method,
    },
    body: mock.body,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to create mock: ${response.statusText}`);
  }
}

/**
 * Delete an existing mock
 */
export async function deleteMock(mock: MockDefinition): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/mocks`, {
    method: 'DELETE',
    headers: {
      'x-mock-host': mock.host,
      'x-mock-uri': mock.endpoint,
      'x-mock-method': mock.method,
    },
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to delete mock: ${response.statusText}`);
  }
}

/**
 * Update an existing mock
 */
export async function updateMock(id: string, mock: MockData): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/mocks/${encodeURIComponent(id)}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'text/plain',
      'x-mock-host': mock.host,
      'x-mock-uri': mock.uri,
      'x-mock-method': mock.method,
    },
    body: mock.body,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to update mock: ${response.statusText}`);
  }
}

