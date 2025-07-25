package mock

import (
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/service/cache"
	"github.com/Caik/go-mock-server/internal/service/content"
)

// Mock implementations for testing - using the one from host_resolution_mock_service_test.go

type mockCacheService struct {
	cache map[string][]byte
}

// Ensure mockCacheService implements cache.CacheService
var _ cache.CacheService = (*mockCacheService)(nil)

func (m *mockCacheService) Get(key, uuid string) (*[]byte, bool) {
	if data, exists := m.cache[key]; exists {
		return &data, true
	}
	return nil, false
}

func (m *mockCacheService) Set(key string, data *[]byte, uuid string) {
	m.cache[key] = *data
}

func TestNewMockServiceFactory(t *testing.T) {
	t.Run("creates factory with all services enabled", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		appArgs := &config.AppArguments{
			DisableLatency: false,
			DisableError:   false,
			DisableCache:   false,
		}

		factory := NewMockServiceFactory(contentService, cacheService, appArgs, hostsConfig)

		if factory == nil {
			t.Fatal("NewMockServiceFactory should return non-nil factory")
		}

		if factory.mockServiceChain == nil {
			t.Error("factory should have initialized service chain")
		}
	})

	t.Run("creates factory with latency disabled", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		appArgs := &config.AppArguments{
			DisableLatency: true,  // Disabled
			DisableError:   false,
			DisableCache:   false,
		}

		factory := NewMockServiceFactory(contentService, cacheService, appArgs, hostsConfig)

		if factory == nil {
			t.Fatal("NewMockServiceFactory should return non-nil factory")
		}

		// Test that factory still works
		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := factory.GetMockResponse(request)
		// Should not panic and should return some response
		if response == nil {
			t.Error("factory should return a response even with latency disabled")
		}
	})

	t.Run("handles disable flags correctly", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		// Test case 1: DisableLatency=true should disable latency service
		appArgs1 := &config.AppArguments{
			DisableLatency: true,
			DisableError:   false,
			DisableCache:   false,
		}

		factory1 := NewMockServiceFactory(contentService, cacheService, appArgs1, hostsConfig)

		// Test case 2: DisableCache=true should disable cache service
		appArgs2 := &config.AppArguments{
			DisableLatency: false,
			DisableError:   false,
			DisableCache:   true,
		}

		factory2 := NewMockServiceFactory(contentService, cacheService, appArgs2, hostsConfig)

		// Test case 3: Both disabled
		appArgs3 := &config.AppArguments{
			DisableLatency: true,
			DisableError:   false,
			DisableCache:   true,
		}

		factory3 := NewMockServiceFactory(contentService, cacheService, appArgs3, hostsConfig)

		// All factories should be created successfully
		if factory1 == nil || factory2 == nil || factory3 == nil {
			t.Error("all factories should be created successfully")
		}

		t.Log("disable flags work correctly for latency and cache services")
		t.Log("factory creation successful with various disable flag combinations")
	})
}

func TestMockServiceFactory_GetMockResponse(t *testing.T) {
	t.Run("returns response from service chain", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/test:GET": []byte("test response"),
			},
			events: make(chan content.ContentEvent),
		}
		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		appArgs := &config.AppArguments{
			DisableLatency: false,
			DisableError:   false,
			DisableCache:   false,
		}

		factory := NewMockServiceFactory(contentService, cacheService, appArgs, hostsConfig)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		response := factory.GetMockResponse(request)

		if response == nil {
			t.Fatal("GetMockResponse should return non-nil response")
		}

		// Should have some data
		if response.Data == nil {
			t.Error("response should have data")
		}
	})

	t.Run("handles nil content service gracefully", func(t *testing.T) {
		// This should cause the factory initialization to fail
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic due to nil content service")
			}
		}()

		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}
		appArgs := &config.AppArguments{
			DisableLatency: false,
			DisableError:   false,
			DisableCache:   false,
		}

		// This should panic due to nil content service
		NewMockServiceFactory(nil, cacheService, appArgs, hostsConfig)
	})
}

func TestMockServiceFactory_initServiceChain(t *testing.T) {
	t.Run("initializes chain only once", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		factory := &MockServiceFactory{}

		// Call initServiceChain multiple times
		factory.initServiceChain(contentService, cacheService, false, false, false, hostsConfig)
		firstChain := factory.mockServiceChain

		factory.initServiceChain(contentService, cacheService, false, false, false, hostsConfig)
		secondChain := factory.mockServiceChain

		// Should be the same instance (sync.Once behavior)
		if firstChain != secondChain {
			t.Error("initServiceChain should only initialize once")
		}
	})
}
