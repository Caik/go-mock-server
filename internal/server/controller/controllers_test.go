package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitAdminRoutes(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("initializes all admin routes correctly", func(t *testing.T) {
		router := gin.New()

		// Create mock controllers (they can be nil for route testing)
		adminMocksController := &AdminMocksController{}
		adminHostsController := &AdminHostsController{}
		trafficController := NewTrafficController(nil)

		// Initialize admin routes
		InitAdminRoutes(router, adminMocksController, adminHostsController, trafficController)

		// Test health endpoint
		t.Run("GET /health", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("health endpoint should return 200, got %d", w.Code)
			}
		})

		// Test admin mocks routes
		adminMocksRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodPost, "/api/v1/mocks"},
			{http.MethodDelete, "/api/v1/mocks"},
		}

		for _, route := range adminMocksRoutes {
			t.Run(route.method+" "+route.path, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				w := httptest.NewRecorder()

				// Expect panic due to nil dependencies, but route should exist
				defer func() {
					if r := recover(); r != nil {
						// Panic is expected due to nil service dependencies
						t.Logf("Expected panic due to nil dependencies: %v", r)
					}
				}()

				router.ServeHTTP(w, req)

				// If we get here without panic, route should exist (not return 404)
				if w.Code == http.StatusNotFound {
					t.Errorf("route %s %s should exist", route.method, route.path)
				}
			})
		}

		// Test admin hosts routes
		adminHostsRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/api/v1/config/hosts"},
			{http.MethodPost, "/api/v1/config/hosts"},
			{http.MethodGet, "/api/v1/config/hosts/example.com"},
			{http.MethodDelete, "/api/v1/config/hosts/example.com"},
			{http.MethodPost, "/api/v1/config/hosts/example.com/latencies"},
			{http.MethodDelete, "/api/v1/config/hosts/example.com/latencies"},
			{http.MethodPost, "/api/v1/config/hosts/example.com/statuses"},
			{http.MethodDelete, "/api/v1/config/hosts/example.com/statuses/500"},
			{http.MethodPost, "/api/v1/config/hosts/example.com/uris"},
		}

		for _, route := range adminHostsRoutes {
			t.Run(route.method+" "+route.path, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				w := httptest.NewRecorder()

				// Expect panic due to nil dependencies, but route should exist
				defer func() {
					if r := recover(); r != nil {
						// Panic is expected due to nil service dependencies
						// The important thing is that the route was matched (not 404)
						t.Logf("Expected panic due to nil dependencies: %v", r)
					}
				}()

				router.ServeHTTP(w, req)

				// If we get here without panic, route should exist (not return 404)
				if w.Code == http.StatusNotFound {
					t.Errorf("route %s %s should exist", route.method, route.path)
				}
			})
		}
	})
}

func TestInitMockRoutes(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("NoRoute handler is set", func(t *testing.T) {
		router := gin.New()

		mocksController := &MocksController{}

		InitMockRoutes(router, mocksController)

		// Test a route that doesn't exist - should be handled by NoRoute (MocksController)
		req := httptest.NewRequest(http.MethodGet, "/some/random/path", nil)
		w := httptest.NewRecorder()

		// Expect panic due to nil dependencies in MocksController
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected due to nil factory in MocksController
				t.Logf("Expected panic due to nil dependencies: %v", r)
			}
		}()

		router.ServeHTTP(w, req)

		// If we get here without panic, NoRoute should handle it (not return 404)
		if w.Code == http.StatusNotFound {
			t.Error("NoRoute handler should be set and handle unmatched routes")
		}
	})
}

func TestInitAdminMocksControllerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets up admin mocks routes", func(t *testing.T) {
		router := gin.New()
		group := router.Group("/api/v1/mocks")
		controller := &AdminMocksController{}

		initAdminMocksController(group, controller)

		// Test POST route
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mocks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Error("POST /api/v1/mocks route should exist")
		}

		// Test DELETE route
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/mocks", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Error("DELETE /api/v1/mocks route should exist")
		}
	})
}

func TestInitAdminHostsControllerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets up admin hosts routes", func(t *testing.T) {
		router := gin.New()
		group := router.Group("/api/v1/config/hosts")
		controller := &AdminHostsController{}

		initAdminHostsController(group, controller)

		// Test base routes
		baseRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/api/v1/config/hosts"},
			{http.MethodPost, "/api/v1/config/hosts"},
		}

		for _, route := range baseRoutes {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("%s %s route should exist", route.method, route.path)
			}
		}

		// Test host-specific routes
		hostRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/api/v1/config/hosts/testhost"},
			{http.MethodDelete, "/api/v1/config/hosts/testhost"},
			{http.MethodPost, "/api/v1/config/hosts/testhost/latencies"},
			{http.MethodDelete, "/api/v1/config/hosts/testhost/latencies"},
			{http.MethodPost, "/api/v1/config/hosts/testhost/statuses"},
			{http.MethodDelete, "/api/v1/config/hosts/testhost/statuses/500"},
			{http.MethodPost, "/api/v1/config/hosts/testhost/uris"},
		}

		for _, route := range hostRoutes {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("%s %s route should exist", route.method, route.path)
			}
		}
	})
}

func TestRouteGrouping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("admin routes are properly grouped under /api/v1", func(t *testing.T) {
		router := gin.New()

		adminMocksController := &AdminMocksController{}
		adminHostsController := &AdminHostsController{}
		trafficController := NewTrafficController(nil)

		InitAdminRoutes(router, adminMocksController, adminHostsController, trafficController)

		// Test that admin routes are under correct paths
		adminPaths := []string{
			"/api/v1/mocks",
			"/api/v1/config/hosts",
		}

		for _, path := range adminPaths {
			// Test with POST method
			req := httptest.NewRequest(http.MethodPost, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not be 404 (route exists)
			if w.Code == http.StatusNotFound {
				t.Errorf("admin route %s should exist", path)
			}
		}
	})

	t.Run("mock routes handled by NoRoute", func(t *testing.T) {
		router := gin.New()

		mocksController := &MocksController{}

		InitMockRoutes(router, mocksController)

		// Test paths that should be handled by MocksController
		mockPaths := []string{
			"/api/users",
			"/",
			"/some/random/path",
		}

		for _, path := range mockPaths {
			t.Run("path "+path, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, path, nil)
				w := httptest.NewRecorder()

				// Expect panic due to nil dependencies in MocksController
				defer func() {
					if r := recover(); r != nil {
						// Panic is expected due to nil factory in MocksController
						t.Logf("Expected panic due to nil dependencies: %v", r)
					}
				}()

				router.ServeHTTP(w, req)

				// If we get here without panic, NoRoute should handle it (not return 404)
				if w.Code == http.StatusNotFound {
					t.Errorf("mock route %s should be handled by NoRoute", path)
				}
			})
		}
	})
}
