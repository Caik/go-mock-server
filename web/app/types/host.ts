// Host configuration types — mirrors backend config.HostsConfig shape

export interface HostConfig {
  hostname: string; // injected from map key by service layer
  latency?: LatencyConfig;
  errors?: Record<string, ErrorConfig>;
  uris?: Record<string, UriConfig>;
}

export interface LatencyConfig {
  min: number;
  max: number;
  p95?: number;
  p99?: number;
}

export interface ErrorConfig {
  percentage: number;
  latency?: LatencyConfig;
}

export interface UriConfig {
  latency?: LatencyConfig;
  errors?: Record<string, ErrorConfig>;
}
