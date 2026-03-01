package controller

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/service/traffic"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

func newTestTrafficService(bufferSize int) *traffic.TrafficLogService {
	return traffic.NewTrafficLogService(&config.AppArguments{
		TrafficLogBufferSize: bufferSize,
	})
}

func TestNewTrafficController(t *testing.T) {
	t.Run("creates controller with service", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		if controller == nil {
			t.Fatal("expected non-nil controller")
		}
	})

	t.Run("creates controller with nil service", func(t *testing.T) {
		controller := NewTrafficController(nil)

		if controller == nil {
			t.Fatal("expected non-nil controller even with nil service")
		}
	})
}

func TestTrafficController_handleTrafficStream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 503 when traffic logging is disabled", func(t *testing.T) {
		controller := NewTrafficController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic", nil)
		c.Set(util.UuidKey, "test-uuid")

		controller.handleTrafficStream(c)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("returns 400 for invalid status code filter", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?status=abc", nil)
		c.Set(util.UuidKey, "test-uuid")

		controller.handleTrafficStream(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns 400 for invalid matched filter", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?matched=invalid", nil)
		c.Set(util.UuidKey, "test-uuid")

		controller.handleTrafficStream(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns 400 for invalid status code range", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?status=999", nil)
		c.Set(util.UuidKey, "test-uuid")

		controller.handleTrafficStream(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestTrafficController_parseFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("parses hosts filter", func(t *testing.T) {
		controller := NewTrafficController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?hosts=example.com,api.test.com", nil)

		filters, err := controller.parseFilters(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filters.Hosts) != 2 {
			t.Errorf("expected 2 hosts, got %d", len(filters.Hosts))
		}
	})

	t.Run("parses status codes filter", func(t *testing.T) {
		controller := NewTrafficController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?status=200,404,500", nil)

		filters, err := controller.parseFilters(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filters.StatusCodes) != 3 {
			t.Errorf("expected 3 status codes, got %d", len(filters.StatusCodes))
		}
	})
}

func TestTrafficController_parseFilters_matched(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("parses matched=true filter", func(t *testing.T) {
		controller := NewTrafficController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?matched=true", nil)

		filters, err := controller.parseFilters(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filters.Matched == nil || !*filters.Matched {
			t.Error("expected matched to be true")
		}
	})

	t.Run("parses matched=false filter", func(t *testing.T) {
		controller := NewTrafficController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic?matched=false", nil)

		filters, err := controller.parseFilters(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filters.Matched == nil || *filters.Matched {
			t.Error("expected matched to be false")
		}
	})

	t.Run("returns nil filters when no params", func(t *testing.T) {
		controller := NewTrafficController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/traffic", nil)

		filters, err := controller.parseFilters(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filters != nil {
			t.Error("expected nil filters for empty query")
		}
	})
}

func TestTrafficController_SSEStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("streams catch-up entries", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		// Add some entries before streaming
		service.Capture(traffic.TrafficEntry{UUID: "entry-1", Request: traffic.TrafficRequest{Host: "test.com"}})
		service.Capture(traffic.TrafficEntry{UUID: "entry-2", Request: traffic.TrafficRequest{Host: "test.com"}})

		// Create router with the endpoint
		router := gin.New()
		router.GET("/api/v1/traffic", func(c *gin.Context) {
			c.Set(util.UuidKey, "test-uuid")
			controller.handleTrafficStream(c)
		})

		// Create request with cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/traffic", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		// Run in goroutine since SSE blocks
		done := make(chan bool)
		go func() {
			router.ServeHTTP(w, req)
			done <- true
		}()

		// Wait a bit for catch-up entries to be written, then cancel
		time.Sleep(50 * time.Millisecond)
		cancel()
		<-done

		// Check that we got catch-up data
		result := w.Body.String()
		if !strings.Contains(result, "entry-1") || !strings.Contains(result, "entry-2") {
			t.Errorf("expected catch-up entries in response, got: %s", result)
		}
	})

	t.Run("sets SSE headers", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		router := gin.New()
		router.GET("/api/v1/traffic", func(c *gin.Context) {
			c.Set(util.UuidKey, "test-uuid")
			controller.handleTrafficStream(c)
		})

		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/traffic", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		done := make(chan bool)
		go func() {
			router.ServeHTTP(w, req)
			done <- true
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		<-done

		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Errorf("expected Content-Type text/event-stream, got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("filters catch-up entries", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		// Add entries with different hosts
		service.Capture(traffic.TrafficEntry{UUID: "match", Request: traffic.TrafficRequest{Host: "example.com"}})
		service.Capture(traffic.TrafficEntry{UUID: "nomatch", Request: traffic.TrafficRequest{Host: "other.com"}})

		router := gin.New()
		router.GET("/api/v1/traffic", func(c *gin.Context) {
			c.Set(util.UuidKey, "test-uuid")
			controller.handleTrafficStream(c)
		})

		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/traffic?hosts=example.com", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		done := make(chan bool)
		go func() {
			router.ServeHTTP(w, req)
			done <- true
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		<-done

		result := w.Body.String()
		if !strings.Contains(result, "match") {
			t.Error("expected filtered match entry")
		}
		if strings.Contains(result, "nomatch") {
			t.Error("should not contain filtered-out entry")
		}
	})
}

func TestTrafficController_SSEFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sends data in SSE format", func(t *testing.T) {
		service := newTestTrafficService(10)
		controller := NewTrafficController(service)

		service.Capture(traffic.TrafficEntry{UUID: "test-entry"})

		router := gin.New()
		router.GET("/api/v1/traffic", func(c *gin.Context) {
			c.Set(util.UuidKey, "test-uuid")
			controller.handleTrafficStream(c)
		})

		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/traffic", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		done := make(chan bool)
		go func() {
			router.ServeHTTP(w, req)
			done <- true
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		<-done

		// Parse SSE response
		scanner := bufio.NewScanner(strings.NewReader(w.Body.String()))
		foundData := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				foundData = true
				jsonData := strings.TrimPrefix(line, "data: ")
				var entry traffic.TrafficEntry
				if err := json.Unmarshal([]byte(jsonData), &entry); err != nil {
					t.Errorf("failed to unmarshal SSE data: %v", err)
				}
				if entry.UUID != "test-entry" {
					t.Errorf("expected UUID 'test-entry', got '%s'", entry.UUID)
				}
			}
		}
		if !foundData {
			t.Error("expected SSE data lines in response")
		}
	})
}
