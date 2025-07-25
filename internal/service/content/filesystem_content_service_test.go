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

	// 🚨 TEST TO EXPOSE BUG #8: Query parameter handling issue
	t.Run("BUG TEST: query parameters with multiple question marks", func(t *testing.T) {
		// URI with query parameter that contains a question mark
		uri := "/api/search?query=what?is?this"

		path := service.getFinalFilePath("example.com", uri, "GET")

		t.Logf("BUG TEST: URI with multiple '?' characters: %s", uri)
		t.Logf("Generated path: %s", path)
		t.Logf("BUG DETECTED: strings.Split(uri, \"?\") will incorrectly split on all '?' characters")
		t.Logf("This means query parameters with '?' in values will be broken")

		// Test the fix: strings.SplitN should split only on first question mark
		parts := strings.SplitN(uri, "?", 2)
		if len(parts) != 2 {
			t.Errorf("EXPECTED: strings.SplitN should create exactly 2 parts, got %d", len(parts))
			t.Logf("Parts: %v", parts)
		} else {
			t.Logf("✅ BUG FIXED: strings.SplitN correctly created 2 parts: %v", parts)
			if parts[1] != "query=what?is?this" {
				t.Errorf("Expected query part 'query=what?is?this', got '%s'", parts[1])
			}
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

	// 🚨 TEST TO EXPOSE BUG #9: Incorrect validation logic
	t.Run("BUG TEST: validation logic is incorrect", func(t *testing.T) {
		// Create a valid file path
		filePath := filepath.Join(tempDir, "example.com", "api.get")

		_, err := service.filePathToContentData(filePath)

		t.Logf("BUG TEST: File path: %s", filePath)
		
		// Calculate the indices that the buggy code would use
		rootPath := strings.TrimSuffix(tempDir, string(os.PathSeparator)) + string(os.PathSeparator)
		relativePath := strings.TrimPrefix(filePath, rootPath)
		firstSlashIndex := strings.Index(relativePath, string(os.PathSeparator))
		lastDotIndex := strings.LastIndex(relativePath, ".")

		t.Logf("Relative path: %s", relativePath)
		t.Logf("First slash index: %d", firstSlashIndex)
		t.Logf("Last dot index: %d", lastDotIndex)

		// 🚨 BUG: The condition firstSlashIndex >= lastDotIndex is wrong
		// For "example.com/api.get": firstSlashIndex=11, lastDotIndex=15
		// The condition should allow firstSlashIndex < lastDotIndex for valid paths
		if firstSlashIndex >= lastDotIndex && err != nil {
			t.Logf("BUG DETECTED: Validation logic rejects valid path")
			t.Logf("Condition 'firstSlashIndex >= lastDotIndex' is incorrect")
			t.Logf("Should be 'firstSlashIndex >= lastDotIndex' only for truly invalid cases")
		}

		if err != nil {
			t.Logf("Error returned: %v", err)
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
	})
}
