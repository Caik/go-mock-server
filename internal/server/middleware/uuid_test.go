package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestUuid(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("sets UUID in context", func(t *testing.T) {
		// Create a test router
		router := gin.New()

		var capturedUUID string

		// Add the UUID middleware and a test handler
		router.Use(Uuid)
		router.GET("/test", func(c *gin.Context) {
			capturedUUID = c.GetString(util.UuidKey)
			c.Status(http.StatusOK)
		})

		// Create a test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Verify the UUID was set
		if capturedUUID == "" {
			t.Error("UUID should be set in context")
		}

		// Verify it's a valid UUID
		if _, err := uuid.Parse(capturedUUID); err != nil {
			t.Errorf("captured UUID should be valid, got error: %v", err)
		}
	})

	t.Run("generates different UUIDs for different requests", func(t *testing.T) {
		router := gin.New()

		var uuids []string

		router.Use(Uuid)
		router.GET("/test", func(c *gin.Context) {
			uuids = append(uuids, c.GetString(util.UuidKey))
			c.Status(http.StatusOK)
		})

		// Make multiple requests
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		// Verify we have 3 different UUIDs
		if len(uuids) != 3 {
			t.Fatalf("expected 3 UUIDs, got %d", len(uuids))
		}

		// Verify all UUIDs are different
		for i := 0; i < len(uuids); i++ {
			for j := i + 1; j < len(uuids); j++ {
				if uuids[i] == uuids[j] {
					t.Errorf("UUIDs should be different, but got duplicate: %s", uuids[i])
				}
			}
		}
	})

	t.Run("calls next middleware", func(t *testing.T) {
		router := gin.New()

		nextCalled := false

		router.Use(Uuid)
		router.Use(func(c *gin.Context) {
			nextCalled = true
			c.Next()
		})
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if !nextCalled {
			t.Error("next middleware should be called")
		}
	})

	t.Run("UUID persists through middleware chain", func(t *testing.T) {
		router := gin.New()

		var uuid1, uuid2 string

		router.Use(Uuid)
		router.Use(func(c *gin.Context) {
			uuid1 = c.GetString(util.UuidKey)
			c.Next()
		})
		router.GET("/test", func(c *gin.Context) {
			uuid2 = c.GetString(util.UuidKey)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if uuid1 == "" || uuid2 == "" {
			t.Error("UUID should be available in all middleware")
		}

		if uuid1 != uuid2 {
			t.Error("UUID should be the same throughout the request lifecycle")
		}
	})
}
