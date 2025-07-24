package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestLogger(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("logs request received and finished", func(t *testing.T) {
		// Capture log output
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

		router := gin.New()
		
		// Set UUID first, then logger
		router.Use(func(c *gin.Context) {
			c.Set(util.UuidKey, "test-uuid-123")
			c.Next()
		})
		router.Use(Logger)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
		req.Host = "example.com"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		logOutput := buf.String()
		
		// Verify request received log
		if !strings.Contains(logOutput, "request received") {
			t.Error("should log 'request received'")
		}

		// Verify request finished log
		if !strings.Contains(logOutput, "request finished") {
			t.Error("should log 'request finished'")
		}

		// Verify UUID is logged
		if !strings.Contains(logOutput, "test-uuid-123") {
			t.Error("should log the UUID")
		}

		// Verify host is logged
		if !strings.Contains(logOutput, "example.com") {
			t.Error("should log the host")
		}

		// Verify URI is logged
		if !strings.Contains(logOutput, "/test?param=value") {
			t.Error("should log the URI")
		}

		// Verify method is logged
		if !strings.Contains(logOutput, "GET") {
			t.Error("should log the HTTP method")
		}
	})

	t.Run("logs status code and latency", func(t *testing.T) {
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

		router := gin.New()
		
		router.Use(func(c *gin.Context) {
			c.Set(util.UuidKey, "test-uuid-456")
			c.Next()
		})
		router.Use(Logger)
		router.GET("/test", func(c *gin.Context) {
			// Add a small delay to ensure measurable latency
			time.Sleep(1 * time.Millisecond)
			c.Status(http.StatusCreated)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		logOutput := buf.String()

		// Verify status code is logged
		if !strings.Contains(logOutput, "201") {
			t.Error("should log the status code")
		}

		// Verify latency is logged (should contain some time measurement)
		if !strings.Contains(logOutput, "latency") {
			t.Error("should log the latency")
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				var buf bytes.Buffer
				log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

				router := gin.New()
				
				router.Use(func(c *gin.Context) {
					c.Set(util.UuidKey, "test-uuid-"+method)
					c.Next()
				})
				router.Use(Logger)
				router.Any("/test", func(c *gin.Context) {
					c.Status(http.StatusOK)
				})

				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				logOutput := buf.String()

				if !strings.Contains(logOutput, method) {
					t.Errorf("should log the HTTP method %s", method)
				}
			})
		}
	})

	t.Run("handles missing UUID gracefully", func(t *testing.T) {
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

		router := gin.New()
		
		// Don't set UUID to test graceful handling
		router.Use(Logger)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not panic and should still log
		logOutput := buf.String()
		if !strings.Contains(logOutput, "request received") {
			t.Error("should still log even without UUID")
		}
	})

	t.Run("logs structured data correctly", func(t *testing.T) {
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

		router := gin.New()
		
		router.Use(func(c *gin.Context) {
			c.Set(util.UuidKey, "structured-test-uuid")
			c.Next()
		})
		router.Use(Logger)
		router.POST("/api/users", func(c *gin.Context) {
			c.Status(http.StatusAccepted)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
		req.Host = "api.example.com:8080"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		logOutput := buf.String()
		
		// Parse the log lines to verify structured logging
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		
		for _, line := range lines {
			if strings.Contains(line, "request received") || strings.Contains(line, "request finished") {
				var logEntry map[string]interface{}
				if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
					t.Errorf("log should be valid JSON: %v", err)
					continue
				}

				// Verify required fields are present
				if logEntry["uuid"] != "structured-test-uuid" {
					t.Error("log should contain correct UUID")
				}

				if logEntry["host"] != "api.example.com:8080" {
					t.Error("log should contain correct host")
				}

				if logEntry["uri"] != "/api/users" {
					t.Error("log should contain correct URI")
				}

				if logEntry["method"] != "POST" {
					t.Error("log should contain correct method")
				}

				if strings.Contains(line, "request finished") {
					if _, exists := logEntry["status_code"]; !exists {
						t.Error("finished log should contain status_code")
					}

					if _, exists := logEntry["latency"]; !exists {
						t.Error("finished log should contain latency")
					}
				}
			}
		}
	})

	t.Run("measures latency correctly", func(t *testing.T) {
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

		router := gin.New()
		
		router.Use(func(c *gin.Context) {
			c.Set(util.UuidKey, "latency-test-uuid")
			c.Next()
		})
		router.Use(Logger)
		router.GET("/slow", func(c *gin.Context) {
			// Add a measurable delay
			time.Sleep(10 * time.Millisecond)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		router.ServeHTTP(w, req)
		actualDuration := time.Since(start)

		logOutput := buf.String()
		
		// Find the finished log line
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		for _, line := range lines {
			if strings.Contains(line, "request finished") {
				var logEntry map[string]interface{}
				if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
					continue
				}

				latencyStr, ok := logEntry["latency"].(string)
				if !ok {
					t.Error("latency should be a string")
					continue
				}

				// Parse the latency duration
				latency, err := time.ParseDuration(latencyStr)
				if err != nil {
					t.Errorf("latency should be a valid duration: %v", err)
					continue
				}

				// Verify latency is reasonable (should be at least 10ms due to our sleep)
				if latency < 10*time.Millisecond {
					t.Errorf("logged latency (%v) should be at least 10ms", latency)
				}

				// Verify latency is not wildly different from actual duration
				if latency > actualDuration*2 {
					t.Errorf("logged latency (%v) seems too high compared to actual (%v)", latency, actualDuration)
				}

				break
			}
		}
	})

	// Restore original logger after tests
	t.Cleanup(func() {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	})
}
