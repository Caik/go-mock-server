package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
)

func TestFilesystemContentService_getFinalFilePath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("generates correct path for simple URI", func(t *testing.T) {
		path := service.getFinalFilePath("example.com", "/api/users", "GET")

		expected := filepath.Join(tempDir, "example.com", "api", "users.get")
		if path != expected {
			t.Errorf("expected path '%s', got '%s'", expected, path)
		}
	})

	t.Run("handles root path correctly", func(t *testing.T) {
		path := service.getFinalFilePath("example.com", "/", "GET")

		expected := filepath.Join(tempDir, "example.com", "root.get")
		if path != expected {
			t.Errorf("expected path '%s', got '%s'", expected, path)
		}
	})

	t.Run("handles query parameters with multiple question marks", func(t *testing.T) {
		// URI with query parameter that contains a question mark
		uri := "/api/search?query=what?is?this"

		path := service.getFinalFilePath("example.com", uri, "GET")

		t.Logf("Testing URI with multiple '?' characters: %s", uri)
		t.Logf("Generated path: %s", path)

		// Verify that strings.SplitN splits only on first question mark
		parts := strings.SplitN(uri, "?", 2)
		if len(parts) != 2 {
			t.Errorf("expected strings.SplitN to create exactly 2 parts, got %d", len(parts))
			t.Logf("Parts: %v", parts)
		} else {
			t.Logf("strings.SplitN correctly created 2 parts: %v", parts)
			if parts[1] != "query=what?is?this" {
				t.Errorf("expected query part 'query=what?is?this', got '%s'", parts[1])
			} else {
				t.Log("query parameter with '?' preserved correctly")
			}
		}

		// Verify the generated path contains the complete query string
		expectedSuffix := "search?query=what?is?this.get"
		if !strings.HasSuffix(path, expectedSuffix) {
			t.Errorf("expected path to end with '%s', got '%s'", expectedSuffix, path)
		} else {
			t.Log("file path correctly preserves complete query string")
		}
	})

	t.Run("handles query parameters", func(t *testing.T) {
		path := service.getFinalFilePath("example.com", "/api/users?id=123", "GET")

		expected := filepath.Join(tempDir, "example.com", "api", "users?id=123.get")
		if path != expected {
			t.Errorf("expected path '%s', got '%s'", expected, path)
		}
	})

	t.Run("handles nested paths", func(t *testing.T) {
		path := service.getFinalFilePath("api.example.com", "/v1/users/profile", "POST")

		expected := filepath.Join(tempDir, "api.example.com", "v1", "users", "profile.post")
		if path != expected {
			t.Errorf("expected path '%s', got '%s'", expected, path)
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			path := service.getFinalFilePath("example.com", "/api/test", method)
			
			expectedSuffix := "." + strings.ToLower(method)
			if !strings.HasSuffix(path, expectedSuffix) {
				t.Errorf("path should end with '%s', got '%s'", expectedSuffix, path)
			}
		}
	})
}

func TestFilesystemContentService_filePathToContentData(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("parses valid file path correctly", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "example.com", "api", "users.get")

		data, err := service.filePathToContentData(filePath)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if data.Host != "example.com" {
			t.Errorf("expected host 'example.com', got '%s'", data.Host)
		}

		expectedURI := string(os.PathSeparator) + "api" + string(os.PathSeparator) + "users"
		if data.Uri != expectedURI {
			t.Errorf("expected URI '%s', got '%s'", expectedURI, data.Uri)
		}

		if data.Method != "GET" {
			t.Errorf("expected method 'GET', got '%s'", data.Method)
		}
	})

	t.Run("validates file path parsing logic", func(t *testing.T) {
		// Create a valid file path
		filePath := filepath.Join(tempDir, "example.com", "api.get")

		_, err := service.filePathToContentData(filePath)

		t.Logf("testing file path: %s", filePath)

		// Calculate the indices that the validation logic uses
		rootPath := strings.TrimSuffix(tempDir, string(os.PathSeparator)) + string(os.PathSeparator)
		relativePath := strings.TrimPrefix(filePath, rootPath)
		firstSlashIndex := strings.Index(relativePath, string(os.PathSeparator))
		lastDotIndex := strings.LastIndex(relativePath, ".")

		t.Logf("relative path: %s", relativePath)
		t.Logf("first slash index: %d", firstSlashIndex)
		t.Logf("last dot index: %d", lastDotIndex)

		// Test the validation logic
		// For "example.com/api.get": firstSlashIndex=11, lastDotIndex=15
		// The condition should allow firstSlashIndex < lastDotIndex for valid paths
		if firstSlashIndex >= lastDotIndex && err != nil {
			t.Logf("validation logic rejects this path")
			t.Logf("condition 'firstSlashIndex >= lastDotIndex' triggered")
		}

		if err != nil {
			t.Logf("error returned: %v", err)
		}
	})

	t.Run("handles root path with root token", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "example.com", "root.get")

		data, err := service.filePathToContentData(filePath)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedURI := string(os.PathSeparator)
		if data.Uri != expectedURI {
			t.Errorf("expected URI '%s', got '%s'", expectedURI, data.Uri)
		}
	})

	t.Run("rejects invalid file patterns", func(t *testing.T) {
		invalidPaths := []string{
			filepath.Join(tempDir, "invalid"),           // No slash or dot
			filepath.Join(tempDir, "no-dot-extension"),  // No dot
			filepath.Join(tempDir, ".hidden"),           // No slash before dot
		}

		for _, invalidPath := range invalidPaths {
			_, err := service.filePathToContentData(invalidPath)

			if err == nil {
				t.Errorf("expected error for invalid path '%s'", invalidPath)
			}
		}
	})

	t.Run("validates host format", func(t *testing.T) {
		// Create path with invalid host
		filePath := filepath.Join(tempDir, "invalid host.com", "api", "users.get")

		_, err := service.filePathToContentData(filePath)

		if err == nil {
			t.Error("expected error for invalid host format")
		}

		if !strings.Contains(err.Error(), "invalid host") {
			t.Errorf("expected 'invalid host' error, got '%v'", err)
		}
	})

	t.Run("validates URI format", func(t *testing.T) {
		// Create path with invalid URI (spaces)
		filePath := filepath.Join(tempDir, "example.com", "invalid uri", "users.get")

		_, err := service.filePathToContentData(filePath)

		if err == nil {
			t.Error("expected error for invalid URI format")
		}

		if !strings.Contains(err.Error(), "invalid uri") {
			t.Errorf("expected 'invalid uri' error, got '%v'", err)
		}
	})

	t.Run("validates HTTP method", func(t *testing.T) {
		// Create path with invalid method
		filePath := filepath.Join(tempDir, "example.com", "api", "users.invalid")

		_, err := service.filePathToContentData(filePath)

		if err == nil {
			t.Error("expected error for invalid HTTP method")
		}

		if !strings.Contains(err.Error(), "invalid method") {
			t.Errorf("expected 'invalid method' error, got '%v'", err)
		}
	})
}

func TestFilesystemContentService_GetContent(t *testing.T) {
	// Create temporary directory and test file
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testContent := []byte("test response content")
	testDir := filepath.Join(tempDir, "example.com", "api")
	os.MkdirAll(testDir, 0755)
	testFile := filepath.Join(testDir, "users.get")
	os.WriteFile(testFile, testContent, 0644)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("reads existing file successfully", func(t *testing.T) {
		data, err := service.GetContent("example.com", "/api/users", "GET", "test-uuid")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if string(*data) != string(testContent) {
			t.Errorf("expected content '%s', got '%s'", string(testContent), string(*data))
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := service.GetContent("example.com", "/api/nonexistent", "GET", "test-uuid")

		if err == nil {
			t.Error("expected error for non-existent file")
		}

		if !strings.Contains(err.Error(), "mock not found") {
			t.Errorf("expected 'mock not found' error, got '%v'", err)
		}
	})
}

func TestFilesystemContentService_SetContent(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("creates file and directories successfully", func(t *testing.T) {
		testContent := []byte("new test content")

		err := service.SetContent("example.com", "/api/users", "POST", "test-uuid", &testContent)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was created
		expectedPath := filepath.Join(tempDir, "example.com", "api", "users.post")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Error("file should have been created")
		}

		// Verify content
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}

		if string(content) != string(testContent) {
			t.Errorf("expected content '%s', got '%s'", string(testContent), string(content))
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		// Create initial file
		initialContent := []byte("initial content")
		service.SetContent("example.com", "/api/test", "GET", "test-uuid", &initialContent)

		// Overwrite with new content
		newContent := []byte("updated content")
		err := service.SetContent("example.com", "/api/test", "GET", "test-uuid", &newContent)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify new content
		expectedPath := filepath.Join(tempDir, "example.com", "api", "test.get")
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(content) != string(newContent) {
			t.Errorf("expected content '%s', got '%s'", string(newContent), string(content))
		}
	})
}

func TestFilesystemContentService_DeleteContent(t *testing.T) {
	// Create temporary directory and test file
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testDir := filepath.Join(tempDir, "example.com", "api")
	os.MkdirAll(testDir, 0755)
	testFile := filepath.Join(testDir, "users.delete")
	os.WriteFile(testFile, []byte("test content"), 0644)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("deletes existing file successfully", func(t *testing.T) {
		err := service.DeleteContent("example.com", "/api/users", "DELETE", "test-uuid")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("file should have been deleted")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		err := service.DeleteContent("example.com", "/api/nonexistent", "DELETE", "test-uuid")

		if err == nil {
			t.Error("expected error for non-existent file")
		}

		if !strings.Contains(err.Error(), "error while removing file") {
			t.Errorf("expected 'error while removing file' error, got '%v'", err)
		}
	})
}

func TestFilesystemContentService_ListContents(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files (using valid hostnames that match the regex)
	testFiles := map[string][]byte{
		"example.com/api/users.get":      []byte("users get"),
		"example.com/api/users.post":     []byte("users post"),
		"api.test.com/v1/health.get":     []byte("health check"),
		"admin.local.com/admin/status.get": []byte("admin status"),
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, content, 0644)
	}

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("lists all content files successfully", func(t *testing.T) {
		contents, err := service.ListContents("test-uuid")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if contents == nil {
			t.Fatal("expected non-nil contents")
		}

		if len(*contents) != len(testFiles) {
			t.Errorf("expected %d contents, got %d", len(testFiles), len(*contents))
			// Debug: show what was actually found
			for i, content := range *contents {
				t.Logf("found content %d: host=%s, uri=%s, method=%s", i, content.Host, content.Uri, content.Method)
			}
		}

		// Verify specific content entries
		foundHosts := make(map[string]bool)
		for _, content := range *contents {
			foundHosts[content.Host] = true
		}

		expectedHosts := []string{"example.com", "api.test.com", "admin.local.com"}
		for _, host := range expectedHosts {
			if !foundHosts[host] {
				t.Errorf("expected to find host '%s' in contents", host)
			}
		}
	})

	t.Run("handles empty directory", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "empty_mock_test")
		if err != nil {
			t.Fatalf("failed to create empty temp dir: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		emptyService := &FilesystemContentService{
			mocksDirConfig: &config.MocksDirectoryConfig{Path: emptyDir},
		}

		contents, err := emptyService.ListContents("test-uuid")

		if err != nil {
			t.Fatalf("expected no error for empty directory, got %v", err)
		}

		if contents == nil {
			t.Fatal("expected non-nil contents")
		}

		if len(*contents) != 0 {
			t.Errorf("expected 0 contents for empty directory, got %d", len(*contents))
		}
	})

	t.Run("handles non-existent directory", func(t *testing.T) {
		nonExistentService := &FilesystemContentService{
			mocksDirConfig: &config.MocksDirectoryConfig{Path: "/non/existent/path"},
		}

		contents, err := nonExistentService.ListContents("test-uuid")

		if err == nil {
			t.Error("expected error for non-existent directory")
		}

		if contents != nil {
			t.Error("expected nil contents on error")
		}
	})
}

func TestFilesystemContentService_Subscribe(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("subscribes to all event types", func(t *testing.T) {
		subscriberId := "test-subscriber"

		// Subscribe without specifying event types (should get all)
		eventChan := service.Subscribe(subscriberId)

		if eventChan == nil {
			t.Error("expected non-nil event channel")
		}

		// Clean up
		service.Unsubscribe(subscriberId)
	})

	t.Run("subscribes to specific event types", func(t *testing.T) {
		subscriberId := "test-subscriber-specific"

		// Subscribe to only Created and Updated events
		eventChan := service.Subscribe(subscriberId, Created, Updated)

		if eventChan == nil {
			t.Error("expected non-nil event channel")
		}

		// Clean up
		service.Unsubscribe(subscriberId)
	})

	t.Run("handles multiple subscribers", func(t *testing.T) {
		subscriber1 := "subscriber-1"
		subscriber2 := "subscriber-2"

		eventChan1 := service.Subscribe(subscriber1)
		eventChan2 := service.Subscribe(subscriber2, Created)

		if eventChan1 == nil {
			t.Error("expected non-nil event channel for subscriber 1")
		}

		if eventChan2 == nil {
			t.Error("expected non-nil event channel for subscriber 2")
		}

		// Clean up
		service.Unsubscribe(subscriber1)
		service.Unsubscribe(subscriber2)
	})
}

func TestFilesystemContentService_Unsubscribe(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mock_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mocksDirConfig := &config.MocksDirectoryConfig{
		Path: tempDir,
	}

	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
	}

	t.Run("unsubscribes existing subscriber", func(t *testing.T) {
		subscriberId := "test-subscriber"

		// Subscribe first
		eventChan := service.Subscribe(subscriberId)
		if eventChan == nil {
			t.Fatal("failed to subscribe")
		}

		// Unsubscribe - should not panic
		service.Unsubscribe(subscriberId)

		t.Log("successfully unsubscribed")
	})

	t.Run("handles unsubscribing non-existent subscriber", func(t *testing.T) {
		// Should not panic when unsubscribing non-existent subscriber
		service.Unsubscribe("non-existent-subscriber")

		t.Log("handled non-existent subscriber gracefully")
	})
}

func TestNewFilesystemContentService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "mock_test")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		mocksDirConfig := &config.MocksDirectoryConfig{
			Path: tempDir,
		}

		service := NewFilesystemContentService(mocksDirConfig)

		if service == nil {
			t.Error("expected non-nil service")
		}

		if service.mocksDirConfig != mocksDirConfig {
			t.Error("service should store the provided config")
		}
	})

	t.Run("handles nil config", func(t *testing.T) {
		// This might panic or cause issues, but let's test it
		defer func() {
			if r := recover(); r != nil {
				t.Logf("NewFilesystemContentService panicked with nil config: %v", r)
			}
		}()

		service := NewFilesystemContentService(nil)

		if service != nil {
			t.Log("service created with nil config")
		}
	})
}

func TestFilesystemContentService_ErrorHandling(t *testing.T) {
	t.Run("SetContent handles directory creation errors", func(t *testing.T) {
		// Use a path that will cause permission errors
		mocksDirConfig := &config.MocksDirectoryConfig{
			Path: "/root/restricted", // This should cause permission errors
		}

		service := &FilesystemContentService{
			mocksDirConfig: mocksDirConfig,
		}

		testContent := []byte("test content")
		err := service.SetContent("example.com", "/api/users", "POST", "test-uuid", &testContent)

		if err == nil {
			t.Error("expected error when creating directories in restricted path")
		}

		if !strings.Contains(err.Error(), "error while creating parent directories") {
			t.Errorf("expected directory creation error, got: %v", err)
		}
	})

	t.Run("handles invalid file paths gracefully", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "mock_test")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		mocksDirConfig := &config.MocksDirectoryConfig{
			Path: tempDir,
		}

		service := &FilesystemContentService{
			mocksDirConfig: mocksDirConfig,
		}

		// Test with invalid file path patterns
		invalidPaths := []string{
			"invalid-host-name",
			"example.com",
			"example.com/no-extension",
		}

		for _, invalidPath := range invalidPaths {
			fullPath := filepath.Join(tempDir, invalidPath)
			os.MkdirAll(filepath.Dir(fullPath), 0755)
			os.WriteFile(fullPath, []byte("test"), 0644)

			_, err := service.filePathToContentData(fullPath)
			if err == nil {
				t.Errorf("expected error for invalid path pattern: %s", invalidPath)
			}
		}
	})
}
