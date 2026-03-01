package traffic

import (
	"testing"
	"time"
)

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

	t.Run("TrafficMock has correct fields", func(t *testing.T) {
		mock := TrafficMock{
			Matched: true,
			Source:  "filesystem",
			Path:    "/mocks/example.com/api.get",
		}

		if !mock.Matched {
			t.Error("expected Matched to be true")
		}

		if mock.Source != "filesystem" {
			t.Errorf("expected Source 'filesystem', got '%s'", mock.Source)
		}
	})

	t.Run("TrafficMock with no match has empty source", func(t *testing.T) {
		mock := TrafficMock{
			Matched: false,
		}

		if mock.Matched {
			t.Error("expected Matched to be false")
		}

		if mock.Source != "" {
			t.Errorf("expected empty Source, got '%s'", mock.Source)
		}

		if mock.Path != "" {
			t.Errorf("expected empty Path, got '%s'", mock.Path)
		}
	})
}

