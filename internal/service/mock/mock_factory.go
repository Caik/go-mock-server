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

		// latency
		if !appConfig.DisableLatency {
			addNextFn(&latencyMockService{})
		}

		// error
		if !appConfig.DisableError {
			addNextFn(&errorMockService{})
		}

		// content type
		addNextFn(&contentTypeMockService{})

		// cache
		if !appConfig.DisableCache {
			addNextFn(&cacheMockService{})
		}

		// file
		addNextFn(&fileMockService{})

		// setting the chain
		mockServiceChain = first
	})
}
