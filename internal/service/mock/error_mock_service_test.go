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

	// 🚨 TEST TO EXPOSE BUG #5: Ignoring strconv.Atoi error
	t.Run("BUG TEST: invalid status code string causes incorrect behavior", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"invalid": { // 🚨 Invalid status code string
				Percentage: intPtr(100),
			},
		}

		service := &errorMockService{}

		wrapper := service.drawError(&errorsConfig)

		if wrapper != nil {
			t.Logf("BUG DETECTED: Invalid status code 'invalid' converted to %d", wrapper.statusCode)
			t.Logf("Bug location: strconv.Atoi error is ignored, resulting in statusCode=0")
			
			if wrapper.statusCode == 0 {
				t.Error("BUG CONFIRMED: Invalid status code resulted in statusCode=0")
			}
		}
	})

	// 🚨 TEST TO EXPOSE BUG #6: Off-by-one error in random range
	t.Run("BUG TEST: random range includes 0 which may never match", func(t *testing.T) {
		errorsConfig := map[string]config.ErrorConfig{
			"500": {
				Percentage: intPtr(100), // Total percentage is exactly 100
			},
		}

		service := &errorMockService{}

		// The bug: rand.Intn(101) generates 0-100, but if draw=0 and total=100,
		// the condition draw <= totalPercentage (0 <= 100) will be true
		// However, if percentages are meant to be 1-100, then 0 should never match

		var zeroDrawCount int
		var totalDraws int = 1000

		// We can't directly test the random draw, but we can test the logic
		// by checking if the method handles edge cases correctly

		for i := 0; i < totalDraws; i++ {
			wrapper := service.drawError(&errorsConfig)
			if wrapper != nil {
				// Error was drawn - this should happen 100% of the time with percentage=100
			} else {
				zeroDrawCount++
			}
		}

		t.Logf("BUG ANALYSIS: With 100%% error rate, got %d non-error responses out of %d", 
			zeroDrawCount, totalDraws)
		t.Logf("Bug location: rand.Intn(101) generates 0-100, but logic may expect 1-100")
		
		// With 100% error rate, we should never get non-error responses
		if zeroDrawCount > 0 {
			t.Logf("Possible bug: %d non-error responses with 100%% error rate", zeroDrawCount)
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
