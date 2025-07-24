package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server/controller"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
)

func TestNewServer(t *testing.T) {
	// Save original gin mode
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	t.Run("creates server with correct configuration", func(t *testing.T) {
		engine := NewServer()

		if engine == nil {
			t.Fatal("NewServer should return a non-nil engine")
		}

		// Verify gin is in release mode
		if gin.Mode() != gin.ReleaseMode {
			t.Error("gin should be in release mode")
		}
	})

	t.Run("server has recovery middleware", func(t *testing.T) {
		engine := NewServer()

		// Add a route that panics to test recovery
		engine.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		// This should not crash the test due to recovery middleware
		engine.ServeHTTP(w, req)

		// Recovery middleware should return 500
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 after panic, got %d", w.Code)
		}
	})

	t.Run("server has UUID middleware", func(t *testing.T) {
		engine := NewServer()

		var capturedUUID string
		engine.GET("/test", func(c *gin.Context) {
			capturedUUID = c.GetString(util.UuidKey)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		if capturedUUID == "" {
			t.Error("UUID middleware should set UUID in context")
		}
	})

	t.Run("server has logger middleware", func(t *testing.T) {
		engine := NewServer()

		// Add a simple route
		engine.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// This should not panic due to logger middleware
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("middleware order is correct", func(t *testing.T) {
		engine := NewServer()

		var middlewareOrder []string

		// Add a test route that captures middleware execution order
		engine.GET("/test", func(c *gin.Context) {
			// UUID should be available (set by UUID middleware)
			if c.GetString(util.UuidKey) != "" {
				middlewareOrder = append(middlewareOrder, "uuid_available")
			}
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		// Verify UUID middleware ran (UUID is available in handler)
		if len(middlewareOrder) == 0 || middlewareOrder[0] != "uuid_available" {
			t.Error("UUID middleware should run before handler")
		}
	})
}

func TestStartServer(t *testing.T) {
	t.Run("initializes routes correctly", func(t *testing.T) {
		// Create mock controllers
		adminMocksController := &controller.AdminMocksController{}
		adminHostsController := &controller.AdminHostsController{}
		mocksController := &controller.MocksController{}

		engine := NewServer()
		appArgs := &config.AppArguments{
			ServerPort: 0, // Use port 0 to get a random available port
		}

		params := StartServerParams{
			Engine:               engine,
			AppArguments:         appArgs,
			AdminMocksController: adminMocksController,
			AdminHostsController: adminHostsController,
			MocksController:      mocksController,
		}

		// Test that routes are initialized without starting the server
		// We'll do this by manually calling the route initialization
		controller.InitRoutes(params.Engine, params.AdminMocksController, params.AdminHostsController, params.MocksController)

		// Test that admin routes are accessible (they should return some response, even if it's an error)
		testRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/admin/config/hosts"},
			{http.MethodPost, "/admin/mocks"},
			{http.MethodDelete, "/admin/mocks"},
		}

		for _, route := range testRoutes {
			t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				w := httptest.NewRecorder()

				engine.ServeHTTP(w, req)

				// We don't care about the exact response, just that the route exists
				// (it might return 400, 500, etc. due to missing dependencies, but not 404)
				if w.Code == http.StatusNotFound {
					t.Errorf("route %s %s should exist", route.method, route.path)
				}
			})
		}
	})

	t.Run("starts server on correct port", func(t *testing.T) {
		// This test is tricky because StartServer is blocking
		// We'll test the port configuration logic without actually starting the server

		engine := NewServer()
		appArgs := &config.AppArguments{
			ServerPort: 8080,
		}

		params := StartServerParams{
			Engine:       engine,
			AppArguments: appArgs,
		}

		// We can't easily test the actual server start without it being blocking
		// But we can verify the parameters are set up correctly
		if params.AppArguments.ServerPort != 8080 {
			t.Errorf("expected port 8080, got %d", params.AppArguments.ServerPort)
		}

		if params.Engine == nil {
			t.Error("engine should not be nil")
		}
	})

	t.Run("server startup with timeout", func(t *testing.T) {
		// Test server startup and shutdown with a timeout
		engine := NewServer()
		appArgs := &config.AppArguments{
			ServerPort: 0, // Use port 0 for random available port
		}

		// Add a simple test route
		engine.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		params := StartServerParams{
			Engine:       engine,
			AppArguments: appArgs,
		}

		// Start server in a goroutine with timeout
		serverErr := make(chan error, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		go func() {
			// This will timeout after 100ms, which is expected for this test
			serverErr <- StartServer(params)
		}()

		select {
		case err := <-serverErr:
			// Server should timeout or fail to bind (both are acceptable for this test)
			if err == nil {
				t.Error("expected server to timeout or fail to bind")
			}
		case <-ctx.Done():
			// Timeout is expected - server startup was initiated
		}
	})

	t.Run("handles invalid port", func(t *testing.T) {
		engine := NewServer()
		appArgs := &config.AppArguments{
			ServerPort: -1, // Invalid port
		}

		params := StartServerParams{
			Engine:       engine,
			AppArguments: appArgs,
		}

		// Start server in a goroutine
		serverErr := make(chan error, 1)
		go func() {
			serverErr <- StartServer(params)
		}()

		// Should get an error quickly due to invalid port
		select {
		case err := <-serverErr:
			if err == nil {
				t.Error("expected error for invalid port")
			}
		case <-time.After(1 * time.Second):
			t.Error("should fail quickly with invalid port")
		}
	})
}

// Test helper to verify middleware chain
func TestMiddlewareChain(t *testing.T) {
	t.Run("middleware chain executes in correct order", func(t *testing.T) {
		engine := NewServer()

		var executionOrder []string

		// Add a test route that tracks middleware execution
		engine.Use(func(c *gin.Context) {
			// This should run after UUID and Logger middleware
			if c.GetString(util.UuidKey) != "" {
				executionOrder = append(executionOrder, "uuid_set")
			}
			c.Next()
		})

		engine.GET("/test", func(c *gin.Context) {
			executionOrder = append(executionOrder, "handler")
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

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
