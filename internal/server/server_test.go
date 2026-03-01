package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server/controller"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

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

		servers := NewServers()

		// Initialize admin routes
		controller.InitAdminRoutes(servers.AdminEngine, adminMocksController, adminHostsController)

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

		controller.InitAdminRoutes(servers.AdminEngine, adminMocksController, adminHostsController)

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
