package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("adds CORS headers to response", func(t *testing.T) {
		router := gin.New()
		router.Use(Cors)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD",
			"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-Mock-Host, X-Mock-Uri, X-Mock-Method",
			"Access-Control-Max-Age":       "86400",
		}

		for header, expectedValue := range expectedHeaders {
			if actualValue := w.Header().Get(header); actualValue != expectedValue {
				t.Errorf("expected header %s to be '%s', got '%s'", header, expectedValue, actualValue)
			}
		}
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		router := gin.New()
		router.Use(Cors)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
		}

		// Verify CORS headers are still set
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("CORS headers should be set on OPTIONS response")
		}
	})

	t.Run("allows request to continue for non-OPTIONS methods", func(t *testing.T) {
		router := gin.New()
		router.Use(Cors)

		handlerCalled := false
		router.POST("/test", func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusCreated)
		})

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if !handlerCalled {
			t.Error("handler should be called for non-OPTIONS requests")
		}

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
	})

	t.Run("OPTIONS request does not call handler", func(t *testing.T) {
		router := gin.New()
		router.Use(Cors)

		handlerCalled := false
		router.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if handlerCalled {
			t.Error("handler should not be called for OPTIONS requests")
		}
	})
}
