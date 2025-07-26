package config

import (
	"testing"
)

// Note: intPtr helper function is defined in config_test.go

// Test HostsConfig methods

func TestHostsConfig_GetHostConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	// Test existing host
	hostConfig := hostsConfig.GetHostConfig("example.com")
	if hostConfig == nil {
		t.Fatal("expected host config to be non-nil")
	}

	if hostConfig.LatencyConfig == nil {
		t.Error("expected latency config to be non-nil")
	}

	if *hostConfig.LatencyConfig.Min != 100 {
		t.Errorf("expected min latency to be 100, got %d", *hostConfig.LatencyConfig.Min)
	}

	// Test non-existing host
	nonExistentConfig := hostsConfig.GetHostConfig("nonexistent.com")
	if nonExistentConfig != nil {
		t.Error("expected nil for non-existent host")
	}
}

func TestHostsConfig_SetHostConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: make(map[string]HostConfig),
	}

	newConfig := HostConfig{
		LatencyConfig: &LatencyConfig{
			Min: intPtr(50),
			Max: intPtr(150),
		},
	}

	// Set new host config
	hostsConfig.SetHostConfig("test.com", newConfig)

	// Verify it was set
	retrievedConfig := hostsConfig.GetHostConfig("test.com")
	if retrievedConfig == nil {
		t.Fatal("expected host config to be set")
	}

	if *retrievedConfig.LatencyConfig.Min != 50 {
		t.Errorf("expected min latency to be 50, got %d", *retrievedConfig.LatencyConfig.Min)
	}

	// Update existing host config
	updatedConfig := HostConfig{
		LatencyConfig: &LatencyConfig{
			Min: intPtr(75),
			Max: intPtr(175),
		},
	}

	hostsConfig.SetHostConfig("test.com", updatedConfig)

	// Verify it was updated
	retrievedConfig = hostsConfig.GetHostConfig("test.com")
	if *retrievedConfig.LatencyConfig.Min != 75 {
		t.Errorf("expected updated min latency to be 75, got %d", *retrievedConfig.LatencyConfig.Min)
	}
}

func TestHostsConfig_DeleteHostConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
			"test.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(50),
					Max: intPtr(150),
				},
			},
		},
	}

	// Verify host exists before deletion
	if hostsConfig.GetHostConfig("example.com") == nil {
		t.Fatal("expected host to exist before deletion")
	}

	// Delete host
	hostsConfig.DeleteHostConfig("example.com")

	// Verify host was deleted
	if hostsConfig.GetHostConfig("example.com") != nil {
		t.Error("expected host to be deleted")
	}

	// Verify other host still exists
	if hostsConfig.GetHostConfig("test.com") == nil {
		t.Error("expected other host to still exist")
	}

	// Delete non-existent host (should not panic)
	hostsConfig.DeleteHostConfig("nonexistent.com")
}

func TestHostsConfig_UpdateHostLatencyConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	// Update existing host
	newLatencyConfig := &LatencyConfig{
		Min: intPtr(50),
		Max: intPtr(150),
		P95: intPtr(120),
	}

	updatedConfig, err := hostsConfig.UpdateHostLatencyConfig("example.com", newLatencyConfig)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedConfig == nil {
		t.Fatal("expected updated config to be non-nil")
	}

	if *updatedConfig.LatencyConfig.Min != 50 {
		t.Errorf("expected updated min latency to be 50, got %d", *updatedConfig.LatencyConfig.Min)
	}

	if *updatedConfig.LatencyConfig.P95 != 120 {
		t.Errorf("expected updated P95 latency to be 120, got %d", *updatedConfig.LatencyConfig.P95)
	}

	// Update non-existent host
	updatedConfig, err = hostsConfig.UpdateHostLatencyConfig("nonexistent.com", newLatencyConfig)

	if err != nil {
		t.Errorf("expected no error for non-existent host, got %v", err)
	}

	if updatedConfig != nil {
		t.Error("expected nil config for non-existent host")
	}
}

func TestHostsConfig_DeleteHostLatencyConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(10),
					},
				},
			},
		},
	}

	// Delete latency config from existing host
	updatedConfig, err := hostsConfig.DeleteHostLatencyConfig("example.com")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedConfig == nil {
		t.Fatal("expected updated config to be non-nil")
	}

	if updatedConfig.LatencyConfig != nil {
		t.Error("expected latency config to be nil after deletion")
	}

	// Verify other configs remain
	if len(updatedConfig.ErrorsConfig) != 1 {
		t.Error("expected errors config to remain")
	}

	// Delete from non-existent host
	updatedConfig, err = hostsConfig.DeleteHostLatencyConfig("nonexistent.com")

	if err != nil {
		t.Errorf("expected no error for non-existent host, got %v", err)
	}

	if updatedConfig != nil {
		t.Error("expected nil config for non-existent host")
	}
}

func TestHostsConfig_UpdateHostErrorsConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	// Update errors config
	newErrorsConfig := map[string]ErrorConfig{
		"500": {
			Percentage: intPtr(20),
		},
		"503": {
			Percentage: intPtr(10),
		},
	}

	updatedConfig, err := hostsConfig.UpdateHostErrorsConfig("example.com", newErrorsConfig)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedConfig == nil {
		t.Fatal("expected updated config to be non-nil")
	}

	if len(updatedConfig.ErrorsConfig) != 2 {
		t.Errorf("expected 2 error configs, got %d", len(updatedConfig.ErrorsConfig))
	}

	if *updatedConfig.ErrorsConfig["500"].Percentage != 20 {
		t.Errorf("expected 500 error percentage to be 20, got %d", *updatedConfig.ErrorsConfig["500"].Percentage)
	}

	// Update non-existent host
	updatedConfig, err = hostsConfig.UpdateHostErrorsConfig("nonexistent.com", newErrorsConfig)

	if err != nil {
		t.Errorf("expected no error for non-existent host, got %v", err)
	}

	if updatedConfig != nil {
		t.Error("expected nil config for non-existent host")
	}
}

func TestHostsConfig_DeleteHostErrorConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(20),
					},
					"503": {
						Percentage: intPtr(10),
					},
				},
			},
		},
	}

	// Delete specific error config
	updatedConfig, err := hostsConfig.DeleteHostErrorConfig("example.com", "500")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedConfig == nil {
		t.Fatal("expected updated config to be non-nil")
	}

	if len(updatedConfig.ErrorsConfig) != 1 {
		t.Errorf("expected 1 error config after deletion, got %d", len(updatedConfig.ErrorsConfig))
	}

	if _, exists := updatedConfig.ErrorsConfig["500"]; exists {
		t.Error("expected 500 error config to be deleted")
	}

	if _, exists := updatedConfig.ErrorsConfig["503"]; !exists {
		t.Error("expected 503 error config to remain")
	}

	// Delete from non-existent host
	updatedConfig, err = hostsConfig.DeleteHostErrorConfig("nonexistent.com", "500")

	if err != nil {
		t.Errorf("expected no error for non-existent host, got %v", err)
	}

	if updatedConfig != nil {
		t.Error("expected nil config for non-existent host")
	}
}

func TestHostsConfig_UpdateHostUrisConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
	}

	// Update URIs config
	newUrisConfig := map[string]UriConfig{
		"/api/v1/users": {
			LatencyConfig: &LatencyConfig{
				Min: intPtr(50),
				Max: intPtr(100),
			},
		},
		"/api/v1/orders": {
			ErrorsConfig: map[string]ErrorConfig{
				"404": {
					Percentage: intPtr(15),
				},
			},
		},
	}

	updatedConfig, err := hostsConfig.UpdateHostUrisConfig("example.com", newUrisConfig)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedConfig == nil {
		t.Fatal("expected updated config to be non-nil")
	}

	if len(updatedConfig.UrisConfig) != 2 {
		t.Errorf("expected 2 URI configs, got %d", len(updatedConfig.UrisConfig))
	}

	usersConfig, exists := updatedConfig.UrisConfig["/api/v1/users"]
	if !exists {
		t.Error("expected /api/v1/users config to exist")
	} else if *usersConfig.LatencyConfig.Min != 50 {
		t.Errorf("expected users URI min latency to be 50, got %d", *usersConfig.LatencyConfig.Min)
	}

	// Update non-existent host
	updatedConfig, err = hostsConfig.UpdateHostUrisConfig("nonexistent.com", newUrisConfig)

	if err != nil {
		t.Errorf("expected no error for non-existent host, got %v", err)
	}

	if updatedConfig != nil {
		t.Error("expected nil config for non-existent host")
	}
}

func TestHostsConfig_GetAppropriateErrorsConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(20),
					},
				},
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						ErrorsConfig: map[string]ErrorConfig{
							"404": {
								Percentage: intPtr(15),
							},
						},
					},
					"/api/v1/orders": {
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
				},
			},
		},
	}

	// Test URI with specific errors config (should override host errors)
	errorsConfig := hostsConfig.GetAppropriateErrorsConfig("example.com", "/api/v1/users")
	if errorsConfig == nil {
		t.Fatal("expected errors config to be non-nil")
	}

	if len(*errorsConfig) != 1 {
		t.Errorf("expected 1 error config, got %d", len(*errorsConfig))
	}

	if _, exists := (*errorsConfig)["404"]; !exists {
		t.Error("expected 404 error config to exist")
	}

	// Test URI without specific errors config (should use host errors)
	errorsConfig = hostsConfig.GetAppropriateErrorsConfig("example.com", "/api/v1/orders")
	if errorsConfig == nil {
		t.Fatal("expected errors config to be non-nil")
	}

	if len(*errorsConfig) != 1 {
		t.Errorf("expected 1 error config, got %d", len(*errorsConfig))
	}

	if _, exists := (*errorsConfig)["500"]; !exists {
		t.Error("expected 500 error config to exist")
	}

	// Test URI that doesn't exist (should use host errors)
	errorsConfig = hostsConfig.GetAppropriateErrorsConfig("example.com", "/nonexistent")
	if errorsConfig == nil {
		t.Fatal("expected errors config to be non-nil")
	}

	if _, exists := (*errorsConfig)["500"]; !exists {
		t.Error("expected 500 error config to exist")
	}

	// Test non-existent host
	errorsConfig = hostsConfig.GetAppropriateErrorsConfig("nonexistent.com", "/api/v1/users")
	if errorsConfig != nil {
		t.Error("expected nil errors config for non-existent host")
	}
}

func TestHostsConfig_GetAppropriateErrorsConfig_EmptyConfigs(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				// No errors config at host level
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						// No errors config at URI level either
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
				},
			},
		},
	}

	// Test when no errors config exists at any level
	errorsConfig := hostsConfig.GetAppropriateErrorsConfig("example.com", "/api/v1/users")
	if errorsConfig != nil {
		t.Error("expected nil errors config when none exist")
	}
}

func TestHostsConfig_GetAppropriateLatencyConfig(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
					"/api/v1/orders": {
						ErrorsConfig: map[string]ErrorConfig{
							"404": {
								Percentage: intPtr(15),
							},
						},
					},
				},
			},
		},
	}

	// Test URI with specific latency config (should override host latency)
	latencyConfig := hostsConfig.GetAppropriateLatencyConfig("example.com", "/api/v1/users")
	if latencyConfig == nil {
		t.Fatal("expected latency config to be non-nil")
	}

	if *latencyConfig.Min != 50 {
		t.Errorf("expected min latency to be 50, got %d", *latencyConfig.Min)
	}

	// Test URI without specific latency config (should use host latency)
	latencyConfig = hostsConfig.GetAppropriateLatencyConfig("example.com", "/api/v1/orders")
	if latencyConfig == nil {
		t.Fatal("expected latency config to be non-nil")
	}

	if *latencyConfig.Min != 100 {
		t.Errorf("expected min latency to be 100, got %d", *latencyConfig.Min)
	}

	// Test URI that doesn't exist (should use host latency)
	latencyConfig = hostsConfig.GetAppropriateLatencyConfig("example.com", "/nonexistent")
	if latencyConfig == nil {
		t.Fatal("expected latency config to be non-nil")
	}

	if *latencyConfig.Min != 100 {
		t.Errorf("expected min latency to be 100, got %d", *latencyConfig.Min)
	}

	// Test non-existent host
	latencyConfig = hostsConfig.GetAppropriateLatencyConfig("nonexistent.com", "/api/v1/users")
	if latencyConfig != nil {
		t.Error("expected nil latency config for non-existent host")
	}
}

func TestHostsConfig_GetAppropriateLatencyConfig_NoHostLatency(t *testing.T) {
	hostsConfig := &HostsConfig{
		Hosts: map[string]HostConfig{
			"example.com": {
				// No latency config at host level
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
					"/api/v1/orders": {
						ErrorsConfig: map[string]ErrorConfig{
							"404": {
								Percentage: intPtr(15),
							},
						},
					},
				},
			},
		},
	}

	// Test URI with specific latency config
	latencyConfig := hostsConfig.GetAppropriateLatencyConfig("example.com", "/api/v1/users")
	if latencyConfig == nil {
		t.Fatal("expected latency config to be non-nil")
	}

	if *latencyConfig.Min != 50 {
		t.Errorf("expected min latency to be 50, got %d", *latencyConfig.Min)
	}

	// Test URI without latency config when host also has none
	latencyConfig = hostsConfig.GetAppropriateLatencyConfig("example.com", "/api/v1/orders")
	if latencyConfig != nil {
		t.Error("expected nil latency config when none exist")
	}
}
