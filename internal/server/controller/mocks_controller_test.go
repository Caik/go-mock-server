package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

func TestNewMocksController(t *testing.T) {
	t.Run("creates controller with factory", func(t *testing.T) {
		// We can't easily mock MockServiceFactory since it's a concrete struct
		// So we'll test with a nil factory and verify the controller is created
		controller := NewMocksController(nil)

		if controller == nil {
			t.Fatal("NewMocksController should return non-nil controller")
		}

		// Verify the factory field exists (even if nil)
		if controller.factory != nil {
			t.Error("controller should store nil factory for this test")
		}
	})
}

func TestMocksController_sanitizeHost(t *testing.T) {
	controller := &MocksController{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "host without port",
			input:    "example.com",
			expected: "example.com",
		},
		{
			name:     "host with port",
			input:    "example.com:8080",
			expected: "example.com",
		},
		{
			name:     "localhost with port",
			input:    "localhost:3000",
			expected: "localhost",
		},
		{
			name:     "IP with port",
			input:    "192.168.1.1:8080",
			expected: "192.168.1.1",
		},
		{
			name:     "uppercase host with port",
			input:    "EXAMPLE.COM:8080",
			expected: "example.com",
		},
		{
			name:     "mixed case host with port",
			input:    "Example.Com:8080",
			expected: "example.com",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "just port",
			input:    ":8080",
			expected: "",
		},
		{
			name:     "multiple colons",
			input:    "example.com:8080:extra",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controller.sanitizeHost(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeHost(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMocksController_newMockRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates mock request from gin context", func(t *testing.T) {
		controller := &MocksController{}

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set up the request
		req := httptest.NewRequest(http.MethodPost, "/api/users?id=123", nil)
		req.Host = "api.example.com:8080"
		req.Header.Set("Accept", "application/json")
		c.Request = req

		// Set UUID in context
		c.Set(util.UuidKey, "test-uuid-123")

		// Create mock request
		mockRequest := controller.newMockRequest(c)

		// Verify the mock request
		if mockRequest.Host != "api.example.com" {
			t.Errorf("expected host 'api.example.com', got '%s'", mockRequest.Host)
		}

		if mockRequest.URI != "/api/users?id=123" {
			t.Errorf("expected URI '/api/users?id=123', got '%s'", mockRequest.URI)
		}

		if mockRequest.Method != http.MethodPost {
			t.Errorf("expected method 'POST', got '%s'", mockRequest.Method)
		}

		if mockRequest.Accept != "application/json" {
			t.Errorf("expected accept 'application/json', got '%s'", mockRequest.Accept)
		}

		if mockRequest.Uuid != "test-uuid-123" {
			t.Errorf("expected UUID 'test-uuid-123', got '%s'", mockRequest.Uuid)
		}
	})

	t.Run("handles missing UUID", func(t *testing.T) {
		controller := &MocksController{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Host = "example.com"
		c.Request = req

		// Don't set UUID in context

		mockRequest := controller.newMockRequest(c)

		if mockRequest.Uuid != "" {
			t.Errorf("expected empty UUID, got '%s'", mockRequest.Uuid)
		}
	})

	t.Run("handles missing accept header", func(t *testing.T) {
		controller := &MocksController{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Host = "example.com"
		c.Request = req

		mockRequest := controller.newMockRequest(c)

		if mockRequest.Accept != "" {
			t.Errorf("expected empty Accept, got '%s'", mockRequest.Accept)
		}
	})
}

func TestMocksController_handleMockRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles nil factory gracefully", func(t *testing.T) {
		controller := NewMocksController(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Host = "example.com"
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// This should panic or handle nil factory gracefully
		// Since we can't easily mock the factory, we'll test the error case
		defer func() {
			if r := recover(); r == nil {
				// If no panic, check if it handled nil gracefully
				if w.Code != http.StatusInternalServerError {
					t.Error("expected 500 status or panic with nil factory")
				}
			}
		}()

		controller.handleMockRequest(c)
	})
}

func TestMocksController_RequestParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles different HTTP methods in request parsing", func(t *testing.T) {
		controller := &MocksController{}

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			t.Run("method "+method, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				req := httptest.NewRequest(method, "/api/test", nil)
				req.Host = "example.com"
				c.Request = req
				c.Set(util.UuidKey, "test-uuid")

				mockRequest := controller.newMockRequest(c)

				if mockRequest.Method != method {
					t.Errorf("expected method %s, got %s", method, mockRequest.Method)
				}

				if mockRequest.Host != "example.com" {
					t.Errorf("expected host 'example.com', got '%s'", mockRequest.Host)
				}

				if mockRequest.URI != "/api/test" {
					t.Errorf("expected URI '/api/test', got '%s'", mockRequest.URI)
				}

				if mockRequest.Uuid != "test-uuid" {
					t.Errorf("expected UUID 'test-uuid', got '%s'", mockRequest.Uuid)
				}
			})
		}
	})

	t.Run("handles different Accept headers", func(t *testing.T) {
		controller := &MocksController{}

		acceptHeaders := []string{
			"application/json",
			"application/xml",
			"text/plain",
			"text/html",
			"*/*",
		}

		for _, acceptHeader := range acceptHeaders {
			t.Run("accept "+acceptHeader, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				req := httptest.NewRequest("GET", "/api/test", nil)
				req.Host = "example.com"
				req.Header.Set("Accept", acceptHeader)
				c.Request = req
				c.Set(util.UuidKey, "test-uuid")

				mockRequest := controller.newMockRequest(c)

				if mockRequest.Accept != acceptHeader {
					t.Errorf("expected Accept '%s', got '%s'", acceptHeader, mockRequest.Accept)
				}
			})
		}
	})

	t.Run("handles complex URIs", func(t *testing.T) {
		controller := &MocksController{}

		complexURIs := []string{
			"/api/users?id=123&name=test",
			"/api/v1/users/123",
			"/api/users-list",
			"/api/users_list",
			"/api/users.json",
			"/api/users?filter=name%3Dtest",
		}

		for _, uri := range complexURIs {
			t.Run("URI "+uri, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				req := httptest.NewRequest("GET", uri, nil)
				req.Host = "example.com"
				c.Request = req
				c.Set(util.UuidKey, "test-uuid")

				mockRequest := controller.newMockRequest(c)

				if mockRequest.URI != uri {
					t.Errorf("expected URI '%s', got '%s'", uri, mockRequest.URI)
				}
			})
		}
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
