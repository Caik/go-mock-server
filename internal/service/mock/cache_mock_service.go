package mock

import (
	"encoding/json"
	"fmt"

	"github.com/Caik/go-mock-server/internal/service/cache"
	log "github.com/sirupsen/logrus"
)

type cacheMockService struct {
	next mockService
}

func (c cacheMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	cacheService := cache.GetCacheService()

	if cacheService == nil {
		log.WithField("uuid", mockRequest.Uuid).
			Warn("bad configuration found, cache service is nil!")

		return c.nextOrNil(mockRequest)
	}

	cacheKey := GenerateCacheKey(mockRequest)
	data, exists := cacheService.Get(cacheKey, mockRequest.Uuid)

	if !exists {
		return c.refreshCache(mockRequest, cacheKey, cacheService)
	}

	// deserializing the data to mockResponse
	mockResponse, err := c.deserialize(data)

	if err != nil {
		log.WithField("uuid", mockRequest.Uuid).
			Error(fmt.Sprintf("error while deserializing data from cache: %v", err))

		return c.nextOrNil(mockRequest)
	}

	log.WithField("uuid", mockRequest.Uuid).
		Info("found mock response on cache")

	// background cache refresh
	go c.refreshCache(mockRequest, cacheKey, cacheService)

	return &mockResponse
}

func (c *cacheMockService) setNext(next mockService) {
	c.next = next
}

func (c cacheMockService) deserialize(data *[]byte) (MockResponse, error) {
	var mockResponse MockResponse

	err := json.Unmarshal(*data, &mockResponse)

	return mockResponse, err
}

func (c cacheMockService) serialize(mockResponse *MockResponse) ([]byte, error) {
	return json.Marshal(mockResponse)
}

func (c cacheMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if c.next == nil {
		return nil
	}

	return c.next.getMockResponse(mockRequest)
}

func (c cacheMockService) refreshCache(mockRequest MockRequest, cacheKey string, cacheService cache.CacheService) *MockResponse {
	freshResponse := c.nextOrNil(mockRequest)

	if freshResponse == nil {
		return nil
	}

	log.WithField("host", mockRequest.Host).
		WithField("uri", mockRequest.URI).
		WithField("method", mockRequest.Method).
		Info("refreshing cache data")

	serializedData, err := c.serialize(freshResponse)

	if err != nil {
		log.WithField("uuid", mockRequest.Uuid).
			Error(fmt.Sprintf("error while serializing data to cache: %v", err))

		return freshResponse
	}

	cacheService.Set(cacheKey, &serializedData, mockRequest.Uuid)

	return freshResponse
}

func NewCacheMockService() *cacheMockService {
	// TODO listen to content changes to remove stuff from cache
	return &cacheMockService{}
}
