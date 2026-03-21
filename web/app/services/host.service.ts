import type { HostConfig, LatencyConfig, StatusConfig, UriConfig } from '~/types/host';

const API_BASE_URL = import.meta.env.DEV ? 'http://localhost:9090' : '';

// --- API response types (mirrors backend config package) ---

interface ApiLatencyConfig {
  min: number;
  max: number;
  p95?: number | null;
  p99?: number | null;
}

interface ApiStatusConfig {
  percentage: number;
  latency?: ApiLatencyConfig | null;
}

interface ApiUriConfig {
  latency?: ApiLatencyConfig | null;
  statuses?: Record<string, ApiStatusConfig> | null;
}

interface ApiHostConfig {
  latency?: ApiLatencyConfig | null;
  statuses?: Record<string, ApiStatusConfig> | null;
  uris?: Record<string, ApiUriConfig> | null;
}

interface ApiHostsConfigData {
  hosts: Record<string, ApiHostConfig>;
}

interface ApiResponse<T> {
  status: string;
  data: T;
}

// --- Mappers ---

function toLatencyConfig(api: ApiLatencyConfig): LatencyConfig {
  return {
    min: api.min,
    max: api.max,
    ...(api.p95 != null && { p95: api.p95 }),
    ...(api.p99 != null && { p99: api.p99 }),
  };
}

function toStatusesConfig(api: Record<string, ApiStatusConfig>): Record<string, StatusConfig> {
  const result: Record<string, StatusConfig> = {};
  for (const [code, cfg] of Object.entries(api)) {
    result[code] = {
      percentage: cfg.percentage,
      ...(cfg.latency && { latency: toLatencyConfig(cfg.latency) }),
    };
  }
  return result;
}

function toUriConfig(api: ApiUriConfig): UriConfig {
  return {
    ...(api.latency && { latency: toLatencyConfig(api.latency) }),
    ...(api.statuses && Object.keys(api.statuses).length > 0 && { statuses: toStatusesConfig(api.statuses) }),
  };
}

function toHostConfig(hostname: string, api: ApiHostConfig): HostConfig {
  return {
    hostname,
    ...(api.latency && { latency: toLatencyConfig(api.latency) }),
    ...(api.statuses && Object.keys(api.statuses).length > 0 && { statuses: toStatusesConfig(api.statuses) }),
    ...(api.uris && Object.keys(api.uris).length > 0 && {
      uris: Object.fromEntries(
        Object.entries(api.uris).map(([pattern, cfg]) => [pattern, toUriConfig(cfg)])
      ),
    }),
  };
}

// --- Public API ---

export async function getHosts(): Promise<HostConfig[]> {
  const response = await fetch(`${API_BASE_URL}/api/v1/config/hosts`);

  if (!response.ok) {
    throw new Error(`Failed to fetch hosts: ${response.statusText}`);
  }

  const data: ApiResponse<ApiHostsConfigData> = await response.json();

  return Object.entries(data.data.hosts).map(([hostname, cfg]) => toHostConfig(hostname, cfg));
}

interface LatencyPayload {
  min: number;
  max: number;
  p95?: number;
  p99?: number;
}

interface ErrorPayload {
  percentage: number;
}

interface UriPayload {
  latency?: LatencyPayload;
  statuses?: Record<string, ErrorPayload>;
}

export interface HostSaveData {
  host: string;
  latency?: LatencyPayload;
  statuses?: Record<string, ErrorPayload>;
  uris?: Record<string, UriPayload>;
}

export async function saveHost(payload: HostSaveData): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/config/hosts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to save host: ${response.statusText}`);
  }
}

export async function deleteHost(hostname: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/config/hosts/${encodeURIComponent(hostname)}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to delete host: ${response.statusText}`);
  }
}
