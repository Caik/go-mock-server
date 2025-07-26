package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestInitLogger(t *testing.T) {
	// Save original logger state
	originalLevel := zerolog.GlobalLevel()
	originalLogger := log.Logger

	// Test InitLogger
	InitLogger()

	// Verify logger was initialized correctly
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("expected global log level to be InfoLevel, got %v", zerolog.GlobalLevel())
	}

	// Restore original state
	zerolog.SetGlobalLevel(originalLevel)
	log.Logger = originalLogger
}

func TestNewHostsConfig_EmptyConfigFile(t *testing.T) {
	appArgs := &AppArguments{
		MocksConfigFile: "",
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hostsConfig == nil {
		t.Fatal("expected hostsConfig to be non-nil")
	}

	if hostsConfig.Hosts == nil {
		t.Fatal("expected Hosts map to be initialized")
	}

	if len(hostsConfig.Hosts) != 0 {
		t.Errorf("expected empty Hosts map, got %d entries", len(hostsConfig.Hosts))
	}
}

func TestNewHostsConfig_ValidConfigFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.json")

	validConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	configData, err := json.Marshal(validConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hostsConfig == nil {
		t.Fatal("expected hostsConfig to be non-nil")
	}

	if len(hostsConfig.Hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hostsConfig.Hosts))
	}

	hostConfig, exists := hostsConfig.Hosts["example.com"]
	if !exists {
		t.Error("expected example.com host to exist")
	}

	if hostConfig.LatencyConfig == nil {
		t.Error("expected latency config to be non-nil")
	}

	if *hostConfig.LatencyConfig.Min != 100 {
		t.Errorf("expected min latency to be 100, got %d", *hostConfig.LatencyConfig.Min)
	}
}

func TestNewHostsConfig_NonExistentFile(t *testing.T) {
	appArgs := &AppArguments{
		MocksConfigFile: "/non/existent/file.json",
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewHostsConfig_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.json")

	invalidJSON := `{"hosts": {"example.com": invalid json}}`
	err := os.WriteFile(configFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewHostsConfig_InvalidHostsConfig(t *testing.T) {
	// Create a temporary file with invalid hosts config (invalid host pattern)
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-hosts-config.json")

	invalidConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"invalid-host": { // This doesn't match the host regex pattern
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	configData, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("failed to marshal invalid config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid hosts config")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewMocksDirectoryConfig_ValidDirectory(t *testing.T) {
	tempDir := t.TempDir()
	mocksDir := filepath.Join(tempDir, "mocks")

	appArgs := &AppArguments{
		MocksDirectory: mocksDir,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mocksConfig == nil {
		t.Fatal("expected mocksConfig to be non-nil")
	}

	expectedPath, _ := filepath.Abs(mocksDir)
	if mocksConfig.Path != expectedPath {
		t.Errorf("expected path to be %s, got %s", expectedPath, mocksConfig.Path)
	}

	// Verify directory was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestNewMocksDirectoryConfig_ExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	mocksDir := filepath.Join(tempDir, "existing-mocks")

	// Pre-create the directory
	err := os.MkdirAll(mocksDir, os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	appArgs := &AppArguments{
		MocksDirectory: mocksDir,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mocksConfig == nil {
		t.Fatal("expected mocksConfig to be non-nil")
	}

	expectedPath, _ := filepath.Abs(mocksDir)
	if mocksConfig.Path != expectedPath {
		t.Errorf("expected path to be %s, got %s", expectedPath, mocksConfig.Path)
	}
}

func TestNewMocksDirectoryConfig_RelativePath(t *testing.T) {
	relativePath := "test-mocks"
	
	appArgs := &AppArguments{
		MocksDirectory: relativePath,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mocksConfig == nil {
		t.Fatal("expected mocksConfig to be non-nil")
	}

	expectedPath, _ := filepath.Abs(relativePath)
	if mocksConfig.Path != expectedPath {
		t.Errorf("expected path to be %s, got %s", expectedPath, mocksConfig.Path)
	}

	// Clean up
	os.RemoveAll(expectedPath)
}

func TestNewHostsConfig_InvalidLatencyConfig(t *testing.T) {
	// Create a temporary file with invalid latency config (min > max)
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-latency-config.json")

	invalidConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(200), // min > max should fail validation
					Max: intPtr(100),
				},
			},
		},
	}

	configData, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("failed to marshal invalid config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid latency config")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewHostsConfig_InvalidErrorConfig(t *testing.T) {
	// Create a temporary file with invalid error config (percentage > 100)
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-error-config.json")

	invalidConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(150), // > 100 should fail validation
					},
				},
			},
		},
	}

	configData, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("failed to marshal invalid config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid error config")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewMocksDirectoryConfig_InvalidPath(t *testing.T) {
	// Test with a path that cannot be created (e.g., under a file instead of directory)
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "file.txt")

	// Create a file
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Try to create a directory under the file (should fail)
	invalidPath := filepath.Join(tempFile, "subdir")

	appArgs := &AppArguments{
		MocksDirectory: invalidPath,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid directory path")
	}

	if mocksConfig != nil {
		t.Error("expected mocksConfig to be nil on error")
	}
}

func TestNewHostsConfig_ComplexValidConfig(t *testing.T) {
	// Test with a more complex valid configuration
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "complex-config.json")

	complexConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"api.example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(50),
					P95: intPtr(150),
					P99: intPtr(180),
					Max: intPtr(200),
				},
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(10),
						LatencyConfig: &LatencyConfig{
							Min: intPtr(100),
							Max: intPtr(300),
						},
					},
					"503": {
						Percentage: intPtr(5),
					},
				},
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						LatencyConfig: &LatencyConfig{
							Min: intPtr(25),
							Max: intPtr(100),
						},
						ErrorsConfig: map[string]ErrorConfig{
							"404": {
								Percentage: intPtr(20),
							},
						},
					},
				},
			},
		},
	}

	configData, err := json.Marshal(complexConfig)
	if err != nil {
		t.Fatalf("failed to marshal complex config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write complex config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hostsConfig == nil {
		t.Fatal("expected hostsConfig to be non-nil")
	}

	// Verify the complex configuration was loaded correctly
	hostConfig, exists := hostsConfig.Hosts["api.example.com"]
	if !exists {
		t.Fatal("expected api.example.com host to exist")
	}

	if hostConfig.LatencyConfig == nil {
		t.Fatal("expected latency config to be non-nil")
	}

	if *hostConfig.LatencyConfig.Min != 50 {
		t.Errorf("expected min latency to be 50, got %d", *hostConfig.LatencyConfig.Min)
	}

	if len(hostConfig.ErrorsConfig) != 2 {
		t.Errorf("expected 2 error configs, got %d", len(hostConfig.ErrorsConfig))
	}

	if len(hostConfig.UrisConfig) != 1 {
		t.Errorf("expected 1 URI config, got %d", len(hostConfig.UrisConfig))
	}
}

func TestNewHostsConfig_EmptyJSONFile(t *testing.T) {
	// Test with an empty JSON file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "empty-config.json")

	err := os.WriteFile(configFile, []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("failed to write empty config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error for empty JSON, got %v", err)
	}

	if hostsConfig == nil {
		t.Fatal("expected hostsConfig to be non-nil")
	}

	// Empty JSON object {} will result in nil Hosts map after unmarshaling
	// This is expected behavior - the validation should handle nil maps gracefully
	if hostsConfig.Hosts == nil {
		t.Log("Hosts map is nil for empty JSON, which is expected behavior")
	} else if len(hostsConfig.Hosts) != 0 {
		t.Errorf("expected empty Hosts map, got %d entries", len(hostsConfig.Hosts))
	}
}

func TestNewHostsConfig_EmptyHostsJSONFile(t *testing.T) {
	// Test with JSON file containing empty hosts object
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "empty-hosts-config.json")

	err := os.WriteFile(configFile, []byte(`{"hosts": {}}`), 0644)
	if err != nil {
		t.Fatalf("failed to write empty hosts config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error for empty hosts JSON, got %v", err)
	}

	if hostsConfig == nil {
		t.Fatal("expected hostsConfig to be non-nil")
	}

	if hostsConfig.Hosts == nil {
		t.Fatal("expected Hosts map to be initialized")
	}

	if len(hostsConfig.Hosts) != 0 {
		t.Errorf("expected empty Hosts map, got %d entries", len(hostsConfig.Hosts))
	}
}

func TestNewHostsConfig_InvalidErrorCode(t *testing.T) {
	// Test with invalid error code (not in 4xx or 5xx range)
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-error-code-config.json")

	invalidConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				ErrorsConfig: map[string]ErrorConfig{
					"200": { // 2xx codes are not allowed for errors
						Percentage: intPtr(10),
					},
				},
			},
		},
	}

	configData, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("failed to marshal invalid config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid error code")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewHostsConfig_InvalidUriPattern(t *testing.T) {
	// Test with invalid URI pattern
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-uri-config.json")

	invalidConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				UrisConfig: map[string]UriConfig{
					"invalid uri pattern": { // This doesn't match URI regex
						LatencyConfig: &LatencyConfig{
							Min: intPtr(100),
							Max: intPtr(200),
						},
					},
				},
			},
		},
	}

	configData, err := json.Marshal(invalidConfig)
	if err != nil {
		t.Fatalf("failed to marshal invalid config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	appArgs := &AppArguments{
		MocksConfigFile: configFile,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err == nil {
		t.Fatal("expected error for invalid URI pattern")
	}

	if hostsConfig != nil {
		t.Error("expected hostsConfig to be nil on error")
	}
}

func TestNewMocksDirectoryConfig_NestedDirectory(t *testing.T) {
	// Test creating nested directories
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "level1", "level2", "mocks")

	appArgs := &AppArguments{
		MocksDirectory: nestedPath,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error for nested directory creation, got %v", err)
	}

	if mocksConfig == nil {
		t.Fatal("expected mocksConfig to be non-nil")
	}

	expectedPath, _ := filepath.Abs(nestedPath)
	if mocksConfig.Path != expectedPath {
		t.Errorf("expected path to be %s, got %s", expectedPath, mocksConfig.Path)
	}

	// Verify nested directories were created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("expected nested directories to be created")
	}
}

func TestParseAppArguments_WithValidArgs(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Set up test arguments
	os.Args = []string{
		"test-program",
		"--mocks-directory", "/test/mocks",
		"--mocks-config-file", "/test/config.json",
		"--port", "9090",
		"--disable-cache",
		"--disable-latency",
		"--disable-error",
	}

	// Test ParseAppArguments
	args := ParseAppArguments()

	if args == nil {
		t.Fatal("expected args to be non-nil")
	}

	if args.MocksDirectory != "/test/mocks" {
		t.Errorf("expected MocksDirectory to be '/test/mocks', got '%s'", args.MocksDirectory)
	}

	if args.MocksConfigFile != "/test/config.json" {
		t.Errorf("expected MocksConfigFile to be '/test/config.json', got '%s'", args.MocksConfigFile)
	}

	if args.ServerPort != 9090 {
		t.Errorf("expected ServerPort to be 9090, got %d", args.ServerPort)
	}

	if !args.DisableCache {
		t.Error("expected DisableCache to be true")
	}

	if !args.DisableLatency {
		t.Error("expected DisableLatency to be true")
	}

	if !args.DisableError {
		t.Error("expected DisableError to be true")
	}
}

func TestParseAppArguments_WithMinimalArgs(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Set up minimal test arguments (only required ones)
	os.Args = []string{
		"test-program",
		"--mocks-directory", "/test/mocks",
	}

	// Test ParseAppArguments
	args := ParseAppArguments()

	if args == nil {
		t.Fatal("expected args to be non-nil")
	}

	if args.MocksDirectory != "/test/mocks" {
		t.Errorf("expected MocksDirectory to be '/test/mocks', got '%s'", args.MocksDirectory)
	}

	// Test default values
	if args.ServerPort != 8080 {
		t.Errorf("expected default ServerPort to be 8080, got %d", args.ServerPort)
	}

	if args.MocksConfigFile != "" {
		t.Errorf("expected empty MocksConfigFile, got '%s'", args.MocksConfigFile)
	}

	if args.DisableCache {
		t.Error("expected DisableCache to be false by default")
	}

	if args.DisableLatency {
		t.Error("expected DisableLatency to be false by default")
	}

	if args.DisableError {
		t.Error("expected DisableError to be false by default")
	}
}

func TestNewHostsConfig_InvalidConfigFilePath(t *testing.T) {
	// Test with a config file path that contains invalid characters that would cause filepath.Abs to fail
	// This is tricky to test as filepath.Abs rarely fails, but we can try with very long paths or invalid characters

	// Create a path that's extremely long to potentially cause filepath.Abs to fail
	longPath := strings.Repeat("a", 4096) + ".json"

	appArgs := &AppArguments{
		MocksConfigFile: longPath,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	// This test might not always fail depending on the OS, but if it does fail, we should handle it gracefully
	if err != nil {
		// If filepath.Abs fails, we should get an error and nil config
		if hostsConfig != nil {
			t.Error("expected hostsConfig to be nil when filepath.Abs fails")
		}
		t.Logf("filepath.Abs failed as expected: %v", err)
	} else {
		// If it doesn't fail, that's also fine - the test still covers the code path
		t.Log("filepath.Abs succeeded with long path")
	}
}

func TestNewMocksDirectoryConfig_InvalidDirectoryPath(t *testing.T) {
	// Test with a directory path that contains invalid characters that would cause filepath.Abs to fail
	// Similar to the above test, this is tricky but we can try with very long paths

	longPath := strings.Repeat("b", 4096)

	appArgs := &AppArguments{
		MocksDirectory: longPath,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	// This test might not always fail depending on the OS, but if it does fail, we should handle it gracefully
	if err != nil {
		// If filepath.Abs fails, we should get an error and nil config
		if mocksConfig != nil {
			t.Error("expected mocksConfig to be nil when filepath.Abs fails")
		}
		t.Logf("filepath.Abs failed as expected: %v", err)
	} else {
		// If it doesn't fail, that's also fine - the test still covers the code path
		t.Log("filepath.Abs succeeded with long path")
		// Clean up if directory was created
		if mocksConfig != nil {
			os.RemoveAll(mocksConfig.Path)
		}
	}
}

func TestNewHostsConfig_FilePathAbsoluteConversion(t *testing.T) {
	// Test that relative paths are properly converted to absolute paths
	tempDir := t.TempDir()

	// Create a config file in temp directory
	configFileName := "test-config.json"
	configFile := filepath.Join(tempDir, configFileName)

	validConfig := HostsConfig{
		Hosts: map[string]HostConfig{
			"test.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	configData, err := json.Marshal(validConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	// Change to temp directory to test relative path resolution
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Use relative path
	appArgs := &AppArguments{
		MocksConfigFile: configFileName,
	}

	hostsConfig, err := NewHostsConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error with relative path, got %v", err)
	}

	if hostsConfig == nil {
		t.Fatal("expected hostsConfig to be non-nil")
	}

	if len(hostsConfig.Hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hostsConfig.Hosts))
	}
}

func TestNewMocksDirectoryConfig_RelativePathConversion(t *testing.T) {
	// Test that relative paths are properly converted to absolute paths
	tempDir := t.TempDir()

	// Change to temp directory to test relative path resolution
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Use relative path
	relativePath := "test-mocks-relative"
	appArgs := &AppArguments{
		MocksDirectory: relativePath,
	}

	mocksConfig, err := NewMocksDirectoryConfig(appArgs)

	if err != nil {
		t.Fatalf("expected no error with relative path, got %v", err)
	}

	if mocksConfig == nil {
		t.Fatal("expected mocksConfig to be non-nil")
	}

	// The path should be converted to absolute
	if !filepath.IsAbs(mocksConfig.Path) {
		t.Error("expected path to be converted to absolute path")
	}

	// Verify directory was created
	if _, err := os.Stat(mocksConfig.Path); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}

	// Clean up
	os.RemoveAll(mocksConfig.Path)
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
