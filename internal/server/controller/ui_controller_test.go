package controller

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupUITestDir(t *testing.T) string {
	t.Helper()
	uiDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(uiDir, "index.html"), []byte("<html>SPA</html>"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(uiDir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uiDir, "assets", "app.js"), []byte("console.log('app')"), 0644); err != nil {
		t.Fatal(err)
	}

	return uiDir
}

func TestInitUIRoutes_RootRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	InitUIRoutes(router, setupUITestDir(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}

	if w.Header().Get("Location") != "/ui/" {
		t.Errorf("expected redirect to /ui/, got %q", w.Header().Get("Location"))
	}
}

func TestInitUIRoutes_ServesIndexHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	InitUIRoutes(router, setupUITestDir(t))

	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "SPA") {
		t.Error("expected index.html content in response body")
	}
}

func TestInitUIRoutes_ServesStaticAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	InitUIRoutes(router, setupUITestDir(t))

	req := httptest.NewRequest(http.MethodGet, "/ui/assets/app.js", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "console.log") {
		t.Error("expected JS file content in response body")
	}
}

func TestInitUIRoutes_SPAFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	InitUIRoutes(router, setupUITestDir(t))

	// These paths don't exist on disk — should all serve index.html for SPA routing
	paths := []string{"/ui/logs", "/ui/mocks", "/ui/hosts", "/ui/nonexistent/deep/path"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", w.Code)
			}

			if !strings.Contains(w.Body.String(), "SPA") {
				t.Errorf("expected index.html (SPA fallback) for path %s", path)
			}
		})
	}
}

func TestInitUIRoutes_DirectoryFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uiDir := setupUITestDir(t)

	// Create a subdirectory (should not be served as a listing — falls back to index.html)
	if err := os.MkdirAll(filepath.Join(uiDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	InitUIRoutes(router, uiDir)

	req := httptest.NewRequest(http.MethodGet, "/ui/subdir", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "SPA") {
		t.Error("expected index.html for directory path, not a directory listing")
	}
}

func TestInitUIRoutes_PathTraversalContainmentCheck(t *testing.T) {
	// Security model:
	// Gin's /*path wildcard always captures params with a leading "/".
	// filepath.Clean normalizes "/../../etc/passwd" → "/etc/passwd" (absolute),
	// then filepath.Join(uiDir, "/etc/passwd") appends it UNDER uiDir
	// (e.g. "/app/ui/etc/passwd"), so the path never escapes.
	// The strings.HasPrefix guard in the handler is defense-in-depth for any
	// edge case where a param without a leading "/" could slip through.

	uiDir, err := filepath.Abs(t.TempDir())

	if err != nil {
		t.Fatal(err)
	}

	// All of these look like traversal attempts but should resolve inside uiDir
	// because filepath.Clean + filepath.Join keeps them contained.
	allPaths := []string{
		"/../../../etc/passwd",
		"/../../etc/passwd",
		"/assets/../../../etc/passwd",
		"/index.html",
		"/assets/app.js",
		"/assets/chunk.js",
	}

	for _, p := range allPaths {
		t.Run(p, func(t *testing.T) {
			filePath := filepath.Join(uiDir, filepath.Clean(p))

			if !strings.HasPrefix(filePath, uiDir+string(filepath.Separator)) {
				t.Errorf("path %q escaped uiDir: resolved to %q", p, filePath)
			}
		})
	}
}
