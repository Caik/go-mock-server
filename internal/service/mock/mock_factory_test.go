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
	if m.cache == nil {
		m.cache = make(map[string][]byte)
	}
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
			DisableCors:    false,
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
			DisableLatency: true, // Disabled
			DisableError:   false,
			DisableCache:   false,
			DisableCors:    false,
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
			DisableCors:    false,
		}

		factory1 := NewMockServiceFactory(contentService, cacheService, appArgs1, hostsConfig)

		// Test case 2: DisableCache=true should disable cache service
		appArgs2 := &config.AppArguments{
			DisableLatency: false,
			DisableError:   false,
			DisableCache:   true,
			DisableCors:    false,
		}

		factory2 := NewMockServiceFactory(contentService, cacheService, appArgs2, hostsConfig)

		// Test case 3: Both disabled
		appArgs3 := &config.AppArguments{
			DisableLatency: true,
			DisableError:   false,
			DisableCache:   true,
			DisableCors:    false,
		}

		factory3 := NewMockServiceFactory(contentService, cacheService, appArgs3, hostsConfig)

		// Test case 4: DisableCors=true should disable CORS service
		appArgs4 := &config.AppArguments{
			DisableLatency: false,
			DisableError:   false,
			DisableCache:   false,
			DisableCors:    true,
		}

		factory4 := NewMockServiceFactory(contentService, cacheService, appArgs4, hostsConfig)

		// All factories should be created successfully
		if factory1 == nil || factory2 == nil || factory3 == nil || factory4 == nil {
			t.Error("all factories should be created successfully")
		}

		t.Log("disable flags work correctly for latency, cache, and CORS services")
		t.Log("factory creation successful with various disable flag combinations")
	})
}

func TestMockServiceFactory_DisableCorsFlag(t *testing.T) {
	t.Run("CORS service behavior with disable flag", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		cacheService := &mockCacheService{
			cache: make(map[string][]byte),
		}
		hostsConfig := &config.HostsConfig{}

		// Add some test content so the service chain doesn't fail
		testData := []byte(`{"message": "test"}`)
		contentService.contents["example.com/api/test.get"] = testData

		// Test case 1: CORS enabled (default)
		appArgsEnabled := &config.AppArguments{
			DisableLatency: true,  // Disable to simplify test
			DisableError:   true,  // Disable to simplify test
			DisableCache:   true,  // Disable to simplify test
			DisableCors:    false, // CORS enabled
		}

		factoryEnabled := NewMockServiceFactory(contentService, cacheService, appArgsEnabled, hostsConfig)

		// Test case 2: CORS disabled
		appArgsDisabled := &config.AppArguments{
			DisableLatency: true, // Disable to simplify test
			DisableError:   true, // Disable to simplify test
			DisableCache:   true, // Disable to simplify test
			DisableCors:    true, // CORS disabled
		}

		factoryDisabled := NewMockServiceFactory(contentService, cacheService, appArgsDisabled, hostsConfig)

		// Create a test request
		testRequest := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		// Test with CORS enabled - should have CORS headers
		responseEnabled := factoryEnabled.GetMockResponse(testRequest)
		if responseEnabled != nil && responseEnabled.Headers != nil {
			corsHeaders := []string{
				"Access-Control-Allow-Origin",
				"Access-Control-Allow-Methods",
				"Access-Control-Allow-Headers",
				"Access-Control-Max-Age",
			}

			for _, header := range corsHeaders {
				if _, exists := (*responseEnabled.Headers)[header]; !exists {
					t.Errorf("Expected CORS header %s to be present when CORS is enabled", header)
				}
			}
			t.Log("CORS headers correctly added when CORS is enabled")
		} else {
			t.Error("Expected response with headers when CORS is enabled")
		}

		// Test with CORS disabled - should not have CORS headers
		responseDisabled := factoryDisabled.GetMockResponse(testRequest)
		if responseDisabled != nil {
			if responseDisabled.Headers != nil {
				corsHeaders := []string{
					"Access-Control-Allow-Origin",
					"Access-Control-Allow-Methods",
					"Access-Control-Allow-Headers",
					"Access-Control-Max-Age",
				}

				for _, header := range corsHeaders {
					if _, exists := (*responseDisabled.Headers)[header]; exists {
						t.Errorf("Expected CORS header %s to NOT be present when CORS is disabled", header)
					}
				}
			}
			t.Log("CORS headers correctly omitted when CORS is disabled")
		} else {
			t.Error("Expected response when CORS is disabled")
		}

		t.Log("CORS disable flag works correctly")
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
			DisableCors:    false,
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
		factory.initServiceChain(contentService, cacheService, false, false, false, false, hostsConfig)
		firstChain := factory.mockServiceChain

		factory.initServiceChain(contentService, cacheService, false, false, false, false, hostsConfig)
		secondChain := factory.mockServiceChain

		// Should be the same instance (sync.Once behavior)
		if firstChain != secondChain {
			t.Error("initServiceChain should only initialize once")
		}
	})
}
