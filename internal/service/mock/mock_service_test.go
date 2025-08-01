package mock

import (
	"testing"
)

func TestMockResponse_AddHeaders(t *testing.T) {
	t.Run("adds headers when none exist", func(t *testing.T) {
		response := &MockResponse{
			StatusCode:  200,
			ContentType: "application/json",
			Headers:     nil,
		}

		headersToAdd := map[string]string{
			"X-Custom-Header": "custom-value",
			"Cache-Control":   "no-cache",
		}

		response.AddHeaders(headersToAdd)

		if response.Headers == nil {
			t.Fatal("expected headers to be set, got nil")
		}

		for key, expectedValue := range headersToAdd {
			if actualValue, exists := (*response.Headers)[key]; !exists {
				t.Errorf("expected header %s to be set", key)
			} else if actualValue != expectedValue {
				t.Errorf("expected header %s to be %s, got %s", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("merges headers with existing ones", func(t *testing.T) {
		existingHeaders := map[string]string{
			"Content-Encoding": "gzip",
			"X-Existing":       "existing-value",
		}

		response := &MockResponse{
			StatusCode:  200,
			ContentType: "application/json",
			Headers:     &existingHeaders,
		}

		headersToAdd := map[string]string{
			"X-Custom-Header": "custom-value",
			"Cache-Control":   "no-cache",
		}

		response.AddHeaders(headersToAdd)

		if response.Headers == nil {
			t.Fatal("expected headers to be set, got nil")
		}

		// Check existing headers are preserved
		if (*response.Headers)["Content-Encoding"] != "gzip" {
			t.Error("expected existing header Content-Encoding to be preserved")
		}
		if (*response.Headers)["X-Existing"] != "existing-value" {
			t.Error("expected existing header X-Existing to be preserved")
		}

		// Check new headers are added
		for key, expectedValue := range headersToAdd {
			if actualValue, exists := (*response.Headers)[key]; !exists {
				t.Errorf("expected header %s to be set", key)
			} else if actualValue != expectedValue {
				t.Errorf("expected header %s to be %s, got %s", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("overwrites existing headers with same key", func(t *testing.T) {
		existingHeaders := map[string]string{
			"X-Custom-Header": "old-value",
			"Cache-Control":   "max-age=3600",
		}

		response := &MockResponse{
			StatusCode:  200,
			ContentType: "application/json",
			Headers:     &existingHeaders,
		}

		headersToAdd := map[string]string{
			"X-Custom-Header": "new-value",
			"X-New-Header":    "new-header-value",
		}

		response.AddHeaders(headersToAdd)

		if response.Headers == nil {
			t.Fatal("expected headers to be set, got nil")
		}

		// Check that existing header was overwritten
		if (*response.Headers)["X-Custom-Header"] != "new-value" {
			t.Errorf("expected X-Custom-Header to be overwritten with new-value, got %s", (*response.Headers)["X-Custom-Header"])
		}

		// Check that other existing header is preserved
		if (*response.Headers)["Cache-Control"] != "max-age=3600" {
			t.Error("expected existing Cache-Control header to be preserved")
		}

		// Check that new header is added
		if (*response.Headers)["X-New-Header"] != "new-header-value" {
			t.Error("expected new header X-New-Header to be added")
		}
	})

	t.Run("handles empty headers map", func(t *testing.T) {
		response := &MockResponse{
			StatusCode:  200,
			ContentType: "application/json",
			Headers:     nil,
		}

		emptyHeaders := map[string]string{}
		response.AddHeaders(emptyHeaders)

		if response.Headers == nil {
			t.Error("expected headers to be initialized even with empty map")
		}

		if len(*response.Headers) != 0 {
			t.Errorf("expected empty headers map, got %d headers", len(*response.Headers))
		}
	})
}

func TestGenerateCacheKey(t *testing.T) {
	t.Run("generates correct cache key", func(t *testing.T) {
		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		key := GenerateCacheKey(request)
		expected := "example.com:GET:/api/test"

		if key != expected {
			t.Errorf("expected cache key %s, got %s", expected, key)
		}
	})

	t.Run("handles empty values", func(t *testing.T) {
		request := MockRequest{
			Host:   "",
			Method: "",
			URI:    "",
		}

		key := GenerateCacheKey(request)
		expected := "::"

		if key != expected {
			t.Errorf("expected cache key %s, got %s", expected, key)
		}
	})
}
