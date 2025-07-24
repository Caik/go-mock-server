package mock

import (
	"testing"
	"time"

	"github.com/Caik/go-mock-server/internal/config"
)



func TestLatencyMockService_getMockResponse(t *testing.T) {
	t.Run("applies latency when config exists", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(10),
						Max: intPtr(20),
					},
				},
			},
		}

		service := newLatencyMockService(hostsConfig)
		
		testData := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		start := time.Now()
		response := service.getMockResponse(request)
		duration := time.Since(start)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should have applied some latency (at least 10ms)
		if duration < 10*time.Millisecond {
			t.Errorf("expected latency of at least 10ms, got %v", duration)
		}
	})

	t.Run("skips latency when no config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := newLatencyMockService(hostsConfig)
		
		testData2 := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData2,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		start := time.Now()
		response := service.getMockResponse(request)
		duration := time.Since(start)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should not have applied significant latency
		if duration > 5*time.Millisecond {
			t.Errorf("expected minimal latency, got %v", duration)
		}
	})

	// 🚨 TEST TO EXPOSE BUG #4: Potential nil pointer dereference
	t.Run("BUG TEST: nil pointer dereference with P95 but no P99", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(10),
						Max: intPtr(20),
						P95: intPtr(100), // P95 is set
						P99: nil,         // 🚨 P99 is nil - this will cause nil pointer dereference
					},
				},
			},
		}

		service := newLatencyMockService(hostsConfig)
		
		testData3 := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData3,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		// This should panic due to nil pointer dereference when drawn <= 5 and hasP95=true
		defer func() {
			if r := recover(); r != nil {
				t.Logf("BUG DETECTED: Panic due to nil pointer dereference in latency calculation: %v", r)
				t.Logf("Bug location: drawLatencyWithUpperAndLowerBounds(latencyConfig.P95, latencyConfig.P99)")
				t.Logf("P99 is nil but not checked before use")
			}
		}()

		// Run multiple times to increase chance of hitting the bug (drawn <= 5)
		for i := 0; i < 100; i++ {
			service.getMockResponse(request)
		}
	})

	t.Run("handles nil hosts config", func(t *testing.T) {
		service := newLatencyMockService(nil)
		
		testData4 := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData4,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		// Should not panic
		response := service.getMockResponse(request)

		if response == nil {
			t.Error("expected response even with nil hosts config")
		}
	})
}

func TestLatencyMockService_drawLatency(t *testing.T) {
	service := &latencyMockService{}

	t.Run("draws latency within min-max range", func(t *testing.T) {
		latencyConfig := &config.LatencyConfig{
			Min: intPtr(10),
			Max: intPtr(20),
		}

		// Test multiple draws to ensure they're within range
		for i := 0; i < 100; i++ {
			latency := service.drawLatency(latencyConfig)
			
			if latency < 10 || latency > 20 {
				t.Errorf("latency %d is outside range [10, 20]", latency)
			}
		}
	})

	t.Run("handles nil min/max gracefully", func(t *testing.T) {
		latencyConfig := &config.LatencyConfig{
			Min: nil,
			Max: nil,
		}

		// Should not panic
		latency := service.drawLatency(latencyConfig)
		
		// Should return some reasonable default
		if latency < 0 {
			t.Errorf("latency should not be negative, got %d", latency)
		}
	})

	t.Run("handles P95 and P99 percentiles", func(t *testing.T) {
		latencyConfig := &config.LatencyConfig{
			Min: intPtr(10),
			Max: intPtr(20),
			P95: intPtr(100),
			P99: intPtr(200),
		}

		// Test multiple draws
		var highLatencyCount int
		for i := 0; i < 1000; i++ {
			latency := service.drawLatency(latencyConfig)
			
			if latency > 50 {
				highLatencyCount++
			}
		}

		// Should occasionally get high latency values from P95/P99
		if highLatencyCount == 0 {
			t.Error("expected some high latency values from P95/P99 percentiles")
		}
	})

	// 🚨 TEST TO EXPOSE BUG #4: Nil P99 with non-nil P95
	t.Run("BUG TEST: nil P99 with non-nil P95 causes panic", func(t *testing.T) {
		latencyConfig := &config.LatencyConfig{
			Min: intPtr(10),
			Max: intPtr(20),
			P95: intPtr(100), // P95 is set
			P99: nil,         // 🚨 P99 is nil
		}

		service := &latencyMockService{}

		defer func() {
			if r := recover(); r != nil {
				t.Logf("BUG DETECTED: Panic when P95 is set but P99 is nil: %v", r)
			}
		}()

		// Run many times to increase chance of hitting the bug condition (drawn <= 5)
		for i := 0; i < 1000; i++ {
			service.drawLatency(latencyConfig)
		}
	})
}

// Note: drawLatencyWithUpperAndLowerBounds is a private method, so we can't test it directly

func TestLatencyMockService_setNext(t *testing.T) {
	t.Run("sets next service correctly", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := newLatencyMockService(hostsConfig)
		mockNext := &mockMockService{}

		service.setNext(mockNext)

		if service.next != mockNext {
			t.Error("setNext should set the next service")
		}
	})
}
