package mock

import (
	"encoding/json"
	"github.com/Caik/go-mock-server/internal/service/cache"
	"github.com/rs/zerolog/log"
)

type cacheMockService struct {
	next         mockService
	cacheService cache.CacheService
}

func (c *cacheMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	cacheKey := GenerateCacheKey(mockRequest)
	data, exists := c.cacheService.Get(cacheKey, mockRequest.Uuid)

	if !exists {
		return c.refreshCache(mockRequest, cacheKey)
	}

	// deserializing the data to mockResponse
	mockResponse, err := c.deserialize(data)

	if err != nil {
		log.Err(err).
			Stack().
			Str("uuid", mockRequest.Uuid).
			Msg("error while deserializing data from cache")

		return c.nextOrNil(mockRequest)
	}

	log.Info().
		Str("uuid", mockRequest.Uuid).
		Msg("found mock response on cache")

	// background cache refresh
	go c.refreshCache(mockRequest, cacheKey)

	return &mockResponse
}

func (c *cacheMockService) setNext(next mockService) {
	c.next = next
}

func (c *cacheMockService) deserialize(data *[]byte) (MockResponse, error) {
	var mockResponse MockResponse

	err := json.Unmarshal(*data, &mockResponse)

	return mockResponse, err
}

func (c *cacheMockService) serialize(mockResponse *MockResponse) ([]byte, error) {
	return json.Marshal(mockResponse)
}

func (c *cacheMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if c.next == nil {
		return nil
	}

	return c.next.getMockResponse(mockRequest)
}

func (c *cacheMockService) refreshCache(mockRequest MockRequest, cacheKey string) *MockResponse {
	freshResponse := c.nextOrNil(mockRequest)

	if freshResponse == nil {
		return nil
	}

	log.Info().
		Str("host", mockRequest.Host).
		Str("uri", mockRequest.URI).
		Str("method", mockRequest.Method).
		Msg("refreshing cache data")

	serializedData, err := c.serialize(freshResponse)

	if err != nil {
		log.Err(err).
			Stack().
			Str("uuid", mockRequest.Uuid).
			Msg("error while serializing data to cache")

		return freshResponse
	}

	c.cacheService.Set(cacheKey, &serializedData, mockRequest.Uuid)

	return freshResponse
}

func newCacheMockService(cacheService cache.CacheService) *cacheMockService {
	// TODO listen to content changes to remove stuff from cache
	return &cacheMockService{
		cacheService: cacheService,
	}
}
