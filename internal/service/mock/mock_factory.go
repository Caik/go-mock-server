package mock

import (
	"fmt"
	"log"
	"sync"

	"github.com/Caik/go-mock-server/internal/config"
)

var (
	once             sync.Once
	mockServiceChain mockService
)

func GetMockResponse(mockRequest MockRequest) *MockResponse {
	ensureInit()

	return mockServiceChain.getMockResponse(mockRequest)
}

func ensureInit() {
	if mockServiceChain != nil {
		return
	}

	once.Do(func() {
		appConfig, err := config.GetAppConfig()

		if err != nil {
			log.Fatalf(fmt.Sprintf("error while getting app config: %v", err))
		}

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

		// TODO: add CORS MockService

		// host resolution
		addNextFn(newHostResolutionMockService())

		// latency
		if !appConfig.DisableLatency {
			addNextFn(newLatencyMockService())
		}

		// error
		if !appConfig.DisableError {
			addNextFn(newErrorMockService())
		}

		// content type
		addNextFn(newContentTypeMockService())

		// cache
		if !appConfig.DisableCache {
			addNextFn(newCacheMockService())
		}

		// content
		addNextFn(newContentMockService())

		// setting the chain
		mockServiceChain = first
	})
}
