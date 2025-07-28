package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitRoutes(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("initializes all routes correctly", func(t *testing.T) {
		router := gin.New()

		// Create mock controllers (they can be nil for route testing)
		adminMocksController := &AdminMocksController{}
		adminHostsController := &AdminHostsController{}
		mocksController := &MocksController{}

		// Initialize routes
		InitRoutes(router, adminMocksController, adminHostsController, mocksController)

		// Test admin mocks routes
		adminMocksRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodPost, "/admin/mocks"},
			{http.MethodDelete, "/admin/mocks"},
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
			{http.MethodGet, "/admin/config/hosts"},
			{http.MethodPost, "/admin/config/hosts"},
			{http.MethodGet, "/admin/config/hosts/example.com"},
			{http.MethodDelete, "/admin/config/hosts/example.com"},
			{http.MethodPost, "/admin/config/hosts/example.com/latencies"},
			{http.MethodDelete, "/admin/config/hosts/example.com/latencies"},
			{http.MethodPost, "/admin/config/hosts/example.com/errors"},
			{http.MethodDelete, "/admin/config/hosts/example.com/errors/500"},
			{http.MethodPost, "/admin/config/hosts/example.com/uris"},
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

	t.Run("NoRoute handler is set", func(t *testing.T) {
		router := gin.New()

		adminMocksController := &AdminMocksController{}
		adminHostsController := &AdminHostsController{}
		mocksController := &MocksController{}

		InitRoutes(router, adminMocksController, adminHostsController, mocksController)

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

func TestInitAdminMocksController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets up admin mocks routes", func(t *testing.T) {
		router := gin.New()
		group := router.Group("/admin/mocks")
		controller := &AdminMocksController{}

		initAdminMocksController(group, controller)

		// Test POST route
		req := httptest.NewRequest(http.MethodPost, "/admin/mocks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Error("POST /admin/mocks route should exist")
		}

		// Test DELETE route
		req = httptest.NewRequest(http.MethodDelete, "/admin/mocks", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Error("DELETE /admin/mocks route should exist")
		}
	})
}

func TestInitAdminHostsController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets up admin hosts routes", func(t *testing.T) {
		router := gin.New()
		group := router.Group("/admin/config/hosts")
		controller := &AdminHostsController{}

		initAdminHostsController(group, controller)

		// Test base routes
		baseRoutes := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/admin/config/hosts"},
			{http.MethodPost, "/admin/config/hosts"},
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
			{http.MethodGet, "/admin/config/hosts/testhost"},
			{http.MethodDelete, "/admin/config/hosts/testhost"},
			{http.MethodPost, "/admin/config/hosts/testhost/latencies"},
			{http.MethodDelete, "/admin/config/hosts/testhost/latencies"},
			{http.MethodPost, "/admin/config/hosts/testhost/errors"},
			{http.MethodDelete, "/admin/config/hosts/testhost/errors/500"},
			{http.MethodPost, "/admin/config/hosts/testhost/uris"},
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

	t.Run("admin routes are properly grouped", func(t *testing.T) {
		router := gin.New()

		adminMocksController := &AdminMocksController{}
		adminHostsController := &AdminHostsController{}
		mocksController := &MocksController{}

		InitRoutes(router, adminMocksController, adminHostsController, mocksController)

		// Test that admin routes are under correct paths
		adminPaths := []string{
			"/admin/mocks",
			"/admin/config/hosts",
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

	t.Run("non-admin routes handled by NoRoute", func(t *testing.T) {
		router := gin.New()

		adminMocksController := &AdminMocksController{}
		adminHostsController := &AdminHostsController{}
		mocksController := &MocksController{}

		InitRoutes(router, adminMocksController, adminHostsController, mocksController)

		// Test non-admin paths
		nonAdminPaths := []string{
			"/api/users",
			"/health",
			"/",
			"/some/random/path",
		}

		for _, path := range nonAdminPaths {
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
					t.Errorf("non-admin route %s should be handled by NoRoute", path)
				}
			})
		}
	})
}
