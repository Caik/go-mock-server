package traffic

import (
	"testing"
	"time"
)

const metadataSource = "Source"

func TestNewTrafficEntry(t *testing.T) {
	t.Run("creates entry with uuid and timestamp", func(t *testing.T) {
		before := time.Now()
		entry := NewTrafficEntry("test-uuid-123")
		after := time.Now()

		if entry == nil {
			t.Fatal("expected non-nil entry")
		}

		if entry.UUID != "test-uuid-123" {
			t.Errorf("expected UUID 'test-uuid-123', got '%s'", entry.UUID)
		}

		if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
			t.Errorf("expected timestamp between %v and %v, got %v", before, after, entry.Timestamp)
		}
	})

	t.Run("creates entry with empty uuid", func(t *testing.T) {
		entry := NewTrafficEntry("")

		if entry == nil {
			t.Fatal("expected non-nil entry")
		}

		if entry.UUID != "" {
			t.Errorf("expected empty UUID, got '%s'", entry.UUID)
		}
	})
}

func TestTrafficEntry_JSONSerialization(t *testing.T) {
	t.Run("TrafficRequest has correct json tags", func(t *testing.T) {
		req := TrafficRequest{
			Method:  "GET",
			Host:    "example.com",
			Path:    "/api/users",
			Query:   "id=123",
			Headers: map[string]string{"Accept": "application/json"},
		}

		if req.Method != "GET" {
			t.Errorf("expected Method 'GET', got '%s'", req.Method)
		}
	})

	t.Run("TrafficResponse has correct fields", func(t *testing.T) {
		resp := TrafficResponse{
			StatusCode:  200,
			ContentType: "application/json",
			BodySize:    1024,
			LatencyMs:   150,
		}

		if resp.StatusCode != 200 {
			t.Errorf("expected StatusCode 200, got %d", resp.StatusCode)
		}
	})

	t.Run("TrafficEntry Metadata stores key-value pairs", func(t *testing.T) {
		entry := TrafficEntry{
			Metadata: map[string]string{
				metadataMatched: "true",
				metadataSource:  "filesystem",
				"Path":    "/mocks/example.com/api.get",
			},
		}

		if entry.Metadata[metadataMatched] != "true" {
			t.Errorf("expected Matched 'true', got '%s'", entry.Metadata[metadataMatched])
		}

		if entry.Metadata[metadataSource] != "filesystem" {
			t.Errorf("expected Source 'filesystem', got '%s'", entry.Metadata[metadataSource])
		}
	})

	t.Run("TrafficEntry Metadata is nil when not matched", func(t *testing.T) {
		entry := TrafficEntry{
			Metadata: map[string]string{metadataMatched: "false"},
		}

		if entry.Metadata[metadataMatched] != "false" {
			t.Errorf("expected Matched 'false', got '%s'", entry.Metadata[metadataMatched])
		}

		if entry.Metadata[metadataSource] != "" {
			t.Errorf("expected empty Source, got '%s'", entry.Metadata[metadataSource])
		}
	})
}
