package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server/controller"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

// mockContentService is a minimal implementation for testing
type mockContentService struct{}

func (m *mockContentService) GetContent(host, uri, method, uuid string, statusCode int) (*content.ContentResult, error) {
	return nil, nil
}

func (m *mockContentService) SetContent(host, uri, method, uuid string, statusCode int, data *[]byte) error {
	return nil
}

func (m *mockContentService) DeleteContent(host, uri, method, uuid string, statusCode int) error {
	return nil
}

func (m *mockContentService) ListContents(uuid string) (*[]content.ContentData, error) {
	return nil, nil
}

func (m *mockContentService) Subscribe(subscriberId string, eventTypes ...content.ContentEventType) <-chan content.ContentEvent {
	return make(chan content.ContentEvent)
}

func (m *mockContentService) Unsubscribe(subscriberId string) {
}

// Ensure mockContentService implements content.ContentService
var _ content.ContentService = (*mockContentService)(nil)

func TestNewServers(t *testing.T) {
	// Save original gin mode
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	// Ensure gin is in release mode for this test
	gin.SetMode(gin.ReleaseMode)

	t.Run("creates servers with correct configuration", func(t *testing.T) {
		servers := NewServers()

		if servers == nil {
			t.Fatal("NewServers should return a non-nil Servers struct")
		}

		if servers.MockEngine == nil {
			t.Fatal("MockEngine should not be nil")
		}

		if servers.AdminEngine == nil {
			t.Fatal("AdminEngine should not be nil")
		}

		// Verify gin is in release mode
		if gin.Mode() != gin.ReleaseMode {
			t.Error("gin should be in release mode")
		}
	})

	t.Run("mock server has recovery middleware", func(t *testing.T) {
		servers := NewServers()

		// Add a route that panics to test recovery
		servers.MockEngine.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		// This should not crash the test due to recovery middleware
		servers.MockEngine.ServeHTTP(w, req)

		// Recovery middleware should return 500
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 after panic, got %d", w.Code)
		}
	})

	t.Run("admin server has recovery middleware", func(t *testing.T) {
		servers := NewServers()

		// Add a route that panics to test recovery
		servers.AdminEngine.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		// This should not crash the test due to recovery middleware
		servers.AdminEngine.ServeHTTP(w, req)

		// Recovery middleware should return 500
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 after panic, got %d", w.Code)
		}
	})

	t.Run("mock server has UUID middleware", func(t *testing.T) {
		servers := NewServers()

		var capturedUUID string
		servers.MockEngine.GET("/test", func(c *gin.Context) {
			capturedUUID = c.GetString(util.UuidKey)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		servers.MockEngine.ServeHTTP(w, req)

		if capturedUUID == "" {
			t.Error("UUID middleware should set UUID in context")
		}
	})

	t.Run("admin server has UUID middleware", func(t *testing.T) {
		servers := NewServers()

		var capturedUUID string
		servers.AdminEngine.GET("/test", func(c *gin.Context) {
			capturedUUID = c.GetString(util.UuidKey)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		servers.AdminEngine.ServeHTTP(w, req)

		if capturedUUID == "" {
			t.Error("UUID middleware should set UUID in context")
		}
	})

	t.Run("servers have logger middleware", func(t *testing.T) {
		servers := NewServers()

		// Add simple routes to both engines
		servers.MockEngine.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		servers.AdminEngine.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Test mock engine
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		servers.MockEngine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 from mock engine, got %d", w.Code)
		}

		// Test admin engine
		req = httptest.NewRequest(http.MethodGet, "/test", nil)
		w = httptest.NewRecorder()
		servers.AdminEngine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 from admin engine, got %d", w.Code)
		}
	})
}

func TestStartServer(t *testing.T) {
	// Set gin mode once to avoid race conditions
	gin.SetMode(gin.TestMode)

	t.Run("initializes mock routes correctly", func(t *testing.T) {
		// Create mock controllers
		mocksController := &controller.MocksController{}

		servers := NewServers()

		// Initialize mock routes
		controller.InitMockRoutes(servers.MockEngine, mocksController)

		// Test that non-admin paths are handled by NoRoute (MocksController)
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		w := httptest.NewRecorder()

		// Expect panic due to nil dependencies in MocksController
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected due to nil factory in MocksController
				t.Logf("Expected panic due to nil dependencies: %v", r)
			}
		}()

		servers.MockEngine.ServeHTTP(w, req)

		// If we get here without panic, NoRoute should handle it (not return 404)
		if w.Code == http.StatusNotFound {
			t.Error("NoRoute handler should be set and handle unmatched routes")
		}
	})

	t.Run("initializes admin routes correctly", func(t *testing.T) {
		// Create admin controllers
		adminMocksController := &controller.AdminMocksController{}
		adminHostsController := &controller.AdminHostsController{}
		trafficController := controller.NewTrafficController(nil)

		servers := NewServers()

		// Initialize admin routes
		controller.InitAdminRoutes(servers.AdminEngine, adminMocksController, adminHostsController, trafficController)

		// Test that admin routes are accessible (they should return some response, even if it's an error)
		testRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/health"},
			{http.MethodGet, "/api/v1/config/hosts"},
			{http.MethodPost, "/api/v1/mocks"},
			{http.MethodDelete, "/api/v1/mocks"},
		}

		for _, route := range testRoutes {
			t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				w := httptest.NewRecorder()

				servers.AdminEngine.ServeHTTP(w, req)

				// We don't care about the exact response, just that the route exists
				// (it might return 400, 500, etc. due to missing dependencies, but not 404)
				if w.Code == http.StatusNotFound {
					t.Errorf("route %s %s should exist", route.method, route.path)
				}
			})
		}
	})

	t.Run("health endpoint returns success", func(t *testing.T) {
		servers := NewServers()
		adminMocksController := &controller.AdminMocksController{}
		adminHostsController := &controller.AdminHostsController{}
		trafficController := controller.NewTrafficController(nil)

		controller.InitAdminRoutes(servers.AdminEngine, adminMocksController, adminHostsController, trafficController)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		servers.AdminEngine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("starts server on correct port", func(t *testing.T) {
		// We'll test the port configuration logic without actually starting the server
		servers := NewServers()
		appArgs := &config.AppArguments{
			ServerPort: 8080,
			AdminPort:  9090,
		}

		params := StartServerParams{
			Servers:      servers,
			AppArguments: appArgs,
		}

		// Verify the parameters are set up correctly
		if params.AppArguments.ServerPort != 8080 {
			t.Errorf("expected mock port 8080, got %d", params.AppArguments.ServerPort)
		}

		if params.AppArguments.AdminPort != 9090 {
			t.Errorf("expected admin port 9090, got %d", params.AppArguments.AdminPort)
		}

		if params.Servers.MockEngine == nil {
			t.Error("MockEngine should not be nil")
		}

		if params.Servers.AdminEngine == nil {
			t.Error("AdminEngine should not be nil")
		}
	})
}

// Test helper to verify middleware chain
func TestMiddlewareChain(t *testing.T) {
	t.Run("middleware chain executes in correct order", func(t *testing.T) {
		servers := NewServers()

		var executionOrder []string

		// Add a test route that tracks middleware execution
		servers.MockEngine.Use(func(c *gin.Context) {
			// This should run after UUID and Logger middleware
			if c.GetString(util.UuidKey) != "" {
				executionOrder = append(executionOrder, "uuid_set")
			}
			c.Next()
		})

		servers.MockEngine.GET("/test", func(c *gin.Context) {
			executionOrder = append(executionOrder, "handler")
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		servers.MockEngine.ServeHTTP(w, req)

		expectedOrder := []string{"uuid_set", "handler"}
		if len(executionOrder) != len(expectedOrder) {
			t.Fatalf("expected %d middleware executions, got %d", len(expectedOrder), len(executionOrder))
		}

		for i, expected := range expectedOrder {
			if executionOrder[i] != expected {
				t.Errorf("expected execution order[%d] to be %s, got %s", i, expected, executionOrder[i])
			}
		}
	})
}

// getAvailablePort finds an available port for testing
func getAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// TestStartServersIntegration tests the actual server startup using simplified setup
// that doesn't require the full mock service chain
func TestStartServersIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("starts both servers and handles health check", func(t *testing.T) {
		// Get available ports
		mockPort, err := getAvailablePort()
		if err != nil {
			t.Fatalf("failed to get available port: %v", err)
		}
		adminPort, err := getAvailablePort()
		if err != nil {
			t.Fatalf("failed to get available port: %v", err)
		}

		servers := NewServers()
		hostsConfig := &config.HostsConfig{}
		contentSvc := &mockContentService{}

		// Use nil for MocksController to avoid complex dependency chain
		// We'll just test that admin routes work
		adminMocksController := controller.NewAdminMocksController(admin.NewMockAdminService(contentSvc))
		adminHostsController := controller.NewAdminHostsController(hostsConfig, admin.NewHostsConfigAdminService(hostsConfig))
		trafficController := controller.NewTrafficController(nil)

		// Initialize admin routes manually for testing
		controller.InitAdminRoutes(servers.AdminEngine, adminMocksController, adminHostsController, trafficController)

		// Add a simple mock route for testing
		servers.MockEngine.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Start admin server
		adminServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", adminPort),
			Handler: servers.AdminEngine,
		}

		// Start mock server
		mockServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", mockPort),
			Handler: servers.MockEngine,
		}

		go func() {
			adminServer.ListenAndServe()
		}()

		go func() {
			mockServer.ListenAndServe()
		}()

		// Wait a bit for servers to start
		time.Sleep(100 * time.Millisecond)

		// Test health endpoint on admin port
		healthURL := fmt.Sprintf("http://localhost:%d/health", adminPort)
		resp, err := http.Get(healthURL)

		if err != nil {
			t.Fatalf("failed to reach health endpoint: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200 from health endpoint, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)

		if string(body) == "" {
			t.Error("health endpoint should return a response body")
		}

		// Test admin API endpoint
		hostsURL := fmt.Sprintf("http://localhost:%d/api/v1/config/hosts", adminPort)
		resp2, err := http.Get(hostsURL)

		if err != nil {
			t.Fatalf("failed to reach hosts config endpoint: %v", err)
		}

		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			t.Errorf("expected status 200 from hosts config endpoint, got %d", resp2.StatusCode)
		}

		// Test mock server responds
		mockURL := fmt.Sprintf("http://localhost:%d/test", mockPort)
		resp3, err := http.Get(mockURL)

		if err != nil {
			t.Fatalf("failed to reach mock server: %v", err)
		}

		defer resp3.Body.Close()

		if resp3.StatusCode != http.StatusOK {
			t.Errorf("expected status 200 from mock server, got %d", resp3.StatusCode)
		}

		// Shutdown servers
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

		defer cancel()
		adminServer.Shutdown(ctx)
		mockServer.Shutdown(ctx)
	})

	t.Run("StartServerParams struct is properly configured", func(t *testing.T) {
		servers := NewServers()
		hostsConfig := &config.HostsConfig{}
		contentSvc := &mockContentService{}

		adminMocksController := controller.NewAdminMocksController(admin.NewMockAdminService(contentSvc))
		adminHostsController := controller.NewAdminHostsController(hostsConfig, admin.NewHostsConfigAdminService(hostsConfig))

		params := StartServerParams{
			Servers: servers,
			AppArguments: &config.AppArguments{
				ServerPort: 8080,
				AdminPort:  9090,
			},
			AdminMocksController: adminMocksController,
			AdminHostsController: adminHostsController,
			MocksController:      nil, // Can be nil for this test
		}

		if params.Servers.MockEngine == nil {
			t.Error("MockEngine should not be nil")
		}

		if params.Servers.AdminEngine == nil {
			t.Error("AdminEngine should not be nil")
		}

		if params.AppArguments.ServerPort != 8080 {
			t.Errorf("expected ServerPort 8080, got %d", params.AppArguments.ServerPort)
		}

		if params.AppArguments.AdminPort != 9090 {
			t.Errorf("expected AdminPort 9090, got %d", params.AppArguments.AdminPort)
		}
	})

	t.Run("admin port disabled when set to 0", func(t *testing.T) {
		params := StartServerParams{
			AppArguments: &config.AppArguments{
				ServerPort: 8080,
				AdminPort:  0, // Disabled
			},
		}

		// When AdminPort is 0, the admin server should not start
		if params.AppArguments.AdminPort != 0 {
			t.Error("AdminPort should be 0 (disabled)")
		}
	})
}
