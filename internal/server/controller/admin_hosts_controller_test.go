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


