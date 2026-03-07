package traffic

import (
	"time"
)

// TrafficRequest captures details about the incoming HTTP request
type TrafficRequest struct {
	Method  string            `json:"method"`
	Host    string            `json:"host"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// TrafficResponse captures details about the mock response
type TrafficResponse struct {
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type,omitempty"`
	BodySize    int    `json:"body_size"`
	LatencyMs   int64  `json:"latency_ms"`
}

// TrafficEntry represents a single traffic log entry
type TrafficEntry struct {
	UUID      string            `json:"uuid"`
	Timestamp time.Time         `json:"timestamp"`
	Request   TrafficRequest    `json:"request"`
	Response  TrafficResponse   `json:"response"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewTrafficEntry creates a new TrafficEntry with the current timestamp
func NewTrafficEntry(uuid string) *TrafficEntry {
	return &TrafficEntry{
		UUID:      uuid,
		Timestamp: time.Now(),
	}
}
