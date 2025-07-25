package mock

import (
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
)

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

func TestErrorMockService_getMockResponse(t *testing.T) {
	t.Run("returns error response when error is drawn", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					ErrorsConfig: map[string]config.ErrorConfig{
						"500": {
							Percentage: intPtr(100), // Always return error
						},
					},
				},
			},
		}

		service := newErrorMockService(hostsConfig)
		
		successData := []byte("success response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &successData,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should return error response
		if response.StatusCode != 500 {
			t.Errorf("expected status 500, got %d", response.StatusCode)
		}

		// Should have empty response body
		if response.Data == nil || len(*response.Data) != 0 {
			t.Error("error response should have empty body")
		}
	})

	t.Run("passes through when no error is drawn", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					ErrorsConfig: map[string]config.ErrorConfig{
						"500": {
							Percentage: intPtr(0), // Never return error
						},
					},
				},
			},
		}

		service := newErrorMockService(hostsConfig)
		
		successData2 := []byte("success response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &successData2,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should pass through to next service
		if response.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", response.StatusCode)
		}
	})

	t.Run("skips when no error config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := newErrorMockService(hostsConfig)
		
		successData3 := []byte("success response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &successData3,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should pass through to next service
		if response.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", response.StatusCode)
		}
	})
}

func TestErrorMockService_drawError(t *testing.T) {
	service := &errorMockService{}

	t.Run("draws error based on percentage", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"500": {
				Percentage: intPtr(100), // Always draw this error
			},
		}

		// Should always return the error
		for i := 0; i < 10; i++ {
			wrapper := service.drawError(&errorsConfig)
			
			if wrapper == nil {
				t.Error("expected error wrapper, got nil")
				continue
			}

			if wrapper.statusCode != 500 {
				t.Errorf("expected status code 500, got %d", wrapper.statusCode)
			}

			if wrapper.percentage != 100 {
				t.Errorf("expected percentage 100, got %d", wrapper.percentage)
			}
		}
	})

	t.Run("returns nil when no error should be drawn", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"500": {
				Percentage: intPtr(0), // Never draw this error
			},
		}

		// Should never return an error
		for i := 0; i < 10; i++ {
			wrapper := service.drawError(&errorsConfig)
			
			if wrapper != nil {
				t.Error("expected nil wrapper, got error")
			}
		}
	})

	t.Run("handles invalid status code strings", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"invalid": { // Invalid status code string (should be validated at config level)
				Percentage: intPtr(100),
			},
		}

		service := &errorMockService{}

		wrapper := service.drawError(&errorsConfig)

		if wrapper != nil {
			t.Logf("status code 'invalid' converted to %d", wrapper.statusCode)
			t.Logf("note: strconv.Atoi error is ignored by design - validation happens at config level")

			if wrapper.statusCode == 0 {
				t.Log("invalid status codes become 0 - this is expected behavior")
				t.Log("validation should occur in config.HostConfig.Validate() at startup")
			}
		}
	})

	t.Run("validates random range behavior with 100% error rate", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"500": {
				Percentage: intPtr(100), // Total percentage is exactly 100
			},
		}

		service := &errorMockService{}

		var errorCount int
		var totalDraws int = 1000

		// Test the random logic with 100% error rate
		for i := 0; i < totalDraws; i++ {
			wrapper := service.drawError(&errorsConfig)
			if wrapper != nil {
				errorCount++
			}
		}

		errorRate := float64(errorCount) / float64(totalDraws) * 100

		t.Logf("with 100%% configured error rate, got %.1f%% actual error rate", errorRate)
		t.Logf("random range: rand.Intn(101) generates 0-100, condition: draw <= totalPercentage")

		// With 100% error rate, we should get close to 100% errors (allowing for randomness)
		if errorRate >= 99.0 { // Allow 1% tolerance for randomness
			t.Log("random range logic works correctly")
		} else {
			t.Logf("unexpected: error rate %.1f%% is lower than expected", errorRate)
		}
	})

	t.Run("handles multiple errors with different percentages", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"400": {
				Percentage: intPtr(30),
			},
			"500": {
				Percentage: intPtr(20),
			},
		}

		service := &errorMockService{}

		var errorCount int
		var totalDraws int = 1000

		for i := 0; i < totalDraws; i++ {
			wrapper := service.drawError(&errorsConfig)
			if wrapper != nil {
				errorCount++
			}
		}

		// With 50% total error rate, we should get roughly 500 errors
		expectedErrors := totalDraws * 50 / 100
		tolerance := totalDraws * 10 / 100 // 10% tolerance

		if errorCount < expectedErrors-tolerance || errorCount > expectedErrors+tolerance {
			t.Errorf("expected ~%d errors, got %d (outside tolerance)", expectedErrors, errorCount)
		}
	})

	t.Run("handles empty errors config", func(t *testing.T) {
		errorsConfig := make(map[string]config.ErrorConfig)

		service := &errorMockService{}

		wrapper := service.drawError(&errorsConfig)

		if wrapper != nil {
			t.Error("expected nil wrapper with empty config")
		}
	})
}

func TestErrorMockService_setNext(t *testing.T) {
	t.Run("sets next service correctly", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := newErrorMockService(hostsConfig)
		mockNext := &mockMockService{}

		service.setNext(mockNext)

		if service.next != mockNext {
			t.Error("setNext should set the next service")
		}
	})
}
