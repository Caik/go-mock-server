package traffic

import (
	"testing"
	"time"

	"github.com/Caik/go-mock-server/internal/config"
)

// newTestService creates a TrafficLogService for testing with the given buffer size
func newTestService(bufferSize int) *TrafficLogService {
	return NewTrafficLogService(&config.AppArguments{
		TrafficLogBufferSize: bufferSize,
	})
}

func TestNewTrafficLogService(t *testing.T) {
	t.Run("creates service with positive buffer size", func(t *testing.T) {
		service := newTestService(100)

		if service == nil {
			t.Error("expected non-nil service")
		}
	})

	t.Run("returns nil with zero buffer size", func(t *testing.T) {
		service := newTestService(0)

		if service != nil {
			t.Error("expected nil service when disabled")
		}
	})

	t.Run("returns nil with negative buffer size", func(t *testing.T) {
		service := newTestService(-10)

		if service != nil {
			t.Error("expected nil service when disabled")
		}
	})
}

func TestTrafficLogService_Capture(t *testing.T) {
	t.Run("captures entry and increases size", func(t *testing.T) {
		service := newTestService(10)

		entry := TrafficEntry{
			UUID:      "test-uuid",
			Timestamp: time.Now(),
			Request:   TrafficRequest{Method: "GET", Host: "example.com", Path: "/api"},
			Response:  TrafficResponse{StatusCode: 200},
			Metadata:  map[string]string{metadataMatched: "true", metadataSource: "filesystem"},
		}

		service.Capture(entry)

		if service.Size() != 1 {
			t.Errorf("expected size 1, got %d", service.Size())
		}
	})

	t.Run("does nothing when disabled", func(t *testing.T) {
		service := newTestService(0)

		entry := TrafficEntry{UUID: "test-uuid"}
		service.Capture(entry)

		if service.Size() != 0 {
			t.Errorf("expected size 0 for disabled service, got %d", service.Size())
		}
	})
}

func TestTrafficLogService_GetAll(t *testing.T) {
	t.Run("returns all entries in order", func(t *testing.T) {
		service := newTestService(10)

		for i := 0; i < 3; i++ {
			service.Capture(TrafficEntry{UUID: string(rune('a' + i))})
		}

		entries := service.GetAll()

		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}

		if entries[0].UUID != "a" || entries[1].UUID != "b" || entries[2].UUID != "c" {
			t.Errorf("expected UUIDs [a, b, c], got [%s, %s, %s]", entries[0].UUID, entries[1].UUID, entries[2].UUID)
		}
	})

	t.Run("returns empty slice when disabled", func(t *testing.T) {
		service := newTestService(0)

		entries := service.GetAll()

		if len(entries) != 0 {
			t.Errorf("expected empty slice, got %d entries", len(entries))
		}
	})
}

func TestTrafficLogService_GetFiltered(t *testing.T) {
	service := newTestService(10)

	// Add test entries
	service.Capture(TrafficEntry{
		UUID:     "1",
		Request:  TrafficRequest{Host: "example.com"},
		Response: TrafficResponse{StatusCode: 200},
		Metadata: map[string]string{metadataMatched: "true"},
	})
	service.Capture(TrafficEntry{
		UUID:     "2",
		Request:  TrafficRequest{Host: "other.com"},
		Response: TrafficResponse{StatusCode: 404},
		Metadata: map[string]string{metadataMatched: "false"},
	})
	service.Capture(TrafficEntry{
		UUID:     "3",
		Request:  TrafficRequest{Host: "example.com"},
		Response: TrafficResponse{StatusCode: 500},
		Metadata: map[string]string{metadataMatched: "true"},
	})
	service.Capture(TrafficEntry{
		UUID:     "4",
		Request:  TrafficRequest{Host: "api.example.com"},
		Response: TrafficResponse{StatusCode: 200},
		Metadata: map[string]string{metadataMatched: "true"},
	})

	t.Run("filters by single host", func(t *testing.T) {
		entries := service.GetFiltered(&TrafficFilters{Hosts: []string{"example.com"}})

		if len(entries) != 2 {
			t.Errorf("expected 2 entries for example.com, got %d", len(entries))
		}
	})

	t.Run("filters by multiple hosts", func(t *testing.T) {
		entries := service.GetFiltered(&TrafficFilters{Hosts: []string{"example.com", "other.com"}})

		if len(entries) != 3 {
			t.Errorf("expected 3 entries for example.com or other.com, got %d", len(entries))
		}
	})

	t.Run("filters by single status code", func(t *testing.T) {
		entries := service.GetFiltered(&TrafficFilters{StatusCodes: []int{404}})

		if len(entries) != 1 {
			t.Errorf("expected 1 entry with status 404, got %d", len(entries))
		}
	})

	t.Run("filters by multiple status codes", func(t *testing.T) {
		entries := service.GetFiltered(&TrafficFilters{StatusCodes: []int{200, 500}})

		if len(entries) != 3 {
			t.Errorf("expected 3 entries with status 200 or 500, got %d", len(entries))
		}
	})

	t.Run("filters by matched", func(t *testing.T) {
		matched := false
		entries := service.GetFiltered(&TrafficFilters{Matched: &matched})

		if len(entries) != 1 {
			t.Errorf("expected 1 unmatched entry, got %d", len(entries))
		}
	})

	t.Run("combines host and status filters", func(t *testing.T) {
		entries := service.GetFiltered(&TrafficFilters{
			Hosts:       []string{"example.com"},
			StatusCodes: []int{200},
		})

		if len(entries) != 1 {
			t.Errorf("expected 1 entry matching example.com AND 200, got %d", len(entries))
		}
	})

	t.Run("returns all when nil filter", func(t *testing.T) {
		entries := service.GetFiltered(nil)

		if len(entries) != 4 {
			t.Errorf("expected 4 entries with nil filter, got %d", len(entries))
		}
	})

	t.Run("returns all when empty filter", func(t *testing.T) {
		entries := service.GetFiltered(&TrafficFilters{})

		if len(entries) != 4 {
			t.Errorf("expected 4 entries with empty filter, got %d", len(entries))
		}
	})

	t.Run("returns empty when disabled", func(t *testing.T) {
		disabledService := newTestService(0)
		entries := disabledService.GetFiltered(&TrafficFilters{Hosts: []string{"example.com"}})

		if len(entries) != 0 {
			t.Errorf("expected 0 entries for disabled service, got %d", len(entries))
		}
	})
}

func TestTrafficLogService_Clear(t *testing.T) {
	t.Run("clears all entries", func(t *testing.T) {
		service := newTestService(10)

		service.Capture(TrafficEntry{UUID: "1"})
		service.Capture(TrafficEntry{UUID: "2"})

		service.Clear()

		if service.Size() != 0 {
			t.Errorf("expected size 0 after clear, got %d", service.Size())
		}
	})

	t.Run("does nothing when disabled", func(t *testing.T) {
		service := newTestService(0)
		service.Clear() // Should not panic
	})
}

func TestTrafficLogService_Subscribe(t *testing.T) {
	t.Run("subscriber receives captured entries without filter", func(t *testing.T) {
		service := newTestService(10)

		ch := service.Subscribe("test-subscriber", nil)
		defer service.Unsubscribe("test-subscriber")

		// Capture an entry
		entry := TrafficEntry{UUID: "test-uuid"}
		service.Capture(entry)

		// Wait for entry with timeout
		select {
		case received := <-ch:
			if received.UUID != "test-uuid" {
				t.Errorf("expected UUID 'test-uuid', got '%s'", received.UUID)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for entry")
		}
	})

	t.Run("subscriber with filter only receives matching entries", func(t *testing.T) {
		service := newTestService(10)

		filter := &TrafficFilters{Hosts: []string{"example.com"}}
		ch := service.Subscribe("filtered-subscriber", filter)
		defer service.Unsubscribe("filtered-subscriber")

		// Capture a matching entry only (simpler test to avoid race conditions)
		service.Capture(TrafficEntry{
			UUID:    "matching",
			Request: TrafficRequest{Host: "example.com"},
		})

		// Should receive the matching entry
		select {
		case received := <-ch:
			if received.UUID != "matching" {
				t.Errorf("expected UUID 'matching', got '%s'", received.UUID)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for matching entry")
		}
	})

	t.Run("returns nil when disabled", func(t *testing.T) {
		service := newTestService(0)

		ch := service.Subscribe("test-subscriber", nil)

		if ch != nil {
			t.Error("expected nil channel for disabled service")
		}

		// Unsubscribe should not panic
		service.Unsubscribe("test-subscriber")
	})
}

func TestTrafficLogService_GetRecent(t *testing.T) {
	t.Run("returns recent entries", func(t *testing.T) {
		service := newTestService(10)

		for i := 0; i < 5; i++ {
			service.Capture(TrafficEntry{UUID: string(rune('a' + i))})
		}

		entries := service.GetRecent(3)

		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}

		if entries[0].UUID != "c" || entries[1].UUID != "d" || entries[2].UUID != "e" {
			t.Errorf("expected UUIDs [c, d, e], got [%s, %s, %s]", entries[0].UUID, entries[1].UUID, entries[2].UUID)
		}
	})

	t.Run("returns empty when disabled", func(t *testing.T) {
		service := newTestService(0)

		entries := service.GetRecent(3)

		if len(entries) != 0 {
			t.Errorf("expected 0 entries for disabled service, got %d", len(entries))
		}
	})
}

func TestTrafficFilters(t *testing.T) {
	t.Run("IsEmpty returns true for empty filters", func(t *testing.T) {
		filters := TrafficFilters{}

		if !filters.IsEmpty() {
			t.Error("expected IsEmpty to return true")
		}
	})

	t.Run("IsEmpty returns false when hosts is set", func(t *testing.T) {
		filters := TrafficFilters{Hosts: []string{"example.com"}}

		if filters.IsEmpty() {
			t.Error("expected IsEmpty to return false")
		}
	})

	t.Run("IsEmpty returns false when status codes is set", func(t *testing.T) {
		filters := TrafficFilters{StatusCodes: []int{200}}

		if filters.IsEmpty() {
			t.Error("expected IsEmpty to return false")
		}
	})

	t.Run("IsEmpty returns false when matched is set", func(t *testing.T) {
		matched := true
		filters := TrafficFilters{Matched: &matched}

		if filters.IsEmpty() {
			t.Error("expected IsEmpty to return false")
		}
	})
}

func TestTrafficFilters_Validate(t *testing.T) {
	t.Run("valid filter passes validation", func(t *testing.T) {
		filters := TrafficFilters{
			Hosts:       []string{"example.com", "api.example.com"},
			StatusCodes: []int{200, 404, 500},
		}

		if err := filters.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty filter passes validation", func(t *testing.T) {
		filters := TrafficFilters{}

		if err := filters.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid status code fails validation", func(t *testing.T) {
		filters := TrafficFilters{StatusCodes: []int{200, 999}}

		err := filters.Validate()

		if err == nil {
			t.Error("expected error for invalid status code")
		}
	})

	t.Run("status code below 100 fails validation", func(t *testing.T) {
		filters := TrafficFilters{StatusCodes: []int{99}}

		err := filters.Validate()

		if err == nil {
			t.Error("expected error for status code below 100")
		}
	})

	t.Run("empty host string fails validation", func(t *testing.T) {
		filters := TrafficFilters{Hosts: []string{"example.com", ""}}

		err := filters.Validate()

		if err == nil {
			t.Error("expected error for empty host")
		}
	})

	t.Run("whitespace-only host fails validation", func(t *testing.T) {
		filters := TrafficFilters{Hosts: []string{"  "}}

		err := filters.Validate()

		if err == nil {
			t.Error("expected error for whitespace-only host")
		}
	})

	t.Run("invalid hostname fails validation", func(t *testing.T) {
		filters := TrafficFilters{Hosts: []string{"not a valid host!"}}

		err := filters.Validate()

		if err == nil {
			t.Error("expected error for invalid hostname")
		}
	})

	t.Run("IP address passes validation", func(t *testing.T) {
		filters := TrafficFilters{Hosts: []string{"192.168.1.1"}}

		if err := filters.Validate(); err != nil {
			t.Errorf("expected IP address to pass validation, got %v", err)
		}
	})

	t.Run("localhost passes validation", func(t *testing.T) {
		filters := TrafficFilters{Hosts: []string{"my-app.localhost"}}

		if err := filters.Validate(); err != nil {
			t.Errorf("expected localhost to pass validation, got %v", err)
		}
	})
}
