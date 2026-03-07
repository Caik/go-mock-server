package mock

import (
	"encoding/json"
	"testing"

	"github.com/Caik/go-mock-server/internal/service/cache"
	"github.com/Caik/go-mock-server/internal/service/content"
)

// inMemoryCacheService is a simple in-memory cache for testing
type inMemoryCacheService struct {
	data map[string][]byte
}

func (c *inMemoryCacheService) Get(key, uuid string) (*[]byte, bool) {
	data, exists := c.data[key]

	if !exists {
		return nil, false
	}

	return &data, true
}

func (c *inMemoryCacheService) Set(key string, data *[]byte, uuid string) {
	if c.data == nil {
		c.data = make(map[string][]byte)
	}

	c.data[key] = *data
}

var _ cache.CacheService = (*inMemoryCacheService)(nil)

func TestCacheMockService_getMockResponse_cacheMiss(t *testing.T) {
	t.Run("cache miss fetches from next service and caches result", func(t *testing.T) {
		cacheStore := &inMemoryCacheService{data: make(map[string][]byte)}
		svc := newCacheMockService(cacheStore)

		data := []byte(`{"key":"value"}`)
		nextSvc := &contentMockService{
			contentService: &mockContentService{
				contents: map[string][]byte{
					"example.com:/api/test:GET": data,
				},
				events: make(chan content.ContentEvent),
			},
		}

		svc.setNext(nextSvc)

		req := MockRequest{
			Host:   "example.com",
			URI:    "/api/test",
			Method: "GET",
			Uuid:   "test-uuid",
		}

		resp := svc.getMockResponse(req)

		if resp == nil {
			t.Fatal("expected non-nil response")
		}

		if resp.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestCacheMockService_getMockResponse_cacheHit(t *testing.T) {
	t.Run("cache hit returns cached response with source=cache", func(t *testing.T) {
		// Pre-populate cache with a serialized response
		cachedResp := MockResponse{
			StatusCode:  200,
			ContentType: "application/json",
			Metadata:    map[string]string{MetadataMatched: "true", MetadataSource: "mock"},
		}
		serialized, err := json.Marshal(cachedResp)

		if err != nil {
			t.Fatalf("failed to serialize: %v", err)
		}

		req := MockRequest{
			Host:   "example.com",
			URI:    "/api/test",
			Method: "GET",
			Uuid:   "test-uuid",
		}
		cacheKey := GenerateCacheKey(req)

		cacheStore := &inMemoryCacheService{
			data: map[string][]byte{cacheKey: serialized},
		}
		svc := newCacheMockService(cacheStore)

		// next service returns something different - should NOT be used for cache hit
		emptyData := []byte("should not be returned")
		nextSvc := &contentMockService{
			contentService: &mockContentService{
				contents: map[string][]byte{
					"example.com:/api/test:GET": emptyData,
				},
				events: make(chan content.ContentEvent),
			},
		}

		svc.setNext(nextSvc)
		resp := svc.getMockResponse(req)

		if resp == nil {
			t.Fatal("expected non-nil response")
		}

		if resp.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Cache hit should override Source to "cache"
		if resp.Metadata[MetadataSource] != "cache" {
			t.Errorf("expected Source=cache, got %q", resp.Metadata[MetadataSource])
		}

		// CacheKey should be stored in Path
		if resp.Metadata[MetadataPath] != cacheKey {
			t.Errorf("expected Path=%q, got %q", cacheKey, resp.Metadata[MetadataPath])
		}
	})
}

func TestCacheMockService_getMockResponse_cacheHitBadJSON(t *testing.T) {
	t.Run("cache hit with invalid JSON falls back to next service", func(t *testing.T) {
		req := MockRequest{
			Host:   "example.com",
			URI:    "/api/test",
			Method: "GET",
			Uuid:   "test-uuid",
		}
		cacheKey := GenerateCacheKey(req)

		cacheStore := &inMemoryCacheService{
			data: map[string][]byte{cacheKey: []byte("not-valid-json")},
		}

		svc := newCacheMockService(cacheStore)

		data := []byte(`{"fallback":true}`)
		nextSvc := &contentMockService{
			contentService: &mockContentService{
				contents: map[string][]byte{
					"example.com:/api/test:GET": data,
				},
				events: make(chan content.ContentEvent),
			},
		}

		svc.setNext(nextSvc)
		resp := svc.getMockResponse(req)

		// Should fall back to next service
		if resp == nil {
			t.Fatal("expected non-nil response from fallback")
		}

		if resp.StatusCode != 200 {
			t.Errorf("expected status 200 from fallback, got %d", resp.StatusCode)
		}
	})
}

func TestCacheMockService_deserialize(t *testing.T) {
	t.Run("deserializes valid JSON", func(t *testing.T) {
		svc := &cacheMockService{}
		original := MockResponse{
			StatusCode:  201,
			ContentType: "application/json",
			Metadata:    map[string]string{MetadataMatched: "true"},
		}

		data, _ := json.Marshal(original)
		result, err := svc.deserialize(&data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.StatusCode != 201 {
			t.Errorf("expected StatusCode 201, got %d", result.StatusCode)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		svc := &cacheMockService{}
		bad := []byte("not json")
		_, err := svc.deserialize(&bad)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
