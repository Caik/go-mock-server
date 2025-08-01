package mock

import (
	"testing"
)

func TestCorsMockService_GetMockResponse(t *testing.T) {
	t.Run("adds CORS headers when response has no existing headers", func(t *testing.T) {
		// Create a mock service that returns a response without headers
		testData := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode:  200,
				Data:        &testData,
				ContentType: "application/json",
				Headers:     nil,
			},
		}

		corsService := newCorsMockService()
		corsService.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		response := corsService.getMockResponse(request)

		// Verify response is not nil
		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Verify headers were added
		if response.Headers == nil {
			t.Fatal("expected headers to be set, got nil")
		}

		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, Accept, Origin",
			"Access-Control-Max-Age":       "86400",
		}

		for key, expectedValue := range expectedHeaders {
			if actualValue, exists := (*response.Headers)[key]; !exists {
				t.Errorf("expected header %s to be set", key)
			} else if actualValue != expectedValue {
				t.Errorf("expected header %s to be %s, got %s", key, expectedValue, actualValue)
			}
		}

		// Verify original response data is preserved
		if response.StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", response.StatusCode)
		}
		if response.ContentType != "application/json" {
			t.Errorf("expected content type application/json, got %s", response.ContentType)
		}
	})

	t.Run("merges CORS headers with existing headers", func(t *testing.T) {
		// Create a mock service that returns a response with existing headers
		testData := []byte("test response")
		existingHeaders := map[string]string{
			"X-Custom-Header": "custom-value",
			"Cache-Control":   "no-cache",
		}
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode:  200,
				Data:        &testData,
				ContentType: "application/json",
				Headers:     &existingHeaders,
			},
		}

		corsService := newCorsMockService()
		corsService.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "POST",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		response := corsService.getMockResponse(request)

		// Verify response is not nil
		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Verify headers were merged
		if response.Headers == nil {
			t.Fatal("expected headers to be set, got nil")
		}

		// Check that existing headers are preserved
		if (*response.Headers)["X-Custom-Header"] != "custom-value" {
			t.Error("expected existing custom header to be preserved")
		}
		if (*response.Headers)["Cache-Control"] != "no-cache" {
			t.Error("expected existing cache-control header to be preserved")
		}

		// Check that CORS headers were added
		expectedCorsHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, Accept, Origin",
			"Access-Control-Max-Age":       "86400",
		}

		for key, expectedValue := range expectedCorsHeaders {
			if actualValue, exists := (*response.Headers)[key]; !exists {
				t.Errorf("expected CORS header %s to be set", key)
			} else if actualValue != expectedValue {
				t.Errorf("expected CORS header %s to be %s, got %s", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("returns nil when next service returns nil", func(t *testing.T) {
		mockNext := &mockMockService{
			response: nil,
		}

		corsService := newCorsMockService()
		corsService.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		response := corsService.getMockResponse(request)

		if response != nil {
			t.Error("expected nil response when next service returns nil")
		}
	})

	t.Run("returns nil when no next service is set", func(t *testing.T) {
		corsService := newCorsMockService()
		// Don't set next service

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		response := corsService.getMockResponse(request)

		if response != nil {
			t.Error("expected nil response when no next service is set")
		}
	})
}

func TestCorsMockService_GetCorsHeaders(t *testing.T) {
	corsService := newCorsMockService()
	headers := corsService.getCorsHeaders()

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, Accept, Origin",
		"Access-Control-Max-Age":       "86400",
	}

	if len(headers) != len(expectedHeaders) {
		t.Errorf("expected %d headers, got %d", len(expectedHeaders), len(headers))
	}

	for key, expectedValue := range expectedHeaders {
		if actualValue, exists := headers[key]; !exists {
			t.Errorf("expected header %s to be present", key)
		} else if actualValue != expectedValue {
			t.Errorf("expected header %s to be %s, got %s", key, expectedValue, actualValue)
		}
	}
}

// Note: Using mockMockService from host_resolution_mock_service_test.go
