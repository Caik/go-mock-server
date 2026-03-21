// Host configuration types — mirrors backend config.HostsConfig shape

export interface HostConfig {
  hostname: string; // injected from map key by service layer
  latency?: LatencyConfig;
  statuses?: Record<string, StatusConfig>;
  uris?: Record<string, UriConfig>;
}

export interface LatencyConfig {
  min: number;
  max: number;
  p95?: number;
  p99?: number;
}

export interface StatusConfig {
  percentage: number;
  latency?: LatencyConfig;
}

export interface UriConfig {
  latency?: LatencyConfig;
  statuses?: Record<string, StatusConfig>;
}
