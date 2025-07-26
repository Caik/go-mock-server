package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

func TestNewAdminHostsController(t *testing.T) {
	t.Run("creates controller with dependencies", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)

		controller := NewAdminHostsController(hostsConfig, service)

		if controller == nil {
			t.Fatal("NewAdminHostsController should return non-nil controller")
		}

		if controller.hostsConfig != hostsConfig {
			t.Error("controller should store the provided hostsConfig")
		}

		if controller.service != service {
			t.Error("controller should store the provided service")
		}
	})
}

func TestAdminHostsController_handleHostsConfigList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns hosts config successfully", func(t *testing.T) {
		// Create test hosts config
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(100),
						Max: intPtr(200),
					},
				},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostsConfigList(c)

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

		if response.Message != "hosts config retrieved with success" {
			t.Errorf("expected success message, got '%s'", response.Message)
		}

		if response.Data == nil {
			t.Error("expected data to be non-nil")
		}
	})
}

func TestAdminHostsController_handleHostConfigAddUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("adds new host config successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request body
		requestBody := AddDeleteGetHostRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(100),
				Max: intPtr(200),
			},
		}

		jsonBody, _ := json.Marshal(requestBody)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admin/config/hosts", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostConfigAddUpdate(c)

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

		// Verify host was added to config
		hostConfig := hostsConfig.GetHostConfig("example.com")
		if hostConfig == nil {
			t.Error("expected host to be added to config")
		}
	})

	t.Run("returns error for invalid request", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create invalid request body (missing required host field)
		requestBody := AddDeleteGetHostRequest{
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(100),
				Max: intPtr(200),
			},
		}

		jsonBody, _ := json.Marshal(requestBody)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admin/config/hosts", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostConfigAddUpdate(c)

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

	t.Run("returns error for invalid host config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid latency config (min > max)
		requestBody := AddDeleteGetHostRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(200),
				Max: intPtr(100), // Invalid: min > max
			},
		}

		jsonBody, _ := json.Marshal(requestBody)

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admin/config/hosts", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostConfigAddUpdate(c)

		// Verify error response (validation errors return 500, not 400)
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

func TestAdminHostsController_handleHostConfigRetrieve(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("retrieves existing host config", func(t *testing.T) {
		// Create hosts config with existing host
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(100),
						Max: intPtr(200),
					},
				},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create test context with host parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: "example.com"}}
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostConfigRetrieve(c)

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

		if response.Data == nil {
			t.Error("expected data to be non-nil")
		}
	})

	t.Run("returns not found for non-existent host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create test context with non-existent host parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: "nonexistent.com"}}
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostConfigRetrieve(c)

		// Verify not found response
		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}

		var response rest.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})
}

func TestAdminHostsController_handleHostConfigDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("deletes existing host config", func(t *testing.T) {
		// Create hosts config with existing host
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(100),
						Max: intPtr(200),
					},
				},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create test context with host parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: "example.com"}}
		c.Set(util.UuidKey, "test-uuid")

		// Call the handler
		controller.handleHostConfigDelete(c)

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

		// Verify host was deleted from config
		hostConfig := hostsConfig.GetHostConfig("example.com")
		if hostConfig != nil {
			t.Error("expected host to be deleted from config")
		}
	})
}

func TestAddDeleteGetHostRequest_validate(t *testing.T) {
	tests := []struct {
		name        string
		request     AddDeleteGetHostRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid host",
			request: AddDeleteGetHostRequest{
				Host: "example.com",
			},
			expectError: false,
		},
		{
			name: "valid IP address",
			request: AddDeleteGetHostRequest{
				Host: "192.168.1.1",
			},
			expectError: false,
		},
		{
			name: "invalid host with spaces",
			request: AddDeleteGetHostRequest{
				Host: "invalid host.com",
			},
			expectError: true,
			errorMsg:    "invalid host provided",
		},
		{
			name: "invalid host with special characters",
			request: AddDeleteGetHostRequest{
				Host: "invalid@host.com",
			},
			expectError: true,
			errorMsg:    "invalid host provided",
		},
		{
			name: "empty host",
			request: AddDeleteGetHostRequest{
				Host: "",
			},
			expectError: true,
			errorMsg:    "invalid host provided",
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate(false, false, false)

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
			}
		})
	}
}

func TestAdminHostsController_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handleHostConfigAddUpdate handles invalid JSON", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid JSON
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(util.UuidKey, "test-uuid")

		controller.handleHostConfigAddUpdate(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})

	t.Run("handleHostConfigAddUpdate handles validation errors", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid host (empty)
		requestBody := `{"host": "", "latency": {"min": 100, "max": 200}}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(util.UuidKey, "test-uuid")

		controller.handleHostConfigAddUpdate(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})

	t.Run("handleHostConfigRetrieve handles invalid host parameter", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid host parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: ""}} // Empty host
		c.Set(util.UuidKey, "test-uuid")

		controller.handleHostConfigRetrieve(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("handleHostConfigRetrieve handles non-existent host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request for non-existent host
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: "nonexistent.com"}}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleHostConfigRetrieve(c)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Fail {
			t.Errorf("expected status 'fail', got '%s'", response.Status)
		}
	})

	t.Run("handleHostConfigDelete handles invalid host parameter", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid host parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: ""}} // Empty host
		c.Set(util.UuidKey, "test-uuid")

		controller.handleHostConfigDelete(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestAdminHostsController_LatencyHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handleLatencyAddUpdate adds latency config successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with latency config
		requestBody := `{"host": "example.com", "latency": {"min": 100, "max": 500, "p95": 400, "p99": 450}}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts/example.com/latencies", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "host", Value: "example.com"}}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleLatencyAddUpdate(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}

		// Verify latency config was added
		hostConfig := hostsConfig.GetHostConfig("example.com")
		if hostConfig == nil || hostConfig.LatencyConfig == nil {
			t.Error("latency config should be added to host")
		}
	})

	t.Run("handleLatencyDelete removes latency config successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(100),
						Max: intPtr(500),
					},
				},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create delete request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: "example.com"}}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleLatencyDelete(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}
	})

	t.Run("handleLatencyAddUpdate handles invalid host parameter", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid host parameter
		requestBody := `{"host": "", "latency": {"min": 100, "max": 500}}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts//latencies", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "host", Value: ""}} // Empty host
		c.Set(util.UuidKey, "test-uuid")

		controller.handleLatencyAddUpdate(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("handleLatencyDelete handles invalid host parameter", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with invalid host parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "host", Value: ""}} // Empty host
		c.Set(util.UuidKey, "test-uuid")

		controller.handleLatencyDelete(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestAdminHostsController_ErrorHandlingEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handleErrorsAddUpdate adds error config successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with error config
		requestBody := `{"host": "example.com", "errors": {"500": {"percentage": 10}, "404": {"percentage": 5}}}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts/example.com/errors", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "host", Value: "example.com"}}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleErrorsAddUpdate(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}

		// Verify error config was added
		hostConfig := hostsConfig.GetHostConfig("example.com")
		if hostConfig == nil || hostConfig.ErrorsConfig == nil {
			t.Error("error config should be added to host")
		}
	})

	t.Run("handleErrorDelete removes specific error successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					ErrorsConfig: map[string]config.ErrorConfig{
						"500": {Percentage: intPtr(10)},
						"404": {Percentage: intPtr(5)},
					},
				},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create delete request for specific error
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "host", Value: "example.com"},
			{Key: "error", Value: "500"},
		}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleErrorDelete(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}
	})

	t.Run("handleUrisAddUpdate adds URI config successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {},
			},
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Create request with URI config
		requestBody := `{"host": "example.com", "uris": {"/api/users": {"methods": ["GET", "POST"]}, "/api/health": {"methods": ["GET"]}}}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts/example.com/uris", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "host", Value: "example.com"}}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleUrisAddUpdate(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response rest.Response
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Status != rest.Success {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}

		// Verify URI config was added
		hostConfig := hostsConfig.GetHostConfig("example.com")
		if hostConfig == nil || hostConfig.UrisConfig == nil {
			t.Error("URI config should be added to host")
		}
	})

	t.Run("error endpoints handle invalid host parameters", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		service := admin.NewHostsConfigAdminService(hostsConfig)
		controller := NewAdminHostsController(hostsConfig, service)

		// Test handleErrorsAddUpdate with empty host
		requestBody := `{"host": "", "errors": {"500": {"percentage": 10}}}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/config/hosts//errors", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "host", Value: ""}}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleErrorsAddUpdate(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		// Test handleErrorDelete with empty host
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "host", Value: ""},
			{Key: "error", Value: "500"},
		}
		c.Set(util.UuidKey, "test-uuid")

		controller.handleErrorDelete(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
