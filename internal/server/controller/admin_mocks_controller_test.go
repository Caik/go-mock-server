package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

// Mock content service for testing
type mockContentService struct {
	shouldError bool
	errorMsg    string
}

func (m *mockContentService) SetContent(host, uri, method, uuid string, data *[]byte) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	return nil
}

func (m *mockContentService) GetContent(host, uri, method, uuid string) (*[]byte, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	data := []byte("mock content")
	return &data, nil
}

func (m *mockContentService) DeleteContent(host, uri, method, uuid string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	return nil
}

func (m *mockContentService) ListContents(uuid string) (*[]content.ContentData, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	contents := []content.ContentData{
		{Host: "example.com", Uri: "/api/users", Method: "GET"},
	}
	return &contents, nil
}

func (m *mockContentService) Subscribe(subscriberId string, eventTypes ...content.ContentEventType) <-chan content.ContentEvent {
	ch := make(chan content.ContentEvent)
	return ch
}

func (m *mockContentService) Unsubscribe(subscriberId string) {
	// Mock implementation
}

func TestNewAdminMocksController(t *testing.T) {
	t.Run("creates controller with service", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		if controller == nil {
			t.Fatal("NewAdminMocksController should return non-nil controller")
		}

		if controller.service != service {
			t.Error("controller should store the provided service")
		}
	})
}

func TestAdminMocksController_handleMockAddUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("adds mock successfully", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request with headers and body
		requestBody := []byte(`{"message": "test mock data"}`)
		req := httptest.NewRequest(http.MethodPost, "/admin/mocks", bytes.NewBuffer(requestBody))
		req.Header.Set("x-mock-host", "example.com")
		req.Header.Set("x-mock-uri", "/api/users")
		req.Header.Set("x-mock-method", "GET")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockAddUpdate(c)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}
	})

	t.Run("returns error for missing headers", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request without required headers
		requestBody := []byte(`{"message": "test mock data"}`)
		req := httptest.NewRequest(http.MethodPost, "/admin/mocks", bytes.NewBuffer(requestBody))
		// Missing x-mock-host header

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockAddUpdate(c)

		// Verify error response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})

	t.Run("returns error for invalid host", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request with invalid host
		requestBody := []byte(`{"message": "test mock data"}`)
		req := httptest.NewRequest(http.MethodPost, "/admin/mocks", bytes.NewBuffer(requestBody))
		req.Header.Set("x-mock-host", "invalid host.com") // Invalid host with space
		req.Header.Set("x-mock-uri", "/api/users")
		req.Header.Set("x-mock-method", "GET")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockAddUpdate(c)

		// Verify error response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})

	t.Run("returns error for empty request body", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request with empty body
		req := httptest.NewRequest(http.MethodPost, "/admin/mocks", bytes.NewBuffer([]byte{}))
		req.Header.Set("x-mock-host", "example.com")
		req.Header.Set("x-mock-uri", "/api/users")
		req.Header.Set("x-mock-method", "GET")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockAddUpdate(c)

		// Verify error response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}

		if !strings.Contains(response.Message, "request body is empty") {
			t.Errorf("expected error message about empty body, got '%s'", response.Message)
		}
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		contentService := &mockContentService{
			shouldError: true,
			errorMsg:    "service error",
		}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request
		requestBody := []byte(`{"message": "test mock data"}`)
		req := httptest.NewRequest(http.MethodPost, "/admin/mocks", bytes.NewBuffer(requestBody))
		req.Header.Set("x-mock-host", "example.com")
		req.Header.Set("x-mock-uri", "/api/users")
		req.Header.Set("x-mock-method", "GET")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockAddUpdate(c)

		// Verify error response
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Error {
			t.Errorf("expected status 'error', got '%s'", response.Status)
		}
	})
}

func TestAdminMocksController_handleMockDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("deletes mock successfully", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request with headers
		req := httptest.NewRequest(http.MethodDelete, "/admin/mocks", nil)
		req.Header.Set("x-mock-host", "example.com")
		req.Header.Set("x-mock-uri", "/api/users")
		req.Header.Set("x-mock-method", "GET")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockDelete(c)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}
	})

	t.Run("returns error for missing headers", func(t *testing.T) {
		contentService := &mockContentService{}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request without required headers
		req := httptest.NewRequest(http.MethodDelete, "/admin/mocks", nil)
		// Missing headers

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockDelete(c)

		// Verify error response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		contentService := &mockContentService{
			shouldError: true,
			errorMsg:    "delete error",
		}
		service := admin.NewMockAdminService(contentService)
		controller := NewAdminMocksController(service)

		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/mocks", nil)
		req.Header.Set("x-mock-host", "example.com")
		req.Header.Set("x-mock-uri", "/api/users")
		req.Header.Set("x-mock-method", "GET")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleMockDelete(c)

		// Verify error response
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Error {
			t.Errorf("expected status 'error', got '%s'", response.Status)
		}
	})
}

func TestAddDeleteMockRequest_validate(t *testing.T) {
	tests := []struct {
		name        string
		request     AddDeleteMockRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: AddDeleteMockRequest{
				Host:   "example.com",
				Uri:    "/api/users",
				Method: "GET",
			},
			expectError: false,
		},
		{
			name: "valid IP address",
			request: AddDeleteMockRequest{
				Host:   "192.168.1.1",
				Uri:    "/api/users",
				Method: "POST",
			},
			expectError: false,
		},
		{
			name: "empty host",
			request: AddDeleteMockRequest{
				Host:   "",
				Uri:    "/api/users",
				Method: "GET",
			},
			expectError: true,
			errorMsg:    "invalid host provided",
		},
		{
			name: "empty URI",
			request: AddDeleteMockRequest{
				Host:   "example.com",
				Uri:    "",
				Method: "GET",
			},
			expectError: true,
			errorMsg:    "invalid uri provided",
		},
		{
			name: "empty method",
			request: AddDeleteMockRequest{
				Host:   "example.com",
				Uri:    "/api/users",
				Method: "",
			},
			expectError: true,
			errorMsg:    "invalid method provided",
		},
		{
			name: "invalid host",
			request: AddDeleteMockRequest{
				Host:   "invalid host.com",
				Uri:    "/api/users",
				Method: "GET",
			},
			expectError: true,
			errorMsg:    "invalid host provided",
		},
		{
			name: "invalid URI",
			request: AddDeleteMockRequest{
				Host:   "example.com",
				Uri:    "/api /users", // Space in URI
				Method: "GET",
			},
			expectError: true,
			errorMsg:    "invalid uri provided",
		},
		{
			name: "invalid method",
			request: AddDeleteMockRequest{
				Host:   "example.com",
				Uri:    "/api/users",
				Method: "INVALID",
			},
			expectError: true,
			errorMsg:    "invalid method provided",
		},
		{
			name: "lowercase method gets converted",
			request: AddDeleteMockRequest{
				Host:   "example.com",
				Uri:    "/api/users",
				Method: "get", // Should be converted to uppercase
			},
			expectError: false, // This should pass after conversion
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}

				// For the lowercase method test, verify it was converted to uppercase
				if tt.name == "lowercase method gets converted" && tt.request.Method != "GET" {
					t.Errorf("expected method to be converted to 'GET', got '%s'", tt.request.Method)
				}
			}
		})
	}
}


