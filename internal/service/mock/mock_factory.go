package mock

import (
	"fmt"
	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/service/cache"
	"github.com/Caik/go-mock-server/internal/service/content"
	"log"
	"sync"
)

type MockServiceFactory struct {
	once             sync.Once
	mockServiceChain mockService
}

func (m *MockServiceFactory) GetMockResponse(mockRequest MockRequest) *MockResponse {
	return m.mockServiceChain.getMockResponse(mockRequest)
}

func (m *MockServiceFactory) initServiceChain(contentService content.ContentService, cacheService cache.CacheService, disableLatency, disableErrors, disableCache bool, hostsConfig *config.HostsConfig) {
	if m.mockServiceChain != nil {
		return
	}

	m.once.Do(func() {
		var first mockService
		var last mockService

		addNextFn := func(next mockService) {
			if first == nil {
				first = next
			}

			if last != nil {
				last.setNext(next)
			}

			last = next
		}

		// host resolution
		hostResolutionMockService, err := newHostResolutionMockService(contentService)

		if err != nil {
			log.Fatalf(fmt.Sprintf("error while starting HostResolutionMockService: %v", err))
		}

		addNextFn(hostResolutionMockService)

		// TODO: add CORS MockService

		// latency
		if !disableLatency {
			addNextFn(newLatencyMockService(hostsConfig))
		}

		// errors
		if !disableErrors {
			addNextFn(newErrorMockService(hostsConfig))
		}

		// content type
		addNextFn(newContentTypeMockService())

		// cache
		if !disableCache {
			addNextFn(newCacheMockService(cacheService))
		}

		// content
		addNextFn(newContentMockService(contentService))

		// setting the chain
		m.mockServiceChain = first
	})
}

func NewMockServiceFactory(contentService content.ContentService, cacheService cache.CacheService, arguments *config.AppArguments, hostsConfig *config.HostsConfig) *MockServiceFactory {
	factory := MockServiceFactory{}
	factory.initServiceChain(contentService, cacheService, arguments.DisableLatency, arguments.DisableError, arguments.DisableCache, hostsConfig)

	return &factory
}
