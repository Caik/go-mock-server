// Traffic log types (matching Go backend TrafficEntry)

export interface TrafficEntry {
  uuid: string;
  timestamp: string;
  request: TrafficRequest;
  response: TrafficResponse;
  metadata?: Record<string, string>;
}

export interface TrafficRequest {
  method: string;
  host: string;
  path: string;
  query?: string;
  headers?: Record<string, string>;
}

export interface TrafficResponse {
  status_code: number;
  content_type?: string;
  body_size: number;
  latency_ms: number;
}

