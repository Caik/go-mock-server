package admin

import (
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
)

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

func TestNewHostsConfigAdminService(t *testing.T) {
	t.Run("creates service with hosts config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)

		if service == nil {
			t.Fatal("NewHostsConfigAdminService should return non-nil service")
		}

		if service.hostsConfig != hostsConfig {
			t.Error("service should store the provided hosts config")
		}
	})

	t.Run("handles nil hosts config", func(t *testing.T) {
		service := NewHostsConfigAdminService(nil)

		if service == nil {
			t.Fatal("NewHostsConfigAdminService should return non-nil service even with nil config")
		}

		if service.hostsConfig != nil {
			t.Error("service should store nil hosts config")
		}
	})
}

func TestHostsConfigAdminService_GetHostsConfig(t *testing.T) {
	t.Run("returns stored hosts config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(10),
						Max: intPtr(20),
					},
				},
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)
		result := service.GetHostsConfig()

		if result != hostsConfig {
			t.Error("GetHostsConfig should return the same hosts config instance")
		}
	})

	t.Run("returns nil when hosts config is nil", func(t *testing.T) {
		service := NewHostsConfigAdminService(nil)
		result := service.GetHostsConfig()

		if result != nil {
			t.Error("GetHostsConfig should return nil when hosts config is nil")
		}
	})
}

func TestHostsConfigAdminService_GetHostConfig(t *testing.T) {
	t.Run("returns existing host config", func(t *testing.T) {
		expectedConfig := config.HostConfig{
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(10),
				Max: intPtr(20),
			},
		}

		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": expectedConfig,
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)
		result := service.GetHostConfig("example.com")

		if result == nil {
			t.Fatal("GetHostConfig should return non-nil config for existing host")
		}

		if result.LatencyConfig == nil {
			t.Error("returned config should have latency config")
		}

		if *result.LatencyConfig.Min != 10 {
			t.Errorf("expected Min latency 10, got %d", *result.LatencyConfig.Min)
		}
	})

	t.Run("returns nil for non-existent host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)
		result := service.GetHostConfig("non-existent.com")

		if result != nil {
			t.Error("GetHostConfig should return nil for non-existent host")
		}
	})
}

func TestHostsConfigAdminService_AddUpdateHost(t *testing.T) {
	t.Run("adds new host successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)

		request := HostAddDeleteRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(10),
				Max: intPtr(20),
			},
		}

		result, err := service.AddUpdateHost(request)

		if err != nil {
			t.Fatalf("AddUpdateHost should not return error: %v", err)
		}

		if result == nil {
			t.Fatal("AddUpdateHost should return non-nil host config")
		}

		// Verify host was added to hosts config
		storedConfig := hostsConfig.GetHostConfig("example.com")
		if storedConfig == nil {
			t.Error("host should be added to hosts config")
		}
	})

	t.Run("updates existing host successfully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(5),
						Max: intPtr(10),
					},
				},
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)

		request := HostAddDeleteRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(20),
				Max: intPtr(30),
			},
		}

		result, err := service.AddUpdateHost(request)

		if err != nil {
			t.Fatalf("AddUpdateHost should not return error: %v", err)
		}

		if result == nil {
			t.Fatal("AddUpdateHost should return non-nil host config")
		}

		// Verify host was updated
		storedConfig := hostsConfig.GetHostConfig("example.com")
		if storedConfig == nil {
			t.Fatal("host should exist in hosts config")
		}

		if *storedConfig.LatencyConfig.Min != 20 {
			t.Errorf("expected updated Min latency 20, got %d", *storedConfig.LatencyConfig.Min)
		}
	})

	t.Run("returns error for invalid host config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)

		request := HostAddDeleteRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(20), // Invalid: Min > Max
				Max: intPtr(10),
			},
		}

		result, err := service.AddUpdateHost(request)

		if err == nil {
			t.Error("AddUpdateHost should return error for invalid config")
		}

		if result != nil {
			t.Error("AddUpdateHost should return nil result on error")
		}

		// Verify host was not added
		storedConfig := hostsConfig.GetHostConfig("example.com")
		if storedConfig != nil {
			t.Error("invalid host should not be added to hosts config")
		}
	})

	t.Run("handles complex host config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)

		request := HostAddDeleteRequest{
			Host: "api.example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(10),
				Max: intPtr(100),
				P95: intPtr(80),
				P99: intPtr(95),
			},
			ErrorConfig: map[string]config.ErrorConfig{
				"500": {
					Percentage: intPtr(5),
				},
				"404": {
					Percentage: intPtr(2),
				},
			},
			UriConfig: map[string]config.UriConfig{
				"/api/users": {
					ErrorsConfig: map[string]config.ErrorConfig{
						"400": {
							Percentage: intPtr(10),
						},
					},
				},
			},
		}

		result, err := service.AddUpdateHost(request)

		if err != nil {
			t.Fatalf("AddUpdateHost should not return error for valid complex config: %v", err)
		}

		if result == nil {
			t.Fatal("AddUpdateHost should return non-nil host config")
		}

		// Verify complex config was stored correctly
		storedConfig := hostsConfig.GetHostConfig("api.example.com")
		if storedConfig == nil {
			t.Fatal("host should be added to hosts config")
		}

		if len(storedConfig.ErrorsConfig) != 2 {
			t.Errorf("expected 2 error configs, got %d", len(storedConfig.ErrorsConfig))
		}

		if len(storedConfig.UrisConfig) != 1 {
			t.Errorf("expected 1 URI config, got %d", len(storedConfig.UrisConfig))
		}
	})
}

func TestHostsConfigAdminService_DeleteHost(t *testing.T) {
	t.Run("deletes existing host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(10),
						Max: intPtr(20),
					},
				},
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)

		// Verify host exists before deletion
		if hostsConfig.GetHostConfig("example.com") == nil {
			t.Fatal("host should exist before deletion")
		}

		service.DeleteHost("example.com")

		// Verify host was deleted
		if hostsConfig.GetHostConfig("example.com") != nil {
			t.Error("host should be deleted from hosts config")
		}
	})

	t.Run("handles deletion of non-existent host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)

		// Should not panic or error
		service.DeleteHost("non-existent.com")

		// Verify hosts config is still empty
		if len(hostsConfig.Hosts) != 0 {
			t.Error("hosts config should remain empty")
		}
	})
}

func TestHostsConfigAdminService_AddUpdateHostLatency(t *testing.T) {
	t.Run("adds latency config to existing host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {}, // Host exists but no latency config
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)

		request := HostAddDeleteRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(10),
				Max: intPtr(20),
			},
		}

		result, err := service.AddUpdateHostLatency(request)

		if err != nil {
			t.Fatalf("AddUpdateHostLatency should not return error: %v", err)
		}

		if result == nil {
			t.Fatal("AddUpdateHostLatency should return non-nil host config")
		}

		// Verify latency config was added
		storedConfig := hostsConfig.GetHostConfig("example.com")
		if storedConfig == nil {
			t.Fatal("host should exist in hosts config")
		}

		if storedConfig.LatencyConfig == nil {
			t.Error("latency config should be added")
		}

		if *storedConfig.LatencyConfig.Min != 10 {
			t.Errorf("expected Min latency 10, got %d", *storedConfig.LatencyConfig.Min)
		}
	})

	t.Run("returns error for invalid latency config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {},
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)

		request := HostAddDeleteRequest{
			Host: "example.com",
			LatencyConfig: &config.LatencyConfig{
				Min: intPtr(20), // Invalid: Min > Max
				Max: intPtr(10),
			},
		}

		result, err := service.AddUpdateHostLatency(request)

		if err == nil {
			t.Error("AddUpdateHostLatency should return error for invalid config")
		}

		if result != nil {
			t.Error("AddUpdateHostLatency should return nil result on error")
		}
	})
}

func TestHostsConfigAdminService_DeleteHostLatency(t *testing.T) {
	t.Run("deletes latency config from existing host", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					LatencyConfig: &config.LatencyConfig{
						Min: intPtr(10),
						Max: intPtr(20),
					},
				},
			},
		}

		service := NewHostsConfigAdminService(hostsConfig)

		result, err := service.DeleteHostLatency("example.com")

		if err != nil {
			t.Fatalf("DeleteHostLatency should not return error: %v", err)
		}

		if result == nil {
			t.Fatal("DeleteHostLatency should return non-nil host config")
		}

		// Verify latency config was deleted
		storedConfig := hostsConfig.GetHostConfig("example.com")
		if storedConfig == nil {
			t.Fatal("host should still exist in hosts config")
		}

		if storedConfig.LatencyConfig != nil {
			t.Error("latency config should be deleted")
		}
	})

	t.Run("handles non-existent host gracefully", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := NewHostsConfigAdminService(hostsConfig)

		result, err := service.DeleteHostLatency("non-existent.com")

		if err != nil {
			t.Errorf("DeleteHostLatency should not return error for non-existent host: %v", err)
		}

		if result != nil {
			t.Error("DeleteHostLatency should return nil result for non-existent host")
		}
	})
}
